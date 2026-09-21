// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"context"
	"fmt"
	"time"
)

// ActivityOpType is the type of action recorded in an activity feed entry
type ActivityOpType string

const (
	ActivityOpCreateRepo                ActivityOpType = "create_repo"
	ActivityOpRenameRepo                ActivityOpType = "rename_repo"
	ActivityOpStarRepo                  ActivityOpType = "star_repo"
	ActivityOpWatchRepo                 ActivityOpType = "watch_repo"
	ActivityOpCommitRepo                ActivityOpType = "commit_repo"
	ActivityOpCreateIssue               ActivityOpType = "create_issue"
	ActivityOpCreatePullRequest         ActivityOpType = "create_pull_request"
	ActivityOpTransferRepo              ActivityOpType = "transfer_repo"
	ActivityOpPushTag                   ActivityOpType = "push_tag"
	ActivityOpCommentIssue              ActivityOpType = "comment_issue"
	ActivityOpMergePullRequest          ActivityOpType = "merge_pull_request"
	ActivityOpCloseIssue                ActivityOpType = "close_issue"
	ActivityOpReopenIssue               ActivityOpType = "reopen_issue"
	ActivityOpClosePullRequest          ActivityOpType = "close_pull_request"
	ActivityOpReopenPullRequest         ActivityOpType = "reopen_pull_request"
	ActivityOpDeleteTag                 ActivityOpType = "delete_tag"
	ActivityOpDeleteBranch              ActivityOpType = "delete_branch"
	ActivityOpMirrorSyncPush            ActivityOpType = "mirror_sync_push"
	ActivityOpMirrorSyncCreate          ActivityOpType = "mirror_sync_create"
	ActivityOpMirrorSyncDelete          ActivityOpType = "mirror_sync_delete"
	ActivityOpApprovePullRequest        ActivityOpType = "approve_pull_request"
	ActivityOpRejectPullRequest         ActivityOpType = "reject_pull_request"
	ActivityOpCommentPull               ActivityOpType = "comment_pull"
	ActivityOpPublishRelease            ActivityOpType = "publish_release"
	ActivityOpPullReviewDismissed       ActivityOpType = "pull_review_dismissed"
	ActivityOpPullRequestReadyForReview ActivityOpType = "pull_request_ready_for_review"
	ActivityOpAutoMergePullRequest      ActivityOpType = "auto_merge_pull_request"
)

// Activity represents a single activity feed entry
type Activity struct {
	ID        int64          `json:"id"`
	UserID    int64          `json:"user_id"`
	OpType    ActivityOpType `json:"op_type"`
	ActUser   *User          `json:"act_user"`
	ActUserID int64          `json:"act_user_id"`
	Repo      *Repository    `json:"repo"`
	RepoID    int64          `json:"repo_id"`
	Comment   *Comment       `json:"comment"`
	CommentID int64          `json:"comment_id"`
	RefName   string         `json:"ref_name"`
	IsPrivate bool           `json:"is_private"`
	Content   string         `json:"content"`
	Created   time.Time      `json:"created"`
}

// ListActivityFeedsOptions holds optional parameters for listing activity feeds
type ListActivityFeedsOptions struct {
	ListOptions
	// Date restricts the feed to activities on this date (YYYY-MM-DD). Optional.
	Date string
}

// QueryEncode encodes options to query parameters
func (opt *ListActivityFeedsOptions) QueryEncode() string {
	query := opt.getURLQuery()
	if opt.Date != "" {
		query.Add("date", opt.Date)
	}
	return query.Encode()
}

// ListOrgActivityFeeds lists an organization's activity feeds
func (c *Client) ListOrgActivityFeeds(ctx context.Context, org string, opt ListActivityFeedsOptions) ([]*Activity, Response, error) {
	if err := escapeValidatePathSegments(&org); err != nil {
		return nil, Response{}, err
	}
	opt.setDefaults()

	var activities []*Activity
	resp, err := c.getParsedResponseWithContext(ctx, fmt.Sprintf("/orgs/%s/activities/feeds?%s", org, opt.QueryEncode()), &activities)
	return activities, resp, err
}

// ListTeamActivityFeeds lists a team's activity feeds
func (c *Client) ListTeamActivityFeeds(ctx context.Context, id int64, opt ListActivityFeedsOptions) ([]*Activity, Response, error) {
	opt.setDefaults()

	var activities []*Activity
	resp, err := c.getParsedResponseWithContext(ctx, fmt.Sprintf("/teams/%d/activities/feeds?%s", id, opt.QueryEncode()), &activities)
	return activities, resp, err
}
