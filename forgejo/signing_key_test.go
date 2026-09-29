// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetSigningKey exercises the GPG signing key endpoint. Whether a
// default signing key exists depends on server configuration
// ([repository.signing] / a generated commit-signing key). Forgejo answers
// 200 with an empty body rather than an error when none is configured, so
// this skips on either signal: no transport error, but also no key.
func TestGetSigningKey(t *testing.T) {
	t.Parallel()
	log.Println("== TestGetSigningKey ==")
	c := newTestClient()

	key, resp, err := c.GetSigningKey(t.Context())
	if err != nil {
		t.Skipf("no default GPG signing key configured on this server: %v", err)
	}
	if key == "" {
		t.Skip("no default GPG signing key configured on this server: empty response")
	}
	assert.NotNil(t, resp)
	assert.Contains(t, key, "PGP PUBLIC KEY BLOCK")
}

// TestGetSSHSigningKey needs the test instance's [repository.signing]
// section (FORMAT = ssh, SIGNING_KEY = a public key file); without it
// /signing-key.ssh answers 404. The route arrived in Forgejo 12.0.0.
func TestGetSSHSigningKey(t *testing.T) {
	t.Parallel()
	log.Println("== TestGetSSHSigningKey ==")
	c := newTestClient()

	key, resp, err := c.GetSSHSigningKey(t.Context())
	if !serverAtLeast(t, c, "12.0.0") {
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
		return
	}
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.True(t, strings.HasPrefix(key, "ssh-ed25519 "), "want an OpenSSH public key, got %q", key)
}
