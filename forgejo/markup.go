// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// MarkdownOption holds the parameters for RenderMarkdown.
type MarkdownOption struct {
	// Context to render
	Context string `json:"Context,omitempty"`
	// Mode to render (comment, gfm, markdown)
	Mode string `json:"Mode,omitempty"`
	// Text markdown to render
	Text string `json:"Text,omitempty"`
	// Wiki tells whether this is a wiki page
	Wiki bool `json:"Wiki,omitempty"`
}

// MarkupOption holds the parameters for RenderMarkup.
type MarkupOption struct {
	// BranchPath is the current branch path where the form gets posted
	BranchPath string `json:"BranchPath,omitempty"`
	// Context to render
	Context string `json:"Context,omitempty"`
	// FilePath is the file path used for detecting the extension in file mode
	FilePath string `json:"FilePath,omitempty"`
	// Mode to render (comment, gfm, markdown, file)
	Mode string `json:"Mode,omitempty"`
	// Text markup to render
	Text string `json:"Text,omitempty"`
	// Wiki tells whether this is a wiki page
	Wiki bool `json:"Wiki,omitempty"`
}

// RenderMarkdown renders a markdown document as HTML.
func (c *Client) RenderMarkdown(ctx context.Context, opt MarkdownOption) (string, Response, error) {
	body, err := json.Marshal(&opt)
	if err != nil {
		return "", Response{}, err
	}

	header := http.Header{"content-type": []string{"application/json"}}
	data, resp, err := c.getResponseWithContext(ctx, "POST", "/markdown", header, bytes.NewReader(body))
	if err != nil {
		return "", resp, err
	}
	return string(data), resp, nil
}

// RenderMarkdownRaw renders a raw markdown document (plain text, not
// wrapped in a MarkdownOption) as HTML.
func (c *Client) RenderMarkdownRaw(ctx context.Context, text string) (string, Response, error) {
	header := http.Header{"content-type": []string{"text/plain"}}
	data, resp, err := c.getResponseWithContext(ctx, "POST", "/markdown/raw", header, strings.NewReader(text))
	if err != nil {
		return "", resp, err
	}
	return string(data), resp, nil
}

// RenderMarkup renders a markup document (markdown, or another markup
// language recognized from FilePath) as HTML.
func (c *Client) RenderMarkup(ctx context.Context, opt MarkupOption) (string, Response, error) {
	body, err := json.Marshal(&opt)
	if err != nil {
		return "", Response{}, err
	}

	header := http.Header{"content-type": []string{"application/json"}}
	data, resp, err := c.getResponseWithContext(ctx, "POST", "/markup", header, bytes.NewReader(body))
	if err != nil {
		return "", resp, err
	}
	return string(data), resp, nil
}
