// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"net/url"
	"time"
)

// ActionRun represents a workflow run
type ActionRun struct {
	ID                int64       `json:"id"`
	RunNumber         int64       `json:"index_in_repo"`
	WorkflowID        string      `json:"workflow_id"`
	Title             string      `json:"title"`
	Status            string      `json:"status"`
	Event             string      `json:"event"`
	CommitSHA         string      `json:"commit_sha"`
	PrettyRef         string      `json:"prettyref"`
	HTMLURL           string      `json:"html_url"`
	TriggerUser       *User       `json:"trigger_user"`
	Repository        *Repository `json:"repository"`
	Created           time.Time   `json:"created"`
	Started           time.Time   `json:"started"`
	Stopped           time.Time   `json:"stopped"`
	Updated           time.Time   `json:"updated"`
	NeedApproval      bool        `json:"need_approval"`
	ApprovedBy        int64       `json:"approved_by"`
	IsForkPullRequest bool        `json:"is_fork_pull_request"`
	IsRefDeleted      bool        `json:"is_ref_deleted"`
}

// ActionTask represents a task within a workflow run
type ActionTask struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	WorkflowID   string    `json:"workflow_id"`
	Status       string    `json:"status"`
	Event        string    `json:"event"`
	DisplayTitle string    `json:"display_title"`
	HeadBranch   string    `json:"head_branch"`
	HeadSHA      string    `json:"head_sha"`
	RunNumber    int64     `json:"run_number"`
	URL          string    `json:"url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	RunStartedAt time.Time `json:"run_started_at"`
}

// ActionVariable represents an action variable
type ActionVariable struct {
	Name    string `json:"name"`
	Data    string `json:"data"`
	OwnerID int64  `json:"owner_id"`
	RepoID  int64  `json:"repo_id"`
}

// ActionRunJob represents a job within a workflow run
type ActionRunJob struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	TaskID int64  `json:"task_id"`
	// RunID identifies the workflow run this job belongs to.
	RunID   int64 `json:"run_id"`
	OwnerID int64 `json:"owner_id"`
	RepoID  int64 `json:"repo_id"`
	// Attempt counts how many times the job has been attempted, including
	// the current one. It is what GetActionJobLogsOption.Attempt selects.
	Attempt int64 `json:"attempt"`
	// Handle opaquely identifies a single attempt of a job.
	Handle string   `json:"handle"`
	RunsOn []string `json:"runs_on"`
	Needs  []string `json:"needs"`
}

// RunnerRegistrationToken represents a runner registration token
type RunnerRegistrationToken struct {
	Token string `json:"token"`
}

// ActionRunner represents a Forgejo Actions runner
type ActionRunner struct {
	ID          int64  `json:"id"`
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	OwnerID     int64  `json:"owner_id"`
	RepoID      int64  `json:"repo_id"`
	Description string `json:"description"`
	// Status is one of "offline", "idle" or "active".
	Status    string   `json:"status"`
	Ephemeral bool     `json:"ephemeral"`
	Labels    []string `json:"labels"`
}

// ListActionRunnersOptions options for listing runners
type ListActionRunnersOptions struct {
	ListOptions
	// Visible includes all runners visible to the caller, not just the ones
	// directly owned by the scope being queried.
	Visible bool
}

// QueryEncode encodes options to query parameters
func (opt *ListActionRunnersOptions) QueryEncode() string {
	query := opt.getURLQuery()
	if opt.Visible {
		query.Add("visible", "true")
	}
	return query.Encode()
}

// RegisterRunnerOption options for registering a new runner
type RegisterRunnerOption struct {
	// Name of the runner to register. It does not have to be unique.
	Name string `json:"name"`
	// Description provides optional details about this runner.
	Description string `json:"description,omitempty"`
	// Ephemeral registers a runner that only takes a single job before being
	// removed. See https://forgejo.org/docs/latest/admin/actions/security/#ephemeral-runner
	Ephemeral bool `json:"ephemeral,omitempty"`
}

// RegisterRunnerResponse contains the details of the just-registered runner,
// including the token needed to configure the runner daemon.
type RegisterRunnerResponse struct {
	ID    int64  `json:"id"`
	Token string `json:"token"`
	UUID  string `json:"uuid"`
}

// ListActionRunsOption options for listing action runs
type ListActionRunsOption struct {
	ListOptions
	Event     string `json:"event"`
	Status    string `json:"status"`
	RunNumber int64  `json:"run_number"`
	HeadSHA   string `json:"head_sha"`
}

// QueryEncode encodes options to query parameters
func (opt *ListActionRunsOption) QueryEncode() string {
	query := opt.getURLQuery()
	if opt.Event != "" {
		query.Add("event", opt.Event)
	}
	if opt.Status != "" {
		query.Add("status", opt.Status)
	}
	if opt.RunNumber > 0 {
		query.Add("run_number", fmt.Sprintf("%d", opt.RunNumber))
	}
	if opt.HeadSHA != "" {
		query.Add("head_sha", opt.HeadSHA)
	}
	return query.Encode()
}

// ListActionRunsResponse paginated list of action runs
type ListActionRunsResponse struct {
	TotalCount   int64        `json:"total_count"`
	WorkflowRuns []*ActionRun `json:"workflow_runs"`
}

// ListActionTasksOption options for listing action tasks
type ListActionTasksOption struct {
	ListOptions
}

// ListActionTasksResponse paginated list of action tasks
type ListActionTasksResponse struct {
	TotalCount   int64         `json:"total_count"`
	WorkflowRuns []*ActionTask `json:"workflow_runs"`
}

// ActionArtifact represents an artifact produced by a workflow run
type ActionArtifact struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	// RunID identifies the workflow run that produced this artifact.
	RunID       int64 `json:"run_id"`
	SizeInBytes int64 `json:"size_in_bytes"`
	// ArchiveDownloadURL is the absolute URL of the zip archive. Prefer
	// DownloadRepoActionArtifact, which goes through the client's
	// authentication.
	ArchiveDownloadURL string `json:"archive_download_url"`
	// Expired reports whether the archive is past its retention and no
	// longer downloadable, even though the record is still listed.
	Expired   bool      `json:"expired"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ListActionArtifactsOption options for listing action artifacts
type ListActionArtifactsOption struct {
	ListOptions
	// Name keeps only the artifacts carrying this exact name. Artifact
	// names are not unique: one name may cover several runs.
	Name string
}

// QueryEncode encodes options to query parameters
func (opt *ListActionArtifactsOption) QueryEncode() string {
	query := opt.getURLQuery()
	if opt.Name != "" {
		query.Add("name", opt.Name)
	}
	return query.Encode()
}

// GetActionJobLogsOption options for reading an action job's logs
type GetActionJobLogsOption struct {
	// Attempt selects one attempt of the job, counting from 1. Zero leaves
	// the choice to the server, which serves the latest attempt.
	Attempt int64
}

// QueryEncode encodes options to query parameters
func (opt *GetActionJobLogsOption) QueryEncode() string {
	query := make(url.Values)
	if opt.Attempt > 0 {
		query.Add("attempt", fmt.Sprintf("%d", opt.Attempt))
	}
	return query.Encode()
}

// ListActionJobsOption options for listing/searching action jobs
type ListActionJobsOption struct {
	Labels string `json:"labels"`
}

// DispatchWorkflowOption options for triggering a workflow
type DispatchWorkflowOption struct {
	Ref           string            `json:"ref"`
	Inputs        map[string]string `json:"inputs,omitempty"`
	ReturnRunInfo bool              `json:"return_run_info,omitempty"`
}

// DispatchWorkflowResponse response from dispatching a workflow
type DispatchWorkflowResponse struct {
	ID        int64    `json:"id"`
	RunNumber int64    `json:"run_number"`
	Jobs      []string `json:"jobs"` // job names, not full ActionRunJob records
}

// CreateVariableOption options for creating/updating a variable
type CreateVariableOption struct {
	Name string `json:"name"`
	Data string `json:"value"`
}

// Validate checks if the CreateVariableOption is valid.
func (opt *CreateVariableOption) Validate() error {
	if len(opt.Name) == 0 {
		return fmt.Errorf("name required")
	}
	if len(opt.Data) == 0 {
		return fmt.Errorf("data required")
	}
	return nil
}

// GetActionsRun returns the workflow run associated with the token used to
// authenticate the request. Unlike the rest of the Client's methods, this
// one must be authenticated with the automatic actions token
// (`forgejo.token` / `ACTIONS_RUNTIME_TOKEN` in a running job, set via
// SetToken like any other token) rather than a personal access token; the
// token is tied to the job and only valid while it is still running.
func (c *Client) GetActionsRun() (*ActionRun, *Response, error) {
	run := new(ActionRun)
	resp, err := c.getParsedResponse("GET", "/actions/run", jsonHeader, nil, run)
	return run, resp, err
}
