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

func TestListAllOrgs(t *testing.T) {
	log.Println("== TestListAllOrgs ==")
	c := newTestClient()

	orgName := "ListAllTestOrg"
	_, _, err := c.GetOrg(orgName)
	if err != nil {
		_, _, err = c.CreateOrg(CreateOrgOption{
			Name:       orgName,
			Visibility: VisibleTypePublic,
		})
		require.NoError(t, err)
	}

	orgs, _, err := c.ListOrgs(ListOrgsOptions{})
	require.NoError(t, err)

	foundOrg := false
	for _, org := range orgs {
		if org.UserName == orgName {
			foundOrg = true
		}
	}
	assert.True(t, foundOrg)

	_, err = c.DeleteOrg(orgName)
	assert.NoError(t, err, "failed to delete org")
}

func TestOrgs_ListMyListUserEdit(t *testing.T) {
	c := newTestClient()
	org := newTestOrg(t, c)

	myOrgs, _, err := c.ListMyOrgs(ListOrgsOptions{})
	require.NoError(t, err)
	assert.True(t, containsOrgName(myOrgs, org.UserName))

	me, _, err := c.GetMyUserInfo()
	require.NoError(t, err)
	userOrgs, _, err := c.ListUserOrgs(me.UserName, ListOrgsOptions{})
	require.NoError(t, err)
	assert.True(t, containsOrgName(userOrgs, org.UserName))

	_, err = c.EditOrg(org.UserName, EditOrgOption{
		FullName:   "Edited by EditOrg",
		Visibility: VisibleTypePublic,
	})
	require.NoError(t, err)

	edited, _, err := c.GetOrg(org.UserName)
	require.NoError(t, err)
	assert.Equal(t, "Edited by EditOrg", edited.FullName)
}

func containsOrgName(orgs []*Organization, name string) bool {
	for _, o := range orgs {
		if o.UserName == name {
			return true
		}
	}
	return false
}

func createTestOrgRepo(t *testing.T, c *Client, name string) (func(), *Repository, error) {
	_, _, err := c.GetOrg(name)
	if err == nil {
		_, _ = c.DeleteOrg(name)
	}
	_, _, err = c.CreateOrg(CreateOrgOption{
		Name:                      name,
		Visibility:                VisibleTypePublic,
		RepoAdminChangeTeamAccess: true,
	})
	if !assert.NoError(t, err) { //nolint
		return nil, nil, err
	}

	_, _, err = c.GetRepo(name, name)
	if err == nil {
		_, _ = c.DeleteRepo(name, name)
	}

	repo, _, err := c.CreateOrgRepo(name, CreateRepoOption{
		Name:        name,
		Description: "A test Repo: " + name,
		AutoInit:    true,
		Gitignores:  "C,C++",
		License:     "MIT",
		Readme:      "Default",
		IssueLabels: "Default",
		Private:     false,
	})
	require.NoError(t, err)
	assert.NotNil(t, repo)

	return func() {
		_, _ = c.DeleteRepo(name, name)
		_, _ = c.DeleteOrg(name)
	}, repo, err
}
