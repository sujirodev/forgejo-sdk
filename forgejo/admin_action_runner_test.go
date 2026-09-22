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

func TestAdminActionRunners(t *testing.T) {
	log.Println("== TestAdminActionRunners ==")
	c := newTestClient()

	// Confirmed absent on a live 13.0.0 instance and present by 15.0.9;
	// below that the SDK's guard refuses the call, which is the documented
	// behavior and worth asserting.
	if !serverAtLeast(t, c, "15.0.0") {
		_, _, err := c.AdminRegisterRunner(RegisterRunnerOption{Name: "sdk-admin-runner-test"})
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
		return
	}

	reg, resp, err := c.AdminRegisterRunner(RegisterRunnerOption{
		Name:        "sdk-admin-runner-test",
		Description: "created by forgejo-sdk admin runner test",
		Ephemeral:   false,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, reg.UUID)
	require.NotEmpty(t, reg.Token)
	require.NotZero(t, reg.ID)
	defer func() { _, _ = c.AdminDeleteRunner(reg.ID) }()

	runner, _, err := c.AdminGetRunner(reg.ID)
	require.NoError(t, err)
	assert.Equal(t, reg.ID, runner.ID)
	assert.Equal(t, reg.UUID, runner.UUID)
	assert.Equal(t, "sdk-admin-runner-test", runner.Name)

	runners, _, err := c.AdminListRunners(AdminListRunnersOptions{})
	require.NoError(t, err)
	found := false
	for _, r := range runners {
		if r.ID == reg.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "registered runner should be present in AdminListRunners")

	resp, err = c.AdminDeleteRunner(reg.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)

	_, _, err = c.AdminGetRunner(reg.ID)
	require.Error(t, err)
}

func TestAdminListActionRunJobs(t *testing.T) {
	log.Println("== TestAdminListActionRunJobs ==")
	c := newTestClient()

	if !serverAtLeast(t, c, "15.0.0") {
		_, _, err := c.AdminListActionRunJobs(ListActionJobsOption{})
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
		return
	}

	// jobs may be empty on a fresh instance, but the call itself must succeed.
	jobs, resp, err := c.AdminListActionRunJobs(ListActionJobsOption{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotNil(t, jobs)
}
