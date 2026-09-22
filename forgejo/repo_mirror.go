// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2023 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
)

type CreatePushMirrorOption struct {
	Interval       string `json:"interval"`
	RemoteAddress  string `json:"remote_address"`
	RemotePassword string `json:"remote_password"`
	RemoteUsername string `json:"remote_username"`
	SyncONCommit   bool   `json:"sync_on_commit"`
}

// PushMirrorResponse returns a git push mirror
type PushMirrorResponse struct {
	Created       string `json:"created"`
	Interval      string `json:"interval"`
	LastError     string `json:"last_error"`
	LastUpdate    string `json:"last_update"`
	RemoteAddress string `json:"remote_address"`
	RemoteName    string `json:"remote_name"`
	RepoName      string `json:"repo_name"`
	SyncONCommit  bool   `json:"sync_on_commit"`
}

// PushMirrors add a push mirror to the repository
func (c *Client) PushMirrors(user, repo string, opt CreatePushMirrorOption) (*PushMirrorResponse, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(opt)
	if err != nil {
		return nil, nil, err
	}
	pm := new(PushMirrorResponse)
	resp, err := c.getParsedResponse("POST", fmt.Sprintf("/repos/%s/%s/push_mirrors", user, repo), jsonHeader, bytes.NewReader(body), &pm)
	return pm, resp, err
}

// ListPushMirrorsOptions options for listing a repository's push mirrors
type ListPushMirrorsOptions struct {
	ListOptions
}

// ListPushMirrors gets all push mirrors of a repository
func (c *Client) ListPushMirrors(user, repo string, opt ListPushMirrorsOptions) ([]*PushMirrorResponse, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()
	link, _ := url.Parse(fmt.Sprintf("/repos/%s/%s/push_mirrors", user, repo))
	link.RawQuery = opt.getURLQuery().Encode()
	pms := make([]*PushMirrorResponse, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", link.String(), jsonHeader, nil, &pms)
	return pms, resp, err
}

// GetPushMirrorByRemoteName gets a push mirror of a repository by its remote name
func (c *Client) GetPushMirrorByRemoteName(user, repo, remoteName string) (*PushMirrorResponse, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo, &remoteName); err != nil {
		return nil, nil, err
	}
	pm := new(PushMirrorResponse)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/push_mirrors/%s", user, repo, remoteName), jsonHeader, nil, pm)
	return pm, resp, err
}

// DeletePushMirror removes a push mirror from a repository by its remote name
func (c *Client) DeletePushMirror(user, repo, remoteName string) (*Response, error) {
	if err := escapeValidatePathSegments(&user, &repo, &remoteName); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/repos/%s/%s/push_mirrors/%s", user, repo, remoteName), nil, nil)
	return resp, err
}

// PushMirrorSync adds all push mirrors of a repository to the sync queue
func (c *Client) PushMirrorSync(user, repo string) (*Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("POST", fmt.Sprintf("/repos/%s/%s/push_mirrors-sync", user, repo), nil, nil)
	return resp, err
}

// ConvertToNormalRepo converts a mirror repository to a normal (non-mirror) repository
func (c *Client) ConvertToNormalRepo(user, repo string) (*Repository, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	r := new(Repository)
	resp, err := c.getParsedResponse("POST", fmt.Sprintf("/repos/%s/%s/convert", user, repo), nil, nil, r)
	return r, resp, err
}
