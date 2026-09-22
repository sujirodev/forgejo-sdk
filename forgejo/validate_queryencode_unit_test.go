// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

// Table tests for every exported option type's Validate and QueryEncode
// method that isn't already covered by an integration test making the
// corresponding call. These are pure functions (Validate additionally
// needs a Client only to read its cached server version, never the
// network) so they belong in the fase 8 batch of TestUnit_* tests: they
// run under `make test-unit`, no Forgejo instance required, and they
// count on every matrix leg regardless of which routes that leg's server
// actually answers.
//
// See TESTING.md, "The four things a new route needs", item 1.

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUnitVersionedClient(t *testing.T, version string) *Client {
	t.Helper()
	c, err := NewClient("http://example.invalid", SetForgejoVersion(version))
	require.NoError(t, err)
	return c
}

// --- action.go ---

func TestUnit_ListActionRunnersOptions_QueryEncode(t *testing.T) {
	assert.NotContains(t, (&ListActionRunnersOptions{}).QueryEncode(), "visible")
	assert.Contains(t, (&ListActionRunnersOptions{Visible: true}).QueryEncode(), "visible=true")
}

func TestUnit_ListActionRunsOption_QueryEncode(t *testing.T) {
	q := (&ListActionRunsOption{Event: "push", Status: "success", RunNumber: 5, HeadSHA: "abc123"}).QueryEncode()
	for _, want := range []string{"event=push", "status=success", "run_number=5", "head_sha=abc123"} {
		assert.Contains(t, q, want)
	}
	empty := (&ListActionRunsOption{}).QueryEncode()
	for _, absent := range []string{"event=", "status=", "run_number=", "head_sha="} {
		assert.NotContains(t, empty, absent)
	}
}

func TestUnit_CreateVariableOption_Validate(t *testing.T) {
	require.EqualError(t, (&CreateVariableOption{}).Validate(), "name required")
	require.EqualError(t, (&CreateVariableOption{Name: "n"}).Validate(), "data required")
	require.NoError(t, (&CreateVariableOption{Name: "n", Data: "v"}).Validate())
}

// --- admin_user.go ---

func TestUnit_CreateUserOption_Validate(t *testing.T) {
	require.EqualError(t, CreateUserOption{}.Validate(), "email is empty")
	require.EqualError(t, CreateUserOption{Email: "a@b.c"}.Validate(), "username is empty")
	require.NoError(t, CreateUserOption{Email: "a@b.c", Username: "u"}.Validate())
}

// --- hook.go ---

func TestUnit_CreateHookOption_Validate(t *testing.T) {
	require.EqualError(t, CreateHookOption{}.Validate(), "hook type needed")
	require.NoError(t, CreateHookOption{Type: "gitea"}.Validate())
}

// --- issue.go ---

func TestUnit_ListIssueOption_QueryEncode(t *testing.T) {
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	before := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	opt := &ListIssueOption{
		State:       StateClosed,
		Type:        IssueTypeIssue,
		Labels:      []string{"bug", "wip"},
		KeyWord:     "crash",
		Milestones:  []string{"v1", "v2"},
		Since:       since,
		Before:      before,
		CreatedBy:   "alice",
		AssignedBy:  "bob",
		MentionedBy: "carol",
		Owner:       "org",
		Team:        "core",
	}
	q, err := url.ParseQuery(opt.QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, "closed", q.Get("state"))
	assert.Equal(t, "bug,wip", q.Get("labels"))
	assert.Equal(t, "crash", q.Get("q"))
	assert.Equal(t, "issues", q.Get("type"))
	assert.Equal(t, "v1,v2", q.Get("milestones"))
	assert.Equal(t, since.Format(time.RFC3339), q.Get("since"))
	assert.Equal(t, before.Format(time.RFC3339), q.Get("before"))
	assert.Equal(t, "alice", q.Get("created_by"))
	assert.Equal(t, "bob", q.Get("assigned_by"))
	assert.Equal(t, "carol", q.Get("mentioned_by"))
	assert.Equal(t, "org", q.Get("owner"))
	// team is a documented bug: the SDK encodes MentionedBy under "team".
	assert.Equal(t, "carol", q.Get("team"))

	minimal := (&ListIssueOption{}).QueryEncode()
	assert.NotContains(t, minimal, "state=")
	assert.NotContains(t, minimal, "labels=")
	assert.NotContains(t, minimal, "since=")
}

func TestUnit_CreateIssueOption_Validate(t *testing.T) {
	require.EqualError(t, CreateIssueOption{}.Validate(), "title is empty")
	require.EqualError(t, CreateIssueOption{Title: "   "}.Validate(), "title is empty")
	require.NoError(t, CreateIssueOption{Title: "ok"}.Validate())
}

func TestUnit_EditIssueOption_Validate(t *testing.T) {
	require.NoError(t, EditIssueOption{}.Validate(), "empty title leaves the field untouched")
	require.EqualError(t, EditIssueOption{Title: "   "}.Validate(), "title is empty")
	require.NoError(t, EditIssueOption{Title: "ok"}.Validate())
}

// --- issue_comment.go ---

func TestUnit_ListRepoIssueCommentOptions_QueryEncode(t *testing.T) {
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	before := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	q, err := url.ParseQuery((&ListRepoIssueCommentOptions{Since: since, Before: before}).QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, since.Format(time.RFC3339), q.Get("since"))
	assert.Equal(t, before.Format(time.RFC3339), q.Get("before"))
	empty := (&ListRepoIssueCommentOptions{}).QueryEncode()
	assert.NotContains(t, empty, "since=")
	assert.NotContains(t, empty, "before=")
}

func TestUnit_ListIssueCommentOptions_QueryEncode(t *testing.T) {
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	q, err := url.ParseQuery((&ListIssueCommentOptions{Since: since}).QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, since.Format(time.RFC3339), q.Get("since"))
	assert.NotContains(t, (&ListIssueCommentOptions{}).QueryEncode(), "since=")
}

func TestUnit_CreateIssueCommentOption_Validate(t *testing.T) {
	require.EqualError(t, CreateIssueCommentOption{}.Validate(), "body is empty")
	require.NoError(t, CreateIssueCommentOption{Body: "ok"}.Validate())
}

func TestUnit_EditIssueCommentOption_Validate(t *testing.T) {
	require.EqualError(t, EditIssueCommentOption{}.Validate(), "body is empty")
	require.NoError(t, EditIssueCommentOption{Body: "ok"}.Validate())
}

// --- issue_label.go ---

func TestUnit_CreateLabelOption_Validate(t *testing.T) {
	require.EqualError(t, CreateLabelOption{Color: "zzzzzz", Name: "x"}.Validate(), "invalid color format")
	require.EqualError(t, CreateLabelOption{Color: "00aabb", Name: "  "}.Validate(), "empty name not allowed")
	require.NoError(t, CreateLabelOption{Color: "#00aabb", Name: "bug"}.Validate())
	require.NoError(t, CreateLabelOption{Color: "00aabb", Name: "bug"}.Validate())
}

func TestUnit_EditLabelOption_Validate(t *testing.T) {
	badColor := "zzzzzz"
	require.EqualError(t, EditLabelOption{Color: &badColor}.Validate(), "invalid color format")
	goodColor := "#00aabb"
	require.NoError(t, EditLabelOption{Color: &goodColor}.Validate())
	blank := "   "
	require.EqualError(t, EditLabelOption{Name: &blank}.Validate(), "empty name not allowed")
	name := "bug"
	require.NoError(t, EditLabelOption{Name: &name}.Validate())
	require.NoError(t, EditLabelOption{}.Validate(), "nil Color/Name leave the fields untouched")
}

// --- issue_milestone.go ---

func TestUnit_ListMilestoneOption_QueryEncode(t *testing.T) {
	q, err := url.ParseQuery((&ListMilestoneOption{State: StateClosed, Name: "v1"}).QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, "closed", q.Get("state"))
	assert.Equal(t, "v1", q.Get("name"))
	empty := (&ListMilestoneOption{}).QueryEncode()
	assert.NotContains(t, empty, "state=")
	assert.NotContains(t, empty, "name=")
}

func TestUnit_CreateMilestoneOption_Validate(t *testing.T) {
	require.EqualError(t, CreateMilestoneOption{}.Validate(), "title is empty")
	require.NoError(t, CreateMilestoneOption{Title: "v1"}.Validate())
}

func TestUnit_EditMilestoneOption_Validate(t *testing.T) {
	require.NoError(t, EditMilestoneOption{}.Validate())
	require.EqualError(t, EditMilestoneOption{Title: "   "}.Validate(), "title is empty")
	require.NoError(t, EditMilestoneOption{Title: "v1"}.Validate())
}

// --- issue_tracked_time.go ---

func TestUnit_ListTrackedTimesOptions_QueryEncode(t *testing.T) {
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	before := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	q, err := url.ParseQuery((&ListTrackedTimesOptions{Since: since, Before: before, User: "alice"}).QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, since.Format(time.RFC3339), q.Get("since"))
	assert.Equal(t, before.Format(time.RFC3339), q.Get("before"))
	assert.Equal(t, "alice", q.Get("user"))
	assert.NotContains(t, (&ListTrackedTimesOptions{}).QueryEncode(), "user=")
}

func TestUnit_AddTimeOption_Validate(t *testing.T) {
	require.EqualError(t, AddTimeOption{}.Validate(), "no time to add")
	require.NoError(t, AddTimeOption{Time: 60}.Validate())
}

// --- notifications.go ---

func TestUnit_ListNotificationOptions_QueryEncode(t *testing.T) {
	since := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	before := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	opt := &ListNotificationOptions{
		Since:        since,
		Before:       before,
		Status:       []NotifyStatus{NotifyStatusUnread},
		SubjectTypes: []NotifySubjectType{NotifySubjectIssue},
	}
	q, err := url.ParseQuery(opt.QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, since.Format(time.RFC3339), q.Get("since"))
	assert.Equal(t, before.Format(time.RFC3339), q.Get("before"))
	assert.Equal(t, "unread", q.Get("status-types"))
	assert.Equal(t, "Issue", q.Get("subject-type"))
}

func TestUnit_ListNotificationOptions_Validate(t *testing.T) {
	old := newUnitVersionedClient(t, "1.12.0")
	require.NoError(t, ListNotificationOptions{}.Validate(old), "no status filter never checks the version")
	require.Error(t, ListNotificationOptions{Status: []NotifyStatus{NotifyStatusUnread}}.Validate(old))

	new := newUnitVersionedClient(t, "1.12.3")
	require.NoError(t, ListNotificationOptions{Status: []NotifyStatus{NotifyStatusUnread}}.Validate(new))
}

func TestUnit_MarkNotificationOptions_QueryEncode(t *testing.T) {
	lastRead := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	opt := &MarkNotificationOptions{
		LastReadAt: lastRead,
		Status:     []NotifyStatus{NotifyStatusUnread},
		ToStatus:   NotifyStatusUnread,
	}
	q, err := url.ParseQuery(opt.QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, lastRead.Format(time.RFC3339), q.Get("last_read_at"))
	assert.Equal(t, "unread", q.Get("status-types"))
	assert.Equal(t, "unread", q.Get("to-status"))
	assert.Empty(t, (&MarkNotificationOptions{}).QueryEncode())
}

func TestUnit_MarkNotificationOptions_Validate(t *testing.T) {
	old := newUnitVersionedClient(t, "1.12.0")
	require.NoError(t, MarkNotificationOptions{}.Validate(old))
	require.Error(t, MarkNotificationOptions{Status: []NotifyStatus{NotifyStatusUnread}}.Validate(old))
	require.Error(t, MarkNotificationOptions{ToStatus: NotifyStatusUnread}.Validate(old))

	new := newUnitVersionedClient(t, "1.12.3")
	require.NoError(t, MarkNotificationOptions{ToStatus: NotifyStatusUnread}.Validate(new))
}

// --- org.go ---

func TestUnit_CreateOrgOption_Validate(t *testing.T) {
	require.EqualError(t, CreateOrgOption{}.Validate(), "empty org name")
	require.EqualError(t, CreateOrgOption{Name: "acme", Visibility: "bogus"}.Validate(), "invalid visibility option")
	require.NoError(t, CreateOrgOption{Name: "acme"}.Validate())
	require.NoError(t, CreateOrgOption{Name: "acme", Visibility: VisibleTypePrivate}.Validate())
}

func TestUnit_EditOrgOption_Validate(t *testing.T) {
	require.EqualError(t, EditOrgOption{Visibility: "bogus"}.Validate(), "invalid visibility option")
	require.NoError(t, EditOrgOption{}.Validate())
	require.NoError(t, EditOrgOption{Visibility: VisibleTypeLimited}.Validate())
}

// --- org_action.go ---

func TestUnit_CreateSecretOption_Validate(t *testing.T) {
	require.EqualError(t, (&CreateSecretOption{}).Validate(), "name required")
	require.EqualError(t, (&CreateSecretOption{Name: strings.Repeat("a", 256)}).Validate(), "name too long (maximum 255 characters)")
	require.EqualError(t, (&CreateSecretOption{Name: "bad name!"}).Validate(), "name must contain only alphanumeric characters and underscores")
	require.EqualError(t, (&CreateSecretOption{Name: "GITEA_TOKEN"}).Validate(), "name cannot start with GITEA_ or GITHUB_ (reserved prefixes)")
	require.EqualError(t, (&CreateSecretOption{Name: "github_pat"}).Validate(), "name cannot start with GITEA_ or GITHUB_ (reserved prefixes)")
	require.EqualError(t, (&CreateSecretOption{Name: "OK_NAME"}).Validate(), "data required")
	require.NoError(t, (&CreateSecretOption{Name: "OK_NAME", Data: "v"}).Validate())
}

// --- org_team.go ---

func TestUnit_CreateTeamOption_Validate(t *testing.T) {
	promoted := &CreateTeamOption{Permission: AccessModeOwner, Name: "n"}
	require.NoError(t, promoted.Validate())
	assert.Equal(t, AccessModeAdmin, promoted.Permission, "owner is downgraded to admin")

	require.EqualError(t, (&CreateTeamOption{Permission: "bogus"}).Validate(), "permission mode invalid")
	require.EqualError(t, (&CreateTeamOption{Permission: AccessModeRead}).Validate(), "name required")
	require.EqualError(t, (&CreateTeamOption{Permission: AccessModeRead, Name: strings.Repeat("a", 256)}).Validate(), "name too long")
	require.EqualError(t, (&CreateTeamOption{Permission: AccessModeRead, Name: "n", Description: strings.Repeat("a", 256)}).Validate(), "description too long")
	require.EqualError(t, (&CreateTeamOption{Permission: AccessModeRead, Name: "n", Units: []RepoUnitType{RepoUnitCode}}).Validate(), "variable Units should be replaced by UnitsMap")
	require.NoError(t, (&CreateTeamOption{Permission: AccessModeWrite, Name: "n"}).Validate())
}

func TestUnit_EditTeamOption_Validate(t *testing.T) {
	promoted := &EditTeamOption{Permission: AccessModeOwner, Name: "n"}
	require.NoError(t, promoted.Validate())
	assert.Equal(t, AccessModeAdmin, promoted.Permission)

	require.EqualError(t, (&EditTeamOption{Permission: "bogus"}).Validate(), "permission mode invalid")
	require.EqualError(t, (&EditTeamOption{Permission: AccessModeAdmin}).Validate(), "name required")
	require.EqualError(t, (&EditTeamOption{Permission: AccessModeAdmin, Name: strings.Repeat("a", 31)}).Validate(), "name to long")
	longDesc := strings.Repeat("a", 256)
	require.EqualError(t, (&EditTeamOption{Permission: AccessModeAdmin, Name: "n", Description: &longDesc}).Validate(), "description to long")
	require.EqualError(t, (&EditTeamOption{Permission: AccessModeAdmin, Name: "n", Units: []RepoUnitType{RepoUnitCode}}).Validate(), "variable Units should be replaced by UnitsMap")
	require.NoError(t, (&EditTeamOption{Permission: AccessModeAdmin, Name: "n"}).Validate())
}

// --- pull.go ---

func TestUnit_ListPullRequestsOptions_QueryEncode(t *testing.T) {
	q, err := url.ParseQuery((&ListPullRequestsOptions{State: StateClosed, Sort: "oldest", Milestone: 3}).QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, "closed", q.Get("state"))
	assert.Equal(t, "oldest", q.Get("sort"))
	assert.Equal(t, "3", q.Get("milestone"))
	empty := (&ListPullRequestsOptions{}).QueryEncode()
	assert.NotContains(t, empty, "state=")
	assert.NotContains(t, empty, "milestone=")
}

func TestUnit_EditPullRequestOption_Validate(t *testing.T) {
	old := newUnitVersionedClient(t, "1.11.0")
	require.EqualError(t, EditPullRequestOption{Title: "   "}.Validate(old), "title is empty")
	require.NoError(t, EditPullRequestOption{}.Validate(old))
	require.EqualError(t, EditPullRequestOption{Base: "main"}.Validate(old), "can not change base forgejo to old")

	new := newUnitVersionedClient(t, "1.12.0")
	require.NoError(t, EditPullRequestOption{Base: "main"}.Validate(new))
}

func TestUnit_MergePullRequestOption_Validate(t *testing.T) {
	old := newUnitVersionedClient(t, "1.11.0")
	require.NoError(t, MergePullRequestOption{Style: MergeStyleMerge}.Validate(old))
	require.Error(t, MergePullRequestOption{Style: MergeStyleSquash}.Validate(old))

	new := newUnitVersionedClient(t, "1.11.5")
	require.NoError(t, MergePullRequestOption{Style: MergeStyleSquash}.Validate(new))
}

// --- pull_review.go ---

func TestUnit_CreatePullReviewOptions_Validate(t *testing.T) {
	require.EqualError(t, CreatePullReviewOptions{}.Validate(), "body is empty")
	require.NoError(t, CreatePullReviewOptions{State: ReviewStateApproved}.Validate(), "an approval needs no body")
	require.NoError(t, CreatePullReviewOptions{Body: "lgtm"}.Validate())
	require.NoError(t, CreatePullReviewOptions{Comments: []CreatePullReviewComment{{Body: "nit"}}}.Validate())

	require.EqualError(t,
		CreatePullReviewOptions{Body: "ok", Comments: []CreatePullReviewComment{{Body: ""}}}.Validate(),
		"body is empty",
		"a nested comment's own Validate is checked too")
}

func TestUnit_SubmitPullReviewOptions_Validate(t *testing.T) {
	require.EqualError(t, SubmitPullReviewOptions{}.Validate(), "body is empty")
	require.NoError(t, SubmitPullReviewOptions{State: ReviewStateApproved}.Validate())
	require.NoError(t, SubmitPullReviewOptions{Body: "lgtm"}.Validate())
}

func TestUnit_CreatePullReviewComment_Validate(t *testing.T) {
	require.EqualError(t, CreatePullReviewComment{}.Validate(), "body is empty")
	require.EqualError(t,
		CreatePullReviewComment{Body: "x", OldLineNum: 1, NewLineNum: 1}.Validate(),
		"old and new line num are set, cant identify the code comment position")
	require.NoError(t, CreatePullReviewComment{Body: "x", NewLineNum: 1}.Validate())
	require.NoError(t, CreatePullReviewComment{Body: "x", OldLineNum: 1}.Validate())
}

// --- release.go ---

func TestUnit_ListReleasesOptions_QueryEncode(t *testing.T) {
	isDraft, isPre := true, false
	q, err := url.ParseQuery((&ListReleasesOptions{IsDraft: &isDraft, IsPreRelease: &isPre}).QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, "true", q.Get("draft"))
	assert.Equal(t, "false", q.Get("pre-release"))
	assert.NotContains(t, (&ListReleasesOptions{}).QueryEncode(), "draft=")
}

func TestUnit_CreateReleaseOption_Validate(t *testing.T) {
	require.EqualError(t, CreateReleaseOption{}.Validate(), "title is empty")
	require.NoError(t, CreateReleaseOption{Title: "v1"}.Validate())
}

// --- repo.go ---

func TestUnit_SearchRepoOptions_QueryEncode(t *testing.T) {
	isPrivate := true
	isArchived := false
	opt := &SearchRepoOptions{
		Keyword:              "sdk",
		KeywordIsTopic:       true,
		KeywordInDescription: true,
		OwnerID:              7,
		StarredByUserID:      9,
		IsPrivate:            &isPrivate,
		IsArchived:           &isArchived,
		ExcludeTemplate:      true,
		Type:                 "fork",
		Sort:                 "updated",
		PrioritizedByOwnerID: 1,
		Order:                "desc",
	}
	q, err := url.ParseQuery(opt.QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, "sdk", q.Get("q"))
	assert.Equal(t, "true", q.Get("topic"))
	assert.Equal(t, "true", q.Get("includeDesc"))
	assert.Equal(t, "7", q.Get("uid"))
	assert.Equal(t, "true", q.Get("exclusive"))
	assert.Equal(t, "9", q.Get("starredBy"))
	assert.NotEmpty(t, q.Get("is_private"))
	assert.NotEmpty(t, q.Get("archived"))
	assert.Equal(t, "false", q.Get("template"))
	assert.Equal(t, "fork", q.Get("mode"))
	assert.Equal(t, "updated", q.Get("sort"))
	assert.Equal(t, "1", q.Get("priority_owner_id"))
	assert.Equal(t, "desc", q.Get("order"))

	empty := (&SearchRepoOptions{}).QueryEncode()
	for _, absent := range []string{"q=", "topic=", "uid=", "starredBy=", "mode=", "sort=", "order="} {
		assert.NotContains(t, empty, absent)
	}
}

func TestUnit_CreateRepoOption_Validate(t *testing.T) {
	c := newUnitVersionedClient(t, "16.0.5")
	require.EqualError(t, CreateRepoOption{}.Validate(c), "name is empty")
	require.EqualError(t, CreateRepoOption{Name: strings.Repeat("a", 101)}.Validate(c), "name has more than 100 chars")
	require.EqualError(t, CreateRepoOption{Name: "n", Description: strings.Repeat("a", 2049)}.Validate(c), "description has more than 2048 chars")
	require.EqualError(t, CreateRepoOption{Name: "n", DefaultBranch: strings.Repeat("a", 101)}.Validate(c), "default branch name has more than 100 chars")
	require.NoError(t, CreateRepoOption{Name: "n"}.Validate(c))
	require.NoError(t, CreateRepoOption{Name: "n", TrustModel: "collaborator"}.Validate(c), "16.0.5 satisfies the >= 1.13.0 gate")

	old := newUnitVersionedClient(t, "1.12.0")
	require.Error(t, CreateRepoOption{Name: "n", TrustModel: "collaborator"}.Validate(old))
}

// --- repo_activity.go ---

func TestUnit_ListRepoActivityFeedsOptions_QueryEncode(t *testing.T) {
	assert.Contains(t, (&ListRepoActivityFeedsOptions{Date: "2024-01-01"}).QueryEncode(), "date=2024-01-01")
	assert.NotContains(t, (&ListRepoActivityFeedsOptions{}).QueryEncode(), "date=")
}

// --- repo_branch.go ---

func TestUnit_CreateBranchOption_Validate(t *testing.T) {
	require.EqualError(t, CreateBranchOption{}.Validate(), "BranchName is empty")
	require.EqualError(t, CreateBranchOption{BranchName: strings.Repeat("a", 101)}.Validate(), "BranchName to long")
	require.EqualError(t, CreateBranchOption{BranchName: "n", OldBranchName: strings.Repeat("a", 101)}.Validate(), "OldBranchName to long")
	require.NoError(t, CreateBranchOption{BranchName: "n"}.Validate())
}

// --- repo_collaborator.go ---

func TestUnit_AddCollaboratorOption_Validate(t *testing.T) {
	require.NoError(t, (&AddCollaboratorOption{}).Validate(), "nil Permission leaves the default")

	owner := AccessModeOwner
	optOwner := &AddCollaboratorOption{Permission: &owner}
	require.NoError(t, optOwner.Validate())
	assert.Equal(t, AccessModeAdmin, *optOwner.Permission, "owner is downgraded to admin")

	none := AccessModeNone
	optNone := &AddCollaboratorOption{Permission: &none}
	require.NoError(t, optNone.Validate())
	assert.Nil(t, optNone.Permission, "none clears the field so the server applies its own default")

	bogus := AccessMode("bogus")
	require.EqualError(t, (&AddCollaboratorOption{Permission: &bogus}).Validate(), "permission mode invalid")

	write := AccessModeWrite
	require.NoError(t, (&AddCollaboratorOption{Permission: &write}).Validate())
}

// --- repo_commit.go ---

func TestUnit_ListCommitOptions_QueryEncode(t *testing.T) {
	q, err := url.ParseQuery((&ListCommitOptions{SHA: "main", Path: "README.md", Stat: true, Not: "old"}).QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, "main", q.Get("sha"))
	assert.Equal(t, "README.md", q.Get("path"))
	assert.Equal(t, "true", q.Get("stat"))
	assert.Equal(t, "true", q.Get("verification"))
	assert.Equal(t, "true", q.Get("files"))
	assert.Equal(t, "old", q.Get("not"))

	empty := (&ListCommitOptions{}).QueryEncode()
	assert.NotContains(t, empty, "sha=")
	assert.NotContains(t, empty, "not=")
}

// --- repo_key.go ---

func TestUnit_ListDeployKeysOptions_QueryEncode(t *testing.T) {
	q, err := url.ParseQuery((&ListDeployKeysOptions{KeyID: 5, Fingerprint: "SHA256:xyz"}).QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, "5", q.Get("key_id"))
	assert.Equal(t, "SHA256:xyz", q.Get("fingerprint"))
	empty := (&ListDeployKeysOptions{}).QueryEncode()
	assert.NotContains(t, empty, "key_id=")
	assert.NotContains(t, empty, "fingerprint=")
}

// --- repo_migrate.go ---

func TestUnit_MigrateRepoOption_Validate(t *testing.T) {
	c := newUnitVersionedClient(t, "16.0.5")
	require.EqualError(t, (&MigrateRepoOption{}).Validate(c), "CloneAddr required")
	require.EqualError(t, (&MigrateRepoOption{CloneAddr: "x"}).Validate(c), "RepoName required")
	require.EqualError(t, (&MigrateRepoOption{CloneAddr: "x", RepoName: strings.Repeat("a", 101)}).Validate(c), "RepoName to long")
	require.EqualError(t, (&MigrateRepoOption{CloneAddr: "x", RepoName: "n", Description: strings.Repeat("a", 2049)}).Validate(c), "description to long")

	require.EqualError(t,
		(&MigrateRepoOption{CloneAddr: "x", RepoName: "n", Service: GitServiceGithub}).Validate(c),
		"github requires token authentication")
	require.NoError(t,
		(&MigrateRepoOption{CloneAddr: "x", RepoName: "n", Service: GitServiceGithub, AuthToken: "t"}).Validate(c))

	require.EqualError(t,
		(&MigrateRepoOption{CloneAddr: "x", RepoName: "n", Service: GitServiceGitlab}).Validate(c),
		"gitlab requires token authentication")
	require.NoError(t,
		(&MigrateRepoOption{CloneAddr: "x", RepoName: "n", Service: GitServiceGitlab, AuthToken: "t"}).Validate(c),
		"16.0.5 satisfies the >= 1.13.0 gate for gitlab/gitea/forgejo")

	old := newUnitVersionedClient(t, "1.12.0")
	require.EqualError(t,
		(&MigrateRepoOption{CloneAddr: "x", RepoName: "n", Service: GitServiceForgejo, AuthToken: "t"}).Validate(old),
		"migrate from service forgejo need forgejo >= 1.13.0")

	require.EqualError(t,
		(&MigrateRepoOption{CloneAddr: "x", RepoName: "n", Service: GitServiceGogs}).Validate(c),
		"gogs requires token authentication")
	require.EqualError(t,
		(&MigrateRepoOption{CloneAddr: "x", RepoName: "n", Service: GitServiceGogs, AuthToken: "t"}).Validate(old),
		"migrate from service gogs need forgejo >= 1.14.0")
	require.NoError(t,
		(&MigrateRepoOption{CloneAddr: "x", RepoName: "n", Service: GitServiceGogs, AuthToken: "t"}).Validate(c))
}

// --- repo_tag.go ---

func TestUnit_CreateTagOption_Validate(t *testing.T) {
	require.EqualError(t, CreateTagOption{}.Validate(), "TagName is required")
	require.NoError(t, CreateTagOption{TagName: "v1"}.Validate())
}

// --- repo_template.go ---

func TestUnit_CreateRepoFromTemplateOption_Validate(t *testing.T) {
	require.EqualError(t, CreateRepoFromTemplateOption{}.Validate(), "field Owner is required")
	require.EqualError(t, CreateRepoFromTemplateOption{Owner: "o"}.Validate(), "field Name is required")
	require.NoError(t, CreateRepoFromTemplateOption{Owner: "o", Name: "n"}.Validate())
}

// --- repo_topics.go ---

func TestUnit_SearchTopicsOptions_QueryEncode(t *testing.T) {
	assert.Contains(t, (&SearchTopicsOptions{Query: "sdk"}).QueryEncode(), "q=sdk")
}

// --- status.go ---

func TestUnit_ListCommitStatusesOptions_QueryEncode(t *testing.T) {
	q, err := url.ParseQuery((&ListCommitStatusesOptions{Sort: "recentupdate", State: "success"}).QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, "recentupdate", q.Get("sort"))
	assert.Equal(t, "success", q.Get("state"))
	empty := (&ListCommitStatusesOptions{}).QueryEncode()
	assert.NotContains(t, empty, "sort=")
	assert.NotContains(t, empty, "state=")
}

// --- user_activity.go ---

func TestUnit_ListActivityFeedsOptions_QueryEncode(t *testing.T) {
	q, err := url.ParseQuery((&ListActivityFeedsOptions{OnlyPerformedBy: true, Date: "2024-01-01"}).QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, "true", q.Get("only-performed-by"))
	assert.Equal(t, "2024-01-01", q.Get("date"))
	empty := (&ListActivityFeedsOptions{}).QueryEncode()
	assert.NotContains(t, empty, "only-performed-by=")
	assert.NotContains(t, empty, "date=")
}

// --- user_search.go ---

func TestUnit_SearchUsersOption_QueryEncode(t *testing.T) {
	q, err := url.ParseQuery((&SearchUsersOption{ListOptions: ListOptions{Page: 2, PageSize: 30}, KeyWord: "alice"}).QueryEncode())
	require.NoError(t, err)
	assert.Equal(t, "2", q.Get("page"))
	assert.Equal(t, "30", q.Get("limit"))
	assert.Equal(t, "alice", q.Get("q"))

	empty := (&SearchUsersOption{}).QueryEncode()
	assert.NotContains(t, empty, "page=")
	assert.NotContains(t, empty, "limit=")
	assert.NotContains(t, empty, "q=")
}
