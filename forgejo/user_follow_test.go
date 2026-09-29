// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestActivityPubFollow follows the throwaway federated peer the suite
// stands up (see activitypub_federation_test.go). That peer is a real
// remote actor as far as the instance under test is concerned: the server
// reads its nodeinfo, fetches its actor document and creates a local
// federated user for it before answering, so a 204 here is evidence the
// whole resolution path ran -- an actor it cannot reach gets a 406.
func TestActivityPubFollow(t *testing.T) {
	t.Parallel()
	log.Println("== TestActivityPubFollow ==")
	c, peer := signingTestClient(t)

	// POST /user/activitypub/follow arrived in Forgejo 16.0.0; below that
	// the SDK's guard refuses the call, which is the documented behavior.
	if !serverAtLeast(t, c, "16.0.0") {
		_, err := c.ActivityPubFollow(APRemoteFollowOption{Target: peer.personURI()})
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
		return
	}

	resp, err := c.ActivityPubFollow(APRemoteFollowOption{Target: peer.personURI()})
	require.NoError(t, err)
	require.NotNil(t, resp)
}
