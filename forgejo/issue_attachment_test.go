// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIssueAttachment exercises the issue attachment CRUD endpoints
func TestIssueAttachment(t *testing.T) {
	log.Println("== TestIssueAttachment ==")

	c := newTestClient()

	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)
	repo, err := createTestRepo(t, "TestIssueAttachmentRepo", c)
	require.NoError(t, err)
	issue, _, err := c.CreateIssue(user.UserName, repo.Name, CreateIssueOption{Title: "issue with attachments", Body: "body"})
	require.NoError(t, err)

	// ListIssueAttachments - starts empty
	attachments, _, err := c.ListIssueAttachments(user.UserName, repo.Name, issue.Index)
	require.NoError(t, err)
	assert.Empty(t, attachments)

	// CreateIssueAttachment
	attachment, _, err := c.CreateIssueAttachment(user.UserName, repo.Name, issue.Index, strings.NewReader("hello world"), "hello.txt")
	require.NoError(t, err)
	assert.Equal(t, "hello.txt", attachment.Name)
	assert.NotZero(t, attachment.ID)

	// ListIssueAttachments - now has one
	attachments, _, err = c.ListIssueAttachments(user.UserName, repo.Name, issue.Index)
	require.NoError(t, err)
	assert.Len(t, attachments, 1)

	// GetIssueAttachment
	got, _, err := c.GetIssueAttachment(user.UserName, repo.Name, issue.Index, attachment.ID)
	require.NoError(t, err)
	assert.Equal(t, attachment.ID, got.ID)
	assert.Equal(t, attachment.Name, got.Name)

	// EditIssueAttachment
	edited, _, err := c.EditIssueAttachment(user.UserName, repo.Name, issue.Index, attachment.ID, EditAttachmentOptions{Name: "renamed.txt"})
	require.NoError(t, err)
	assert.Equal(t, "renamed.txt", edited.Name)

	// DeleteIssueAttachment
	_, err = c.DeleteIssueAttachment(user.UserName, repo.Name, issue.Index, attachment.ID)
	require.NoError(t, err)

	attachments, _, err = c.ListIssueAttachments(user.UserName, repo.Name, issue.Index)
	require.NoError(t, err)
	assert.Empty(t, attachments)

	_, resp, err := c.GetIssueAttachment(user.UserName, repo.Name, issue.Index, attachment.ID)
	require.Error(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, 404, resp.StatusCode)
	}
}
