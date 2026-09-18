// Copyright 2024 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
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

func newTestClient() *Client {
	c, _ := NewClient(getForgejoURL(), newTestClientAuth())
	return c
}

func newTestClientAuth() ClientOption {
	token := getForgejoToken()
	if token == "" {
		return SetBasicAuth(getForgejoUsername(), getForgejoPassword())
	}
	return SetToken(getForgejoToken())
}

// renovate: datasource=docker depName=codeberg.org/forgejo/forgejo
const testForgejoVersion = "15.0.1"

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

	_, err = cfg.WriteString(`[security]
INTERNAL_TOKEN = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYmYiOjE1NTg4MzY4ODB9.LoKQyK5TN_0kMJFVHWUW0uDAyoGjDP6Mkup4ps2VJN4
INSTALL_LOCK   = true
SECRET_KEY     = 2crAW4UANgvLipDS6U5obRcFosjSJHQANll6MNfX7P0G3se3fKcCwwK3szPyGcbo
[database]
DB_TYPE  = sqlite3
[log]
MODE = console
LEVEL = Trace
REDIRECT_MACARON_LOG = true
MACARON = ,
ROUTER = ,`)
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
	exit(exitCode)
}

func exit(code int) {
	os.Exit(code)
}
