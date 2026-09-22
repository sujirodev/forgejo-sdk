// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import "context"

// GetSigningKey returns the server's default GPG signing key, armored.
func (c *Client) GetSigningKey(ctx context.Context) (string, Response, error) {
	data, resp, err := c.getResponseWithContext(ctx, "GET", "/signing-key.gpg", nil, nil)
	if err != nil {
		return "", resp, err
	}
	return string(data), resp, nil
}

// GetSSHSigningKey returns the server's default SSH signing key, in OpenSSH
// authorized-key format.
func (c *Client) GetSSHSigningKey(ctx context.Context) (string, Response, error) {
	data, resp, err := c.getResponseWithContext(ctx, "GET", "/signing-key.ssh", nil, nil)
	if err != nil {
		return "", resp, err
	}
	return string(data), resp, nil
}
