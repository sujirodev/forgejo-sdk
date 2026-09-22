// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullReview(t *testing.T) {
	log.Println("== TestPullReview ==")
	c := newTestClient()

	repoName := "Reviews"
	repo, pull, submitter, reviewer, success := preparePullReviewTest(t, c, repoName)
	if !success {
		return
	}

	// CreatePullReview
	r1, _, err := c.CreatePullReview(repo.Owner.UserName, repo.Name, pull.Index, CreatePullReviewOptions{
		State: ReviewStateComment,
		Body:  "I'll have a look at it later",
	})
	require.NoError(t, err)
	if assert.NotNil(t, r1) {
		assert.Equal(t, ReviewStateComment, r1.State)
		assert.Equal(t, int64(1), r1.Reviewer.ID)
	}

	c.SetSudo(submitter.UserName)
	_, _, err = c.CreatePullReview(repo.Owner.UserName, repo.Name, pull.Index, CreatePullReviewOptions{
		State: ReviewStateApproved,
		Body:  "lgtm it myself",
	})
	require.Error(t, err)
	r2, _, err := c.CreatePullReview(repo.Owner.UserName, repo.Name, pull.Index, CreatePullReviewOptions{
		State: ReviewStateComment,
		Body:  "no seriously please have a look at it",
	})
	require.NoError(t, err)
	assert.NotNil(t, r2)

	c.SetSudo(reviewer.UserName)
	r3, _, err := c.CreatePullReview(repo.Owner.UserName, repo.Name, pull.Index, CreatePullReviewOptions{
		State: ReviewStateApproved,
		Body:  "lgtm",
		Comments: []CreatePullReviewComment{
			{
				Path:       "WOW-file",
				Body:       "no better name - really?",
				NewLineNum: 1,
			},
		},
	})
	require.NoError(t, err)
	assert.NotNil(t, r3)

	// ListPullReviews
	c.SetSudo("")
	rl, _, err := c.ListPullReviews(repo.Owner.UserName, repo.Name, pull.Index, ListPullReviewsOptions{})
	require.NoError(t, err)
	assert.Len(t, rl, 3)
	for i := range rl {
		assert.Equal(t, pull.HTMLURL, rl[i].HTMLPullURL)
		if rl[i].CodeCommentsCount == 1 {
			assert.Equal(t, reviewer.ID, rl[i].Reviewer.ID)
		}
	}

	// GetPullReview
	rNew, _, err := c.GetPullReview(repo.Owner.UserName, repo.Name, pull.Index, r3.ID)
	require.NoError(t, err)
	assert.Equal(t, r3, rNew)

	// DeletePullReview
	c.SetSudo(submitter.UserName)
	_, err = c.DeletePullReview(repo.Owner.UserName, repo.Name, pull.Index, r2.ID)
	require.NoError(t, err)
	_, err = c.DeletePullReview(repo.Owner.UserName, repo.Name, pull.Index, r3.ID)
	require.Error(t, err)

	// SubmitPullReview
	c.SetSudo("")
	r4, resp, err := c.CreatePullReview(repo.Owner.UserName, repo.Name, pull.Index, CreatePullReviewOptions{
		Body: "...",
		Comments: []CreatePullReviewComment{
			{
				Path:       "WOW-file",
				Body:       "its ok",
				NewLineNum: 1,
			},
		},
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)
	r5, _, err := c.CreatePullReview(repo.Owner.UserName, repo.Name, pull.Index, CreatePullReviewOptions{
		Body: "...",
		Comments: []CreatePullReviewComment{
			{
				Path:       "WOW-file",
				Body:       "hehe and here it is",
				NewLineNum: 3,
			},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, r4.ID, r5.ID)

	r, _, err := c.SubmitPullReview(repo.Owner.UserName, repo.Name, pull.Index, r4.ID, SubmitPullReviewOptions{
		State: ReviewStateRequestChanges,
		Body:  "one nit",
	})
	require.NoError(t, err)
	assert.Equal(t, r4.ID, r.ID)
	assert.Equal(t, ReviewStateRequestChanges, r.State)

	// ListPullReviewComments
	rcl, _, err := c.ListPullReviewComments(repo.Owner.UserName, repo.Name, pull.Index, r.ID)
	require.NoError(t, err)
	assert.Len(t, rcl, r.CodeCommentsCount)
	for _, rc := range rcl {
		assert.Equal(t, pull.HTMLURL, rc.HTMLPullURL)
		if rc.LineNum == 3 {
			assert.Equal(t, "hehe and here it is", rc.Body)
		} else {
			assert.Equal(t, uint64(1), rc.LineNum)
			assert.Equal(t, "its ok", rc.Body)
		}
	}

	r, _, err = c.GetPullReview(repo.Owner.UserName, repo.Name, pull.Index, r.ID)
	require.NoError(t, err)
	assert.False(t, r.Dismissed)

	// DismissPullReview
	resp, err = c.DismissPullReview(repo.Owner.UserName, repo.Name, pull.Index, r.ID, DismissPullReviewOptions{Message: "stale"})
	require.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, 200, resp.StatusCode)
	}
	r, _, _ = c.GetPullReview(repo.Owner.UserName, repo.Name, pull.Index, r.ID)
	assert.True(t, r.Dismissed)

	// UnDismissPullReview
	resp, err = c.UnDismissPullReview(repo.Owner.UserName, repo.Name, pull.Index, r.ID)
	require.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, 200, resp.StatusCode)
	}
	r, _, _ = c.GetPullReview(repo.Owner.UserName, repo.Name, pull.Index, r.ID)
	assert.False(t, r.Dismissed)

	rl, _, err = c.ListPullReviews(repo.Owner.UserName, repo.Name, pull.Index, ListPullReviewsOptions{})
	require.NoError(t, err)
	assert.Len(t, rl, 3)

	c.SetSudo(submitter.UserName)
	resp, err = c.CreateReviewRequests(repo.Owner.UserName, repo.Name, pull.Index, PullReviewRequestOptions{Reviewers: []string{reviewer.UserName}})
	require.NoError(t, err)
	assert.NotNil(t, resp)

	rl, _, _ = c.ListPullReviews(repo.Owner.UserName, repo.Name, pull.Index, ListPullReviewsOptions{})
	if assert.Len(t, rl, 4) {
		assert.Equal(t, ReviewStateRequestReview, rl[3].State)
	}

	c.SetSudo(reviewer.UserName)
	resp, err = c.DeleteReviewRequests(repo.Owner.UserName, repo.Name, pull.Index, PullReviewRequestOptions{Reviewers: []string{reviewer.UserName}})
	require.NoError(t, err)
	assert.NotNil(t, resp)

	rl, _, _ = c.ListPullReviews(repo.Owner.UserName, repo.Name, pull.Index, ListPullReviewsOptions{})
	assert.Len(t, rl, 3)

	c.SetSudo("")
	_, err = c.AdminDeleteUser(reviewer.UserName)
	require.NoError(t, err)
	_, err = c.AdminDeleteUser(submitter.UserName)
	require.NoError(t, err)
}

func TestPullReviewComments(t *testing.T) {
	log.Println("== TestPullReviewComments ==")
	c := newTestClient()

	repoName := "ReviewComments"
	repo, pull, submitter, reviewer, success := preparePullReviewTest(t, c, repoName)
	if !success {
		return
	}

	r, _, err := c.CreatePullReview(repo.Owner.UserName, repo.Name, pull.Index, CreatePullReviewOptions{
		State: ReviewStateComment,
		Body:  "let's discuss this",
	})
	require.NoError(t, err)
	require.NotNil(t, r)

	rc, _, err := c.CreatePullReviewComment(repo.Owner.UserName, repo.Name, pull.Index, r.ID, CreatePullReviewCommentOptions{
		Path:       "WOW-file",
		Body:       "why this name?",
		NewLineNum: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, rc)
	assert.Equal(t, "why this name?", rc.Body)

	got, _, err := c.GetPullReviewComment(repo.Owner.UserName, repo.Name, pull.Index, r.ID, rc.ID)
	require.NoError(t, err)
	assert.Equal(t, rc.ID, got.ID)
	assert.Equal(t, rc.Body, got.Body)

	_, err = c.DeletePullReviewComment(repo.Owner.UserName, repo.Name, pull.Index, r.ID, rc.ID)
	require.NoError(t, err)

	_, _, err = c.GetPullReviewComment(repo.Owner.UserName, repo.Name, pull.Index, r.ID, rc.ID)
	require.Error(t, err)

	_, err = c.AdminDeleteUser(reviewer.UserName)
	require.NoError(t, err)
	_, err = c.AdminDeleteUser(submitter.UserName)
	require.NoError(t, err)
}

func preparePullReviewTest(t *testing.T, c *Client, repoName string) (*Repository, *PullRequest, *User, *User, bool) {
	repo, err := createTestRepo(t, repoName, c)
	if !assert.NoError(t, err) { //nolint
		return nil, nil, nil, nil, false
	}

	pullSubmitter := createTestUser(t, uniqueName(t, "pullsub"), c)
	write := AccessModeWrite
	_, err = c.AddCollaborator(repo.Owner.UserName, repo.Name, pullSubmitter.UserName, AddCollaboratorOption{
		Permission: &write,
	})
	require.NoError(t, err)

	c.SetSudo(pullSubmitter.UserName)

	newFile, _, err := c.CreateFile(repo.Owner.UserName, repo.Name, "WOW-file", CreateFileOptions{
		Content: "QSBuZXcgRmlsZQoKYW5kIHNvbWUgbGluZXMK",
		FileOptions: FileOptions{
			Message:       "creat a new file",
			BranchName:    "main",
			NewBranchName: "new_file",
		},
	})

	if !assert.NoError(t, err) || !assert.NotNil(t, newFile) { //nolint
		return nil, nil, nil, nil, false
	}

	pull, _, err := c.CreatePullRequest(c.username, repoName, CreatePullRequestOption{
		Base:  "main",
		Head:  "new_file",
		Title: "Creat a NewFile",
	})
	require.NoError(t, err)
	assert.NotNil(t, pull)

	c.SetSudo("")

	reviewer := createTestUser(t, uniqueName(t, "pullrev"), c)
	admin := AccessModeAdmin
	_, err = c.AddCollaborator(repo.Owner.UserName, repo.Name, pullSubmitter.UserName, AddCollaboratorOption{
		Permission: &admin,
	})
	require.NoError(t, err)

	return repo, pull, pullSubmitter, reviewer, pull.Poster.ID == pullSubmitter.ID
}
