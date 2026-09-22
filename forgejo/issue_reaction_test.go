// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueReactions(t *testing.T) {
	t.Parallel()
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})
	issue := createTestIssue(t, c, repo.Name, "Reaction target", "", nil, nil, 0, nil, false, false)
	comment, _, err := c.CreateIssueComment(repo.Owner.UserName, repo.Name, issue.Index, CreateIssueCommentOption{Body: "a comment"})
	require.NoError(t, err)

	reactions, _, err := c.GetIssueReactions(repo.Owner.UserName, repo.Name, issue.Index)
	require.NoError(t, err)
	assert.Empty(t, reactions)

	added, _, err := c.PostIssueReaction(repo.Owner.UserName, repo.Name, issue.Index, "+1")
	require.NoError(t, err)
	assert.Equal(t, "+1", added.Reaction)
	assert.Equal(t, "test01", added.User.UserName)

	reactions, _, err = c.GetIssueReactions(repo.Owner.UserName, repo.Name, issue.Index)
	require.NoError(t, err)
	assert.Len(t, reactions, 1)
	assert.Equal(t, "+1", reactions[0].Reaction)

	_, err = c.DeleteIssueReaction(repo.Owner.UserName, repo.Name, issue.Index, "+1")
	require.NoError(t, err)
	reactions, _, err = c.GetIssueReactions(repo.Owner.UserName, repo.Name, issue.Index)
	require.NoError(t, err)
	assert.Empty(t, reactions)

	commentReactions, _, err := c.GetIssueCommentReactions(repo.Owner.UserName, repo.Name, comment.ID)
	require.NoError(t, err)
	assert.Empty(t, commentReactions)

	addedComment, _, err := c.PostIssueCommentReaction(repo.Owner.UserName, repo.Name, comment.ID, "heart")
	require.NoError(t, err)
	assert.Equal(t, "heart", addedComment.Reaction)

	commentReactions, _, err = c.GetIssueCommentReactions(repo.Owner.UserName, repo.Name, comment.ID)
	require.NoError(t, err)
	assert.Len(t, commentReactions, 1)

	_, err = c.DeleteIssueCommentReaction(repo.Owner.UserName, repo.Name, comment.ID, "heart")
	require.NoError(t, err)
	commentReactions, _, err = c.GetIssueCommentReactions(repo.Owner.UserName, repo.Name, comment.ID)
	require.NoError(t, err)
	assert.Empty(t, commentReactions)
}
