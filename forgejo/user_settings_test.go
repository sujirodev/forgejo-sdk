// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserSettings(t *testing.T) {
	log.Println("== TestUserSettings ==")
	c := newTestClient()
	owner := createTestUser(t, uniqueName(t, "setowner"), c)
	c.SetSudo(owner.UserName)
	t.Cleanup(func() { c.SetSudo("") })

	userConf, _, err := c.GetUserSettings()
	require.NoError(t, err)
	assert.NotNil(t, userConf)
	assert.Equal(t, UserSettings{
		Theme:        "forgejo-auto",
		HideEmail:    false,
		HideActivity: false,
	}, *userConf)

	userConf, _, err = c.UpdateUserSettings(UserSettingsOptions{
		FullName:  OptionalString("Admin User on Test"),
		Language:  OptionalString("de_de"),
		HideEmail: OptionalBool(true),
	})
	require.NoError(t, err)
	assert.NotNil(t, userConf)
	assert.Equal(t, UserSettings{
		FullName:     "Admin User on Test",
		Theme:        "forgejo-auto",
		Language:     "de_de",
		HideEmail:    true,
		HideActivity: false,
	}, *userConf)

	_, _, err = c.UpdateUserSettings(UserSettingsOptions{
		FullName:  OptionalString(""),
		Language:  OptionalString(""),
		HideEmail: OptionalBool(false),
	})
	require.NoError(t, err)
}
