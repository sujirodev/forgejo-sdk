// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	log.Printf("== TestVersion ==")
	c := newTestClient()
	rawVersion, _, err := c.ServerVersion()
	require.NoError(t, err)
	assert.NotEmpty(t, rawVersion)

	require.NoError(t, c.checkServerVersionGreaterThanOrEqual(version8_0_3))
	require.Error(t, c.CheckServerVersionConstraint("< 8.0.3"))

	require.NoError(t, c.checkServerVersionGreaterThanOrEqual(version1_11_0))
	require.Error(t, c.CheckServerVersionConstraint("< 1.11.0"))

	c.serverVersion = version1_11_0
	require.Error(t, c.checkServerVersionGreaterThanOrEqual(version1_15_0))
	c.ignoreVersion = true
	require.NoError(t, c.checkServerVersionGreaterThanOrEqual(version1_15_0))

	c, err = NewClient(getForgejoURL(), newTestClientAuth(), SetForgejoVersion("1.12.123"))
	require.NoError(t, err)
	require.NoError(t, c.CheckServerVersionConstraint("=1.12.123"))
}
