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

func TestSyncFork(t *testing.T) {
	t.Parallel()
	log.Println("== TestSyncFork ==")
	c := newTestClient()

	origRepo, err := createTestRepo(t, "SyncForkBase", c)
	require.NoError(t, err)

	forkOrg := "SyncForkOrg"
	_, _ = c.DeleteRepo(forkOrg, origRepo.Name)
	_, _ = c.DeleteOrg(forkOrg)
	org, _, err := c.CreateOrg(CreateOrgOption{Name: forkOrg})
	require.NoError(t, err)
	forkRepo, _, err := c.CreateFork(origRepo.Owner.UserName, origRepo.Name, CreateForkOption{Organization: &org.UserName})
	require.NoError(t, err)
	require.NotNil(t, forkRepo)

	// Confirmed absent on a live 11.0.16 instance and present by 15.0.9;
	// below that the SDK's guard refuses the call, which is the documented
	// behavior and worth asserting.
	if !serverAtLeast(t, c, "15.0.0") {
		_, _, err := c.GetSyncForkDefaultInfo(forkRepo.Owner.UserName, forkRepo.Name)
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
		return
	}

	// nothing to sync right after forking
	info, _, err := c.GetSyncForkDefaultInfo(forkRepo.Owner.UserName, forkRepo.Name)
	require.NoError(t, err)
	assert.False(t, info.Allowed)

	branchInfo, _, err := c.GetSyncForkBranchInfo(forkRepo.Owner.UserName, forkRepo.Name, forkRepo.DefaultBranch)
	require.NoError(t, err)
	assert.False(t, branchInfo.Allowed)

	// calling GetSyncForkDefaultInfo/GetSyncForkBranchInfo on a repo that
	// isn't a fork at all must fail
	_, resp, err := c.GetSyncForkDefaultInfo(origRepo.Owner.UserName, origRepo.Name)
	require.Error(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// advance the base repo so that a sync becomes possible
	license, _, err := c.GetContents(origRepo.Owner.UserName, origRepo.Name, origRepo.DefaultBranch, "LICENSE")
	require.NoError(t, err)
	_, _, err = c.UpdateFile(origRepo.Owner.UserName, origRepo.Name, "LICENSE", UpdateFileOptions{
		FileOptions: FileOptions{
			Message:    "advance base for sync",
			BranchName: origRepo.DefaultBranch,
		},
		SHA:     license.SHA,
		Content: "U3luYyBGb3JrIFRlc3QK",
	})
	require.NoError(t, err)

	info, _, err = c.GetSyncForkDefaultInfo(forkRepo.Owner.UserName, forkRepo.Name)
	require.NoError(t, err)
	assert.True(t, info.Allowed)

	_, err = c.SyncForkDefault(forkRepo.Owner.UserName, forkRepo.Name)
	require.NoError(t, err)

	_, err = c.SyncForkBranch(forkRepo.Owner.UserName, forkRepo.Name, forkRepo.DefaultBranch)
	require.Error(t, err)
}
