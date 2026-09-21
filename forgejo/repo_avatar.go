// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// UpdateRepoAvatarOption options when updating a repository's avatar
type UpdateRepoAvatarOption struct {
	// image must be base64 encoded
	Image string `json:"image"`
}

// UpdateRepoAvatar updates a repository's avatar
func (c *Client) UpdateRepoAvatar(owner, repo string, opt UpdateRepoAvatarOption) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("POST", fmt.Sprintf("/repos/%s/%s/avatar", owner, repo), jsonHeader, bytes.NewReader(body))
	return resp, err
}

// DeleteRepoAvatar deletes a repository's avatar
func (c *Client) DeleteRepoAvatar(owner, repo string) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/repos/%s/%s/avatar", owner, repo), nil, nil)
	return resp, err
}
