// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestActionRunUnmarshal verifies that ActionRun correctly unmarshals all fields
// from JSON, including index_in_repo (RunNumber) and is_ref_deleted (IsRefDeleted)
func TestActionRunUnmarshal(t *testing.T) {
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
	assert.Equal(t, true, run.IsRefDeleted, "IsRefDeleted should be populated from is_ref_deleted")
}
