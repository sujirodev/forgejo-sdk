// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoFromTemplate(t *testing.T) {
	t.Parallel()
	log.Println("== TestRepoFromTemplate ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "TemplateRepo", c)
	require.NoError(t, err)
	repo, _, err = c.EditRepo(repo.Owner.UserName, repo.Name, EditRepoOption{Template: OptionalBool(true)})
	require.NoError(t, err)
	_, err = c.SetRepoTopics(repo.Owner.UserName, repo.Name, []string{"abc", "def", "ghi"})
	require.NoError(t, err)

	newRepo, resp, err := c.CreateRepoFromTemplate(repo.Owner.UserName, repo.Name, CreateRepoFromTemplateOption{
		Owner:       repo.Owner.UserName,
		Name:        "repoFromTemplate",
		Description: "",
		Topics:      true,
		Labels:      true,
	})
	require.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
	assert.False(t, newRepo.Template)

	labels, _, err := c.ListRepoLabels(repo.Owner.UserName, repo.Name, ListLabelsOptions{})
	require.NoError(t, err)
	assert.Len(t, labels, 7)

	topics, _, _ := c.ListRepoTopics(repo.Owner.UserName, repo.Name, ListRepoTopicsOptions{})
	assert.Equal(t, []string{"abc", "def", "ghi"}, topics)

	_, err = c.DeleteRepo(repo.Owner.UserName, "repoFromTemplate")
	require.NoError(t, err)
}
