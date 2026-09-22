// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListForks(t *testing.T) {
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})

	forks, _, err := c.ListForks(repo.Owner.UserName, repo.Name, ListForksOptions{})
	require.NoError(t, err)
	assert.Empty(t, forks)

	forker := createTestUser(t, uniqueName(t, "forker"), c)
	asUser(t, c, forker.UserName, func() {
		fork, _, err := c.CreateFork(repo.Owner.UserName, repo.Name, CreateForkOption{})
		require.NoError(t, err)
		t.Cleanup(func() { _, _ = c.DeleteRepo(forker.UserName, fork.Name) })
	})

	forks, _, err = c.ListForks(repo.Owner.UserName, repo.Name, ListForksOptions{})
	require.NoError(t, err)
	require.Len(t, forks, 1)
	assert.Equal(t, forker.UserName, forks[0].Owner.UserName)
}
