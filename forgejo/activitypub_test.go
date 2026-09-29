// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
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

// requireActor asserts that an ActivityPub GET succeeded and decoded into an
// ActivityStreams document.
func requireActor(t *testing.T, obj *ActivityPubObject, resp *Response, err error) {
	t.Helper()
	skipIfFederationDisabled(t, resp, err)
	require.NoError(t, err)
	require.NotNil(t, obj)
	assert.Contains(t, *obj, "@context")
}

// requireDelivered asserts that an inbox accepted a signed delivery. The SDK
// turns any non-2xx into an error, so a nil error here already means the
// activity was taken.
func requireDelivered(t *testing.T, resp *Response, err error) {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestActivityPubInstanceActor(t *testing.T) {
	t.Parallel()
	log.Println("== TestActivityPubInstanceActor ==")
	c, _ := signingTestClient(t)

	actor, resp, err := c.GetActivityPubActor()
	requireActor(t, actor, resp, err)

	// /activitypub/actor/outbox arrived in Forgejo 14.0.0; below that the
	// SDK's guard refuses the call, which is the documented behavior.
	if serverAtLeast(t, c, "14.0.0") {
		outbox, resp, err := c.GetActivityPubActorOutbox()
		requireActor(t, outbox, resp, err)
	} else {
		_, _, err := c.GetActivityPubActorOutbox()
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
	}
}

func TestActivityPubInstanceActorInbox(t *testing.T) {
	t.Parallel()
	log.Println("== TestActivityPubInstanceActorInbox ==")
	c, peer := signingTestClient(t)

	actor, resp, err := c.GetActivityPubActor()
	requireActor(t, actor, resp, err)

	resp, err = c.SendActivityPubActorInbox(peer.followActivity(c.url + "/api/v1/activitypub/actor"))
	requireDelivered(t, resp, err)
}

func TestActivityPubRepository(t *testing.T) {
	t.Parallel()
	log.Println("== TestActivityPubRepository ==")
	c, _ := signingTestClient(t)
	repo, err := createTestRepo(t, "activitypub-repo", c)
	require.NoError(t, err)

	actor, resp, err := c.GetActivityPubRepository(repo.ID)
	requireActor(t, actor, resp, err)

	// /activitypub/repository-id/{id}/outbox arrived in Forgejo 14.0.0.
	if serverAtLeast(t, c, "14.0.0") {
		outbox, resp, err := c.GetActivityPubRepositoryOutbox(repo.ID)
		requireActor(t, outbox, resp, err)
	} else {
		_, _, err := c.GetActivityPubRepositoryOutbox(repo.ID)
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
	}
}

func TestActivityPubRepositoryInbox(t *testing.T) {
	t.Parallel()
	log.Println("== TestActivityPubRepositoryInbox ==")
	c, peer := signingTestClient(t)
	repo, err := createTestRepo(t, "activitypub-repo-inbox", c)
	require.NoError(t, err)

	actor, resp, err := c.GetActivityPubRepository(repo.ID)
	requireActor(t, actor, resp, err)

	// A repository inbox only accepts Like (forgefed.ForgeLike); a Follow
	// there is refused as a validation error, not as a routing failure.
	resp, err = c.SendActivityPubRepositoryInbox(repo.ID, peer.likeActivity(
		fmt.Sprintf("%s/api/v1/activitypub/repository-id/%d", c.url, repo.ID)))
	requireDelivered(t, resp, err)
}

func TestActivityPubPerson(t *testing.T) {
	t.Parallel()
	log.Println("== TestActivityPubPerson ==")
	c, _ := signingTestClient(t)
	me, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	actor, resp, err := c.GetActivityPubPerson(me.ID)
	requireActor(t, actor, resp, err)

	// /activitypub/user-id/{id}/outbox arrived in Forgejo 13.0.0.
	if serverAtLeast(t, c, "13.0.0") {
		outbox, resp, err := c.GetActivityPubPersonOutbox(me.ID)
		requireActor(t, outbox, resp, err)
	} else {
		_, _, err := c.GetActivityPubPersonOutbox(me.ID)
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
	}
}

func TestActivityPubPersonInbox(t *testing.T) {
	t.Parallel()
	log.Println("== TestActivityPubPersonInbox ==")
	c, peer := signingTestClient(t)
	me, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	actor, resp, err := c.GetActivityPubPerson(me.ID)
	requireActor(t, actor, resp, err)

	resp, err = c.SendActivityPubPersonInbox(me.ID, peer.followActivity(
		fmt.Sprintf("%s/api/v1/activitypub/user-id/%d", c.url, me.ID)))
	requireDelivered(t, resp, err)
}

func TestActivityPubPersonActivity(t *testing.T) {
	t.Parallel()
	log.Println("== TestActivityPubPersonActivity ==")
	c, _ := signingTestClient(t)
	me, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	// The activity routes arrived in Forgejo 13.0.0.
	if !serverAtLeast(t, c, "13.0.0") {
		_, _, err := c.GetActivityPubPersonActivityNote(me.ID, 1)
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
		_, _, err = c.GetActivityPubPersonActivity(me.ID, 1)
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
		return
	}

	// These routes read an ordinary activity feed entry, not a federation
	// specific one (routers/api/v1/activitypub/person.go, getActivity), and
	// only one the user performed on themselves in a public repository. So
	// seed one by creating a public repo, then ask the feed for its ID:
	// there is no fixed activity ID to hard-code, and a missing one is a
	// 404 rather than an empty document.
	activityID := seedPersonActivity(t, c, me.UserName)

	note, resp, err := c.GetActivityPubPersonActivityNote(me.ID, activityID)
	requireActor(t, note, resp, err)

	activity, resp, err := c.GetActivityPubPersonActivity(me.ID, activityID)
	requireActor(t, activity, resp, err)
}

// seedPersonActivity creates a public repository and returns the ID of the
// resulting feed entry, which is what the ActivityPub activity routes serve.
func seedPersonActivity(t *testing.T, c *Client, username string) int64 {
	t.Helper()

	_, err := createTestRepo(t, "activitypub-activity", c)
	require.NoError(t, err)

	feeds, _, err := c.ListUserActivityFeeds(username, ListActivityFeedsOptions{
		OnlyPerformedBy: true,
		ListOptions:     ListOptions{PageSize: 20},
	})
	require.NoError(t, err)

	for _, f := range feeds {
		// getActivity serves only an action the user performed on their own
		// feed, and only a public one.
		if !f.IsPrivate && f.UserID == f.ActUserID {
			return f.ID
		}
	}
	t.Skip("no public self-performed activity in the user's feed to read back")
	return 0
}
