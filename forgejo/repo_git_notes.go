// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Note contains information related to a git note
type Note struct {
	Message string  `json:"message"`
	Commit  *Commit `json:"commit"`
}

// NoteOptions options when setting a git note
type NoteOptions struct {
	Message string `json:"message"`
}

// GetNote gets a note corresponding to a single commit from a repository
func (c *Client) GetNote(owner, repo, sha string) (*Note, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo, &sha); err != nil {
		return nil, nil, err
	}
	note := new(Note)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/repos/%s/%s/git/notes/%s", owner, repo, sha), jsonHeader, nil, note)
	return note, resp, err
}

// SetNote sets a note corresponding to a single commit from a repository
func (c *Client) SetNote(owner, repo, sha string, opt NoteOptions) (*Note, *Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo, &sha); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}
	note := new(Note)
	resp, err := c.getParsedResponse("POST", fmt.Sprintf("/repos/%s/%s/git/notes/%s", owner, repo, sha), jsonHeader, bytes.NewReader(body), note)
	return note, resp, err
}

// RemoveNote removes a note corresponding to a single commit from a repository
func (c *Client) RemoveNote(owner, repo, sha string) (*Response, error) {
	if err := escapeValidatePathSegments(&owner, &repo, &sha); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/repos/%s/%s/git/notes/%s", owner, repo, sha), nil, nil)
	return resp, err
}
