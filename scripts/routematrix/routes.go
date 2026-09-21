// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// route is one exported *Client method: one HTTP call to the Forgejo API.
type route struct {
	Name string
	File string // base name, e.g. "repo_tag.go"
	Line int
	Doc  string // first line of the godoc block, method name stripped
	// Guard is the minimum server version the SDK itself enforces, from an
	// unconditional checkServerVersionGreaterThanOrEqual. Empty when the
	// method has no guard.
	Guard string
}

// parseRoutes reads every non-test .go file in dir and returns its routes,
// sorted by name. It also returns the routes that have no godoc at all: the
// caller fails on those rather than shipping a table with an empty column.
func parseRoutes(dir string) (routes []route, undocumented []route, err error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		return nil, nil, err
	}

	fset := token.NewFileSet()
	files := map[string]*ast.File{}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, nil, err
		}
		files[path] = f
	}

	versions := map[string]string{}
	for _, f := range files {
		collectVersionConsts(f, versions)
	}

	for path, f := range files {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !isClientMethod(fn) || !ast.IsExported(fn.Name.Name) {
				continue
			}

			r := route{
				Name:  fn.Name.Name,
				File:  filepath.Base(path),
				Line:  fset.Position(fn.Pos()).Line,
				Doc:   docSummary(fn),
				Guard: guardVersion(fn, versions),
			}
			if r.Doc == "" {
				undocumented = append(undocumented, r)
			}
			routes = append(routes, r)
		}
	}

	sort.Slice(routes, func(i, j int) bool { return routes[i].Name < routes[j].Name })
	sort.Slice(undocumented, func(i, j int) bool { return undocumented[i].Name < undocumented[j].Name })
	return routes, undocumented, nil
}

// isClientMethod reports whether fn is a method on *Client.
func isClientMethod(fn *ast.FuncDecl) bool {
	if fn.Recv == nil || len(fn.Recv.List) != 1 {
		return false
	}
	star, ok := fn.Recv.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	ident, ok := star.X.(*ast.Ident)
	return ok && ident.Name == "Client"
}

// docSummary returns the first line of fn's doc block with the conventional
// "MethodName " prefix removed, truncated so the table stays readable.
func docSummary(fn *ast.FuncDecl) string {
	if fn.Doc == nil || len(fn.Doc.List) == 0 {
		return ""
	}
	// The first line of the contiguous block, not the last: multi-line docs
	// end mid-sentence and would read as nonsense in a table cell.
	line := strings.TrimSpace(strings.TrimPrefix(fn.Doc.List[0].Text, "//"))
	line = strings.TrimSpace(strings.TrimPrefix(line, fn.Name.Name))
	line = strings.TrimPrefix(line, "-")
	line = strings.TrimSpace(line)
	line = strings.ReplaceAll(line, "|", `\|`)

	const maxLen = 90
	if len(line) > maxLen {
		line = strings.TrimSpace(line[:maxLen]) + "..."
	}
	return line
}

// collectVersionConsts maps "version1_13_0" to "1.13.0" from declarations of
// the form `version1_13_0 = version.Must(version.NewVersion("1.13.0"))`.
func collectVersionConsts(f *ast.File, out map[string]string) {
	ast.Inspect(f, func(n ast.Node) bool {
		spec, ok := n.(*ast.ValueSpec)
		if !ok || len(spec.Names) != 1 || len(spec.Values) != 1 {
			return true
		}
		name := spec.Names[0].Name
		if !strings.HasPrefix(name, "version") {
			return true
		}
		var lit string
		ast.Inspect(spec.Values[0], func(n ast.Node) bool {
			if bl, ok := n.(*ast.BasicLit); ok && bl.Kind == token.STRING && lit == "" {
				if s, err := strconv.Unquote(bl.Value); err == nil {
					lit = s
				}
			}
			return true
		})
		if lit != "" {
			out[name] = lit
		}
		return true
	})
}

// guardVersion returns the highest version fn refuses to run below, or "".
//
// Only the unconditional form counts:
//
//	if err := c.checkServerVersionGreaterThanOrEqual(version1_13_0); err != nil {
//		return ..., err
//	}
//
// The SDK also uses the same call to *choose a path* rather than to refuse
// (`if c.check...(v) != nil { <compat path> }` in DeleteMilestoneByName, and
// `err == nil` to opt into a newer field). Those methods do work below the
// version, so counting them as guards would wrongly mark them "n/a" on the
// older legs.
func guardVersion(fn *ast.FuncDecl, versions map[string]string) string {
	best := ""
	ast.Inspect(fn, func(n ast.Node) bool {
		stmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		errName, constName, ok := guardInit(stmt.Init)
		if !ok {
			return true
		}
		if !isErrNotNil(stmt.Cond, errName) || !returnsErr(stmt.Body, errName) {
			return true
		}
		if v, ok := versions[constName]; ok && higher(v, best) {
			best = v
		}
		return true
	})
	return best
}

// guardInit matches `err := c.checkServerVersionGreaterThanOrEqual(constName)`
// (or `err = ...`) and returns the error variable and the version constant.
func guardInit(init ast.Stmt) (errName, constName string, ok bool) {
	assign, ok := init.(*ast.AssignStmt)
	if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return "", "", false
	}
	lhs, ok := assign.Lhs[0].(*ast.Ident)
	if !ok {
		return "", "", false
	}
	call, ok := assign.Rhs[0].(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return "", "", false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "checkServerVersionGreaterThanOrEqual" {
		return "", "", false
	}
	arg, ok := call.Args[0].(*ast.Ident)
	if !ok {
		return "", "", false
	}
	return lhs.Name, arg.Name, true
}

// isErrNotNil matches exactly `err != nil`. A compound condition
// (`err != nil && opt.IsPrivate != nil`) guards a field, not the route.
func isErrNotNil(cond ast.Expr, errName string) bool {
	bin, ok := cond.(*ast.BinaryExpr)
	if !ok || bin.Op != token.NEQ {
		return false
	}
	x, ok := bin.X.(*ast.Ident)
	if !ok || x.Name != errName {
		return false
	}
	y, ok := bin.Y.(*ast.Ident)
	return ok && y.Name == "nil"
}

// returnsErr reports whether the if-body returns the checked error, which is
// what makes the guard a refusal instead of a branch.
func returnsErr(body *ast.BlockStmt, errName string) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, res := range ret.Results {
			if id, ok := res.(*ast.Ident); ok && id.Name == errName {
				found = true
			}
		}
		return true
	})
	return found
}

// higher compares two dotted versions numerically, field by field.
func higher(a, b string) bool {
	if b == "" {
		return true
	}
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		av, bv := field(as, i), field(bs, i)
		if av != bv {
			return av > bv
		}
	}
	return false
}

func field(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimLeft(parts[i], "v"))
	if err != nil {
		return 0
	}
	return n
}

// atLeast reports whether have >= want, both dotted versions. A version that
// cannot be read at all is treated as new enough: the report already proves
// what ran, and inventing a "n/a" from an unparseable string would hide it.
func atLeast(have, want string) bool {
	if want == "" {
		return true
	}
	if have == "" {
		return true
	}
	return !higher(want, have)
}

func fmtGuard(v string) string {
	return fmt.Sprintf("n/a (>= %s)", v)
}
