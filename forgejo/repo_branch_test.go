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

func TestRepoBranches(t *testing.T) {
	log.Println("== TestRepoBranches ==")
	c := newTestClient()
	repoName := "branches"

	repo := prepareBranchTest(t, c, repoName)
	if repo == nil {
		return
	}
	time.Sleep(1 * time.Second)
	bl, _, err := c.ListRepoBranches(repo.Owner.UserName, repo.Name, ListRepoBranchesOptions{})
	require.NoError(t, err)
	assert.Len(t, bl, 3)

	branchNames := make([]string, len(bl))
	branches := make(map[string]Branch, len(bl))
	for index, branch := range bl {
		branchNames[index] = branch.Name
		branches[branch.Name] = *branch
	}
	assert.ElementsMatch(t, []string{"feature", "main", "update"}, branchNames)

	b, _, err := c.GetRepoBranch(repo.Owner.UserName, repo.Name, "update")
	require.NoError(t, err)
	assert.Equal(t, branches["update"].Commit.ID, b.Commit.ID)
	assert.Equal(t, branches["update"].Commit.Added, b.Commit.Added)

	s, _, err := c.DeleteRepoBranch(repo.Owner.UserName, repo.Name, "main")
	require.NoError(t, err)
	assert.False(t, s)
	s, _, err = c.DeleteRepoBranch(repo.Owner.UserName, repo.Name, "feature")
	require.NoError(t, err)
	assert.True(t, s)

	bl, _, err = c.ListRepoBranches(repo.Owner.UserName, repo.Name, ListRepoBranchesOptions{})
	require.NoError(t, err)
	assert.Len(t, bl, 2)

	b, _, err = c.GetRepoBranch(repo.Owner.UserName, repo.Name, "feature")
	require.Error(t, err)
	assert.Nil(t, b)

	bNew, _, err := c.CreateBranch(repo.Owner.UserName, repo.Name, CreateBranchOption{BranchName: "NewBranch"})
	require.NoError(t, err)

	b, _, err = c.GetRepoBranch(repo.Owner.UserName, repo.Name, bNew.Name)
	require.NoError(t, err)
	assert.Equal(t, bNew, b)
}

func TestRepoBranchProtection(t *testing.T) {
	log.Println("== TestRepoBranchProtection ==")
	c := newTestClient()
	repoName := "BranchProtection"

	repo := prepareBranchTest(t, c, repoName)
	if repo == nil {
		return
	}
	assert.NotNil(t, repo)

	// ListBranchProtections
	bpl, _, err := c.ListBranchProtections(repo.Owner.UserName, repo.Name, ListBranchProtectionsOptions{})
	require.NoError(t, err)
	assert.Empty(t, bpl)

	// CreateBranchProtection
	bp, _, err := c.CreateBranchProtection(repo.Owner.UserName, repo.Name, CreateBranchProtectionOption{
		BranchName:              "main",
		EnablePush:              true,
		EnablePushWhitelist:     true,
		PushWhitelistUsernames:  []string{"test01"},
		EnableMergeWhitelist:    true,
		MergeWhitelistUsernames: []string{"test01"},
		BlockOnOutdatedBranch:   true,
	})
	require.NoError(t, err)
	assert.Equal(t, "main", bp.BranchName)
	assert.False(t, bp.EnableStatusCheck)
	assert.True(t, bp.EnablePush)
	assert.True(t, bp.EnablePushWhitelist)
	assert.Equal(t, []string{"test01"}, bp.PushWhitelistUsernames)

	bp, _, err = c.CreateBranchProtection(repo.Owner.UserName, repo.Name, CreateBranchProtectionOption{
		BranchName:              "update",
		EnablePush:              false,
		EnableMergeWhitelist:    true,
		MergeWhitelistUsernames: []string{"test01"},
	})
	require.NoError(t, err)
	assert.NotNil(t, bp)

	bpl, _, err = c.ListBranchProtections(repo.Owner.UserName, repo.Name, ListBranchProtectionsOptions{})
	require.NoError(t, err)
	assert.Len(t, bpl, 2)

	// GetBranchProtection
	bp, _, err = c.GetBranchProtection(repo.Owner.UserName, repo.Name, bpl[0].BranchName)
	require.NoError(t, err)
	assert.Equal(t, bpl[0], bp)

	// EditBranchProtection
	bp, _, err = c.EditBranchProtection(repo.Owner.UserName, repo.Name, bpl[0].BranchName, EditBranchProtectionOption{
		EnablePush:                  OptionalBool(false),
		EnablePushWhitelist:         OptionalBool(false),
		PushWhitelistUsernames:      nil,
		RequiredApprovals:           OptionalInt64(1),
		EnableApprovalsWhitelist:    OptionalBool(true),
		ApprovalsWhitelistUsernames: []string{"test01"},
	})
	require.NoError(t, err)
	assert.NotEqual(t, bpl[0], bp)
	assert.Equal(t, bpl[0].BranchName, bp.BranchName)
	assert.Equal(t, bpl[0].EnableMergeWhitelist, bp.EnableMergeWhitelist)
	assert.Equal(t, bpl[0].Created, bp.Created)

	// DeleteBranchProtection
	_, err = c.DeleteBranchProtection(repo.Owner.UserName, repo.Name, bpl[1].BranchName)
	require.NoError(t, err)
	bpl, _, err = c.ListBranchProtections(repo.Owner.UserName, repo.Name, ListBranchProtectionsOptions{})
	require.NoError(t, err)
	assert.Len(t, bpl, 1)
}

func prepareBranchTest(t *testing.T, c *Client, repoName string) *Repository {
	origRepo, err := createTestRepo(t, repoName, c)
	if !assert.NoError(t, err) { //nolint:testifylint
		return nil
	}

	mainLicense, _, err := c.GetContents(origRepo.Owner.UserName, origRepo.Name, "main", "README.md")
	if !assert.NoError(t, err) || !assert.NotNil(t, mainLicense) { //nolint:testifylint
		return nil
	}

	updatedFile, _, err := c.UpdateFile(origRepo.Owner.UserName, origRepo.Name, "README.md", UpdateFileOptions{
		FileOptions: FileOptions{
			Message:       "update it",
			BranchName:    "main",
			NewBranchName: "update",
		},
		SHA:     mainLicense.SHA,
		Content: "Tk9USElORyBJUyBIRVJFIEFOWU1PUkUKSUYgWU9VIExJS0UgVE8gRklORCBTT01FVEhJTkcKV0FJVCBGT1IgVEhFIEZVVFVSRQo=",
	})
	if !assert.NoError(t, err) || !assert.NotNil(t, updatedFile) { //nolint:testifylint
		return nil
	}

	newFile, _, err := c.CreateFile(origRepo.Owner.UserName, origRepo.Name, "WOW-file", CreateFileOptions{
		Content: "QSBuZXcgRmlsZQo=",
		FileOptions: FileOptions{
			Message:       "creat a new file",
			BranchName:    "main",
			NewBranchName: "feature",
		},
	})
	if !assert.NoError(t, err) || !assert.NotNil(t, newFile) { //nolint:testifylint
		return nil
	}

	return origRepo
}
