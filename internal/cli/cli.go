// Package cli wires the hardenme commands.
package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/XexusCode/hardenme/checks"
	"github.com/XexusCode/hardenme/docs/guide"
	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/fix"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/render"
	"github.com/XexusCode/hardenme/internal/runner"
	"github.com/XexusCode/hardenme/internal/scan"
)

// ExitError carries a process exit code without printing an error.
type ExitError struct{ Code int }

func (e ExitError) Error() string { return fmt.Sprintf("exit %d", e.Code) }

// Execute runs the CLI and returns the process exit code.
func Execute(version string) int {
	root := newRoot()
	root.Version = version
	err := root.Execute()
	var exit ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &exit):
		return exit.Code
	default:
		fmt.Fprintln(os.Stderr, "error:", err)
		return 3
	}
}

type options struct {
	lang string
}

func newRoot() *cobra.Command {
	opts := &options{}
	root := &cobra.Command{
		Use:           "hardenme",
		Short:         "Check your device's security and learn how to improve it",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringVar(&opts.lang, "lang", defaultLang(), "language: en or es")
	root.AddCommand(newScan(opts), newFix(opts), newChecklist(opts), newProfile(opts), newGuide(opts))
	return root
}

// defaultLang picks Spanish when the system language is Spanish. It reads
// the locale variables first and falls back to the macOS language list,
// because apps opened from Finder often start without LANG.
func defaultLang() string {
	return pickLang(os.Getenv, systemLanguages)
}

func pickLang(getenv func(string) string, system func() string) string {
	for _, v := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if val := getenv(v); val != "" && val != "C" && val != "POSIX" {
			return langOf(val)
		}
	}
	return langOf(system())
}

func langOf(val string) string {
	if strings.HasPrefix(strings.ToLower(val), "es") {
		return "es"
	}
	return "en"
}

// systemLanguages returns the first preferred language on macOS, e.g.
// "es-ES" from `defaults read -g AppleLanguages`, or "" elsewhere.
func systemLanguages() string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := (runner.Exec{}).Run(ctx, "defaults", "read", "-g", "AppleLanguages")
	if err != nil || out.ExitCode != 0 {
		return ""
	}
	return firstAppleLanguage(out.Stdout)
}

// firstAppleLanguage parses the plist array printed by defaults:
// ( "es-ES", en ) -> "es-ES".
func firstAppleLanguage(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "(")
	first, _, _ := strings.Cut(s, ",")
	return strings.Trim(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(first), ")")), `"`)
}

func currentPlatform() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "macos", nil
	}
	return "", fmt.Errorf("%s is not supported yet: hardenme currently audits macOS (Android and Windows are coming)", runtime.GOOS)
}

func newScan(opts *options) *cobra.Command {
	var format, only string
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Audit this device and print a weighted security score",
		RunE: func(cmd *cobra.Command, _ []string) error {
			lang, err := check.ParseLang(opts.lang)
			if err != nil {
				return err
			}
			f, err := render.ParseFormat(format)
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
			selected, err := filterChecks(all, only)
			if err != nil {
				return err
			}
			ctx := cmd.Context()
			r := runner.Exec{}
			p, err := loadOrCreateProfile(ctx, r, lang, cmd.InOrStdin(), cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			report := scan.Scan(ctx, platform, selected, p, r)
			report.OSVersion = osVersion(ctx, r)
			if err := render.Render(cmd.OutOrStdout(), report, f, lang); err != nil {
				return err
			}
			if f == render.Text && isTerminal(cmd.InOrStdin()) && os.Getenv("HARDENME_NO_FIX_HINT") == "" && len(fix.Candidates(report)) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), t(lang, "scan_fix_hint"))
			}
			if code := report.ExitCode(); code != 0 {
				return ExitError{Code: code}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "text", "output format: text, json or markdown")
	cmd.Flags().StringVar(&only, "only", "", "comma-separated check IDs to run (e.g. filevault,firewall)")
	return cmd
}

// filterChecks keeps the requested IDs. The platform prefix is optional:
// "filevault" matches "macos.filevault".
func filterChecks(all []check.Check, only string) ([]check.Check, error) {
	if strings.TrimSpace(only) == "" {
		return all, nil
	}
	var out []check.Check
	for _, id := range strings.Split(only, ",") {
		id = strings.TrimSpace(id)
		found := false
		for _, c := range all {
			if c.ID == id || strings.HasSuffix(c.ID, "."+id) {
				out = append(out, c)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown check %q", id)
		}
	}
	return out, nil
}

func osVersion(ctx context.Context, r runner.Runner) string {
	out, err := r.Run(ctx, "sw_vers", "-productVersion")
	if err != nil || out.ExitCode != 0 {
		return ""
	}
	return strings.TrimSpace(out.Stdout)
}

// loadOrCreateProfile returns the saved profile. On first run it asks the
// questions when a person is at the terminal, or uses safe defaults.
func loadOrCreateProfile(ctx context.Context, r runner.Runner, lang check.Lang, in io.Reader, errOut io.Writer) (profile.Profile, error) {
	store, err := profile.DefaultStore()
	if err != nil {
		return profile.Profile{}, err
	}
	p, err := store.Load()
	if err == nil {
		return p, nil
	}
	if !errors.Is(err, profile.ErrNoProfile) {
		return profile.Profile{}, err
	}
	portable, _ := profile.DetectPortable(ctx, r, runtime.GOOS)
	p = profile.Default(portable)
	if !isTerminal(in) {
		fmt.Fprintln(errOut, t(lang, "no_profile_default"))
		return p, nil
	}
	fmt.Fprintln(errOut, t(lang, "first_run"))
	p, err = askProfile(p, lang, in, errOut)
	if err != nil {
		return profile.Profile{}, err
	}
	if err := store.Save(p); err != nil {
		return profile.Profile{}, err
	}
	fmt.Fprintf(errOut, t(lang, "saved")+"\n\n", store.Path)
	return p, nil
}

func isTerminal(in io.Reader) bool {
	f, ok := in.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

func newProfile(opts *options) *cobra.Command {
	var portable, leavesHome, shared, sensitive bool
	var risk string
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Answer a few questions so the score fits how you use this device",
		RunE: func(cmd *cobra.Command, _ []string) error {
			lang, err := check.ParseLang(opts.lang)
			if err != nil {
				return err
			}
			store, err := profile.DefaultStore()
			if err != nil {
				return err
			}
			p, err := store.Load()
			if errors.Is(err, profile.ErrNoProfile) {
				detected, _ := profile.DetectPortable(cmd.Context(), runner.Exec{}, runtime.GOOS)
				p, err = profile.Default(detected), nil
			}
			if err != nil {
				return err
			}
			flags := cmd.Flags()
			if flags.NFlag() == 0 || (flags.NFlag() == 1 && flags.Changed("lang")) {
				p, err = askProfile(p, lang, cmd.InOrStdin(), cmd.ErrOrStderr())
				if err != nil {
					return err
				}
			} else {
				if flags.Changed("portable") {
					p.Portable = portable
				}
				if flags.Changed("leaves-home") {
					p.LeavesHome = leavesHome
				}
				if flags.Changed("shared") {
					p.Shared = shared
				}
				if flags.Changed("sensitive-data") {
					p.SensitiveData = sensitive
				}
				if flags.Changed("risk") {
					if p.Risk, err = profile.ParseRisk(risk); err != nil {
						return err
					}
				}
			}
			if err := store.Save(p); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), t(lang, "saved")+"\n", store.Path)
			return nil
		},
	}
	f := cmd.Flags()
	f.BoolVar(&portable, "portable", false, "this device is a laptop or tablet")
	f.BoolVar(&leavesHome, "leaves-home", false, "you take this device outside your home")
	f.BoolVar(&shared, "shared", false, "other people use this device")
	f.BoolVar(&sensitive, "sensitive-data", false, "it holds work or banking data")
	f.StringVar(&risk, "risk", "normal", "risk level: normal or elevated")
	return cmd
}

// askProfile asks the profile questions, using p's values as defaults.
func askProfile(p profile.Profile, lang check.Lang, in io.Reader, out io.Writer) (profile.Profile, error) {
	return askProfileWith(p, lang, bufio.NewScanner(in), out, false)
}

// askProfileWith asks the profile questions from sc. A phone is always
// portable, so the phone version skips that question.
func askProfileWith(p profile.Profile, lang check.Lang, sc *bufio.Scanner, out io.Writer, phone bool) (profile.Profile, error) {
	ask := func(key string, def bool) (bool, error) {
		hint := t(lang, "yes_no_default_no")
		if def {
			hint = t(lang, "yes_no_default_yes")
		}
		for {
			fmt.Fprintf(out, "%s %s ", t(lang, key), hint)
			if !sc.Scan() {
				if err := sc.Err(); err != nil {
					return false, err
				}
				return def, nil
			}
			switch strings.ToLower(strings.TrimSpace(sc.Text())) {
			case "":
				return def, nil
			case "y", "yes", "s", "si", "sí":
				return true, nil
			case "n", "no":
				return false, nil
			}
		}
	}
	var err error
	if phone {
		p.Portable = true
	} else if p.Portable, err = ask("q_portable", p.Portable); err != nil {
		return p, err
	}
	if p.LeavesHome, err = ask("q_leaves_home", p.LeavesHome || p.Portable); err != nil {
		return p, err
	}
	if p.Shared, err = ask("q_shared", p.Shared); err != nil {
		return p, err
	}
	if p.SensitiveData, err = ask("q_sensitive", p.SensitiveData); err != nil {
		return p, err
	}
	elevated, err := ask("q_risk", p.Risk == profile.RiskElevated)
	if err != nil {
		return p, err
	}
	p.Risk = profile.RiskNormal
	if elevated {
		p.Risk = profile.RiskElevated
	}
	return p, nil
}

func newGuide(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:       "guide [page]",
		Short:     "Read the guide: what to do before and after a device theft",
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: guide.Pages(),
		RunE: func(cmd *cobra.Command, args []string) error {
			lang, err := check.ParseLang(opts.lang)
			if err != nil {
				return err
			}
			if len(args) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), t(lang, "guide_pages"))
				for _, page := range guide.Pages() {
					fmt.Fprintf(cmd.OutOrStdout(), "  hardenme guide %s\n", page)
				}
				return nil
			}
			text, err := guide.Page(string(lang), args[0])
			if err != nil {
				return err
			}
			_, err = io.WriteString(cmd.OutOrStdout(), text)
			return err
		},
	}
}
