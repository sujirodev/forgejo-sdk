// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"io"
)

// ListIssueAttachments list an issue's attachments
func (c *Client) ListIssueAttachments(owner, repo string, index int64) ([]*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	return c.listAttachments(fmt.Sprintf("/repos/%s/%s/issues/%d/assets", owner, repo, index))
}

// GetIssueAttachment returns the requested issue attachment
func (c *Client) GetIssueAttachment(owner, repo string, index, attachmentID int64) (*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	return c.getAttachment(fmt.Sprintf("/repos/%s/%s/issues/%d/assets/%d", owner, repo, index, attachmentID))
}

// CreateIssueAttachment creates an attachment for the given issue with the given file
// and filename. The filename is also used as the attachment's name unless overridden
// with EditIssueAttachment afterwards.
func (c *Client) CreateIssueAttachment(owner, repo string, index int64, file io.Reader, filename string) (*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	return c.uploadAttachment(fmt.Sprintf("/repos/%s/%s/issues/%d/assets", owner, repo, index), filename, file)
}

// EditIssueAttachment updates the given issue attachment with the given options
func (c *Client) EditIssueAttachment(owner, repo string, index, attachmentID int64, form EditAttachmentOptions) (*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, nil, err
	}
	return c.editAttachment(fmt.Sprintf("/repos/%s/%s/issues/%d/assets/%d", owner, repo, index, attachmentID), form)
}

// DeleteIssueAttachment deletes the given issue attachment including the uploaded file
func (c *Client) DeleteIssueAttachment(owner, repo string, index, attachmentID int64) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo); err != nil {
		return nil, err
	}
	return c.deleteAttachment(fmt.Sprintf("/repos/%s/%s/issues/%d/assets/%d", owner, repo, index, attachmentID))
}
