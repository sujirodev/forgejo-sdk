// Copyright 2026 The Forgejo Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package forgejo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The route recorder answers the question the static check-route-coverage.sh
// cannot: which routes did the suite *actually* exercise against this server,
// and with what result. It wraps the test client's transport, so every request
// the SDK emits passes through it, and attributes each request to the exported
// *Client methods on the call stack.
//
// The report it writes (forgejo/route-report.json) is the per-version evidence
// that feeds ROUTES.md. See docs/PLANO-CONTRATO-ROTAS.md, section 2.1.

// routeRecorderFrames bounds the stack walk. The frames between an exported
// *Client method and RoundTrip are http.Client internals (a handful), plus
// whatever exported methods call each other; 64 leaves ample room.
const routeRecorderFrames = 64

// routeStat is one route's line in the report.
type routeStat struct {
	HTTPMethods []string       `json:"http_methods"`
	Statuses    map[string]int `json:"statuses"`
	Calls       int            `json:"calls"`
	OK          bool           `json:"ok"`
}

// unattributedCall is a request that reached the server without any exported
// *Client method on the stack. The generator fails when this list is not
// empty: a request nobody can be credited for is a hole in the evidence.
type unattributedCall struct {
	HTTPMethod string `json:"http_method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
}

// routeReport is the on-disk format, one file per test leg.
type routeReport struct {
	Label         string               `json:"label"`
	GeneratedAt   string               `json:"generated_at"`
	ServerURL     string               `json:"server_url"`
	ServerVersion string               `json:"server_version"`
	Routes        map[string]routeStat `json:"routes"`
	Unattributed  []unattributedCall   `json:"unattributed"`
}

type routeEntry struct {
	httpMethods map[string]bool
	statuses    map[int]int
	calls       int
	ok          bool
}

type routeRecorder struct {
	base http.RoundTripper

	mu           sync.Mutex
	routes       map[string]*routeEntry
	unattributed []unattributedCall
}

func newRouteRecorder() *routeRecorder {
	return &routeRecorder{base: http.DefaultTransport, routes: map[string]*routeEntry{}}
}

func (r *routeRecorder) RoundTrip(req *http.Request) (*http.Response, error) {
	// Capture the stack before the request goes out: RoundTrip runs on the
	// caller's goroutine, but reading it first keeps the attribution honest
	// even if that ever changes.
	names := exportedClientMethodsOnStack()

	resp, err := r.base.RoundTrip(req)

	status := 0
	if resp != nil {
		status = resp.StatusCode
	}
	r.record(names, req, status)

	return resp, err
}

func (r *routeRecorder) record(names []string, req *http.Request, status int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(names) == 0 {
		r.unattributed = append(r.unattributed, unattributedCall{
			HTTPMethod: req.Method,
			Path:       req.URL.Path,
			Status:     status,
		})
		return
	}

	for _, name := range names {
		e := r.routes[name]
		if e == nil {
			e = &routeEntry{httpMethods: map[string]bool{}, statuses: map[int]int{}}
			r.routes[name] = e
		}
		e.httpMethods[req.Method] = true
		e.statuses[status]++
		e.calls++
		if status >= 200 && status < 300 {
			e.ok = true
		}
	}
}

func (r *routeRecorder) report(label, serverURL, serverVersion string) routeReport {
	r.mu.Lock()
	defer r.mu.Unlock()

	rep := routeReport{
		Label:         label,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		ServerURL:     serverURL,
		ServerVersion: serverVersion,
		Routes:        make(map[string]routeStat, len(r.routes)),
		Unattributed:  r.unattributed,
	}
	if rep.Unattributed == nil {
		rep.Unattributed = []unattributedCall{}
	}

	for name, e := range r.routes {
		methods := make([]string, 0, len(e.httpMethods))
		for m := range e.httpMethods {
			methods = append(methods, m)
		}
		sort.Strings(methods)

		statuses := make(map[string]int, len(e.statuses))
		for code, n := range e.statuses {
			statuses[strconv.Itoa(code)] = n
		}

		rep.Routes[name] = routeStat{
			HTTPMethods: methods,
			Statuses:    statuses,
			Calls:       e.calls,
			OK:          e.ok,
		}
	}

	return rep
}

func (r *routeRecorder) writeReport(path, label, serverURL, serverVersion string) error {
	b, err := json.MarshalIndent(r.report(label, serverURL, serverVersion), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o600)
}

// exportedClientMethodsOnStack returns every exported *Client method currently
// on the call stack, outermost last. Crediting all of them (not just the
// innermost) is what makes a composed method such as DeleteMilestoneByName
// credit the routes it delegates to as well.
func exportedClientMethodsOnStack() []string {
	var pcs [routeRecorderFrames]uintptr
	n := runtime.Callers(2, pcs[:])
	if n == 0 {
		return nil
	}

	frames := runtime.CallersFrames(pcs[:n])
	var names []string
	seen := map[string]bool{}
	for {
		frame, more := frames.Next()
		if name, ok := clientMethodName(frame.Function); ok && !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
		if !more {
			break
		}
	}
	return names
}

// sdkClientMarker is this package's own "(*Client)." prefix, resolved at init
// from a real function name. Anchoring on the package path matters: net/http
// also has a *Client, and net/http.(*Client).Do sits on the stack between
// every route and RoundTrip. A package-agnostic marker would record a route
// called "Do" on every single request.
var sdkClientMarker = sdkPackagePath() + ".(*Client)."

func sdkPackagePath() string {
	pc, _, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	fn := runtime.FuncForPC(pc).Name() // "<package path>.sdkPackagePath"
	if i := strings.LastIndex(fn, "."); i >= 0 {
		return fn[:i]
	}
	return ""
}

// clientMethodName extracts "ListRepoTags" from a runtime function name such as
// "codeberg.org/sujirodev/forgejo-sdk/forgejo/v3.(*Client).ListRepoTags"
// (and from its closures, "...(*Client).ListRepoTags.func1"). Unexported
// methods are not routes and are skipped, and so is any other package's
// *Client.
func clientMethodName(fn string) (string, bool) {
	if !strings.HasPrefix(fn, sdkClientMarker) {
		return "", false
	}
	name := fn[len(sdkClientMarker):]
	if j := strings.Index(name, "."); j >= 0 {
		name = name[:j]
	}
	if name == "" || name[0] < 'A' || name[0] > 'Z' {
		return "", false
	}
	return name, true
}

// testRouteRecorder is installed by newTestClient and drained by TestMain.
var testRouteRecorder = newRouteRecorder()

// routeReportLabel names the test leg this report came from. The version
// matrix sets it per leg (lts-11, lts-15, latest-stable); a local run keeps
// the default, which the generator refuses to publish as a matrix column.
func routeReportLabel() string {
	if l := os.Getenv("FORGEJO_SDK_TEST_LABEL"); l != "" {
		return l
	}
	return "local"
}

const routeReportPath = "route-report.json"

// writeRouteReport is called from TestMain after m.Run(). A failure to write is
// reported but does not change the suite's exit code: the report is evidence,
// not a test.
func writeRouteReport() {
	url := getForgejoURL()
	if err := testRouteRecorder.writeReport(routeReportPath, routeReportLabel(), url, serverVersionOf(url)); err != nil {
		fmt.Fprintf(os.Stderr, "route report: %v\n", err)
	}
}

// serverVersionOf asks the instance which version answered this leg, so the
// generated table can name the version instead of trusting the CI label. It
// uses a bare client on purpose: going through the recorder would credit
// ServerVersion for a call no test made.
func serverVersionOf(url string) string {
	if url == "" {
		return ""
	}
	resp, err := http.Get(url + "/api/v1/version") //nolint:noctx // test-only, runs after the suite
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var v struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return ""
	}
	return v.Version
}
