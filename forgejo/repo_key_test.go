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

func TestGetRepoSigningKey(t *testing.T) {
	log.Println("== TestGetRepoSigningKey ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "SigningKey", c)
	require.NoError(t, err)

	key, resp, err := c.GetRepoSigningKey(repo.Owner.UserName, repo.Name)
	assert.NotNil(t, resp)
	// The test instance may or may not have instance signing configured.
	// Either way the request itself must reach the route: a configured
	// instance returns 200 with an ASCII-armored key, an unconfigured one
	// returns 404 since there is no key to serve.
	if err == nil {
		assert.IsType(t, "", key)
	} else {
		assert.Equal(t, 404, resp.StatusCode)
	}
}
