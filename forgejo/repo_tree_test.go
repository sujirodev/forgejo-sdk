// Copyright 2025 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoTrees(t *testing.T) {
	log.Println("== TestRepoTrees ==")
	c := newTestClient()
	rawVersion, _, err := c.ServerVersion()
	require.NoError(t, err)
	assert.NotEmpty(t, rawVersion)

	repoName := "gettrees"
	repo := prepareTreeTest(t, c, repoName)
	if repo == nil {
		return
	}
	time.Sleep(1 * time.Second)

	// test without recursive option set
	tl, _, err := c.GetTrees(repo.Owner.UserName, repo.Name, "main", GetTreesOptions{Recursive: false})
	require.NoError(t, err)
	assert.Len(t, tl.Entries, 4)

	// test with recursive option set
	tl, _, err = c.GetTrees(repo.Owner.UserName, repo.Name, "main", GetTreesOptions{Recursive: true})
	require.NoError(t, err)
	assert.False(t, tl.Truncated)
	assert.Len(t, tl.Entries, 24)

	// test with recursive option set and page size set to 3
	tl, _, err = c.GetTrees(repo.Owner.UserName, repo.Name, "main", GetTreesOptions{
		Recursive:   true,
		ListOptions: ListOptions{Page: 1, PageSize: 3},
	})
	require.NoError(t, err)
	assert.Len(t, tl.Entries, 3)
	assert.True(t, tl.Truncated)
	assert.Equal(t, 24, tl.TotalCount)

	// test with recursive option set and page size set to 3, making sure page 4 also has 3 entries
	tl, _, err = c.GetTrees(repo.Owner.UserName, repo.Name, "main", GetTreesOptions{
		Recursive:   true,
		ListOptions: ListOptions{Page: 4, PageSize: 3},
	})
	require.NoError(t, err)
	assert.Len(t, tl.Entries, 3)
	assert.True(t, tl.Truncated)
	assert.Equal(t, 4, tl.Page)
	assert.Equal(t, 24, tl.TotalCount)
}

func prepareTreeTest(t *testing.T, c *Client, repoName string) *Repository {
	origRepo, err := createTestRepo(t, repoName, c)
	if !assert.NoError(t, err) {
		return nil
	}

	// Add 20 files in subdirectory dirA so we can test recursive option
	for i := 0; i < 20; i++ {
		newFile, _, err := c.CreateFile(origRepo.Owner.UserName, origRepo.Name, fmt.Sprintf("dirA/file-%d", i), CreateFileOptions{
			Content: "QSBuZXcgRmlsZQo=",
			FileOptions: FileOptions{
				Message:    fmt.Sprintf("create a new file %d", i),
				BranchName: "main",
			},
		})
		if !assert.NoError(t, err) || !assert.NotNil(t, newFile) { //nolint:testifylint
			return nil
		}
	}

	return origRepo
}
