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

func TestGetGPGKeyVerificationToken(t *testing.T) {
	log.Println("== TestGetGPGKeyVerificationToken ==")
	c := newTestClient()

	token, resp, err := c.GetGPGKeyVerificationToken()
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, token)
}

// TestVerifyGPGKeyRejectsBogusSignature only exercises the failure path.
// A full round trip would need a real GPG keypair to sign the verification
// token from GetGPGKeyVerificationToken, which this environment does not
// have available, so it is not live-verified here.
func TestVerifyGPGKeyRejectsBogusSignature(t *testing.T) {
	log.Println("== TestVerifyGPGKeyRejectsBogusSignature ==")
	c := newTestClient()

	_, _, err := c.VerifyGPGKey(VerifyGPGKeyOption{
		KeyID:     "0000000000000000",
		Signature: "not a real signature",
	})
	require.Error(t, err)
}

// TestUserGPGKeys uses the fixture in testdata/gpg_test01.asc, whose UID
// email matches the CI/local test instance's admin account
// (test01@forgejo.org), so it attaches to the default test client's own
// user rather than needing a second one.
func TestUserGPGKeys(t *testing.T) {
	c := newTestClient()

	baseline, _, err := c.ListMyGPGKeys(&ListGPGKeysOptions{})
	require.NoError(t, err)

	created, _, err := c.CreateGPGKey(CreateGPGKeyOption{ArmoredKey: testGPGPublicKey(t)})
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = c.DeleteGPGKey(created.ID) })
	require.Len(t, created.Emails, 1)
	assert.Equal(t, "test01@forgejo.org", created.Emails[0].Email)

	fetched, _, err := c.GetGPGKey(created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.KeyID, fetched.KeyID)

	mine, _, err := c.ListMyGPGKeys(&ListGPGKeysOptions{})
	require.NoError(t, err)
	assert.Len(t, mine, len(baseline)+1)

	all, _, err := c.ListGPGKeys("test01", ListGPGKeysOptions{})
	require.NoError(t, err)
	assert.Len(t, all, len(baseline)+1)

	_, err = c.DeleteGPGKey(created.ID)
	require.NoError(t, err)

	mine, _, err = c.ListMyGPGKeys(&ListGPGKeysOptions{})
	require.NoError(t, err)
	assert.Len(t, mine, len(baseline))
}
