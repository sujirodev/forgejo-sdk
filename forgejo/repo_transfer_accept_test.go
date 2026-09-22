// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Accept/RejectRepoTransfer only have anything to accept or reject when the
// transfer actually goes through the pending-transfer flow. A transfer
// initiated by a site admin (as newTestRepo's owner, the admin test client,
// would be) applies immediately instead: see TestRepoTransfer in
// repo_transfer_test.go. So both tests below use a non-admin source owner,
// created via AdminCreateRepo, and drive TransferRepo itself through asUser
// as that owner, to get a genuinely pending transfer.

func newPendingTransfer(t *testing.T, c *Client) (source, recipient *User, repoName string) {
	t.Helper()
	source = createTestUser(t, uniqueName(t, "xfersrc"), c)
	recipient = createTestUser(t, uniqueName(t, "xferdst"), c)
	repoName = uniqueName(t, "repo")

	repo, _, err := c.AdminCreateRepo(source.UserName, CreateRepoOption{Name: repoName, AutoInit: true})
	require.NoError(t, err)
	require.Equal(t, source.UserName, repo.Owner.UserName)

	asUser(t, c, source.UserName, func() {
		pending, _, err := c.TransferRepo(source.UserName, repoName, TransferRepoOption{NewOwner: recipient.UserName})
		require.NoError(t, err)
		require.NotNil(t, pending)
		// still under source: the transfer is pending, not yet applied.
		assert.Equal(t, source.UserName, pending.Owner.UserName)
	})
	return source, recipient, repoName
}

func TestRepoTransfer_AcceptByRecipient(t *testing.T) {
	t.Parallel()
	c := newTestClient()
	source, recipient, repoName := newPendingTransfer(t, c)

	var accepted *Repository
	asUser(t, c, recipient.UserName, func() {
		var err error
		accepted, _, err = c.AcceptRepoTransfer(source.UserName, repoName)
		require.NoError(t, err)
	})
	require.NotNil(t, accepted)
	assert.Equal(t, recipient.UserName, accepted.Owner.UserName)
	assert.Equal(t, repoName, accepted.Name)

	t.Cleanup(func() { _, _ = c.DeleteRepo(recipient.UserName, repoName) })
}

func TestRepoTransfer_RejectByRecipient(t *testing.T) {
	t.Parallel()
	c := newTestClient()
	source, recipient, repoName := newPendingTransfer(t, c)
	t.Cleanup(func() { _, _ = c.DeleteRepo(source.UserName, repoName) })

	asUser(t, c, recipient.UserName, func() {
		rejected, _, err := c.RejectRepoTransfer(source.UserName, repoName)
		require.NoError(t, err)
		require.NotNil(t, rejected)
		assert.Equal(t, source.UserName, rejected.Owner.UserName)
	})
}
