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

func TestListUserActivityFeeds(t *testing.T) {
	log.Println("== TestListUserActivityFeeds ==")
	c := newTestClient()

	me, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	// The freshly created test user has no activity yet, so this only
	// verifies the call succeeds and returns a (possibly empty) list, not
	// any specific content.
	feeds, resp, err := c.ListUserActivityFeeds(me.UserName, ListActivityFeedsOptions{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotNil(t, feeds) //nolint:staticcheck // feeds may legitimately be empty
}

func TestGetUserHeatmapData(t *testing.T) {
	log.Println("== TestGetUserHeatmapData ==")
	c := newTestClient()

	me, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	heatmap, resp, err := c.GetUserHeatmapData(me.UserName)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, heatmap)
}
