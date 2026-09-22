// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// UpdateOrgAvatarOption options for updating an organization's avatar
type UpdateOrgAvatarOption struct {
	// Image must be base64 encoded
	Image string `json:"image"`
}

// UpdateOrgAvatar updates an organization's avatar
func (c *Client) UpdateOrgAvatar(ctx context.Context, org string, opt UpdateOrgAvatarOption) (Response, error) {
	if err := escapeValidatePathSegments(&org); err != nil {
		return Response{}, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return Response{}, err
	}

	_, resp, err := c.getResponseWithContext(ctx, "POST", fmt.Sprintf("/orgs/%s/avatar", org), jsonHeader, bytes.NewReader(body))
	return resp, err
}

// DeleteOrgAvatar deletes an organization's avatar. It will be replaced by a default one.
func (c *Client) DeleteOrgAvatar(ctx context.Context, org string) (Response, error) {
	if err := escapeValidatePathSegments(&org); err != nil {
		return Response{}, err
	}

	_, resp, err := c.getResponseWithContext(ctx, "DELETE", fmt.Sprintf("/orgs/%s/avatar", org), jsonHeader, nil)
	return resp, err
}
