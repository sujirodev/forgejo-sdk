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

func TestGetIssueTemplates(t *testing.T) {
	t.Parallel()
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})

	templates, _, err := c.GetIssueTemplates(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.Empty(t, templates)

	content := "---\n" +
		"name: Bug report\n" +
		"about: Report a bug\n" +
		"title: \"[BUG] \"\n" +
		"labels: bug\n" +
		"---\n\n" +
		"Describe the bug.\n"
	_, _, err = c.CreateFile(repo.Owner.UserName, repo.Name, ".forgejo/ISSUE_TEMPLATE/bug.md", CreateFileOptions{
		Content: base64.StdEncoding.EncodeToString([]byte(content)),
		FileOptions: FileOptions{
			Message: "add bug issue template",
		},
	})
	require.NoError(t, err)

	templates, _, err = c.GetIssueTemplates(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	require.Len(t, templates, 1)
	assert.Equal(t, "Bug report", templates[0].Name)
	assert.Equal(t, "Report a bug", templates[0].About)
	assert.False(t, templates[0].IsForm())
	assert.Contains(t, templates[0].MarkdownContent, "Describe the bug.")
}
