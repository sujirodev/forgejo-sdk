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

func TestRepoTeamManagement(t *testing.T) {
	log.Println("== TestRepoTeamManagement ==")
	c := newTestClient()

	// prepare for test
	clean, repo, err := createTestOrgRepo(t, c, "RepoTeamManagement")
	if err != nil {
		return
	}
	defer clean()
	if _, err = createTestOrgTeams(t, c, repo.Owner.UserName, "Admins", AccessModeAdmin, map[string]string{RepoUnitCode.String(): "read", RepoUnitIssues.String(): "read", RepoUnitPulls.String(): "read", RepoUnitReleases.String(): "read"}); err != nil {
		return
	}
	if _, err = createTestOrgTeams(t, c, repo.Owner.UserName, "CodeManager", AccessModeWrite, map[string]string{RepoUnitCode.String(): "read"}); err != nil {
		return
	}
	if _, err = createTestOrgTeams(t, c, repo.Owner.UserName, "IssueManager", AccessModeWrite, map[string]string{RepoUnitIssues.String(): "read", RepoUnitPulls.String(): "read"}); err != nil {
		return
	}

	// test
	teams, _, err := c.GetRepoTeams(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	if !assert.Len(t, teams, 1) {
		return
	}
	assert.Equal(t, AccessModeOwner, teams[0].Permission)

	team, _, err := c.CheckRepoTeam(repo.Owner.UserName, repo.Name, "Admins")
	require.NoError(t, err)
	assert.Nil(t, team)

	resp, err := c.AddRepoTeam(repo.Owner.UserName, repo.Name, "Admins")
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
	resp, err = c.AddRepoTeam(repo.Owner.UserName, repo.Name, "CodeManager")
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
	resp, err = c.AddRepoTeam(repo.Owner.UserName, repo.Name, "IssueManager")
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	team, _, err = c.CheckRepoTeam(repo.Owner.UserName, repo.Name, "Admins")
	require.NoError(t, err)
	if assert.NotNil(t, team) {
		assert.Equal(t, "Admins", team.Name)
		assert.Equal(t, AccessModeAdmin, team.Permission)
	}

	teams, _, err = c.GetRepoTeams(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.Len(t, teams, 4)

	resp, err = c.RemoveRepoTeam(repo.Owner.UserName, repo.Name, "IssueManager")
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	team, _, err = c.CheckRepoTeam(repo.Owner.UserName, repo.Name, "IssueManager")
	require.NoError(t, err)
	assert.Nil(t, team)
}
