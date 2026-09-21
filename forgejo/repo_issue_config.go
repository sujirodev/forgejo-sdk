// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
)

// IssueConfigContactLink is a contact link declared in a repository's issue config
type IssueConfigContactLink struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	About string `json:"about"`
}

// IssueConfig is a repository's issue config, as read from .forgejo/ISSUE_TEMPLATE/config.yaml
// (or the .gitea/.github equivalents)
type IssueConfig struct {
	BlankIssuesEnabled bool                     `json:"blank_issues_enabled"`
	ContactLinks       []IssueConfigContactLink `json:"contact_links"`
}

// IssueConfigValidation is the result of validating a repository's issue config
type IssueConfigValidation struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

// GetIssueConfig returns the issue config for a repository
func (c *Client) GetIssueConfig(owner, repo string) (*IssueConfig, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	cfg := new(IssueConfig)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/issue_config", owner, repo), jsonHeader, nil, cfg)
	return cfg, resp, err
}

// ValidateIssueConfig returns the validation information for a repository's issue config
func (c *Client) ValidateIssueConfig(owner, repo string) (*IssueConfigValidation, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	v := new(IssueConfigValidation)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/issue_config/validate", owner, repo), jsonHeader, nil, v)
	return v, resp, err
}
