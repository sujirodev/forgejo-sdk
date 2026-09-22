// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserPublicKeys(t *testing.T) {
	c := newTestClient()
	owner := createTestUser(t, uniqueName(t, "keyowner"), c)

	var created *PublicKey
	asUser(t, c, owner.UserName, func() {
		var err error
		created, _, err = c.CreatePublicKey(CreateKeyOption{
			Title: "sdk-test-key",
			Key:   genSSHPublicKey(t, "sdk-test"),
		})
		require.NoError(t, err)
	})
	assert.Equal(t, "sdk-test-key", created.Title)

	var mine []*PublicKey
	asUser(t, c, owner.UserName, func() {
		var err error
		mine, _, err = c.ListMyPublicKeys(ListPublicKeysOptions{})
		require.NoError(t, err)
	})
	require.Len(t, mine, 1)
	assert.Equal(t, created.ID, mine[0].ID)

	keys, _, err := c.ListPublicKeys(owner.UserName, ListPublicKeysOptions{})
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, created.ID, keys[0].ID)

	var fetched *PublicKey
	asUser(t, c, owner.UserName, func() {
		var err error
		fetched, _, err = c.GetPublicKey(created.ID)
		require.NoError(t, err)
	})
	assert.Equal(t, created.Key, fetched.Key)

	asUser(t, c, owner.UserName, func() {
		_, err := c.DeletePublicKey(created.ID)
		require.NoError(t, err)
	})

	keys, _, err = c.ListPublicKeys(owner.UserName, ListPublicKeysOptions{})
	require.NoError(t, err)
	assert.Empty(t, keys)
}
