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

func TestGetRepoAllGitRefs(t *testing.T) {
	log.Println("== TestGetRepoAllGitRefs ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "AllGitRefs", c)
	require.NoError(t, err)

	refs, _, err := c.GetRepoAllGitRefs(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	require.NotEmpty(t, refs)

	found := false
	for _, r := range refs {
		if r.Ref == "refs/heads/main" {
			found = true
		}
	}
	assert.True(t, found)
}
