// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2015 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"bytes"
	"encoding/json"
	"fmt"

	"codeberg.org/MatheusAlves96/forgejo-sdk/forgejo/v3/models"
)

// AdminListUsersOptions options for listing admin users
type AdminListUsersOptions struct {
	ListOptions
}

// AdminListUsers lists all users
func (c *Client) AdminListUsers(opt AdminListUsersOptions) ([]*User, *Response, error) {
	opt.setDefaults()
	users := make([]*User, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/admin/users?%s", opt.getURLQuery().Encode()), nil, nil, &users)
	return users, resp, err
}

// CreateUserOption create user options
type CreateUserOption struct {
	SourceID           int64        `json:"source_id"`
	LoginName          string       `json:"login_name"`
	Username           string       `json:"username"`
	FullName           string       `json:"full_name"`
	Email              string       `json:"email"`
	Password           string       `json:"password"`
	MustChangePassword *bool        `json:"must_change_password"`
	SendNotify         bool         `json:"send_notify"`
	Visibility         *VisibleType `json:"visibility"`
}

// Validate the CreateUserOption struct
func (opt CreateUserOption) Validate() error {
	if len(opt.Email) == 0 {
		return fmt.Errorf("email is empty")
	}
	if len(opt.Username) == 0 {
		return fmt.Errorf("username is empty")
	}
	return nil
}

// AdminCreateUser create a user
func (c *Client) AdminCreateUser(opt CreateUserOption) (*User, *Response, error) {
	if err := opt.Validate(); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}
	user := new(User)
	resp, err := c.getParsedResponse("POST", "/admin/users", jsonHeader, bytes.NewReader(body), user)
	return user, resp, err
}

// EditUserOption edit user options
type EditUserOption struct {
	SourceID                int64        `json:"source_id"`
	LoginName               string       `json:"login_name"`
	Email                   *string      `json:"email"`
	FullName                *string      `json:"full_name"`
	Password                string       `json:"password"`
	Description             *string      `json:"description"`
	MustChangePassword      *bool        `json:"must_change_password"`
	Website                 *string      `json:"website"`
	Location                *string      `json:"location"`
	Active                  *bool        `json:"active"`
	Admin                   *bool        `json:"admin"`
	AllowGitHook            *bool        `json:"allow_git_hook"`
	AllowImportLocal        *bool        `json:"allow_import_local"`
	MaxRepoCreation         *int         `json:"max_repo_creation"`
	ProhibitLogin           *bool        `json:"prohibit_login"`
	AllowCreateOrganization *bool        `json:"allow_create_organization"`
	Restricted              *bool        `json:"restricted"`
	Visibility              *VisibleType `json:"visibility"`
}

// AdminEditUser modify user informations
func (c *Client) AdminEditUser(user string, opt EditUserOption) (*Response, error) {
	if err := escapeValidatePathSegments(&user); err != nil {
		return nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("PATCH", fmt.Sprintf("/admin/users/%s", user), jsonHeader, bytes.NewReader(body))
	return resp, err
}

// AdminDeleteUser delete one user according name
func (c *Client) AdminDeleteUser(user string) (*Response, error) {
	if err := escapeValidatePathSegments(&user); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/admin/users/%s", user), nil, nil)
	return resp, err
}

// AdminCreateUserPublicKey adds a public key for the user
func (c *Client) AdminCreateUserPublicKey(user string, opt CreateKeyOption) (*PublicKey, *Response, error) {
	if err := escapeValidatePathSegments(&user); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}
	key := new(PublicKey)
	resp, err := c.getParsedResponse("POST", fmt.Sprintf("/admin/users/%s/keys", user), jsonHeader, bytes.NewReader(body), key)
	return key, resp, err
}

// AdminDeleteUserPublicKey deletes a user's public key
func (c *Client) AdminDeleteUserPublicKey(user string, keyID int) (*Response, error) {
	if err := escapeValidatePathSegments(&user); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/admin/users/%s/keys/%d", user, keyID), nil, nil)
	return resp, err
}

// AdminListUserEmails lists all email addresses for a user.
func (c *Client) AdminListUserEmails(user string) ([]*Email, *Response, error) {
	if err := escapeValidatePathSegments(&user); err != nil {
		return nil, nil, err
	}
	emails := make([]*Email, 0)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/admin/users/%s/emails", user), jsonHeader, nil, &emails)
	return emails, resp, err
}

// AdminDeleteUserEmails deletes email addresses from a user's account.
func (c *Client) AdminDeleteUserEmails(user string, opt DeleteEmailOption) (*Response, error) {
	if err := escapeValidatePathSegments(&user); err != nil {
		return nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/admin/users/%s/emails", user), jsonHeader, bytes.NewReader(body))
	return resp, err
}

// AdminGetUserQuota returns the given user's quota info.
func (c *Client) AdminGetUserQuota(user string) (models.QuotaInfo, *Response, error) {
	if err := escapeValidatePathSegments(&user); err != nil {
		return models.QuotaInfo{}, nil, err
	}
	var quota models.QuotaInfo
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/admin/users/%s/quota", user), jsonHeader, nil, &quota)
	return quota, resp, err
}

// SetUserQuotaGroupsOption options for setting a user's quota groups
type SetUserQuotaGroupsOption struct {
	// Groups is the complete list of quota groups the user shall be a member of.
	Groups []string `json:"groups"`
}

// AdminSetUserQuotaGroups sets the user's quota groups to the given list,
// replacing any existing group membership.
func (c *Client) AdminSetUserQuotaGroups(user string, opt SetUserQuotaGroupsOption) (*Response, error) {
	if err := escapeValidatePathSegments(&user); err != nil {
		return nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("POST", fmt.Sprintf("/admin/users/%s/quota/groups", user), jsonHeader, bytes.NewReader(body))
	return resp, err
}

// RenameUserOption options for renaming a user
type RenameUserOption struct {
	// NewName is the new username for this user. It cannot already be in use
	// by any other user.
	NewName string `json:"new_username"`
}

// AdminRenameUser renames a user.
func (c *Client) AdminRenameUser(user string, opt RenameUserOption) (*Response, error) {
	if err := escapeValidatePathSegments(&user); err != nil {
		return nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("POST", fmt.Sprintf("/admin/users/%s/rename", user), jsonHeader, bytes.NewReader(body))
	return resp, err
}

// AdminListUserAccessTokensOptions options for listing a user's access tokens as an admin
type AdminListUserAccessTokensOptions struct {
	ListOptions
}

// AdminListUserAccessTokens lists the specified user's access tokens.
func (c *Client) AdminListUserAccessTokens(user string, opt AdminListUserAccessTokensOptions) ([]*AccessToken, *Response, error) {
	if err := escapeValidatePathSegments(&user); err != nil {
		return nil, nil, err
	}
	opt.setDefaults()
	tokens := make([]*AccessToken, 0, opt.PageSize)
	resp, err := c.getParsedResponse("GET", fmt.Sprintf("/admin/users/%s/tokens?%s", user, opt.getURLQuery().Encode()), jsonHeader, nil, &tokens)
	return tokens, resp, err
}

// AdminCreateUserAccessToken creates an access token for the specified user,
// without needing that user's password.
func (c *Client) AdminCreateUserAccessToken(user string, opt CreateAccessTokenOption) (*AccessToken, *Response, error) {
	if err := escapeValidatePathSegments(&user); err != nil {
		return nil, nil, err
	}
	body, err := json.Marshal(&opt)
	if err != nil {
		return nil, nil, err
	}
	t := new(AccessToken)
	resp, err := c.getParsedResponse("POST", fmt.Sprintf("/admin/users/%s/tokens", user), jsonHeader, bytes.NewReader(body), t)
	return t, resp, err
}

// AdminDeleteUserAccessToken deletes an access token for the specified user,
// identified by ID and if not available by name.
func (c *Client) AdminDeleteUserAccessToken(user, token string) (*Response, error) {
	if err := escapeValidatePathSegments(&user, &token); err != nil {
		return nil, err
	}
	_, resp, err := c.getResponse("DELETE", fmt.Sprintf("/admin/users/%s/tokens/%s", user, token), nil, nil)
	return resp, err
}
