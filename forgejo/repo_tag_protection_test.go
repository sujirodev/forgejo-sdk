// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoTagProtection(t *testing.T) {
	t.Parallel()
	log.Println("== TestRepoTagProtection ==")
	c := newTestClient()
	repoName := "TagProtection"

	repo, err := createTestRepo(t, repoName, c)
	require.NoError(t, err)
	assert.NotNil(t, repo)

	// ListTagProtections - should be empty
	tpl, _, err := c.ListTagProtections(repo.Owner.UserName, repo.Name, ListTagProtectionsOptions{})
	require.NoError(t, err)
	assert.Empty(t, tpl)

	// CreateTagProtection
	tp, _, err := c.CreateTagProtection(repo.Owner.UserName, repo.Name, CreateTagProtectionOption{
		NamePattern:        "v*",
		WhitelistUsernames: []string{"test01"},
	})
	require.NoError(t, err)
	assert.Equal(t, "v*", tp.NamePattern)
	assert.Equal(t, []string{"test01"}, tp.WhitelistUsernames)

	// ListTagProtections - should have 1 entry
	tpl, _, err = c.ListTagProtections(repo.Owner.UserName, repo.Name, ListTagProtectionsOptions{})
	require.NoError(t, err)
	assert.Len(t, tpl, 1)

	// GetTagProtection
	tp2, _, err := c.GetTagProtection(repo.Owner.UserName, repo.Name, tpl[0].ID)
	require.NoError(t, err)
	assert.Equal(t, tpl[0], tp2)

	// EditTagProtection
	newPattern := "release-*"
	tp3, _, err := c.EditTagProtection(repo.Owner.UserName, repo.Name, tpl[0].ID, EditTagProtectionOption{
		NamePattern: &newPattern,
	})
	require.NoError(t, err)
	assert.Equal(t, "release-*", tp3.NamePattern)
	assert.Equal(t, tpl[0].ID, tp3.ID)
	assert.Equal(t, tpl[0].Created, tp3.Created)

	// DeleteTagProtection
	_, err = c.DeleteTagProtection(repo.Owner.UserName, repo.Name, tpl[0].ID)
	require.NoError(t, err)

	// ListTagProtections - should be empty again
	tpl, _, err = c.ListTagProtections(repo.Owner.UserName, repo.Name, ListTagProtectionsOptions{})
	require.NoError(t, err)
	assert.Empty(t, tpl)
}
