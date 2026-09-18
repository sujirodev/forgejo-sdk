// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdminCreateRepo(t *testing.T) {
	c := newTestClient()
	owner := createTestUser(t, uniqueName(t, "adminrepo"), c)

	name := uniqueName(t, "repo")
	repo, _, err := c.AdminCreateRepo(owner.UserName, CreateRepoOption{Name: name, AutoInit: true})
	require.NoError(t, err)
	require.NotNil(t, repo)
	assert.Equal(t, name, repo.Name)
	assert.Equal(t, owner.UserName, repo.Owner.UserName)
	t.Cleanup(func() { _, _ = c.DeleteRepo(owner.UserName, name) })
}
