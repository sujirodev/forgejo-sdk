// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

// Unit tests for the debug-logging branches of doRequest,
// doRequestWithContext and getWebResponse (client.go), and for the error
// paths those three functions take before a request ever reaches the
// network. None of the existing integration tests turn debug mode on, so
// the `if debug { ... }` branches went unexercised even though every
// request in the suite runs through this code.

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_DoRequest_DebugMode(t *testing.T) {
	t.Parallel()
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	c := newUnitTestClient(t, srv, SetDebugMode())

	_, err := c.doRequest("POST", "/x", nil, bytes.NewReader([]byte(`{"a":1}`)))
	require.NoError(t, err, "debug logging must not change the outcome of a request with a body")
}

func TestUnit_DoRequestWithContext_DebugMode(t *testing.T) {
	t.Parallel()
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	c := newUnitTestClient(t, srv, SetDebugMode())

	resp, err := c.doRequestWithContext(t.Context(), "POST", "/x", nil, bytes.NewReader([]byte(`{"a":1}`)))
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUnit_GetWebResponse_DebugMode(t *testing.T) {
	t.Parallel()
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("diff body"))
	})
	c := newUnitTestClient(t, srv, SetDebugMode())

	data, resp, err := c.getWebResponse("GET", "/x.diff", nil)
	require.NoError(t, err)
	assert.Equal(t, "diff body", string(data))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUnit_DoRequest_BadMethod(t *testing.T) {
	t.Parallel()
	c, err := NewClient("http://example.invalid", SetForgejoVersion("16.0.5"))
	require.NoError(t, err)

	// A control character in the method makes http.NewRequestWithContext
	// itself fail, before any network I/O.
	_, err = c.doRequest("BAD\nMETHOD", "/x", nil, nil)
	require.Error(t, err)
}

func TestUnit_DoRequestWithContext_BadMethod(t *testing.T) {
	t.Parallel()
	c, err := NewClient("http://example.invalid", SetForgejoVersion("16.0.5"))
	require.NoError(t, err)

	_, err = c.doRequestWithContext(t.Context(), "BAD\nMETHOD", "/x", nil, nil)
	require.Error(t, err)
}

func TestUnit_GetWebResponse_BadMethod(t *testing.T) {
	t.Parallel()
	c, err := NewClient("http://example.invalid", SetForgejoVersion("16.0.5"))
	require.NoError(t, err)

	_, _, err = c.getWebResponse("BAD\nMETHOD", "/x", nil)
	require.Error(t, err)
}

func TestUnit_DoRequest_UnreachableHost(t *testing.T) {
	t.Parallel()
	// A well-formed request that http.Client.Do can never complete: no
	// listener on this port.
	c, err := NewClient("http://127.0.0.1:1", SetForgejoVersion("16.0.5"))
	require.NoError(t, err)

	_, err = c.doRequest("GET", "/x", nil, nil)
	require.Error(t, err)
}

func TestUnit_DoRequestWithContext_UnreachableHost(t *testing.T) {
	t.Parallel()
	c, err := NewClient("http://127.0.0.1:1", SetForgejoVersion("16.0.5"))
	require.NoError(t, err)

	_, err = c.doRequestWithContext(t.Context(), "GET", "/x", nil, nil)
	require.Error(t, err)
}

func TestUnit_GetWebResponse_UnreachableHost(t *testing.T) {
	t.Parallel()
	c, err := NewClient("http://127.0.0.1:1", SetForgejoVersion("16.0.5"))
	require.NoError(t, err)

	_, _, err = c.getWebResponse("GET", "/x", nil)
	require.Error(t, err)
}
