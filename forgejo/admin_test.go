// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminOrg(t *testing.T) {
	log.Println("== TestAdminOrg ==")
	c := newTestClient()
	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	orgName := "NewTestOrg"
	newOrg, _, err := c.AdminCreateOrg(user.UserName, CreateOrgOption{
		Name:        orgName,
		FullName:    orgName + " FullName",
		Description: "test adminCreateOrg",
		Visibility:  VisibleTypePublic,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, newOrg)
	assert.Equal(t, orgName, newOrg.UserName)

	orgs, _, err := c.AdminListOrgs(AdminListOrgsOptions{})
	require.NoError(t, err)
	if assert.GreaterOrEqual(t, len(orgs), 1) {
		orgs = orgs[len(orgs)-1:]
		assert.Equal(t, newOrg.ID, orgs[0].ID)
	}

	_, err = c.DeleteOrg(orgName)
	require.NoError(t, err)
}

func TestAdminListEditUsers(t *testing.T) {
	c := newTestClient()
	user := createTestUser(t, uniqueName(t, "adminedit"), c)

	users, _, err := c.AdminListUsers(AdminListUsersOptions{})
	require.NoError(t, err)
	found := false
	for _, u := range users {
		if u.UserName == user.UserName {
			found = true
		}
	}
	assert.True(t, found, "AdminListUsers should list the just-created user")

	fullName := "Edited By AdminEditUser"
	_, err = c.AdminEditUser(user.UserName, EditUserOption{
		LoginName: user.UserName,
		FullName:  &fullName,
	})
	require.NoError(t, err)

	edited, _, err := c.GetUserInfo(user.UserName)
	require.NoError(t, err)
	assert.Equal(t, fullName, edited.FullName)
}

func TestAdminUserPublicKeys(t *testing.T) {
	c := newTestClient()
	owner := createTestUser(t, uniqueName(t, "adminkey"), c)

	key, _, err := c.AdminCreateUserPublicKey(owner.UserName, CreateKeyOption{
		Title: "sdk-test-admin-key",
		Key:   genSSHPublicKey(t, "sdk-admin-test"),
	})
	require.NoError(t, err)
	assert.Equal(t, "sdk-test-admin-key", key.Title)

	keys, _, err := c.ListPublicKeys(owner.UserName, ListPublicKeysOptions{})
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, key.ID, keys[0].ID)

	_, err = c.AdminDeleteUserPublicKey(owner.UserName, int(key.ID))
	require.NoError(t, err)

	keys, _, err = c.ListPublicKeys(owner.UserName, ListPublicKeysOptions{})
	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestAdminCronTasks(t *testing.T) {
	log.Println("== TestAdminCronTasks ==")
	c := newTestClient()

	tasks, _, err := c.ListCronTasks(ListCronTaskOptions{})
	require.NoError(t, err)
	require.Greater(t, len(tasks), 15)
	_, err = c.RunCronTasks(tasks[0].Name)
	require.NoError(t, err)
}
