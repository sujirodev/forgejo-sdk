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

func TestListLabelTemplates(t *testing.T) {
	log.Println("== TestListLabelTemplates ==")
	c := newTestClient()

	templates, _, err := c.ListLabelTemplates(t.Context())
	require.NoError(t, err)
	assert.NotEmpty(t, templates)
}

func TestGetLabelTemplate(t *testing.T) {
	log.Println("== TestGetLabelTemplate ==")
	c := newTestClient()

	templates, _, err := c.ListLabelTemplates(t.Context())
	require.NoError(t, err)
	require.NotEmpty(t, templates)

	labels, _, err := c.GetLabelTemplate(t.Context(), templates[0])
	require.NoError(t, err)
	require.NotEmpty(t, labels)
	assert.NotEmpty(t, labels[0].Name)
	assert.NotEmpty(t, labels[0].Color)
}

func TestGetLabelTemplate_NotFound(t *testing.T) {
	log.Println("== TestGetLabelTemplate_NotFound ==")
	c := newTestClient()

	_, _, err := c.GetLabelTemplate(t.Context(), "DefinitelyNotARealTemplateName")
	require.Error(t, err)
}
