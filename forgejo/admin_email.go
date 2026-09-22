// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"net/url"
)

// AdminListAllEmailsOptions options for listing all users' email addresses
type AdminListAllEmailsOptions struct {
	ListOptions
}

// AdminListAllEmails lists all users' email addresses.
func (c *Client) AdminListAllEmails(opt AdminListAllEmailsOptions) ([]*Email, *Response, error) {
	opt.setDefaults()
	emails := make([]*Email, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/admin/emails?%s", opt.getURLQuery().Encode()), jsonHeader, nil, &emails)
	return emails, resp, err
}

// AdminSearchEmailsOptions options for searching users' email addresses
type AdminSearchEmailsOptions struct {
	ListOptions
	// Keyword to search for
	Keyword string
}

// AdminSearchEmails searches users' email addresses.
func (c *Client) AdminSearchEmails(opt AdminSearchEmailsOptions) ([]*Email, *Response, error) {
	opt.setDefaults()
	query := opt.getURLQuery()
	if opt.Keyword != "" {
		query.Add("q", opt.Keyword)
	}

	link, _ := url.Parse("/admin/emails/search")
	link.RawQuery = query.Encode()

	emails := make([]*Email, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", link.String(), jsonHeader, nil, &emails)
	return emails, resp, err
}
