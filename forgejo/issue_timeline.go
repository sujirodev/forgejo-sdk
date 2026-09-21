// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"net/url"
	"time"
)

// TimelineComment represents a timeline comment (comment of any type) on an issue or pull request
type TimelineComment struct {
	ID              int64        `json:"id"`
	HTMLURL         string       `json:"html_url"`
	PRURL           string       `json:"pull_request_url"`
	IssueURL        string       `json:"issue_url"`
	Poster          *User        `json:"user"`
	Body            string       `json:"body"`
	Created         time.Time    `json:"created_at"`
	Updated         time.Time    `json:"updated_at"`
	Type            string       `json:"type"`
	Assignee        *User        `json:"assignee"`
	AssigneeTeam    *Team        `json:"assignee_team"`
	RemovedAssignee bool         `json:"removed_assignee"`
	ResolveDoer     *User        `json:"resolve_doer"`
	Label           *Label       `json:"label"`
	Milestone       *Milestone   `json:"milestone"`
	OldMilestone    *Milestone   `json:"old_milestone"`
	ProjectID       int64        `json:"project_id"`
	OldProjectID    int64        `json:"old_project_id"`
	NewTitle        string       `json:"new_title"`
	OldTitle        string       `json:"old_title"`
	NewRef          string       `json:"new_ref"`
	OldRef          string       `json:"old_ref"`
	DependentIssue  *Issue       `json:"dependent_issue"`
	RefIssue        *Issue       `json:"ref_issue"`
	RefComment      *Comment     `json:"ref_comment"`
	RefAction       string       `json:"ref_action"`
	RefCommitSHA    string       `json:"ref_commit_sha"`
	ReviewID        int64        `json:"review_id"`
	TrackedTime     *TrackedTime `json:"tracked_time"`
}

// ListIssueTimelineOptions list timeline options
type ListIssueTimelineOptions struct {
	ListOptions
	Since  time.Time
	Before time.Time
}

// QueryEncode turns options into querystring argument
func (opt *ListIssueTimelineOptions) QueryEncode() string {
	query := opt.getURLQuery()
	if !opt.Since.IsZero() {
		query.Add("since", opt.Since.Format(time.RFC3339))
	}
	if !opt.Before.IsZero() {
		query.Add("before", opt.Before.Format(time.RFC3339))
	}
	return query.Encode()
}

// ListIssueTimeline lists all comments and events on an issue
func (c *Client) ListIssueTimeline(owner, repo string, index int64, opt ListIssueTimelineOptions) ([]*TimelineComment, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()
	link, _ := url.Parse(fmt.Sprintf("/repos/%s/%s/issues/%d/timeline", owner, repo, index))
	link.RawQuery = opt.QueryEncode()
	timeline := make([]*TimelineComment, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", link.String(), nil, nil, &timeline)
	return timeline, resp, err
}
