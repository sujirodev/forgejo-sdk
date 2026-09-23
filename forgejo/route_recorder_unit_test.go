// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

// The route recorder is the source of the per-version evidence in ROUTES.md,
// so its attribution has to be proven, not assumed. These tests pin the four
// properties the generator relies on: a route is credited, a composed method
// credits the routes it delegates to, a route that never answered 2xx is not
// "ok", and a request nobody can be credited for lands in the unattributed
// bucket instead of being silently dropped.

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newRecordingClient builds a client against srv whose transport is rec.
func newRecordingClient(t *testing.T, url string, rec *routeRecorder) *Client {
	t.Helper()
	c, err := NewClient(url,
		SetForgejoVersion("16.0.5"),
		SetHTTPClient(&http.Client{Transport: rec}),
	)
	require.NoError(t, err)
	return c
}

func TestUnit_RouteRecorder_CreditsTheRoute(t *testing.T) {
	t.Parallel()
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/repos/u/r/tags", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	})

	rec := newRouteRecorder()
	c := newRecordingClient(t, srv.URL, rec)

	_, _, err := c.ListRepoTags("u", "r", ListRepoTagsOptions{})
	require.NoError(t, err)

	rep := rec.report("unit", srv.URL, "16.0.5")
	require.Contains(t, rep.Routes, "ListRepoTags")
	got := rep.Routes["ListRepoTags"]
	assert.Equal(t, []string{"GET"}, got.HTTPMethods)
	assert.Equal(t, 1, got.Calls)
	assert.Equal(t, map[string]int{"200": 1}, got.Statuses)
	assert.True(t, got.OK)
	assert.Empty(t, rep.Unattributed)
}

// A method that delegates to other exported methods must credit all of them:
// DeleteMilestoneByName on a pre-1.13 server resolves the milestone through
// ListRepoMilestones and then calls DeleteMilestone.
func TestUnit_RouteRecorder_CreditsComposedRoutes(t *testing.T) {
	t.Parallel()
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`[{"id":7,"title":"v1"}]`))
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	rec := newRouteRecorder()
	c, err := NewClient(srv.URL,
		SetForgejoVersion("1.12.0"), // below 1.13.0: takes the compat path
		SetHTTPClient(&http.Client{Transport: rec}),
	)
	require.NoError(t, err)

	_, err = c.DeleteMilestoneByName("u", "r", "v1")
	require.NoError(t, err)

	rep := rec.report("unit", srv.URL, "16.0.5")
	assert.Contains(t, rep.Routes, "DeleteMilestoneByName")
	assert.Contains(t, rep.Routes, "ListRepoMilestones")
	assert.Contains(t, rep.Routes, "DeleteMilestone")
	assert.Empty(t, rep.Unattributed)
}

// A route called 40 times that only ever got a 404 is not covered on that
// version. This is the distinction the static checker cannot make.
func TestUnit_RouteRecorder_FailedCallsAreNotOK(t *testing.T) {
	t.Parallel()
	srv := newUnitTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	rec := newRouteRecorder()
	c := newRecordingClient(t, srv.URL, rec)

	_, _, err := c.ListRepoTags("u", "r", ListRepoTagsOptions{})
	require.Error(t, err)

	got := rec.report("unit", srv.URL, "16.0.5").Routes["ListRepoTags"]
	assert.Equal(t, 1, got.Calls)
	assert.Equal(t, map[string]int{"404": 1}, got.Statuses)
	assert.False(t, got.OK)
}

func TestUnit_RouteRecorder_UnattributedRequests(t *testing.T) {
	t.Parallel()
	srv := newUnitTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rec := newRouteRecorder()
	client := &http.Client{Transport: rec}
	resp, err := client.Get(srv.URL + "/api/v1/version")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	rep := rec.report("unit", srv.URL, "16.0.5")
	assert.Empty(t, rep.Routes)
	require.Len(t, rep.Unattributed, 1)
	assert.Equal(t, "GET", rep.Unattributed[0].HTTPMethod)
	assert.Equal(t, "/api/v1/version", rep.Unattributed[0].Path)
	assert.Equal(t, http.StatusOK, rep.Unattributed[0].Status)
}

func TestUnit_RouteRecorder_ClientMethodName(t *testing.T) {
	t.Parallel()
	const pkg = "codeberg.org/sujirodev/forgejo-sdk/forgejo/v3"

	for _, tc := range []struct {
		fn   string
		want string
		ok   bool
	}{
		{pkg + ".(*Client).ListRepoTags", "ListRepoTags", true},
		{pkg + ".(*Client).ListRepoTags.func1", "ListRepoTags", true},
		{pkg + ".(*Client).checkServerVersionGreaterThanOrEqual", "", false},
		{pkg + ".NewClient", "", false},
		// net/http.(*Client).Do is on the stack of every request; crediting
		// it would invent a route named "Do".
		{"net/http.(*Client).Do", "", false},
		{"example.com/other.(*Client).ListRepoTags", "", false},
	} {
		got, ok := clientMethodName(tc.fn)
		assert.Equal(t, tc.ok, ok, tc.fn)
		assert.Equal(t, tc.want, got, tc.fn)
	}
}
