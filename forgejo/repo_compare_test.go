// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareCommits(t *testing.T) {
	t.Parallel()
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})

	commits, _, err := c.ListRepoCommits(repo.Owner.UserName, repo.Name, ListCommitOptions{})
	require.NoError(t, err)
	require.Len(t, commits, 1)
	initialSHA := commits[0].SHA

	_, _, err = c.CreateFile(repo.Owner.UserName, repo.Name, "second.txt", CreateFileOptions{
		Content:     "c2Vjb25kIGZpbGU=", // "second file"
		FileOptions: FileOptions{Message: "add second file"},
	})
	require.NoError(t, err)

	cmp, _, err := c.CompareCommits(repo.Owner.UserName, repo.Name, initialSHA, "main")
	require.NoError(t, err)
	assert.Equal(t, 1, cmp.TotalCommits)
	require.Len(t, cmp.Commits, 1)
	assert.Equal(t, "add second file\n", cmp.Commits[0].RepoCommit.Message)
}
