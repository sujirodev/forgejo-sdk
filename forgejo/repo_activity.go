// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"net/url"
	"time"
)

// Activity is an activity feed item for a repository, user or organization
type Activity struct {
	ID     int64 `json:"id"`
	UserID int64 `json:"user_id"`
	// OpType is the type of action, e.g. "create_repo", "close_issue", "merge_pull_request", ...
	OpType    string      `json:"op_type"`
	ActUserID int64       `json:"act_user_id"`
	ActUser   *User       `json:"act_user"`
	RepoID    int64       `json:"repo_id"`
	Repo      *Repository `json:"repo"`
	CommentID int64       `json:"comment_id"`
	Comment   *Comment    `json:"comment"`
	RefName   string      `json:"ref_name"`
	IsPrivate bool        `json:"is_private"`
	Content   string      `json:"content"`
	Created   time.Time   `json:"created"`
}

// ListRepoActivityFeedsOptions options for listing a repository's activity feeds
type ListRepoActivityFeedsOptions struct {
	ListOptions
	// Date restricts the feed to activities on this date, in YYYY-MM-DD format
	Date string
}

// QueryEncode turns options into querystring argument
func (opt *ListRepoActivityFeedsOptions) QueryEncode() string {
	query := opt.getURLQuery()
	if len(opt.Date) > 0 {
		query.Add("date", opt.Date)
	}
	return query.Encode()
}

// ListRepoActivityFeeds lists a repository's activity feeds
func (c *Client) ListRepoActivityFeeds(owner, repo string, opt ListRepoActivityFeedsOptions) ([]*Activity, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()
	link, _ := url.Parse(fmt.Sprintf("/repos/%s/%s/activities/feeds", owner, repo))
	link.RawQuery = opt.QueryEncode()
	activities := make([]*Activity, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", link.String(), jsonHeader, nil, &activities)
	return activities, resp, err
}
