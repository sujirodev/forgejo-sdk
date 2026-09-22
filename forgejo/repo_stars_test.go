// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoStaring(t *testing.T) {
	t.Parallel()
	log.Println("== TestRepoStaring ==")

	// init user2
	c := newTestClient()

	user1 := createTestUser(t, uniqueName(t, "starowner"), c)
	userA := createTestUser(t, uniqueName(t, "stargazerA"), c)
	userB := createTestUser(t, uniqueName(t, "stargazerB"), c)

	// Only now switch identity: createTestUser needs the admin account.
	c.SetSudo(user1.UserName)
	t.Cleanup(func() { c.SetSudo("") })

	repo, _ := createTestRepo(t, uniqueName(t, "toStar"), c)
	if repo == nil {
		t.Skip()
	}

	is, _, err := c.IsRepoStarring(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.False(t, is)

	repos, _, err := c.GetMyStarredRepos()
	require.NoError(t, err)
	assert.Empty(t, repos)

	resp, err := c.StarRepo(repo.Owner.UserName, repo.Name)
	if err != nil && resp != nil && resp.StatusCode == http.StatusInternalServerError {
		// Known bug: starring 500s on Forgejo 16.0.5 once [federation] is
		// enabled (needed for GetNodeInfo/ActivityPub elsewhere in this
		// suite); see issue #59. Not reproducible on 11.0.16/15.0.9.
		t.Skip("StarRepo 500s on this server with federation enabled (issue #59)")
	}
	require.NoError(t, err)

	// The "yes" answer too: IsRepoStarring maps 404 to false and 204 to
	// true, and only the false branch was ever exercised.
	is, _, err = c.IsRepoStarring(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.True(t, is)

	c.SetSudo(userA.UserName)
	_, err = c.StarRepo(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	c.SetSudo(userB.UserName)
	_, err = c.StarRepo(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)

	users, _, err := c.ListRepoStargazers(repo.Owner.UserName, repo.Name, ListStargazersOptions{})
	require.NoError(t, err)
	assert.Len(t, users, 3)
	assert.Equal(t, user1.UserName, users[0].UserName)

	_, err = c.UnStarRepo(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	_, err = c.UnStarRepo(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)

	// Back to the repo's owner, who is also the first stargazer: the
	// assertions below are about *their* starred list. This used to be
	// SetSudo("") because the owner was the shared admin account.
	c.SetSudo(user1.UserName)

	users, _, err = c.ListRepoStargazers(repo.Owner.UserName, repo.Name, ListStargazersOptions{})
	require.NoError(t, err)
	assert.Len(t, users, 2)

	repos, _, err = c.GetMyStarredRepos()
	require.NoError(t, err)
	assert.Len(t, repos, 1)

	reposNew, _, err := c.GetStarredRepos(user1.UserName)
	require.NoError(t, err)
	assert.Len(t, repos, 1)
	assert.Equal(t, repos, reposNew)
}
