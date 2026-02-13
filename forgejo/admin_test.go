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

func TestAdminCronTasks(t *testing.T) {
	log.Println("== TestAdminCronTasks ==")
	c := newTestClient()

	tasks, _, err := c.ListCronTasks(ListCronTaskOptions{})
	require.NoError(t, err)
	require.Greater(t, len(tasks), 15)
	_, err = c.RunCronTasks(tasks[0].Name)
	require.NoError(t, err)
}
