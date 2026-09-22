// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"strings"
)

// GitBlobResponse represents a git blob
type GitBlobResponse struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
	URL      string `json:"url"`
	SHA      string `json:"sha"`
	Size     int64  `json:"size"`
}

// GetBlob get the blob of a repository file
func (c *Client) GetBlob(user, repo, sha string) (*GitBlobResponse, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo, &sha); err != nil {
		return nil, nil, err
	}
	blob := new(GitBlobResponse)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/git/blobs/%s", user, repo, sha), nil, nil, blob)
	return blob, resp, err
}

// GetBlobs gets multiple blobs of a repository, given a list of blob SHAs
func (c *Client) GetBlobs(user, repo string, shas []string) ([]*GitBlobResponse, *Response, error) {
	// Confirmed absent on a live 11.0.16 instance (404) and present by
	// 15.0.9.
	if err := c.checkServerVersionGreaterThanOrEqual(version15_0_0); err != nil {
		return nil, nil, err
	}
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	blobs := make([]*GitBlobResponse, 0, len(shas))
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/git/blobs?shas=%s", user, repo, strings.Join(shas, ",")), nil, nil, &blobs)
	return blobs, resp, err
}
