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

func TestPushMirrors(t *testing.T) {
	log.Println("== TestPushMirrors ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "PushMirrors", c)
	require.NoError(t, err)

	pml, _, err := c.ListPushMirrors(repo.Owner.UserName, repo.Name, ListPushMirrorsOptions{})
	require.NoError(t, err)
	assert.Empty(t, pml)

	pm, _, err := c.PushMirrors(repo.Owner.UserName, repo.Name, CreatePushMirrorOption{
		Interval:      "8h",
		RemoteAddress: "https://example.invalid/push-mirror-target.git",
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
	log.Println("== TestConvertToNormalRepo ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "ConvertMirror", c)
	require.NoError(t, err)

	// the repo isn't a mirror, so converting it must fail
	_, resp, err := c.ConvertToNormalRepo(repo.Owner.UserName, repo.Name)
	require.Error(t, err)
	assert.NotNil(t, resp)
}
