// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"net/url"
)

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
