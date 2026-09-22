// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReleaseAttachments(t *testing.T) {
	t.Parallel()
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})

	release, _, err := c.CreateRelease(repo.Owner.UserName, repo.Name, CreateReleaseOption{
		TagName: "v1.0.0",
		Title:   "v1.0.0",
	})
	require.NoError(t, err)

	attachments, _, err := c.ListReleaseAttachments(repo.Owner.UserName, repo.Name, release.ID)
	require.NoError(t, err)
	assert.Empty(t, attachments)

	created, _, err := c.CreateReleaseAttachment(repo.Owner.UserName, repo.Name, release.ID,
		strings.NewReader("attachment contents"), "notes.txt")
	require.NoError(t, err)
	assert.Equal(t, "notes.txt", created.Name)
	assert.EqualValues(t, len("attachment contents"), created.Size)

	fetched, _, err := c.GetReleaseAttachment(repo.Owner.UserName, repo.Name, release.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)

	attachments, _, err = c.ListReleaseAttachments(repo.Owner.UserName, repo.Name, release.ID)
	require.NoError(t, err)
	require.Len(t, attachments, 1)

	edited, _, err := c.EditReleaseAttachment(repo.Owner.UserName, repo.Name, release.ID, created.ID, EditAttachmentOptions{
		Name: "renamed.txt",
	})
	require.NoError(t, err)
	assert.Equal(t, "renamed.txt", edited.Name)

	_, err = c.DeleteReleaseAttachment(repo.Owner.UserName, repo.Name, release.ID, created.ID)
	require.NoError(t, err)

	attachments, _, err = c.ListReleaseAttachments(repo.Owner.UserName, repo.Name, release.ID)
	require.NoError(t, err)
	assert.Empty(t, attachments)
}
