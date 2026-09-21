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

func TestOrgQuota(t *testing.T) {
	log.Println("== TestOrgQuota ==")
	c := newTestClient()

	orgName := "OrgQuotaTestOrg"
	_, _, err := c.GetOrg(orgName)
	if err == nil {
		_, _ = c.DeleteOrg(orgName)
	}
	_, _, err = c.CreateOrg(CreateOrgOption{Name: orgName, Visibility: VisibleTypePublic})
	require.NoError(t, err)
	defer func() { _, _ = c.DeleteOrg(orgName) }()

	t.Run("GetOrgQuota", func(t *testing.T) {
		quota, resp, err := c.GetOrgQuota(t.Context(), orgName)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.GreaterOrEqual(t, quota.Used.Size.Repos.Public, int64(0))
		assert.GreaterOrEqual(t, quota.Used.Size.Repos.Private, int64(0))
	})

	t.Run("ListOrgQuotaArtifacts", func(t *testing.T) {
		artifacts, resp, err := c.ListOrgQuotaArtifacts(t.Context(), orgName, ListOrgQuotaArtifactsOptions{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotNil(t, artifacts)
	})

	t.Run("ListOrgQuotaAttachments", func(t *testing.T) {
		attachments, resp, err := c.ListOrgQuotaAttachments(t.Context(), orgName, ListOrgQuotaAttachmentsOptions{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotNil(t, attachments)
	})

	t.Run("ListOrgQuotaPackages", func(t *testing.T) {
		packages, resp, err := c.ListOrgQuotaPackages(t.Context(), orgName, ListOrgQuotaPackagesOptions{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotNil(t, packages)
	})

	t.Run("CheckOrgQuota", func(t *testing.T) {
		overQuota, resp, err := c.CheckOrgQuota(t.Context(), orgName, QuotaSubjectSizeAll)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.IsType(t, false, overQuota)
	})

	t.Run("CheckOrgQuotaRequiresSubject", func(t *testing.T) {
		_, _, err := c.CheckOrgQuota(t.Context(), orgName, "")
		require.Error(t, err)
	})
}
