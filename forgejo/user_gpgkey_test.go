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
