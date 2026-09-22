// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The generator decides what the published table says about every route, so
// the rules it applies are worth pinning: a wrong "ok" is a promise the SDK
// cannot keep, and a wrong "n/a" hides a route nobody tests.

func writeGoFile(t *testing.T, dir, name, src string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
}

const versionConsts = `package forgejo

var (
	version1_13_0 = version.Must(version.NewVersion("1.13.0"))
	version1_22_0 = version.Must(version.NewVersion("1.22.0"))
)
`

func TestParseRoutes_GuardVsCompatPath(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, "version.go", versionConsts)
	writeGoFile(t, dir, "routes.go", `package forgejo

// Guarded refuses to run on an old server.
func (c *Client) Guarded() error {
	if err := c.checkServerVersionGreaterThanOrEqual(version1_22_0); err != nil {
		return err
	}
	return nil
}

// Compat falls back instead of refusing, so it is not guarded.
func (c *Client) Compat() error {
	if c.checkServerVersionGreaterThanOrEqual(version1_13_0) != nil {
		return c.Guarded()
	}
	return nil
}

// OptIn only uses the newer field when the server has it.
func (c *Client) OptIn() error {
	if err := c.checkServerVersionGreaterThanOrEqual(version1_13_0); err == nil {
		return nil
	}
	return nil
}

// FieldGuard guards a field, not the route.
func (c *Client) FieldGuard(private *bool) error {
	if err := c.checkServerVersionGreaterThanOrEqual(version1_13_0); err != nil && private != nil {
		return err
	}
	return nil
}

func (c *Client) unexported() error { return nil }
`)

	routes, undocumented, err := parseRoutes(dir)
	if err != nil {
		t.Fatal(err)
	}

	guards := map[string]string{}
	for _, r := range routes {
		guards[r.Name] = r.Guard
	}

	for name, want := range map[string]string{
		"Guarded":    "1.22.0",
		"Compat":     "",
		"OptIn":      "",
		"FieldGuard": "",
	} {
		if got := guards[name]; got != want {
			t.Errorf("%s: guard = %q, want %q", name, got, want)
		}
	}
	if _, ok := guards["unexported"]; ok {
		t.Error("unexported methods are not routes")
	}
	if len(undocumented) != 0 {
		t.Errorf("undocumented = %v, want none", undocumented)
	}
}

func TestParseRoutes_UndocumentedAndDescription(t *testing.T) {
	dir := t.TempDir()
	writeGoFile(t, dir, "routes.go", `package forgejo

// ListRepoTags list all the tags of one repository.
// It is the caller's responsibility to paginate.
func (c *Client) ListRepoTags() error { return nil }

func (c *Client) Bare() error { return nil }
`)

	routes, undocumented, err := parseRoutes(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(routes) != 2 {
		t.Fatalf("routes = %d, want 2", len(routes))
	}
	// The first line of the block, not the last: the last is a continuation.
	if want := "list all the tags of one repository."; routes[1].Doc != want {
		t.Errorf("Doc = %q, want %q", routes[1].Doc, want)
	}
	if len(undocumented) != 1 || undocumented[0].Name != "Bare" {
		t.Errorf("undocumented = %v, want [Bare]", undocumented)
	}
}

func TestDecide(t *testing.T) {
	old := leg{Label: "lts-11", ServerVersion: "11.0.16", Routes: map[string]routeStat{}}
	guarded := route{Name: "New", Guard: "1.22.0"}

	// A guard the leg satisfies is not an excuse for missing evidence.
	if got := decide(guarded, leg{Label: "lts-15", ServerVersion: "15.0.9", Routes: map[string]routeStat{}}, exceptionSet{}); got.Kind != kindMissing {
		t.Errorf("satisfied guard, no evidence: kind = %v, want missing", got.Kind)
	}

	// A guard above the leg's version resolves itself, with no hand-written
	// exception. The leg's build suffix must not confuse the comparison.
	withSuffix := leg{Label: "lts-11", ServerVersion: "11.0.16+gitea-1.22.0", Routes: map[string]routeStat{}}
	if got := decide(route{Name: "New", Guard: "16.0.0"}, withSuffix, exceptionSet{}); got.Kind != kindGuarded || got.Text != "n/a (>= 16.0.0)" {
		t.Errorf("unsatisfied guard: %+v, want n/a", got)
	}
	if got := decide(route{Name: "New", Guard: "1.22.0"}, withSuffix, exceptionSet{}); got.Kind == kindGuarded {
		t.Error("11.0.16 satisfies a 1.22.0 guard; the build suffix must not break the comparison")
	}

	withCalls := func(ok bool) leg {
		return leg{Label: "lts-11", ServerVersion: "11.0.16", Routes: map[string]routeStat{
			"R": {Calls: 3, OK: ok, Statuses: map[string]int{"404": 3}},
		}}
	}
	if got := decide(route{Name: "R"}, withCalls(true), exceptionSet{}); got.Kind != kindOK {
		t.Errorf("successful call: kind = %v, want ok", got.Kind)
	}
	// Called 3 times, never a 2xx: that is a masked regression, not coverage.
	if got := decide(route{Name: "R"}, withCalls(false), exceptionSet{}); got.Kind != kindFailed {
		t.Errorf("failed calls: kind = %v, want fail", got.Kind)
	}

	set := exceptionSet{Exceptions: []exception{
		{Route: "R", Versions: []string{"lts-11"}, Reason: "known-bug"},
	}}
	if got := decide(route{Name: "R"}, old, set); got.Kind != kindException || got.Text != "known-bug" {
		t.Errorf("declared exception: %+v", got)
	}
	// A reason that does not cover failing evidence never silences a call
	// that actually failed.
	todo := exceptionSet{Exceptions: []exception{
		{Route: "R", Versions: []string{"lts-11"}, Reason: "todo"},
	}}
	if got := decide(route{Name: "R"}, withCalls(false), todo); got.Kind != kindFailed {
		t.Errorf("a \"todo\" must not mask a failing call: kind = %v", got.Kind)
	}
	// One that does covers it, and the cell says the reason -- never "ok".
	negative := exceptionSet{Exceptions: []exception{
		{Route: "R", Versions: []string{"lts-11"}, Reason: "negative-only"},
	}}
	if got := decide(route{Name: "R"}, withCalls(false), negative); got.Kind != kindException || got.Text != "negative-only" {
		t.Errorf("negative-only should cover failing evidence: %+v", got)
	}
}

func TestExceptionProblems(t *testing.T) {
	legs := []leg{{Label: "lts-11"}}
	known := map[string]bool{"R": true}
	stale := time.Now().Add(-100 * 24 * time.Hour).Format("2006-01-02")

	for name, tc := range map[string]struct {
		e    exception
		want string
	}{
		"unknown route":   {exception{Route: "Nope", Versions: []string{"*"}, Reason: "known-bug", Detail: "d", Issue: "i"}, "does not exist"},
		"unknown leg":     {exception{Route: "R", Versions: []string{"lts-99"}, Reason: "known-bug", Detail: "d", Issue: "i"}, "unknown leg"},
		"free-text":       {exception{Route: "R", Versions: []string{"*"}, Reason: "because", Detail: "d"}, "not one of"},
		"issue required":  {exception{Route: "R", Versions: []string{"*"}, Reason: "known-bug", Detail: "d"}, "requires an issue"},
		"detail required": {exception{Route: "R", Versions: []string{"*"}, Reason: "known-bug", Issue: "i"}, "no detail"},
		"todo undated":    {exception{Route: "R", Versions: []string{"*"}, Reason: "todo", Detail: "d", Issue: "i"}, "since date"},
		"todo too old":    {exception{Route: "R", Versions: []string{"*"}, Reason: "todo", Detail: "d", Issue: "i", Since: stale}, "days"},
	} {
		got := exceptionSet{Exceptions: []exception{tc.e}}.problems(known, legs)
		if len(got) == 0 {
			t.Errorf("%s: no problem reported", name)
			continue
		}
		if !strings.Contains(strings.Join(got, "\n"), tc.want) {
			t.Errorf("%s: problems = %v, want one containing %q", name, got, tc.want)
		}
	}

	fresh := exception{Route: "R", Versions: []string{"lts-11"}, Reason: "todo", Detail: "d", Issue: "i", Since: time.Now().Format("2006-01-02")}
	if got := (exceptionSet{Exceptions: []exception{fresh}}).problems(known, legs); len(got) != 0 {
		t.Errorf("a fresh todo is allowed, got %v", got)
	}
}

func TestReplaceReadmeBlock(t *testing.T) {
	readme := []byte("# Title\n\n" + markerStart + "\nold\n" + markerEnd + "\n\n## Next\n")
	got, err := replaceReadmeBlock(readme, markerStart+"\nnew\n"+markerEnd)
	if err != nil {
		t.Fatal(err)
	}
	want := "# Title\n\n" + markerStart + "\nnew\n" + markerEnd + "\n\n## Next\n"
	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}

	if _, err := replaceReadmeBlock([]byte("no markers here"), "x"); err == nil {
		t.Error("a README without markers must be an error, not a silent no-op")
	}
}

func TestLegTitleDropsBuildSuffix(t *testing.T) {
	for _, tc := range []struct {
		l    leg
		want string
	}{
		{leg{Label: "lts-11", ServerVersion: "11.0.16+gitea-1.22.0"}, "V11.x"},
		{leg{Label: "latest-stable", ServerVersion: "16.0.5"}, "V16.x (latest)"},
		{leg{Label: localLabel, ServerVersion: "16.0.5"}, "V16.x (local)"},
		{leg{Label: "lts-15", ServerVersion: ""}, "lts-15"},
	} {
		if got := tc.l.Title(); got != tc.want {
			t.Errorf("Title() = %q, want %q", got, tc.want)
		}
	}
}

func TestLoadReportsOrdersLegsAndRejectsDuplicates(t *testing.T) {
	dir := t.TempDir()
	write := func(name, label string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(`{"label":"`+label+`","routes":{}}`), 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}

	legs, err := loadReports([]string{write("c.json", "latest-stable"), write("a.json", "lts-11"), write("b.json", "lts-15")})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"lts-11", "lts-15", "latest-stable"}
	for i, l := range legs {
		if l.Label != want[i] {
			t.Errorf("leg %d = %q, want %q", i, l.Label, want[i])
		}
	}

	if _, err := loadReports([]string{write("d.json", "lts-11"), write("e.json", "lts-11")}); err == nil {
		t.Error("two reports with the same label must be an error")
	}
}
