// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssueTrackedTime(t *testing.T) {
	t.Parallel()
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})
	issue := createTestIssue(t, c, repo.Name, "Tracked time target", "", nil, nil, 0, nil, false, false)

	tt, _, err := c.AddTime(repo.Owner.UserName, repo.Name, issue.Index, AddTimeOption{Time: 120})
	require.NoError(t, err)
	assert.Equal(t, int64(120), tt.Time)
	assert.Equal(t, "test01", tt.UserName)

	issueTimes, _, err := c.ListIssueTrackedTimes(repo.Owner.UserName, repo.Name, issue.Index, ListTrackedTimesOptions{})
	require.NoError(t, err)
	require.Len(t, issueTimes, 1)
	assert.Equal(t, tt.ID, issueTimes[0].ID)

	repoTimes, _, err := c.ListRepoTrackedTimes(repo.Owner.UserName, repo.Name, ListTrackedTimesOptions{})
	require.NoError(t, err)
	require.Len(t, repoTimes, 1)
	assert.Equal(t, tt.ID, repoTimes[0].ID)

	myTimes, _, err := c.GetMyTrackedTimes()
	require.NoError(t, err)
	found := false
	for _, m := range myTimes {
		if m.ID == tt.ID {
			found = true
		}
	}
	assert.True(t, found, "GetMyTrackedTimes should include the time entry just added")

	tt2, _, err := c.AddTime(repo.Owner.UserName, repo.Name, issue.Index, AddTimeOption{Time: 60})
	require.NoError(t, err)

	_, err = c.DeleteTime(repo.Owner.UserName, repo.Name, issue.Index, tt2.ID)
	require.NoError(t, err)
	issueTimes, _, err = c.ListIssueTrackedTimes(repo.Owner.UserName, repo.Name, issue.Index, ListTrackedTimesOptions{})
	require.NoError(t, err)
	assert.Len(t, issueTimes, 1)

	_, err = c.ResetIssueTime(repo.Owner.UserName, repo.Name, issue.Index)
	require.NoError(t, err)
	issueTimes, _, err = c.ListIssueTrackedTimes(repo.Owner.UserName, repo.Name, issue.Index, ListTrackedTimesOptions{})
	require.NoError(t, err)
	assert.Empty(t, issueTimes)
}
