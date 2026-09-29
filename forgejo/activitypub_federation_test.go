// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// The ActivityPub actor sub-routes answer a *remote federated server*, and
// authenticate it with an HTTP Signature whose key ID Forgejo resolves by
// fetching it over HTTP (services/federation/signature_service.go,
// FindOrCreateActorKey). So a test that wants a real 2xx out of them has to
// be a federated server: it has to serve an actor document holding the
// public key, plus the nodeinfo documents Forgejo reads on the way there.
//
// That is what federatedActor is. It is a throwaway peer, unique per test
// run, that exists only so the Forgejo instance under test can resolve the
// suite's signing key.

// federatedActor is a minimal ActivityPub instance the suite speaks for: a
// single Person actor publishing one RSA key, plus the two nodeinfo
// documents Forgejo insists on before it will believe in that Person.
type federatedActor struct {
	// hostPort is the authority the Forgejo instance under test reaches
	// this server at, e.g. "172.17.0.4:41293". It is not necessarily the
	// address the listener is bound to.
	hostPort string

	// id is unique per run. Forgejo caches a peer's key under its key ID
	// and creates a local user named after preferredUsername, so a fixed
	// identity would make the second run against a surviving instance fail
	// with either a stale key or "user already exists".
	id string

	privatePEM []byte
	publicPEM  string

	// fetched counts the requests Forgejo made to this server, which is how
	// the candidate loop below tells "reachable" from "not reachable".
	fetched atomic.Int32
}

func (a *federatedActor) baseURL() string { return "http://" + a.hostPort }
func (a *federatedActor) personURI() string {
	// Forgejo validates a Person's URI against its own route shape
	// (modules/forgefed/actor_person.go, personIDapiPathV1): anything not
	// under api/v1/activitypub/user-id is rejected before the key is read.
	return fmt.Sprintf("%s/api/v1/activitypub/user-id/%s", a.baseURL(), a.id)
}
func (a *federatedActor) keyID() string { return a.personURI() + "#main-key" }

// handler serves the documents Forgejo fetches, deriving every URI from the
// request's Host header so that one listener can answer for whichever
// candidate address turns out to be the reachable one.
func (a *federatedActor) handler() http.Handler {
	mux := http.NewServeMux()
	count := func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			a.fetched.Add(1)
			h(w, r)
		}
	}
	base := func(r *http.Request) string { return "http://" + r.Host }

	mux.HandleFunc("/.well-known/nodeinfo", count(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"links":[{"rel":"http://nodeinfo.diaspora.software/ns/schema/2.1","href":%q}]}`,
			base(r)+"/api/v1/nodeinfo")
	}))

	mux.HandleFunc("/api/v1/nodeinfo", count(func(w http.ResponseWriter, r *http.Request) {
		// "forgejo" is one of the four software names Forgejo federates
		// with (models/forgefed/nodeinfo.go, KnownSourceTypes); an unknown
		// one fails validation before the actor is ever fetched.
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"version":"2.1","software":{"name":"forgejo","version":"`+testForgejoVersion+`"},`+
			`"protocols":["activitypub"],"services":{"inbound":[],"outbound":[]},`+
			`"openRegistrations":false,"usage":{"users":{"total":1}},"metadata":{}}`)
	}))

	mux.HandleFunc("/api/v1/activitypub/user-id/"+a.id, count(func(w http.ResponseWriter, r *http.Request) {
		person := fmt.Sprintf("%s/api/v1/activitypub/user-id/%s", base(r), a.id)
		w.Header().Set("Content-Type", ActivityPubContentType)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"@context":          []any{"https://www.w3.org/ns/activitystreams", "https://w3id.org/security/v1"},
			"id":                person,
			"type":              "Person",
			"preferredUsername": "sdktest" + a.id,
			"name":              "sdktest" + a.id,
			"inbox":             person + "/inbox",
			"outbox":            person + "/outbox",
			"publicKey": map[string]any{
				"id":           person + "#main-key",
				"owner":        person,
				"publicKeyPem": a.publicPEM,
			},
		})
	}))

	mux.HandleFunc("/api/v1/activitypub/user-id/"+a.id+"/inbox", count(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	return mux
}

// followActivity is a well-formed Follow from this actor to object, the
// shape Forgejo's inbox handlers validate against (ForgeFollow).
func (a *federatedActor) followActivity(object string) ActivityPubObject {
	return ActivityPubObject{
		"@context": "https://www.w3.org/ns/activitystreams",
		"id":       a.personURI() + "/follows/" + a.id,
		"type":     "Follow",
		"actor":    a.personURI(),
		"object":   object,
	}
}

// likeActivity is a well-formed Like, which is the only activity type a
// repository inbox accepts (ForgeLike).
func (a *federatedActor) likeActivity(object string) ActivityPubObject {
	return ActivityPubObject{
		"@context":  "https://www.w3.org/ns/activitystreams",
		"id":        a.personURI() + "/likes/" + a.id,
		"type":      "Like",
		"actor":     a.personURI(),
		"object":    object,
		"startTime": time.Now().UTC().Format(time.RFC3339),
	}
}

var (
	federationOnce  sync.Once
	federationActor *federatedActor
	federationWhy   string // why there is no actor, for the skip message
)

// signingTestClient returns an integration client that signs its ActivityPub
// requests as the shared federated actor, and the actor itself. It skips the
// calling test when this environment cannot host a peer the Forgejo instance
// under test can fetch a key from -- reaching back into the test process is
// the one thing that is not under the suite's control.
func signingTestClient(t *testing.T) (*Client, *federatedActor) {
	t.Helper()

	federationOnce.Do(setUpFederatedActor)
	if federationActor == nil {
		t.Skipf("no federated actor the Forgejo instance can reach: %s", federationWhy)
	}

	c, err := newTestClientOpts(UseActivityPubSignature(federationActor.keyID(), federationActor.privatePEM))
	if err != nil {
		t.Fatalf("building a signing client: %v", err)
	}
	return c, federationActor
}

// setUpFederatedActor starts the peer and picks the address Forgejo can
// reach it at. "Reachable" is not guessed: each candidate is tried with a
// real signed request, and only an address Forgejo actually fetched the
// actor document from is kept.
func setUpFederatedActor() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		federationWhy = fmt.Sprintf("generating an RSA key: %v", err)
		return
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		federationWhy = fmt.Sprintf("encoding the public key: %v", err)
		return
	}

	actor := &federatedActor{
		id:         fmt.Sprintf("%d", time.Now().UnixNano()),
		privatePEM: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}),
		publicPEM:  string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})),
	}

	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		federationWhy = fmt.Sprintf("listening for the federated actor: %v", err)
		return
	}
	go func() { _ = http.Serve(ln, actor.handler()) }() //nolint:gosec // a throwaway test server

	port := fmt.Sprintf("%d", ln.Addr().(*net.TCPAddr).Port)
	candidates := callbackCandidates(port)
	if len(candidates) == 0 {
		federationWhy = "no candidate callback address"
		return
	}

	for _, hostPort := range candidates {
		actor.hostPort = hostPort
		if federatedActorReachable(actor) {
			federationActor = actor
			return
		}
	}
	federationWhy = fmt.Sprintf("none of %v was reachable from %s; set FORGEJO_SDK_TEST_CALLBACK_HOST",
		candidates, getForgejoURL())
}

// callbackCandidates lists the authorities worth trying, most reliable
// first. FORGEJO_SDK_TEST_CALLBACK_HOST overrides the guessing entirely.
func callbackCandidates(port string) []string {
	if host := os.Getenv("FORGEJO_SDK_TEST_CALLBACK_HOST"); host != "" {
		return []string{net.JoinHostPort(host, port)}
	}

	var out []string
	// The address this process presents to the Forgejo instance. On a
	// container network -- which is what CI runs -- that is exactly the
	// address Forgejo can dial back.
	if u, err := url.Parse(getForgejoURL()); err == nil && u.Host != "" {
		target := u.Host
		if u.Port() == "" {
			defaultPort := "80"
			if u.Scheme == "https" {
				defaultPort = "443"
			}
			target = net.JoinHostPort(u.Hostname(), defaultPort)
		}
		if conn, err := net.DialTimeout("tcp", target, 5*time.Second); err == nil {
			local, ok := conn.LocalAddr().(*net.TCPAddr)
			_ = conn.Close()
			// A loopback source address means the two sides do not share a
			// network stack's view of each other (a published container
			// port); Forgejo would dial its own loopback instead of ours.
			if ok && !local.IP.IsLoopback() {
				out = append(out, net.JoinHostPort(local.IP.String(), port))
			}
		}
	}
	// Docker Desktop's name for the host from inside a container, which is
	// how a locally published `make test-instance-docker` instance reaches
	// a suite running on the host.
	out = append(out, net.JoinHostPort("host.docker.internal", port))
	return out
}

// federatedActorReachable makes one signed request to a route that requires
// a signature on every supported Forgejo version, and reports whether
// Forgejo came back to fetch the actor document. The request's own status
// code is deliberately not the signal: it is the callback that proves the
// address works.
func federatedActorReachable(actor *federatedActor) bool {
	c, err := newTestClientOpts(UseActivityPubSignature(actor.keyID(), actor.privatePEM))
	if err != nil {
		return false
	}
	me, _, err := c.GetMyUserInfo()
	if err != nil {
		return false
	}

	before := actor.fetched.Load()
	_, _ = c.SendActivityPubPersonInbox(me.ID, actor.followActivity(
		fmt.Sprintf("%s/api/v1/activitypub/user-id/%d", c.url, me.ID)))
	return actor.fetched.Load() > before
}
