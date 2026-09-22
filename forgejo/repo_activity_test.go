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

func TestListRepoActivityFeeds(t *testing.T) {
	log.Println("== TestListRepoActivityFeeds ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "ActivityFeeds", c)
	require.NoError(t, err)

	activities, resp, err := c.ListRepoActivityFeeds(repo.Owner.UserName, repo.Name, ListRepoActivityFeedsOptions{})
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, activities)
}
