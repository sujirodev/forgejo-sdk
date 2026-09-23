// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"io"
	"net/url"
)

// ListRepoActionArtifacts lists a repository's artifacts, across every
// workflow run. Set opt.Name to keep only the artifacts with that exact name.
func (c *Client) ListRepoActionArtifacts(owner, repo string, opt ListActionArtifactsOption) ([]*ActionArtifact, *Response, error) {
	// The artifact routes arrived in Forgejo 16.0.0: 15.0.9 has no such
	// route and answers a bare 404 page rather than an API error.
	if err := c.checkServerVersionGreaterThanOrEqual(version16_0_0); err != nil {
		return nil, nil, err
	}
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()

	link, _ := url.Parse(fmt.Sprintf("/repos/%s/%s/actions/artifacts", owner, repo))
	link.RawQuery = opt.QueryEncode()

	artifacts := make([]*ActionArtifact, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", link.String(), jsonHeader, nil, &artifacts)
	return artifacts, resp, err
}

// ListRepoActionRunArtifacts lists the artifacts a single workflow run
// produced.
func (c *Client) ListRepoActionRunArtifacts(owner, repo string, runID int64, opt ListActionArtifactsOption) ([]*ActionArtifact, *Response, error) {
	if err := c.checkServerVersionGreaterThanOrEqual(version16_0_0); err != nil {
		return nil, nil, err
	}
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()

	link, _ := url.Parse(fmt.Sprintf("/repos/%s/%s/actions/runs/%d/artifacts", owner, repo, runID))
	link.RawQuery = opt.QueryEncode()

	artifacts := make([]*ActionArtifact, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", link.String(), jsonHeader, nil, &artifacts)
	return artifacts, resp, err
}

// GetRepoActionArtifact gets one of a repository's artifacts by ID.
func (c *Client) GetRepoActionArtifact(owner, repo string, artifactID int64) (*ActionArtifact, *Response, error) {
	if err := c.checkServerVersionGreaterThanOrEqual(version16_0_0); err != nil {
		return nil, nil, err
	}
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}

	artifact := new(ActionArtifact)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/actions/artifacts/%d", owner, repo, artifactID), jsonHeader, nil, artifact)
	return artifact, resp, err
}

// DeleteRepoActionArtifact marks an artifact for deletion. The server removes
// the stored archive later, so the artifact may still be listed right after
// this call returns.
func (c *Client) DeleteRepoActionArtifact(owner, repo string, artifactID int64) (*Response, error) {
	if err := c.checkServerVersionGreaterThanOrEqual(version16_0_0); err != nil {
		return nil, err
	}
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}

	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/repos/%s/%s/actions/artifacts/%d", owner, repo, artifactID), jsonHeader, nil)
	return resp, err
}

// DownloadRepoActionArtifact downloads an artifact's zip archive. The archive
// is returned as a byte stream in a ReadCloser; closing it is the caller's
// responsibility.
func (c *Client) DownloadRepoActionArtifact(owner, repo string, artifactID int64) (io.ReadCloser, *Response, error) {
	if err := c.checkServerVersionGreaterThanOrEqual(version16_0_0); err != nil {
		return nil, nil, err
	}
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}

	return c.getResponseReader("GET", fmt.Sprintf("/repos/%s/%s/actions/artifacts/%d/zip", owner, repo, artifactID), nil, nil)
}
