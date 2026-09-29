// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func getForgejoURL() string {
	return os.Getenv("FORGEJO_SDK_TEST_URL")
}

func getForgejoToken() string {
	return os.Getenv("FORGEJO_SDK_TEST_TOKEN")
}

func getForgejoUsername() string {
	return os.Getenv("FORGEJO_SDK_TEST_USERNAME")
}

func getForgejoPassword() string {
	return os.Getenv("FORGEJO_SDK_TEST_PASSWORD")
}

func enableRunForgejo() bool {
	r, _ := strconv.ParseBool(os.Getenv("FORGEJO_SDK_TEST_RUN_FORGEJO"))
	return r
}

// runnerAttached reports whether the harness started a forgejo-runner with
// the "host" label next to the instance (scripts/start-test-runner.sh, run by
// `make test-instance-docker` and by CI). Tests that need a job to execute
// run only then; without a runner they would wait for a job that never starts.
func runnerAttached() bool {
	r, _ := strconv.ParseBool(os.Getenv("FORGEJO_SDK_TEST_RUNNER"))
	return r
}

func newTestClient() *Client {
	c, _ := newTestClientOpts()
	return c
}

// newTestClientOpts builds an integration test client with extra options on
// top of the standard auth and the route recorder. Every integration test
// builds its client here, never with a bare NewClient: a client built outside
// this helper bypasses the recorder and its requests land in the report's
// "unattributed" bucket.
func newTestClientOpts(opts ...ClientOption) (*Client, error) {
	all := make([]ClientOption, 0, len(opts)+2)
	all = append(all,
		newTestClientAuth(),
		SetHTTPClient(&http.Client{Transport: testRouteRecorder}),
	)
	all = append(all, opts...)
	return NewClient(getForgejoURL(), all...)
}

func newTestClientAuth() ClientOption {
	token := getForgejoToken()
	if token == "" {
		return SetBasicAuth(getForgejoUsername(), getForgejoPassword())
	}
	return SetToken(getForgejoToken())
}

// serverAtLeast reports whether c's server satisfies the given minimum
// version. Use it when a route's *documented* behavior genuinely differs
// across Forgejo versions (not a bug, and nothing the SDK can or should
// paper over): assert one thing when true, the other, explicitly, when
// false. A version check with no else branch on either side is not a
// version-aware test, it is a test that only runs half the time.
func serverAtLeast(t *testing.T, c *Client, minVersion string) bool {
	t.Helper()
	err := c.CheckServerVersionConstraint(">= " + minVersion)
	if err != nil && !strings.Contains(err.Error(), "does not satisfy version constraint") {
		t.Fatalf("serverAtLeast(%q): %v", minVersion, err)
	}
	return err == nil
}

// renovate: datasource=docker depName=codeberg.org/forgejo/forgejo
const testForgejoVersion = "16.0.5"

// TODO: replace with proper forgejo path
func forgejoMasterPath() string {
	switch runtime.GOOS {
	case "darwin":
		return fmt.Sprintf("https://codeberg.org/forgejo/forgejo/releases/download/v%[1]s/forgejo-%[1]s-%[2]s", testForgejoVersion, runtime.GOARCH)
	case "linux":
		return fmt.Sprintf("https://codeberg.org/forgejo/forgejo/releases/download/v%[1]s/forgejo-%[1]s-linux-%[2]s", testForgejoVersion, runtime.GOARCH)
	case "windows":
		return fmt.Sprintf("https://codeberg.org/forgejo/forgejo/releases/download/v%[1]s/forgejo-%[1]s-%[2]s.exe", testForgejoVersion, runtime.GOARCH)
	}
	return ""
}

func downForgejo() (string, error) {
	for i := 3; i > 0; i-- {
		resp, err := http.Get(forgejoMasterPath())
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		f, err := os.CreateTemp(os.TempDir(), "forgejo")
		if err != nil {
			continue
		}
		_, err = io.Copy(f, resp.Body)
		f.Close()
		if err != nil {
			continue
		}

		if err = os.Chmod(f.Name(), 0o700); err != nil {
			return "", err
		}

		return f.Name(), nil
	}

	return "", fmt.Errorf("Download forgejo from %v failed", forgejoMasterPath())
}

func runForgejo() (*os.Process, error) {
	log.Println("Downloading Forgejo from", forgejoMasterPath())
	p, err := downForgejo()
	if err != nil {
		log.Fatal(err)
	}

	forgejoDir := filepath.Dir(p)
	cfgDir := filepath.Join(forgejoDir, "custom", "conf")
	err = os.MkdirAll(cfgDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := os.Create(filepath.Join(cfgDir, "app.ini"))
	if err != nil {
		log.Fatal(err)
	}

	// The instance's SSH signing key: GET /signing-key.ssh only answers when
	// [repository.signing] names one. Only the public half is read, so only
	// it is written, freshly generated on every run. The sign-when options
	// below are "never" because the private half does not exist: with a
	// key configured Forgejo would otherwise try to sign every commit it
	// makes (repo init, file edits, merges) and fail.
	signingPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	sshPub, err := ssh.NewPublicKey(signingPub)
	if err != nil {
		log.Fatal(err)
	}
	signingKeyPath := filepath.Join(cfgDir, "signing_key.pub")
	if err = os.WriteFile(signingKeyPath, ssh.MarshalAuthorizedKey(sshPub), 0o644); err != nil {
		log.Fatal(err)
	}

	_, err = fmt.Fprintf(cfg, `[security]
INTERNAL_TOKEN = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYmYiOjE1NTg4MzY4ODB9.LoKQyK5TN_0kMJFVHWUW0uDAyoGjDP6Mkup4ps2VJN4
INSTALL_LOCK   = true
SECRET_KEY     = 2crAW4UANgvLipDS6U5obRcFosjSJHQANll6MNfX7P0G3se3fKcCwwK3szPyGcbo
DISABLE_GIT_HOOKS = false
[migrations]
ALLOW_LOCALNETWORKS = true
[repository]
ENABLE_FLAGS = true
[repository.signing]
FORMAT = ssh
SIGNING_KEY = %s
INITIAL_COMMIT = never
CRUD_ACTIONS = never
WIKI = never
MERGES = never
[quota]
ENABLED = true
[federation]
ENABLED = true
INSECURE_ALLOW_INVALID_HOSTS = true
[database]
DB_TYPE  = sqlite3
[log]
MODE = console
LEVEL = Trace
REDIRECT_MACARON_LOG = true
MACARON = ,
ROUTER = ,`, filepath.ToSlash(signingKeyPath))
	cfg.Close()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Run forgejo migrate", p)
	err = exec.Command(p, "migrate").Run()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Run forgejo admin", p)
	err = exec.Command(p, "admin", "create-user", "--username=test01", "--password=test01", "--email=test01@forgejo.org", "--admin=true", "--must-change-password=false", "--access-token").Run()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Start Forgejo", p)
	return os.StartProcess(filepath.Base(p), []string{}, &os.ProcAttr{
		Dir: forgejoDir,
	})
}

func TestMain(m *testing.M) {
	if enableRunForgejo() {
		p, err := runForgejo()
		if err != nil {
			log.Fatal(err)
			return
		}
		defer func() {
			if err := p.Kill(); err != nil {
				log.Fatal(err)
			}
		}()
	}
	log.Printf("testing with %v, %v, %v\n", getForgejoURL(), getForgejoUsername(), getForgejoPassword())
	exitCode := m.Run()
	writeRouteReport()
	exit(exitCode)
}

func exit(code int) {
	os.Exit(code)
}
