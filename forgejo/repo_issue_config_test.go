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

func TestRepoIssueConfig(t *testing.T) {
	log.Println("== TestRepoIssueConfig ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "IssueConfig", c)
	require.NoError(t, err)

	// no .forgejo/ISSUE_TEMPLATE/config.yaml exists yet, so the default config applies
	cfg, resp, err := c.GetIssueConfig(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.True(t, cfg.BlankIssuesEnabled)

	v, _, err := c.ValidateIssueConfig(repo.Owner.UserName, repo.Name)
	require.NoError(t, err)
	assert.True(t, v.Valid)
}
