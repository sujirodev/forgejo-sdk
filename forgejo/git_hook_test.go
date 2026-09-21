// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Needs [security] DISABLE_GIT_HOOKS = false on the test instance (see
// docs/PLANO-COBERTURA-TESTES.md section 4 / scripts/check-test-instance-settings.sh):
// with the default (true), these routes 404.
func TestRepoGitHooks(t *testing.T) {
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})

	hooks, _, err := c.ListRepoGitHooks(repo.Owner.UserName, repo.Name, ListRepoGitHooksOptions{})
	require.NoError(t, err)
	require.Len(t, hooks, 3) // pre-receive, update, post-receive
	for _, h := range hooks {
		assert.False(t, h.IsActive)
	}

	hook, _, err := c.GetRepoGitHook(repo.Owner.UserName, repo.Name, "pre-receive")
	require.NoError(t, err)
	assert.Equal(t, "pre-receive", hook.Name)
	assert.False(t, hook.IsActive)

	_, err = c.EditRepoGitHook(repo.Owner.UserName, repo.Name, "pre-receive", EditGitHookOption{
		Content: "#!/bin/sh\nexit 0\n",
	})
	require.NoError(t, err)

	hook, _, err = c.GetRepoGitHook(repo.Owner.UserName, repo.Name, "pre-receive")
	require.NoError(t, err)
	assert.True(t, hook.IsActive)
	assert.Equal(t, "#!/bin/sh\nexit 0\n", hook.Content)

	_, err = c.DeleteRepoGitHook(repo.Owner.UserName, repo.Name, "pre-receive")
	require.NoError(t, err)

	hook, _, err = c.GetRepoGitHook(repo.Owner.UserName, repo.Name, "pre-receive")
	require.NoError(t, err)
	assert.False(t, hook.IsActive)
}
