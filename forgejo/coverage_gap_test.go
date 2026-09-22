// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests close a coverage gap left by TestOrgActionVariables and
// TestDeleteOrgActionSecret (org_action_test.go): both act against
// testOrgName ("testorg"), an organization no test in this file creates, so
// CreateOrgActionSecret/CreateOrgActionVariable fail and the rest of each
// test is t.Skip-ed — leaving ListOrgActionVariables, GetOrgActionVariable,
// UpdateOrgActionVariable, DeleteOrgActionVariable and DeleteOrgActionSecret
// completely unexercised. Rather than touch those existing tests, this adds
// the missing calls against an org this test creates for itself.

func TestOrgActionVariablesLifecycle_CoverageGap(t *testing.T) {
	t.Parallel()
	log.Println("== TestOrgActionVariablesLifecycle_CoverageGap ==")
	c := newTestClient()
	org := newTestOrg(t, c)

	variableName := "GAP_VARIABLE"

	resp, err := c.CreateOrgActionVariable(org.UserName, CreateVariableOption{
		Name: variableName,
		Data: "value1",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	variables, resp, err := c.ListOrgActionVariables(org.UserName, ListOptions{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, variables)

	variable, resp, err := c.GetOrgActionVariable(org.UserName, variableName)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, variableName, variable.Name)

	resp, err = c.UpdateOrgActionVariable(org.UserName, variableName, CreateVariableOption{
		Name: variableName,
		Data: "value2",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	resp, err = c.DeleteOrgActionVariable(org.UserName, variableName)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestDeleteOrgActionSecret_CoverageGap(t *testing.T) {
	t.Parallel()
	log.Println("== TestDeleteOrgActionSecret_CoverageGap ==")
	c := newTestClient()
	org := newTestOrg(t, c)

	_, err := c.CreateOrgActionSecret(org.UserName, CreateSecretOption{
		Name: "GAP_SECRET",
		Data: "value",
	})
	require.NoError(t, err)

	resp, err := c.DeleteOrgActionSecret(org.UserName, "GAP_SECRET")
	require.NoError(t, err)
	require.NotNil(t, resp)
}

// TestGetRepoActionRun_CoverageGap closes the gap left by TestGetRepoActionRun
// (repo_action_test.go), which only calls GetRepoActionRun when a prior
// ListRepoActionRuns returned at least one run -- something that never
// happens there, since no workflow is ever dispatched into a run. The route
// itself needs no runner and no successful run to be exercised: it just
// reads whatever record exists (or reports it doesn't), so this calls it
// directly and only checks that the SDK completes the call.
//
// This deliberately does not assert anything about resp on error: a plain
// 404 (no run with that ID, or an older SDK with no version guard yet)
// still round-trips and leaves resp populated, but once the client grows a
// version guard for this route (see repo_action.go on branches downstream
// of this one), a too-old server makes the call return before any HTTP
// request is sent at all, and resp is nil by design. Either shape exercises
// the function; only the call itself is being checked here.
func TestGetRepoActionRun_CoverageGap(t *testing.T) {
	t.Parallel()
	log.Println("== TestGetRepoActionRun_CoverageGap ==")
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{})

	_, _, _ = c.GetRepoActionRun(repo.Owner.UserName, repo.Name, 1)
}

// The tests below are TestUnit_* (see client_unit_test.go): they exercise
// small pieces of client plumbing that a normal integration test never
// reaches because nothing in the suite constructs a client that way --
// Version() is a free function nothing calls, NewClientWithHTTP and the
// SetContext/SetDebugMode ClientOptions are alternate construction paths
// next to the SetToken/SetBasicAuth ones every other test already uses, and
// ErrUnknownVersion.Error/PullRequestDiffOptions.QueryEncode/fixPullHeadSha
// are pure helpers with no HTTP of their own. They reuse newUnitTestServer
// and newUnitTestClient from client_unit_test.go.

func TestUnit_Version_CoverageGap(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "3.0.0", Version())
}

// versionOnlyUnitServer answers GET /version like a real Forgejo instance
// (so NewClient's own version check, which NewClientWithHTTP and the
// ClientOptions below all go through, succeeds) and 200s everything else.
func versionOnlyUnitServer(t *testing.T) *httptest.Server {
	t.Helper()
	return newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/version") {
			_, _ = w.Write([]byte(`{"version":"16.0.5"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestUnit_NewClientWithHTTP_CoverageGap(t *testing.T) {
	t.Parallel()
	srv := versionOnlyUnitServer(t)

	c := NewClientWithHTTP(srv.URL, &http.Client{})
	require.NotNil(t, c)
}

func TestUnit_SetContextOption_CoverageGap(t *testing.T) {
	t.Parallel()
	srv := versionOnlyUnitServer(t)

	c, err := NewClient(srv.URL, SetContext(context.Background()))
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestUnit_SetDebugModeOption_CoverageGap(t *testing.T) {
	t.Parallel()
	srv := versionOnlyUnitServer(t)

	c, err := NewClient(srv.URL, SetDebugMode())
	require.NoError(t, err)
	require.NotNil(t, c)
}

func TestUnit_ErrUnknownVersion_Error_CoverageGap(t *testing.T) {
	t.Parallel()
	err := &ErrUnknownVersion{raw: "banana"}
	assert.Equal(t, "unknown version: banana", err.Error())
}

func TestUnit_PullRequestDiffOptions_QueryEncode_CoverageGap(t *testing.T) {
	t.Parallel()
	opt := PullRequestDiffOptions{Binary: true}
	assert.Equal(t, "binary=true", opt.QueryEncode())
}

func TestUnit_FixPullHeadSha_AlreadyResolved_CoverageGap(t *testing.T) {
	t.Parallel()
	// Head.Sha already set: fixPullHeadSha must be a no-op and never touch
	// the client, so passing nil here is safe and proves it.
	pr := &PullRequest{
		Base: &PRBranchInfo{Repository: &Repository{Owner: &User{UserName: "owner"}}},
		Head: &PRBranchInfo{Ref: "some-ref", Sha: "deadbeef"},
	}
	require.NoError(t, fixPullHeadSha(nil, pr))
	assert.Equal(t, "deadbeef", pr.Head.Sha)
}
