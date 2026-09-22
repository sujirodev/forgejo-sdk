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

func TestRepoTransfer(t *testing.T) {
	t.Parallel()
	log.Printf("== TestRepoTransfer ==")
	c := newTestClient()

	org, _, err := c.AdminCreateOrg(c.username, CreateOrgOption{Name: "TransferOrg"})
	require.NoError(t, err)
	repo, err := createTestRepo(t, "ToMove", c)
	require.NoError(t, err)

	newRepo, _, err := c.TransferRepo(c.username, repo.Name, TransferRepoOption{NewOwner: org.UserName})
	require.NoError(t, err) // admin transfer repository will execute immediately but not set as pendding.
	assert.NotNil(t, newRepo)
	assert.Equal(t, "ToMove", newRepo.Name)

	repo, err = createTestRepo(t, "ToMove", c)
	require.NoError(t, err)
	_, resp, err := c.TransferRepo(c.username, repo.Name, TransferRepoOption{NewOwner: org.UserName})
	assert.Equal(t, 422, resp.StatusCode)
	require.Error(t, err)

	_, err = c.DeleteRepo(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	_, err = c.DeleteRepo(newRepo.Owner.UserName, newRepo.Name)
	require.NoError(t, err)
	_, err = c.DeleteOrg(org.UserName)
	require.NoError(t, err)
}
