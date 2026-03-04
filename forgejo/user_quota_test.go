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

func TestGetMyQuota(t *testing.T) {
	log.Println("== TestGetMyQuota ==")
	c := newTestClient()

	quota, _, err := c.GetMyQuota(t.Context())
	require.NoError(t, err)

	// Verify structure is populated
	// Even with no usage, these should be 0, not a crash
	assert.GreaterOrEqual(t, quota.Used.Size.Repos.Public, int64(0))

	// Groups may be nil if no quota groups are assigned.
	assert.GreaterOrEqual(t, len(quota.Groups), 0)
}

func TestListMyQuotaArtifacts(t *testing.T) {
	log.Println("== TestListMyQuotaArtifacts ==")
	c := newTestClient()

	artifacts, resp, err := c.ListMyQuotaArtifacts(t.Context(), ListMyQuotaArtifactsOptions{})
	require.NoError(t, err)
	assert.NotNil(t, artifacts)
	assert.NotNil(t, resp)

	// Artifacts list may be empty if user has no artifacts
	// but should return a valid slice
	for _, artifact := range artifacts {
		assert.NotEmpty(t, artifact.Name)
		assert.GreaterOrEqual(t, artifact.Size, int64(0))
	}
}

func TestListMyQuotaArtifactsWithPagination(t *testing.T) {
	log.Println("== TestListMyQuotaArtifactsWithPagination ==")
	c := newTestClient()

	// Test with pagination options
	artifacts, resp, err := c.ListMyQuotaArtifacts(t.Context(), ListMyQuotaArtifactsOptions{
		ListOptions: ListOptions{
			Page:     1,
			PageSize: 10,
		},
	})
	require.NoError(t, err)
	assert.NotNil(t, artifacts)
	assert.NotNil(t, resp)

	// Verify pagination doesn't break the response
	assert.LessOrEqual(t, len(artifacts), 10)
}

func TestListMyQuotaAttachments(t *testing.T) {
	log.Println("== TestListMyQuotaAttachments ==")
	c := newTestClient()

	attachments, resp, err := c.ListMyQuotaAttachments(t.Context(), ListMyQuotaAttachmentsOptions{})
	require.NoError(t, err)
	assert.NotNil(t, attachments)
	assert.NotNil(t, resp)

	// Attachments list may be empty if user has no attachments
	// but should return a valid slice
	for _, attachment := range attachments {
		assert.NotEmpty(t, attachment.Name)
		assert.GreaterOrEqual(t, attachment.Size, int64(0))
		// APIURL should be present
		assert.NotEmpty(t, attachment.APIURL)
	}
}

func TestListMyQuotaAttachmentsWithPagination(t *testing.T) {
	log.Println("== TestListMyQuotaAttachmentsWithPagination ==")
	c := newTestClient()

	// Test with pagination options
	attachments, resp, err := c.ListMyQuotaAttachments(t.Context(), ListMyQuotaAttachmentsOptions{
		ListOptions: ListOptions{
			Page:     1,
			PageSize: 5,
		},
	})
	require.NoError(t, err)
	assert.NotNil(t, attachments)
	assert.NotNil(t, resp)

	// Verify pagination doesn't break the response
	assert.LessOrEqual(t, len(attachments), 5)
}

func TestListMyQuotaPackages(t *testing.T) {
	log.Println("== TestListMyQuotaPackages ==")
	c := newTestClient()

	packages, resp, err := c.ListMyQuotaPackages(t.Context(), ListMyQuotaPackagesOptions{})
	require.NoError(t, err)
	assert.NotNil(t, packages)
	assert.NotNil(t, resp)

	// Packages list may be empty if user has no packages
	// but should return a valid slice
	for _, pkg := range packages {
		assert.NotEmpty(t, pkg.Name)
		assert.GreaterOrEqual(t, pkg.Size, int64(0))
		assert.NotEmpty(t, pkg.Type)
	}
}

func TestListMyQuotaPackagesWithPagination(t *testing.T) {
	log.Println("== TestListMyQuotaPackagesWithPagination ==")
	c := newTestClient()

	// Test with pagination options
	packages, resp, err := c.ListMyQuotaPackages(t.Context(), ListMyQuotaPackagesOptions{
		ListOptions: ListOptions{
			Page:     1,
			PageSize: 10,
		},
	})
	require.NoError(t, err)
	assert.NotNil(t, packages)
	assert.NotNil(t, resp)

	// Verify pagination doesn't break the response
	assert.LessOrEqual(t, len(packages), 10)
}

func TestCheckMyQuota(t *testing.T) {
	log.Println("== TestCheckMyQuota ==")
	c := newTestClient()

	overQuota, _, err := c.CheckMyQuota(t.Context(), QuotaSubjectSizeAll)
	require.NoError(t, err)

	// The result should be a boolean (true or false)
	// Most users won't be over quota, but we can't assert the value
	// Just verify the call succeeds and returns a boolean
	assert.IsType(t, false, overQuota)
}

func TestQuotaInfoStructure(t *testing.T) {
	log.Println("== TestQuotaInfoStructure ==")
	c := newTestClient()

	quota, _, err := c.GetMyQuota(t.Context())
	require.NoError(t, err)

	// Test that nested structures are accessible without nil checks
	assert.GreaterOrEqual(t, quota.Used.Size.Repos.Public, int64(0))
	assert.GreaterOrEqual(t, quota.Used.Size.Repos.Private, int64(0))
	assert.GreaterOrEqual(t, quota.Used.Size.Git.LFS, int64(0))
	assert.GreaterOrEqual(t, quota.Used.Size.Assets.Artifacts, int64(0))
	assert.GreaterOrEqual(t, quota.Used.Size.Assets.Attachments.Issues, int64(0))
	assert.GreaterOrEqual(t, quota.Used.Size.Assets.Attachments.Releases, int64(0))
	assert.GreaterOrEqual(t, quota.Used.Size.Assets.Packages.All, int64(0))
}

func TestQuotaGroupsAndRules(t *testing.T) {
	log.Println("== TestQuotaGroupsAndRules ==")
	c := newTestClient()

	quota, _, err := c.GetMyQuota(t.Context())
	require.NoError(t, err)

	// If quota groups are assigned, verify their structure
	if len(quota.Groups) > 0 {
		for _, group := range quota.Groups {
			// Group should have a name
			assert.NotEmpty(t, group.Name)

			// If rules exist, verify their structure
			if len(group.Rules) > 0 {
				for _, rule := range group.Rules {
					// Rules should have a limit
					assert.GreaterOrEqual(t, rule.Limit, int64(0))

					// Subjects may be empty but should be a valid slice
					assert.NotNil(t, rule.Subjects)
				}
			}
		}
	}
}

func TestQuotaAttachmentContainedIn(t *testing.T) {
	log.Println("== TestQuotaAttachmentContainedIn ==")
	c := newTestClient()

	attachments, _, err := c.ListMyQuotaAttachments(t.Context(), ListMyQuotaAttachmentsOptions{})
	require.NoError(t, err)

	// If attachments exist, verify ContainedIn structure
	for _, attachment := range attachments {
		if attachment.ContainedIn != nil {
			// ContainedIn should have at least one URL
			assert.True(t,
				attachment.ContainedIn.APIURL != "" || attachment.ContainedIn.HTMLURL != "",
				"ContainedIn should have at least one URL populated")
		}
	}
}
