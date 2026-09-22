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

func newHookOption() CreateHookOption {
	return CreateHookOption{
		Type: HookTypeForgejo,
		Config: map[string]string{
			"url":          "https://example.com/hook",
			"content_type": "json",
		},
		Events: []string{"push"},
		Active: false,
	}
}

func TestRepoHooks(t *testing.T) {
	t.Parallel()
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})

	hooks, _, err := c.ListRepoHooks(repo.Owner.UserName, repo.Name, ListHooksOptions{})
	require.NoError(t, err)
	assert.Empty(t, hooks)

	created, _, err := c.CreateRepoHook(repo.Owner.UserName, repo.Name, newHookOption())
	require.NoError(t, err)
	assert.Equal(t, string(HookTypeForgejo), created.Type)
	assert.False(t, created.Active)

	fetched, _, err := c.GetRepoHook(repo.Owner.UserName, repo.Name, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)

	hooks, _, err = c.ListRepoHooks(repo.Owner.UserName, repo.Name, ListHooksOptions{})
	require.NoError(t, err)
	require.Len(t, hooks, 1)

	active := true
	_, err = c.EditRepoHook(repo.Owner.UserName, repo.Name, created.ID, EditHookOption{Active: &active})
	require.NoError(t, err)

	fetched, _, err = c.GetRepoHook(repo.Owner.UserName, repo.Name, created.ID)
	require.NoError(t, err)
	assert.True(t, fetched.Active)

	_, err = c.DeleteRepoHook(repo.Owner.UserName, repo.Name, created.ID)
	require.NoError(t, err)

	hooks, _, err = c.ListRepoHooks(repo.Owner.UserName, repo.Name, ListHooksOptions{})
	require.NoError(t, err)
	assert.Empty(t, hooks)
}

func TestOrgHooks(t *testing.T) {
	t.Parallel()
	c := newTestClient()
	org := newTestOrg(t, c)

	hooks, _, err := c.ListOrgHooks(org.UserName, ListHooksOptions{})
	require.NoError(t, err)
	assert.Empty(t, hooks)

	created, _, err := c.CreateOrgHook(org.UserName, newHookOption())
	require.NoError(t, err)

	fetched, _, err := c.GetOrgHook(org.UserName, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)

	hooks, _, err = c.ListOrgHooks(org.UserName, ListHooksOptions{})
	require.NoError(t, err)
	require.Len(t, hooks, 1)

	active := true
	_, err = c.EditOrgHook(org.UserName, created.ID, EditHookOption{Active: &active})
	require.NoError(t, err)

	fetched, _, err = c.GetOrgHook(org.UserName, created.ID)
	require.NoError(t, err)
	assert.True(t, fetched.Active)

	_, err = c.DeleteOrgHook(org.UserName, created.ID)
	require.NoError(t, err)

	hooks, _, err = c.ListOrgHooks(org.UserName, ListHooksOptions{})
	require.NoError(t, err)
	assert.Empty(t, hooks)
}

func TestMyHooks(t *testing.T) {
	t.Parallel()
	c := newTestClient()
	owner := createTestUser(t, uniqueName(t, "myhooks"), c)

	var created *Hook
	asUser(t, c, owner.UserName, func() {
		var err error
		hooks, _, err := c.ListMyHooks(ListHooksOptions{})
		require.NoError(t, err)
		assert.Empty(t, hooks)

		created, _, err = c.CreateMyHook(newHookOption())
		require.NoError(t, err)
	})

	var fetched *Hook
	asUser(t, c, owner.UserName, func() {
		var err error
		fetched, _, err = c.GetMyHook(created.ID)
		require.NoError(t, err)
	})
	assert.Equal(t, created.ID, fetched.ID)

	asUser(t, c, owner.UserName, func() {
		hooks, _, err := c.ListMyHooks(ListHooksOptions{})
		require.NoError(t, err)
		assert.Len(t, hooks, 1)
	})

	active := true
	asUser(t, c, owner.UserName, func() {
		_, err := c.EditMyHook(created.ID, EditHookOption{Active: &active})
		require.NoError(t, err)
	})

	asUser(t, c, owner.UserName, func() {
		fetched, _, err := c.GetMyHook(created.ID)
		require.NoError(t, err)
		assert.True(t, fetched.Active)
	})

	asUser(t, c, owner.UserName, func() {
		_, err := c.DeleteMyHook(created.ID)
		require.NoError(t, err)

		hooks, _, err := c.ListMyHooks(ListHooksOptions{})
		require.NoError(t, err)
		assert.Empty(t, hooks)
	})
}

func TestTestRepoHook(t *testing.T) {
	t.Parallel()
	log.Println("== TestTestRepoHook ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "TestHook", c)
	require.NoError(t, err)

	h, _, err := c.CreateRepoHook(repo.Owner.UserName, repo.Name, CreateHookOption{
		Type: HookTypeForgejo,
		Config: map[string]string{
			"url":          "http://localhost:1/webhook",
			"content_type": "json",
		},
		Events: []string{"push"},
		Active: false,
	})
	require.NoError(t, err)
	require.NotNil(t, h)
	assert.False(t, h.Active)

	_, err = c.TestRepoHook(repo.Owner.UserName, repo.Name, h.ID, "main")
	require.NoError(t, err)
}
