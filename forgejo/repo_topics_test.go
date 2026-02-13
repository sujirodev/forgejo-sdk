// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoTopics(t *testing.T) {
	log.Println("== TestRepoTopics ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "RandomTopic", c)
	require.NoError(t, err)

	// Add
	_, err = c.AddRepoTopic(repo.Owner.UserName, repo.Name, "best")
	require.NoError(t, err)
	_, err = c.AddRepoTopic(repo.Owner.UserName, repo.Name, "git")
	require.NoError(t, err)
	_, err = c.AddRepoTopic(repo.Owner.UserName, repo.Name, "forgejo")
	require.NoError(t, err)
	_, err = c.AddRepoTopic(repo.Owner.UserName, repo.Name, "drone")
	require.NoError(t, err)

	// Get List
	tl, _, err := c.ListRepoTopics(repo.Owner.UserName, repo.Name, ListRepoTopicsOptions{})
	require.NoError(t, err)
	assert.Len(t, tl, 4)

	// Del
	_, err = c.DeleteRepoTopic(repo.Owner.UserName, repo.Name, "drone")
	require.NoError(t, err)
	_, err = c.DeleteRepoTopic(repo.Owner.UserName, repo.Name, "best")
	require.NoError(t, err)
	tl, _, err = c.ListRepoTopics(repo.Owner.UserName, repo.Name, ListRepoTopicsOptions{})
	require.NoError(t, err)
	assert.Len(t, tl, 2)

	// Set List
	newTopics := []string{"analog", "digital", "cat"}
	_, err = c.SetRepoTopics(repo.Owner.UserName, repo.Name, newTopics)
	require.NoError(t, err)
	tl, _, _ = c.ListRepoTopics(repo.Owner.UserName, repo.Name, ListRepoTopicsOptions{})
	assert.Len(t, tl, 3)

	sort.Strings(tl)
	sort.Strings(newTopics)
	assert.Equal(t, newTopics, tl)
}
