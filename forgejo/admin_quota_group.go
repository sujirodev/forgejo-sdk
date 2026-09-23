// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"encoding/json"
	"fmt"

	"codeberg.org/sujirodev/forgejo-sdk/forgejo/v3/models"
)

// AdminListQuotaGroups lists the available quota groups.
func (c *Client) AdminListQuotaGroups() ([]*models.QuotaGroup, *Response, error) {
	groups := make([]*models.QuotaGroup, 0)
	resp, err := c.getParsedResponse("GET", "/admin/quota/groups", jsonHeader, nil, &groups)
	return groups, resp, err
}

// CreateQuotaGroupOption options for creating a quota group
type CreateQuotaGroupOption struct {
	Name string `json:"name"`
	// Rules to add to the newly created group. If a rule with a given name
	// does not exist yet, it will be created. See CreateQuotaRuleOption.
	Rules []CreateQuotaRuleOption `json:"rules,omitempty"`
}

// AdminCreateQuotaGroup creates a new quota group.
func (c *Client) AdminCreateQuotaGroup(opt CreateQuotaGroupOption) (*models.QuotaGroup, *Response, error) {
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}
	group := new(models.QuotaGroup)
	resp, err := c.getParsedResponse("POST", "/admin/quota/groups", jsonHeader, bytes.NewReader(body), group)
	return group, resp, err
}

// AdminGetQuotaGroup gets information about the given quota group.
func (c *Client) AdminGetQuotaGroup(quotaGroup string) (*models.QuotaGroup, *Response, error) {
	if err := escapeValidatePathSegments(&quotaGroup); err != nil {
		return nil, nil, err
	}
	group := new(models.QuotaGroup)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/admin/quota/groups/%s", quotaGroup), jsonHeader, nil, group)
	return group, resp, err
}

// AdminDeleteQuotaGroup deletes the given quota group.
func (c *Client) AdminDeleteQuotaGroup(quotaGroup string) (*Response, error) {
	if err := escapeValidatePathSegments(&quotaGroup); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/admin/quota/groups/%s", quotaGroup), nil, nil)
	return resp, err
}

// AdminAddRuleToQuotaGroup adds an existing quota rule to a quota group.
func (c *Client) AdminAddRuleToQuotaGroup(quotaGroup, quotaRule string) (*Response, error) {
	if err := escapeValidatePathSegments(&quotaGroup, &quotaRule); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("PUT", fmt.Sprintf("/admin/quota/groups/%s/rules/%s", quotaGroup, quotaRule), nil, nil)
	return resp, err
}

// AdminRemoveRuleFromQuotaGroup removes a quota rule from a quota group.
func (c *Client) AdminRemoveRuleFromQuotaGroup(quotaGroup, quotaRule string) (*Response, error) {
	if err := escapeValidatePathSegments(&quotaGroup, &quotaRule); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/admin/quota/groups/%s/rules/%s", quotaGroup, quotaRule), nil, nil)
	return resp, err
}

// AdminListUsersInQuotaGroup lists the users in a quota group.
func (c *Client) AdminListUsersInQuotaGroup(quotaGroup string) ([]*User, *Response, error) {
	if err := escapeValidatePathSegments(&quotaGroup); err != nil {
		return nil, nil, err
	}
	users := make([]*User, 0)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/admin/quota/groups/%s/users", quotaGroup), jsonHeader, nil, &users)
	return users, resp, err
}

// AdminAddUserToQuotaGroup adds a user to a quota group.
func (c *Client) AdminAddUserToQuotaGroup(quotaGroup, username string) (*Response, error) {
	if err := escapeValidatePathSegments(&quotaGroup, &username); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("PUT", fmt.Sprintf("/admin/quota/groups/%s/users/%s", quotaGroup, username), nil, nil)
	return resp, err
}

// AdminRemoveUserFromQuotaGroup removes a user from a quota group.
func (c *Client) AdminRemoveUserFromQuotaGroup(quotaGroup, username string) (*Response, error) {
	if err := escapeValidatePathSegments(&quotaGroup, &username); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/admin/quota/groups/%s/users/%s", quotaGroup, username), nil, nil)
	return resp, err
}
