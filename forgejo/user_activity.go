// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"net/url"
	"time"
)

// Activity represents a single entry in a user's activity feed: an action
// the user performed, or that happened on something visible to them (such
// as a repository they follow).
type Activity struct {
	ID        int64       `json:"id"`
	UserID    int64       `json:"user_id"`
	OpType    string      `json:"op_type"`
	ActUserID int64       `json:"act_user_id"`
	ActUser   *User       `json:"act_user"`
	RepoID    int64       `json:"repo_id"`
	Repo      *Repository `json:"repo"`
	CommentID int64       `json:"comment_id"`
	Comment   *Comment    `json:"comment"`
	IsPrivate bool        `json:"is_private"`
	Content   string      `json:"content"`
	Created   time.Time   `json:"created"`
	RefName   string      `json:"ref_name"`
}

// ListActivityFeedsOptions options for listing a user's activity feeds
type ListActivityFeedsOptions struct {
	ListOptions
	// OnlyPerformedBy restricts the feed to actions performed by the user
	// themselves.
	OnlyPerformedBy bool
	// Date restricts the feed to activities on this date, formatted as
	// YYYY-MM-DD.
	Date string
}

// QueryEncode encodes options to query parameters
func (opt *ListActivityFeedsOptions) QueryEncode() string {
	query := opt.getURLQuery()
	if opt.OnlyPerformedBy {
		query.Add("only-performed-by", "true")
	}
	if opt.Date != "" {
		query.Add("date", opt.Date)
	}
	return query.Encode()
}

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
