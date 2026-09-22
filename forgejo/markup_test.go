// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderMarkdown(t *testing.T) {
	t.Parallel()
	log.Println("== TestRenderMarkdown ==")
	c := newTestClient()

	html, _, err := c.RenderMarkdown(t.Context(), MarkdownOption{
		Mode: "markdown",
		Text: "# Hello\n\nThis is **bold**.",
	})
	require.NoError(t, err)
	assert.Contains(t, html, "Hello")
	assert.Contains(t, strings.ToLower(html), "<strong>bold</strong>")
}

func TestRenderMarkdownRaw(t *testing.T) {
	t.Parallel()
	log.Println("== TestRenderMarkdownRaw ==")
	c := newTestClient()

	html, _, err := c.RenderMarkdownRaw(t.Context(), "# Hello Raw")
	require.NoError(t, err)
	assert.Contains(t, html, "Hello Raw")
}

func TestRenderMarkup(t *testing.T) {
	t.Parallel()
	log.Println("== TestRenderMarkup ==")
	c := newTestClient()

	html, _, err := c.RenderMarkup(t.Context(), MarkupOption{
		Mode: "markdown",
		Text: "# Hello Markup",
	})
	require.NoError(t, err)
	assert.Contains(t, html, "Hello Markup")
}
