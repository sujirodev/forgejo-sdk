// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueStopwatch_StartStop(t *testing.T) {
	c := newTestClient()
	owner := createTestUser(t, uniqueName(t, "swowner"), c)
	c.SetSudo(owner.UserName)
	t.Cleanup(func() { c.SetSudo("") })
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})
	issue := createTestIssue(t, c, repo.Name, "Stopwatch target", "", nil, nil, 0, nil, false, false)

	_, err := c.StartIssueStopWatch(repo.Owner.UserName, repo.Name, issue.Index)
	require.NoError(t, err)

	watches, _, err := c.GetMyStopwatches()
	require.NoError(t, err)
	require.Len(t, watches, 1)
	assert.Equal(t, issue.Index, watches[0].IssueIndex)
	assert.Equal(t, repo.Name, watches[0].RepoName)

	_, err = c.StopIssueStopWatch(repo.Owner.UserName, repo.Name, issue.Index)
	require.NoError(t, err)

	watches, _, err = c.GetMyStopwatches()
	require.NoError(t, err)
	assert.Empty(t, watches)

	// Stopping the stopwatch converts it into tracked time; confirm via the
	// tracked-time API (covered fully in issue_tracked_time_test.go).
	times, _, err := c.ListIssueTrackedTimes(repo.Owner.UserName, repo.Name, issue.Index, ListTrackedTimesOptions{})
	require.NoError(t, err)
	assert.Len(t, times, 1)
}

func TestIssueStopwatch_Delete(t *testing.T) {
	c := newTestClient()
	owner := createTestUser(t, uniqueName(t, "swowner"), c)
	c.SetSudo(owner.UserName)
	t.Cleanup(func() { c.SetSudo("") })
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})
	issue := createTestIssue(t, c, repo.Name, "Stopwatch delete target", "", nil, nil, 0, nil, false, false)

	_, err := c.StartIssueStopWatch(repo.Owner.UserName, repo.Name, issue.Index)
	require.NoError(t, err)

	watches, _, err := c.GetMyStopwatches()
	require.NoError(t, err)
	require.Len(t, watches, 1)

	_, err = c.DeleteIssueStopwatch(repo.Owner.UserName, repo.Name, issue.Index)
	require.NoError(t, err)

	watches, _, err = c.GetMyStopwatches()
	require.NoError(t, err)
	assert.Empty(t, watches)

	// Deleting (as opposed to stopping) the stopwatch discards the time.
	times, _, err := c.ListIssueTrackedTimes(repo.Owner.UserName, repo.Name, issue.Index, ListTrackedTimesOptions{})
	require.NoError(t, err)
	assert.Empty(t, times)
}
