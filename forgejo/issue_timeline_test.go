// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIssueTimeline exercises listing the comments and events on an issue
func TestIssueTimeline(t *testing.T) {
	log.Println("== TestIssueTimeline ==")

	c := newTestClient()

	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)
	repo, err := createTestRepo(t, "TestIssueTimelineRepo", c)
	require.NoError(t, err)
	issue, _, err := c.CreateIssue(user.UserName, repo.Name, CreateIssueOption{Title: "has a timeline"})
	require.NoError(t, err)
	_, _, err = c.CreateIssueComment(user.UserName, repo.Name, issue.Index, CreateIssueCommentOption{Body: "a comment"})
	require.NoError(t, err)
	state := StateClosed
	_, _, err = c.EditIssue(user.UserName, repo.Name, issue.Index, EditIssueOption{State: &state})
	require.NoError(t, err)

	timeline, _, err := c.ListIssueTimeline(user.UserName, repo.Name, issue.Index, ListIssueTimelineOptions{})
	require.NoError(t, err)
	// At least the comment and the close event show up on the timeline
	assert.GreaterOrEqual(t, len(timeline), 2)

	var sawComment, sawClose bool
	for _, entry := range timeline {
		switch entry.Type {
		case "comment":
			sawComment = true
			assert.Equal(t, "a comment", entry.Body)
		case "close":
			sawClose = true
		}
	}
	assert.True(t, sawComment, "expected a comment entry in the timeline")
	assert.True(t, sawClose, "expected a close event entry in the timeline")
}
