// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoCollaborator(t *testing.T) {
	log.Println("== TestRepoCollaborator ==")
	c := newTestClient()

	repo, _ := createTestRepo(t, "RepoCollaborators", c)
	createTestUser(t, "ping", c)
	createTestUser(t, "pong", c)
	defer func() {
		_, err := c.AdminDeleteUser("ping")
		require.NoError(t, err)
		_, err = c.AdminDeleteUser("pong")
		require.NoError(t, err)
	}()

	collaborators, _, err := c.ListCollaborators(repo.Owner.UserName, repo.Name, ListCollaboratorsOptions{})
	require.NoError(t, err)
	assert.Empty(t, collaborators)

	mode := AccessModeAdmin
	resp, err := c.AddCollaborator(repo.Owner.UserName, repo.Name, "ping", AddCollaboratorOption{Permission: &mode})
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	permissonPing, resp, err := c.CollaboratorPermission(repo.Owner.UserName, repo.Name, "ping")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, AccessModeAdmin, permissonPing.Permission)
	assert.Equal(t, "ping", permissonPing.User.UserName)

	mode = AccessModeRead
	_, err = c.AddCollaborator(repo.Owner.UserName, repo.Name, "pong", AddCollaboratorOption{Permission: &mode})
	require.NoError(t, err)

	permissonPong, resp, err := c.CollaboratorPermission(repo.Owner.UserName, repo.Name, "pong")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, AccessModeRead, permissonPong.Permission)
	assert.Equal(t, "pong", permissonPong.User.UserName)

	collaborators, _, err = c.ListCollaborators(repo.Owner.UserName, repo.Name, ListCollaboratorsOptions{})
	require.NoError(t, err)
	assert.Len(t, collaborators, 2)
	assert.Equal(t, []string{"ping", "pong"}, userToStringSlice(collaborators))

	reviewers, _, err := c.GetReviewers(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	// Forgejo 16 narrowed GetReviewers to only list collaborators with write
	// access or higher (same criteria as GetAssignees below), so "pong"
	// (added with AccessModeRead) no longer qualifies.
	assert.Len(t, reviewers, 2)
	assert.Equal(t, []string{"ping", "test01"}, userToStringSlice(reviewers))

	assignees, _, err := c.GetAssignees(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.Len(t, assignees, 2)
	assert.Equal(t, []string{"ping", "test01"}, userToStringSlice(assignees))

	resp, err = c.DeleteCollaborator(repo.Owner.UserName, repo.Name, "ping")
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	collaborators, _, err = c.ListCollaborators(repo.Owner.UserName, repo.Name, ListCollaboratorsOptions{})
	require.NoError(t, err)
	assert.Len(t, collaborators, 1)

	permissonNotExists, resp, err := c.CollaboratorPermission(repo.Owner.UserName, repo.Name, "user_that_not_exists")
	require.Error(t, err)
	assert.Equal(t, 404, resp.StatusCode)
	assert.Nil(t, permissonNotExists)
}
