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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIssue is main func witch call all Tests for Issue API
// (to make sure they are on correct order)
func TestIssue(t *testing.T) {
	t.Parallel()
	c := newTestClient()

	createIssue(t, c)
	// Forgejo stores/indexes the issue asynchronously; poll instead of
	// sleeping a fixed amount, which stops being enough under load.
	eventually(t, func() bool {
		issues, _, err := c.ListRepoIssues("test01", "IssueTestsRepo", ListIssueOption{State: StateOpen})
		return err == nil && len(issues) > 0
	})
	editIssues(t, c)
	listIssues(t, c)
	deleteIssue(t, c)
}

func createIssue(t *testing.T, c *Client) {
	log.Println("== TestCreateIssues ==")

	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)
	repo, _ := createTestRepo(t, "IssueTestsRepo", c)

	nowTime := time.Now()
	mile, _, _ := c.CreateMilestone(user.UserName, repo.Name, CreateMilestoneOption{Title: "mile1"})
	label1, _, _ := c.CreateLabel(user.UserName, repo.Name, CreateLabelOption{Name: "Label1", Description: "a", Color: "#ee0701"})
	label2, _, _ := c.CreateLabel(user.UserName, repo.Name, CreateLabelOption{Name: "Label2", Description: "b", Color: "#128a0c"})

	createTestIssue(t, c, repo.Name, "First Issue", "", nil, nil, 0, nil, false, false)
	createTestIssue(t, c, repo.Name, "Issue 2", "closed isn't it?", nil, nil, 0, nil, true, false)
	createTestIssue(t, c, repo.Name, "Issue 3", "", nil, nil, 0, nil, true, false)
	createTestIssue(t, c, repo.Name, "Feature: spam protect 4", "explain explain explain", []string{user.UserName}, &nowTime, 0, nil, true, false)
	createTestIssue(t, c, repo.Name, "W 123", "", nil, &nowTime, mile.ID, nil, false, false)
	createTestIssue(t, c, repo.Name, "First Issue", "", nil, nil, 0, nil, false, false)
	createTestIssue(t, c, repo.Name, "Do it soon!", "is important!", []string{user.UserName}, &nowTime, mile.ID, []int64{label1.ID, label2.ID}, false, false)
	createTestIssue(t, c, repo.Name, "Job Done", "you never know", nil, nil, mile.ID, []int64{label2.ID}, true, false)
	createTestIssue(t, c, repo.Name, "", "you never know", nil, nil, mile.ID, nil, true, true)
}

func deleteIssue(t *testing.T, c *Client) {
	log.Println("== TestDeleteIssues ==")

	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)
	repo, _ := createTestRepo(t, "IssueTestsRepo", c)

	issue := createTestIssue(t, c, repo.Name, "Deleteable Issue", "", nil, nil, 0, nil, false, false)
	_, err = c.DeleteIssue(user.UserName, repo.Name, issue.Index)
	require.NoError(t, err)
}

func editIssues(t *testing.T, c *Client) {
	log.Println("== TestEditIssues ==")
	il, _, err := c.ListIssues(ListIssueOption{KeyWord: "soon!"})
	require.NoError(t, err)
	issue, _, err := c.GetIssue(il[0].Poster.UserName, il[0].Repository.Name, il[0].Index)
	require.NoError(t, err)

	state := StateClosed
	issueNew, _, err := c.EditIssue(issue.Poster.UserName, issue.Repository.Name, issue.Index, EditIssueOption{
		Title: "Edited",
		Body:  OptionalString("123 test and go"),
		State: &state,
		Ref:   OptionalString("main"),
	})
	require.NoError(t, err)
	assert.Equal(t, issue.ID, issueNew.ID)
	assert.Equal(t, "123 test and go", issueNew.Body)
	assert.Equal(t, "Edited", issueNew.Title)
	assert.Equal(t, "main", issueNew.Ref)
}

func listIssues(t *testing.T, c *Client) {
	log.Println("== TestListIssues ==")

	issues, _, err := c.ListRepoIssues("test01", "IssueTestsRepo", ListIssueOption{
		Labels:  []string{"Label1", "Label2"},
		KeyWord: "",
		State:   "all",
	})
	require.NoError(t, err)
	assert.Len(t, issues, 1)

	issues, _, err = c.ListIssues(ListIssueOption{
		Labels:  []string{"Label2"},
		KeyWord: "Done",
		State:   "all",
	})
	require.NoError(t, err)
	assert.Len(t, issues, 1)

	issues, _, err = c.ListRepoIssues("test01", "IssueTestsRepo", ListIssueOption{
		Milestones: []string{"mile1"},
		State:      "all",
	})
	require.NoError(t, err)
	assert.Len(t, issues, 3)
	for i := range issues {
		if assert.NotNil(t, issues[i].Milestone) {
			assert.Equal(t, "mile1", issues[i].Milestone.Title)
		}
	}

	issues, _, err = c.ListRepoIssues("test01", "IssueTestsRepo", ListIssueOption{})
	require.NoError(t, err)
	assert.Len(t, issues, 3)
}

func createTestIssue(t *testing.T, c *Client, repoName, title, body string, assignees []string, deadline *time.Time, milestone int64, labels []int64, closed, shouldFail bool) *Issue {
	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)
	issue, _, e := c.CreateIssue(user.UserName, repoName, CreateIssueOption{
		Title:     title,
		Body:      body,
		Assignees: assignees,
		Deadline:  deadline,
		Milestone: milestone,
		Labels:    labels,
		Closed:    closed,
	})
	if shouldFail {
		require.Error(t, e)
		return nil
	}
	require.NoError(t, e)
	assert.NotEmpty(t, issue)
	assert.Equal(t, title, issue.Title)
	assert.Equal(t, body, issue.Body)
	assert.Len(t, assignees, len(issue.Assignees))
	for i, a := range issue.Assignees {
		assert.Equal(t, assignees[i], a.UserName)
	}
	if milestone > 0 {
		assert.Equal(t, milestone, issue.Milestone.ID)
	}
	assert.Len(t, labels, len(issue.Labels))
	if closed {
		assert.False(t, issue.Closed.IsZero())
	} else {
		assert.Empty(t, issue.Closed)
	}
	return issue
}
