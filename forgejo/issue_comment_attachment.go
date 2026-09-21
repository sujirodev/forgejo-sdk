// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"io"
)

// ListIssueCommentAttachments list a comment's attachments
func (c *Client) ListIssueCommentAttachments(owner, repo string, commentID int64) ([]*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	return c.listAttachments(fmt.Sprintf("/repos/%s/%s/issues/comments/%d/assets", owner, repo, commentID))
}

// GetIssueCommentAttachment returns the requested comment attachment
func (c *Client) GetIssueCommentAttachment(owner, repo string, commentID, attachmentID int64) (*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	return c.getAttachment(fmt.Sprintf("/repos/%s/%s/issues/comments/%d/assets/%d", owner, repo, commentID, attachmentID))
}

// CreateIssueCommentAttachment creates an attachment for the given comment with the
// given file and filename. The filename is also used as the attachment's name unless
// overridden with EditIssueCommentAttachment afterwards.
func (c *Client) CreateIssueCommentAttachment(owner, repo string, commentID int64, file io.Reader, filename string) (*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	return c.uploadAttachment(fmt.Sprintf("/repos/%s/%s/issues/comments/%d/assets", owner, repo, commentID), filename, file)
}

// EditIssueCommentAttachment updates the given comment attachment with the given options
func (c *Client) EditIssueCommentAttachment(owner, repo string, commentID, attachmentID int64, form EditAttachmentOptions) (*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	return c.editAttachment(fmt.Sprintf("/repos/%s/%s/issues/comments/%d/assets/%d", owner, repo, commentID, attachmentID), form)
}

// DeleteIssueCommentAttachment deletes the given comment attachment including the uploaded file
func (c *Client) DeleteIssueCommentAttachment(owner, repo string, commentID, attachmentID int64) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}
	return c.deleteAttachment(fmt.Sprintf("/repos/%s/%s/issues/comments/%d/assets/%d", owner, repo, commentID, attachmentID))
}
