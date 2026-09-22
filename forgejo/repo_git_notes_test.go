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

func TestRepoGitNotes(t *testing.T) {
	t.Parallel()
	log.Println("== TestRepoGitNotes ==")
	c := newTestClient()
	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	repoName := "GitNotes"
	repo, err := createTestRepo(t, repoName, c)
	require.NoError(t, err)

	commits, _, err := c.ListRepoCommits(user.UserName, repoName, ListCommitOptions{
		SHA: repo.DefaultBranch,
	})
	require.NoError(t, err)
	require.Len(t, commits, 1)
	sha := commits[0].SHA

	_, resp, err := c.GetNote(repo.Owner.UserName, repo.Name, sha)
	require.Error(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	// Forgejo stores the note via `git notes add`, which appends a trailing
	// newline the same way a commit message would, so the message read back
	// is never byte-identical to the one sent.
	note, _, err := c.SetNote(repo.Owner.UserName, repo.Name, sha, NoteOptions{Message: "a note"})
	require.NoError(t, err)
	require.NotNil(t, note)
	assert.Equal(t, "a note\n", note.Message)

	got, _, err := c.GetNote(repo.Owner.UserName, repo.Name, sha)
	require.NoError(t, err)
	assert.Equal(t, "a note\n", got.Message)

	_, err = c.RemoveNote(repo.Owner.UserName, repo.Name, sha)
	require.NoError(t, err)

	_, resp, err = c.GetNote(repo.Owner.UserName, repo.Name, sha)
	require.Error(t, err)
	assert.Equal(t, 404, resp.StatusCode)
}
