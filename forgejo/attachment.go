// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo // import "codeberg.org/sujirodev/forgejo-sdk/forgejo/v3"
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// Attachment a generic attachment
type Attachment struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Size          int64     `json:"size"`
	DownloadCount int64     `json:"download_count"`
	Created       time.Time `json:"created_at"`
	UUID          string    `json:"uuid"`
	DownloadURL   string    `json:"browser_download_url"`
}

// ListReleaseAttachments list release's attachments
func (c *Client) ListReleaseAttachments(user, repo string, release int64) ([]*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	return c.listAttachments(fmt.Sprintf("/repos/%s/%s/releases/%d/assets", user, repo, release))
}

// GetReleaseAttachment returns the requested attachment
func (c *Client) GetReleaseAttachment(user, repo string, release, id int64) (*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	return c.getAttachment(fmt.Sprintf("/repos/%s/%s/releases/%d/assets/%d", user, repo, release, id))
}

// CreateReleaseAttachment creates an attachment for the given release
func (c *Client) CreateReleaseAttachment(user, repo string, release int64, file io.Reader, filename string) (*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	return c.uploadAttachment(fmt.Sprintf("/repos/%s/%s/releases/%d/assets", user, repo, release), filename, file)
}

// EditAttachmentOptions options for editing attachments
type EditAttachmentOptions struct {
	Name string `json:"name"`
}

// EditReleaseAttachment updates the given attachment with the given options
func (c *Client) EditReleaseAttachment(user, repo string, release, attachment int64, form EditAttachmentOptions) (*Attachment, *Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, nil, err
	}
	return c.editAttachment(fmt.Sprintf("/repos/%s/%s/releases/%d/assets/%d", user, repo, release, attachment), form)
}

// DeleteReleaseAttachment deletes the given attachment including the uploaded file
func (c *Client) DeleteReleaseAttachment(user, repo string, release, id int64) (*Response, error) {
	if err := escapeValidatePathSegments(&user, &repo); err != nil {
		return nil, err
	}
	return c.deleteAttachment(fmt.Sprintf("/repos/%s/%s/releases/%d/assets/%d", user, repo, release, id))
}

// listAttachments lists the attachments at apiPath (an issue's, comment's or
// release's /assets endpoint)
func (c *Client) listAttachments(apiPath string) ([]*Attachment, *Response, error) {
	attachments := make([]*Attachment, 0)
	resp, err := c.getParsedResponse("GET", apiPath, nil, nil, &attachments)
	return attachments, resp, err
}

// getAttachment fetches the single attachment at apiPath (an issue's,
// comment's or release's /assets/{id} endpoint)
func (c *Client) getAttachment(apiPath string) (*Attachment, *Response, error) {
	a := new(Attachment)
	resp, err := c.getParsedResponse("GET", apiPath, nil, nil, &a)
	return a, resp, err
}

// uploadAttachment uploads file as a multipart/form-data attachment to apiPath
// (an issue's, comment's or release's /assets endpoint) and decodes the
// created Attachment from the response
func (c *Client) uploadAttachment(apiPath, filename string, file io.Reader) (*Attachment, *Response, error) {
	// Write file to body
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("attachment", filename)
	if err != nil {
		return nil, nil, err
	}

	if _, err = io.Copy(part, file); err != nil {
		return nil, nil, err
	}
	if err = writer.Close(); err != nil {
		return nil, nil, err
	}

	// Send request
	attachment := new(Attachment)
	resp, err := c.getParsedResponse("POST", apiPath,
		http.Header{"Content-Type": []string{writer.FormDataContentType()}}, body, &attachment)
	return attachment, resp, err
}

// editAttachment updates the attachment at apiPath (an issue's, comment's or
// release's /assets/{id} endpoint) with the given options
func (c *Client) editAttachment(apiPath string, form EditAttachmentOptions) (*Attachment, *Response, error) {
	body, err := json.Marshal(&form)
	if err != nil {
		return nil, nil, err
	}
	attach := new(Attachment)
	resp, err := c.getParsedResponse("PATCH", apiPath, jsonHeader, bytes.NewReader(body), attach)
	return attach, resp, err
}

// deleteAttachment deletes the attachment at apiPath (an issue's, comment's
// or release's /assets/{id} endpoint) including the uploaded file
func (c *Client) deleteAttachment(apiPath string) (*Response, error) {
	_, resp, err := c.getResponse("DELETE", apiPath, nil, nil)
	return resp, err
}
