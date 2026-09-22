// Copyright 2026 The Forgejo Authors. All rights reserved.
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

// TestUnit_CheckRepoFlag_UnexpectedStatus exercises the branch CheckRepoFlag
// takes when the server answers with neither 204 (flag present) nor 404
// (flag absent) -- unreachable from the integration test below, which only
// ever sees those two.
func TestUnit_CheckRepoFlag_UnexpectedStatus(t *testing.T) {
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	c := newUnitTestClient(t, srv)

	has, _, err := c.CheckRepoFlag("o", "r", "flag")
	require.Error(t, err)
	assert.False(t, has)
	assert.Contains(t, err.Error(), "unexpected Status: 500")
}

// TestRepoFlags exercises the full lifecycle of the six repository-flag
// routes. It needs [repository] ENABLE_FLAGS = true on the test instance
// (see TESTING.md, "Server configuration"); without it every one of these
// calls 404s. Closes issue #62.
func TestRepoFlags(t *testing.T) {
	log.Println("== TestRepoFlags ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "RepoFlags", c)
	require.NoError(t, err)

	fl, _, err := c.ListRepoFlags(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
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
