// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccessToken_DeleteByIDAndByName(t *testing.T) {
	t.Parallel()
	c := newTestClient()
	owner := createTestUser(t, uniqueName(t, "tokenowner"), c)

	var byID *AccessToken
	asUser(t, c, owner.UserName, func() {
		var err error
		byID, _, err = c.CreateAccessToken(owner.UserName, CreateAccessTokenOption{Name: "delete-by-id", Scopes: []AccessTokenScope{AccessTokenScopeAll}})
		require.NoError(t, err)
		_, _, err = c.CreateAccessToken(owner.UserName, CreateAccessTokenOption{Name: "delete-by-name", Scopes: []AccessTokenScope{AccessTokenScopeAll}})
		require.NoError(t, err)
	})

	ctx := context.Background()

	resp, err := c.DeleteAccessTokenByID(ctx, owner.UserName, byID.ID)
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	resp, err = c.DeleteAccessTokenByName(ctx, owner.UserName, "delete-by-name")
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	tokens, _, err := c.ListAccessTokens(owner.UserName, ListAccessTokensOptions{})
	require.NoError(t, err)
	assert.Empty(t, tokens)
}
