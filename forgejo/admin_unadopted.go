// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"net/url"
)

// AdminListUnadoptedOptions options for listing unadopted repositories
type AdminListUnadoptedOptions struct {
	ListOptions
	// Pattern of repositories to search for
	Pattern string
}

// AdminListUnadopted lists unadopted repositories, i.e. directories in the
// repository root that are not tracked as repositories in the database.
func (c *Client) AdminListUnadopted(opt AdminListUnadoptedOptions) ([]string, *Response, error) {
	opt.setDefaults()
	query := opt.getURLQuery()
	if opt.Pattern != "" {
		query.Add("pattern", opt.Pattern)
	}

	link, _ := url.Parse("/admin/unadopted")
	link.RawQuery = query.Encode()

	repos := make([]string, 0)
	resp, err := c.getParsedResponse("GET", link.String(), jsonHeader, nil, &repos)
	return repos, resp, err
}

// AdminAdoptRepository adopts unadopted files as a repository owned by the
// given owner.
func (c *Client) AdminAdoptRepository(owner, repo string) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("POST", fmt.Sprintf("/admin/unadopted/%s/%s", owner, repo), nil, nil)
	return resp, err
}

// AdminDeleteUnadoptedRepository deletes unadopted files that would otherwise
// have been adopted as owner/repo.
func (c *Client) AdminDeleteUnadoptedRepository(owner, repo string) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/admin/unadopted/%s/%s", owner, repo), nil, nil)
	return resp, err
}
