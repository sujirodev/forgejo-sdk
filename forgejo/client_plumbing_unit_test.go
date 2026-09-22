// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

// Unit tests for the *Client setter methods (as opposed to the ClientOption
// wrappers around them, which the rest of this package's tests already
// exercise via NewClient) and for SignRequest / parseLinkHeader.

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestUnit_ClientMethod_SetHTTPClient(t *testing.T) {
	c, err := NewClient("http://example.invalid", SetForgejoVersion("16.0.5"))
	require.NoError(t, err)

	custom := &http.Client{Timeout: 42 * time.Second}
	c.SetHTTPClient(custom)
	assert.Same(t, custom, c.client)
}

func TestUnit_ClientMethod_SetOTP(t *testing.T) {
	var gotOTP string
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotOTP = r.Header.Get("X-FORGEJO-OTP")
		w.WriteHeader(http.StatusOK)
	})
	c := newUnitTestClient(t, srv)

	c.SetOTP("654321")
	_, err := c.doRequest("GET", "/x", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "654321", gotOTP)
}

func TestUnit_ClientMethod_SetContext(t *testing.T) {
	c, err := NewClient("http://example.invalid", SetForgejoVersion("16.0.5"))
	require.NoError(t, err)

	type key int
	ctx := context.WithValue(context.Background(), key(0), "marker")
	c.SetContext(ctx)
	assert.Equal(t, ctx, c.ctx)
}

func TestUnit_ClientMethod_SetUserAgent(t *testing.T) {
	var gotUA string
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	})
	c := newUnitTestClient(t, srv)

	c.SetUserAgent("sdk-test-agent")
	_, err := c.doRequest("GET", "/x", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "sdk-test-agent", gotUA)
}

// writeUnencryptedEd25519Key generates a fresh ed25519 keypair and writes
// its unencrypted OpenSSH-format private key to a temp file, returning the
// path. UseSSHPubkey(..., sshKey, "") reads exactly this format.
func writeUnencryptedEd25519Key(t *testing.T) string {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	block, err := ssh.MarshalPrivateKey(priv, "")
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "id_ed25519")
	require.NoError(t, os.WriteFile(path, pem.EncodeToMemory(block), 0o600))
	return path
}

func TestUnit_SignRequest(t *testing.T) {
	keyPath := writeUnencryptedEd25519Key(t)

	c, err := NewClient("http://example.invalid", SetForgejoVersion("16.0.5"), UseSSHPubkey("", keyPath, ""))
	require.NoError(t, err)
	require.NotNil(t, c.httpsigner)

	req, err := http.NewRequest(http.MethodGet, "http://example.invalid/api/v1/version", nil)
	require.NoError(t, err)
	before := len(req.Header)

	require.NoError(t, c.SignRequest(req))
	assert.Greater(t, len(req.Header), before, "SignRequest should have added signature headers")
}

func TestUnit_ParseLinkHeader(t *testing.T) {
	link := `<https://example.com/api/v1/repos/search?page=1>; rel="first", ` +
		`<https://example.com/api/v1/repos/search?page=2>; rel="prev", ` +
		`<https://example.com/api/v1/repos/search?page=4>; rel="next", ` +
		`<https://example.com/api/v1/repos/search?page=10>; rel="last"`

	resp := &Response{Response: &http.Response{Header: http.Header{}}}
	resp.Header.Set("Link", link)
	resp.parseLinkHeader()

	assert.Equal(t, 1, resp.FirstPage)
	assert.Equal(t, 2, resp.PrevPage)
	assert.Equal(t, 4, resp.NextPage)
	assert.Equal(t, 10, resp.LastPage)
}
