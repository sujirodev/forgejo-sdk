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

	"codeberg.org/MatheusAlves96/forgejo-sdk/forgejo/v3/models"
)

// QuotaSubject represents a quota limit subject for use with CheckMyQuota.
type QuotaSubject string

const (
	QuotaSubjectNone                          QuotaSubject = "none"
	QuotaSubjectSizeAll                       QuotaSubject = "size:all"
	QuotaSubjectSizeReposAll                  QuotaSubject = "size:repos:all"
	QuotaSubjectSizeReposPublic               QuotaSubject = "size:repos:public"
	QuotaSubjectSizeReposPrivate              QuotaSubject = "size:repos:private"
	QuotaSubjectSizeGitAll                    QuotaSubject = "size:git:all"
	QuotaSubjectSizeGitLFS                    QuotaSubject = "size:git:lfs"
	QuotaSubjectSizeAssetsAll                 QuotaSubject = "size:assets:all"
	QuotaSubjectSizeAssetsAttachmentsAll      QuotaSubject = "size:assets:attachments:all"
	QuotaSubjectSizeAssetsAttachmentsIssues   QuotaSubject = "size:assets:attachments:issues"
	QuotaSubjectSizeAssetsAttachmentsReleases QuotaSubject = "size:assets:attachments:releases"
	QuotaSubjectSizeAssetsArtifacts           QuotaSubject = "size:assets:artifacts"
	QuotaSubjectSizeAssetsPackagesAll         QuotaSubject = "size:assets:packages:all"
	QuotaSubjectSizeWiki                      QuotaSubject = "size:assets:wiki"
)

// GetMyQuota returns quota information for the authenticated user
func (c *Client) GetMyQuota(ctx context.Context) (models.QuotaInfo, Response, error) {
	var quota models.QuotaInfo

	resp, err := c.getParsedResponseWithContext(ctx, "/user/quota", jsonHeader, nil, &quota)
	if err != nil {
		return models.QuotaInfo{}, resp, err
	}

	return quota, resp, nil
}

// ListMyQuotaArtifactsOptions holds optional parameters for listing quota artifacts
type ListMyQuotaArtifactsOptions struct {
	ListOptions
}

// ListMyQuotaArtifacts lists artifacts counting towards the authenticated user's quota
func (c *Client) ListMyQuotaArtifacts(ctx context.Context, opt ListMyQuotaArtifactsOptions) ([]models.QuotaUsedArtifact, Response, error) {
	link := opt.getURLQuery().Encode()
	var artifacts []models.QuotaUsedArtifact

	path := "/user/quota/artifacts"
	if link != "" {
		path += "?" + link
	}

	resp, err := c.getParsedResponseWithContext(ctx, path, jsonHeader, nil, &artifacts)
	if err != nil {
		return nil, resp, err
	}

	return artifacts, resp, nil
}

// ListMyQuotaAttachmentsOptions holds optional parameters for listing quota attachments
type ListMyQuotaAttachmentsOptions struct {
	ListOptions
}

// ListMyQuotaAttachments lists attachments counting towards the authenticated user's quota
func (c *Client) ListMyQuotaAttachments(ctx context.Context, opt ListMyQuotaAttachmentsOptions) ([]models.QuotaUsedAttachment, Response, error) {
	link := opt.getURLQuery().Encode()
	var attachments []models.QuotaUsedAttachment

	path := "/user/quota/attachments"
	if link != "" {
		path += "?" + link
	}

	resp, err := c.getParsedResponseWithContext(ctx, path, jsonHeader, nil, &attachments)
	if err != nil {
		return nil, resp, err
	}

	return attachments, resp, nil
}

// CheckMyQuota checks if the authenticated user is over quota.
func (c *Client) CheckMyQuota(ctx context.Context, subject QuotaSubject) (bool, Response, error) {
	s := strings.TrimSpace(string(subject))
	if s == "" {
		return false, Response{}, fmt.Errorf("quota subject is required")
	}

	data, resp, err := c.getResponseWithContext(ctx, "GET", "/user/quota/check?subject="+url.QueryEscape(s), nil, nil)
	if err != nil {
		return false, resp, err
	}

	var b bool
	if err := json.Unmarshal(data, &b); err != nil {
		return false, resp, fmt.Errorf("unexpected /user/quota/check response: %s", string(data))
	}

	return b, resp, nil
}

// ListMyQuotaPackagesOptions holds optional parameters for listing quota packages
type ListMyQuotaPackagesOptions struct {
	ListOptions
}

// ListMyQuotaPackages lists packages counting towards the authenticated user's quota
func (c *Client) ListMyQuotaPackages(ctx context.Context, opt ListMyQuotaPackagesOptions) ([]models.QuotaUsedPackage, Response, error) {
	link := opt.getURLQuery().Encode()
	var packages []models.QuotaUsedPackage

	path := "/user/quota/packages"
	if link != "" {
		path += "?" + link
	}

	resp, err := c.getParsedResponseWithContext(ctx, path, jsonHeader, nil, &packages)
	if err != nil {
		return nil, resp, err
	}

	return packages, resp, nil
}
