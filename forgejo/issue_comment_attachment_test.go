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

// TestIssueCommentAttachment exercises the comment attachment CRUD endpoints
func TestIssueCommentAttachment(t *testing.T) {
	log.Println("== TestIssueCommentAttachment ==")

	c := newTestClient()

	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)
	repo, err := createTestRepo(t, "TestIssueCommentAttachmentRepo", c)
	require.NoError(t, err)
	issue, _, err := c.CreateIssue(user.UserName, repo.Name, CreateIssueOption{Title: "issue with a commented attachment", Body: "body"})
	require.NoError(t, err)
	comment, _, err := c.CreateIssueComment(user.UserName, repo.Name, issue.Index, CreateIssueCommentOption{Body: "see attached"})
	require.NoError(t, err)

	// ListIssueCommentAttachments - starts empty
	attachments, _, err := c.ListIssueCommentAttachments(user.UserName, repo.Name, comment.ID)
	require.NoError(t, err)
	assert.Empty(t, attachments)

	// CreateIssueCommentAttachment
	attachment, _, err := c.CreateIssueCommentAttachment(user.UserName, repo.Name, comment.ID, strings.NewReader("hello world"), "hello.txt")
	require.NoError(t, err)
	assert.Equal(t, "hello.txt", attachment.Name)
	assert.NotZero(t, attachment.ID)

	// ListIssueCommentAttachments - now has one
	attachments, _, err = c.ListIssueCommentAttachments(user.UserName, repo.Name, comment.ID)
	require.NoError(t, err)
	assert.Len(t, attachments, 1)

	// GetIssueCommentAttachment
	got, _, err := c.GetIssueCommentAttachment(user.UserName, repo.Name, comment.ID, attachment.ID)
	require.NoError(t, err)
	assert.Equal(t, attachment.ID, got.ID)
	assert.Equal(t, attachment.Name, got.Name)

	// EditIssueCommentAttachment
	edited, _, err := c.EditIssueCommentAttachment(user.UserName, repo.Name, comment.ID, attachment.ID, EditAttachmentOptions{Name: "renamed.txt"})
	require.NoError(t, err)
	assert.Equal(t, "renamed.txt", edited.Name)

	// DeleteIssueCommentAttachment
	_, err = c.DeleteIssueCommentAttachment(user.UserName, repo.Name, comment.ID, attachment.ID)
	require.NoError(t, err)

	attachments, _, err = c.ListIssueCommentAttachments(user.UserName, repo.Name, comment.ID)
	require.NoError(t, err)
	assert.Empty(t, attachments)

	_, resp, err := c.GetIssueCommentAttachment(user.UserName, repo.Name, comment.ID, attachment.ID)
	require.Error(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, 404, resp.StatusCode)
	}
}
