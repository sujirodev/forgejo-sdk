// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIssueComment creat a issue and test comment creation/edit/deletion on it
func TestIssueComment(t *testing.T) {
	log.Println("== TestIssueComment ==")

	c := newTestClient()

	user, _, err := c.GetMyUserInfo()

	require.NoError(t, err)
	repo, err := createTestRepo(t, "TestIssueCommentRepo", c)
	require.NoError(t, err)
	issue1, _, err := c.CreateIssue(user.UserName, repo.Name, CreateIssueOption{Title: "issue1", Body: "body", Closed: false})
	require.NoError(t, err)
	assert.Equal(t, int64(1), issue1.Index)
	issue2, _, err := c.CreateIssue(user.UserName, repo.Name, CreateIssueOption{Title: "issue1", Body: "body", Closed: false})
	assert.Equal(t, int64(2), issue2.Index)
	require.NoError(t, err)
	tUser2 := createTestUser(t, "Commenter2", c)
	tUser3 := createTestUser(t, "Commenter3", c)

	createOne := func(u *User, issue int64, text string) {
		c.sudo = u.UserName
		comment, _, e := c.CreateIssueComment(user.UserName, repo.Name, issue, CreateIssueCommentOption{Body: text})
		c.sudo = ""
		require.NoError(t, e)
		assert.NotEmpty(t, comment)
		assert.Equal(t, text, comment.Body)
		assert.Equal(t, u.ID, comment.Poster.ID)
	}

	// CreateIssue
	createOne(user, 1, "what a nice issue")
	createOne(tUser2, 1, "dont think so")
	createOne(tUser3, 1, "weow weow")
	createOne(user, 1, "spam isn't it?")
	createOne(tUser3, 2, "hehe first commit")
	createOne(tUser2, 2, "second")
	createOne(user, 2, "3")

	_, err = c.AdminDeleteUser(tUser3.UserName)
	require.NoError(t, err)

	// ListRepoIssueComments
	comments, _, err := c.ListRepoIssueComments(user.UserName, repo.Name, ListRepoIssueCommentOptions{})
	require.NoError(t, err)
	assert.Len(t, comments, 7)

	// ListIssueComments
	comments, _, err = c.ListIssueComments(user.UserName, repo.Name, 2, ListIssueCommentOptions{})
	require.NoError(t, err)
	assert.Len(t, comments, 3)

	// GetIssueComment
	comment, _, err := c.GetIssueComment(user.UserName, repo.Name, comments[1].ID)
	require.NoError(t, err)
	assert.Equal(t, comment.Poster.ID, comments[1].Poster.ID)
	assert.Equal(t, comment.Body, comments[1].Body)
	assert.Equal(t, comment.Updated.Unix(), comments[1].Updated.Unix())

	// EditIssueComment
	comment, _, err = c.EditIssueComment(user.UserName, repo.Name, comments[1].ID, EditIssueCommentOption{
		Body: "changed my mind",
	})
	require.NoError(t, err)
	assert.Equal(t, "changed my mind", comment.Body)

	// DeleteIssueComment
	_, err = c.DeleteIssueComment(user.UserName, repo.Name, comments[1].ID)
	require.NoError(t, err)
	_, _, err = c.GetIssueComment(user.UserName, repo.Name, comments[1].ID)
	require.Error(t, err)
}
