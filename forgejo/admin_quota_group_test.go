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

func TestAdminQuotaGroups(t *testing.T) {
	log.Println("== TestAdminQuotaGroups ==")
	c := newTestClient()

	groupName := "sdk-admin-quota-group-test"
	ruleName := "sdk-admin-quota-group-rule-test"
	extraRuleName := "sdk-admin-quota-group-extra-rule-test"

	// best-effort cleanup from a previous failed run
	_, _ = c.AdminDeleteQuotaGroup(groupName)
	_, _ = c.AdminDeleteQuotaRule(ruleName)
	_, _ = c.AdminDeleteQuotaRule(extraRuleName)

	group, resp, err := c.AdminCreateQuotaGroup(CreateQuotaGroupOption{
		Name: groupName,
		Rules: []CreateQuotaRuleOption{
			{Name: ruleName, Limit: 4096, Subjects: []string{"size:all"}},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, groupName, group.Name)
	require.Len(t, group.Rules, 1)
	assert.Equal(t, ruleName, group.Rules[0].Name)
	defer func() {
		_, _ = c.AdminDeleteQuotaGroup(groupName)
		_, _ = c.AdminDeleteQuotaRule(ruleName)
		_, _ = c.AdminDeleteQuotaRule(extraRuleName)
	}()

	got, _, err := c.AdminGetQuotaGroup(groupName)
	require.NoError(t, err)
	assert.Equal(t, groupName, got.Name)

	groups, _, err := c.AdminListQuotaGroups()
	require.NoError(t, err)
	found := false
	for _, g := range groups {
		if g.Name == groupName {
			found = true
			break
		}
	}
	assert.True(t, found, "created group should be present in AdminListQuotaGroups")

	// add a second, standalone rule and attach/detach it from the group
	_, _, err = c.AdminCreateQuotaRule(CreateQuotaRuleOption{Name: extraRuleName, Limit: 1, Subjects: []string{"size:all"}})
	require.NoError(t, err)

	resp, err = c.AdminAddRuleToQuotaGroup(groupName, extraRuleName)
	require.NoError(t, err)
	require.NotNil(t, resp)

	got, _, err = c.AdminGetQuotaGroup(groupName)
	require.NoError(t, err)
	assert.Len(t, got.Rules, 2)

	resp, err = c.AdminRemoveRuleFromQuotaGroup(groupName, extraRuleName)
	require.NoError(t, err)
	require.NotNil(t, resp)

	got, _, err = c.AdminGetQuotaGroup(groupName)
	require.NoError(t, err)
	assert.Len(t, got.Rules, 1)

	// membership
	me, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	resp, err = c.AdminAddUserToQuotaGroup(groupName, me.UserName)
	require.NoError(t, err)
	require.NotNil(t, resp)

	users, _, err := c.AdminListUsersInQuotaGroup(groupName)
	require.NoError(t, err)
	foundUser := false
	for _, u := range users {
		if u.UserName == me.UserName {
			foundUser = true
			break
		}
	}
	assert.True(t, foundUser, "added user should be present in AdminListUsersInQuotaGroup")

	resp, err = c.AdminRemoveUserFromQuotaGroup(groupName, me.UserName)
	require.NoError(t, err)
	require.NotNil(t, resp)

	users, _, err = c.AdminListUsersInQuotaGroup(groupName)
	require.NoError(t, err)
	for _, u := range users {
		assert.NotEqual(t, me.UserName, u.UserName)
	}

	resp, err = c.AdminDeleteQuotaGroup(groupName)
	require.NoError(t, err)
	require.NotNil(t, resp)

	_, _, err = c.AdminGetQuotaGroup(groupName)
	require.Error(t, err)
}
