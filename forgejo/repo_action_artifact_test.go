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

// noSuchArtifactID is an id no artifact in a freshly created repository can
// have. The routes that take an artifact id cannot be exercised positively
// here: producing an artifact needs a runner daemon to execute a job and
// upload it over the Actions pipeline protocol, and the test harness stands
// up a Forgejo instance with no runner attached. Declared in
// route-exceptions.json as negative-only.
const noSuchArtifactID = int64(999999)

func TestRepoActionArtifacts(t *testing.T) {
	t.Parallel()
	log.Println("== TestRepoActionArtifacts ==")
	c := newTestClient()

	repo, cleanup := newWorkflowRepo(t, c, "artifacts")
	defer cleanup()

	// The artifact routes arrived in Forgejo 16.0.0. Below that the SDK's
	// guard refuses the call without contacting the server.
	if !serverAtLeast(t, c, "16.0.0") {
		_, _, err := c.ListRepoActionArtifacts(repo.Owner.UserName, repo.Name, ListActionArtifactsOption{})
		require.Error(t, err, "the version guard must refuse the call on a server without the route")
		_, _, err = c.ListRepoActionRunArtifacts(repo.Owner.UserName, repo.Name, 1, ListActionArtifactsOption{})
		require.Error(t, err)
		_, _, err = c.GetRepoActionArtifact(repo.Owner.UserName, repo.Name, noSuchArtifactID)
		require.Error(t, err)
		_, err = c.DeleteRepoActionArtifact(repo.Owner.UserName, repo.Name, noSuchArtifactID)
		require.Error(t, err)
		_, _, err = c.DownloadRepoActionArtifact(repo.Owner.UserName, repo.Name, noSuchArtifactID)
		require.Error(t, err)
		return
	}

	// Dispatching gives the repository a real run to scope the per-run
	// listing to. No runner picks it up, so the run produces no artifact:
	// an empty list is the correct answer, not a missing one.
	_, _, err := c.DispatchRepoWorkflow(repo.Owner.UserName, repo.Name, "test.yml", DispatchWorkflowOption{Ref: "main"})
	require.NoError(t, err)

	runs, _, err := c.ListRepoActionRuns(repo.Owner.UserName, repo.Name, ListActionRunsOption{})
	require.NoError(t, err)
	require.NotEmpty(t, runs.WorkflowRuns, "the dispatch above should have created a run")
	runID := runs.WorkflowRuns[0].ID

	artifacts, resp, err := c.ListRepoActionArtifacts(repo.Owner.UserName, repo.Name, ListActionArtifactsOption{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, artifacts, "no runner ran the job, so the repository has no artifact")

	// The name filter is a query parameter, so it has to survive a call
	// that returns nothing just as well.
	named, _, err := c.ListRepoActionArtifacts(repo.Owner.UserName, repo.Name, ListActionArtifactsOption{Name: "build-output"})
	require.NoError(t, err)
	assert.Empty(t, named)

	runArtifacts, resp, err := c.ListRepoActionRunArtifacts(repo.Owner.UserName, repo.Name, runID, ListActionArtifactsOption{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, runArtifacts)

	// A run that does not exist is a 404, which also proves the per-run
	// path is the one being requested and not the repository-wide one.
	_, _, err = c.ListRepoActionRunArtifacts(repo.Owner.UserName, repo.Name, runID+10000, ListActionArtifactsOption{})
	require.Error(t, err)

	// The three id-taking routes: with no artifact to point at, the
	// documented 404 is the outcome the test asserts.
	_, _, err = c.GetRepoActionArtifact(repo.Owner.UserName, repo.Name, noSuchArtifactID)
	require.EqualError(t, err, "resource does not exist")

	_, err = c.DeleteRepoActionArtifact(repo.Owner.UserName, repo.Name, noSuchArtifactID)
	require.EqualError(t, err, "resource does not exist")

	body, _, err := c.DownloadRepoActionArtifact(repo.Owner.UserName, repo.Name, noSuchArtifactID)
	require.EqualError(t, err, "resource does not exist")
	// getResponseReader hands back the error body rather than nil, so it is
	// still the caller's to close.
	require.NotNil(t, body)
	require.NoError(t, body.Close())
}
