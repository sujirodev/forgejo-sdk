// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/sujirodev/forgejo-sdk/forgejo/v3/models"
)

// AdminListQuotaRules lists the available quota rules.
func (c *Client) AdminListQuotaRules() ([]*models.QuotaRuleInfo, *Response, error) {
	rules := make([]*models.QuotaRuleInfo, 0)
	resp, err := c.getParsedResponse("GET", "/admin/quota/rules", jsonHeader, nil, &rules)
	return rules, resp, err
}

// CreateQuotaRuleOption options for creating a quota rule
type CreateQuotaRuleOption struct {
	Name     string   `json:"name"`
	Limit    int64    `json:"limit"`
	Subjects []string `json:"subjects,omitempty"`
}

// AdminCreateQuotaRule creates a new quota rule.
func (c *Client) AdminCreateQuotaRule(opt CreateQuotaRuleOption) (*models.QuotaRuleInfo, *Response, error) {
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}
	rule := new(models.QuotaRuleInfo)
	resp, err := c.getParsedResponse("POST", "/admin/quota/rules", jsonHeader, bytes.NewReader(body), rule)
	return rule, resp, err
}

// AdminGetQuotaRule gets information about the given quota rule.
func (c *Client) AdminGetQuotaRule(quotaRule string) (*models.QuotaRuleInfo, *Response, error) {
	if err := escapeValidatePathSegments(&quotaRule); err != nil {
		return nil, nil, err
	}
	rule := new(models.QuotaRuleInfo)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/admin/quota/rules/%s", quotaRule), jsonHeader, nil, rule)
	return rule, resp, err
}

// AdminDeleteQuotaRule deletes the given quota rule.
func (c *Client) AdminDeleteQuotaRule(quotaRule string) (*Response, error) {
	if err := escapeValidatePathSegments(&quotaRule); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/admin/quota/rules/%s", quotaRule), nil, nil)
	return resp, err
}

// EditQuotaRuleOption options for editing an existing quota rule
type EditQuotaRuleOption struct {
	Limit    *int64   `json:"limit,omitempty"`
	Subjects []string `json:"subjects,omitempty"`
}

// AdminEditQuotaRule changes an existing quota rule.
func (c *Client) AdminEditQuotaRule(quotaRule string, opt EditQuotaRuleOption) (*models.QuotaRuleInfo, *Response, error) {
	if err := escapeValidatePathSegments(&quotaRule); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}
	rule := new(models.QuotaRuleInfo)
	resp, err := c.getParsedResponse("PATCH", fmt.Sprintf("/admin/quota/rules/%s", quotaRule), jsonHeader, bytes.NewReader(body), rule)
	return rule, resp, err
}
