// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIssuePin exercises pinning, moving and unpinning issues
func TestIssuePin(t *testing.T) {
	log.Println("== TestIssuePin ==")

	c := newTestClient()

	user, _, err := c.GetMyUserInfo()
	require.NoError(t, err)
	repo, err := createTestRepo(t, "TestIssuePinRepo", c)
	require.NoError(t, err)
	issue1, _, err := c.CreateIssue(user.UserName, repo.Name, CreateIssueOption{Title: "pin me first"})
	require.NoError(t, err)
	issue2, _, err := c.CreateIssue(user.UserName, repo.Name, CreateIssueOption{Title: "pin me second"})
	require.NoError(t, err)

	// PinIssue
	_, err = c.PinIssue(user.UserName, repo.Name, issue1.Index)
	require.NoError(t, err)
	_, err = c.PinIssue(user.UserName, repo.Name, issue2.Index)
	require.NoError(t, err)

	// MoveIssuePin - move the second pinned issue to the first position
	_, err = c.MoveIssuePin(user.UserName, repo.Name, issue2.Index, 1)
	require.NoError(t, err)

	// UnpinIssue
	_, err = c.UnpinIssue(user.UserName, repo.Name, issue1.Index)
	require.NoError(t, err)
	_, err = c.UnpinIssue(user.UserName, repo.Name, issue2.Index)
	require.NoError(t, err)
}
