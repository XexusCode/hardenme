package ui_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/XexusCode/hardenme/checks"
	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/fix"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/runner"
	"github.com/XexusCode/hardenme/internal/story"
	"github.com/XexusCode/hardenme/internal/ui"
)

const (
	host    = "127.0.0.1:4321"
	token   = "secret-token"
	fwProbe = "/usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate"
	fvProbe = "fdesetup status"
)

// applyRunner stands in for osascript: it answers with a fixed output and
// runs a hook, so a test can flip what the probes read.
type applyRunner struct {
	out    runner.Output
	onCall func()
	calls  int
}

func (a *applyRunner) Run(context.Context, string, ...string) (runner.Output, error) {
	a.calls++
	if a.onCall != nil {
		a.onCall()
	}
	return a.out, nil
}

type env struct {
	h     http.Handler
	probe *runner.Fake
	apply *applyRunner
	log   fix.Log
	prof  profile.Store
}

// newEnv serves two real checks: the firewall (automatic fix) and
// FileVault (guided), both failing.
func newEnv(t *testing.T) *env {
	t.Helper()
	all, err := check.Load(checks.FS, "macos")
	if err != nil {
		t.Fatal(err)
	}
	var catalog []check.Check
	for _, c := range all {
		if c.ID == "macos.firewall" || c.ID == "macos.filevault" {
			catalog = append(catalog, c)
		}
	}
	stories, err := story.Load(checks.FS, "macos")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	e := &env{
		probe: &runner.Fake{Outputs: map[string]runner.Output{
			fwProbe: {Stdout: "Firewall is disabled. (State = 0)"},
			fvProbe: {Stdout: "FileVault is Off.\n"},
		}},
		apply: &applyRunner{out: runner.Output{Stdout: "HARDENME_STEP 0 0\n"}},
		log:   fix.Log{Path: filepath.Join(dir, "changes.json")},
		prof:  profile.Store{Path: filepath.Join(dir, "profile.json")},
	}
	e.apply.onCall = func() {
		e.probe.Outputs[fwProbe] = runner.Output{Stdout: "Firewall is enabled. (State = 1)"}
	}
	srv := ui.New(ui.Deps{
		Platform: "macos", OSVersion: "26.2", Lang: check.ES, Catalog: catalog, Stories: stories,
		Probe: e.probe, Apply: e.apply,
		Log:      func() (fix.Log, error) { return e.log, nil },
		Profiles: e.prof, Fallback: profile.Default(true),
		Now: func() time.Time { return time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC) },
	}, host, token)
	e.h = srv.Handler()
	return e
}

type reqOpt func(*http.Request)

func (e *env) do(method, path, body string, opts ...reqOpt) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "http://"+host+path, bytes.NewBufferString(body))
	r.Header.Set("X-Hardenme-Token", token)
	for _, o := range opts {
		o(r)
	}
	w := httptest.NewRecorder()
	e.h.ServeHTTP(w, r)
	return w
}

func (e *env) reply(t *testing.T, method, path, body string) ui.Reply {
	t.Helper()
	w := e.do(method, path, body)
	if w.Code != http.StatusOK {
		t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body)
	}
	var out ui.Reply
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func item(t *testing.T, st ui.State, id string) ui.Item {
	t.Helper()
	for _, it := range st.Items {
		if it.ID == id {
			return it
		}
	}
	t.Fatalf("no item %s", id)
	return ui.Item{}
}

func TestOnlyThisRunsPageGetsIn(t *testing.T) {
	e := newEnv(t)
	tests := map[string]struct {
		method, path string
		opt          reqOpt
		want         int
	}{
		"good request":              {"GET", "/api/state", func(*http.Request) {}, 200},
		"no token":                  {"GET", "/api/state", func(r *http.Request) { r.Header.Del("X-Hardenme-Token") }, 403},
		"wrong token":               {"POST", "/api/fix", func(r *http.Request) { r.Header.Set("X-Hardenme-Token", "nope") }, 403},
		"token in the query on api": {"GET", "/api/state?t=" + token, func(r *http.Request) { r.Header.Del("X-Hardenme-Token") }, 403},
		"another host (rebinding)":  {"GET", "/api/state", func(r *http.Request) { r.Host = "evil.example:4321" }, 403},
		"localhost alias":           {"GET", "/api/state", func(r *http.Request) { r.Host = "localhost:4321" }, 403},
		"another site's script":     {"POST", "/api/fix", func(r *http.Request) { r.Header.Set("Origin", "https://evil.example") }, 403},
		"own origin":                {"POST", "/api/recheck", func(r *http.Request) { r.Header.Set("Origin", "http://"+host) }, 400},
		"page with the token":       {"GET", "/?t=" + token, func(r *http.Request) { r.Header.Del("X-Hardenme-Token") }, 200},
		"page without the token":    {"GET", "/", func(r *http.Request) { r.Header.Del("X-Hardenme-Token") }, 403},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if w := e.do(tc.method, tc.path, "", tc.opt); w.Code != tc.want {
				t.Errorf("status %d, want %d", w.Code, tc.want)
			}
		})
	}
	if e.apply.calls != 0 {
		t.Errorf("a rejected request ran a fix (%d calls)", e.apply.calls)
	}
}

func TestStateShowsScoreStoriesAndFixKinds(t *testing.T) {
	e := newEnv(t)
	st := e.reply(t, "GET", "/api/state", "").State
	if st.Score == nil || st.Score.Value != 0 || st.Score.Label != "En riesgo" {
		t.Errorf("score %+v", st.Score)
	}
	if len(st.Stories) != 1 || !strings.Contains(st.Stories[0], "Si te roban el Mac") {
		t.Errorf("stories %v", st.Stories)
	}
	if st.AutoFixes != 1 || st.CanUndo || st.ProfileSaved {
		t.Errorf("auto %d undo %v saved %v", st.AutoFixes, st.CanUndo, st.ProfileSaved)
	}
	fw, fv := item(t, st, "macos.firewall"), item(t, st, "macos.filevault")
	if fw.Fix != check.FixAuto || fw.Title != "Activa el cortafuegos" || fw.Why == "" || len(fw.Steps) == 0 {
		t.Errorf("firewall %+v", fw)
	}
	if fv.Fix != check.FixGuided || !fv.CanOpen || fv.Section != "essential" {
		t.Errorf("filevault %+v", fv)
	}
}

func TestFixEverythingThenUndo(t *testing.T) {
	e := newEnv(t)
	out := e.reply(t, "POST", "/api/fix", "{}")
	if e.apply.calls != 1 {
		t.Fatalf("password dialog shown %d times, want 1", e.apply.calls)
	}
	if out.Outcomes["macos.firewall"] != fix.Fixed {
		t.Errorf("outcomes %v", out.Outcomes)
	}
	if item(t, out.State, "macos.firewall").Status != check.Pass || out.State.AutoFixes != 0 || !out.State.CanUndo {
		t.Errorf("state after fix: %+v", out.State)
	}
	if out.State.Score.Value <= 0 {
		t.Errorf("the score did not go up: %+v", out.State.Score)
	}
	if _, ok := out.Outcomes["macos.filevault"]; ok {
		t.Error("a guided fix must never be applied by the button")
	}

	// Nothing left: the button does nothing and shows no dialog.
	again := e.reply(t, "POST", "/api/fix", "{}")
	if e.apply.calls != 1 || again.Message == "" {
		t.Errorf("second fix: calls %d, message %q", e.apply.calls, again.Message)
	}

	e.apply.onCall = func() {
		e.probe.Outputs[fwProbe] = runner.Output{Stdout: "Firewall is disabled. (State = 0)"}
	}
	undone := e.reply(t, "POST", "/api/undo", "{}")
	if item(t, undone.State, "macos.firewall").Status != check.Fail || undone.State.CanUndo {
		t.Errorf("state after undo: %+v", undone.State)
	}
	if !strings.Contains(undone.Message, "como antes") {
		t.Errorf("message %q", undone.Message)
	}
}

func TestCancelledDialogChangesNothing(t *testing.T) {
	e := newEnv(t)
	e.apply.out = runner.Output{Stderr: "User canceled. (-128)", ExitCode: 1}
	e.apply.onCall = nil
	out := e.reply(t, "POST", "/api/fix", "{}")
	if !strings.Contains(out.Message, "ventana de la contraseña") || out.State.CanUndo {
		t.Errorf("message %q, undo %v", out.Message, out.State.CanUndo)
	}
	if item(t, out.State, "macos.firewall").Status != check.Fail {
		t.Error("the firewall should still fail")
	}
}

func TestGuidedFixTicksItself(t *testing.T) {
	e := newEnv(t)
	e.reply(t, "GET", "/api/state", "")
	e.reply(t, "POST", "/api/open", `{"id":"macos.filevault"}`)
	opened := false
	for _, call := range e.probe.Calls {
		if strings.HasPrefix(call, "/usr/bin/open x-apple.systempreferences:") {
			opened = true
		}
	}
	if !opened {
		t.Errorf("the setting was not opened: %v", e.probe.Calls)
	}
	if st := e.reply(t, "POST", "/api/recheck", `{"id":"macos.filevault"}`).State; item(t, st, "macos.filevault").Status != check.Fail {
		t.Error("still off, should still fail")
	}
	// The person turns FileVault on in System Settings.
	e.probe.Outputs[fvProbe] = runner.Output{Stdout: "FileVault is On.\n"}
	st := e.reply(t, "POST", "/api/recheck", `{"id":"macos.filevault"}`).State
	if item(t, st, "macos.filevault").Status != check.Pass || len(st.Stories) != 0 {
		t.Errorf("after turning it on: %+v, stories %v", item(t, st, "macos.filevault"), st.Stories)
	}
	// The risk turns into a reassurance.
	if len(st.Safe) != 1 || !strings.Contains(st.Safe[0], "cifrado") {
		t.Errorf("safe stories %v", st.Safe)
	}
}

func TestThePageCannotChooseWhatRuns(t *testing.T) {
	e := newEnv(t)
	for name, body := range map[string]string{
		"unknown check":     `{"id":"macos.nope"}`,
		"a link, not an id": `{"id":"x-apple.systempreferences:evil"}`,
		"a command":         `{"id":"macos.firewall; rm -rf /"}`,
		"not json":          `id=macos.firewall`,
	} {
		t.Run(name, func(t *testing.T) {
			if w := e.do("POST", "/api/open", body); w.Code == http.StatusOK {
				t.Errorf("accepted: %s", w.Body)
			}
		})
	}
	// The firewall check has a pane; a check without one opens nothing.
	for _, call := range e.probe.Calls {
		if strings.HasPrefix(call, "/usr/bin/open") {
			t.Errorf("opened something: %s", call)
		}
	}
}

func TestProfileIsSavedAndChangesTheScore(t *testing.T) {
	e := newEnv(t)
	st := e.reply(t, "POST", "/api/profile", `{"portable":false,"leaves_home":false,"shared":false,"sensitive_data":false,"risk":"normal"}`).State
	if !st.ProfileSaved || st.Profile.Portable {
		t.Errorf("profile %+v saved %v", st.Profile, st.ProfileSaved)
	}
	if p, err := e.prof.Load(); err != nil || p.Portable {
		t.Errorf("stored profile %+v %v", p, err)
	}
	for name, body := range map[string]string{
		"bad risk":      `{"risk":"extreme"}`,
		"unknown field": `{"risk":"normal","admin":true}`,
	} {
		if w := e.do("POST", "/api/profile", body); w.Code != http.StatusBadRequest {
			t.Errorf("%s: status %d", name, w.Code)
		}
	}
}

func TestPageHasNoRemoteResources(t *testing.T) {
	e := newEnv(t)
	w := e.do("GET", "/?t="+token, "")
	body := w.Body.String()
	for _, bad := range []string{"https://", "http://", "innerHTML", "eval("} {
		if strings.Contains(strings.ReplaceAll(body, "http://www.w3.org/2000/svg", ""), bad) {
			t.Errorf("the page contains %q", bad)
		}
	}
	if csp := w.Header().Get("Content-Security-Policy"); !strings.Contains(csp, "default-src 'none'") || !strings.Contains(csp, "connect-src 'self'") {
		t.Errorf("CSP %q", csp)
	}
}
