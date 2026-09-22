// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import "context"

// LicenseTemplateListEntry is one entry of the list returned by
// ListLicenseTemplates.
type LicenseTemplateListEntry struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// LicenseTemplateInfo holds the full text and metadata of a single license
// template.
type LicenseTemplateInfo struct {
	Key            string `json:"key"`
	Name           string `json:"name"`
	URL            string `json:"url"`
	Implementation string `json:"implementation"`
	Body           string `json:"body"`
}

// ListLicenseTemplates returns every license template known to the server.
func (c *Client) ListLicenseTemplates(ctx context.Context) ([]LicenseTemplateListEntry, Response, error) {
	var licenses []LicenseTemplateListEntry
	resp, err := c.getParsedResponseWithContext(ctx, "/licenses", &licenses)
	return licenses, resp, err
}

// GetLicenseTemplate returns the full text and metadata of a single license
// template.
func (c *Client) GetLicenseTemplate(ctx context.Context, name string) (LicenseTemplateInfo, Response, error) {
	if err := escapeValidatePathSegments(&name); err != nil {
		return LicenseTemplateInfo{}, Response{}, err
	}

	var info LicenseTemplateInfo
	resp, err := c.getParsedResponseWithContext(ctx, "/licenses/"+name, &info)
	return info, resp, err
}
