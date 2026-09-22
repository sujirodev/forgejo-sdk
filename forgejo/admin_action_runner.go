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

// ActionRunner represents a Forgejo Actions runner, global or scoped to an
// organization, user, or repository.
type ActionRunner struct {
	ID          int64    `json:"id"`
	UUID        string   `json:"uuid"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	OwnerID     int64    `json:"owner_id"`
	RepoID      int64    `json:"repo_id"`
	Description string   `json:"description"`
	Status      string   `json:"status"`
	Labels      []string `json:"labels"`
	Ephemeral   bool     `json:"ephemeral"`
}

// AdminListRunnersOptions options for listing admin runners
type AdminListRunnersOptions struct {
	ListOptions
	// Visible, when true, includes all runners visible to the instance
	// (global, and org/user/repo scoped). When false, only runners directly
	// owned by the instance are returned.
	Visible bool
}

// AdminListRunners lists all runners, no matter whether they are global
// runners or scoped to an organization, user, or repository.
func (c *Client) AdminListRunners(opt AdminListRunnersOptions) ([]*ActionRunner, *Response, error) {
	// Confirmed absent on a live 13.0.0 instance (404) and present by
	// 15.0.9; 14.x was not verified (no such image tag exists to test
	// against), so this guard may be more conservative than the true floor.
	if err := c.checkServerVersionGreaterThanOrEqual(version15_0_0); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()
	query := opt.getURLQuery()
	if opt.Visible {
		query.Add("visible", "true")
	}

	link, _ := url.Parse("/admin/actions/runners")
	link.RawQuery = query.Encode()

	runners := make([]*ActionRunner, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", link.String(), jsonHeader, nil, &runners)
	return runners, resp, err
}

// RegisterRunnerOption options for registering a new global runner
type RegisterRunnerOption struct {
	// Name of the runner to register. Does not have to be unique.
	Name string `json:"name"`
	// Description provides optional details about the runner.
	Description string `json:"description,omitempty"`
	// Ephemeral registers the runner as an ephemeral runner.
	// See https://forgejo.org/docs/latest/admin/actions/security/#ephemeral-runner
	Ephemeral bool `json:"ephemeral,omitempty"`
}

// RegisterRunnerResponse contains the details of the just-registered runner.
// Unlike a registration token, Token here is an actual authentication token
// the runner daemon can use directly.
type RegisterRunnerResponse struct {
	ID    int64  `json:"id"`
	UUID  string `json:"uuid"`
	Token string `json:"token"`
}

// AdminRegisterRunner registers a new global runner.
func (c *Client) AdminRegisterRunner(opt RegisterRunnerOption) (*RegisterRunnerResponse, *Response, error) {
	if err := c.checkServerVersionGreaterThanOrEqual(version15_0_0); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}
	reg := new(RegisterRunnerResponse)
	resp, err := c.getParsedResponse("POST", "/admin/actions/runners", jsonHeader, bytes.NewReader(body), reg)
	return reg, resp, err
}

// AdminGetRunner gets a particular runner, no matter whether it is a global
// runner or scoped to an organization, user, or repository.
func (c *Client) AdminGetRunner(runnerID int64) (*ActionRunner, *Response, error) {
	if err := c.checkServerVersionGreaterThanOrEqual(version15_0_0); err != nil {
		return nil, nil, err
	}
	runner := new(ActionRunner)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/admin/actions/runners/%d", runnerID), jsonHeader, nil, runner)
	return runner, resp, err
}

// AdminDeleteRunner deletes a particular runner, no matter whether it is a
// global runner or scoped to an organization, user, or repository.
func (c *Client) AdminDeleteRunner(runnerID int64) (*Response, error) {
	if err := c.checkServerVersionGreaterThanOrEqual(version15_0_0); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/admin/actions/runners/%d", runnerID), jsonHeader, nil)
	return resp, err
}

// AdminListActionRunJobs gets action run jobs across the whole instance,
// optionally filtered by a comma separated list of labels.
func (c *Client) AdminListActionRunJobs(opt ListActionJobsOption) ([]*ActionRunJob, *Response, error) {
	if err := c.checkServerVersionGreaterThanOrEqual(version15_0_0); err != nil {
		return nil, nil, err
	}
	link, _ := url.Parse("/admin/actions/runners/jobs")
	if opt.Labels != "" {
		query := link.Query()
		query.Add("labels", opt.Labels)
		link.RawQuery = query.Encode()
	}

	jobs := make([]*ActionRunJob, 0)
	resp, err := c.getParsedResponse("GET", link.String(), jsonHeader, nil, &jobs)
	return jobs, resp, err
}
