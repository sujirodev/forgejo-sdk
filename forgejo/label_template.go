// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import "context"

// LabelTemplate is a single label defined by a label template.
type LabelTemplate struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
	Exclusive   bool   `json:"exclusive"`
}

// ListLabelTemplates returns the names of all label templates known to the
// server.
func (c *Client) ListLabelTemplates(ctx context.Context) ([]string, Response, error) {
	var templates []string
	resp, err := c.getParsedResponseWithContext(ctx, "/label/templates", &templates)
	return templates, resp, err
}

// GetLabelTemplate returns all the labels defined by a single label
// template.
func (c *Client) GetLabelTemplate(ctx context.Context, name string) ([]LabelTemplate, Response, error) {
	if err := escapeValidatePathSegments(&name); err != nil {
		return nil, Response{}, err
	}

	var labels []LabelTemplate
	resp, err := c.getParsedResponseWithContext(ctx, "/label/templates/"+name, &labels)
	return labels, resp, err
}
