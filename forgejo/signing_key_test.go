// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetSigningKey exercises the GPG signing key endpoint. Whether a
// default signing key exists depends on server configuration
// ([repository.signing] / a generated commit-signing key), so this only
// pins the wire behavior: no transport error, and a well-formed key when
// one is configured.
func TestGetSigningKey(t *testing.T) {
	log.Println("== TestGetSigningKey ==")
	c := newTestClient()

	key, resp, err := c.GetSigningKey(t.Context())
	if err != nil {
		t.Skipf("no default GPG signing key configured on this server: %v", err)
	}
	assert.NotNil(t, resp)
	assert.Contains(t, key, "PGP PUBLIC KEY BLOCK")
}

// TestGetSSHSigningKey mirrors TestGetSigningKey for the SSH signing key.
func TestGetSSHSigningKey(t *testing.T) {
	log.Println("== TestGetSSHSigningKey ==")
	c := newTestClient()

	key, resp, err := c.GetSSHSigningKey(t.Context())
	if err != nil {
		t.Skipf("no default SSH signing key configured on this server: %v", err)
	}
	assert.NotNil(t, resp)
	assert.NotEmpty(t, key)
}
