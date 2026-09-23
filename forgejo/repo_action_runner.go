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

// ListRepoRunners gets the runners that belong to the repository.
// Set opt.Visible to also include the runners the repository can use but does
// not own (the organization's and the instance's).
func (c *Client) ListRepoRunners(owner, repo string, opt ListActionRunnersOptions) ([]*ActionRunner, *Response, error) {
	// The repository-scoped runner routes arrived with the simplified
	// registration API of Forgejo 15.0.0; 11.0.16 has no such route. Same
	// floor as the organization-scoped ones in org_action.go.
	if err := c.checkServerVersionGreaterThanOrEqual(version15_0_0); err != nil {
		return nil, nil, err
	}
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()

	link, _ := url.Parse(fmt.Sprintf("/repos/%s/%s/actions/runners", owner, repo))
	link.RawQuery = opt.QueryEncode()

	runners := make([]*ActionRunner, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", link.String(), jsonHeader, nil, &runners)
	return runners, resp, err
}

// RegisterRepoRunner registers a new repository-level runner and returns the
// token the runner daemon needs to authenticate. The token is not a
// registration token: it is the runner's own credential, usable directly.
func (c *Client) RegisterRepoRunner(owner, repo string, opt RegisterRunnerOption) (*RegisterRunnerResponse, *Response, error) {
	if err := c.checkServerVersionGreaterThanOrEqual(version15_0_0); err != nil {
		return nil, nil, err
	}
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}

	runner := new(RegisterRunnerResponse)
	resp, err := c.getParsedResponse("POST", fmt.Sprintf("/repos/%s/%s/actions/runners", owner, repo), jsonHeader, bytes.NewReader(body), runner)
	return runner, resp, err
}

// GetRepoRunner gets a particular runner that belongs to the repository.
func (c *Client) GetRepoRunner(owner, repo string, runnerID int64) (*ActionRunner, *Response, error) {
	if err := c.checkServerVersionGreaterThanOrEqual(version15_0_0); err != nil {
		return nil, nil, err
	}
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}

	runner := new(ActionRunner)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/actions/runners/%d", owner, repo, runnerID), jsonHeader, nil, runner)
	return runner, resp, err
}

// DeleteRepoRunner deletes a particular runner that belongs to the repository.
func (c *Client) DeleteRepoRunner(owner, repo string, runnerID int64) (*Response, error) {
	if err := c.checkServerVersionGreaterThanOrEqual(version15_0_0); err != nil {
		return nil, err
	}
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}

	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/repos/%s/%s/actions/runners/%d", owner, repo, runnerID), jsonHeader, nil)
	return resp, err
}
