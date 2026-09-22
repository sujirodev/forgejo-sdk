// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

// Unit tests for httpsign.go: the SSH-certificate and SSH-pubkey signed
// HTTP request scheme UseSSHCert/UseSSHPubkey enable. These run against
// generated keys and an httptest.Server, no Forgejo instance needed. The
// end-to-end proof that a signed request actually authenticates against a
// real server lives in httpsig_test.go (fase 8, closes issue #58's
// "needs-config" ActivityPub routes only by way of establishing that the
// two auth schemes are unrelated -- see that file's comment).

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// writeEd25519Key generates a fresh ed25519 keypair and writes its OpenSSH
// PEM private key to <dir>/id_ed25519, encrypted with passphrase when
// non-empty. Returns the key path and the raw public key.
func writeEd25519Key(t *testing.T, dir, passphrase string) (string, ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	var block *pem.Block
	if passphrase == "" {
		block, err = ssh.MarshalPrivateKey(priv, "")
	} else {
		block, err = ssh.MarshalPrivateKeyWithPassphrase(priv, "", []byte(passphrase))
	}
	require.NoError(t, err)

	path := filepath.Join(dir, "id_ed25519")
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(block), 0o600))
	return path, pub
}

// writeSSHCert signs a fresh ed25519 user key with a fresh CA key for the
// given principal and writes both the private key and the
// "<path>-cert.pub" certificate getSignerFromFile expects. validFor <= 0
// produces an already-expired certificate.
func writeSSHCert(t *testing.T, dir, principal string, validFor time.Duration) string {
	t.Helper()
	_, caPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	caSigner, err := ssh.NewSignerFromKey(caPriv)
	require.NoError(t, err)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	sshPub, err := ssh.NewPublicKey(pub)
	require.NoError(t, err)

	validBefore := uint64(time.Now().Add(validFor).Unix())
	cert := &ssh.Certificate{
		Key:             sshPub,
		CertType:        ssh.UserCert,
		ValidPrincipals: []string{principal},
		ValidAfter:      uint64(time.Now().Add(-time.Hour).Unix()),
		ValidBefore:     validBefore,
	}
	require.NoError(t, cert.SignCert(rand.Reader, caSigner))

	path := filepath.Join(dir, "id_ed25519")
	block, err := ssh.MarshalPrivateKey(priv, "")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(block), 0o600))
	require.NoError(t, os.WriteFile(path+"-cert.pub", ssh.MarshalAuthorizedKey(cert), 0o600))
	return path
}

// --- getSignerFromFile / newHTTPSign ---

func TestUnit_NewHTTPSignWithPubkey_UnencryptedKey(t *testing.T) {
	path, pub := writeEd25519Key(t, t.TempDir(), "")

	sign, err := NewHTTPSignWithPubkey("", path, "")
	require.NoError(t, err)
	require.NotNil(t, sign)
	assert.False(t, sign.cert)
	assert.Equal(t, pub, sign.Signer.PublicKey().(ssh.CryptoPublicKey).CryptoPublicKey())
}

func TestUnit_NewHTTPSignWithPubkey_EncryptedKey(t *testing.T) {
	dir := t.TempDir()
	path, _ := writeEd25519Key(t, dir, "s3cret")

	_, err := NewHTTPSignWithPubkey("", path, "wrong")
	require.Error(t, err)

	sign, err := NewHTTPSignWithPubkey("", path, "s3cret")
	require.NoError(t, err)
	require.NotNil(t, sign)
}

func TestUnit_NewHTTPSignWithPubkey_MissingFile(t *testing.T) {
	_, err := NewHTTPSignWithPubkey("", filepath.Join(t.TempDir(), "nope"), "")
	require.Error(t, err)
}

func TestUnit_NewHTTPSignWithCert_Valid(t *testing.T) {
	path := writeSSHCert(t, t.TempDir(), "test01", time.Hour)

	sign, err := NewHTTPSignWithCert("test01", path, "")
	require.NoError(t, err)
	require.NotNil(t, sign)
	assert.True(t, sign.cert)
}

func TestUnit_NewHTTPSignWithCert_MissingCertFile(t *testing.T) {
	// A plain (non-certificate) key has no "-cert.pub" sibling.
	path, _ := writeEd25519Key(t, t.TempDir(), "")

	_, err := NewHTTPSignWithCert("test01", path, "")
	require.Error(t, err)
}

// --- findCertSigner / findPubkeySigner ---

func certSigner(t *testing.T, principal string, validFor time.Duration) ssh.Signer {
	t.Helper()
	_, caPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	caSigner, err := ssh.NewSignerFromKey(caPriv)
	require.NoError(t, err)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	sshPub, err := ssh.NewPublicKey(pub)
	require.NoError(t, err)
	userSigner, err := ssh.NewSignerFromKey(priv)
	require.NoError(t, err)

	cert := &ssh.Certificate{
		Key:             sshPub,
		CertType:        ssh.UserCert,
		ValidPrincipals: []string{principal},
		ValidAfter:      uint64(time.Now().Add(-time.Hour).Unix()),
		ValidBefore:     uint64(time.Now().Add(validFor).Unix()),
	}
	require.NoError(t, cert.SignCert(rand.Reader, caSigner))

	certSigner, err := ssh.NewCertSigner(cert, userSigner)
	require.NoError(t, err)
	return certSigner
}

func plainSigner(t *testing.T) ssh.Signer {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	s, err := ssh.NewSignerFromKey(priv)
	require.NoError(t, err)
	return s
}

// fingerprintOf identifies an ssh.Signer by its public key fingerprint, since
// the concrete types findCertSigner/findPubkeySigner hand back aren't always
// comparable with assert.Same.
func fingerprintOf(s ssh.Signer) string {
	if s == nil {
		return ""
	}
	return ssh.FingerprintSHA256(s.PublicKey())
}

func TestUnit_FindCertSigner(t *testing.T) {
	forAlice := certSigner(t, "alice", time.Hour)
	forBob := certSigner(t, "bob", time.Hour)
	expired := certSigner(t, "alice", time.Second) // < the 10s grace window
	plain := plainSigner(t)

	signers := []ssh.Signer{plain, expired, forAlice, forBob}

	assert.Equal(t, fingerprintOf(forAlice), fingerprintOf(findCertSigner(signers, "alice")))
	assert.Equal(t, fingerprintOf(forBob), fingerprintOf(findCertSigner(signers, "bob")))
	assert.Nil(t, findCertSigner(signers, "carol"), "no cert matches this principal")
	assert.Nil(t, findCertSigner([]ssh.Signer{plain}, ""), "a non-certificate key is never returned")

	// No principal requested: the first valid (non-expired) certificate wins.
	first := findCertSigner(signers, "")
	require.NotNil(t, first)
	assert.NotEqual(t, fingerprintOf(expired), fingerprintOf(first))
}

func TestUnit_FindPubkeySigner(t *testing.T) {
	a := plainSigner(t)
	b := plainSigner(t)
	cert := certSigner(t, "alice", time.Hour)

	signers := []ssh.Signer{cert, a, b}

	assert.Equal(t, fingerprintOf(a), fingerprintOf(findPubkeySigner(signers, "")), "no fingerprint: first non-certificate key")

	fp := ssh.FingerprintSHA256(b.PublicKey())
	assert.Equal(t, fingerprintOf(b), fingerprintOf(findPubkeySigner(signers, fp)), "matches by SHA256 fingerprint")

	authorizedLine := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(b.PublicKey())))
	assert.Equal(t, fingerprintOf(b), fingerprintOf(findPubkeySigner(signers, authorizedLine)), "matches by authorized_keys line")

	assert.Nil(t, findPubkeySigner(signers, "SHA256:doesnotexist"))
	assert.Nil(t, findPubkeySigner([]ssh.Signer{cert}, ""), "a certificate is never returned as a pubkey signer")
}

// --- UseSSHCert / UseSSHPubkey version gate ---

func TestUnit_UseSSHPubkey_VersionGate(t *testing.T) {
	path, _ := writeEd25519Key(t, t.TempDir(), "")

	_, err := NewClient("http://example.invalid", SetForgejoVersion("1.16.0"), UseSSHPubkey("", path, ""))
	require.Error(t, err, "signed requests need >= 1.17.0")

	c, err := NewClient("http://example.invalid", SetForgejoVersion("1.17.0"), UseSSHPubkey("", path, ""))
	require.NoError(t, err)
	require.NotNil(t, c.httpsigner)
}

func TestUnit_UseSSHCert_VersionGate(t *testing.T) {
	path := writeSSHCert(t, t.TempDir(), "test01", time.Hour)

	_, err := NewClient("http://example.invalid", SetForgejoVersion("1.16.0"), UseSSHCert("test01", path, ""))
	require.Error(t, err)

	c, err := NewClient("http://example.invalid", SetForgejoVersion("1.17.0"), UseSSHCert("test01", path, ""))
	require.NoError(t, err)
	require.NotNil(t, c.httpsigner)
	assert.True(t, c.httpsigner.cert)
}

// --- SignRequest ---

func TestUnit_SignRequest_PubkeyMode(t *testing.T) {
	path, _ := writeEd25519Key(t, t.TempDir(), "")

	for _, tc := range []struct {
		name    string
		version string
	}{
		{"legacy (< 1.23.0)", "1.17.0"},
		{"current (>= 1.23.0)", "16.0.5"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewClient("http://example.invalid", SetForgejoVersion(tc.version), UseSSHPubkey("", path, ""))
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodGet, "http://example.invalid/api/v1/version", nil)
			require.NoError(t, err)

			require.NoError(t, c.SignRequest(req))
			assert.NotEmpty(t, req.Header.Get("Signature"), "SignRequest sets the Signature header")
			assert.Contains(t, req.Header.Get("Signature"), "keyId=", "the signature identifies the signing key")
			assert.Empty(t, req.Header.Get("x-ssh-certificate"), "pubkey mode never sends a certificate")
		})
	}
}

func TestUnit_SignRequest_CertMode(t *testing.T) {
	path := writeSSHCert(t, t.TempDir(), "test01", time.Hour)
	c, err := NewClient("http://example.invalid", SetForgejoVersion("16.0.5"), UseSSHCert("test01", path, ""))
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/api/v1/version", nil)
	require.NoError(t, err)

	require.NoError(t, c.SignRequest(req))
	assert.NotEmpty(t, req.Header.Get("x-ssh-certificate"), "cert mode attaches the certificate")
	assert.NotEmpty(t, req.Header.Get("Signature"))
}

func TestUnit_SignRequest_WithBody(t *testing.T) {
	path, _ := writeEd25519Key(t, t.TempDir(), "")
	c, err := NewClient("http://example.invalid", SetForgejoVersion("16.0.5"), UseSSHPubkey("", path, ""))
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "http://example.invalid/api/v1/x", strings.NewReader(`{"a":1}`))
	require.NoError(t, err)

	require.NoError(t, c.SignRequest(req))
	assert.Contains(t, req.Header.Get("Signature"), "digest", "a request with a body signs the Digest header too")
}

func TestUnit_SignRequest_GetBodyFailure(t *testing.T) {
	path, _ := writeEd25519Key(t, t.TempDir(), "")
	c, err := NewClient("http://example.invalid", SetForgejoVersion("16.0.5"), UseSSHPubkey("", path, ""))
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "http://example.invalid/api/v1/x", strings.NewReader(`{"a":1}`))
	require.NoError(t, err)
	req.GetBody = func() (io.ReadCloser, error) { return nil, fmt.Errorf("boom") }

	err = c.SignRequest(req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getBody() failed")
}

func TestUnit_SignRequest_CertMode_NonCertificateKey(t *testing.T) {
	// A client whose httpsigner.cert is true but whose Signer carries an
	// ordinary (non-certificate) public key: SignRequest must refuse to
	// sign rather than send a bogus x-ssh-certificate header. This state
	// is unreachable through UseSSHCert (NewHTTPSignWithCert always fails
	// first when the key has no certificate), so it's built by hand.
	c := &Client{httpsigner: &HTTPSign{cert: true, Signer: plainSigner(t)}}

	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/api/v1/x", nil)
	require.NoError(t, err)

	err = c.SignRequest(req)
	require.EqualError(t, err, "no ssh certificate found")
}
