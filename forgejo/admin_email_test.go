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

func TestAdminListAllEmails(t *testing.T) {
	log.Println("== TestAdminListAllEmails ==")
	c := newTestClient()

	me, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	// A large PageSize, not the server's DEFAULT_PAGING_NUM: this suite
	// creates hundreds of users across its full run, and the admin's own
	// email can land past the first page of an unpaginated request.
	emails, resp, err := c.AdminListAllEmails(AdminListAllEmailsOptions{ListOptions: ListOptions{PageSize: 500}})
	require.NoError(t, err)
	require.NotNil(t, resp)

	found := false
	for _, e := range emails {
		if e.Email == me.Email {
			found = true
			assert.Equal(t, me.UserName, e.UserName)
			break
		}
	}
	assert.True(t, found, "AdminListAllEmails should include the admin's own email")
}

func TestAdminSearchEmails(t *testing.T) {
	log.Println("== TestAdminSearchEmails ==")
	c := newTestClient()

	me, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	emails, resp, err := c.AdminSearchEmails(AdminSearchEmailsOptions{Keyword: me.UserName})
	require.NoError(t, err)
	require.NotNil(t, resp)

	found := false
	for _, e := range emails {
		if e.Email == me.Email {
			found = true
			break
		}
	}
	assert.True(t, found, "AdminSearchEmails(%q) should find the admin's own email", me.UserName)
}
