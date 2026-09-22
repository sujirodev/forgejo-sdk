// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import "fmt"

// PinIssue pins an issue
func (c *Client) PinIssue(owner, repo string, index int64) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("POST", fmt.Sprintf("/repos/%s/%s/issues/%d/pin", owner, repo, index), nil, nil)
	return resp, err
}

// UnpinIssue unpins an issue
func (c *Client) UnpinIssue(owner, repo string, index int64) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/repos/%s/%s/issues/%d/pin", owner, repo, index), nil, nil)
	return resp, err
}

// MoveIssuePin moves the pinned issue to the given position among the repository's pinned issues
func (c *Client) MoveIssuePin(owner, repo string, index, position int64) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("PATCH", fmt.Sprintf("/repos/%s/%s/issues/%d/pin/%d", owner, repo, index, position), nil, nil)
	return resp, err
}
