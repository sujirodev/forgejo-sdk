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

func TestRepoFlags(t *testing.T) {
	log.Println("== TestRepoFlags ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "RepoFlags", c)
	require.NoError(t, err)

	fl, _, err := c.ListRepoFlags(repo.Owner.UserName, repo.Name)
	if err != nil {
		t.Skipf("repo flags unavailable on this server (see issue #62): %v", err)
	}
	assert.Empty(t, fl)

	has, _, err := c.CheckRepoFlag(repo.Owner.UserName, repo.Name, "takedown")
	require.NoError(t, err)
	assert.False(t, has)

	_, err = c.AddRepoFlag(repo.Owner.UserName, repo.Name, "takedown")
	require.NoError(t, err)

	has, _, err = c.CheckRepoFlag(repo.Owner.UserName, repo.Name, "takedown")
	require.NoError(t, err)
	assert.True(t, has)

	fl, _, err = c.ListRepoFlags(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.Equal(t, []string{"takedown"}, fl)

	_, err = c.ReplaceAllRepoFlags(repo.Owner.UserName, repo.Name, ReplaceFlagsOption{Flags: []string{"a", "b"}})
	require.NoError(t, err)
	fl, _, err = c.ListRepoFlags(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"a", "b"}, fl)

	_, err = c.DeleteRepoFlag(repo.Owner.UserName, repo.Name, "a")
	require.NoError(t, err)
	fl, _, err = c.ListRepoFlags(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.Equal(t, []string{"b"}, fl)

	_, err = c.DeleteAllRepoFlags(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	fl, _, err = c.ListRepoFlags(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.Empty(t, fl)
}
