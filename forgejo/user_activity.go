// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"net/url"
)

// ListUserActivityFeeds lists a user's activity feeds.
func (c *Client) ListUserActivityFeeds(username string, opt ListActivityFeedsOptions) ([]*Activity, *Response, error) {
	if err := escapeValidatePathSegments(&username); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()

	link, _ := url.Parse(fmt.Sprintf("/users/%s/activities/feeds", username))
	link.RawQuery = opt.QueryEncode()

	feeds := make([]*Activity, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", link.String(), jsonHeader, nil, &feeds)
	return feeds, resp, err
}

// UserHeatmapData represents one point of a user's contribution heatmap:
// the number of contributions made at a given Unix timestamp.
type UserHeatmapData struct {
	Timestamp     int64 `json:"timestamp"`
	Contributions int64 `json:"contributions"`
}

// GetUserHeatmapData gets a user's heatmap data.
func (c *Client) GetUserHeatmapData(username string) ([]*UserHeatmapData, *Response, error) {
	if err := escapeValidatePathSegments(&username); err != nil {
		return nil, nil, err
	}
	heatmap := make([]*UserHeatmapData, 0)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/users/%s/heatmap", username), jsonHeader, nil, &heatmap)
	return heatmap, resp, err
}
