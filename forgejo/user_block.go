// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"time"
)

// BlockedUser represents a user blocked by the authenticated user.
type BlockedUser struct {
	BlockID int64     `json:"block_id"`
	Created time.Time `json:"created_at"`
}

// ListBlockedUsersOptions options for listing the users blocked by the
// authenticated user.
type ListBlockedUsersOptions struct {
	ListOptions
}

// ListBlockedUsers lists the users blocked by the currently authenticated user.
func (c *Client) ListBlockedUsers(opt ListBlockedUsersOptions) ([]*BlockedUser, *Response, error) {
	opt.setDefaults()
	blocked := make([]*BlockedUser, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/user/list_blocked?%s", opt.getURLQuery().Encode()), jsonHeader, nil, &blocked)
	return blocked, resp, err
}

// BlockUser blocks the given user from interacting with the authenticated user.
func (c *Client) BlockUser(username string) (*Response, error) {
	if err := escapeValidatePathSegments(&username); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("PUT", fmt.Sprintf("/user/block/%s", username), jsonHeader, nil)
	return resp, err
}

// UnblockUser removes a block placed on the given user by the authenticated user.
func (c *Client) UnblockUser(username string) (*Response, error) {
	if err := escapeValidatePathSegments(&username); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("PUT", fmt.Sprintf("/user/unblock/%s", username), jsonHeader, nil)
	return resp, err
}
