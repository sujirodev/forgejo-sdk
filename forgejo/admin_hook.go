// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// AdminListHooks lists all the global (system) webhooks.
func (c *Client) AdminListHooks(opt ListHooksOptions) ([]*Hook, *Response, error) {
	opt.setDefaults()
	hooks := make([]*Hook, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/admin/hooks?%s", opt.getURLQuery().Encode()), jsonHeader, nil, &hooks)
	return hooks, resp, err
}

// AdminCreateHook creates a global (system) webhook.
func (c *Client) AdminCreateHook(opt CreateHookOption) (*Hook, *Response, error) {
	if err := opt.Validate(); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}
	h := new(Hook)
	resp, err := c.getParsedResponse("POST", "/admin/hooks", jsonHeader, bytes.NewReader(body), h)
	return h, resp, err
}

// AdminGetHook gets a global (system) webhook by its ID.
func (c *Client) AdminGetHook(id int64) (*Hook, *Response, error) {
	h := new(Hook)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/admin/hooks/%d", id), jsonHeader, nil, h)
	return h, resp, err
}

// AdminDeleteHook deletes a global (system) webhook by its ID.
func (c *Client) AdminDeleteHook(id int64) (*Response, error) {
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/admin/hooks/%d", id), nil, nil)
	return resp, err
}

// AdminEditHook edits a global (system) webhook by its ID.
func (c *Client) AdminEditHook(id int64, opt EditHookOption) (*Hook, *Response, error) {
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}
	h := new(Hook)
	resp, err := c.getParsedResponse("PATCH", fmt.Sprintf("/admin/hooks/%d", id), jsonHeader, bytes.NewReader(body), h)
	return h, resp, err
}
