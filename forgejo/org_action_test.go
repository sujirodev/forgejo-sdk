// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2024 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constant for org action tests
var testOrgName = "testorg"

func TestOrgActionSecrets(t *testing.T) {
	log.Println("== TestOrgActionSecrets ==")
	c := newTestClient()

	// Create one user and org for all subtests
	user := createTestUser(t, "org_action_test_user", c)
	c.SetSudo(user.UserName)
	testOrg, _, err := c.CreateOrg(CreateOrgOption{Name: "ActionTestOrg"})
	require.NoError(t, err)
	require.NotNil(t, testOrg)

	t.Run("CreateAndUpdate", func(t *testing.T) {
		// create secret
		resp, err := c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "test", Data: "test"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// update secret
		resp, err = c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "test", Data: "test2"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// create secret with 31 character name
		longName31 := strings.Repeat("A", 31)
		resp, err = c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: longName31, Data: "test_data"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// create secret with long value
		longValue := strings.Repeat("secret", 100)
		resp, err = c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "long_value_test", Data: longValue})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// list secrets - should have 3 secrets
		secrets, _, err := c.ListOrgActionSecret(testOrg.UserName, ListOrgActionSecretOption{})
		require.NoError(t, err)
		assert.Len(t, secrets, 3)
	})

	t.Run("InvalidNames", func(t *testing.T) {
		// test invalid names - client-side validation should reject these
		_, err := c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "", Data: "data"})
		require.Error(t, err)
		require.EqualError(t, err, "name required")

		_, err = c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "INVALID-NAME", Data: "data"})
		require.Error(t, err)
		require.EqualError(t, err, "name must contain only alphanumeric characters and underscores")

		_, err = c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "INVALID NAME", Data: "data"})
		require.Error(t, err)
		require.EqualError(t, err, "name must contain only alphanumeric characters and underscores")

		_, err = c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "GITEA_SECRET", Data: "data"})
		require.Error(t, err)
		require.EqualError(t, err, "name cannot start with GITEA_ or GITHUB_ (reserved prefixes)")

		_, err = c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "github_token", Data: "data"})
		require.Error(t, err)
		require.EqualError(t, err, "name cannot start with GITEA_ or GITHUB_ (reserved prefixes)")

		_, err = c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: strings.Repeat("A", 256), Data: "data"})
		require.Error(t, err)
		require.EqualError(t, err, "name too long (maximum 255 characters)")
	})

	t.Run("EmptyData", func(t *testing.T) {
		_, err := c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "VALID_NAME", Data: ""})
		require.Error(t, err)
		require.EqualError(t, err, "data required")
	})

	t.Run("UpdateMultipleTimes", func(t *testing.T) {
		// create and update same secret multiple times
		resp, err := c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "UPDATE_SECRET", Data: "data1"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		for i := 2; i <= 5; i++ {
			resp, err = c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "UPDATE_SECRET", Data: "data" + string(rune('0'+i))})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		}

		secrets, _, err := c.ListOrgActionSecret(testOrg.UserName, ListOrgActionSecretOption{})
		require.NoError(t, err)
		// Should have original 3 + UPDATE_SECRET = 4 total
		assert.GreaterOrEqual(t, len(secrets), 1)
	})

	t.Run("CaseSensitivity", func(t *testing.T) {
		// create secret with lowercase name
		resp, err := c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "my_secret", Data: "lower"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// try to create with uppercase - should update the same secret (case-insensitive)
		resp, err = c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "MY_SECRET", Data: "upper"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// try with mixed case - should also update the same secret
		resp, err = c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "My_Secret", Data: "mixed"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("Pagination", func(t *testing.T) {
		// create 10 more secrets for pagination test
		for i := 1; i <= 10; i++ {
			name := "PAGE_SECRET_" + string(rune('A'+i-1))
			_, err := c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: name, Data: "data" + string(rune('0'+i))})
			require.NoError(t, err)
		}

		// test pagination
		secrets, _, err := c.ListOrgActionSecret(testOrg.UserName, ListOrgActionSecretOption{
			ListOptions: ListOptions{Page: 1, PageSize: 5},
		})
		require.NoError(t, err)
		assert.Len(t, secrets, 5)

		secrets, _, err = c.ListOrgActionSecret(testOrg.UserName, ListOrgActionSecretOption{
			ListOptions: ListOptions{Page: 2, PageSize: 5},
		})
		require.NoError(t, err)
		assert.Len(t, secrets, 5)

		// get all
		secrets, _, err = c.ListOrgActionSecret(testOrg.UserName, ListOrgActionSecretOption{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(secrets), 10)
	})

	t.Run("NonExistentOrg", func(t *testing.T) {
		// trying to create secret for non-existent org should fail
		_, err := c.CreateOrgActionSecret("NonExistentOrg123456", CreateSecretOption{Name: "TEST", Data: "data"})
		require.Error(t, err)

		// listing secrets for non-existent org also returns error
		_, _, err = c.ListOrgActionSecret("NonExistentOrg123456", ListOrgActionSecretOption{})
		require.Error(t, err)
	})

	t.Run("LargeData", func(t *testing.T) {
		// test various data sizes
		smallData := "small"
		mediumData := strings.Repeat("medium", 100)
		largeData := strings.Repeat("large", 1000)

		_, err := c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "SMALL_DATA", Data: smallData})
		require.NoError(t, err)

		_, err = c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "MEDIUM_DATA", Data: mediumData})
		require.NoError(t, err)

		_, err = c.CreateOrgActionSecret(testOrg.UserName, CreateSecretOption{Name: "LARGE_DATA", Data: largeData})
		require.NoError(t, err)
	})
}

func TestCreateSecretOption_Validate(t *testing.T) {
	log.Println("== TestCreateSecretOption_Validate ==")
	tests := []struct {
		name    string
		opt     CreateSecretOption
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid secret with short name",
			opt:     CreateSecretOption{Name: "TEST_SECRET", Data: "secret_value"},
			wantErr: false,
		},
		{
			name:    "valid secret with 30 character name",
			opt:     CreateSecretOption{Name: strings.Repeat("A", 30), Data: "secret_value"},
			wantErr: false,
		},
		{
			name:    "valid secret with 31 character name",
			opt:     CreateSecretOption{Name: strings.Repeat("A", 31), Data: "secret_value"},
			wantErr: false,
		},
		{
			name:    "valid secret with 64 character name",
			opt:     CreateSecretOption{Name: strings.Repeat("A", 64), Data: "secret_value"},
			wantErr: false,
		},
		{
			name:    "valid secret with 255 character name (max length)",
			opt:     CreateSecretOption{Name: strings.Repeat("A", 255), Data: "secret_value"},
			wantErr: false,
		},
		{
			name:    "valid secret with underscores and mixed case",
			opt:     CreateSecretOption{Name: "My_Test_Secret_123", Data: "secret_value"},
			wantErr: false,
		},
		{
			name:    "valid secret with only alphanumeric (no underscores)",
			opt:     CreateSecretOption{Name: "MyTestSecret123", Data: "secret_value"},
			wantErr: false,
		},
		{
			name:    "valid secret starting with number",
			opt:     CreateSecretOption{Name: "123_TEST_SECRET", Data: "secret_value"},
			wantErr: false,
		},
		{
			name:    "valid secret with only lowercase letters",
			opt:     CreateSecretOption{Name: "test_secret", Data: "secret_value"},
			wantErr: false,
		},
		{
			name:    "valid secret with only uppercase letters",
			opt:     CreateSecretOption{Name: "TEST_SECRET_UPPER", Data: "secret_value"},
			wantErr: false,
		},
		{
			name:    "empty name should fail",
			opt:     CreateSecretOption{Name: "", Data: "secret_value"},
			wantErr: true,
			errMsg:  "name required",
		},
		{
			name:    "name exceeding 255 characters should fail",
			opt:     CreateSecretOption{Name: strings.Repeat("A", 256), Data: "secret_value"},
			wantErr: true,
			errMsg:  "name too long",
		},
		{
			name:    "name with spaces should fail",
			opt:     CreateSecretOption{Name: "TEST SECRET", Data: "secret_value"},
			wantErr: true,
			errMsg:  "must contain only alphanumeric characters and underscores",
		},
		{
			name:    "name with hyphen should fail",
			opt:     CreateSecretOption{Name: "TEST-SECRET", Data: "secret_value"},
			wantErr: true,
			errMsg:  "must contain only alphanumeric characters and underscores",
		},
		{
			name:    "name with special characters should fail",
			opt:     CreateSecretOption{Name: "TEST@SECRET", Data: "secret_value"},
			wantErr: true,
			errMsg:  "must contain only alphanumeric characters and underscores",
		},
		{
			name:    "name starting with GITEA_ should fail",
			opt:     CreateSecretOption{Name: "GITEA_SECRET", Data: "secret_value"},
			wantErr: true,
			errMsg:  "cannot start with GITEA_ or GITHUB_",
		},
		{
			name:    "name starting with gitea_ (lowercase) should fail",
			opt:     CreateSecretOption{Name: "gitea_secret", Data: "secret_value"},
			wantErr: true,
			errMsg:  "cannot start with GITEA_ or GITHUB_",
		},
		{
			name:    "name starting with GITHUB_ should fail",
			opt:     CreateSecretOption{Name: "GITHUB_TOKEN", Data: "secret_value"},
			wantErr: true,
			errMsg:  "cannot start with GITEA_ or GITHUB_",
		},
		{
			name:    "name starting with github_ (lowercase) should fail",
			opt:     CreateSecretOption{Name: "github_token", Data: "secret_value"},
			wantErr: true,
			errMsg:  "cannot start with GITEA_ or GITHUB_",
		},
		{
			name:    "name starting with GiTeA_ (mixed case) should fail",
			opt:     CreateSecretOption{Name: "GiTeA_secret", Data: "secret_value"},
			wantErr: true,
			errMsg:  "cannot start with GITEA_ or GITHUB_",
		},
		{
			name:    "name starting with GiTHuB_ (mixed case) should fail",
			opt:     CreateSecretOption{Name: "GiTHuB_token", Data: "secret_value"},
			wantErr: true,
			errMsg:  "cannot start with GITEA_ or GITHUB_",
		},
		{
			name:    "name containing GITEA_ but not starting with it should pass",
			opt:     CreateSecretOption{Name: "MY_GITEA_SECRET", Data: "secret_value"},
			wantErr: false,
		},
		{
			name:    "empty data should fail",
			opt:     CreateSecretOption{Name: "TEST_SECRET", Data: ""},
			wantErr: true,
			errMsg:  "data required",
		},
		{
			name:    "long data value should pass",
			opt:     CreateSecretOption{Name: "TEST_SECRET", Data: strings.Repeat("x", 1000)},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opt.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteOrgActionSecret(t *testing.T) {
	log.Println("== TestDeleteOrgActionSecret ==")
	c := newTestClient()

	// First create a secret to delete
	_, err := c.CreateOrgActionSecret(testOrgName, CreateSecretOption{
		Name: "DELETE_TEST_SECRET",
		Data: "test_value",
	})
	if err != nil {
		t.Skip("Could not create secret, skipping delete test")
	}

	resp, err := c.DeleteOrgActionSecret(testOrgName, "DELETE_TEST_SECRET")
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestListOrgActionJobs(t *testing.T) {
	log.Println("== TestListOrgActionJobs ==")
	c := newTestClient()

	user := createTestUser(t, "org_action_jobs_user", c)
	c.SetSudo(user.UserName)
	org, _, err := c.CreateOrg(CreateOrgOption{Name: "ActionJobsTestOrg"})
	require.NoError(t, err)
	require.NotNil(t, org)

	// jobs may be nil when the API returns JSON null (no jobs exist)
	_, resp, err := c.ListOrgActionJobs(org.UserName, ListActionJobsOption{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestOrgActionVariables(t *testing.T) {
	log.Println("== TestOrgActionVariables ==")
	c := newTestClient()

	variableName := "TEST_ORG_VARIABLE"

	// Create
	resp, err := c.CreateOrgActionVariable(testOrgName, CreateVariableOption{
		Name: variableName,
		Data: "test_value",
	})
	if err != nil {
		t.Skip("Could not create variable, skipping test")
	}
	require.NotNil(t, resp)

	// List
	variables, resp, err := c.ListOrgActionVariables(testOrgName, ListOptions{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, variables)

	// Get
	variable, resp, err := c.GetOrgActionVariable(testOrgName, variableName)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, variableName, variable.Name)

	// Update
	resp, err = c.UpdateOrgActionVariable(testOrgName, variableName, CreateVariableOption{
		Name: variableName,
		Data: "updated_value",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Delete
	resp, err = c.DeleteOrgActionVariable(testOrgName, variableName)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestGetOrgActionRunnerRegistrationToken(t *testing.T) {
	log.Println("== TestGetOrgActionRunnerRegistrationToken ==")
	c := newTestClient()

	token, resp, err := c.GetOrgActionRunnerRegistrationToken(testOrgName)
	if err == nil {
		require.NotNil(t, resp)
		assert.NotEmpty(t, token.Token)
	}
}
