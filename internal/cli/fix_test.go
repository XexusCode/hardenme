package cli

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/XexusCode/hardenme/internal/fix"
	"github.com/XexusCode/hardenme/internal/runner"
)

const fwProbe = "/usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate"

// switchRunner answers osascript calls with a fixed output and runs a hook,
// so a test can flip the probe's answer when the "change" happens.
type switchRunner struct {
	out    runner.Output
	onCall func()
	calls  int
}

func (s *switchRunner) Run(_ context.Context, name string, _ ...string) (runner.Output, error) {
	s.calls++
	if s.onCall != nil {
		s.onCall()
	}
	return s.out, nil
}

func runFix(t *testing.T, deps fixDeps, input string, args ...string) (string, error) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	opts := &options{lang: "es"}
	cmd := newFixWith(opts, deps)
	var out bytes.Buffer
	cmd.SetIn(strings.NewReader(input))
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), err
}

func testDeps(t *testing.T, probe *runner.Fake, apply runner.Runner) (fixDeps, fix.Log) {
	l := fix.Log{Path: filepath.Join(t.TempDir(), "changes.json")}
	return fixDeps{
		probe:    probe,
		apply:    apply,
		log:      func() (fix.Log, error) { return l, nil },
		terminal: func(io.Reader) bool { return true },
		now:      func() time.Time { return time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC) },
	}, l
}

func TestFixAppliesRechecksAndUndoes(t *testing.T) {
	probe := &runner.Fake{Outputs: map[string]runner.Output{
		fwProbe: {Stdout: "Firewall is disabled. (State = 0)"},
	}}
	apply := &switchRunner{
		out: runner.Output{Stdout: "HARDENME_STEP 0 0\n"},
		onCall: func() {
			probe.Outputs[fwProbe] = runner.Output{Stdout: "Firewall is enabled. (State = 1)"}
		},
	}
	deps, l := testDeps(t, probe, apply)

	out, err := runFix(t, deps, "1\n\n", "--only", "firewall")
	if err != nil {
		t.Fatalf("fix: %v\n%s", err, out)
	}
	for _, want := range []string{"1) Activa el cortafuegos  [lo hago yo]", "macOS te pedirá tu contraseña", "✓ El cortafuegos está activado", "hardenme fix --undo"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	if apply.calls != 1 {
		t.Fatalf("osascript called %d times, want 1", apply.calls)
	}
	batch, ok, _ := l.Last()
	if !ok || len(batch.Changes) != 1 || batch.Changes[0].Undo[2] != "off" {
		t.Fatalf("log = %+v", batch)
	}

	out, err = runFix(t, deps, "\n", "--undo")
	if err != nil {
		t.Fatalf("undo: %v\n%s", err, out)
	}
	if !strings.Contains(out, "Activa el cortafuegos") || !strings.Contains(out, "vuelven a estar como antes") {
		t.Errorf("undo output:\n%s", out)
	}
	if _, ok, _ := l.Last(); ok {
		t.Fatal("the undone batch must be removed from the log")
	}
}

func TestFixCancelledDialogKeepsNothing(t *testing.T) {
	probe := &runner.Fake{Outputs: map[string]runner.Output{fwProbe: {Stdout: "Firewall is disabled. (State = 0)"}}}
	apply := &switchRunner{out: runner.Output{Stderr: "User canceled. (-128)", ExitCode: 1}}
	deps, l := testDeps(t, probe, apply)
	out, err := runFix(t, deps, "1\n\n", "--only", "firewall")
	if code, ok := err.(ExitError); !ok || code.Code != 1 {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(out, "Se cerró la ventana de la contraseña") {
		t.Errorf("output:\n%s", out)
	}
	if _, ok, _ := l.Last(); ok {
		t.Fatal("nothing must be logged when the dialog is cancelled")
	}
}

func TestFixDryRunChangesNothing(t *testing.T) {
	probe := &runner.Fake{Outputs: map[string]runner.Output{fwProbe: {Stdout: "Firewall is disabled. (State = 0)"}}}
	apply := &switchRunner{}
	deps, _ := testDeps(t, probe, apply)
	out, err := runFix(t, deps, "", "--only", "firewall", "--dry-run")
	if err != nil {
		t.Fatal(err)
	}
	if apply.calls != 0 {
		t.Fatal("dry run must not apply anything")
	}
	if !strings.Contains(out, "[admin] /usr/libexec/ApplicationFirewall/socketfilterfw --setglobalstate on") {
		t.Errorf("output:\n%s", out)
	}
}

func TestFixRefusesWithoutTerminal(t *testing.T) {
	probe := &runner.Fake{Outputs: map[string]runner.Output{fwProbe: {Stdout: "Firewall is disabled. (State = 0)"}}}
	deps, _ := testDeps(t, probe, &switchRunner{})
	deps.terminal = func(io.Reader) bool { return false }
	if _, err := runFix(t, deps, "", "--only", "firewall"); err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("err = %v", err)
	}
}

func TestParseSelection(t *testing.T) {
	cases := []struct {
		in   string
		want []int
		ok   bool
	}{
		{"", nil, true},
		{"1,3", []int{1, 3}, true},
		{"3 1 3", []int{3, 1}, true},
		{"2-4", []int{2, 3, 4}, true},
		{"todo", []int{1, 2, 3, 4}, true},
		{"5", nil, false},
		{"0", nil, false},
		{"uno", nil, false},
		{"3-2", nil, false},
	}
	for _, c := range cases {
		got, ok := parseSelection(c.in, 4)
		if ok != c.ok || !reflect.DeepEqual(got, c.want) {
			t.Errorf("parseSelection(%q) = %v, %v; want %v, %v", c.in, got, ok, c.want, c.ok)
		}
	}
}
