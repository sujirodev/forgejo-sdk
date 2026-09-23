// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/42wim/httpsig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func unitTestRSAKeyPEM(t *testing.T) (*rsa.PrivateKey, []byte) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return key, pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
}

func TestUnit_NewActivityPubSigner(t *testing.T) {
	key, pkcs1 := unitTestRSAKeyPEM(t)

	pkcs8Bytes, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	pkcs8 := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Bytes})

	for _, tc := range []struct {
		name  string
		keyID string
		pem   []byte
		want  string // substring of the expected error, "" when it must succeed
	}{
		{name: "pkcs1", keyID: "https://peer.example/actor#main-key", pem: pkcs1},
		{name: "pkcs8", keyID: "https://peer.example/actor#main-key", pem: pkcs8},
		{name: "empty key ID", keyID: "  ", pem: pkcs1, want: "empty key ID"},
		{name: "not PEM", keyID: "https://peer.example/actor#main-key", pem: []byte("nope"), want: "not PEM encoded"},
		{
			name:  "wrong PEM block",
			keyID: "https://peer.example/actor#main-key",
			pem:   pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: []byte{1, 2, 3}}),
			want:  "unexpected PEM block",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			signer, err := NewActivityPubSigner(tc.keyID, tc.pem)
			if tc.want != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.keyID, signer.KeyID())

			pubPEM, err := signer.PublicKeyPEM()
			require.NoError(t, err)
			assert.True(t, strings.HasPrefix(pubPEM, "-----BEGIN PUBLIC KEY-----"))
		})
	}
}

// TestUnit_ActivityPubSignatureVerifies checks the SDK's signature against
// the very library Forgejo verifies with (github.com/42wim/httpsig, see
// routers/api/v1/activitypub/reqsignature.go), over the algorithm Forgejo
// defaults to. A signature this rejects is one the server would reject.
func TestUnit_ActivityPubSignatureVerifies(t *testing.T) {
	key, keyPEM := unitTestRSAKeyPEM(t)
	const keyID = "http://peer.example/api/v1/activitypub/user-id/1#main-key"

	var got *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Clone(r.Context())
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"@context":"https://www.w3.org/ns/activitystreams"}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL,
		SetForgejoVersion("16.0.5"),
		UseActivityPubSignature(keyID, keyPEM),
	)
	require.NoError(t, err)

	for _, tc := range []struct {
		name    string
		call    func() error
		headers []string
	}{
		{
			name:    "GET",
			call:    func() error { _, _, err := c.GetActivityPubPerson(1); return err },
			headers: activityPubGetHeaders,
		},
		{
			name: "POST",
			call: func() error {
				_, err := c.SendActivityPubPersonInbox(1, ActivityPubObject{"type": "Follow"})
				return err
			},
			headers: activityPubPostHeaders,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got = nil
			require.NoError(t, tc.call())
			require.NotNil(t, got)

			verifier, err := httpsig.NewVerifier(got)
			require.NoError(t, err)
			assert.Equal(t, keyID, verifier.KeyId())
			require.NoError(t, verifier.Verify(&key.PublicKey, httpsig.RSA_SHA256))

			// httpsig writes the covered header names lowercased into the
			// `headers=` parameter.
			signature := got.Header.Get("Signature")
			for _, h := range tc.headers {
				assert.Contains(t, signature, strings.ToLower(h),
					"the %s header must be covered by the signature", h)
			}
		})
	}
}

// TestUnit_ActivityPubSignatureScope pins that only federation routes are
// signed: a Signature header on the rest of the API means nothing to the
// server and costs a failed authentication attempt in its log.
func TestUnit_ActivityPubSignatureScope(t *testing.T) {
	_, keyPEM := unitTestRSAKeyPEM(t)

	var signed []bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signed = append(signed, r.Header.Get("Signature") != "")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL,
		SetForgejoVersion("16.0.5"),
		UseActivityPubSignature("http://peer.example/api/v1/activitypub/user-id/1#main-key", keyPEM),
	)
	require.NoError(t, err)

	_, _, err = c.GetActivityPubActor()
	require.NoError(t, err)
	_, _, err = c.ServerVersion()
	require.NoError(t, err)

	// Clearing the signer stops the signing, so a client can be handed back
	// to ordinary use without rebuilding it.
	c.SetActivityPubSigner(nil)
	_, _, err = c.GetActivityPubActor()
	require.NoError(t, err)

	assert.Equal(t, []bool{true, false, false}, signed,
		"/activitypub/ is signed while a signer is set, the rest of the API never is")
}
