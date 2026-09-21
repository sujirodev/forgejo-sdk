// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Command routematrix turns the route evidence the test suite recorded into
// the versioned route contract: ROUTES.md (one line per route, one column per
// Forgejo version the suite ran against) and the summary block in README.md.
//
// It is the generator of docs/PLANO-CONTRATO-ROTAS.md, sections 2.3 and 2.4.
// The point of generating both files is that a hand-written table lies by the
// third pull request; here CI regenerates and diffs them, so a stale table is
// a red build instead of misleading documentation.
//
// Usage:
//
//	routematrix [-root DIR] [-check] [report.json...]
//
// With no report arguments it reads forgejo/route-report.json. With -check it
// writes nothing and fails when the files on disk differ from the generated
// output.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	root := flag.String("root", "", "repository root (default: the directory two levels above this program)")
	check := flag.Bool("check", false, "do not write; fail if ROUTES.md or README.md differ from the generated output")
	flag.Parse()

	if err := run(*root, *check, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "routematrix: %v\n", err)
		os.Exit(1)
	}
}

func run(root string, check bool, reportPaths []string) error {
	if root == "" {
		var err error
		if root, err = defaultRoot(); err != nil {
			return err
		}
	}

	routes, undocumented, err := parseRoutes(filepath.Join(root, "forgejo"))
	if err != nil {
		return err
	}
	if len(routes) == 0 {
		return fmt.Errorf("no routes found under %s/forgejo: wrong -root?", root)
	}

	if len(reportPaths) == 0 {
		reportPaths = []string{filepath.Join(root, "forgejo", "route-report.json")}
	}
	legs, err := loadReports(reportPaths)
	if err != nil {
		return err
	}

	exceptions, err := loadExceptions(filepath.Join(root, exceptionsFile))
	if err != nil {
		return err
	}

	m := buildMatrix(routes, legs, exceptions)

	// Write first, then report the problems: a developer fixing a gap wants
	// to read the generated table, not to have the generator refuse to
	// produce it.
	var writeErr error
	if check {
		writeErr = m.verifyFiles(root)
	} else {
		writeErr = m.writeFiles(root)
	}

	m.printSummary(os.Stdout)

	problems := m.problems(routes, undocumented, legs, exceptions)
	if len(problems) > 0 {
		for _, p := range problems {
			fmt.Fprintf(os.Stderr, "routematrix: %s\n", p)
		}
		return fmt.Errorf("%d problem(s); see above", len(problems))
	}
	return writeErr
}

// defaultRoot resolves the repository root from this program's own location,
// so `go run .` from scripts/routematrix needs no arguments.
func defaultRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// scripts/routematrix -> scripts -> root
	root := filepath.Dir(filepath.Dir(wd))
	if _, err := os.Stat(filepath.Join(root, "forgejo", "client.go")); err != nil {
		return "", fmt.Errorf("cannot find the repository root from %s: pass -root", wd)
	}
	return root, nil
}

// cell is one route's state on one leg.
type cell struct {
	Text string // what goes in the table
	Kind kind
}

type kind int

const (
	kindOK kind = iota
	kindGuarded
	kindException
	kindFailed  // called, never answered 2xx
	kindMissing // no evidence, no exception
)

type matrix struct {
	legs   []leg
	routes []route
	cells  map[string][]cell // route name -> one cell per leg, in legs order
}

func buildMatrix(routes []route, legs []leg, exceptions exceptionSet) matrix {
	m := matrix{legs: legs, routes: routes, cells: map[string][]cell{}}

	for _, r := range routes {
		row := make([]cell, 0, len(legs))
		for _, l := range legs {
			row = append(row, decide(r, l, exceptions))
		}
		m.cells[r.Name] = row
	}
	return m
}

// decide applies the rules of PLANO-CONTRATO-ROTAS section 2.3, in order.
func decide(r route, l leg, exceptions exceptionSet) cell {
	// 1. The SDK itself refuses to call this route on this version. Derived
	//    from the source, never declared by hand.
	if r.Guard != "" && !atLeast(shortVersion(l.ServerVersion), r.Guard) {
		return cell{Text: fmtGuard(r.Guard), Kind: kindGuarded}
	}

	stat, called := l.Routes[r.Name]
	if called && stat.OK {
		return cell{Text: "ok", Kind: kindOK}
	}

	e, declared := exceptions.find(r.Name, l.Label)

	if called {
		// 3. Called and never succeeded. A 404 can be the honest answer (an
		//    IsFollowing that means "no"), but by default this is a masked
		//    regression: only a reason that explicitly covers failing
		//    evidence may stand in for it, and the cell still says so
		//    instead of claiming "ok".
		if declared && e.coversFailure() {
			return cell{Text: e.Reason, Kind: kindException}
		}
		return cell{Text: "fail", Kind: kindFailed}
	}

	if declared {
		return cell{Text: e.Reason, Kind: kindException}
	}
	return cell{Text: "no test", Kind: kindMissing}
}

// problems collects everything that must turn the build red.
func (m matrix) problems(routes, undocumented []route, legs []leg, exceptions exceptionSet) []string {
	var out []string

	for _, r := range undocumented {
		out = append(out, fmt.Sprintf("%s:%d: %s has no doc comment; the Description column would be empty", r.File, r.Line, r.Name))
	}

	for _, l := range legs {
		for _, u := range l.Unattributed {
			out = append(out, fmt.Sprintf("leg %s: %s %s (status %d) could not be attributed to any *Client method; a client built outside newTestClientOpts?", l.Label, u.HTTPMethod, u.Path, u.Status))
		}
	}

	known := map[string]bool{}
	for _, r := range routes {
		known[r.Name] = true
	}
	out = append(out, exceptions.problems(known, legs)...)

	for _, r := range m.routes {
		for i, c := range m.cells[r.Name] {
			switch c.Kind {
			case kindFailed:
				stat := legs[i].Routes[r.Name]
				out = append(out, fmt.Sprintf("leg %s: %s was called %d time(s) and never answered 2xx (statuses: %s)", legs[i].Label, r.Name, stat.Calls, statusList(stat)))
			case kindMissing:
				out = append(out, fmt.Sprintf("leg %s: %s has no evidence and no declared exception in %s", legs[i].Label, r.Name, exceptionsFile))
			}
		}
	}

	sort.Strings(out)
	return out
}

func statusList(s routeStat) string {
	codes := make([]string, 0, len(s.Statuses))
	for code, n := range s.Statuses {
		codes = append(codes, fmt.Sprintf("%s x%d", code, n))
	}
	sort.Strings(codes)
	return strings.Join(codes, ", ")
}

// tally counts one leg's cells by kind.
type tally struct {
	ok, guarded, exception, failed, missing int
}

func (m matrix) tally(i int) tally {
	var t tally
	for _, r := range m.routes {
		switch m.cells[r.Name][i].Kind {
		case kindOK:
			t.ok++
		case kindGuarded:
			t.guarded++
		case kindException:
			t.exception++
		case kindFailed:
			t.failed++
		case kindMissing:
			t.missing++
		}
	}
	return t
}

func (m matrix) printSummary(w *os.File) {
	fmt.Fprintf(w, "%d routes, %d leg(s)\n", len(m.routes), len(m.legs))
	for i, l := range m.legs {
		t := m.tally(i)
		fmt.Fprintf(w, "  %-16s ok=%d guarded=%d exception=%d fail=%d no-test=%d\n",
			l.Title(), t.ok, t.guarded, t.exception, t.failed, t.missing)
	}
	if len(m.legs) == 1 && m.legs[0].Label == localLabel {
		fmt.Fprintf(w, "\n  note: this is a single local leg, not the CI matrix.\n"+
			"  Do not commit the files it generates; CI regenerates them from all legs.\n")
	}
}
