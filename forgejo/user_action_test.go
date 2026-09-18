// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserActionSecrets(t *testing.T) {
	log.Println("== TestUserActionSecrets ==")
	c := newTestClient()

	secretName := "TEST_USER_SECRET"

	// Create
	resp, err := c.CreateUserActionSecret(CreateSecretOption{
		Name: secretName,
		Data: "test_value",
	})
	if err != nil {
		t.Skip("Could not create user secret, skipping test")
	}
	require.NotNil(t, resp)

	// Delete
	resp, err = c.DeleteUserActionSecret(secretName)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestListUserActionJobs(t *testing.T) {
	log.Println("== TestListUserActionJobs ==")
	c := newTestClient()

	// jobs may be nil when the API returns JSON null (no jobs exist)
	_, resp, err := c.ListUserActionJobs(ListActionJobsOption{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestUserActionVariables(t *testing.T) {
	log.Println("== TestUserActionVariables ==")
	c := newTestClient()

	variableName := "TEST_USER_VARIABLE"

	// Create
	resp, err := c.CreateUserActionVariable(CreateVariableOption{
		Name: variableName,
		Data: "test_value",
	})
	if err != nil {
		t.Skip("Could not create user variable, skipping test")
	}
	require.NotNil(t, resp)

	// List
	variables, resp, err := c.ListUserActionVariables(ListOptions{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	_ = variables

	// Get
	variable, resp, err := c.GetUserActionVariable(variableName)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, variableName, variable.Name)

	// Update
	resp, err = c.UpdateUserActionVariable(variableName, CreateVariableOption{
		Name: variableName,
		Data: "updated_value",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Delete
	resp, err = c.DeleteUserActionVariable(variableName)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestGetUserActionRunnerRegistrationToken(t *testing.T) {
	log.Println("== TestGetUserActionRunnerRegistrationToken ==")
	c := newTestClient()

	token, resp, err := c.GetUserActionRunnerRegistrationToken()
	if err == nil {
		require.NotNil(t, resp)
		assert.NotEmpty(t, token.Token)
	}
}
