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

func TestGetBlobs(t *testing.T) {
	log.Println("== TestGetBlobs ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "GetBlobs", c)
	require.NoError(t, err)

	dir, _, err := c.ListContents(repo.Owner.UserName, repo.Name, "", "")
	require.NoError(t, err)
	require.NotEmpty(t, dir)

	shas := make([]string, 0, len(dir))
	for _, entry := range dir {
		shas = append(shas, entry.SHA)
	}

	blob, _, err := c.GetBlob(repo.Owner.UserName, repo.Name, shas[0])
	require.NoError(t, err)
	require.NotNil(t, blob)

	blobs, _, err := c.GetBlobs(repo.Owner.UserName, repo.Name, shas)
	require.NoError(t, err)
	assert.Len(t, blobs, len(shas))
}
