// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

// EditDeadlineOption options for creating or updating an issue deadline
type EditDeadlineOption struct {
	// Deadline is the new due date. Passing nil unsets the deadline.
	Deadline *time.Time `json:"due_date"`
}

// IssueDeadline holds an issue's due date
type IssueDeadline struct {
	Deadline *time.Time `json:"due_date"`
}

// EditIssueDeadline sets or removes the deadline of an issue. Passing a nil
// Deadline in opt removes it; only the date portion of a non-nil Deadline is
// taken into account, the time of day is ignored.
func (c *Client) EditIssueDeadline(owner, repo string, index int64, opt EditDeadlineOption) (*IssueDeadline, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}
	deadline := new(IssueDeadline)
	resp, err := c.getParsedResponse("POST",
		fmt.Sprintf("/repos/%s/%s/issues/%d/deadline", owner, repo, index),
		jsonHeader, bytes.NewReader(body), deadline)
	return deadline, resp, err
}
