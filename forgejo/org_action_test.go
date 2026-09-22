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

// newActionTestOrg creates an organization owned by a throwaway user and
// returns its name.
//
// The tests below used to target a fixed "testorg" that nothing in the suite
// ever created. Every call against it answered 404, the test skipped, and the
// eight routes it was meant to cover were never exercised while the suite
// stayed green -- see issue #27. Each test now owns the organization it acts
// on, like TestOrgActionSecrets already did.
func newActionTestOrg(t *testing.T, c *Client, prefix string) string {
	t.Helper()
	user := createTestUser(t, uniqueName(t, prefix+"u"), c)
	c.SetSudo(user.UserName)
	org, _, err := c.CreateOrg(CreateOrgOption{Name: uniqueName(t, prefix)})
	require.NoError(t, err)
	require.NotNil(t, org)
	return org.UserName
}

func TestOrgActionSecrets(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	log.Println("== TestDeleteOrgActionSecret ==")
	c := newTestClient()

	org := newActionTestOrg(t, c, "secdel")

	// First create a secret to delete. A failure here is a failure, not a
	// reason to skip: skipping is what hid this route for so long.
	_, err := c.CreateOrgActionSecret(org, CreateSecretOption{
		Name: "DELETE_TEST_SECRET",
		Data: "test_value",
	})
	require.NoError(t, err)

	resp, err := c.DeleteOrgActionSecret(org, "DELETE_TEST_SECRET")
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestListOrgActionJobs(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	log.Println("== TestOrgActionVariables ==")
	c := newTestClient()

	org := newActionTestOrg(t, c, "orgvar")
	variableName := "TEST_ORG_VARIABLE"

	// Create. A failure here used to skip the whole test, taking the four
	// routes below with it.
	resp, err := c.CreateOrgActionVariable(org, CreateVariableOption{
		Name: variableName,
		Data: "test_value",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// List
	variables, resp, err := c.ListOrgActionVariables(org, ListOptions{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, variables)

	// Get
	variable, resp, err := c.GetOrgActionVariable(org, variableName)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, variableName, variable.Name)

	// Update
	resp, err = c.UpdateOrgActionVariable(org, variableName, CreateVariableOption{
		Name: variableName,
		Data: "updated_value",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Delete
	resp, err = c.DeleteOrgActionVariable(org, variableName)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestGetOrgActionRunnerRegistrationToken(t *testing.T) {
	t.Parallel()
	log.Println("== TestGetOrgActionRunnerRegistrationToken ==")
	c := newTestClient()

	org := newActionTestOrg(t, c, "runtok")

	token, resp, err := c.GetOrgActionRunnerRegistrationToken(org)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, token.Token)
}

func TestOrgRunnersCRUD(t *testing.T) {
	log.Println("== TestOrgRunnersCRUD ==")
	c := newTestClient()

	orgName := "OrgRunnersTestOrg"
	_, _, err := c.GetOrg(orgName)
	if err == nil {
		_, _ = c.DeleteOrg(orgName)
	}
	_, _, err = c.CreateOrg(CreateOrgOption{Name: orgName, Visibility: VisibleTypePublic})
	require.NoError(t, err)
	defer func() { _, _ = c.DeleteOrg(orgName) }()

	// Confirmed absent on a live 13.0.0 instance and present by 15.0.9;
	// below that the SDK's guard refuses the call, which is the documented
	// behavior and worth asserting.
	if !serverAtLeast(t, c, "15.0.0") {
		_, _, err := c.ListOrgRunners(orgName, ListOrgRunnersOption{})
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
		return
	}

	// Listing should always succeed, even with zero runners registered.
	runners, resp, err := c.ListOrgRunners(orgName, ListOrgRunnersOption{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotNil(t, runners)

	registered, resp, err := c.RegisterOrgRunner(orgName, RegisterRunnerOption{
		Name:        "sdk-test-runner",
		Description: "runner registered by the SDK test suite",
	})
	if err != nil {
		t.Skipf("could not register an org runner, skipping: %v", err)
	}
	require.NotNil(t, resp)
	require.NotEmpty(t, registered.Token)
	require.NotZero(t, registered.ID)
	defer func() { _, _ = c.DeleteOrgRunner(orgName, registered.ID) }()

	runner, resp, err := c.GetOrgRunner(orgName, registered.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "sdk-test-runner", runner.Name)
	assert.Equal(t, registered.UUID, runner.UUID)

	resp, err = c.DeleteOrgRunner(orgName, registered.ID)
	require.NoError(t, err)
	require.NotNil(t, resp)

	_, _, err = c.GetOrgRunner(orgName, registered.ID)
	assert.Error(t, err)
}
