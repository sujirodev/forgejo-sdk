// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"io"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRepo(t *testing.T) {
	log.Println("== TestCreateRepo ==")
	c := newTestClient()
	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	repoName := "test1"
	_, _, err = c.GetRepo(user.UserName, repoName)
	if err != nil {
		repo, _, err := c.CreateRepo(CreateRepoOption{
			Name: repoName,
		})
		require.NoError(t, err)
		assert.NotNil(t, repo)
	}

	_, err = c.DeleteRepo(user.UserName, repoName)
	require.NoError(t, err)
}

func TestRepoMigrateAndLanguages(t *testing.T) {
	log.Println("== TestMigrateRepo ==")
	c := newTestClient()
	user, _, uErr := c.GetMyUserInfo()
	require.NoError(t, uErr)
	_, _, err := c.GetRepo(user.UserName, "sdk-mirror")
	if err == nil {
		_, _ = c.DeleteRepo(user.UserName, "sdk-mirror")
	}

	// TODO: replace by proper url for forgejo
	repoM, _, err := c.MigrateRepo(MigrateRepoOption{
		CloneAddr:   "https://codeberg.org/mvdkleijn/forgejo-sdk.git",
		RepoName:    "sdk-mirror",
		RepoOwner:   user.UserName,
		Mirror:      true,
		Private:     false,
		Description: "mirror sdk",
	})
	require.NoError(t, err)

	repoG, _, err := c.GetRepo(repoM.Owner.UserName, repoM.Name)
	require.NoError(t, err)
	assert.Equal(t, repoM.ID, repoG.ID)
	assert.Equal(t, "main", repoG.DefaultBranch)
	assert.True(t, repoG.Mirror)
	assert.False(t, repoG.Empty)
	assert.Equal(t, 1, repoG.Watchers)
	var zeroTime time.Time
	assert.NotEqual(t, zeroTime, repoG.MirrorUpdated)

	log.Println("== TestRepoLanguages ==")
	time.Sleep(time.Second * 2)
	lang, _, err := c.GetRepoLanguages(repoM.Owner.UserName, repoM.Name)
	require.NoError(t, err)
	assert.Len(t, lang, 2)
	assert.Less(t, int64(217441), lang["Go"])
	assert.True(t, 3614 < lang["Makefile"] && 15000 > lang["Makefile"])
}

func TestSearchRepo(t *testing.T) {
	log.Println("== TestSearchRepo ==")
	c := newTestClient()

	repo, err := createTestRepo(t, "RepoSearch1", c)
	require.NoError(t, err)
	_, err = c.AddRepoTopic(repo.Owner.UserName, repo.Name, "TestTopic1")
	require.NoError(t, err)
	_, err = c.AddRepoTopic(repo.Owner.UserName, repo.Name, "TestTopic2")
	require.NoError(t, err)

	repo, err = createTestRepo(t, "RepoSearch2", c)
	require.NoError(t, err)
	_, err = c.AddRepoTopic(repo.Owner.UserName, repo.Name, "TestTopic1")
	require.NoError(t, err)

	repos, _, err := c.SearchRepos(SearchRepoOptions{
		Keyword:              "Search1",
		KeywordInDescription: true,
	})
	require.NoError(t, err)
	assert.NotNil(t, repos)
	assert.Len(t, repos, 1)

	repos, _, err = c.SearchRepos(SearchRepoOptions{
		Keyword:              "Search",
		KeywordInDescription: true,
	})
	require.NoError(t, err)
	assert.NotNil(t, repos)
	assert.Len(t, repos, 2)

	repos, _, err = c.SearchRepos(SearchRepoOptions{
		Keyword:              "TestTopic1",
		KeywordInDescription: true,
	})
	require.NoError(t, err)
	assert.NotNil(t, repos)
	assert.Len(t, repos, 2)

	repos, _, err = c.SearchRepos(SearchRepoOptions{
		Keyword:              "TestTopic2",
		KeywordInDescription: true,
	})
	require.NoError(t, err)
	assert.NotNil(t, repos)
	assert.Len(t, repos, 1)

	_, err = c.DeleteRepo(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
}

func TestDeleteRepo(t *testing.T) {
	log.Println("== TestDeleteRepo ==")
	c := newTestClient()
	repo, _ := createTestRepo(t, "TestDeleteRepo", c)
	_, err := c.DeleteRepo(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
}

func TestGetArchive(t *testing.T) {
	log.Println("== TestGetArchive ==")
	c := newTestClient()
	repo, _ := createTestRepo(t, "ToDownload", c)
	time.Sleep(time.Second / 2)
	archive, _, err := c.GetArchive(repo.Owner.UserName, repo.Name, "main", ZipArchive)
	require.NoError(t, err)
	expectedSizeMin, expectedSizeMax := 1500, 1800
	assert.True(t, len(archive) > expectedSizeMin && len(archive) < expectedSizeMax, "archive size: %d is not within range of (%d, %d)", len(archive), expectedSizeMin, expectedSizeMax)
}

func TestGetArchiveReader(t *testing.T) {
	log.Println("== TestGetArchiveReader ==")
	c := newTestClient()
	repo, _ := createTestRepo(t, "ToDownload", c)
	time.Sleep(time.Second / 2)
	r, _, err := c.GetArchiveReader(repo.Owner.UserName, repo.Name, "main", ZipArchive)
	require.NoError(t, err)
	defer r.Close()

	archive := bytes.NewBuffer(nil)
	nBytes, err := io.Copy(archive, r)
	require.NoError(t, err)
	assert.Greater(t, nBytes, int64(1500))
	assert.Equal(t, nBytes, int64(len(archive.Bytes())))
}

func TestGetRepoByID(t *testing.T) {
	log.Println("== TestGetRepoByID ==")
	c := newTestClient()
	testrepo, _ := createTestRepo(t, "TestGetRepoByID", c)

	repo, _, err := c.GetRepoByID(testrepo.ID)
	require.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, testrepo.ID, repo.ID)

	_, err = c.DeleteRepo(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
}

// standard func to create a init repo for test routines
func createTestRepo(t *testing.T, name string, c *Client) (*Repository, error) {
	user, _, uErr := c.GetMyUserInfo()
	require.NoError(t, uErr)
	repo, _, err := c.GetRepo(user.UserName, name)
	// We need to check that the received repo is not a
	// redirected one, it could be the case that forgejo redirect us
	// to a new repo(because it e.g. was transferred or renamed).
	if err == nil && repo.Owner.UserName == user.UserName {
		_, _ = c.DeleteRepo(user.UserName, name)
	}

	repo, _, err = c.CreateRepo(CreateRepoOption{
		Name:        name,
		Description: "A test Repo: " + name,
		AutoInit:    true,
		Gitignores:  "C,C++",
		License:     "MIT",
		Readme:      "Default",
		IssueLabels: "Default",
		Private:     false,
	})
	require.NoError(t, err)
	assert.NotNil(t, repo)

	return repo, err
}

func TestRepos_ListMyListUserListOrg(t *testing.T) {
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})

	myRepos, _, err := c.ListMyRepos(ListReposOptions{})
	require.NoError(t, err)
	assert.True(t, containsRepoName(myRepos, repo.Name))

	userRepos, _, err := c.ListUserRepos(repo.Owner.UserName, ListReposOptions{})
	require.NoError(t, err)
	assert.True(t, containsRepoName(userRepos, repo.Name))

	org := newTestOrg(t, c)
	orgRepo, _, err := c.CreateOrgRepo(org.UserName, CreateRepoOption{Name: uniqueName(t, "orgrepo"), AutoInit: true})
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = c.DeleteRepo(org.UserName, orgRepo.Name) })

	orgRepos, _, err := c.ListOrgRepos(org.UserName, ListOrgReposOptions{})
	require.NoError(t, err)
	assert.True(t, containsRepoName(orgRepos, orgRepo.Name))
}

// Needs [migrations] ALLOW_LOCALNETWORKS = true on the test instance (see
// docs/PLANO-COBERTURA-TESTES.md section 4 / scripts/check-test-instance-settings.sh):
// otherwise Forgejo refuses a clone address that resolves to itself.
func TestMirrorSync(t *testing.T) {
	c := newTestClient()
	source := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true})

	mirror, _, err := c.MigrateRepo(MigrateRepoOption{
		CloneAddr: source.CloneURL,
		RepoName:  uniqueName(t, "mirror"),
		RepoOwner: source.Owner.UserName,
		Mirror:    true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = c.DeleteRepo(mirror.Owner.UserName, mirror.Name) })
	assert.True(t, mirror.Mirror)

	_, err = c.MirrorSync(mirror.Owner.UserName, mirror.Name)
	require.NoError(t, err)
}
