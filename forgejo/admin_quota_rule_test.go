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

func TestAdminQuotaRules(t *testing.T) {
	log.Println("== TestAdminQuotaRules ==")
	c := newTestClient()

	ruleName := "sdk-admin-quota-rule-test"
	_, _ = c.AdminDeleteQuotaRule(ruleName) // best-effort cleanup from a previous failed run

	rule, resp, err := c.AdminCreateQuotaRule(CreateQuotaRuleOption{
		Name:     ruleName,
		Limit:    1024,
		Subjects: []string{"size:all"},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, ruleName, rule.Name)
	assert.Equal(t, int64(1024), rule.Limit)
	defer func() { _, _ = c.AdminDeleteQuotaRule(ruleName) }()

	got, _, err := c.AdminGetQuotaRule(ruleName)
	require.NoError(t, err)
	assert.Equal(t, ruleName, got.Name)
	assert.Equal(t, int64(1024), got.Limit)

	rules, _, err := c.AdminListQuotaRules()
	require.NoError(t, err)
	found := false
	for _, r := range rules {
		if r.Name == ruleName {
			found = true
			break
		}
	}
	assert.True(t, found, "created rule should be present in AdminListQuotaRules")

	newLimit := int64(2048)
	edited, resp, err := c.AdminEditQuotaRule(ruleName, EditQuotaRuleOption{Limit: &newLimit})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, newLimit, edited.Limit)

	resp, err = c.AdminDeleteQuotaRule(ruleName)
	require.NoError(t, err)
	require.NotNil(t, resp)

	_, _, err = c.AdminGetQuotaRule(ruleName)
	require.Error(t, err)
}
