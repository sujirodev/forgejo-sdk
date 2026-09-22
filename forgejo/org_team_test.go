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

func createTestOrgTeams(t *testing.T, c *Client, org, name string, accessMode AccessMode, unitsMap map[string]string) (*Team, error) {
	team, _, e := c.CreateTeam(org, CreateTeamOption{
		Name:                    name,
		Description:             name + "'s team desc",
		Permission:              accessMode,
		CanCreateOrgRepo:        false,
		IncludesAllRepositories: false,
		UnitsMap:                unitsMap,
	})
	require.NoError(t, e)
	assert.NotNil(t, team)
	return team, e
}

func TestDeprecationErrorOnUnitsUse(t *testing.T) {
	t.Parallel()
	log.Println("== TestDeprecationErrorOnUnitsUse ==")

	c := newTestClient()
	org := "TestTeamsOrg"
	_, _, e := c.CreateTeam(org, CreateTeamOption{
		Name:                    "testTeam",
		Description:             "testTeam" + "'s team desc",
		Permission:              AccessModeAdmin,
		CanCreateOrgRepo:        false,
		IncludesAllRepositories: false,
		Units:                   []RepoUnitType{RepoUnitCode, RepoUnitIssues, RepoUnitPulls, RepoUnitReleases},
	})
	require.Error(t, e)
	assert.EqualError(t, e, "variable Units should be replaced by UnitsMap")
}

func TestTeamSearch(t *testing.T) {
	t.Parallel()
	log.Println("== TestTeamSearch ==")
	c := newTestClient()

	orgName := "TestTeamsOrg"
	// prepare for test
	_, _, err := c.CreateOrg(CreateOrgOption{
		Name:                      orgName,
		Visibility:                VisibleTypePublic,
		RepoAdminChangeTeamAccess: true,
	})
	defer func() {
		_, _ = c.DeleteOrg(orgName)
	}()

	require.NoError(t, err)

	if _, err = createTestOrgTeams(t, c, orgName, "Admins", AccessModeAdmin, map[string]string{RepoUnitCode.String(): string(AccessModeRead), RepoUnitIssues.String(): string(AccessModeRead), RepoUnitPulls.String(): string(AccessModeRead), RepoUnitReleases.String(): string(AccessModeRead)}); err != nil {
		return
	}

	teams, _, err := c.SearchOrgTeams(orgName, &SearchTeamsOptions{
		Query: "Admins",
	})
	require.NoError(t, err)
	if assert.Len(t, teams, 1) {
		assert.Equal(t, "Admins", teams[0].Name)
	}
}

func TestCreateTeamWithUnitsMap(t *testing.T) {
	t.Parallel()
	log.Println("== TestCreateTeamWithUnitsMap ==")
	c := newTestClient()

	orgName := "TestUnitsMapOrg"
	// Create test organization
	_, _, err := c.CreateOrg(CreateOrgOption{
		Name:                      orgName,
		Visibility:                VisibleTypePublic,
		RepoAdminChangeTeamAccess: true,
	})
	defer func() {
		_, _ = c.DeleteOrg(orgName)
	}()

	require.NoError(t, err)

	// Create team with per-unit permissions using UnitsMap
	unitsMap := map[string]string{
		RepoUnitCode.String():    string(AccessModeWrite),
		RepoUnitIssues.String():  string(AccessModeWrite),
		RepoUnitPulls.String():   string(AccessModeRead),
		RepoUnitWiki.String():    string(AccessModeRead),
		RepoUnitActions.String(): string(AccessModeNone),
	}

	team, _, err := c.CreateTeam(orgName, CreateTeamOption{
		Name:                    "developers",
		Description:             "Developers team with per-unit permissions",
		Permission:              AccessModeWrite,
		CanCreateOrgRepo:        false,
		IncludesAllRepositories: false,
		UnitsMap:                unitsMap,
	})

	require.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "developers", team.Name)
	assert.Equal(t, unitsMap, team.UnitsMap)
}

func TestEditTeamWithUnitsMap(t *testing.T) {
	t.Parallel()
	log.Println("== TestEditTeamWithUnitsMap ==")
	c := newTestClient()

	orgName := "TestEditUnitsMapOrg"
	// Create test organization
	_, _, err := c.CreateOrg(CreateOrgOption{
		Name:                      orgName,
		Visibility:                VisibleTypePublic,
		RepoAdminChangeTeamAccess: true,
	})
	defer func() {
		_, _ = c.DeleteOrg(orgName)
	}()

	require.NoError(t, err)

	// Create team first with UnitsMap for consistency
	initialTeam, _, err := c.CreateTeam(orgName, CreateTeamOption{
		Name:                    "readers",
		Description:             "Initial readers team",
		Permission:              AccessModeRead,
		CanCreateOrgRepo:        false,
		IncludesAllRepositories: false,
		UnitsMap: map[string]string{
			RepoUnitCode.String(): string(AccessModeRead),
		},
	})

	require.NoError(t, err)
	assert.NotNil(t, initialTeam)

	// Edit team to add per-unit permissions
	updatedUnitsMap := map[string]string{
		RepoUnitCode.String():   string(AccessModeRead),
		RepoUnitIssues.String(): string(AccessModeWrite),
		RepoUnitPulls.String():  string(AccessModeNone),
	}

	_, err = c.EditTeam(initialTeam.ID, EditTeamOption{
		Name:       "readers",
		Permission: AccessModeRead,
		UnitsMap:   updatedUnitsMap,
	})

	require.NoError(t, err)

	// Verify the team was updated
	updatedTeam, _, err := c.GetTeam(initialTeam.ID)
	require.NoError(t, err)
	assert.NotNil(t, updatedTeam)
	assert.Equal(t, updatedUnitsMap, updatedTeam.UnitsMap)
}

func TestUnitsMapSerialization(t *testing.T) {
	t.Parallel()
	log.Println("== TestUnitsMapSerialization ==")

	// Test CreateTeamOption serialization
	opt := CreateTeamOption{
		Name:        "test-team",
		Description: "Test team",
		Permission:  AccessModeWrite,
		UnitsMap: map[string]string{
			"repo.code":   "write",
			"repo.issues": "read",
		},
	}

	// Verify UnitsMap field is set correctly
	assert.NotNil(t, opt.UnitsMap)
	assert.Len(t, opt.UnitsMap, 2)
	assert.Equal(t, "write", opt.UnitsMap["repo.code"])
	assert.Equal(t, "read", opt.UnitsMap["repo.issues"])

	// Test EditTeamOption serialization
	editOpt := EditTeamOption{
		Name: "test-team",
		UnitsMap: map[string]string{
			RepoUnitCode.String(): string(AccessModeNone),
		},
	}

	assert.NotNil(t, editOpt.UnitsMap)
	assert.Len(t, editOpt.UnitsMap, 1)
	assert.Equal(t, "none", editOpt.UnitsMap["repo.code"])

	// Test Team struct
	team := &Team{
		ID:   1,
		Name: "test-team",
		UnitsMap: map[string]string{
			RepoUnitWiki.String():     string(AccessModeRead),
			RepoUnitReleases.String(): string(AccessModeWrite),
		},
	}

	assert.NotNil(t, team.UnitsMap)
	assert.Len(t, team.UnitsMap, 2)
	assert.Equal(t, "read", team.UnitsMap["repo.wiki"])
	assert.Equal(t, "write", team.UnitsMap["repo.releases"])
}

func TestUnitsMapBackwardCompatibility(t *testing.T) {
	t.Parallel()
	log.Println("== TestUnitsMapBackwardCompatibility ==")

	// Test that teams without UnitsMap still work
	opt := CreateTeamOption{
		Name:                    "legacy-team",
		Permission:              AccessModeRead,
		CanCreateOrgRepo:        true,
		IncludesAllRepositories: true,
		Units:                   []RepoUnitType{RepoUnitCode, RepoUnitIssues},
	}

	// UnitsMap should be nil (not set)
	assert.Nil(t, opt.UnitsMap)

	// Verify other fields are still accessible
	assert.Equal(t, "legacy-team", opt.Name)
	assert.Equal(t, AccessModeRead, opt.Permission)
	assert.True(t, opt.CanCreateOrgRepo)
	assert.True(t, opt.IncludesAllRepositories)
	assert.Len(t, opt.Units, 2)

	// Test EditTeamOption without UnitsMap
	editOpt := EditTeamOption{
		Name:  "legacy-team",
		Units: []RepoUnitType{RepoUnitCode},
	}

	assert.Nil(t, editOpt.UnitsMap)
	assert.Len(t, editOpt.Units, 1)
}

func TestOrgTeams_ListsMembersAndRepos(t *testing.T) {
	t.Parallel()
	c := newTestClient()
	org := newTestOrg(t, c)

	team, err := createTestOrgTeams(t, c, org.UserName, uniqueName(t, "team"), AccessModeWrite,
		map[string]string{RepoUnitCode.String(): string(AccessModeWrite)})
	require.NoError(t, err)

	// ListOrgTeams / ListMyTeams: an org always has a default "Owners" team,
	// so assert our team is present rather than an exact count.
	orgTeams, _, err := c.ListOrgTeams(org.UserName, ListTeamsOptions{})
	require.NoError(t, err)
	assert.True(t, containsTeamName(orgTeams, team.Name))

	// The creator isn't automatically added to a newly created custom team
	// (only to the org's default "Owners" team), so just check the call
	// itself works and returns something.
	myTeams, _, err := c.ListMyTeams(&ListTeamsOptions{})
	require.NoError(t, err)
	assert.NotEmpty(t, myTeams)

	// AddTeamMember / GetTeamMember / ListTeamMembers / RemoveTeamMember
	mate := createTestUser(t, uniqueName(t, "teammate"), c)

	_, err = c.AddTeamMember(team.ID, mate.UserName)
	require.NoError(t, err)

	member, _, err := c.GetTeamMember(team.ID, mate.UserName)
	require.NoError(t, err)
	assert.Equal(t, mate.UserName, member.UserName)

	members, _, err := c.ListTeamMembers(team.ID, ListTeamMembersOptions{})
	require.NoError(t, err)
	assert.True(t, containsUserName(members, mate.UserName))

	_, err = c.RemoveTeamMember(team.ID, mate.UserName)
	require.NoError(t, err)

	members, _, err = c.ListTeamMembers(team.ID, ListTeamMembersOptions{})
	require.NoError(t, err)
	assert.False(t, containsUserName(members, mate.UserName))

	// AddTeamRepository / ListTeamRepositories / RemoveTeamRepository
	repo, _, err := c.CreateOrgRepo(org.UserName, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = c.DeleteRepo(org.UserName, repo.Name) })

	repos, _, err := c.ListTeamRepositories(team.ID, ListTeamRepositoriesOptions{})
	require.NoError(t, err)
	assert.False(t, containsRepoName(repos, repo.Name))

	_, err = c.AddTeamRepository(team.ID, org.UserName, repo.Name)
	require.NoError(t, err)

	repos, _, err = c.ListTeamRepositories(team.ID, ListTeamRepositoriesOptions{})
	require.NoError(t, err)
	assert.True(t, containsRepoName(repos, repo.Name))

	_, err = c.RemoveTeamRepository(team.ID, org.UserName, repo.Name)
	require.NoError(t, err)

	repos, _, err = c.ListTeamRepositories(team.ID, ListTeamRepositoriesOptions{})
	require.NoError(t, err)
	assert.False(t, containsRepoName(repos, repo.Name))

	// DeleteTeam
	_, err = c.DeleteTeam(team.ID)
	require.NoError(t, err)
	_, _, err = c.GetTeam(team.ID)
	require.Error(t, err)
}

func containsTeamName(teams []*Team, name string) bool {
	for _, tm := range teams {
		if tm.Name == name {
			return true
		}
	}
	return false
}

func containsUserName(users []*User, name string) bool {
	for _, u := range users {
		if u.UserName == name {
			return true
		}
	}
	return false
}

func containsRepoName(repos []*Repository, name string) bool {
	for _, r := range repos {
		if r.Name == name {
			return true
		}
	}
	return false
}
