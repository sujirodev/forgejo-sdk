// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/42wim/httpsig"
)

// ActivityPubContentType is the media type a federated ActivityPub payload
// is sent and accepted with.
const ActivityPubContentType = `application/ld+json; profile="https://www.w3.org/ns/activitystreams"`

// How long a signature stays valid, in seconds. Forgejo's own federation
// client uses 60; matching it keeps the clock-skew tolerance symmetric
// between the two ends.
const activityPubSignatureExpiry = 60

// The signature parameters Forgejo's federation module defaults to, from
// modules/setting/federation.go: SignatureAlgorithms[0] is what
// routers/api/v1/activitypub/reqsignature.go verifies with, and GetHeaders /
// PostHeaders are the header sets its own outgoing requests sign. A server
// with non-default [federation] settings would need a different set; the SDK
// does not try to discover them, because they are not exposed over the API.
var (
	activityPubAlgorithms  = []httpsig.Algorithm{httpsig.RSA_SHA256}
	activityPubGetHeaders  = []string{httpsig.RequestTarget, "Date", "Host"}
	activityPubPostHeaders = []string{httpsig.RequestTarget, "Date", "Host", "Digest"}
)

// ActivityPubSigner holds the RSA key a client signs its /activitypub/
// requests with, and the key ID that tells the server where to find the
// matching public key.
//
// This is the authentication the actor sub-routes (/activitypub/user-id/...
// and /activitypub/repository-id/...) expect. They are not token-protected
// API endpoints: they answer a *remote federated server*, and Forgejo
// authenticates that peer with an HTTP Signature over the Cavage draft
// scheme, not with the client's access token. A token- or password-
// authenticated request to them is rejected with "request signature
// verification failed".
//
// keyID must be the URL of the "publicKey" of an ActivityStreams actor
// document the Forgejo server can fetch over HTTP, conventionally
// "<actor-id>#main-key". On receiving a signed request Forgejo GETs that URL
// and requires the document to be a Person or Application actor whose
// publicKey.id equals keyID and whose publicKey.publicKeyPem is the PEM of
// the key used here (PublicKeyPEM returns exactly that). Publishing that
// actor document is the caller's job -- it belongs to the federated instance
// the caller speaks for, not to this HTTP client.
//
// Note that Forgejo stores a peer's key the first time it resolves keyID and
// keeps serving it from the database afterwards: changing the key behind a
// key ID it has already seen makes every later signature fail. Give a new
// key a new key ID.
type ActivityPubSigner struct {
	keyID string
	key   *rsa.PrivateKey
}

// NewActivityPubSigner returns a signer for the given key ID and RSA private
// key. The key is PEM encoded, in either PKCS#1 ("RSA PRIVATE KEY") or
// PKCS#8 ("PRIVATE KEY") form.
func NewActivityPubSigner(keyID string, privateKeyPEM []byte) (*ActivityPubSigner, error) {
	if strings.TrimSpace(keyID) == "" {
		return nil, errors.New("activitypub: empty key ID")
	}

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("activitypub: private key is not PEM encoded")
	}

	var key *rsa.PrivateKey
	switch block.Type {
	case "RSA PRIVATE KEY":
		parsed, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("activitypub: parsing the PKCS#1 private key: %w", err)
		}
		key = parsed
	case "PRIVATE KEY":
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("activitypub: parsing the PKCS#8 private key: %w", err)
		}
		rsaKey, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("activitypub: the private key is %T, not RSA", parsed)
		}
		key = rsaKey
	default:
		return nil, fmt.Errorf("activitypub: unexpected PEM block %q, want an RSA private key", block.Type)
	}

	return &ActivityPubSigner{keyID: keyID, key: key}, nil
}

// KeyID returns the key ID the signer signs with.
func (s *ActivityPubSigner) KeyID() string {
	return s.keyID
}

// PublicKeyPEM returns the PKIX PEM encoding of the signer's public key: the
// exact string that has to appear as publicKey.publicKeyPem in the actor
// document published at KeyID, for the server to accept the signature.
func (s *ActivityPubSigner) PublicKeyPEM() (string, error) {
	der, err := x509.MarshalPKIXPublicKey(&s.key.PublicKey)
	if err != nil {
		return "", fmt.Errorf("activitypub: encoding the public key: %w", err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})), nil
}

// UseActivityPubSignature is an option for NewClient to sign the client's
// ActivityPub requests as a federated actor. See ActivityPubSigner for what
// keyID has to point at.
func UseActivityPubSignature(keyID string, privateKeyPEM []byte) ClientOption {
	return func(client *Client) error {
		signer, err := NewActivityPubSigner(keyID, privateKeyPEM)
		if err != nil {
			return err
		}
		client.SetActivityPubSigner(signer)
		return nil
	}
}

// SetActivityPubSigner sets (or, with nil, clears) the signer used for the
// client's ActivityPub requests.
func (c *Client) SetActivityPubSigner(signer *ActivityPubSigner) {
	c.mutex.Lock()
	c.apSigner = signer
	c.mutex.Unlock()
}

// isActivityPubPath reports whether an API path addresses a federation
// route. Only those are signed: the signature names a remote actor, which
// means nothing on the rest of the API, and Forgejo's ordinary auth chain
// logs a failed authentication attempt for every unusable Signature header
// it is handed.
func isActivityPubPath(path string) bool {
	return strings.HasPrefix(path, "/activitypub/")
}

// signActivityPubRequest adds the Digest (on a request with a body) and
// Signature headers, over the same header set Forgejo signs with.
func (c *Client) signActivityPubRequest(signer *ActivityPubSigner, req *http.Request) error {
	var body []byte
	headers := activityPubGetHeaders

	if req.Body != nil && req.GetBody != nil {
		rc, err := req.GetBody()
		if err != nil {
			return fmt.Errorf("activitypub: rereading the request body: %w", err)
		}
		defer rc.Close()

		body, err = io.ReadAll(rc)
		if err != nil {
			return fmt.Errorf("activitypub: reading the request body: %w", err)
		}
		headers = activityPubPostHeaders
	}

	// httpsig signs whatever it finds in the header map, and Go keeps the
	// Host header in its own field rather than there, so it has to be put
	// back for the signature to cover it.
	req.Header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	req.Header.Set("Host", req.URL.Host)

	httpSigner, _, err := httpsig.NewSigner(
		activityPubAlgorithms,
		httpsig.DigestSha256,
		headers,
		httpsig.Signature,
		activityPubSignatureExpiry,
	)
	if err != nil {
		return fmt.Errorf("activitypub: httpsig.NewSigner: %w", err)
	}

	if err := httpSigner.SignRequest(signer.key, signer.keyID, req, body); err != nil {
		return fmt.Errorf("activitypub: signing the request: %w", err)
	}
	return nil
}
