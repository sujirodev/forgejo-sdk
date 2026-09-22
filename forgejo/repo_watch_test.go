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

func TestRepoWatch(t *testing.T) {
	log.Printf("== TestRepoWatch ==")
	c := newTestClient()
	rawVersion, _, err := c.ServerVersion()
	require.NoError(t, err)
	assert.NotEmpty(t, rawVersion)

	owner := createTestUser(t, uniqueName(t, "wtowner"), c)
	c.SetSudo(owner.UserName)
	t.Cleanup(func() { c.SetSudo("") })
	repo1, _ := createTestRepo(t, "TestRepoWatch_1", c)
	repo2, _ := createTestRepo(t, "TestRepoWatch_2", c)
	assert.NotEqual(t, repo1, repo2)

	// GetWatchedRepos
	wl, _, err := c.GetWatchedRepos(owner.UserName)
	require.NoError(t, err)
	assert.NotNil(t, wl)
	maxcount := len(wl)

	// GetMyWatchedRepos
	wl, _, err = c.GetMyWatchedRepos()
	require.NoError(t, err)
	assert.Len(t, wl, maxcount)

	// CheckRepoWatch
	isWatching, _, err := c.CheckRepoWatch(repo1.Owner.UserName, repo1.Name)
	require.NoError(t, err)
	assert.True(t, isWatching)

	// UnWatchRepo
	_, err = c.UnWatchRepo(repo1.Owner.UserName, repo1.Name)
	require.NoError(t, err)
	isWatching, _, _ = c.CheckRepoWatch(repo1.Owner.UserName, repo1.Name)
	assert.False(t, isWatching)

	// WatchRepo
	_, err = c.WatchRepo(repo1.Owner.UserName, repo1.Name)
	require.NoError(t, err)
	isWatching, _, _ = c.CheckRepoWatch(repo1.Owner.UserName, repo1.Name)
	assert.True(t, isWatching)

	// ListRepoSubscribers
	subs, _, err := c.ListRepoSubscribers(repo1.Owner.UserName, repo1.Name, ListRepoSubscribersOptions{})
	require.NoError(t, err)
	if assert.Len(t, subs, 1) {
		assert.Equal(t, owner.UserName, subs[0].UserName)
	}
}
