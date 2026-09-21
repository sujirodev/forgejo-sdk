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

func TestOrgBlockUnblockUser(t *testing.T) {
	log.Println("== TestOrgBlockUnblockUser ==")
	c := newTestClient()

	orgName := "OrgBlockTestOrg"
	_, _, err := c.GetOrg(orgName)
	if err == nil {
		_, _ = c.DeleteOrg(orgName)
	}
	_, _, err = c.CreateOrg(CreateOrgOption{Name: orgName, Visibility: VisibleTypePublic})
	require.NoError(t, err)
	defer func() { _, _ = c.DeleteOrg(orgName) }()

	user := createTestUser(t, "org_block_test_user", c)

	resp, err := c.OrgBlockUser(t.Context(), orgName, user.UserName)
	require.NoError(t, err)
	require.NotNil(t, resp)

	blocked, resp, err := c.ListOrgBlockedUsers(t.Context(), orgName, ListOrgBlockedUsersOptions{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	found := false
	for _, b := range blocked {
		if b.BlockID == user.ID {
			found = true
		}
	}
	assert.True(t, found, "blocked user should appear in the organization's blocked user list")

	resp, err = c.OrgUnblockUser(t.Context(), orgName, user.UserName)
	require.NoError(t, err)
	require.NotNil(t, resp)

	blocked, _, err = c.ListOrgBlockedUsers(t.Context(), orgName, ListOrgBlockedUsersOptions{})
	require.NoError(t, err)
	for _, b := range blocked {
		assert.NotEqual(t, user.ID, b.BlockID, "user should no longer be blocked")
	}
}
