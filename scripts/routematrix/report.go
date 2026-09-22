// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	exceptionsFile = "route-exceptions.json"
	localLabel     = "local"
	todoMaxAge     = 90 * 24 * time.Hour
)

// leg is one test run: one Forgejo version, one route report. It mirrors the
// JSON that forgejo/route_recorder_test.go writes.
type leg struct {
	Label         string               `json:"label"`
	GeneratedAt   string               `json:"generated_at"`
	ServerURL     string               `json:"server_url"`
	ServerVersion string               `json:"server_version"`
	Routes        map[string]routeStat `json:"routes"`
	Unattributed  []unattributed       `json:"unattributed"`
}

type routeStat struct {
	HTTPMethods []string       `json:"http_methods"`
	Statuses    map[string]int `json:"statuses"`
	Calls       int            `json:"calls"`
	OK          bool           `json:"ok"`
}

type unattributed struct {
	HTTPMethod string `json:"http_method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
}

// legOrder fixes the column order of the CI matrix; anything else sorts after
// these, alphabetically, so an added leg does not reshuffle the table.
var legOrder = map[string]int{
	"lts-11":        0,
	"lts-15":        1,
	"latest-stable": 2,
}

// Title names the leg's column: the version that actually answered, which is
// stronger than the label (a leg labelled lts-11 running 15.x would otherwise
// go unnoticed).
func (l leg) Title() string {
	version := shortVersion(l.ServerVersion)
	if version == "" {
		return l.Label
	}

	major := version
	if i := strings.Index(major, "."); i > 0 {
		major = major[:i]
	}

	switch l.Label {
	case "latest-stable":
		return fmt.Sprintf("V%s.x (latest)", major)
	case localLabel:
		return fmt.Sprintf("V%s.x (local)", major)
	default:
		return fmt.Sprintf("V%s.x", major)
	}
}

// Version names the leg in the README summary, where the exact patch level is
// what a reader wants.
func (l leg) Version() string {
	version := shortVersion(l.ServerVersion)
	if version == "" {
		return l.Label
	}
	if strings.HasPrefix(l.Label, "lts-") {
		return version + " (LTS)"
	}
	if l.Label == "latest-stable" {
		return version + " (latest stable)"
	}
	return version
}

// shortVersion drops Forgejo's build suffix: the server reports
// "11.0.16+gitea-1.22.0", and everything here (comparisons included) cares
// only about the leading X.Y.Z.
func shortVersion(v string) string {
	if i := strings.IndexAny(v, "+-"); i > 0 {
		return v[:i]
	}
	return v
}

func loadReports(paths []string) ([]leg, error) {
	var legs []leg
	seen := map[string]string{}

	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("reading report: %w", err)
		}
		var l leg
		if err := json.Unmarshal(b, &l); err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		if l.Label == "" {
			return nil, fmt.Errorf("%s: report has no label", p)
		}
		if prev, dup := seen[l.Label]; dup {
			return nil, fmt.Errorf("%s and %s both carry the label %q", prev, p, l.Label)
		}
		seen[l.Label] = p
		legs = append(legs, l)
	}

	sort.Slice(legs, func(i, j int) bool {
		oi, oki := legOrder[legs[i].Label]
		oj, okj := legOrder[legs[j].Label]
		switch {
		case oki && okj:
			return oi < oj
		case oki:
			return true
		case okj:
			return false
		default:
			return legs[i].Label < legs[j].Label
		}
	})
	return legs, nil
}

// exception is one declared, reviewed reason for a route to have no evidence
// on one or more legs. "No evidence and no exception" is a red build: the
// reason lives in a file the machine reads, not in prose.
type exception struct {
	Route    string   `json:"route"`
	Versions []string `json:"versions"` // leg labels, or ["*"] for all
	Reason   string   `json:"reason"`
	Detail   string   `json:"detail"`
	Issue    string   `json:"issue"`
	Since    string   `json:"since"` // YYYY-MM-DD, required for "todo"
}

type exceptionSet struct {
	Exceptions []exception `json:"exceptions"`
}

// reasons is closed on purpose: a free-text reason is a place to hide.
//
// coversFailure marks the reasons that may also stand in for evidence that
// exists but never succeeded. Without one of those, a route whose only
// recorded calls failed is an error no declaration can silence.
var reasons = map[string]struct {
	needsIssue    bool
	coversFailure bool
}{
	"version-guard":  {},
	"no-http":        {},
	"server-missing": {needsIssue: true, coversFailure: true},
	"needs-config":   {needsIssue: true, coversFailure: true},
	"known-bug":      {needsIssue: true, coversFailure: true},
	"negative-only":  {needsIssue: true, coversFailure: true},
	"todo":           {needsIssue: true},
}

// coversFailure reports whether e may account for a route that was called and
// never answered 2xx.
func (e exception) coversFailure() bool {
	return reasons[e.Reason].coversFailure
}

func loadExceptions(path string) (exceptionSet, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return exceptionSet{}, nil
	}
	if err != nil {
		return exceptionSet{}, err
	}
	var set exceptionSet
	if err := json.Unmarshal(b, &set); err != nil {
		return exceptionSet{}, fmt.Errorf("%s: %w", path, err)
	}
	return set, nil
}

func (s exceptionSet) find(route, label string) (exception, bool) {
	for _, e := range s.Exceptions {
		if e.Route != route {
			continue
		}
		for _, v := range e.Versions {
			if v == "*" || v == label {
				return e, true
			}
		}
	}
	return exception{}, false
}

func (s exceptionSet) problems(knownRoutes map[string]bool, legs []leg) []string {
	var out []string

	knownLabels := map[string]bool{}
	for _, l := range legs {
		knownLabels[l.Label] = true
	}

	for _, e := range s.Exceptions {
		where := fmt.Sprintf("%s: exception for %s", exceptionsFile, e.Route)

		if !knownRoutes[e.Route] {
			out = append(out, where+" names a route that does not exist")
		}
		if len(e.Versions) == 0 {
			out = append(out, where+" has no versions")
		}
		for _, v := range e.Versions {
			if v != "*" && !knownLabels[v] {
				out = append(out, fmt.Sprintf("%s names the unknown leg %q", where, v))
			}
		}

		rule, ok := reasons[e.Reason]
		if !ok {
			out = append(out, fmt.Sprintf("%s uses the reason %q, which is not one of %s", where, e.Reason, reasonList()))
			continue
		}
		if rule.needsIssue && e.Issue == "" {
			out = append(out, fmt.Sprintf("%s has reason %q, which requires an issue link", where, e.Reason))
		}
		if e.Detail == "" {
			out = append(out, where+" has no detail")
		}

		if e.Reason == "todo" {
			age, err := exceptionAge(e.Since)
			switch {
			case err != nil:
				out = append(out, fmt.Sprintf("%s has reason \"todo\" and needs a since date (YYYY-MM-DD): %v", where, err))
			case age > todoMaxAge:
				out = append(out, fmt.Sprintf("%s has been \"todo\" for %d days (limit: %d)", where, int(age.Hours()/24), int(todoMaxAge.Hours()/24)))
			default:
				fmt.Fprintf(os.Stdout, "  todo exception: %s, %d days old\n", e.Route, int(age.Hours()/24))
			}
		}
	}

	return out
}

func exceptionAge(since string) (time.Duration, error) {
	if since == "" {
		return 0, fmt.Errorf("missing")
	}
	t, err := time.Parse("2006-01-02", since)
	if err != nil {
		return 0, err
	}
	return time.Since(t), nil
}

func reasonList() string {
	names := make([]string, 0, len(reasons))
	for r := range reasons {
		names = append(names, r)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
