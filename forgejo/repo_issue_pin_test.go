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

func TestRepoIssuePins(t *testing.T) {
	t.Parallel()
	log.Println("== TestRepoIssuePins ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "IssuePins", c)
	require.NoError(t, err)

	pinned, _, err := c.ListPinnedIssues(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.Empty(t, pinned)

	allowed, _, err := c.NewPinAllowed(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.True(t, allowed.Issues)
	assert.True(t, allowed.PullRequests)
}
