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

func TestRepoActionRunners(t *testing.T) {
	t.Parallel()
	log.Println("== TestRepoActionRunners ==")
	c := newTestClient()

	// Unique names on purpose: unlike the fixed-name helper the older
	// action tests share, this one owns a runner registration, and a rerun
	// against a surviving instance must not inherit the previous run's.
	user := createTestUser(t, uniqueName(t, "repoRunnerU"), c)
	c.SetSudo(user.UserName)
	defer c.SetSudo("")
	repo, _, err := c.CreateRepo(CreateRepoOption{Name: uniqueName(t, "repoRunner")})
	require.NoError(t, err)
	require.NotNil(t, repo)

	// The repository-scoped runner routes arrived in Forgejo 15.0.0. Below
	// that the SDK's guard refuses the call without contacting the server,
	// which is the documented behavior and worth asserting on its own leg.
	if !serverAtLeast(t, c, "15.0.0") {
		_, _, err := c.ListRepoRunners(repo.Owner.UserName, repo.Name, ListActionRunnersOptions{})
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
		_, _, err = c.RegisterRepoRunner(repo.Owner.UserName, repo.Name, RegisterRunnerOption{Name: "sdk-repo-runner-test"})
		require.Error(t, err)
		_, _, err = c.GetRepoRunner(repo.Owner.UserName, repo.Name, 1)
		require.Error(t, err)
		_, err = c.DeleteRepoRunner(repo.Owner.UserName, repo.Name, 1)
		require.Error(t, err)
		return
	}

	// A fresh repository owns no runner yet, so the list is a reliable
	// before/after anchor for the registration below.
	runners, resp, err := c.ListRepoRunners(repo.Owner.UserName, repo.Name, ListActionRunnersOptions{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Empty(t, runners, "a fresh repository owns no runner")

	reg, resp, err := c.RegisterRepoRunner(repo.Owner.UserName, repo.Name, RegisterRunnerOption{
		Name:        "sdk-repo-runner-test",
		Description: "created by forgejo-sdk repo runner test",
		Ephemeral:   false,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotZero(t, reg.ID)
	require.NotEmpty(t, reg.UUID)
	// The token is the runner's own credential, not a registration token:
	// the daemon authenticates with it directly.
	require.NotEmpty(t, reg.Token)

	runner, _, err := c.GetRepoRunner(repo.Owner.UserName, repo.Name, reg.ID)
	require.NoError(t, err)
	assert.Equal(t, reg.ID, runner.ID)
	assert.Equal(t, reg.UUID, runner.UUID)
	assert.Equal(t, "sdk-repo-runner-test", runner.Name)
	assert.Equal(t, "created by forgejo-sdk repo runner test", runner.Description)
	assert.Equal(t, repo.ID, runner.RepoID, "a repository-scoped runner carries the repo id, not an owner id")
	assert.Zero(t, runner.OwnerID)
	// No daemon ever connects with this token, so the runner stays offline.
	assert.Equal(t, "offline", runner.Status)

	runners, _, err = c.ListRepoRunners(repo.Owner.UserName, repo.Name, ListActionRunnersOptions{})
	require.NoError(t, err)
	found := false
	for _, r := range runners {
		if r.ID == reg.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "the registered runner should be present in ListRepoRunners")

	// Visible: true widens the list to the runners the repository may use
	// without owning them; its own runner is still in there.
	visible, _, err := c.ListRepoRunners(repo.Owner.UserName, repo.Name, ListActionRunnersOptions{Visible: true})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(visible), len(runners))

	resp, err = c.DeleteRepoRunner(repo.Owner.UserName, repo.Name, reg.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)

	_, _, err = c.GetRepoRunner(repo.Owner.UserName, repo.Name, reg.ID)
	require.Error(t, err, "the runner is gone after DeleteRepoRunner")
}
