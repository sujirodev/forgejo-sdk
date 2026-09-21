// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Needs [migrations] ALLOW_LOCALNETWORKS = true on the test instance (see
// docs/PLANO-COBERTURA-TESTES.md section 4 / scripts/check-test-instance-settings.sh):
// otherwise Forgejo refuses a remote address that resolves to itself.
func TestPushMirrors(t *testing.T) {
	c := newTestClient()
	source := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})
	target := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})

	mirror, _, err := c.PushMirrors(source.Owner.UserName, source.Name, CreatePushMirrorOption{
		RemoteAddress:  target.CloneURL,
		RemoteUsername: getForgejoUsername(),
		RemotePassword: getForgejoPassword(),
		Interval:       "8h0m0s",
	})
	require.NoError(t, err)
	assert.Equal(t, source.Name, mirror.RepoName)
	assert.Contains(t, mirror.RemoteAddress, target.Name)
}
