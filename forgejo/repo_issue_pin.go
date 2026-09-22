// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
)

// NewIssuePinsAllowed tells whether new issue or pull request pins are allowed for a repository
type NewIssuePinsAllowed struct {
	Issues       bool `json:"issues"`
	PullRequests bool `json:"pull_requests"`
}

// ListPinnedIssues lists a repo's pinned issues
func (c *Client) ListPinnedIssues(owner, repo string) ([]*Issue, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	issues := make([]*Issue, 0, 5)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/issues/pinned", owner, repo), jsonHeader, nil, &issues)
	return issues, resp, err
}

// NewPinAllowed returns whether new issue and pull request pins are allowed for a repository
func (c *Client) NewPinAllowed(owner, repo string) (*NewIssuePinsAllowed, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	allowed := new(NewIssuePinsAllowed)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/new_pin_allowed", owner, repo), jsonHeader, nil, allowed)
	return allowed, resp, err
}
