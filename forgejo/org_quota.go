// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"codeberg.org/sujirodev/forgejo-sdk/forgejo/v3/models"
)

// GetOrgQuota returns quota information for an organization
func (c *Client) GetOrgQuota(ctx context.Context, org string) (models.QuotaInfo, Response, error) {
	if err := escapeValidatePathSegments(&org); err != nil {
		return models.QuotaInfo{}, Response{}, err
	}

	var quota models.QuotaInfo
	resp, err := c.getParsedResponseWithContext(ctx, fmt.Sprintf("/orgs/%s/quota", org), &quota)
	if err != nil {
		return models.QuotaInfo{}, resp, err
	}

	return quota, resp, nil
}

// ListOrgQuotaArtifactsOptions holds optional parameters for listing an organization's quota artifacts
type ListOrgQuotaArtifactsOptions struct {
	ListOptions
}

// ListOrgQuotaArtifacts lists artifacts counting towards an organization's quota
func (c *Client) ListOrgQuotaArtifacts(ctx context.Context, org string, opt ListOrgQuotaArtifactsOptions) ([]models.QuotaUsedArtifact, Response, error) {
	if err := escapeValidatePathSegments(&org); err != nil {
		return nil, Response{}, err
	}

	link := opt.getURLQuery().Encode()
	var artifacts []models.QuotaUsedArtifact

	path := fmt.Sprintf("/orgs/%s/quota/artifacts", org)
	if link != "" {
		path += "?" + link
	}

	resp, err := c.getParsedResponseWithContext(ctx, path, &artifacts)
	if err != nil {
		return nil, resp, err
	}

	return artifacts, resp, nil
}

// ListOrgQuotaAttachmentsOptions holds optional parameters for listing an organization's quota attachments
type ListOrgQuotaAttachmentsOptions struct {
	ListOptions
}

// ListOrgQuotaAttachments lists attachments counting towards an organization's quota
func (c *Client) ListOrgQuotaAttachments(ctx context.Context, org string, opt ListOrgQuotaAttachmentsOptions) ([]models.QuotaUsedAttachment, Response, error) {
	if err := escapeValidatePathSegments(&org); err != nil {
		return nil, Response{}, err
	}

	link := opt.getURLQuery().Encode()
	var attachments []models.QuotaUsedAttachment

	path := fmt.Sprintf("/orgs/%s/quota/attachments", org)
	if link != "" {
		path += "?" + link
	}

	resp, err := c.getParsedResponseWithContext(ctx, path, &attachments)
	if err != nil {
		return nil, resp, err
	}

	return attachments, resp, nil
}

// CheckOrgQuota checks if the organization is over quota for the given subject.
func (c *Client) CheckOrgQuota(ctx context.Context, org string, subject QuotaSubject) (bool, Response, error) {
	if err := escapeValidatePathSegments(&org); err != nil {
		return false, Response{}, err
	}

	s := strings.TrimSpace(string(subject))
	if s == "" {
		return false, Response{}, fmt.Errorf("quota subject is required")
	}

	data, resp, err := c.getResponseWithContext(ctx, "GET", fmt.Sprintf("/orgs/%s/quota/check?subject=%s", org, url.QueryEscape(s)), nil, nil)
	if err != nil {
		return false, resp, err
	}

	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return false, resp, fmt.Errorf("unexpected /orgs/%s/quota/check response: %s", org, string(data))
	}

	return b, resp, nil
}

// ListOrgQuotaPackagesOptions holds optional parameters for listing an organization's quota packages
type ListOrgQuotaPackagesOptions struct {
	ListOptions
}

// ListOrgQuotaPackages lists packages counting towards an organization's quota
func (c *Client) ListOrgQuotaPackages(ctx context.Context, org string, opt ListOrgQuotaPackagesOptions) ([]models.QuotaUsedPackage, Response, error) {
	if err := escapeValidatePathSegments(&org); err != nil {
		return nil, Response{}, err
	}

	link := opt.getURLQuery().Encode()
	var packages []models.QuotaUsedPackage

	path := fmt.Sprintf("/orgs/%s/quota/packages", org)
	if link != "" {
		path += "?" + link
	}

	resp, err := c.getParsedResponseWithContext(ctx, path, &packages)
	if err != nil {
		return nil, resp, err
	}

	return packages, resp, nil
}
