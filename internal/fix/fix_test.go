package fix_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/XexusCode/hardenme/checks"
	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/fix"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/runner"
	"github.com/XexusCode/hardenme/internal/scan"
)

// fakeRunner answers by command name and records every call.
type fakeRunner struct {
	handle func(name string, args []string) (runner.Output, error)
	calls  [][]string
}

func (f *fakeRunner) Run(_ context.Context, name string, args ...string) (runner.Output, error) {
	f.calls = append(f.calls, append([]string{name}, args...))
	return f.handle(name, args)
}

func catalog(t *testing.T) []check.Check {
	t.Helper()
	all, err := check.Load(checks.FS, "macos")
	if err != nil {
		t.Fatal(err)
	}
	return all
}

func byID(t *testing.T, all []check.Check, ids ...string) []check.Check {
	t.Helper()
	var out []check.Check
	for _, id := range ids {
		found := false
		for _, c := range all {
			if c.ID == id {
				out = append(out, c)
				found = true
			}
		}
		if !found {
			t.Fatalf("no check %s", id)
		}
	}
	return out
}

func TestCandidatesOrderAndFilter(t *testing.T) {
	all := catalog(t)
	entry := func(id string, rel check.Relevance, st check.Status) scan.Entry {
		return scan.Entry{Check: byID(t, all, id)[0], Relevance: rel, Result: check.Result{CheckID: id, Status: st}}
	}
	report := scan.Report{Entries: []scan.Entry{
		entry("macos.firewall_stealth", check.Recommended, check.Fail),
		entry("macos.sip", check.Essential, check.Fail),             // info only
		entry("macos.firewall", check.Essential, check.Fail),        // auto
		entry("macos.filevault", check.Essential, check.Fail),       // guided
		entry("macos.guest_account", check.Recommended, check.Pass), // passes
		entry("macos.airdrop", check.NotApplicable, check.Fail),     // not applicable
		entry("macos.auto_lock", check.Essential, check.Warn),
	}}
	var got []string
	for _, it := range fix.Candidates(report) {
		got = append(got, string(it.Kind)+":"+it.Entry.Check.ID)
	}
	want := []string{"guided:macos.filevault", "auto:macos.firewall", "auto:macos.auto_lock", "auto:macos.firewall_stealth"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestEveryAutoFixHasAnUndoAndIsVerified(t *testing.T) {
	// Only the fixes verified on a real Mac (plan, phase 1) may be automatic.
	verified := map[string]bool{
		"macos.firewall": true, "macos.firewall_stealth": true, "macos.guest_account": true,
		"macos.auto_updates": true, "macos.security_responses": true, "macos.auto_lock": true,
	}
	for _, c := range catalog(t) {
		if c.FixKind() != check.FixAuto {
			continue
		}
		if !verified[c.ID] {
			t.Errorf("%s has an automatic fix that was not verified on a real Mac", c.ID)
		}
		for i, s := range c.Fix.Steps {
			if len(s.Undo) == 0 {
				t.Errorf("%s step %d has no undo", c.ID, i)
			}
		}
	}
}

func TestPrepareReadsPreviousValues(t *testing.T) {
	all := catalog(t)
	r := &fakeRunner{handle: func(name string, args []string) (runner.Output, error) {
		key := runner.Key(name, args...)
		switch {
		case key == "/usr/bin/pmset -g custom":
			return runner.Output{Stdout: "Battery Power:\n displaysleep         5\nAC Power:\n displaysleep         120\n"}, nil
		case strings.Contains(key, "idleTime"):
			return runner.Output{Stderr: "does not exist", ExitCode: 1}, nil
		case strings.HasSuffix(key, "AutomaticCheckEnabled"):
			return runner.Output{Stdout: "1\n"}, nil
		case strings.HasSuffix(key, "AutomaticDownload"):
			return runner.Output{Stdout: "0\n"}, nil
		case strings.HasSuffix(key, "AutomaticallyInstallMacOSUpdates"):
			return runner.Output{Stdout: "weird\n"}, nil
		}
		t.Fatalf("unexpected command %s", key)
		return runner.Output{}, nil
	}}
	steps, skips := fix.Prepare(context.Background(), r, byID(t, all, "macos.firewall_stealth", "macos.auto_lock", "macos.firewall", "macos.auto_updates"))

	var ids []string
	undo := map[string]string{}
	for _, s := range steps {
		ids = append(ids, s.CheckID)
		undo[strings.Join(s.Apply, " ")] = strings.Join(s.Undo, " ")
	}
	// Sorted by check ID: the firewall goes on before stealth mode.
	want := []string{"macos.auto_lock", "macos.auto_lock", "macos.auto_lock", "macos.auto_updates", "macos.auto_updates", "macos.firewall", "macos.firewall_stealth"}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("steps %v, want %v", ids, want)
	}
	checks := map[string]string{
		"/usr/bin/pmset -c displaysleep 5":                                                                       "/usr/bin/pmset -c displaysleep 120",
		"/usr/bin/pmset -b displaysleep 5":                                                                       "/usr/bin/pmset -b displaysleep 5",
		"/usr/bin/defaults -currentHost write com.apple.screensaver idleTime -int 300":                           "/usr/bin/defaults -currentHost delete com.apple.screensaver idleTime",
		"/usr/bin/defaults write /Library/Preferences/com.apple.SoftwareUpdate AutomaticDownload -bool true":     "/usr/bin/defaults write /Library/Preferences/com.apple.SoftwareUpdate AutomaticDownload -bool false",
		"/usr/bin/defaults write /Library/Preferences/com.apple.SoftwareUpdate AutomaticCheckEnabled -bool true": "/usr/bin/defaults write /Library/Preferences/com.apple.SoftwareUpdate AutomaticCheckEnabled -bool true",
	}
	for apply, wantUndo := range checks {
		if undo[apply] != wantUndo {
			t.Errorf("undo of %q = %q, want %q", apply, undo[apply], wantUndo)
		}
	}
	// An unexpected value cannot be undone safely, so that step is skipped.
	if len(skips) != 1 || skips[0].CheckID != "macos.auto_updates" || !strings.Contains(skips[0].Reason, "weird") {
		t.Fatalf("skips = %+v", skips)
	}
}

func TestPrepareDropsBatteryStepOnDesktop(t *testing.T) {
	r := &fakeRunner{handle: func(name string, args []string) (runner.Output, error) {
		if name == "/usr/bin/pmset" {
			return runner.Output{Stdout: "AC Power:\n displaysleep         30\n"}, nil
		}
		return runner.Output{Stdout: "600\n"}, nil
	}}
	steps, skips := fix.Prepare(context.Background(), r, byID(t, catalog(t), "macos.auto_lock"))
	if len(skips) != 0 {
		t.Fatalf("skips = %+v", skips)
	}
	for _, s := range steps {
		if strings.Contains(strings.Join(s.Apply, " "), "pmset -b") {
			t.Fatal("battery step prepared on a Mac without a battery")
		}
	}
	if len(steps) != 2 {
		t.Fatalf("got %d steps, want 2", len(steps))
	}
}

func TestScriptQuotingRunsInARealShell(t *testing.T) {
	dir := t.TempDir()
	tricky := filepath.Join(dir, `it's "a" file`)
	steps := []fix.Step{
		{Apply: []string{"/usr/bin/touch", tricky}},
		{Apply: []string{"/usr/bin/false"}},
		{Apply: []string{"/bin/echo", `back\slash`}},
	}
	out, err := exec.Command("/bin/sh", "-c", fix.Script(steps)).Output()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(tricky); err != nil {
		t.Fatalf("the quoted path was not created: %v", err)
	}
	got := fix.ParseResults(string(out))
	if !reflect.DeepEqual(got, map[int]int{0: 0, 1: 1, 2: 0}) {
		t.Fatalf("results %v from %q", got, out)
	}
}

func TestAppleScriptEscaping(t *testing.T) {
	got := fix.AppleScript(`echo "x" 'a\b'`)
	want := `do shell script "echo \"x\" 'a\\b'" with administrator privileges without altering line endings`
	if got != want {
		t.Fatalf("got  %s\nwant %s", got, want)
	}
}

func TestRunUsesOnePasswordDialog(t *testing.T) {
	r := &fakeRunner{handle: func(name string, args []string) (runner.Output, error) {
		if name == fix.Osascript {
			return runner.Output{Stdout: "HARDENME_STEP 0 0\rHARDENME_STEP 1 3\r"}, nil
		}
		return runner.Output{}, nil
	}}
	steps := []fix.Step{
		{CheckID: "a", Admin: true, Apply: []string{"/bin/a"}, Undo: []string{"/bin/undo-a"}},
		{CheckID: "u", Apply: []string{"/bin/u"}, Undo: []string{"/bin/undo-u"}},
		{CheckID: "b", Admin: true, Apply: []string{"/bin/b"}, Undo: []string{"/bin/undo-b"}},
	}
	res := fix.Run(context.Background(), r, steps)
	osa := 0
	for _, c := range r.calls {
		if c[0] == fix.Osascript {
			osa++
		}
	}
	if osa != 1 || len(r.calls) != 2 {
		t.Fatalf("calls %v: want one osascript call and one user command", r.calls)
	}
	changes := res.Changes()
	var ids []string
	for _, c := range changes {
		ids = append(ids, c.CheckID)
	}
	// b exited 3, so only a and u are logged.
	if !reflect.DeepEqual(ids, []string{"a", "u"}) {
		t.Fatalf("changes %v", ids)
	}
}

func TestRunCancelledDialog(t *testing.T) {
	r := &fakeRunner{handle: func(name string, args []string) (runner.Output, error) {
		return runner.Output{Stderr: "execution error: User canceled. (-128)", ExitCode: 1}, nil
	}}
	res := fix.Run(context.Background(), r, []fix.Step{{CheckID: "a", Admin: true, Apply: []string{"/bin/a"}, Undo: []string{"/bin/x"}}})
	if !res.Cancelled || len(res.Changes()) != 0 {
		t.Fatalf("got %+v", res)
	}
	if ran, cancelled := res.Touched("a"); ran || !cancelled {
		t.Fatalf("touched = %v, %v", ran, cancelled)
	}
}

func TestRecheckDecidesTheOutcome(t *testing.T) {
	all := catalog(t)
	fw := byID(t, all, "macos.firewall")[0]
	report := scan.Report{Entries: []scan.Entry{{Check: fw, Relevance: check.Essential, Result: check.Result{Status: check.Fail}}}}
	env := check.Env{Profile: profile.Default(true), Runner: &runner.Fake{Outputs: map[string]runner.Output{
		"/usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate": {Stdout: "Firewall is disabled. (State = 0)"},
	}}}
	// The step "succeeded" but the firewall is still off: not changed.
	res := fix.Result{Steps: []fix.StepResult{{Step: fix.Step{CheckID: fw.ID, Admin: true}, Ran: true}}}
	out := fix.Recheck(context.Background(), env, &report, res, []string{fw.ID})
	if out[fw.ID] != fix.NotChanged {
		t.Fatalf("got %s", out[fw.ID])
	}
	env.Runner = &runner.Fake{Outputs: map[string]runner.Output{
		"/usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate": {Stdout: "Firewall is enabled. (State = 1)"},
	}}
	out = fix.Recheck(context.Background(), env, &report, res, []string{fw.ID})
	if out[fw.ID] != fix.Fixed || report.Entries[0].Result.Status != check.Pass {
		t.Fatalf("got %s, %s", out[fw.ID], report.Entries[0].Result.Status)
	}
}

func TestUndoStepsOnlyAllowsCatalogCommands(t *testing.T) {
	all := catalog(t)
	ok := fix.Batch{Changes: []fix.Change{
		{CheckID: "macos.firewall", Admin: true, Undo: []string{"/usr/libexec/ApplicationFirewall/socketfilterfw", "--setglobalstate", "off"}},
		{CheckID: "macos.auto_lock", Admin: true, Undo: []string{"/usr/bin/pmset", "-c", "displaysleep", "120"}},
		{CheckID: "macos.auto_lock", Undo: []string{"/usr/bin/defaults", "-currentHost", "delete", "com.apple.screensaver", "idleTime"}},
	}}
	steps, err := fix.UndoSteps(ok, all)
	if err != nil {
		t.Fatal(err)
	}
	if steps[0].CheckID != "macos.auto_lock" || steps[2].CheckID != "macos.firewall" {
		t.Fatalf("undo must run newest first: %+v", steps)
	}

	bad := []fix.Change{
		{CheckID: "macos.firewall", Admin: true, Undo: []string{"/bin/rm", "-rf", "/"}},
		{CheckID: "macos.auto_lock", Admin: true, Undo: []string{"/usr/bin/pmset", "-c", "displaysleep", "1 && rm"}},
		{CheckID: "macos.auto_lock", Admin: true, Undo: []string{"/usr/bin/defaults", "-currentHost", "delete", "com.apple.screensaver", "idleTime"}}, // user step marked admin
		{CheckID: "macos.unknown", Undo: []string{"/usr/bin/true"}},
	}
	for _, ch := range bad {
		if _, err := fix.UndoSteps(fix.Batch{Changes: []fix.Change{ch}}, all); err == nil {
			t.Errorf("undo %v was allowed", ch)
		}
	}
}

func TestLogAppendLastDrop(t *testing.T) {
	l := fix.Log{Path: filepath.Join(t.TempDir(), "state", "changes.json")}
	if _, ok, err := l.Last(); err != nil || ok {
		t.Fatalf("empty log: %v %v", ok, err)
	}
	if err := l.Append(fix.Batch{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(l.Path); !os.IsNotExist(err) {
		t.Fatal("an empty batch must not create the log")
	}
	first := fix.Batch{Time: time.Unix(1, 0).UTC(), Changes: []fix.Change{{CheckID: "a", Undo: []string{"/x"}}}}
	second := fix.Batch{Time: time.Unix(2, 0).UTC(), Changes: []fix.Change{{CheckID: "b", Undo: []string{"/y"}}}}
	if err := l.Append(first); err != nil {
		t.Fatal(err)
	}
	if err := l.Append(second); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(l.Path)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("log mode %v", info.Mode().Perm())
	}
	last, ok, err := l.Last()
	if err != nil || !ok || !reflect.DeepEqual(last, second) {
		t.Fatalf("last = %+v %v %v", last, ok, err)
	}
	if err := l.DropLast(); err != nil {
		t.Fatal(err)
	}
	last, _, _ = l.Last()
	if !reflect.DeepEqual(last, first) {
		t.Fatalf("after drop: %+v", last)
	}
}
