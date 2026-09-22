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

// TestOrgLabels test organization-wide label related func
func TestOrgLabels(t *testing.T) {
	t.Parallel()
	log.Println("== TestOrgLabels ==")
	c := newTestClient()

	org := "OrgLabelTestOrg"
	if _, _, err := c.GetOrg(org); err == nil {
		_, _ = c.DeleteOrg(org)
	}
	_, _, err := c.CreateOrg(CreateOrgOption{Name: org, Visibility: VisibleTypePublic})
	require.NoError(t, err)
	defer func() { _, _ = c.DeleteOrg(org) }()

	plain, _, err := c.CreateOrgLabel(org, CreateLabelOption{
		Name:        "blue",
		Color:       "#0000FF",
		Description: "non-exclusive label",
	})
	require.NoError(t, err)
	assert.Equal(t, "blue", plain.Name)
	assert.Equal(t, "0000ff", plain.Color)
	assert.False(t, plain.Exclusive)

	scoped, _, err := c.CreateOrgLabel(org, CreateLabelOption{
		Name:      "kind/bug",
		Color:     "ee0701",
		Exclusive: true,
	})
	require.NoError(t, err)
	assert.True(t, scoped.Exclusive)

	labels, _, err := c.ListOrgLabels(org, ListLabelsOptions{})
	require.NoError(t, err)
	assert.Len(t, labels, 2)

	got, _, err := c.GetOrgLabel(org, plain.ID)
	require.NoError(t, err)
	assert.Equal(t, plain, got)

	edited, _, err := c.EditOrgLabel(org, plain.ID, EditLabelOption{
		Color:       OptionalString("#0e0175"),
		Description: OptionalString("edited"),
		IsArchived:  OptionalBool(true),
	})
	require.NoError(t, err)
	assert.Equal(t, "0e0175", edited.Color)
	assert.Equal(t, "edited", edited.Description)
	assert.True(t, edited.IsArchived)

	_, err = c.DeleteOrgLabel(org, plain.ID)
	require.NoError(t, err)
	labels, _, err = c.ListOrgLabels(org, ListLabelsOptions{})
	require.NoError(t, err)
	assert.Len(t, labels, 1)
}
