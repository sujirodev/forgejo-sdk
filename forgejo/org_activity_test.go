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

func TestListOrgActivityFeeds(t *testing.T) {
	log.Println("== TestListOrgActivityFeeds ==")
	c := newTestClient()

	orgName := "OrgActivityTestOrg"
	cleanup, repo, err := createTestOrgRepo(t, c, orgName)
	require.NoError(t, err)
	defer cleanup()
	require.NotNil(t, repo)

	activities, resp, err := c.ListOrgActivityFeeds(t.Context(), orgName, ListActivityFeedsOptions{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotNil(t, activities)
}

func TestListTeamActivityFeeds(t *testing.T) {
	log.Println("== TestListTeamActivityFeeds ==")
	c := newTestClient()

	orgName := "TeamActivityTestOrg"
	cleanup, repo, err := createTestOrgRepo(t, c, orgName)
	require.NoError(t, err)
	defer cleanup()
	require.NotNil(t, repo)

	team, err := createTestOrgTeams(t, c, orgName, "activity-team", AccessModeRead, map[string]string{RepoUnitCode.String(): string(AccessModeRead)})
	require.NoError(t, err)

	activities, resp, err := c.ListTeamActivityFeeds(t.Context(), team.ID, ListActivityFeedsOptions{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotNil(t, activities)
}
