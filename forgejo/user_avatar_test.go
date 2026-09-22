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

// tinyPNGBase64 is a 1x1 transparent PNG, base64 encoded, used as a minimal
// valid avatar image.
const tinyPNGBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII="

func TestUserAvatar(t *testing.T) {
	t.Parallel()
	log.Println("== TestUserAvatar ==")
	c := newTestClient()

	resp, err := c.UpdateUserAvatar(UpdateUserAvatarOption{Image: tinyPNGBase64})
	require.NoError(t, err)
	require.NotNil(t, resp)

	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)
	assert.NotEmpty(t, user.AvatarURL)

	// Restore the default avatar so other tests that assert on the
	// deterministic default avatar URL (e.g. TestMyUser) are unaffected.
	resp, err = c.DeleteUserAvatar()
	require.NoError(t, err)
	require.NotNil(t, resp)
}
