// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// ReplaceFlagsOption options when replacing the flags of a repository
type ReplaceFlagsOption struct {
	Flags []string `json:"flags"`
}

// ListRepoFlags lists the flags of a repository
func (c *Client) ListRepoFlags(owner, repo string) ([]string, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	flags := make([]string, 0)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/flags", owner, repo), jsonHeader, nil, &flags)
	return flags, resp, err
}

// ReplaceAllRepoFlags replaces all the flags of a repository
func (c *Client) ReplaceAllRepoFlags(owner, repo string, opt ReplaceFlagsOption) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("PUT", fmt.Sprintf("/repos/%s/%s/flags", owner, repo), jsonHeader, bytes.NewReader(body))
	return resp, err
}

// DeleteAllRepoFlags removes all flags from a repository
func (c *Client) DeleteAllRepoFlags(owner, repo string) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/repos/%s/%s/flags", owner, repo), nil, nil)
	return resp, err
}

// CheckRepoFlag checks whether a repository has a given flag
func (c *Client) CheckRepoFlag(owner, repo, flag string) (bool, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo, &flag); err != nil {
		return false, nil, err
	}
	status, resp, err := c.getStatusCode("GET", fmt.Sprintf("/repos/%s/%s/flags/%s", owner, repo, flag), nil, nil)
	if err != nil {
		return false, resp, err
	}
	switch status {
	case http.StatusNoContent:
		return true, resp, nil
	case http.StatusNotFound:
		return false, resp, nil
	default:
		return false, resp, fmt.Errorf("unexpected Status: %d", status)
	}
}

// AddRepoFlag adds a flag to a repository
func (c *Client) AddRepoFlag(owner, repo, flag string) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo, &flag); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("PUT", fmt.Sprintf("/repos/%s/%s/flags/%s", owner, repo, flag), nil, nil)
	return resp, err
}

// DeleteRepoFlag removes a flag from a repository
func (c *Client) DeleteRepoFlag(owner, repo, flag string) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo, &flag); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/repos/%s/%s/flags/%s", owner, repo, flag), nil, nil)
	return resp, err
}
