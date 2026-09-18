// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

// Shared helpers for the integration test suite (the Test* functions that
// need a running Forgejo instance; see main_test.go). TestUnit_* files have
// their own httptest-based helpers (see client_unit_test.go).

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"hash/fnv"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// uniqueSeq makes uniqueName collision-proof even when called twice within
// the same nanosecond from the same test.
var uniqueSeq int64

// uniqueName returns a short, collision-resistant name derived from the
// test's own name, so a resource is traceable back to the test that created
// it and reruns don't collide with each other or with the legacy fixed-name
// tests (see docs/PLANO-COBERTURA-TESTES.md section 5: a new test must never
// touch a resource it didn't create).
//
// The test name goes in hashed, not literal: an earlier version embedded the
// sanitized test name directly, and a name like
// "TestIssueSubscription_AddDeleteOtherUser" contains the substring "other",
// which is exactly the kind of keyword several legacy tests search for
// globally (TestUserSearch, TestNotifications) — polluting their exact-count
// assertions. A hash keeps names unique without risking an accidental match.
func uniqueName(t *testing.T, prefix string) string {
	t.Helper()
	h := fnv.New32a()
	_, _ = h.Write([]byte(t.Name()))
	seq := atomic.AddInt64(&uniqueSeq, 1)
	name := fmt.Sprintf("%s-%x-%d%d", prefix, h.Sum32(), time.Now().UnixNano()%1_000_000, seq)
	if len(name) > 60 {
		name = name[:60]
	}
	return name
}

// newTestRepo creates a personal repository for the authenticated user with
// a name unique to the calling test, and registers its deletion on cleanup.
// It fails the test on error.
func newTestRepo(t *testing.T, c *Client, opt CreateRepoOption) *Repository {
	t.Helper()
	if opt.Name == "" {
		opt.Name = uniqueName(t, "repo")
	}
	repo, _, err := c.CreateRepo(opt)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = c.DeleteRepo(repo.Owner.UserName, repo.Name)
	})
	return repo
}

// newTestOrg creates an organization with a name unique to the calling test
// and registers its deletion on cleanup. It fails the test on error.
func newTestOrg(t *testing.T, c *Client) *Organization {
	t.Helper()
	name := uniqueName(t, "org")
	org, _, err := c.CreateOrg(CreateOrgOption{
		Name:       name,
		Visibility: VisibleTypePublic,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = c.DeleteOrg(org.UserName)
	})
	return org
}

// asUser runs fn with c impersonating username via the sudo header (c must
// be an admin client), then restores c's previous sudo value. This is the
// established pattern in this suite (see createTestRepoForActions in
// repo_action_test.go) for exercising an endpoint that must be called by a
// specific non-admin identity, without standing up a second *Client.
func asUser(t *testing.T, c *Client, username string, fn func()) {
	t.Helper()
	c.mutex.RLock()
	previous := c.sudo
	c.mutex.RUnlock()
	c.SetSudo(username)
	defer c.SetSudo(previous)
	fn()
}

// genSSHPublicKey generates a fresh ed25519 keypair and returns its
// authorized_keys line. Every call returns a distinct key: Forgejo rejects
// registering the same public key twice (as a user key, deploy key, or
// across the two), so tests that need more than one key must call this more
// than once.
func genSSHPublicKey(t *testing.T, comment string) string {
	t.Helper()
	pub, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	sshPub, err := ssh.NewPublicKey(pub)
	require.NoError(t, err)
	line := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPub)))
	if comment != "" {
		line += " " + comment
	}
	return line
}

// testGPGPublicKey returns the armored public key in testdata/gpg_test01.asc:
// an ed25519 OpenPGP key generated for this suite with the UID
// "forgejo-sdk test <test01@forgejo.org>", matching the CI/local test
// instance's admin user email so Forgejo can verify it against that account.
// It's a fixture, not a secret: only the public key is committed.
func testGPGPublicKey(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("testdata/gpg_test01.asc")
	require.NoError(t, err)
	return string(data)
}
