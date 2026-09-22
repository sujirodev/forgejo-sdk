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

// WikiCommit page commit/revision
type WikiCommit struct {
	ID        string      `json:"sha"`
	Author    *CommitUser `json:"author"`
	Committer *CommitUser `json:"commiter"`
	Message   string      `json:"message"`
}

// WikiPageMetaData wiki page meta information
type WikiPageMetaData struct {
	Title      string      `json:"title"`
	HTMLURL    string      `json:"html_url"`
	SubURL     string      `json:"sub_url"`
	LastCommit *WikiCommit `json:"last_commit"`
}

// WikiPage a wiki page
type WikiPage struct {
	*WikiPageMetaData
	// Page content, base64 encoded
	ContentBase64 string `json:"content_base64"`
	CommitCount   int64  `json:"commit_count"`
	Sidebar       string `json:"sidebar"`
	Footer        string `json:"footer"`
}

// WikiCommitList commit/revision list
type WikiCommitList struct {
	WikiCommits []*WikiCommit `json:"commits"`
	Count       int64         `json:"count"`
}

// CreateWikiPageOptions form for creating a wiki page
type CreateWikiPageOptions struct {
	// page title. leave empty to keep unchanged
	Title string `json:"title"`
	// content must be base64 encoded
	ContentBase64 string `json:"content_base64"`
	// optional commit message summarizing the change
	Message string `json:"message"`
}

// CreateWikiPage creates a wiki page for a repository
func (c *Client) CreateWikiPage(owner, repo string, opt CreateWikiPageOptions) (*WikiPage, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}
	page := new(WikiPage)
	resp, err := c.getParsedResponse("POST", fmt.Sprintf("/repos/%s/%s/wiki/new", owner, repo), jsonHeader, bytes.NewReader(body), page)
	return page, resp, err
}

// GetWikiPage gets a wiki page. pageName may be empty to fetch the "Home" page
func (c *Client) GetWikiPage(owner, repo, pageName string) (*WikiPage, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	page := new(WikiPage)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/wiki/page/%s", owner, repo, url.PathEscape(pageName)), jsonHeader, nil, page)
	return page, resp, err
}

// EditWikiPage edits an existing wiki page. pageName may be empty to fetch the "Home" page
func (c *Client) EditWikiPage(owner, repo, pageName string, opt CreateWikiPageOptions) (*WikiPage, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}
	page := new(WikiPage)
	resp, err := c.getParsedResponse("PATCH", fmt.Sprintf("/repos/%s/%s/wiki/page/%s", owner, repo, url.PathEscape(pageName)), jsonHeader, bytes.NewReader(body), page)
	return page, resp, err
}

// DeleteWikiPage deletes a wiki page. pageName may be empty to delete the "Home" page
func (c *Client) DeleteWikiPage(owner, repo, pageName string) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/repos/%s/%s/wiki/page/%s", owner, repo, url.PathEscape(pageName)), nil, nil)
	return resp, err
}

// ListWikiPagesOptions options for listing a repository's wiki pages
type ListWikiPagesOptions struct {
	ListOptions
}

// ListWikiPages gets all wiki pages of a repository
func (c *Client) ListWikiPages(owner, repo string, opt ListWikiPagesOptions) ([]*WikiPageMetaData, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()
	link, _ := url.Parse(fmt.Sprintf("/repos/%s/%s/wiki/pages", owner, repo))
	link.RawQuery = opt.getURLQuery().Encode()
	pages := make([]*WikiPageMetaData, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", link.String(), jsonHeader, nil, &pages)
	return pages, resp, err
}

// ListWikiPageRevisionsOptions options for listing a wiki page's revisions
type ListWikiPageRevisionsOptions struct {
	ListOptions
}

// ListWikiPageRevisions gets the revisions of a wiki page. pageName may be empty for the "Home" page
func (c *Client) ListWikiPageRevisions(owner, repo, pageName string, opt ListWikiPageRevisionsOptions) (*WikiCommitList, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()
	link, _ := url.Parse(fmt.Sprintf("/repos/%s/%s/wiki/revisions/%s", owner, repo, url.PathEscape(pageName)))
	link.RawQuery = opt.getURLQuery().Encode()
	list := new(WikiCommitList)
	resp, err := c.getParsedResponse("GET", link.String(), jsonHeader, nil, list)
	return list, resp, err
}
