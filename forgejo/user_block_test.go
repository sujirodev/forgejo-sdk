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

func TestUserBlock(t *testing.T) {
	t.Parallel()
	log.Println("== TestUserBlock ==")
	c := newTestClient()

	target := createTestUser(t, "userBlockTarget", c)

	resp, err := c.BlockUser(target.UserName)
	require.NoError(t, err)
	require.NotNil(t, resp)

	blocked, resp, err := c.ListBlockedUsers(ListBlockedUsersOptions{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, blocked, 1)
	assert.Equal(t, target.ID, blocked[0].BlockID)

	resp, err = c.UnblockUser(target.UserName)
	require.NoError(t, err)
	require.NotNil(t, resp)

	blocked, resp, err = c.ListBlockedUsers(ListBlockedUsersOptions{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, blocked)
}
