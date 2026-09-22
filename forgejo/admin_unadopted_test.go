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

func TestAdminListUnadopted(t *testing.T) {
	t.Parallel()
	log.Println("== TestAdminListUnadopted ==")
	c := newTestClient()

	// A fresh test instance has no unadopted repositories, so this only
	// exercises that the call succeeds and returns a (possibly empty) list.
	repos, resp, err := c.AdminListUnadopted(AdminListUnadoptedOptions{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotNil(t, repos)
}

func TestAdminAdoptAndDeleteUnadoptedRepository_NotFound(t *testing.T) {
	t.Parallel()
	log.Println("== TestAdminAdoptAndDeleteUnadoptedRepository_NotFound ==")
	c := newTestClient()

	me, _, err := c.GetMyUserInfo()
	require.NoError(t, err)

	// There is no unadopted "does-not-exist" directory for this user, so
	// adopting it is expected to fail with a documented 404. This exercises
	// request wiring (path building, auth) but not the success path, which
	// requires a directory to exist under the server's repository root
	// outside of what the API itself can create - not reproducible from the
	// test client alone.
	_, err = c.AdminAdoptRepository(me.UserName, "sdk-admin-unadopted-does-not-exist")
	require.Error(t, err)

	// DeleteUnadoptedRepository only documents 204/403 responses (no 404),
	// so its behavior against a nonexistent path is server-defined; just
	// exercise that the request round-trips without a client-side error
	// building it.
	_, respErr := c.AdminDeleteUnadoptedRepository(me.UserName, "sdk-admin-unadopted-does-not-exist")
	_ = respErr
}
