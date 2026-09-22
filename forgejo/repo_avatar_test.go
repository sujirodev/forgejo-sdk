// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"encoding/base64"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// a minimal 1x1 transparent PNG
const testAvatarPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII="

func TestRepoAvatar(t *testing.T) {
	log.Println("== TestRepoAvatar ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "RepoAvatar", c)
	require.NoError(t, err)

	_, err = base64.StdEncoding.DecodeString(testAvatarPNGBase64)
	require.NoError(t, err)

	_, err = c.UpdateRepoAvatar(repo.Owner.UserName, repo.Name, UpdateRepoAvatarOption{Image: testAvatarPNGBase64})
	require.NoError(t, err)

	updated, _, err := c.GetRepo(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.NotEmpty(t, updated.AvatarURL)

	_, err = c.DeleteRepoAvatar(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)

	updated, _, err = c.GetRepo(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.Empty(t, updated.AvatarURL)
}
