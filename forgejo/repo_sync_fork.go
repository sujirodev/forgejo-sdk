// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
)

// SyncForkInfo has information about syncing a fork branch with its base branch
type SyncForkInfo struct {
	Allowed       bool   `json:"allowed"`
	ForkCommit    string `json:"fork_commit"`
	BaseCommit    string `json:"base_commit"`
	CommitsBehind int    `json:"commits_behind"`
}

// GetSyncForkDefaultInfo gets information about syncing a fork's default branch with its base branch
func (c *Client) GetSyncForkDefaultInfo(owner, repo string) (*SyncForkInfo, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	info := new(SyncForkInfo)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/sync_fork", owner, repo), jsonHeader, nil, info)
	return info, resp, err
}

// SyncForkDefault syncs a fork's default branch with its base branch
func (c *Client) SyncForkDefault(owner, repo string) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("POST", fmt.Sprintf("/repos/%s/%s/sync_fork", owner, repo), nil, nil)
	return resp, err
}

// GetSyncForkBranchInfo gets information about syncing a fork branch with its base branch
func (c *Client) GetSyncForkBranchInfo(owner, repo, branch string) (*SyncForkInfo, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo, &branch); err != nil {
		return nil, nil, err
	}
	info := new(SyncForkInfo)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/sync_fork/%s", owner, repo, branch), jsonHeader, nil, info)
	return info, resp, err
}

// SyncForkBranch syncs a fork branch with its base branch
func (c *Client) SyncForkBranch(owner, repo, branch string) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo, &branch); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("POST", fmt.Sprintf("/repos/%s/%s/sync_fork/%s", owner, repo, branch), nil, nil)
	return resp, err
}
