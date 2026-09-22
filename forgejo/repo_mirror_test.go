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

// Needs [migrations] ALLOW_LOCALNETWORKS = true on the test instance (see
// docs/PLANO-COBERTURA-TESTES.md section 4 / scripts/check-test-instance-settings.sh):
// otherwise Forgejo refuses a remote address that resolves to itself.
func TestPushMirrors(t *testing.T) {
	t.Parallel()
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

// TestPushMirrorsLifecycle exercises the rest of the push-mirror surface
// (list, get-by-name, sync, delete) that TestPushMirrors above doesn't
// reach. Needs [migrations] ALLOW_LOCALNETWORKS = true on the test
// instance, same as TestPushMirrors: an "example.invalid" remote gets
// "Permission denied" instead of a network error, so this points at a
// real repo too.
func TestPushMirrorsLifecycle(t *testing.T) {
	t.Parallel()
	log.Println("== TestPushMirrorsLifecycle ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "PushMirrorsLifecycle", c)
	require.NoError(t, err)
	target, err := createTestRepo(t, "PushMirrorsLifecycleTarget", c)
	require.NoError(t, err)

	pml, _, err := c.ListPushMirrors(repo.Owner.UserName, repo.Name, ListPushMirrorsOptions{})
	require.NoError(t, err)
	assert.Empty(t, pml)

	pm, _, err := c.PushMirrors(repo.Owner.UserName, repo.Name, CreatePushMirrorOption{
		Interval:       "8h",
		RemoteAddress:  target.CloneURL,
		RemoteUsername: getForgejoUsername(),
		RemotePassword: getForgejoPassword(),
	})
	require.NoError(t, err)
	require.NotNil(t, pm)
	assert.NotEmpty(t, pm.RemoteName)

	pml, _, err = c.ListPushMirrors(repo.Owner.UserName, repo.Name, ListPushMirrorsOptions{})
	require.NoError(t, err)
	require.Len(t, pml, 1)

	got, _, err := c.GetPushMirrorByRemoteName(repo.Owner.UserName, repo.Name, pml[0].RemoteName)
	require.NoError(t, err)
	assert.Equal(t, pml[0].RemoteName, got.RemoteName)

	_, err = c.PushMirrorSync(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)

	_, err = c.DeletePushMirror(repo.Owner.UserName, repo.Name, pml[0].RemoteName)
	require.NoError(t, err)

	pml, _, err = c.ListPushMirrors(repo.Owner.UserName, repo.Name, ListPushMirrorsOptions{})
	require.NoError(t, err)
	assert.Empty(t, pml)
}

func TestConvertToNormalRepo(t *testing.T) {
	t.Parallel()
	log.Println("== TestConvertToNormalRepo ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "ConvertMirror", c)
	require.NoError(t, err)

	// the repo isn't a mirror, so converting it must fail
	_, resp, err := c.ConvertToNormalRepo(repo.Owner.UserName, repo.Name)
	require.Error(t, err)
	assert.NotNil(t, resp)
}
