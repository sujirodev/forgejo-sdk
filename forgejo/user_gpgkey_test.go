// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/openpgp"       //nolint:staticcheck // frozen but enough for a throwaway test key; avoids a new dependency
	"golang.org/x/crypto/openpgp/armor" //nolint:staticcheck // see above
)

func TestGetGPGKeyVerificationToken(t *testing.T) {
	t.Parallel()
	log.Println("== TestGetGPGKeyVerificationToken ==")
	c := newTestClient()

	token, resp, err := c.GetGPGKeyVerificationToken()
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, token)
}

func TestVerifyGPGKeyRejectsBogusSignature(t *testing.T) {
	t.Parallel()
	log.Println("== TestVerifyGPGKeyRejectsBogusSignature ==")
	c := newTestClient()

	_, _, err := c.VerifyGPGKey(VerifyGPGKeyOption{
		KeyID:     "0000000000000000",
		Signature: "not a real signature",
	})
	require.Error(t, err)
}

// newSignedGPGKey generates a throwaway RSA OpenPGP key whose UID email is
// email, and returns its armored public key, its key ID as Forgejo reports it
// (16 upper-case hex digits), and an armored detached signature of message.
func newSignedGPGKey(t *testing.T, email, message string) (armoredKey, keyID, signature string) {
	t.Helper()
	entity, err := openpgp.NewEntity("forgejo-sdk verify test", "", email, nil)
	require.NoError(t, err)

	var pub bytes.Buffer
	w, err := armor.Encode(&pub, openpgp.PublicKeyType, nil)
	require.NoError(t, err)
	require.NoError(t, entity.Serialize(w))
	require.NoError(t, w.Close())

	var sig bytes.Buffer
	require.NoError(t, openpgp.ArmoredDetachSign(&sig, entity, strings.NewReader(message), nil))

	return pub.String(), fmt.Sprintf("%016X", entity.PrimaryKey.KeyId), sig.String()
}

// TestVerifyGPGKey runs the real verification flow: it generates a keypair,
// takes the token from GetGPGKeyVerificationToken, signs it with the private
// key and submits the armored detached signature. It runs as a dedicated
// user, not the shared admin, so the key it adds does not disturb
// TestUserGPGKeys' exact key counts on that account.
func TestVerifyGPGKey(t *testing.T) {
	t.Parallel()
	log.Println("== TestVerifyGPGKey ==")
	c := newTestClient()
	user := createTestUser(t, uniqueName(t, "gpgverify"), c)

	asUser(t, c, user.UserName, func() {
		token, _, err := c.GetGPGKeyVerificationToken()
		require.NoError(t, err)
		require.NotEmpty(t, token)

		armoredKey, keyID, signature := newSignedGPGKey(t, user.Email, token)

		added, _, err := c.CreateGPGKey(CreateGPGKeyOption{ArmoredKey: armoredKey})
		require.NoError(t, err)
		t.Cleanup(func() {
			asUser(t, c, user.UserName, func() { _, _ = c.DeleteGPGKey(added.ID) })
		})
		require.Equal(t, keyID, strings.ToUpper(added.KeyID))

		verified, resp, err := c.VerifyGPGKey(VerifyGPGKeyOption{KeyID: keyID, Signature: signature})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, keyID, strings.ToUpper(verified.KeyID))
		assert.NotZero(t, verified.ID)
	})
}

// TestUserGPGKeys uses the fixture in testdata/gpg_test01.asc, whose UID
// email matches the CI/local test instance's admin account
// (test01@forgejo.org), so it attaches to the default test client's own
// user rather than needing a second one.
func TestUserGPGKeys(t *testing.T) {
	t.Parallel()
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
