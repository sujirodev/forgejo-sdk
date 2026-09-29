// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"archive/zip"
	"bytes"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noSuchArtifactID is an id no artifact in a freshly created repository can
// have. It backs the negative path of the artifact routes: TestRepoActionArtifacts
// runs against a workflow no runner executes, so it can only assert the
// documented 404. The positive path, against an artifact a runner really
// uploaded, is TestRepoActionExecutedRun.
const noSuchArtifactID = int64(999999)

// runnerWorkflow is a workflow the harness runner (label "host", see
// scripts/start-test-runner.sh) picks up: one job that writes a file, prints
// a marker and uploads the file as an artifact.
const runnerWorkflow = `on: workflow_dispatch
jobs:
  build:
    runs-on: host
    steps:
      - run: echo hello-from-the-runner | tee out.txt
      - uses: https://code.forgejo.org/forgejo/upload-artifact@v4
        with:
          name: build-output
          path: out.txt
`

// runFinishDeadline bounds the wait for the runner to finish the job. A job
// that writes a line and uploads a few bytes takes seconds once the runner is
// up; the rest of the allowance is the runner still installing node and
// registering when the suite starts.
const runFinishDeadline = 3 * time.Minute

// waitForRun polls the run until it reaches a terminal status, or fails the
// test at the deadline.
func waitForRun(t *testing.T, c *Client, owner, repo string, runID int64) *ActionRun {
	t.Helper()
	deadline := time.Now().Add(runFinishDeadline)
	for {
		run, _, err := c.GetRepoActionRun(owner, repo, runID)
		require.NoError(t, err)
		switch run.Status {
		case "success", "failure", "cancelled", "skipped":
			return run
		}
		require.True(t, time.Now().Before(deadline), "run %d is still %q after %s: is the harness runner up?", runID, run.Status, runFinishDeadline)
		time.Sleep(time.Second)
	}
}

// TestRepoActionExecutedRun runs a workflow on the harness runner and
// exercises what only an executed job leaves behind: its logs and the
// artifact it uploaded (issue #77).
func TestRepoActionExecutedRun(t *testing.T) {
	t.Parallel()
	log.Println("== TestRepoActionExecutedRun ==")
	c := newTestClient()

	if !runnerAttached() {
		t.Skip("no runner attached (FORGEJO_SDK_TEST_RUNNER unset): see scripts/start-test-runner.sh")
	}
	if !serverAtLeast(t, c, "16.0.0") {
		t.Skip("the artifact and job-log routes arrived in Forgejo 16.0.0")
	}

	repo, cleanup := newWorkflowRepoWith(t, c, "executed", "build.yml", runnerWorkflow)
	defer cleanup()
	owner, name := repo.Owner.UserName, repo.Name

	dispatched, _, err := c.DispatchRepoWorkflow(owner, name, "build.yml", DispatchWorkflowOption{Ref: "main", ReturnRunInfo: true})
	require.NoError(t, err)
	require.NotNil(t, dispatched)
	runID := dispatched.ID

	run := waitForRun(t, c, owner, name, runID)
	require.Equal(t, "success", run.Status)

	jobs, _, err := c.ListRepoActionRunJobs(owner, name, runID)
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, "build", jobs[0].Name)
	assert.Equal(t, "success", jobs[0].Status)

	// The job's own log, now that it has run.
	logs, resp, err := c.GetRepoActionJobLogs(owner, name, jobs[0].ID, GetActionJobLogsOption{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	logText, err := io.ReadAll(logs)
	require.NoError(t, err)
	require.NoError(t, logs.Close())
	assert.Contains(t, string(logText), "hello-from-the-runner")

	// Attempt goes on the wire as a query parameter and picks one attempt.
	attempt := jobs[0].Attempt
	require.NotZero(t, attempt)
	logs, _, err = c.GetRepoActionJobLogs(owner, name, jobs[0].ID, GetActionJobLogsOption{Attempt: attempt})
	require.NoError(t, err)
	logText, err = io.ReadAll(logs)
	require.NoError(t, err)
	require.NoError(t, logs.Close())
	assert.Contains(t, string(logText), "hello-from-the-runner")

	// The artifact the upload step produced, through the three listings and
	// the by-id routes.
	artifacts, _, err := c.ListRepoActionRunArtifacts(owner, name, runID, ListActionArtifactsOption{})
	require.NoError(t, err)
	require.Len(t, artifacts, 1)
	artifactID := artifacts[0].ID
	assert.Equal(t, "build-output", artifacts[0].Name)
	assert.Equal(t, runID, artifacts[0].RunID)

	byName, _, err := c.ListRepoActionArtifacts(owner, name, ListActionArtifactsOption{Name: "build-output"})
	require.NoError(t, err)
	require.Len(t, byName, 1)
	assert.Equal(t, artifactID, byName[0].ID)

	artifact, resp, err := c.GetRepoActionArtifact(owner, name, artifactID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, artifactID, artifact.ID)
	assert.Equal(t, "build-output", artifact.Name)
	assert.Equal(t, runID, artifact.RunID)
	assert.False(t, artifact.Expired)

	// The archive is a zip holding the file the job wrote.
	body, resp, err := c.DownloadRepoActionArtifact(owner, name, artifactID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	archive, err := io.ReadAll(body)
	require.NoError(t, err)
	require.NoError(t, body.Close())
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	require.NoError(t, err)
	require.Len(t, zr.File, 1)
	assert.Equal(t, "out.txt", zr.File[0].Name)
	entry, err := zr.File[0].Open()
	require.NoError(t, err)
	content, err := io.ReadAll(entry)
	require.NoError(t, err)
	require.NoError(t, entry.Close())
	assert.Equal(t, "hello-from-the-runner\n", string(content))

	// Deleting only marks the artifact; the archive is removed later.
	resp, err = c.DeleteRepoActionArtifact(owner, name, artifactID)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

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
	// listing to. No runner takes its job (label docker), so no artifact:
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
