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

func TestAdminHooks(t *testing.T) {
	log.Println("== TestAdminHooks ==")
	c := newTestClient()

	h, resp, err := c.AdminCreateHook(CreateHookOption{
		Type: HookTypeGitea,
		Config: map[string]string{
			"url":          "http://example.com/admin-hook-sdk-test",
			"content_type": "json",
		},
		Active: false,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotZero(t, h.ID)
	defer func() { _, _ = c.AdminDeleteHook(h.ID) }()

	got, _, err := c.AdminGetHook(h.ID)
	require.NoError(t, err)
	assert.Equal(t, h.ID, got.ID)
	assert.Equal(t, "http://example.com/admin-hook-sdk-test", got.Config["url"])

	hooks, _, err := c.AdminListHooks(ListHooksOptions{})
	require.NoError(t, err)
	found := false
	for _, lh := range hooks {
		if lh.ID == h.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "created hook should be present in AdminListHooks")

	active := true
	edited, resp, err := c.AdminEditHook(h.ID, EditHookOption{
		Config: map[string]string{
			"url":          "http://example.com/admin-hook-sdk-test-edited",
			"content_type": "json",
		},
		Active: &active,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, edited.Active)
	assert.Equal(t, "http://example.com/admin-hook-sdk-test-edited", edited.Config["url"])

	resp, err = c.AdminDeleteHook(h.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)

	_, _, err = c.AdminGetHook(h.ID)
	require.Error(t, err)
}
