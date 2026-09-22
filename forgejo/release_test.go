// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelease(t *testing.T) {
	t.Parallel()
	log.Println("== TestRelease ==")
	c := newTestClient()

	repo, _ := createTestRepo(t, "ReleaseTests", c)

	// ListReleases
	rl, _, err := c.ListReleases(repo.Owner.UserName, repo.Name, ListReleasesOptions{})
	require.NoError(t, err)
	assert.Empty(t, rl)

	// CreateRelease
	r, _, err := c.CreateRelease(repo.Owner.UserName, repo.Name, CreateReleaseOption{
		TagName:      "awesome",
		Target:       "main",
		Title:        "Release 1",
		Note:         "yes it's awesome",
		IsDraft:      true,
		IsPrerelease: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "awesome", r.TagName)
	assert.True(t, r.IsPrerelease)
	assert.True(t, r.IsDraft)
	assert.Equal(t, "Release 1", r.Title)
	assert.Equal(t, fmt.Sprintf("%s/api/v1/repos/%s/releases/%d", c.url, repo.FullName, r.ID), r.URL)
	assert.Equal(t, "main", r.Target)
	assert.Equal(t, "yes it's awesome", r.Note)
	assert.Equal(t, c.username, r.Publisher.UserName)
	rl, _, _ = c.ListReleases(repo.Owner.UserName, repo.Name, ListReleasesOptions{})
	assert.Len(t, rl, 1)

	// GetRelease
	r2, _, err := c.GetRelease(repo.Owner.UserName, repo.Name, r.ID)
	require.NoError(t, err)
	assert.Equal(t, r, r2)
	r2, _, err = c.GetReleaseByTag(repo.Owner.UserName, repo.Name, r.TagName)
	require.NoError(t, err)
	assert.Equal(t, r, r2)
	// ListRelease without pre-releases
	tr := true
	rl, _, err = c.ListReleases(repo.Owner.UserName, repo.Name, ListReleasesOptions{
		IsPreRelease: &tr,
	})
	require.NoError(t, err)
	assert.Len(t, rl, 1) // created release is a pre-release
	// test fallback
	r2, _, err = c.fallbackGetReleaseByTag(repo.Owner.UserName, repo.Name, r.TagName)
	require.NoError(t, err)
	assert.Equal(t, r, r2)

	// EditRelease
	r2, _, err = c.EditRelease(repo.Owner.UserName, repo.Name, r.ID, EditReleaseOption{
		Title:        "Release Awesome",
		Note:         "",
		IsDraft:      OptionalBool(false),
		IsPrerelease: OptionalBool(false),
	})
	require.NoError(t, err)
	assert.Equal(t, r.Target, r2.Target)
	assert.False(t, r2.IsDraft)
	assert.False(t, r2.IsPrerelease)
	assert.Equal(t, r.Note, r2.Note)

	// GetLatestRelease
	r3, _, err := c.GetLatestRelease(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.Equal(t, r2, r3)

	// DeleteRelease
	_, err = c.DeleteRelease(repo.Owner.UserName, repo.Name, r.ID)
	require.NoError(t, err)
	rl, _, _ = c.ListReleases(repo.Owner.UserName, repo.Name, ListReleasesOptions{})
	assert.Empty(t, rl)

	// CreateRelease
	_, _, err = c.CreateRelease(repo.Owner.UserName, repo.Name, CreateReleaseOption{
		TagName: "aNewReleaseTag",
		Target:  "main",
		Title:   "Title of aNewReleaseTag",
	})
	require.NoError(t, err)

	// DeleteReleaseByTag
	_, err = c.DeleteReleaseByTag(repo.Owner.UserName, repo.Name, "aNewReleaseTag")
	require.NoError(t, err)
	rl, _, _ = c.ListReleases(repo.Owner.UserName, repo.Name, ListReleasesOptions{})
	assert.Empty(t, rl)
	_, err = c.DeleteReleaseByTag(repo.Owner.UserName, repo.Name, "aNewReleaseTag")
	require.Error(t, err)

	// Test Response if try to get not existing release
	_, resp, err := c.GetRelease(repo.Owner.UserName, repo.Name, 1234)
	require.Error(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, 404, resp.StatusCode)
	}
	_, resp, err = c.GetReleaseByTag(repo.Owner.UserName, repo.Name, "not_here")
	require.Error(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, 404, resp.StatusCode)
	}
	_, resp, err = c.fallbackGetReleaseByTag(repo.Owner.UserName, repo.Name, "not_here")
	require.Error(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, 404, resp.StatusCode)
	}
}
