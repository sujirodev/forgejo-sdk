// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

// Tests in this file exercise the Client's own plumbing (auth headers,
// error mapping, version gating, context cancellation) against an
// httptest.Server instead of a real Forgejo instance. They don't need
// FORGEJO_SDK_TEST_* and don't need `make test-instance`; see
// `make test-unit`.
//
// They're named TestUnit_* on purpose, so `make test-unit` can select them
// with `go test -run '^TestUnit_'` without needing a build tag on every
// other (server-dependent) test file in this package.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUnitTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// newUnitTestClient builds a Client against srv without ever hitting the
// network for version detection: SetForgejoVersion pre-fills serverVersion,
// consuming the client's sync.Once, so NewClient's own version check and any
// later checkServerVersionGreaterThanOrEqual call are answered from memory.
func newUnitTestClient(t *testing.T, srv *httptest.Server, opts ...ClientOption) *Client {
	t.Helper()
	c, err := NewClient(srv.URL, append([]ClientOption{SetForgejoVersion("16.0.5")}, opts...)...)
	require.NoError(t, err)
	return c
}

func TestUnit_DoRequest_TokenAuthHeader(t *testing.T) {
	t.Parallel()
	var gotAuth, gotUA string
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
	})

	c := newUnitTestClient(t, srv, SetToken("mytoken"), SetUserAgent("forgejo-sdk-test"))

	_, err := c.doRequest("GET", "/some/path", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "token mytoken", gotAuth)
	assert.Equal(t, "forgejo-sdk-test", gotUA)
}

func TestUnit_DoRequest_BasicAuthAndOTPHeader(t *testing.T) {
	t.Parallel()
	var gotUser, gotPass string
	var gotOK bool
	var gotOTP string
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, gotOK = r.BasicAuth()
		gotOTP = r.Header.Get("X-FORGEJO-OTP")
		w.WriteHeader(http.StatusOK)
	})

	c := newUnitTestClient(t, srv, SetBasicAuth("alice", "s3cret"), SetOTP("123456"))

	_, err := c.doRequest("GET", "/some/path", nil, nil)
	require.NoError(t, err)
	assert.True(t, gotOK)
	assert.Equal(t, "alice", gotUser)
	assert.Equal(t, "s3cret", gotPass)
	assert.Equal(t, "123456", gotOTP)
}

func TestUnit_DoRequest_SudoHeader(t *testing.T) {
	t.Parallel()
	var gotSudo string
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotSudo = r.Header.Get("Sudo")
		w.WriteHeader(http.StatusOK)
	})

	c := newUnitTestClient(t, srv, SetToken("t"), SetSudo("root-user"))

	_, err := c.doRequest("GET", "/some/path", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "root-user", gotSudo)
}

func TestUnit_StatusCodeToErr_SuccessNoError(t *testing.T) {
	t.Parallel()
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	c := newUnitTestClient(t, srv)

	data, _, err := c.getResponse("GET", "/ok", nil, nil)
	require.NoError(t, err)
	assert.JSONEq(t, `{"ok":true}`, string(data))
}

func TestUnit_StatusCodeToErr_JSONMessageBody(t *testing.T) {
	t.Parallel()
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "repository not found"})
	})

	c := newUnitTestClient(t, srv)

	_, _, err := c.getResponse("GET", "/missing", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "repository not found")
}

func TestUnit_StatusCodeToErr_EmptyJSONMessage(t *testing.T) {
	t.Parallel()
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "",
			"url":     "http://localhost:3000/api/swagger",
		})
	})

	c := newUnitTestClient(t, srv)

	// Forgejo strips the internal detail of a 5XX for callers who are not
	// instance admins and answers {"message":""}. Taking that message as the
	// error would give the caller a non-nil error that says nothing, so the
	// status fallback has to win.
	_, _, err := c.getResponse("GET", "/masked", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500 Internal Server Error")
	assert.Contains(t, err.Error(), "/api/swagger")
}

func TestUnit_StatusCodeToErr_NonJSONBody(t *testing.T) {
	t.Parallel()
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom, not json"))
	})

	c := newUnitTestClient(t, srv)

	_, _, err := c.getResponse("GET", "/broken", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown API Error: 500")
	assert.Contains(t, err.Error(), "/api/v1/broken")
}

func TestUnit_SetForgejoVersion_SkipsVersionCheck(t *testing.T) {
	t.Parallel()
	called := false
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusInternalServerError)
	})

	// SetForgejoVersion("") sets ignoreVersion, so NewClient must not call
	// GET /version at all, even though the server would fail that request.
	c, err := NewClient(srv.URL, SetForgejoVersion(""))
	require.NoError(t, err)
	require.NotNil(t, c)
	assert.False(t, called, "server should not have been contacted")
}

func TestUnit_NewClient_RejectsServerOlderThanMinimum(t *testing.T) {
	t.Parallel()
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.10.0"})
	})

	c, err := NewClient(srv.URL)
	require.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "older than 1.11.0")
}

func TestUnit_NewClient_UnknownVersionStillReturnsClient(t *testing.T) {
	t.Parallel()
	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"version": "not-a-real-version"})
	})

	c, err := NewClient(srv.URL)
	require.Error(t, err)
	require.NotNil(t, c, "an unrecognized version should still hand back a usable client")
	assert.ErrorIs(t, err, &ErrUnknownVersion{})
}

func TestUnit_DoRequestWithContext_RespectsCancellation(t *testing.T) {
	t.Parallel()
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })

	srv := newUnitTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	})

	c := newUnitTestClient(t, srv)

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	_, err := c.doRequestWithContext(ctx, "GET", "/slow", nil, nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
