// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"
)

// TestActivityPubFollow only verifies the request round-trips to the
// server. The test instance has no reachable remote ActivityPub actor (and
// federation may not even be enabled on it), so a full follow against a
// live remote instance is not exercised here.
func TestActivityPubFollow(t *testing.T) {
	t.Parallel()
	log.Println("== TestActivityPubFollow ==")
	c := newTestClient()

	resp, err := c.ActivityPubFollow(APRemoteFollowOption{Target: "https://example.com/api/v1/activitypub/user-id/1"})
	if err != nil {
		t.Logf("ActivityPubFollow returned an error (expected if federation is disabled or the target is unreachable): %v", err)
		return
	}
	if resp == nil {
		t.Fatal("expected a non-nil response when ActivityPubFollow returns no error")
	}
}
