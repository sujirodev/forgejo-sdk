// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBlob(t *testing.T) {
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true, Readme: "Default"})

	tree, _, err := c.GetTrees(repo.Owner.UserName, repo.Name, "main", GetTreesOptions{Recursive: true})
	require.NoError(t, err)
	var readmeSHA string
	for _, entry := range tree.Entries {
		if entry.Path == "README.md" {
			readmeSHA = entry.SHA
		}
	}
	require.NotEmpty(t, readmeSHA, "AutoInit with a Readme should create README.md")

	blob, _, err := c.GetBlob(repo.Owner.UserName, repo.Name, readmeSHA)
	require.NoError(t, err)
	assert.Equal(t, readmeSHA, blob.SHA)
	assert.Equal(t, "base64", blob.Encoding)

	decoded, err := base64.StdEncoding.DecodeString(blob.Content)
	require.NoError(t, err)
	assert.NotEmpty(t, decoded)
}
