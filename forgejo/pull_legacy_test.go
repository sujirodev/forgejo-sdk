// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

// TestGetPullRequestDiff_LegacyPath exercises getPullRequestDiffOrPatch's
// pre-1.13.0 branch, which fetches the diff/patch from the web UI route
// instead of the API's dedicated one. It does this against the same real
// server as every other integration test, just with the *client's* cached
// version overridden to look old -- the same technique version_test.go uses
// for TestCheckServerVersionConstraint. The server itself never needs to
// actually be an old Forgejo.
import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPullRequestDiff_LegacyPath(t *testing.T) {
	t.Parallel()
	log.Println("== TestGetPullRequestDiff_LegacyPath ==")
	c := newTestClient()

	repoName := uniqueName(t, "pull_legacy")
	repo, err := createTestRepo(t, repoName, c)
	require.NoError(t, err)

	main, _, err := c.GetContents(repo.Owner.UserName, repo.Name, "main", "README.md")
	require.NoError(t, err)

	_, _, err = c.UpdateFile(repo.Owner.UserName, repo.Name, "README.md", UpdateFileOptions{
		FileOptions: FileOptions{
			Message:       "change",
			BranchName:    "main",
			NewBranchName: "legacy-diff-branch",
		},
		SHA:     main.SHA,
		Content: "bGVnYWN5IGRpZmYgdGVzdA==", // "legacy diff test"
	})
	require.NoError(t, err)

	pr, _, err := c.CreatePullRequest(repo.Owner.UserName, repo.Name, CreatePullRequestOption{
		Base:  "main",
		Head:  "legacy-diff-branch",
		Title: "legacy diff test",
	})
	require.NoError(t, err)

	legacy, err := newTestClientOpts(SetForgejoVersion("1.12.0"))
	require.NoError(t, err)

	diff, _, err := legacy.GetPullRequestDiff(repo.Owner.UserName, repo.Name, pr.Index, PullRequestDiffOptions{})
	require.NoError(t, err)
	assert.NotEmpty(t, diff)

	patch, _, err := legacy.GetPullRequestPatch(repo.Owner.UserName, repo.Name, pr.Index)
	require.NoError(t, err)
	assert.NotEmpty(t, patch)
}

// TestFixPullHeadSha_Resolves exercises fixPullHeadSha's resolving branch:
// pr.Head.Sha comes back empty and pr.Head.Ref has to be resolved through
// the base repo's refs instead (the real-world trigger is gitea issue
// #12675: the head branch got deleted and Forgejo repoints Ref into the
// base repo). fixPullHeadSha forwards Ref to GetRepoRefs completely
// unmodified, so this proves the resolving mechanics work for any ref
// GetRepoRefs itself can resolve -- confirmed with plain curl to need the
// "heads/" qualifier a bare branch name doesn't have.
func TestFixPullHeadSha_Resolves(t *testing.T) {
	t.Parallel()
	log.Println("== TestFixPullHeadSha_Resolves ==")
	c := newTestClient()

	repoName := uniqueName(t, "fix_head_sha")
	repo, err := createTestRepo(t, repoName, c)
	require.NoError(t, err)

	main, _, err := c.GetContents(repo.Owner.UserName, repo.Name, "main", "README.md")
	require.NoError(t, err)
	_, _, err = c.UpdateFile(repo.Owner.UserName, repo.Name, "README.md", UpdateFileOptions{
		FileOptions: FileOptions{
			Message:       "change",
			BranchName:    "main",
			NewBranchName: "fix-head-sha-branch",
		},
		SHA:     main.SHA,
		Content: "Zml4IGhlYWQgc2hhIHRlc3Q=", // "fix head sha test"
	})
	require.NoError(t, err)

	refs, _, err := c.GetRepoRefs(repo.Owner.UserName, repo.Name, "heads/fix-head-sha-branch")
	require.NoError(t, err)
	require.NotEmpty(t, refs)

	pr := &PullRequest{
		Base: &PRBranchInfo{Repository: repo},
		Head: &PRBranchInfo{Ref: "heads/fix-head-sha-branch", Sha: ""},
	}
	require.NoError(t, fixPullHeadSha(c, pr))
	assert.Equal(t, refs[0].Object.SHA, pr.Head.Sha)
}
