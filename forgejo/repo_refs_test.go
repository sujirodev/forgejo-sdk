// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoRefs(t *testing.T) {
	c := newTestClient()
	repo := newTestRepo(t, c, CreateRepoOption{Name: uniqueName(t, "repo"), AutoInit: true, DefaultBranch: "main"})

	refs, _, err := c.GetRepoRefs(repo.Owner.UserName, repo.Name, "heads")
	require.NoError(t, err)
	require.Len(t, refs, 1)
	assert.Equal(t, "refs/heads/main", refs[0].Ref)
	assert.Equal(t, "commit", refs[0].Object.Type)
	assert.NotEmpty(t, refs[0].Object.SHA)

	// As of Forgejo 15, GET .../git/refs/{ref} always answers with a JSON
	// array (even for a single exact match: see GetRepoRefs above), so
	// GetRepoRef's "decode a single object, and treat an array as ambiguous"
	// contract never succeeds against a current server. Documenting the
	// actual behavior here rather than silently skipping the route: this
	// looks like a real, separate bug in GetRepoRef worth its own fix.
	_, _, err = c.GetRepoRef(repo.Owner.UserName, repo.Name, "heads/main")
	require.EqualError(t, err, "no exact match found for this ref")

	_, _, err = c.GetRepoRef(repo.Owner.UserName, repo.Name, "heads/does-not-exist")
	require.Error(t, err)
}
