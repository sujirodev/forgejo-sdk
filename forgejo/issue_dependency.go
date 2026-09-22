// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
)

// IssueMeta identifies an issue by owner, repo and index, for use as the
// request body of endpoints that reference another issue (e.g. blocking and
// dependency relations)
type IssueMeta struct {
	Owner string `json:"owner"`
	Name  string `json:"repo"`
	Index int64  `json:"index"`
}

// ListIssueBlocks lists the issues that are blocked by the given issue
func (c *Client) ListIssueBlocks(owner, repo string, index int64, opt ListOptions) ([]*Issue, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()
	link, _ := url.Parse(fmt.Sprintf("/repos/%s/%s/issues/%d/blocks", owner, repo, index))
	link.RawQuery = opt.getURLQuery().Encode()
	issues := make([]*Issue, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", link.String(), nil, nil, &issues)
	return issues, resp, err
}

// CreateIssueBlocking blocks the issue given in meta by the issue given in index
func (c *Client) CreateIssueBlocking(owner, repo string, index int64, meta IssueMeta) (*Issue, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&meta)
	if err != nil {
		return nil, nil, err
	}
	issue := new(Issue)
	resp, err := c.getParsedResponse("POST",
		fmt.Sprintf("/repos/%s/%s/issues/%d/blocks", owner, repo, index),
		jsonHeader, bytes.NewReader(body), issue)
	return issue, resp, err
}

// RemoveIssueBlocking unblocks the issue given in meta from the issue given in index
func (c *Client) RemoveIssueBlocking(owner, repo string, index int64, meta IssueMeta) (*Issue, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&meta)
	if err != nil {
		return nil, nil, err
	}
	issue := new(Issue)
	resp, err := c.getParsedResponse("DELETE",
		fmt.Sprintf("/repos/%s/%s/issues/%d/blocks", owner, repo, index),
		jsonHeader, bytes.NewReader(body), issue)
	return issue, resp, err
}

// ListIssueDependencies lists all issues that block the given issue
func (c *Client) ListIssueDependencies(owner, repo string, index int64, opt ListOptions) ([]*Issue, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()
	link, _ := url.Parse(fmt.Sprintf("/repos/%s/%s/issues/%d/dependencies", owner, repo, index))
	link.RawQuery = opt.getURLQuery().Encode()
	issues := make([]*Issue, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", link.String(), nil, nil, &issues)
	return issues, resp, err
}

// CreateIssueDependency makes the issue given in index depend on the issue given in meta
func (c *Client) CreateIssueDependency(owner, repo string, index int64, meta IssueMeta) (*Issue, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&meta)
	if err != nil {
		return nil, nil, err
	}
	issue := new(Issue)
	resp, err := c.getParsedResponse("POST",
		fmt.Sprintf("/repos/%s/%s/issues/%d/dependencies", owner, repo, index),
		jsonHeader, bytes.NewReader(body), issue)
	return issue, resp, err
}

// RemoveIssueDependency removes the issue given in meta as a dependency of the issue given in index
func (c *Client) RemoveIssueDependency(owner, repo string, index int64, meta IssueMeta) (*Issue, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&meta)
	if err != nil {
		return nil, nil, err
	}
	issue := new(Issue)
	resp, err := c.getParsedResponse("DELETE",
		fmt.Sprintf("/repos/%s/%s/issues/%d/dependencies", owner, repo, index),
		jsonHeader, bytes.NewReader(body), issue)
	return issue, resp, err
}
