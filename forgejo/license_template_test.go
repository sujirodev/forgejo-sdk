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

func TestListLicenseTemplates(t *testing.T) {
	t.Parallel()
	log.Println("== TestListLicenseTemplates ==")
	c := newTestClient()

	licenses, _, err := c.ListLicenseTemplates(t.Context())
	require.NoError(t, err)
	assert.NotEmpty(t, licenses)

	found := false
	for _, l := range licenses {
		if l.Key == "MIT" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected the MIT license to be in the list")
}

func TestGetLicenseTemplate(t *testing.T) {
	t.Parallel()
	log.Println("== TestGetLicenseTemplate ==")
	c := newTestClient()

	info, _, err := c.GetLicenseTemplate(t.Context(), "MIT")
	require.NoError(t, err)
	assert.Equal(t, "MIT", info.Key)
	assert.NotEmpty(t, info.Name)
	assert.NotEmpty(t, info.Body)
}

func TestGetLicenseTemplate_NotFound(t *testing.T) {
	t.Parallel()
	log.Println("== TestGetLicenseTemplate_NotFound ==")
	c := newTestClient()

	_, _, err := c.GetLicenseTemplate(t.Context(), "definitely-not-a-real-license")
	require.Error(t, err)
}
