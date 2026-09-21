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

func TestCommitStatus(t *testing.T) {
	log.Println("== TestCommitStatus ==")
	c := newTestClient()
	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	repoName := "CommitStatuses"
	origRepo, err := createTestRepo(t, repoName, c)
	if !assert.NoError(t, err) { //nolint:testifylint
		return
	}

	commits, _, _ := c.ListRepoCommits(user.UserName, repoName, ListCommitOptions{
		ListOptions: ListOptions{},
		SHA:         origRepo.DefaultBranch,
	})
	if !assert.Len(t, commits, 1) {
		return
	}
	sha := commits[0].SHA

	combiStats, resp, err := c.GetCombinedStatus(user.UserName, repoName, sha)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, combiStats)
	assert.Equal(t, 0, combiStats.TotalCount)

	statuses, resp, err := c.ListStatuses(user.UserName, repoName, sha, ListStatusesOption{})
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, statuses)
	assert.Empty(t, statuses)

	createStatus(t, c, user.UserName, repoName, sha, "http://dummy.test", "start testing", "ultraCI", StatusPending)
	createStatus(t, c, user.UserName, repoName, sha, "https://more.secure", "just a warning", "warn/bot", StatusWarning)
	createStatus(t, c, user.UserName, repoName, sha, "http://dummy.test", "test failed", "ultraCI", StatusFailure)
	createStatus(t, c, user.UserName, repoName, sha, "http://dummy.test", "start testing", "ultraCI", StatusPending)
	createStatus(t, c, user.UserName, repoName, sha, "http://dummy.test", "test passed", "ultraCI", StatusSuccess)

	statuses, resp, err = c.ListStatuses(user.UserName, repoName, sha, ListStatusesOption{})
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, statuses)
	assert.Len(t, statuses, 5)

	combiStats, resp, err = c.GetCombinedStatus(user.UserName, repoName, sha)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, combiStats)
	assert.Equal(t, 2, combiStats.TotalCount)
	assert.Equal(t, StatusState("warning"), combiStats.State)
	assert.Len(t, combiStats.Statuses, 2)
}

func TestListCommitStatuses(t *testing.T) {
	log.Println("== TestListCommitStatuses ==")
	c := newTestClient()
	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	repoName := "CommitStatusesBySHA"
	origRepo, err := createTestRepo(t, repoName, c)
	if !assert.NoError(t, err) { //nolint:testifylint
		return
	}

	commits, _, _ := c.ListRepoCommits(user.UserName, repoName, ListCommitOptions{
		ListOptions: ListOptions{},
		SHA:         origRepo.DefaultBranch,
	})
	if !assert.Len(t, commits, 1) {
		return
	}
	sha := commits[0].SHA

	statuses, resp, err := c.ListCommitStatuses(user.UserName, repoName, sha, ListCommitStatusesOptions{})
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, statuses)

	createStatus(t, c, user.UserName, repoName, sha, "http://dummy.test", "start testing", "ultraCI", StatusPending)
	createStatus(t, c, user.UserName, repoName, sha, "http://dummy.test", "test passed", "ultraCI", StatusSuccess)

	statuses, resp, err = c.ListCommitStatuses(user.UserName, repoName, sha, ListCommitStatusesOptions{State: StatusSuccess})
	require.NoError(t, err)
	assert.NotNil(t, resp)
	if assert.Len(t, statuses, 1) {
		assert.Equal(t, StatusSuccess, statuses[0].State)
	}
}

func createStatus(t *testing.T, c *Client, userName, repoName, sha, url, desc, context string, state StatusState) { //nolint
	stats, resp, err := c.CreateStatus(userName, repoName, sha, CreateStatusOption{
		State:       state,
		TargetURL:   url,
		Description: desc,
		Context:     context,
	})
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, stats)
	assert.Equal(t, state, stats.State)
}
