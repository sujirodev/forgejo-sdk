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

func TestAdminUserEmails(t *testing.T) {
	t.Parallel()
	log.Println("== TestAdminUserEmails ==")
	c := newTestClient()

	// Confirmed absent on a live 13.0.0 instance and present by 15.0.9;
	// below that the SDK's guard refuses the call, which is the documented
	// behavior and worth asserting.
	if !serverAtLeast(t, c, "15.0.0") {
		_, _, err := c.AdminListUserEmails("test01")
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
		return
	}

	user := createTestUser(t, "admin_user_emails_test", c)

	emails, resp, err := c.AdminListUserEmails(user.UserName)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.GreaterOrEqual(t, len(emails), 1)
	assert.Equal(t, user.Email, emails[0].Email)

	extra := "admin_user_emails_test_extra@forgejo.org"

	// Forgejo's admin email endpoints only manage existing addresses, so add
	// one for the user first via the regular user email API, impersonating
	// them with sudo.
	c.SetSudo(user.UserName)
	_, _, err = c.AddEmail(CreateEmailOption{Emails: []string{extra}})
	c.SetSudo("")
	require.NoError(t, err)

	emails, _, err = c.AdminListUserEmails(user.UserName)
	require.NoError(t, err)
	found := false
	for _, e := range emails {
		if e.Email == extra {
			found = true
			break
		}
	}
	assert.True(t, found, "AdminListUserEmails should include the newly added address")

	resp, err = c.AdminDeleteUserEmails(user.UserName, DeleteEmailOption{Emails: []string{extra}})
	require.NoError(t, err)
	require.NotNil(t, resp)

	emails, _, err = c.AdminListUserEmails(user.UserName)
	require.NoError(t, err)
	for _, e := range emails {
		assert.NotEqual(t, extra, e.Email)
	}
}

func TestAdminUserQuota(t *testing.T) {
	t.Parallel()
	log.Println("== TestAdminUserQuota ==")
	c := newTestClient()

	user := createTestUser(t, "admin_user_quota_test", c)

	quota, resp, err := c.AdminGetUserQuota(user.UserName)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.GreaterOrEqual(t, quota.Used.Size.Repos.Public, int64(0))

	groupName := "sdk-admin-user-quota-group-test"
	_, _ = c.AdminDeleteQuotaGroup(groupName)
	_, _, err = c.AdminCreateQuotaGroup(CreateQuotaGroupOption{Name: groupName})
	require.NoError(t, err)
	defer func() { _, _ = c.AdminDeleteQuotaGroup(groupName) }()

	resp, err = c.AdminSetUserQuotaGroups(user.UserName, SetUserQuotaGroupsOption{Groups: []string{groupName}})
	require.NoError(t, err)
	require.NotNil(t, resp)

	quota, _, err = c.AdminGetUserQuota(user.UserName)
	require.NoError(t, err)
	found := false
	for _, g := range quota.Groups {
		if g.Name == groupName {
			found = true
			break
		}
	}
	assert.True(t, found, "AdminGetUserQuota should reflect the assigned quota group")

	resp, err = c.AdminSetUserQuotaGroups(user.UserName, SetUserQuotaGroupsOption{Groups: []string{}})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestAdminRenameUser(t *testing.T) {
	t.Parallel()
	log.Println("== TestAdminRenameUser ==")
	c := newTestClient()

	user := createTestUser(t, "admin_rename_user_test", c)
	newName := "admin_rename_user_test_renamed"

	resp, err := c.AdminRenameUser(user.UserName, RenameUserOption{NewName: newName})
	require.NoError(t, err)
	require.NotNil(t, resp)
	defer func() { _, _ = c.AdminRenameUser(newName, RenameUserOption{NewName: user.UserName}) }()

	renamed, _, err := c.GetUserInfo(newName)
	require.NoError(t, err)
	assert.Equal(t, user.ID, renamed.ID)

	// Confirmed against a live Forgejo 16.0.5 instance with plain curl:
	// the server answers a redirect (307) for the old username, not a
	// 404, and the SDK's HTTP client follows it -- so this is expected to
	// keep resolving, to the same user, not to error out.
	stillResolves, _, err := c.GetUserInfo(user.UserName)
	require.NoError(t, err)
	assert.Equal(t, user.ID, stillResolves.ID)
}

func TestAdminUserAccessTokens(t *testing.T) {
	t.Parallel()
	log.Println("== TestAdminUserAccessTokens ==")
	c := newTestClient()

	// Confirmed absent on a live 15.0.9 instance (404) and present by 16.0.0.
	if !serverAtLeast(t, c, "16.0.0") {
		_, _, err := c.AdminCreateUserAccessToken("test01", CreateAccessTokenOption{Name: "sdk-admin-token-test"})
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
		return
	}

	user := createTestUser(t, "admin_user_tokens_test", c)

	token, resp, err := c.AdminCreateUserAccessToken(user.UserName, CreateAccessTokenOption{
		Name:   "sdk-admin-token-test",
		Scopes: []AccessTokenScope{AccessTokenScopeRepositoryRead},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "sdk-admin-token-test", token.Name)
	require.NotEmpty(t, token.Token)

	tokens, _, err := c.AdminListUserAccessTokens(user.UserName, AdminListUserAccessTokensOptions{})
	require.NoError(t, err)
	found := false
	for _, tk := range tokens {
		if tk.ID == token.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "created token should be present in AdminListUserAccessTokens")

	resp, err = c.AdminDeleteUserAccessToken(user.UserName, token.Name)
	require.NoError(t, err)
	require.NotNil(t, resp)

	tokens, _, err = c.AdminListUserAccessTokens(user.UserName, AdminListUserAccessTokensOptions{})
	require.NoError(t, err)
	for _, tk := range tokens {
		assert.NotEqual(t, token.ID, tk.ID)
	}
}
