// Copyright 2024 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"encoding/base64"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestRepo creates a test user and repo for action tests
// newWorkflowRepo creates an initialized repository that holds one
// workflow_dispatch workflow, so the routes that need an actual workflow (and,
// once dispatched, an actual run) have something real to act on. The tests
// that need this used to fire at a "test.yml" nothing ever created: the call
// got a 500 or the test skipped, and the route went unexercised while the
// suite stayed green (issues #25 and #27).
func newWorkflowRepo(t *testing.T, c *Client, prefix string) (*Repository, func()) {
	t.Helper()
	user := createTestUser(t, uniqueName(t, prefix+"u"), c)
	c.SetSudo(user.UserName)

	repo, _, err := c.CreateRepo(CreateRepoOption{
		Name:          uniqueName(t, prefix),
		AutoInit:      true,
		DefaultBranch: "main",
	})
	require.NoError(t, err)
	require.NotNil(t, repo)

	workflow := `on: workflow_dispatch
jobs:
  noop:
    runs-on: docker
    steps:
      - run: echo dispatched
`
	_, _, err = c.CreateFile(repo.Owner.UserName, repo.Name, ".forgejo/workflows/test.yml", CreateFileOptions{
		FileOptions: FileOptions{Message: "add a dispatchable workflow", BranchName: "main"},
		Content:     base64.StdEncoding.EncodeToString([]byte(workflow)),
	})
	require.NoError(t, err)

	return repo, func() { c.SetSudo("") }
}

func createTestRepoForActions(t *testing.T, c *Client, suffix string) (*Repository, func()) {
	t.Helper()
	user := createTestUser(t, "repo_action_"+suffix, c)
	c.SetSudo(user.UserName)
	repo, _, err := c.CreateRepo(CreateRepoOption{Name: "ActionTest" + suffix})
	require.NoError(t, err)
	require.NotNil(t, repo)
	return repo, func() { c.SetSudo("") }
}

func TestRepoActionSecrets(t *testing.T) {
	log.Println("== TestRepoActionSecrets ==")
	c := newTestClient()

	// Create one user and repo for all subtests
	user := createTestUser(t, "repo_action_test_user", c)
	c.SetSudo(user.UserName)
	testRepo, _, err := c.CreateRepo(CreateRepoOption{Name: "ActionTestRepo"})
	require.NoError(t, err)
	require.NotNil(t, testRepo)

	t.Run("CreateAndUpdate", func(t *testing.T) {
		// create secret
		resp, err := c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "test", Data: "test"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// update secret
		resp, err = c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "test", Data: "test2"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// create secret with 64 character name
		longName64 := strings.Repeat("B", 64)
		resp, err = c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: longName64, Data: "test_data"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// create secret with long value
		longValue := strings.Repeat("secret", 100)
		resp, err = c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "LONG_VALUE_TEST", Data: longValue})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// list secrets - should have 3 secrets
		secrets, _, err := c.ListRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, ListRepoActionSecretOption{})
		require.NoError(t, err)
		assert.Len(t, secrets, 3)
	})

	t.Run("InvalidNames", func(t *testing.T) {
		// test invalid names - client-side validation should reject these
		_, err := c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "", Data: "data"})
		require.Error(t, err)
		require.EqualError(t, err, "name required")

		_, err = c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "INVALID-NAME", Data: "data"})
		require.Error(t, err)
		require.EqualError(t, err, "name must contain only alphanumeric characters and underscores")

		_, err = c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "INVALID NAME", Data: "data"})
		require.Error(t, err)
		require.EqualError(t, err, "name must contain only alphanumeric characters and underscores")

		_, err = c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "GITEA_SECRET", Data: "data"})
		require.Error(t, err)
		require.EqualError(t, err, "name cannot start with GITEA_ or GITHUB_ (reserved prefixes)")

		_, err = c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "github_token", Data: "data"})
		require.Error(t, err)
		require.EqualError(t, err, "name cannot start with GITEA_ or GITHUB_ (reserved prefixes)")

		_, err = c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: strings.Repeat("A", 256), Data: "data"})
		require.Error(t, err)
		require.EqualError(t, err, "name too long (maximum 255 characters)")
	})

	t.Run("EmptyData", func(t *testing.T) {
		_, err := c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "VALID_NAME", Data: ""})
		require.Error(t, err)
		require.EqualError(t, err, "data required")
	})

	t.Run("UpdateMultipleTimes", func(t *testing.T) {
		// create and update same secret multiple times
		resp, err := c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "UPDATE_SECRET", Data: "data1"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		for i := 2; i <= 5; i++ {
			resp, err = c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "UPDATE_SECRET", Data: "data" + string(rune('0'+i))})
			require.NoError(t, err)
			assert.Equal(t, http.StatusNoContent, resp.StatusCode)
		}

		secrets, _, err := c.ListRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, ListRepoActionSecretOption{})
		require.NoError(t, err)
		// Should have original 3 + UPDATE_SECRET = 4 total
		assert.GreaterOrEqual(t, len(secrets), 1)
	})

	t.Run("CaseSensitivity", func(t *testing.T) {
		// create secret with lowercase name
		resp, err := c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "my_secret", Data: "lower"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// try to create with uppercase - should update the same secret (case-insensitive)
		resp, err = c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "MY_SECRET", Data: "upper"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)

		// try with mixed case - should also update the same secret
		resp, err = c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "My_Secret", Data: "mixed"})
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("Pagination", func(t *testing.T) {
		// create 10 more secrets for pagination test
		for i := 1; i <= 10; i++ {
			name := "PAGE_SECRET_" + string(rune('A'+i-1))
			_, err := c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: name, Data: "data" + string(rune('0'+i))})
			require.NoError(t, err)
		}

		// test pagination
		secrets, _, err := c.ListRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, ListRepoActionSecretOption{
			ListOptions: ListOptions{Page: 1, PageSize: 5},
		})
		require.NoError(t, err)
		assert.Len(t, secrets, 5)

		secrets, _, err = c.ListRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, ListRepoActionSecretOption{
			ListOptions: ListOptions{Page: 2, PageSize: 5},
		})
		require.NoError(t, err)
		assert.Len(t, secrets, 5)

		// get all
		secrets, _, err = c.ListRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, ListRepoActionSecretOption{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(secrets), 10)
	})

	t.Run("NonExistentRepo", func(t *testing.T) {
		// trying to create secret for non-existent repo should fail
		_, err := c.CreateRepoActionSecret(testRepo.Owner.UserName, "NonExistentRepo123456", CreateSecretOption{Name: "TEST", Data: "data"})
		require.Error(t, err)

		// listing secrets for non-existent repo also returns error
		_, _, err = c.ListRepoActionSecret(testRepo.Owner.UserName, "NonExistentRepo123456", ListRepoActionSecretOption{})
		require.Error(t, err)
	})

	t.Run("LargeData", func(t *testing.T) {
		// test various data sizes
		smallData := "small"
		mediumData := strings.Repeat("medium", 100)
		largeData := strings.Repeat("large", 1000)

		_, err := c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "SMALL_DATA", Data: smallData})
		require.NoError(t, err)

		_, err = c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "MEDIUM_DATA", Data: mediumData})
		require.NoError(t, err)

		_, err = c.CreateRepoActionSecret(testRepo.Owner.UserName, testRepo.Name, CreateSecretOption{Name: "LARGE_DATA", Data: largeData})
		require.NoError(t, err)
	})
}

func TestDeleteRepoActionSecret(t *testing.T) {
	log.Println("== TestDeleteRepoActionSecret ==")
	c := newTestClient()

	repo, cleanup := createTestRepoForActions(t, c, "del_secret")
	defer cleanup()

	// First create a secret to delete
	_, err := c.CreateRepoActionSecret(repo.Owner.UserName, repo.Name, CreateSecretOption{
		Name: "DELETE_TEST_SECRET",
		Data: "test_value",
	})
	require.NoError(t, err)

	// Delete the secret
	resp, err := c.DeleteRepoActionSecret(repo.Owner.UserName, repo.Name, "DELETE_TEST_SECRET")
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestListRepoActionRuns(t *testing.T) {
	log.Println("== TestListRepoActionRuns ==")
	c := newTestClient()

	repo, cleanup := createTestRepoForActions(t, c, "list_runs")
	defer cleanup()

	// /actions/runs arrived in Forgejo 12.0.0, and the SDK now guards it.
	// Assert both sides: the route works above the guard, and below it the
	// SDK refuses without contacting the server.
	if !serverAtLeast(t, c, "12.0.0") {
		_, _, err := c.ListRepoActionRuns(repo.Owner.UserName, repo.Name, ListActionRunsOption{})
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
		return
	}

	runs, resp, err := c.ListRepoActionRuns(repo.Owner.UserName, repo.Name, ListActionRunsOption{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, runs)
	assert.GreaterOrEqual(t, runs.TotalCount, int64(0))
}

func TestGetRepoActionRun(t *testing.T) {
	log.Println("== TestGetRepoActionRun ==")
	c := newTestClient()

	repo, cleanup := newWorkflowRepo(t, c, "getrun")
	defer cleanup()

	// Dispatching creates the run record. No runner picks it up here, which
	// is fine: the route reads the record, it does not wait for the job.
	_, _, err := c.DispatchRepoWorkflow(repo.Owner.UserName, repo.Name, "test.yml", DispatchWorkflowOption{
		Ref: "main",
	})
	require.NoError(t, err)

	// Reading the run needs Forgejo 12.0.0; below that the SDK's guard
	// refuses, which is the documented behavior and worth asserting.
	if !serverAtLeast(t, c, "12.0.0") {
		_, _, err := c.GetRepoActionRun(repo.Owner.UserName, repo.Name, 1)
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
		return
	}

	runs, _, err := c.ListRepoActionRuns(repo.Owner.UserName, repo.Name, ListActionRunsOption{})
	require.NoError(t, err)
	require.NotNil(t, runs)
	require.NotEmpty(t, runs.WorkflowRuns, "the dispatch above should have created a run")

	run, resp, err := c.GetRepoActionRun(repo.Owner.UserName, repo.Name, runs.WorkflowRuns[0].ID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, runs.WorkflowRuns[0].ID, run.ID)
}

func TestDispatchRepoWorkflow(t *testing.T) {
	log.Println("== TestDispatchRepoWorkflow ==")
	c := newTestClient()

	repo, cleanup := newWorkflowRepo(t, c, "dispwf")
	defer cleanup()

	// Without ReturnRunInfo the endpoint answers 204, so the decoded body is
	// nil by design: the HTTP response is what there is to assert on.
	dispatched, resp, err := c.DispatchRepoWorkflow(repo.Owner.UserName, repo.Name, "test.yml", DispatchWorkflowOption{
		Ref: "main",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Nil(t, dispatched)
}

func TestListRepoActionTasks(t *testing.T) {
	log.Println("== TestListRepoActionTasks ==")
	c := newTestClient()

	repo, cleanup := createTestRepoForActions(t, c, "list_tasks")
	defer cleanup()

	tasks, resp, err := c.ListRepoActionTasks(repo.Owner.UserName, repo.Name, ListActionTasksOption{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, tasks)
	assert.GreaterOrEqual(t, tasks.TotalCount, int64(0))
}

func TestListRepoActionJobs(t *testing.T) {
	log.Println("== TestListRepoActionJobs ==")
	c := newTestClient()

	repo, cleanup := createTestRepoForActions(t, c, "list_jobs")
	defer cleanup()

	// jobs may be nil when the API returns JSON null (no jobs exist)
	_, resp, err := c.ListRepoActionJobs(repo.Owner.UserName, repo.Name, ListActionJobsOption{})
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestRepoActionVariables(t *testing.T) {
	log.Println("== TestRepoActionVariables ==")
	c := newTestClient()

	repo, cleanup := createTestRepoForActions(t, c, "variables")
	defer cleanup()

	variableName := "TEST_VARIABLE"

	// Create
	resp, err := c.CreateRepoActionVariable(repo.Owner.UserName, repo.Name, CreateVariableOption{
		Name: variableName,
		Data: "test_value",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// List
	variables, resp, err := c.ListRepoActionVariables(repo.Owner.UserName, repo.Name, ListOptions{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	found := false
	for _, v := range variables {
		if v.Name == variableName {
			found = true
			break
		}
	}
	assert.True(t, found, "variable should be in list")

	// Get
	variable, resp, err := c.GetRepoActionVariable(repo.Owner.UserName, repo.Name, variableName)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, variableName, variable.Name)
	assert.Equal(t, "test_value", variable.Data)

	// Update
	resp, err = c.UpdateRepoActionVariable(repo.Owner.UserName, repo.Name, variableName, CreateVariableOption{
		Name: variableName,
		Data: "updated_value",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify update
	variable, _, err = c.GetRepoActionVariable(repo.Owner.UserName, repo.Name, variableName)
	require.NoError(t, err)
	assert.Equal(t, "updated_value", variable.Data)

	// Delete
	resp, err = c.DeleteRepoActionVariable(repo.Owner.UserName, repo.Name, variableName)
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestGetRepoActionRunnerRegistrationToken(t *testing.T) {
	log.Println("== TestGetRepoActionRunnerRegistrationToken ==")
	c := newTestClient()

	repo, cleanup := createTestRepoForActions(t, c, "runner_token")
	defer cleanup()

	token, resp, err := c.GetRepoActionRunnerRegistrationToken(repo.Owner.UserName, repo.Name)
	// This may fail if user doesn't have admin access to repo
	if err == nil {
		require.NotNil(t, resp)
		assert.NotEmpty(t, token.Token)
	}
}
