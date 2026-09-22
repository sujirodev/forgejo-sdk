// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import "context"

// NodeInfo is the standardized way of exposing metadata about a server
// running one of the distributed social networks, as described by
// https://nodeinfo.diaspora.software/.
type NodeInfo struct {
	Version           string           `json:"version"`
	Software          NodeInfoSoftware `json:"software"`
	Protocols         []string         `json:"protocols"`
	Services          NodeInfoServices `json:"services"`
	OpenRegistrations bool             `json:"openRegistrations"`
	Usage             NodeInfoUsage    `json:"usage"`
	Metadata          interface{}      `json:"metadata"`
}

// NodeInfoSoftware holds metadata about the server software in use.
type NodeInfoSoftware struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Repository string `json:"repository"`
	Homepage   string `json:"homepage"`
}

// NodeInfoServices lists the third-party sites this server can connect to,
// and be connected from, via their application API.
type NodeInfoServices struct {
	Inbound  []string `json:"inbound"`
	Outbound []string `json:"outbound"`
}

// NodeInfoUsage holds usage statistics for the server.
type NodeInfoUsage struct {
	Users         NodeInfoUsageUsers `json:"users"`
	LocalPosts    int64              `json:"localPosts"`
	LocalComments int64              `json:"localComments"`
}

// NodeInfoUsageUsers holds statistics about the users of the server.
type NodeInfoUsageUsers struct {
	Total          int64 `json:"total"`
	ActiveMonth    int64 `json:"activeMonth"`
	ActiveHalfyear int64 `json:"activeHalfyear"`
}

// GetNodeInfo returns the nodeinfo of the Forgejo application.
func (c *Client) GetNodeInfo(ctx context.Context) (NodeInfo, Response, error) {
	var info NodeInfo
	resp, err := c.getParsedResponseWithContext(ctx, "/nodeinfo", &info)
	return info, resp, err
}
