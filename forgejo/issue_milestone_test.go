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

func TestMilestones(t *testing.T) {
	log.Println("== TestMilestones ==")
	c := newTestClient()

	repo, _ := createTestRepo(t, "TestMilestones", c)
	now := time.Now()
	future := time.Unix(1896134400, 0) // 2030-02-01
	closed := "closed"
	sClosed := StateClosed

	// CreateMilestone 4x
	m1, _, err := c.CreateMilestone(repo.Owner.UserName, repo.Name, CreateMilestoneOption{Title: "v1.0", Description: "First Version", Deadline: &now})
	require.NoError(t, err)
	_, _, err = c.CreateMilestone(repo.Owner.UserName, repo.Name, CreateMilestoneOption{Title: "v2.0", Description: "Second Version", Deadline: &future})
	require.NoError(t, err)
	_, _, err = c.CreateMilestone(repo.Owner.UserName, repo.Name, CreateMilestoneOption{Title: "v3.0", Description: "Third Version", Deadline: nil})
	require.NoError(t, err)
	m4, _, err := c.CreateMilestone(repo.Owner.UserName, repo.Name, CreateMilestoneOption{Title: "temp", Description: "part time milestone"})
	require.NoError(t, err)

	// EditMilestone
	m1, _, err = c.EditMilestone(repo.Owner.UserName, repo.Name, m1.ID, EditMilestoneOption{Description: &closed, State: &sClosed})
	require.NoError(t, err)

	// DeleteMilestone
	_, err = c.DeleteMilestone(repo.Owner.UserName, repo.Name, m4.ID)
	require.NoError(t, err)

	// ListRepoMilestones
	ml, _, err := c.ListRepoMilestones(repo.Owner.UserName, repo.Name, ListMilestoneOption{})
	require.NoError(t, err)
	assert.Len(t, ml, 3)
	ml, _, err = c.ListRepoMilestones(repo.Owner.UserName, repo.Name, ListMilestoneOption{State: StateClosed})
	require.NoError(t, err)
	assert.Len(t, ml, 1)
	ml, _, err = c.ListRepoMilestones(repo.Owner.UserName, repo.Name, ListMilestoneOption{State: StateAll})
	require.NoError(t, err)
	assert.Len(t, ml, 3)
	ml, _, err = c.ListRepoMilestones(repo.Owner.UserName, repo.Name, ListMilestoneOption{State: StateAll, Name: "V3.0"})
	require.NoError(t, err)
	assert.Len(t, ml, 1)
	assert.Equal(t, "v3.0", ml[0].Title)

	// test fallback resolveMilestoneByName
	m, _, err := c.resolveMilestoneByName(repo.Owner.UserName, repo.Name, "V3.0")
	require.NoError(t, err)
	assert.Equal(t, ml[0].ID, m.ID)
	_, _, err = c.resolveMilestoneByName(repo.Owner.UserName, repo.Name, "NoEvidenceOfExist")
	require.Error(t, err)
	assert.Equal(t, "milestone 'NoEvidenceOfExist' do not exist", err.Error())

	// GetMilestone
	_, _, err = c.GetMilestone(repo.Owner.UserName, repo.Name, m4.ID)
	require.Error(t, err)
	m, _, err = c.GetMilestone(repo.Owner.UserName, repo.Name, m1.ID)
	require.NoError(t, err)
	assert.True(t, m1.Updated.Before(*m.Updated) || m1.Updated.Equal(*m.Updated))
	assert.True(t, m.Updated.Before(time.Now()) || m.Updated.Equal(time.Now()))
	m1.Updated = m.Updated
	assert.Equal(t, m1, m)
	m2, _, err := c.GetMilestoneByName(repo.Owner.UserName, repo.Name, m.Title)
	require.NoError(t, err)
	assert.True(t, m.Updated.Before(*m2.Updated) || m.Updated.Equal(*m2.Updated))
	assert.True(t, m2.Updated.Before(time.Now()) || m2.Updated.Equal(time.Now()))
	m.Updated = m2.Updated
	assert.Equal(t, m, m2)
}

func TestMilestoneByName_EditDelete(t *testing.T) {
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})

	_, _, err := c.CreateMilestone(repo.Owner.UserName, repo.Name, CreateMilestoneOption{Title: "v1.0"})
	require.NoError(t, err)

	newDescription := "renamed via EditMilestoneByName"
	edited, _, err := c.EditMilestoneByName(repo.Owner.UserName, repo.Name, "v1.0", EditMilestoneOption{
		Description: &newDescription,
	})
	require.NoError(t, err)
	assert.Equal(t, newDescription, edited.Description)
	assert.Equal(t, "v1.0", edited.Title)

	_, err = c.DeleteMilestoneByName(repo.Owner.UserName, repo.Name, "v1.0")
	require.NoError(t, err)

	_, _, err = c.GetMilestoneByName(repo.Owner.UserName, repo.Name, "v1.0")
	require.Error(t, err)
}

// TestMilestoneByName_EditLegacyPath exercises EditMilestoneByName's
// pre-1.13.0 branch, which resolves the name to an ID with
// resolveMilestoneByName and edits it through EditMilestone -- against the
// same real server, just with the client's cached version overridden to
// look old (see TestGetPullRequestDiff_LegacyPath for why that's enough).
func TestMilestoneByName_EditLegacyPath(t *testing.T) {
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})

	_, _, err := c.CreateMilestone(repo.Owner.UserName, repo.Name, CreateMilestoneOption{Title: "v2.0"})
	require.NoError(t, err)

	legacy, err := newTestClientOpts(SetForgejoVersion("1.12.0"))
	require.NoError(t, err)

	newDescription := "renamed via the legacy path"
	edited, _, err := legacy.EditMilestoneByName(repo.Owner.UserName, repo.Name, "v2.0", EditMilestoneOption{
		Description: &newDescription,
	})
	require.NoError(t, err)
	assert.Equal(t, newDescription, edited.Description)

	_, _, err = legacy.EditMilestoneByName(repo.Owner.UserName, repo.Name, "NoEvidenceOfExist", EditMilestoneOption{})
	require.Error(t, err, "resolveMilestoneByName itself fails, before any edit is attempted")
}
