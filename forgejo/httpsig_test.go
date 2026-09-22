// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

// TestHTTPSigAuth is the end-to-end proof that UseSSHPubkey/SignRequest
// authenticate a request against a real Forgejo server, closing out fase 8's
// httpsign.go work with more than an assertion on header shapes.
//
// It is worth being explicit about what this does and doesn't prove,
// because "signed HTTP request" is used for two unrelated things in this
// codebase:
//
//   - UseSSHCert/UseSSHPubkey (httpsign.go, tested here): an alternative to
//     basic-auth/token authentication, keyed to an SSH public key the user
//     registered via CreatePublicKey. This is what the test below exercises:
//     a client with *no* basic auth and *no* token, authenticated purely by
//     a signed request, resolves to the user who owns the key.
//   - The ActivityPub routes' own HTTP Signature check (issue #58): remote
//     federated instances signing requests with a key served from their own
//     actor document, verified by dereferencing that document. httpsign.go
//     has no part in this; there is no local action that makes an
//     ActivityPub route accept a request from this SDK, short of running a
//     second, publicly reachable Forgejo instance to act as the remote
//     peer. That is why the ActivityPub routes are still `negative-only`/
//     `needs-config` exceptions after this change.
import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestHTTPSigAuth(t *testing.T) {
	t.Parallel()
	log.Println("== TestHTTPSigAuth ==")
	admin := newTestClient()

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	sshPub, err := ssh.NewPublicKey(pub)
	require.NoError(t, err)
	pubLine := string(ssh.MarshalAuthorizedKey(sshPub))

	key, _, err := admin.CreatePublicKey(CreateKeyOption{Title: uniqueName(t, "httpsig"), Key: pubLine})
	require.NoError(t, err)
	t.Cleanup(func() { _, _ = admin.DeletePublicKey(key.ID) })

	block, err := ssh.MarshalPrivateKey(priv, "")
	require.NoError(t, err)
	keyPath := filepath.Join(t.TempDir(), "id_ed25519")
	require.NoError(t, os.WriteFile(keyPath, pem.EncodeToMemory(block), 0o600))

	// No SetBasicAuth, no SetToken: UseSSHPubkey is this client's only
	// authentication.
	c, err := NewClient(getForgejoURL(), UseSSHPubkey("", keyPath, ""))
	require.NoError(t, err)

	me, _, err := c.GetMyUserInfo()
	require.NoError(t, err)
	assert.Equal(t, getForgejoUsername(), me.UserName, "the signature resolves to the user who owns the registered key")
}
