package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/XexusCode/hardenme/checks"
	"github.com/XexusCode/hardenme/internal/answers"
	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/render"
	"github.com/XexusCode/hardenme/internal/scan"
	"github.com/XexusCode/hardenme/internal/story"
)

// phonePlatforms are the devices covered by the guided checklist.
var phonePlatforms = []string{"ios", "android"}

// checklistDeps are the parts of `hardenme checklist` that tests replace.
type checklistDeps struct {
	catalog  fs.FS
	answers  func(platform string) (answers.Store, error)
	profile  func() (profile.Store, error)
	terminal func(io.Reader) bool
	now      func() time.Time
}

func defaultChecklistDeps() checklistDeps {
	return checklistDeps{
		catalog:  checks.FS,
		answers:  answers.DefaultStore,
		profile:  profile.PhoneStore,
		terminal: isTerminal,
		now:      time.Now,
	}
}

type checklistOptions struct {
	again, reprofile bool
	only, format     string
}

func newChecklist(opts *options) *cobra.Command {
	return newChecklistWith(opts, defaultChecklistDeps())
}

func newChecklistWith(opts *options, deps checklistDeps) *cobra.Command {
	co := &checklistOptions{}
	cmd := &cobra.Command{
		Use:       "checklist ios|android",
		Short:     "Check your phone's security by answering a few questions",
		Args:      cobra.ExactArgs(1),
		ValidArgs: phonePlatforms,
		RunE: func(cmd *cobra.Command, args []string) error {
			lang, err := check.ParseLang(opts.lang)
			if err != nil {
				return err
			}
			platform := strings.ToLower(args[0])
			if platform == "iphone" {
				platform = "ios"
			}
			if platform != "ios" && platform != "android" {
				return fmt.Errorf("unknown phone %q (use ios or android)", args[0])
			}
			f, err := render.ParseFormat(co.format)
			if err != nil {
				return err
			}
			all, err := check.Load(deps.catalog, platform)
			if err != nil {
				return err
			}
			selected, err := filterChecks(all, co.only)
			if err != nil {
				return err
			}
			if !deps.terminal(cmd.InOrStdin()) {
				return errors.New(t(lang, "checklist_needs_terminal"))
			}
			c := &checklister{
				lang: lang, platform: platform, deps: deps, opts: co,
				in: bufio.NewScanner(cmd.InOrStdin()), out: cmd.OutOrStdout(),
			}
			report, err := c.run(cmd, selected)
			if err != nil {
				return err
			}
			if err := render.Render(cmd.OutOrStdout(), report, f, lang); err != nil {
				return err
			}
			if code := report.ExitCode(); code != 0 {
				return ExitError{Code: code}
			}
			return nil
		},
	}
	fl := cmd.Flags()
	fl.BoolVar(&co.again, "again", false, "ask every question again, even the ones that were already right")
	fl.BoolVar(&co.reprofile, "profile", false, "answer the questions about how you use the phone again")
	fl.StringVar(&co.only, "only", "", "comma-separated check IDs to ask (e.g. stolen_device_protection,sim_pin)")
	fl.StringVar(&co.format, "format", "text", "output format: text, json or markdown")
	return cmd
}

type checklister struct {
	lang     check.Lang
	platform string
	deps     checklistDeps
	opts     *checklistOptions
	in       *bufio.Scanner
	out      io.Writer
}

func (c *checklister) say(key string, args ...any) {
	fmt.Fprintf(c.out, t(c.lang, key)+"\n", args...)
}

func (c *checklister) run(cmd *cobra.Command, selected []check.Check) (scan.Report, error) {
	c.say("checklist_intro_" + c.platform)
	p, err := c.phoneProfile()
	if err != nil {
		return scan.Report{}, err
	}
	store, err := c.deps.answers(c.platform)
	if err != nil {
		return scan.Report{}, err
	}
	saved, err := store.Load()
	if err != nil {
		return scan.Report{}, err
	}

	// Ask the most important questions first, so someone who stops halfway
	// has answered what matters most for them.
	selected = append([]check.Check(nil), selected...)
	sort.SliceStable(selected, func(i, j int) bool {
		return selected[i].RelevanceFor(p).Weight() > selected[j].RelevanceFor(p).Weight()
	})

	// Count the questions first so each one can say "3/12".
	total, reused := 0, 0
	for _, ch := range selected {
		if ch.RelevanceFor(p) == check.NotApplicable {
			continue
		}
		if a, ok := saved[ch.ID]; !c.opts.again && answers.Reusable(a, ok) {
			reused++
		} else {
			total++
		}
	}
	if reused > 0 {
		c.say("checklist_reusing")
	}

	n, quit := 0, false
	ask := func(ch check.Check) check.Answer {
		if a, ok := saved[ch.ID]; !c.opts.again && answers.Reusable(a, ok) {
			return check.Yes
		}
		if quit {
			return check.Quit
		}
		n++
		answer := c.ask(ch, n, total)
		if answer == check.Quit {
			quit = true
			return answer
		}
		saved[ch.ID] = answers.Saved{Answer: answer, Time: c.deps.now()}
		return answer
	}
	report := scan.ScanEnv(cmd.Context(), c.platform, selected, check.Env{Profile: p, Ask: ask})
	stories, err := story.Load(c.deps.catalog, c.platform)
	if err != nil {
		return scan.Report{}, err
	}
	story.Add(&report, stories)
	if err := store.Save(saved); err != nil {
		return scan.Report{}, err
	}
	fmt.Fprintln(c.out)
	if quit {
		c.say("checklist_quit", c.platform)
	}
	return report, nil
}

// phoneProfile loads the phone's profile, asking for it on the first run or
// with --profile.
func (c *checklister) phoneProfile() (profile.Profile, error) {
	store, err := c.deps.profile()
	if err != nil {
		return profile.Profile{}, err
	}
	p, err := store.Load()
	switch {
	case err == nil && !c.opts.reprofile:
		return p, nil
	case err == nil:
	case errors.Is(err, profile.ErrNoProfile):
		p = profile.DefaultPhone()
		c.say("checklist_first_run")
	default:
		return profile.Profile{}, err
	}
	p, err = askProfileWith(p, c.lang, c.in, c.out, true)
	if err != nil {
		return profile.Profile{}, err
	}
	if err := store.Save(p); err != nil {
		return profile.Profile{}, err
	}
	fmt.Fprintln(c.out)
	return p, nil
}

// ask shows one question until it gets an answer it understands. The end
// of the input counts as quitting.
func (c *checklister) ask(ch check.Check, n, total int) check.Answer {
	text := ch.TextFor(c.lang)
	fmt.Fprintf(c.out, "\n%d/%d %s\n", n, total, text.Question)
	if text.Where != "" {
		fmt.Fprintf(c.out, "    %s: %s\n", t(c.lang, "checklist_where"), text.Where)
	}
	for {
		fmt.Fprintf(c.out, "    %s ", t(c.lang, "checklist_prompt"))
		if !c.in.Scan() {
			fmt.Fprintln(c.out)
			return check.Quit
		}
		if a, ok := parseAnswer(c.in.Text()); ok {
			return a
		}
		c.say("checklist_ask_again")
	}
}

// parseAnswer understands yes, no, "I don't know" and quit, in EN and ES.
func parseAnswer(s string) (check.Answer, bool) {
	s = strings.Join(strings.Fields(strings.ToLower(s)), " ")
	switch s {
	case "s", "si", "sí", "y", "yes":
		return check.Yes, true
	case "n", "no":
		return check.No, true
	case "?", "ns", "no sé", "no se", "nose", "no lo sé", "no lo se", "d", "dk", "idk", "don't know", "dont know", "i don't know", "i dont know":
		return check.DontKnow, true
	case "q", "quit", "salir", "x", "exit":
		return check.Quit, true
	}
	return "", false
}
