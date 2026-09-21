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

// issueRelationEndpoints abstracts over the issue blocking and issue
// dependency endpoints, which are structurally identical: list the related
// issues, add one via IssueMeta, remove it again.
type issueRelationEndpoints struct {
	list   func(c *Client, owner, repo string, index int64, opt ListOptions) ([]*Issue, *Response, error)
	add    func(c *Client, owner, repo string, index int64, meta IssueMeta) (*Issue, *Response, error)
	remove func(c *Client, owner, repo string, index int64, meta IssueMeta) (*Issue, *Response, error)
}

// testIssueRelation exercises list/add/remove for the given issue relation endpoints
func testIssueRelation(t *testing.T, repoName string, ep issueRelationEndpoints) {
	c := newTestClient()

	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)
	repo, err := createTestRepo(t, repoName, c)
	require.NoError(t, err)
	from, _, err := c.CreateIssue(user.UserName, repo.Name, CreateIssueOption{Title: "from"})
	require.NoError(t, err)
	to, _, err := c.CreateIssue(user.UserName, repo.Name, CreateIssueOption{Title: "to"})
	require.NoError(t, err)

	// starts empty
	related, _, err := ep.list(c, user.UserName, repo.Name, from.Index, ListOptions{})
	require.NoError(t, err)
	assert.Empty(t, related)

	// add
	_, _, err = ep.add(c, user.UserName, repo.Name, from.Index, IssueMeta{
		Owner: user.UserName,
		Name:  repo.Name,
		Index: to.Index,
	})
	require.NoError(t, err)

	related, _, err = ep.list(c, user.UserName, repo.Name, from.Index, ListOptions{})
	require.NoError(t, err)
	if assert.Len(t, related, 1) {
		assert.Equal(t, to.Index, related[0].Index)
	}

	// remove
	_, _, err = ep.remove(c, user.UserName, repo.Name, from.Index, IssueMeta{
		Owner: user.UserName,
		Name:  repo.Name,
		Index: to.Index,
	})
	require.NoError(t, err)

	related, _, err = ep.list(c, user.UserName, repo.Name, from.Index, ListOptions{})
	require.NoError(t, err)
	assert.Empty(t, related)
}

// TestIssueBlocking exercises the issue blocking relation endpoints
func TestIssueBlocking(t *testing.T) {
	log.Println("== TestIssueBlocking ==")
	testIssueRelation(t, "TestIssueBlockingRepo", issueRelationEndpoints{
		list:   (*Client).ListIssueBlocks,
		add:    (*Client).CreateIssueBlocking,
		remove: (*Client).RemoveIssueBlocking,
	})
}

// TestIssueDependency exercises the issue dependency endpoints
func TestIssueDependency(t *testing.T) {
	log.Println("== TestIssueDependency ==")
	testIssueRelation(t, "TestIssueDependencyRepo", issueRelationEndpoints{
		list:   (*Client).ListIssueDependencies,
		add:    (*Client).CreateIssueDependency,
		remove: (*Client).RemoveIssueDependency,
	})
}
