// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIssueDeadline exercises setting and removing an issue's deadline
func TestIssueDeadline(t *testing.T) {
	log.Println("== TestIssueDeadline ==")

	c := newTestClient()

	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)
	repo, err := createTestRepo(t, "TestIssueDeadlineRepo", c)
	require.NoError(t, err)
	issue, _, err := c.CreateIssue(user.UserName, repo.Name, CreateIssueOption{Title: "has a deadline"})
	require.NoError(t, err)

	// EditIssueDeadline - set. Per the API doc, only the date is taken into
	// account and the time of day is ignored: the server normalizes it to
	// the end of that day (23:59:59 UTC), not midnight, so only the date
	// portion is compared here.
	deadline := time.Now().Add(24 * time.Hour).UTC().Truncate(24 * time.Hour)
	got, _, err := c.EditIssueDeadline(user.UserName, repo.Name, issue.Index, EditDeadlineOption{Deadline: &deadline})
	require.NoError(t, err)
	if assert.NotNil(t, got.Deadline) {
		assert.Equal(t, deadline.Format("2006-01-02"), got.Deadline.Format("2006-01-02"))
	}

	updated, _, err := c.GetIssue(user.UserName, repo.Name, issue.Index)
	require.NoError(t, err)
	if assert.NotNil(t, updated.Deadline) {
		assert.Equal(t, deadline.Format("2006-01-02"), updated.Deadline.Format("2006-01-02"))
	}

	// EditIssueDeadline - unset. The deadline endpoint itself echoes back a
	// zero-value due_date (not a JSON null) when clearing it, so Deadline
	// comes back non-nil but zero; GetIssue afterwards correctly reports nil.
	got, _, err = c.EditIssueDeadline(user.UserName, repo.Name, issue.Index, EditDeadlineOption{Deadline: nil})
	require.NoError(t, err)
	if assert.NotNil(t, got.Deadline) {
		assert.True(t, got.Deadline.IsZero())
	}

	updated, _, err = c.GetIssue(user.UserName, repo.Name, issue.Index)
	require.NoError(t, err)
	assert.Nil(t, updated.Deadline)
}
