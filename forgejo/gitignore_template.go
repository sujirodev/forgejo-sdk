// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import "context"

// GitignoreTemplateInfo holds the name and content of a gitignore template.
type GitignoreTemplateInfo struct {
	Name   string `json:"name"`
	Source string `json:"source"`
}

// ListGitignoreTemplates returns the names of all gitignore templates known
// to the server.
func (c *Client) ListGitignoreTemplates(ctx context.Context) ([]string, Response, error) {
	var templates []string
	resp, err := c.getParsedResponseWithContext(ctx, "/gitignore/templates", &templates)
	return templates, resp, err
}

// GetGitignoreTemplate returns the name and content of a single gitignore
// template.
func (c *Client) GetGitignoreTemplate(ctx context.Context, name string) (GitignoreTemplateInfo, Response, error) {
	if err := escapeValidatePathSegments(&name); err != nil {
		return GitignoreTemplateInfo{}, Response{}, err
	}

	var info GitignoreTemplateInfo
	resp, err := c.getParsedResponseWithContext(ctx, "/gitignore/templates/"+name, &info)
	return info, resp, err
}
