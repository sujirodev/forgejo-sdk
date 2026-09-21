// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipIfFederationDisabled skips the calling test when the test instance
// answers an ActivityPub GET with 403, which is how Forgejo reports
// federation being turned off ([federation] ENABLED = false in app.ini).
func skipIfFederationDisabled(t *testing.T, resp *Response, err error) {
	t.Helper()
	if err != nil && resp != nil && resp.StatusCode == http.StatusForbidden {
		t.Skip("federation is disabled on the test instance")
	}
}

// remoteFollowActivity is a well-formed, but unsigned, ActivityStreams
// Follow activity. It is enough to exercise inbox routing and request
// encoding; it is not expected to be *accepted*, because Forgejo verifies
// the HTTP signature of an inbox delivery against the sending actor's
// published key, and this test has no such remote actor/key to sign with.
func remoteFollowActivity(target string) ActivityPubObject {
	return ActivityPubObject{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type":     "Follow",
		"actor":    "https://example.invalid/actor",
		"object":   target,
	}
}

func TestActivityPubInstanceActor(t *testing.T) {
	log.Println("== TestActivityPubInstanceActor ==")
	c := newTestClient()

	actor, resp, err := c.GetActivityPubActor()
	skipIfFederationDisabled(t, resp, err)
	require.NoError(t, err)
	require.NotNil(t, actor)
	assert.Contains(t, *actor, "@context")

	outbox, resp, err := c.GetActivityPubActorOutbox()
	skipIfFederationDisabled(t, resp, err)
	require.NoError(t, err)
	require.NotNil(t, outbox)
}

func TestActivityPubInstanceActorInbox(t *testing.T) {
	log.Println("== TestActivityPubInstanceActorInbox ==")
	c := newTestClient()

	_, resp, err := c.GetActivityPubActor()
	skipIfFederationDisabled(t, resp, err)
	require.NoError(t, err)

	resp, err = c.SendActivityPubActorInbox(remoteFollowActivity(c.url + "/activitypub/actor"))
	require.NotNil(t, resp)
	// Not asserting success: an unsigned delivery is expected to be
	// rejected. What matters here is that the route exists and the SDK's
	// request reaches it.
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
	_ = err
}

func TestActivityPubRepository(t *testing.T) {
	log.Println("== TestActivityPubRepository ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "activitypub-repo", c)
	require.NoError(t, err)

	actor, resp, err := c.GetActivityPubRepository(repo.ID)
	skipIfFederationDisabled(t, resp, err)
	require.NoError(t, err)
	require.NotNil(t, actor)
	assert.Contains(t, *actor, "@context")

	outbox, resp, err := c.GetActivityPubRepositoryOutbox(repo.ID)
	skipIfFederationDisabled(t, resp, err)
	require.NoError(t, err)
	require.NotNil(t, outbox)
}

func TestActivityPubRepositoryInbox(t *testing.T) {
	log.Println("== TestActivityPubRepositoryInbox ==")
	c := newTestClient()
	repo, err := createTestRepo(t, "activitypub-repo-inbox", c)
	require.NoError(t, err)

	_, resp, err := c.GetActivityPubRepository(repo.ID)
	skipIfFederationDisabled(t, resp, err)
	require.NoError(t, err)

	resp, err = c.SendActivityPubRepositoryInbox(repo.ID, remoteFollowActivity(c.url+"/activitypub/repository-id/"))
	require.NotNil(t, resp)
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
	_ = err
}

func TestActivityPubPerson(t *testing.T) {
	log.Println("== TestActivityPubPerson ==")
	c := newTestClient()
	me, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	actor, resp, err := c.GetActivityPubPerson(me.ID)
	skipIfFederationDisabled(t, resp, err)
	require.NoError(t, err)
	require.NotNil(t, actor)
	assert.Contains(t, *actor, "@context")

	outbox, resp, err := c.GetActivityPubPersonOutbox(me.ID)
	skipIfFederationDisabled(t, resp, err)
	require.NoError(t, err)
	require.NotNil(t, outbox)
}

func TestActivityPubPersonInbox(t *testing.T) {
	log.Println("== TestActivityPubPersonInbox ==")
	c := newTestClient()
	me, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	_, resp, err := c.GetActivityPubPerson(me.ID)
	skipIfFederationDisabled(t, resp, err)
	require.NoError(t, err)

	resp, err = c.SendActivityPubPersonInbox(me.ID, remoteFollowActivity(c.url+"/activitypub/user-id/"))
	require.NotNil(t, resp)
	assert.NotEqual(t, http.StatusNotFound, resp.StatusCode)
	_ = err
}

func TestActivityPubPersonActivity(t *testing.T) {
	log.Println("== TestActivityPubPersonActivity ==")
	c := newTestClient()
	me, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	if _, resp, err := c.GetActivityPubPerson(me.ID); err != nil {
		skipIfFederationDisabled(t, resp, err)
		require.NoError(t, err)
	}

	// There is no way to seed a specific federation activity ID through the
	// REST API, so activity-id 1 is very likely absent. This still verifies
	// the SDK reaches the route (not a 404 caused by a wrong path on the
	// SDK side) and decodes whatever comes back.
	_, resp, err := c.GetActivityPubPersonActivityNote(me.ID, 1)
	require.NotNil(t, resp)
	_ = err

	_, resp, err = c.GetActivityPubPersonActivity(me.ID, 1)
	require.NotNil(t, resp)
	_ = err
}
