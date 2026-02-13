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

// TestLabels test label related func
func TestLabels(t *testing.T) {
	log.Println("== TestLabels ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "LabelTestsRepo", c)
	require.NoError(t, err)

	createOpts := CreateLabelOption{
		Name:        " ",
		Description: "",
		Color:       "",
	}
	err = createOpts.Validate()
	require.Error(t, err)
	assert.Equal(t, "invalid color format", err.Error())
	createOpts.Color = "12345f"
	err = createOpts.Validate()
	require.Error(t, err)
	assert.Equal(t, "empty name not allowed", err.Error())
	createOpts.Name = "label one"

	labelOne, _, err := c.CreateLabel(repo.Owner.UserName, repo.Name, createOpts)
	require.NoError(t, err)
	assert.Equal(t, createOpts.Name, labelOne.Name)
	assert.Equal(t, createOpts.Color, labelOne.Color)

	labelTwo, _, err := c.CreateLabel(repo.Owner.UserName, repo.Name, CreateLabelOption{
		Name:        "blue",
		Color:       "#0000FF",
		Description: "CMYB(100%, 100%, 0%, 0%)",
	})
	require.NoError(t, err)
	_, _, err = c.CreateLabel(repo.Owner.UserName, repo.Name, CreateLabelOption{
		Name:        "gray",
		Color:       "808080",
		Description: "CMYB(0%, 0%, 0%, 50%)",
	})
	require.NoError(t, err)
	_, _, err = c.CreateLabel(repo.Owner.UserName, repo.Name, CreateLabelOption{
		Name:        "green",
		Color:       "#98F76C",
		Description: "CMYB(38%, 0%, 56%, 3%)",
	})
	require.NoError(t, err)

	labels, resp, err := c.ListRepoLabels(repo.Owner.UserName, repo.Name, ListLabelsOptions{ListOptions: ListOptions{PageSize: 3}})
	require.NoError(t, err)
	assert.Len(t, labels, 3)
	assert.NotNil(t, resp)
	assert.Contains(t, labels, labelTwo)
	assert.NotContains(t, labels, labelOne)

	label, _, err := c.GetRepoLabel(repo.Owner.UserName, repo.Name, labelTwo.ID)
	require.NoError(t, err)
	assert.Equal(t, labelTwo, label)

	label, _, err = c.EditLabel(repo.Owner.UserName, repo.Name, labelTwo.ID, EditLabelOption{
		Color:       OptionalString("#0e0175"),
		Description: OptionalString("blueish"),
	})
	require.NoError(t, err)
	assert.Equal(t, &Label{
		ID:          labelTwo.ID,
		Name:        labelTwo.Name,
		Color:       "0e0175",
		Description: "blueish",
		URL:         labelTwo.URL,
	}, label)
	labels, _, _ = c.ListRepoLabels(repo.Owner.UserName, repo.Name, ListLabelsOptions{ListOptions: ListOptions{PageSize: 3}})

	createTestIssue(t, c, repo.Name, "test-issue", "", nil, nil, 0, []int64{label.ID}, false, false)
	issueIndex := int64(1)

	issueLabels, _, err := c.GetIssueLabels(repo.Owner.UserName, repo.Name, issueIndex, ListLabelsOptions{})
	require.NoError(t, err)
	assert.Len(t, issueLabels, 1)
	assert.Equal(t, label, issueLabels[0])

	_, _, err = c.AddIssueLabels(repo.Owner.UserName, repo.Name, issueIndex, IssueLabelsOption{Labels: []int64{labels[0].ID}})
	require.NoError(t, err)

	issueLabels, _, err = c.AddIssueLabels(repo.Owner.UserName, repo.Name, issueIndex, IssueLabelsOption{Labels: []int64{labels[1].ID, labels[2].ID}})
	require.NoError(t, err)
	assert.Len(t, issueLabels, 3)
	assert.Equal(t, labels, issueLabels)

	labels, _, _ = c.ListRepoLabels(repo.Owner.UserName, repo.Name, ListLabelsOptions{})
	assert.Len(t, labels, 11)

	issueLabels, _, err = c.ReplaceIssueLabels(repo.Owner.UserName, repo.Name, issueIndex, IssueLabelsOption{Labels: []int64{labels[0].ID, labels[1].ID}})
	require.NoError(t, err)
	assert.Len(t, issueLabels, 2)

	_, err = c.DeleteIssueLabel(repo.Owner.UserName, repo.Name, issueIndex, labels[0].ID)
	require.NoError(t, err)
	issueLabels, _, _ = c.GetIssueLabels(repo.Owner.UserName, repo.Name, issueIndex, ListLabelsOptions{})
	assert.Len(t, issueLabels, 1)

	_, err = c.ClearIssueLabels(repo.Owner.UserName, repo.Name, issueIndex)
	require.NoError(t, err)
	issueLabels, _, _ = c.GetIssueLabels(repo.Owner.UserName, repo.Name, issueIndex, ListLabelsOptions{})
	assert.Empty(t, issueLabels)

	_, err = c.DeleteLabel(repo.Owner.UserName, repo.Name, labelTwo.ID)
	require.NoError(t, err)
	labels, _, _ = c.ListRepoLabels(repo.Owner.UserName, repo.Name, ListLabelsOptions{})
	assert.Len(t, labels, 10)
}
