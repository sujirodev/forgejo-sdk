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

// TestGetNodeInfo needs [federation] ENABLED = true on the test instance
// (see TESTING.md, "Server configuration"); without it GetNodeInfo 404s.
func TestGetNodeInfo(t *testing.T) {
	log.Println("== TestGetNodeInfo ==")
	c := newTestClient()

	info, _, err := c.GetNodeInfo(t.Context())
	require.NoError(t, err)
	assert.NotEmpty(t, info.Version)
	assert.Equal(t, "forgejo", info.Software.Name)
	assert.NotEmpty(t, info.Software.Version)
	assert.GreaterOrEqual(t, info.Usage.Users.Total, int64(0))
}
