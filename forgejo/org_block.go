// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"context"
	"fmt"
)

// OrgBlockUser blocks a user from the organization
func (c *Client) OrgBlockUser(ctx context.Context, org, username string) (Response, error) {
	if err := escapeValidatePathSegments(&org, &username); err != nil {
		return Response{}, err
	}

	_, resp, err := c.getResponseWithContext(ctx, "PUT", fmt.Sprintf("/orgs/%s/block/%s", org, username), nil, nil)
	return resp, err
}

// OrgUnblockUser unblocks a user from the organization
func (c *Client) OrgUnblockUser(ctx context.Context, org, username string) (Response, error) {
	if err := escapeValidatePathSegments(&org, &username); err != nil {
		return Response{}, err
	}

	_, resp, err := c.getResponseWithContext(ctx, "PUT", fmt.Sprintf("/orgs/%s/unblock/%s", org, username), nil, nil)
	return resp, err
}

// ListOrgBlockedUsersOptions holds optional parameters for listing an organization's blocked users
type ListOrgBlockedUsersOptions struct {
	ListOptions
}

// ListOrgBlockedUsers lists the organization's blocked users
func (c *Client) ListOrgBlockedUsers(ctx context.Context, org string, opt ListOrgBlockedUsersOptions) ([]*BlockedUser, Response, error) {
	if err := escapeValidatePathSegments(&org); err != nil {
		return nil, Response{}, err
	}
	opt.setDefaults()

	var blocked []*BlockedUser
	resp, err := c.getParsedResponseWithContext(ctx, fmt.Sprintf("/orgs/%s/list_blocked?%s", org, opt.getURLQuery().Encode()), &blocked)
	return blocked, resp, err
}
