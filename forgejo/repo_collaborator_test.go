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
	// Forgejo 15.0.8 narrowed GetReviewers to only list collaborators with
	// write access or higher (same criteria as GetAssignees below), so
	// "pong" (added with AccessModeRead) no longer qualifies. Confirmed by
	// bisecting the release line directly against the live endpoint: 15.0.7
	// still returns pong, 15.0.8 does not (16.0.5 inherits the new
	// behavior). This is the server's own access-control decision, not
	// something GetReviewers computes or could normalize away, so the test
	// asserts per version instead of picking one and breaking the other —
	// found by running this suite against Forgejo 11.0.16 and 15.0.9, not
	// just the newest release.
	if serverAtLeast(t, c, "15.0.8") {
		assert.Len(t, reviewers, 2)
		assert.Equal(t, []string{"ping", "test01"}, userToStringSlice(reviewers))
	} else {
		assert.Len(t, reviewers, 3)
		assert.Equal(t, []string{"ping", "pong", "test01"}, userToStringSlice(reviewers))
	}

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

func TestIsCollaborator(t *testing.T) {
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})
	collaborator := createTestUser(t, uniqueName(t, "collab"), c)

	is, _, err := c.IsCollaborator(repo.Owner.UserName, repo.Name, collaborator.UserName)
	require.NoError(t, err)
	assert.False(t, is)

	mode := AccessModeRead
	_, err = c.AddCollaborator(repo.Owner.UserName, repo.Name, collaborator.UserName, AddCollaboratorOption{Permission: &mode})
	require.NoError(t, err)

	is, _, err = c.IsCollaborator(repo.Owner.UserName, repo.Name, collaborator.UserName)
	require.NoError(t, err)
	assert.True(t, is)
}
