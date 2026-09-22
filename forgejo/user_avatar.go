// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"encoding/json"
)

// UpdateUserAvatarOption options for updating the authenticated user's avatar
type UpdateUserAvatarOption struct {
	// Image must be base64 encoded image data.
	Image string `json:"image"`
}

// UpdateUserAvatar updates the avatar of the currently authenticated user.
func (c *Client) UpdateUserAvatar(opt UpdateUserAvatarOption) (*Response, error) {
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("POST", "/user/avatar", jsonHeader, bytes.NewReader(body))
	return resp, err
}

// DeleteUserAvatar deletes the avatar of the currently authenticated user.
// It will be replaced by a default one.
func (c *Client) DeleteUserAvatar() (*Response, error) {
	_, resp, err := c.getResponse("DELETE", "/user/avatar", jsonHeader, nil)
	return resp, err
}
