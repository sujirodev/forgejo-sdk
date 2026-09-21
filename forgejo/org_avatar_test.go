// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

// a 1x1 transparent PNG, base64 encoded
const testAvatarPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII="

func TestOrgAvatar(t *testing.T) {
	log.Println("== TestOrgAvatar ==")
	c := newTestClient()

	orgName := "OrgAvatarTestOrg"
	_, _, err := c.GetOrg(orgName)
	if err == nil {
		_, _ = c.DeleteOrg(orgName)
	}
	_, _, err = c.CreateOrg(CreateOrgOption{Name: orgName, Visibility: VisibleTypePublic})
	require.NoError(t, err)
	defer func() { _, _ = c.DeleteOrg(orgName) }()

	resp, err := c.UpdateOrgAvatar(t.Context(), orgName, UpdateOrgAvatarOption{Image: testAvatarPNGBase64})
	require.NoError(t, err)
	require.NotNil(t, resp)

	org, _, err := c.GetOrg(orgName)
	require.NoError(t, err)
	require.NotEmpty(t, org.AvatarURL)

	resp, err = c.DeleteOrgAvatar(t.Context(), orgName)
	require.NoError(t, err)
	require.NotNil(t, resp)
}
