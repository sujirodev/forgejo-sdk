// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestActionRunUnmarshal verifies that ActionRun correctly unmarshals all fields
// from JSON, including index_in_repo (RunNumber) and is_ref_deleted (IsRefDeleted)
func TestActionRunUnmarshal(t *testing.T) {
	t.Parallel()
	jsonData := `{
		"id": 4827,
		"workflow_id": "test.yml",
		"title": "Test Workflow",
		"status": "success",
		"event": "push",
		"commit_sha": "abc123",
		"prettyref": "main",
		"html_url": "https://example.com/run/4827",
		"trigger_user": null,
		"repository": null,
		"created": "2026-02-28T12:00:00Z",
		"started": "2026-02-28T12:01:00Z",
		"stopped": "2026-02-28T12:05:00Z",
		"updated": "2026-02-28T12:05:00Z",
		"need_approval": false,
		"approved_by": 0,
		"is_fork_pull_request": false,
		"index_in_repo": 561,
		"is_ref_deleted": true
	}`

	var run ActionRun
	err := json.Unmarshal([]byte(jsonData), &run)
	require.NoError(t, err)

	// Verify existing fields unmarshal correctly
	assert.Equal(t, int64(4827), run.ID)
	assert.Equal(t, "test.yml", run.WorkflowID)
	assert.Equal(t, "Test Workflow", run.Title)
	assert.Equal(t, "success", run.Status)
	assert.Equal(t, "push", run.Event)
	assert.Equal(t, "abc123", run.CommitSHA)

	// Verify timestamps
	expectedCreated, _ := time.Parse(time.RFC3339, "2026-02-28T12:00:00Z")
	assert.Equal(t, expectedCreated, run.Created)

	// Verify the missing fields are now populated
	assert.Equal(t, int64(561), run.RunNumber, "RunNumber should be populated from index_in_repo")
	assert.True(t, run.IsRefDeleted, "IsRefDeleted should be populated from is_ref_deleted")
}

// TestUnit_GetActionsRun exercises GetActionsRun against an httptest server
// instead of a real Forgejo instance: the endpoint is only meaningful when
// authenticated with an ephemeral actions-job token
// (ACTIONS_RUNTIME_TOKEN/forgejo.token), which a plain integration test has
// no way to obtain, so the wire behavior (path, auth header, body decoding)
// is what gets pinned here.
func TestUnit_GetActionsRun(t *testing.T) {
	t.Parallel()
	var gotPath, gotAuth string
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ActionRun{
			ID:         42,
			WorkflowID: "test.yml",
			Status:     "running",
		})
	})

	c := newUnitTestClient(t, srv, SetToken("actions-job-token"))

	run, resp, err := c.GetActionsRun()
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, run)
	assert.Equal(t, "/api/v1/actions/run", gotPath)
	assert.Equal(t, "token actions-job-token", gotAuth)
	assert.Equal(t, int64(42), run.ID)
	assert.Equal(t, "test.yml", run.WorkflowID)
	assert.Equal(t, "running", run.Status)
}

// TestUnit_GetActionsRun_Unauthenticated pins the error path: without a
// valid actions-job token, the server rejects the request and the SDK
// surfaces that as an error rather than a zero-value ActionRun.
func TestUnit_GetActionsRun_Unauthenticated(t *testing.T) {
	t.Parallel()
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "unauthorized"})
	})

	c := newUnitTestClient(t, srv)

	_, resp, err := c.GetActionsRun()
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
