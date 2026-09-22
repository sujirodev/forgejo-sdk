// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

// A small table-driven runner for "contract of wire" unit tests: for a
// given *Client method, assert the HTTP verb, path, query string and body it
// sends, and (for version-gated methods) that it refuses to even contact
// the server on an old version.
//
// Each phase of the plan adds its own TestUnit_Routes_<area> function in the
// area's own file, built from a []routeCase and run through runRouteCases.
// See TestUnit_Routes_AdminCron below for the pattern.

import (
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// routeCase describes one *Client method's wire contract.
type routeCase struct {
	// name identifies the case (and, by convention, names the method under
	// test) for `go test -run` and for scripts/check-route-coverage.sh
	// --strict, which looks for "<method>(" inside a TestUnit_ function.
	name string

	// handler runs inside the fake server and must assert everything this
	// case cares about (method, URL path, query, headers, body) before
	// writing a response. Call t.Error/t.Fatal via the t passed in, not the
	// outer test's t, so failures are attributed to the right sub-test.
	handler func(t *testing.T, w http.ResponseWriter, r *http.Request)

	// call invokes the *Client method under test against c and returns its
	// error. It must call the fake server exactly once on success. It may
	// use t for extra assertions on the decoded return value.
	call func(t *testing.T, c *Client) error

	// minVersion/belowVersion, when both set, additionally verify the
	// method's version guard: calling it against a client pinned to
	// belowVersion must fail *and* must never reach the handler, and calling
	// it pinned to exactly minVersion must succeed.
	minVersion   string
	belowVersion string
}

// runRouteCases runs each case as its own subtest.
func runRouteCases(t *testing.T, cases []routeCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var hits int32
			srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&hits, 1)
				tc.handler(t, w, r)
			})

			version := tc.minVersion
			if version == "" {
				version = "16.0.5" // an arbitrary version newer than every guard
			}
			// SetForgejoVersion pins the version through a sync.Once, so
			// newUnitTestClient (which pre-applies its own default) can't be
			// reused here: build the client directly with the exact pin.
			c, err := NewClient(srv.URL, SetForgejoVersion(version))
			require.NoError(t, err)

			err = tc.call(t, c)
			require.NoError(t, err, "call against a server on the method's minimum (or default) version")
			assert.Equal(t, int32(1), atomic.LoadInt32(&hits), "handler should be hit exactly once")

			if tc.minVersion == "" {
				return
			}
			if tc.belowVersion == "" {
				t.Fatalf("routeCase %q sets minVersion without belowVersion", tc.name)
			}

			t.Run("blocked_below_min_version", func(t *testing.T) {
				var belowHits int32
				belowSrv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
					atomic.AddInt32(&belowHits, 1)
					w.WriteHeader(http.StatusOK)
				})
				c, err := NewClient(belowSrv.URL, SetForgejoVersion(tc.belowVersion))
				require.NoError(t, err)

				err = tc.call(t, c)
				require.Error(t, err, "must refuse to call an endpoint the pinned version doesn't have")
				assert.Equal(t, int32(0), atomic.LoadInt32(&belowHits), "must not contact the server when the version guard blocks the call")
			})
		})
	}
}

// TestUnit_Routes_AdminCron doubles as the runner's own worked example.
func TestUnit_Routes_AdminCron(t *testing.T) {
	t.Parallel()
	runRouteCases(t, []routeCase{
		{
			name:         "ListCronTasks",
			minVersion:   "1.13.0",
			belowVersion: "1.12.9",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/v1/admin/cron", r.URL.Path)
				assert.Equal(t, "limit=0&page=1", r.URL.RawQuery)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[{"name":"delete_old_actions","schedule":"@every 24h"}]`))
			},
			call: func(t *testing.T, c *Client) error {
				tasks, _, err := c.ListCronTasks(ListCronTaskOptions{})
				if err != nil {
					return err
				}
				assert.Len(t, tasks, 1)
				return nil
			},
		},
		{
			name:         "RunCronTasks",
			minVersion:   "1.13.0",
			belowVersion: "1.12.9",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				// r.URL.Path is the decoded path; use r.URL.EscapedPath() to
				// assert on the raw escaping a case sends over the wire.
				assert.Equal(t, "/api/v1/admin/cron/delete old actions", r.URL.Path)
				assert.Equal(t, "/api/v1/admin/cron/delete%20old%20actions", r.URL.EscapedPath())
				w.WriteHeader(http.StatusNoContent)
			},
			call: func(t *testing.T, c *Client) error {
				_, err := c.RunCronTasks("delete old actions")
				return err
			},
		},
	})
}

// TestUnit_Routes_RepoActionRuns pins the guard added for Forgejo 12.0.0:
// /repos/{owner}/{repo}/actions/runs does not exist on 11.x, where the router
// answers a bare 404 page instead of an API error.
func TestUnit_Routes_RepoActionRuns(t *testing.T) {
	t.Parallel()
	runRouteCases(t, []routeCase{
		{
			name:         "ListRepoActionRuns",
			minVersion:   "12.0.0",
			belowVersion: "11.0.16",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/v1/repos/u/r/actions/runs", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"total_count":1,"workflow_runs":[{"id":7}]}`))
			},
			call: func(t *testing.T, c *Client) error {
				runs, _, err := c.ListRepoActionRuns("u", "r", ListActionRunsOption{})
				if err != nil {
					return err
				}
				assert.Len(t, runs.WorkflowRuns, 1)
				return nil
			},
		},
		{
			name:         "GetRepoActionRun",
			minVersion:   "12.0.0",
			belowVersion: "11.0.16",
			handler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/api/v1/repos/u/r/actions/runs/7", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"id":7}`))
			},
			call: func(t *testing.T, c *Client) error {
				run, _, err := c.GetRepoActionRun("u", "r", 7)
				if err != nil {
					return err
				}
				assert.Equal(t, int64(7), run.ID)
				return nil
			},
		},
	})
}
