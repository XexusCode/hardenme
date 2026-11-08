package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/XexusCode/hardenme/checks"
	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/fix"
	"github.com/XexusCode/hardenme/internal/runner"
	"github.com/XexusCode/hardenme/internal/scan"
)

// fixDeps are the parts of `hardenme fix` that tests replace.
type fixDeps struct {
	probe    runner.Runner // reads state; short timeout
	apply    runner.Runner // applies changes; waits for the password dialog
	log      func() (fix.Log, error)
	terminal func(io.Reader) bool
	now      func() time.Time
}

func defaultFixDeps() fixDeps {
	return fixDeps{
		probe:    runner.Exec{},
		apply:    runner.Exec{Timeout: 5 * time.Minute},
		log:      fix.DefaultLog,
		terminal: isTerminal,
		now:      time.Now,
	}
}

type fixOptions struct {
	only   string
	dryRun bool
	yes    bool
	undo   bool
}

func newFix(opts *options) *cobra.Command {
	return newFixWith(opts, defaultFixDeps())
}

func newFixWith(opts *options, deps fixDeps) *cobra.Command {
	fo := &fixOptions{}
	cmd := &cobra.Command{
		Use:   "fix",
		Short: "Pick the improvements you want and let hardenme apply them",
		RunE: func(cmd *cobra.Command, _ []string) error {
			lang, err := check.ParseLang(opts.lang)
			if err != nil {
				return err
			}
			platform, err := currentPlatform()
			if err != nil {
				return err
			}
			all, err := check.Load(checks.FS, platform)
			if err != nil {
				return err
			}
			f := &fixer{
				ctx: cmd.Context(), lang: lang, all: all, deps: deps, opts: fo,
				in: bufio.NewScanner(cmd.InOrStdin()), rawIn: cmd.InOrStdin(),
				out: cmd.OutOrStdout(), errOut: cmd.ErrOrStderr(),
			}
			if fo.undo {
				return f.undo()
			}
			return f.run(platform)
		},
	}
	fl := cmd.Flags()
	fl.StringVar(&fo.only, "only", "", "comma-separated check IDs to consider (e.g. firewall,auto_lock)")
	fl.BoolVar(&fo.dryRun, "dry-run", false, "show what would change, change nothing")
	fl.BoolVar(&fo.yes, "yes", false, "apply every automatic fix without asking (macOS still asks for your password)")
	fl.BoolVar(&fo.undo, "undo", false, "undo the last batch of changes made by hardenme fix")
	return cmd
}

type fixer struct {
	ctx    context.Context
	lang   check.Lang
	all    []check.Check
	deps   fixDeps
	opts   *fixOptions
	in     *bufio.Scanner
	rawIn  io.Reader
	out    io.Writer
	errOut io.Writer
}

func (f *fixer) say(key string, args ...any) {
	fmt.Fprintf(f.out, t(f.lang, key)+"\n", args...)
}

func (f *fixer) run(platform string) error {
	p, err := loadOrCreateProfile(f.ctx, f.deps.probe, f.lang, f.rawIn, f.errOut)
	if err != nil {
		return err
	}
	report := scan.Scan(f.ctx, platform, f.all, p, f.deps.probe)
	items := fix.Candidates(report)
	if f.opts.only != "" {
		wanted, err := filterChecks(f.all, f.opts.only)
		if err != nil {
			return err
		}
		items = keepIDs(items, wanted)
	}
	before := report.Score()
	f.say("fix_score_now", before.Value, before.Grade)
	if len(items) == 0 {
		f.say("fix_nothing")
		return nil
	}
	f.printMenu(items)

	if f.opts.dryRun {
		f.dryRun(items)
		return nil
	}

	var selected []fix.Item
	switch {
	case f.opts.yes:
		for _, it := range items {
			if it.Kind == check.FixAuto {
				selected = append(selected, it)
			}
		}
	case !f.deps.terminal(f.rawIn):
		return errors.New(t(f.lang, "fix_needs_terminal"))
	default:
		selected = f.askSelection(items)
	}
	if len(selected) == 0 {
		f.say("fix_none_selected")
		return nil
	}

	var auto, guided []fix.Item
	for _, it := range selected {
		if it.Kind == check.FixAuto {
			auto = append(auto, it)
		} else {
			guided = append(guided, it)
		}
	}

	outcomes := map[string]fix.Outcome{}
	logged := false
	env := check.Env{Runner: f.deps.probe, Profile: p}
	if len(auto) > 0 {
		if !f.confirmAuto(auto) {
			f.say("fix_cancelled")
			return ExitError{Code: 1}
		}
		var cs []check.Check
		var ids []string
		for _, it := range auto {
			cs = append(cs, it.Entry.Check)
			ids = append(ids, it.Entry.Check.ID)
		}
		steps, skips := fix.Prepare(f.ctx, f.deps.probe, cs)
		for _, s := range skips {
			f.say("fix_step_skipped", f.title(s.CheckID), s.Reason)
		}
		if len(steps) > 0 {
			f.say("fix_applying")
		}
		res := fix.Run(f.ctx, f.deps.apply, steps)
		switch {
		case res.Cancelled:
			f.say("fix_dialog_cancelled")
		case res.AdminErr != nil:
			f.say("fix_admin_error", res.AdminErr)
		}
		if changes := res.Changes(); len(changes) > 0 {
			l, err := f.deps.log()
			if err == nil {
				err = l.Append(fix.Batch{Time: f.deps.now().UTC(), Changes: changes})
			}
			if err != nil {
				f.say("fix_log_error", err)
			} else {
				logged = true
			}
		}
		for id, o := range fix.Recheck(f.ctx, env, &report, res, ids) {
			outcomes[id] = o
		}
	}

	for _, it := range guided {
		outcomes[it.Entry.Check.ID] = f.guide(it, env, &report)
	}

	fmt.Fprintln(f.out)
	f.say("fix_results")
	allFixed := true
	for _, it := range selected {
		id := it.Entry.Check.ID
		text := it.Entry.Check.TextFor(f.lang)
		switch outcomes[id] {
		case fix.Fixed:
			fmt.Fprintf(f.out, "  ✓ %s\n", text.TitlePass)
		case fix.Cancelled:
			allFixed = false
			fmt.Fprintf(f.out, "  - %s: %s\n", text.TitleFail, t(f.lang, "fix_outcome_cancelled"))
		case fix.Skipped:
			allFixed = false
			fmt.Fprintf(f.out, "  - %s: %s\n", text.TitleFail, t(f.lang, "fix_outcome_skipped"))
		default:
			allFixed = false
			fmt.Fprintf(f.out, "  ✗ %s: %s\n", text.TitleFail, t(f.lang, "fix_outcome_not_changed"))
		}
	}
	after := report.Score()
	fmt.Fprintln(f.out)
	f.say("fix_score_before_after", before.Value, before.Grade, after.Value, after.Grade)
	if logged {
		f.say("fix_undo_hint")
	}
	if !allFixed {
		return ExitError{Code: 1}
	}
	return nil
}

func keepIDs(items []fix.Item, wanted []check.Check) []fix.Item {
	keep := map[string]bool{}
	for _, c := range wanted {
		keep[c.ID] = true
	}
	var out []fix.Item
	for _, it := range items {
		if keep[it.Entry.Check.ID] {
			out = append(out, it)
		}
	}
	return out
}

func (f *fixer) title(id string) string {
	for _, c := range f.all {
		if c.ID == id {
			return c.TextFor(f.lang).TitleFail
		}
	}
	return id
}

func (f *fixer) printMenu(items []fix.Item) {
	fmt.Fprintln(f.out)
	f.say("fix_menu_intro")
	for i, it := range items {
		label := t(f.lang, "fix_label_auto")
		if it.Kind == check.FixGuided {
			label = t(f.lang, "fix_label_guided")
		}
		fmt.Fprintf(f.out, "  %2d) %s  [%s]\n", i+1, it.Entry.Check.TextFor(f.lang).TitleFail, label)
	}
	fmt.Fprintln(f.out)
}

func (f *fixer) dryRun(items []fix.Item) {
	f.say("fix_dry_run")
	for i, it := range items {
		text := it.Entry.Check.TextFor(f.lang)
		fmt.Fprintf(f.out, "\n%d) %s\n", i+1, text.TitleFail)
		if it.Kind == check.FixGuided {
			fmt.Fprintf(f.out, "   open %s\n", text.DeepLink)
			continue
		}
		steps, skips := fix.Prepare(f.ctx, f.deps.probe, []check.Check{it.Entry.Check})
		for _, s := range steps {
			prefix := "  "
			if s.Admin {
				prefix = t(f.lang, "fix_admin_prefix")
			}
			fmt.Fprintf(f.out, "   %s %s\n", prefix, strings.Join(s.Apply, " "))
			fmt.Fprintf(f.out, "      %s %s\n", t(f.lang, "fix_undo_prefix"), strings.Join(s.Undo, " "))
		}
		for _, s := range skips {
			fmt.Fprintf(f.out, "   - %s\n", s.Reason)
		}
	}
}

// askSelection reads numbers ("1,3", "1 3", "2-4"), "all"/"todo", or a
// blank line to exit. Invalid input is asked again.
func (f *fixer) askSelection(items []fix.Item) []fix.Item {
	for {
		fmt.Fprint(f.out, t(f.lang, "fix_ask")+" ")
		if !f.in.Scan() {
			return nil
		}
		picked, ok := parseSelection(f.in.Text(), len(items))
		if !ok {
			f.say("fix_ask_again", len(items))
			continue
		}
		out := make([]fix.Item, 0, len(picked))
		for _, n := range picked {
			out = append(out, items[n-1])
		}
		return out
	}
}

func parseSelection(s string, max int) ([]int, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "":
		return nil, true
	case "all", "a", "todo", "todos", "t":
		out := make([]int, max)
		for i := range out {
			out[i] = i + 1
		}
		return out, true
	}
	seen := map[int]bool{}
	var out []int
	for _, field := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == ';' }) {
		lo, hi := field, field
		if a, b, ok := strings.Cut(field, "-"); ok {
			lo, hi = a, b
		}
		from, err1 := strconv.Atoi(lo)
		to, err2 := strconv.Atoi(hi)
		if err1 != nil || err2 != nil || from < 1 || to > max || from > to {
			return nil, false
		}
		for n := from; n <= to; n++ {
			if !seen[n] {
				seen[n] = true
				out = append(out, n)
			}
		}
	}
	return out, len(out) > 0
}

func (f *fixer) confirmAuto(items []fix.Item) bool {
	fmt.Fprintln(f.out)
	f.say("fix_will_change")
	admin := false
	for _, it := range items {
		text := it.Entry.Check.TextFor(f.lang)
		fmt.Fprintf(f.out, "  • %s\n", text.TitleFail)
		if text.FixWarning != "" {
			fmt.Fprintf(f.out, "    %s %s\n", t(f.lang, "fix_warning_prefix"), text.FixWarning)
		}
		for _, s := range it.Entry.Check.Fix.Steps {
			admin = admin || s.Admin
		}
	}
	if admin {
		f.say("fix_password_note")
	}
	if f.opts.yes {
		return true
	}
	fmt.Fprintf(f.out, "%s %s ", t(f.lang, "fix_confirm"), t(f.lang, "yes_no_default_yes"))
	if !f.in.Scan() {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(f.in.Text())) {
	case "", "y", "yes", "s", "si", "sí":
		return true
	}
	return false
}

// guide opens the setting, waits for the user and checks again.
func (f *fixer) guide(it fix.Item, env check.Env, report *scan.Report) fix.Outcome {
	text := it.Entry.Check.TextFor(f.lang)
	fmt.Fprintf(f.out, "\n%s\n", text.TitleFail)
	for i, step := range text.FixSteps {
		fmt.Fprintf(f.out, "  %d. %s\n", i+1, step)
	}
	if strings.HasPrefix(text.DeepLink, "x-apple.systempreferences:") {
		_, _ = f.deps.probe.Run(f.ctx, "/usr/bin/open", text.DeepLink)
		f.say("fix_opened_settings")
	}
	fmt.Fprint(f.out, t(f.lang, "fix_guided_wait")+" ")
	if !f.in.Scan() || strings.EqualFold(strings.TrimSpace(f.in.Text()), "s") {
		return fix.Skipped
	}
	for i := range report.Entries {
		e := &report.Entries[i]
		if e.Check.ID == it.Entry.Check.ID {
			e.Result = e.Check.Run(f.ctx, env)
			if e.Result.Status == check.Pass {
				return fix.Fixed
			}
			if e.Result.Status == check.Manual {
				return fix.Skipped
			}
		}
	}
	return fix.NotChanged
}

func (f *fixer) undo() error {
	l, err := f.deps.log()
	if err != nil {
		return err
	}
	batch, ok, err := l.Last()
	if err != nil {
		return err
	}
	if !ok {
		f.say("fix_undo_nothing")
		return nil
	}
	steps, err := fix.UndoSteps(batch, f.all)
	if err != nil {
		return err
	}
	f.say("fix_undo_intro", batch.Time.Local().Format("2006-01-02 15:04"))
	seen := map[string]bool{}
	for _, s := range steps {
		if !seen[s.CheckID] {
			seen[s.CheckID] = true
			fmt.Fprintf(f.out, "  • %s\n", f.title(s.CheckID))
		}
	}
	if !f.opts.yes {
		if !f.deps.terminal(f.rawIn) {
			return errors.New(t(f.lang, "fix_needs_terminal"))
		}
		fmt.Fprintf(f.out, "%s %s ", t(f.lang, "fix_confirm_undo"), t(f.lang, "yes_no_default_yes"))
		if !f.in.Scan() {
			return ExitError{Code: 1}
		}
		switch strings.ToLower(strings.TrimSpace(f.in.Text())) {
		case "", "y", "yes", "s", "si", "sí":
		default:
			f.say("fix_cancelled")
			return ExitError{Code: 1}
		}
	}
	res := fix.Run(f.ctx, f.deps.apply, steps)
	if res.Cancelled {
		f.say("fix_dialog_cancelled")
		return ExitError{Code: 1}
	}
	if res.AdminErr != nil {
		f.say("fix_admin_error", res.AdminErr)
		return ExitError{Code: 1}
	}
	for _, s := range res.Steps {
		if !s.Ran || s.ExitCode != 0 {
			f.say("fix_undo_failed", f.title(s.CheckID))
			return ExitError{Code: 1}
		}
	}
	if err := l.DropLast(); err != nil {
		return err
	}
	f.say("fix_undo_done")
	return nil
}
