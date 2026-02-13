// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTags(t *testing.T) {
	log.Println("== TestTags ==")
	c := newTestClient()

	repo, _ := createTestRepo(t, "TestTags", c)

	// Create Tags
	cTagMSG := "A tag message.\n\n:)"
	cTag, resp, err := c.CreateTag(repo.Owner.UserName, repo.Name, CreateTagOption{
		TagName: "tag1",
		Message: cTagMSG,
		Target:  "main",
	})
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
	assert.Equal(t, cTagMSG, cTag.Message)
	assert.Equal(t, fmt.Sprintf("%s/%s/TestTags/archive/tag1.zip", c.url, c.username), cTag.ZipballURL)

	tags, _, err := c.ListRepoTags(repo.Owner.UserName, repo.Name, ListRepoTagsOptions{})
	require.NoError(t, err)
	assert.Len(t, tags, 1)
	assert.Equal(t, cTag, tags[0])

	// get tag
	gTag, _, err := c.GetTag(repo.Owner.UserName, repo.Name, cTag.Name)
	require.NoError(t, err)
	assert.Equal(t, cTag, gTag)

	aTag, _, err := c.GetAnnotatedTag(repo.Owner.UserName, repo.Name, cTag.ID)
	require.NoError(t, err)
	assert.Equal(t, cTag.Name, aTag.Tag)
	assert.Equal(t, cTag.ID, aTag.SHA)
	assert.Equal(t, fmt.Sprintf("%s/api/v1/repos/%s/TestTags/git/tags/%s", c.url, c.username, cTag.ID), aTag.URL)
	assert.Equal(t, cTag.Message+"\n", aTag.Message)
	assert.Equal(t, "commit", aTag.Object.Type)

	// DeleteReleaseTag
	resp, err = c.DeleteTag(repo.Owner.UserName, repo.Name, "tag1")
	require.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)
	tags, _, err = c.ListRepoTags(repo.Owner.UserName, repo.Name, ListRepoTagsOptions{})
	require.NoError(t, err)
	assert.Empty(t, tags)
}
