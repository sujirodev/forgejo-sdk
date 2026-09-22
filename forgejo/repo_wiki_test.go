// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"encoding/base64"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoWiki(t *testing.T) {
	t.Parallel()
	log.Println("== TestRepoWiki ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "RepoWiki", c)
	require.NoError(t, err)

	content := base64.StdEncoding.EncodeToString([]byte("# Hello Wiki\n"))
	page, _, err := c.CreateWikiPage(repo.Owner.UserName, repo.Name, CreateWikiPageOptions{
		Title:         "First Page",
		ContentBase64: content,
		Message:       "add first page",
	})
	require.NoError(t, err)
	require.NotNil(t, page)
	assert.Equal(t, "First Page", page.Title)

	got, _, err := c.GetWikiPage(repo.Owner.UserName, repo.Name, "First Page")
	require.NoError(t, err)
	assert.Equal(t, content, got.ContentBase64)

	newContent := base64.StdEncoding.EncodeToString([]byte("# Hello Wiki\n\nUpdated.\n"))
	updated, _, err := c.EditWikiPage(repo.Owner.UserName, repo.Name, "First Page", CreateWikiPageOptions{
		Title:         "First Page",
		ContentBase64: newContent,
		Message:       "update page",
	})
	require.NoError(t, err)
	assert.Equal(t, newContent, updated.ContentBase64)

	pages, _, err := c.ListWikiPages(repo.Owner.UserName, repo.Name, ListWikiPagesOptions{})
	require.NoError(t, err)
	assert.Len(t, pages, 1)

	revisions, _, err := c.ListWikiPageRevisions(repo.Owner.UserName, repo.Name, "First Page", ListWikiPageRevisionsOptions{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, revisions.Count, int64(2))

	_, err = c.DeleteWikiPage(repo.Owner.UserName, repo.Name, "First Page")
	require.NoError(t, err)

	pages, _, err = c.ListWikiPages(repo.Owner.UserName, repo.Name, ListWikiPagesOptions{})
	require.NoError(t, err)
	assert.Empty(t, pages)
}
