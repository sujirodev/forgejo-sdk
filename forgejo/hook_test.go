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

func TestTestRepoHook(t *testing.T) {
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
