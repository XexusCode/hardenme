// Package fix applies the fixes the user picked after a scan: it captures the
// previous values, runs every admin change behind one native macOS password
// dialog, re-checks the result and keeps a log so the changes can be undone.
package fix

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/runner"
	"github.com/XexusCode/hardenme/internal/scan"
)

// Item is a finding the user can fix.
type Item struct {
	Entry scan.Entry
	Kind  check.FixKind
}

// Candidates lists the failing or warned checks that apply to the profile
// and that hardenme can apply or guide, most important first.
func Candidates(r scan.Report) []Item {
	var items []Item
	for _, e := range r.Entries {
		if e.Relevance == check.NotApplicable {
			continue
		}
		if e.Result.Status != check.Fail && e.Result.Status != check.Warn {
			continue
		}
		kind := e.Check.FixKind()
		if kind == check.FixInfo {
			continue
		}
		items = append(items, Item{Entry: e, Kind: kind})
	}
	sort.SliceStable(items, func(i, j int) bool {
		wi, wj := items[i].Entry.Relevance.Weight(), items[j].Entry.Relevance.Weight()
		if wi != wj {
			return wi > wj
		}
		if (items[i].Entry.Result.Status == check.Fail) != (items[j].Entry.Result.Status == check.Fail) {
			return items[i].Entry.Result.Status == check.Fail
		}
		return items[i].Entry.Check.ID < items[j].Entry.Check.ID
	})
	return items
}

// Step is one concrete change, ready to run.
type Step struct {
	CheckID string
	Admin   bool
	Apply   []string
	Undo    []string
}

// Skip is a step that was not prepared, with the reason.
type Skip struct {
	CheckID string
	Reason  string
}

// safeValue is what a captured previous value may look like.
var safeValue = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// Prepare turns the checks' fix steps into concrete steps, reading each
// previous value first. Steps that do not apply to this device (for example
// the battery setting on a desktop) are dropped silently; steps that could
// not be undone safely are reported as skipped. Steps are ordered by check
// ID, so macos.firewall runs before macos.firewall_stealth.
func Prepare(ctx context.Context, r runner.Runner, checks []check.Check) ([]Step, []Skip) {
	sorted := append([]check.Check(nil), checks...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	var steps []Step
	var skips []Skip
	for _, c := range sorted {
		if c.Fix == nil {
			continue
		}
		for _, fs := range c.Fix.Steps {
			step, ok, reason := prepareStep(ctx, r, c.ID, fs)
			switch {
			case ok:
				steps = append(steps, step)
			case reason != "":
				skips = append(skips, Skip{CheckID: c.ID, Reason: reason})
			}
		}
	}
	return steps, skips
}

func prepareStep(ctx context.Context, r runner.Runner, id string, fs check.FixStep) (Step, bool, string) {
	step := Step{CheckID: id, Admin: fs.Admin, Apply: fs.Apply}
	if len(fs.Read) == 0 {
		step.Undo = fs.Undo
		return step, true, ""
	}
	out, err := r.Run(ctx, fs.Read[0], fs.Read[1:]...)
	if err != nil {
		return step, false, "could not read the current value: " + err.Error()
	}
	if out.ExitCode != 0 {
		if len(fs.UndoIfMissing) > 0 {
			step.Undo = fs.UndoIfMissing
			return step, true, ""
		}
		return step, false, "could not read the current value"
	}
	prev := strings.TrimSpace(out.Stdout)
	if re := fs.Matcher(); re != nil {
		m := re.FindStringSubmatch(out.Combined())
		if m == nil {
			return step, false, "" // does not apply to this device
		}
		prev = m[1]
	}
	if fs.PreviousMap != nil {
		mapped, ok := fs.PreviousMap[prev]
		if !ok {
			return step, false, fmt.Sprintf("unexpected current value %q", prev)
		}
		prev = mapped
	}
	if !safeValue.MatchString(prev) {
		return step, false, fmt.Sprintf("unexpected current value %q", prev)
	}
	step.Undo = fill(fs.Undo, prev)
	return step, true, ""
}

func fill(argv []string, prev string) []string {
	out := make([]string, len(argv))
	for i, a := range argv {
		out[i] = strings.ReplaceAll(a, "{previous}", prev)
	}
	return out
}

// StepResult is what happened to one step.
type StepResult struct {
	Step
	Ran      bool
	ExitCode int
}

// Result is the outcome of running a list of steps.
type Result struct {
	Steps []StepResult
	// Cancelled is true when the user closed the password dialog.
	Cancelled bool
	// AdminErr is set when the admin batch could not run at all.
	AdminErr error
}

// Osascript is the program that shows the native password dialog.
const Osascript = "/usr/bin/osascript"

// Run applies the steps: every admin step in one osascript call (one
// password dialog), then the user-level steps directly.
func Run(ctx context.Context, r runner.Runner, steps []Step) Result {
	var res Result
	var admin []int
	res.Steps = make([]StepResult, len(steps))
	for i, s := range steps {
		res.Steps[i] = StepResult{Step: s}
		if s.Admin {
			admin = append(admin, i)
		}
	}
	if len(admin) > 0 {
		batch := make([]Step, len(admin))
		for k, i := range admin {
			batch[k] = steps[i]
		}
		out, err := r.Run(ctx, Osascript, "-e", AppleScript(Script(batch)))
		switch {
		case err != nil:
			res.AdminErr = err
		case out.ExitCode != 0 && strings.Contains(out.Stderr, "-128"):
			res.Cancelled = true
		case out.ExitCode != 0:
			res.AdminErr = fmt.Errorf("osascript: %s", strings.TrimSpace(out.Combined()))
		default:
			codes := ParseResults(out.Stdout)
			for k, i := range admin {
				if code, ok := codes[k]; ok {
					res.Steps[i].Ran = true
					res.Steps[i].ExitCode = code
				}
			}
		}
	}
	for i, s := range steps {
		if s.Admin {
			continue
		}
		out, err := r.Run(ctx, s.Apply[0], s.Apply[1:]...)
		if err != nil {
			continue
		}
		res.Steps[i].Ran = true
		res.Steps[i].ExitCode = out.ExitCode
	}
	return res
}

// Script builds the POSIX shell script for an admin batch. Every argument is
// single-quoted, and each step prints "HARDENME_STEP <n> <exit code>"
// without stopping the batch when a step fails.
func Script(steps []Step) string {
	parts := make([]string, 0, len(steps))
	for i, s := range steps {
		quoted := make([]string, len(s.Apply))
		for j, a := range s.Apply {
			quoted[j] = shellQuote(a)
		}
		parts = append(parts, fmt.Sprintf(`%s >/dev/null 2>&1; echo "HARDENME_STEP %d $?"`, strings.Join(quoted, " "), i))
	}
	return strings.Join(parts, "; ")
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// AppleScript wraps a shell script so macOS runs it as administrator after
// its own password dialog. hardenme never sees the password.
func AppleScript(script string) string {
	esc := strings.ReplaceAll(script, `\`, `\\`)
	esc = strings.ReplaceAll(esc, `"`, `\"`)
	return `do shell script "` + esc + `" with administrator privileges without altering line endings`
}

var resultLine = regexp.MustCompile(`HARDENME_STEP (\d+) (\d+)`)

// ParseResults reads the exit code of each step from the batch output.
func ParseResults(out string) map[int]int {
	codes := map[int]int{}
	for _, m := range resultLine.FindAllStringSubmatch(out, -1) {
		n, _ := strconv.Atoi(m[1])
		code, _ := strconv.Atoi(m[2])
		codes[n] = code
	}
	return codes
}

// Outcome is what happened to a check after fixing it.
type Outcome string

const (
	Fixed      Outcome = "fixed"
	NotChanged Outcome = "not_changed"
	Cancelled  Outcome = "cancelled"
	Skipped    Outcome = "skipped"
)

// Changes returns the steps that ran successfully and can be undone, for
// the change log.
func (r Result) Changes() []Change {
	var out []Change
	for _, s := range r.Steps {
		if s.Ran && s.ExitCode == 0 && len(s.Undo) > 0 {
			out = append(out, Change{CheckID: s.CheckID, Admin: s.Admin, Undo: s.Undo})
		}
	}
	return out
}

// Touched reports whether any step of the check ran, and whether its admin
// steps were cancelled.
func (r Result) Touched(id string) (ran, cancelled bool) {
	for _, s := range r.Steps {
		if s.CheckID != id {
			continue
		}
		if s.Ran {
			ran = true
		}
		if s.Admin && r.Cancelled {
			cancelled = true
		}
	}
	return ran, cancelled
}

// Recheck runs the checks again, updates the report and returns the outcome
// of each one.
func Recheck(ctx context.Context, env check.Env, report *scan.Report, res Result, ids []string) map[string]Outcome {
	outcomes := map[string]Outcome{}
	for _, id := range ids {
		for i := range report.Entries {
			e := &report.Entries[i]
			if e.Check.ID != id {
				continue
			}
			e.Result = e.Check.Run(ctx, env)
			ran, cancelled := res.Touched(id)
			switch {
			case e.Result.Status == check.Pass:
				outcomes[id] = Fixed
			case cancelled:
				outcomes[id] = Cancelled
			case !ran:
				outcomes[id] = Skipped
			default:
				outcomes[id] = NotChanged
			}
		}
	}
	return outcomes
}

// ErrNotAllowed is returned when a logged undo does not match the catalog.
var ErrNotAllowed = errors.New("undo command does not match any fix in this version of hardenme")

// UndoSteps turns a logged batch into steps, newest change first. The log
// is a file the user (or anything running as the user) can edit, and admin
// undos run as root, so every command must match an undo defined in the
// embedded catalog, with {previous} filled by a safe value.
func UndoSteps(b Batch, catalog []check.Check) ([]Step, error) {
	byID := map[string]check.Check{}
	for _, c := range catalog {
		byID[c.ID] = c
	}
	steps := make([]Step, 0, len(b.Changes))
	for i := len(b.Changes) - 1; i >= 0; i-- {
		ch := b.Changes[i]
		c, ok := byID[ch.CheckID]
		if !ok || c.Fix == nil || !allowed(c.Fix.Steps, ch) {
			return nil, fmt.Errorf("%s: %w", ch.CheckID, ErrNotAllowed)
		}
		steps = append(steps, Step{CheckID: ch.CheckID, Admin: ch.Admin, Apply: ch.Undo})
	}
	return steps, nil
}

func allowed(defs []check.FixStep, ch Change) bool {
	for _, d := range defs {
		if d.Admin != ch.Admin {
			continue
		}
		if matchTemplate(d.Undo, ch.Undo) || (len(d.UndoIfMissing) > 0 && matchTemplate(d.UndoIfMissing, ch.Undo)) {
			return true
		}
	}
	return false
}

func matchTemplate(tmpl, argv []string) bool {
	if len(tmpl) != len(argv) {
		return false
	}
	for i := range tmpl {
		if tmpl[i] == argv[i] {
			continue
		}
		if !strings.Contains(tmpl[i], "{previous}") {
			return false
		}
		prefix, suffix, _ := strings.Cut(tmpl[i], "{previous}")
		if !strings.HasPrefix(argv[i], prefix) || !strings.HasSuffix(argv[i], suffix) {
			return false
		}
		if !safeValue.MatchString(strings.TrimSuffix(strings.TrimPrefix(argv[i], prefix), suffix)) {
			return false
		}
	}
	return true
}
