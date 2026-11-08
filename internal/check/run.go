package check

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/runner"
)

// Env is what a probe needs to run.
type Env struct {
	Runner  runner.Runner
	Profile profile.Profile
	// Ask gets the person's answer for ask probes. When nil, ask checks are
	// reported as manual.
	Ask func(Check) Answer
}

// Phone platforms have nothing hardenme can read, so their checks ask.
var askOnly = map[string]bool{"ios": true, "android": true}

// Builtin is a probe written in Go, for checks whose logic does not fit
// declarative rules (thresholds that depend on the profile, several reads...).
type Builtin func(ctx context.Context, env Env) Result

// Builtins is the registry of Go probes, keyed by the name used in YAML.
var Builtins = map[string]Builtin{}

// compile prepares a loaded check and reports definition errors.
func (c *Check) compile() error {
	p := &c.Probe
	kinds := 0
	if len(p.Command) > 0 {
		kinds++
	}
	if p.Builtin != "" {
		kinds++
	}
	if p.Manual {
		kinds++
	}
	if p.Ask {
		kinds++
	}
	if kinds != 1 {
		return errors.New("probe must set exactly one of command, builtin, manual or ask")
	}
	if askOnly[c.Platform] && !p.Ask && !p.Manual {
		return fmt.Errorf("%s checks can only use ask or manual probes", c.Platform)
	}
	if p.Builtin != "" {
		if _, ok := Builtins[p.Builtin]; !ok {
			return fmt.Errorf("unknown builtin %q", p.Builtin)
		}
	}
	if len(p.Command) > 0 && len(p.Rules) == 0 {
		return errors.New("command probe needs at least one rule")
	}
	for i := range p.Rules {
		r := &p.Rules[i]
		if r.Match == "" && r.Exit == nil {
			return fmt.Errorf("rule %d must set match or exit", i)
		}
		if r.Match != "" {
			re, err := regexp.Compile(r.Match)
			if err != nil {
				return fmt.Errorf("rule %d: %w", i, err)
			}
			r.re = re
		}
		if !validStatus(r.Status) {
			return fmt.Errorf("rule %d: invalid status %q", i, r.Status)
		}
	}
	if p.Default != "" && !validStatus(p.Default) {
		return fmt.Errorf("invalid default status %q", p.Default)
	}
	if !c.Relevance.Default.valid() {
		return fmt.Errorf("invalid default relevance %q", c.Relevance.Default)
	}
	for i, rule := range c.Relevance.Rules {
		if !rule.Level.valid() {
			return fmt.Errorf("relevance rule %d: invalid level %q", i, rule.Level)
		}
	}
	for _, lang := range []Lang{EN, ES} {
		t, ok := c.Text[lang]
		if !ok || t.Title == "" || t.TitlePass == "" || t.TitleFail == "" || t.Why == "" || len(t.FixSteps) == 0 {
			return fmt.Errorf("missing or incomplete %s text", lang)
		}
		if p.Ask != (t.Question != "") {
			return fmt.Errorf("%s text: question is required with ask and only allowed with ask", lang)
		}
		if t.Where != "" && !p.Ask {
			return fmt.Errorf("%s text: where is only allowed with ask", lang)
		}
	}
	if len(c.References) == 0 {
		return errors.New("at least one reference is required")
	}
	if c.Fix != nil {
		if err := c.Fix.compile(); err != nil {
			return fmt.Errorf("fix: %w", err)
		}
	}
	return nil
}

// unsafeArg matches characters that have no business in a fix command.
// Commands never go through a shell, but the admin batch is a script, so
// this is a second line of defence on top of quoting.
var unsafeArg = regexp.MustCompile("[$`;|&<>\\\n\r]")

func (f *Fix) compile() error {
	if len(f.Steps) == 0 {
		return errors.New("needs at least one step")
	}
	for i := range f.Steps {
		s := &f.Steps[i]
		if err := checkArgv(s.Apply, false); err != nil {
			return fmt.Errorf("step %d apply: %w", i, err)
		}
		if err := checkArgv(s.Undo, true); err != nil {
			return fmt.Errorf("step %d undo: %w", i, err)
		}
		usesPrevious := strings.Contains(strings.Join(s.Undo, " "), "{previous}")
		if usesPrevious && len(s.Read) == 0 {
			return fmt.Errorf("step %d: undo uses {previous} but there is no read", i)
		}
		if len(s.Read) > 0 {
			if err := checkArgv(s.Read, false); err != nil {
				return fmt.Errorf("step %d read: %w", i, err)
			}
		}
		if s.ReadMatch != "" {
			re, err := regexp.Compile(s.ReadMatch)
			if err != nil {
				return fmt.Errorf("step %d read_match: %w", i, err)
			}
			if re.NumSubexp() != 1 {
				return fmt.Errorf("step %d read_match needs exactly one group", i)
			}
			s.re = re
		}
		if len(s.UndoIfMissing) > 0 {
			if err := checkArgv(s.UndoIfMissing, false); err != nil {
				return fmt.Errorf("step %d undo_if_missing: %w", i, err)
			}
		}
	}
	return nil
}

func checkArgv(argv []string, allowPrevious bool) error {
	if len(argv) == 0 {
		return errors.New("is empty")
	}
	if !strings.HasPrefix(argv[0], "/") {
		return fmt.Errorf("command %q must be an absolute path", argv[0])
	}
	for _, a := range argv {
		if unsafeArg.MatchString(a) {
			return fmt.Errorf("argument %q has unsafe characters", a)
		}
		if !allowPrevious && strings.Contains(a, "{previous}") {
			return fmt.Errorf("argument %q cannot use {previous}", a)
		}
	}
	return nil
}

func validStatus(s Status) bool {
	switch s {
	case Pass, Fail, Warn, Unknown, Manual, Skipped:
		return true
	}
	return false
}

// Run executes the check's probe.
func (c Check) Run(ctx context.Context, env Env) Result {
	var res Result
	switch {
	case c.Probe.Manual:
		res = Result{Status: Manual}
	case c.Probe.Ask:
		res = c.ask(env)
	case c.Probe.Builtin != "":
		res = Builtins[c.Probe.Builtin](ctx, env)
	default:
		res = c.runCommand(ctx, env)
	}
	res.CheckID = c.ID
	if res.Status == Pass {
		res.Credit = 1
	}
	if res.Status == Fail {
		res.Credit = 0
	}
	if res.Status == Warn && res.Credit == 0 {
		res.Credit = 0.5
	}
	return res
}

func (c Check) ask(env Env) Result {
	if env.Ask == nil {
		return Result{Status: Manual}
	}
	switch env.Ask(c) {
	case Yes:
		return Result{Status: Pass}
	case No:
		return Result{Status: Fail}
	case Quit:
		return Result{Status: Skipped}
	default:
		return Result{Status: Manual}
	}
}

func (c Check) runCommand(ctx context.Context, env Env) Result {
	name, args := c.Probe.Command[0], c.Probe.Command[1:]
	out, err := env.Runner.Run(ctx, name, args...)
	evidence := Evidence(name, args, out, err)
	if err != nil {
		return Result{Status: Unknown, Current: err.Error(), Evidence: evidence}
	}
	text := out.Combined()
	for _, rule := range c.Probe.Rules {
		if rule.Exit != nil && *rule.Exit != out.ExitCode {
			continue
		}
		current := firstLine(text)
		if rule.re != nil {
			loc := rule.re.FindStringIndex(text)
			if loc == nil {
				continue
			}
			current = lineAround(text, loc[0])
			if len(current) > 100 {
				current = text[loc[0]:loc[1]]
			}
		}
		res := Result{Status: rule.Status, Current: current, Evidence: evidence}
		if rule.Status == Warn {
			res.Credit = 0.5
			if rule.Credit != nil {
				res.Credit = *rule.Credit
			}
		}
		return res
	}
	status := c.Probe.Default
	if status == "" {
		status = Unknown
	}
	res := Result{Status: status, Current: firstLine(text), Evidence: evidence}
	if status == Warn {
		res.Credit = 0.5
	}
	return res
}

// Evidence renders a command and its output like a terminal transcript.
func Evidence(name string, args []string, out runner.Output, err error) string {
	var b strings.Builder
	b.WriteString("$ " + runner.Key(name, args...) + "\n")
	if err != nil {
		b.WriteString("error: " + err.Error() + "\n")
		return b.String()
	}
	text := strings.TrimSpace(out.Combined())
	const max = 2000
	if len(text) > max {
		text = text[:max] + "\n[...]"
	}
	if text != "" {
		b.WriteString(text + "\n")
	}
	if out.ExitCode != 0 {
		fmt.Fprintf(&b, "[exit %d]\n", out.ExitCode)
	}
	return b.String()
}

// logPrefix matches the timestamp macOS tools prepend to stderr lines,
// e.g. "2026-11-06 15:47:44.510 sysadminctl[99242:21639140] ".
var logPrefix = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} [\d:.]+ \S+\[\d+:\d+\] `)

func clean(line string) string {
	return strings.TrimSpace(logPrefix.ReplaceAllString(strings.TrimSpace(line), ""))
}

func firstLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		if l := clean(line); l != "" {
			return l
		}
	}
	return ""
}

func lineAround(text string, offset int) string {
	start := strings.LastIndex(text[:offset], "\n") + 1
	end := strings.Index(text[offset:], "\n")
	if end < 0 {
		end = len(text)
	} else {
		end += offset
	}
	return clean(text[start:end])
}
