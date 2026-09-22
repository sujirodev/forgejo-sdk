// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListGitignoreTemplates(t *testing.T) {
	t.Parallel()
	log.Println("== TestListGitignoreTemplates ==")
	c := newTestClient()

	templates, _, err := c.ListGitignoreTemplates(t.Context())
	require.NoError(t, err)
	assert.NotEmpty(t, templates)
	assert.Contains(t, templates, "Go")
}

func TestGetGitignoreTemplate(t *testing.T) {
	t.Parallel()
	log.Println("== TestGetGitignoreTemplate ==")
	c := newTestClient()

	info, _, err := c.GetGitignoreTemplate(t.Context(), "Go")
	require.NoError(t, err)
	assert.Equal(t, "Go", info.Name)
	assert.NotEmpty(t, info.Source)
}

func TestGetGitignoreTemplate_NotFound(t *testing.T) {
	t.Parallel()
	log.Println("== TestGetGitignoreTemplate_NotFound ==")
	c := newTestClient()

	_, _, err := c.GetGitignoreTemplate(t.Context(), "DefinitelyNotARealTemplateName")
	require.Error(t, err)
}
