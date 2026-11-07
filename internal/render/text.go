// Package render turns a scan report into something a person can read.
package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/scan"
)

// Format is an output format.
type Format string

const Text Format = "text"

// ParseFormat validates an output format.
func ParseFormat(s string) (Format, error) {
	if Format(s) == Text {
		return Text, nil
	}
	return "", fmt.Errorf("unsupported format %q (only text for now)", s)
}

const width = 78

// Render writes the report in the given format and language.
func Render(w io.Writer, r scan.Report, f Format, lang check.Lang) error {
	if f != Text {
		return fmt.Errorf("unsupported format %q", f)
	}
	t := &textWriter{w: w, lang: lang}
	t.report(r)
	return t.err
}

type textWriter struct {
	w    io.Writer
	lang check.Lang
	err  error
}

func (t *textWriter) printf(format string, args ...any) {
	if t.err == nil {
		_, t.err = fmt.Fprintf(t.w, format, args...)
	}
}

func (t *textWriter) report(r scan.Report) {
	title := "hardenme · " + platformName(r.Platform)
	if r.OSVersion != "" {
		title += " " + r.OSVersion
	}
	t.printf("%s\n", title)
	t.printf("%s: %s\n\n", msg(t.lang, "profile"), describeProfile(r.Profile, t.lang))

	groups := group(r.Entries)
	score := r.Score()
	if score.Value < 0 {
		t.printf("%s\n", msg(t.lang, "no_score"))
	} else {
		t.printf("%s: %d/100 (%s, %s)\n", msg(t.lang, "score"), score.Value, score.Grade, msg(t.lang, "label_"+score.Grade))
	}
	toFix := len(groups[string(check.Essential)]) + len(groups[string(check.Recommended)])
	t.printf(msg(t.lang, "summary")+"\n", toFix, len(groups[string(check.Optional)]), len(groups["manual"]))
	t.printf("%s\n", wrap(msg(t.lang, "scoring_note"), width, ""))

	t.section("sec_essential", "✗", groups[string(check.Essential)], true)
	t.section("sec_recommended", "✗", groups[string(check.Recommended)], true)
	t.section("sec_optional", "·", groups[string(check.Optional)], true)
	t.section("sec_manual", "?", groups["manual"], true)
	t.section("sec_unknown", "!", groups["unknown"], false)
	t.section("sec_pass", "✓", groups["pass"], false)
	t.section("sec_na", "-", groups["na"], false)

	t.printf("\n%s\n", msg(t.lang, "edit_profile"))
}

// group sorts entries into the report sections. Problems are grouped by
// relevance; everything else by status.
func group(entries []scan.Entry) map[string][]scan.Entry {
	g := map[string][]scan.Entry{}
	for _, e := range entries {
		var key string
		switch {
		case e.Relevance == check.NotApplicable:
			key = "na"
		case e.Result.Status == check.Pass:
			key = "pass"
		case e.Result.Status == check.Manual:
			key = "manual"
		case e.Result.Status == check.Unknown, e.Result.Status == check.Skipped:
			key = "unknown"
		default: // fail or warn
			key = string(e.Relevance)
		}
		g[key] = append(g[key], e)
	}
	return g
}

func (t *textWriter) section(key, mark string, entries []scan.Entry, detailed bool) {
	if len(entries) == 0 {
		return
	}
	t.printf("\n%s (%d)\n", msg(t.lang, key), len(entries))
	for _, e := range entries {
		text := e.Check.TextFor(t.lang)
		title := text.TitleFail
		if e.Result.Status == check.Pass {
			title = text.TitlePass
		}
		if !detailed {
			line := mark + " " + title
			if e.Result.Status == check.Unknown && e.Result.Current != "" {
				line += " (" + e.Result.Current + ")"
			}
			t.printf("%s\n", line)
			continue
		}
		t.printf("\n%s %s\n", mark, title)
		if e.Result.Current != "" && e.Result.Status != check.Manual {
			t.printf("  %s: %s\n", msg(t.lang, "now"), e.Result.Current)
		}
		t.printf("  %s:\n%s\n", msg(t.lang, "why"), wrap(text.Why, width, "    "))
		t.printf("  %s:\n", msg(t.lang, "how"))
		for i, step := range text.FixSteps {
			prefix := fmt.Sprintf("    %d. ", i+1)
			t.printf("%s%s\n", prefix, strings.TrimPrefix(wrap(step, width, strings.Repeat(" ", len(prefix))), strings.Repeat(" ", len(prefix))))
		}
		if text.DeepLink != "" {
			t.printf("  %s: open \"%s\"\n", msg(t.lang, "open"), text.DeepLink)
		}
		if text.Command != "" {
			t.printf("  %s: %s\n", msg(t.lang, "command"), text.Command)
		}
	}
}

func platformName(p string) string {
	switch p {
	case "macos":
		return "macOS"
	case "windows":
		return "Windows"
	case "android":
		return "Android"
	case "ios":
		return "iOS"
	}
	return p
}

func describeProfile(p profile.Profile, lang check.Lang) string {
	parts := []string{msg(lang, "desktop")}
	if p.Portable {
		parts[0] = msg(lang, "laptop")
	}
	if p.LeavesHome {
		parts = append(parts, msg(lang, "leaves_home"))
	} else {
		parts = append(parts, msg(lang, "stays_home"))
	}
	if p.Shared {
		parts = append(parts, msg(lang, "shared"))
	}
	if p.SensitiveData {
		parts = append(parts, msg(lang, "sensitive"))
	}
	if p.Risk == profile.RiskElevated {
		parts = append(parts, msg(lang, "risk_elevated"))
	} else {
		parts = append(parts, msg(lang, "risk_normal"))
	}
	return strings.Join(parts, ", ")
}

// wrap breaks text into lines of at most width runes, each starting with indent.
func wrap(text string, width int, indent string) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return indent
	}
	var lines []string
	line := indent + words[0]
	for _, word := range words[1:] {
		if len([]rune(line))+1+len([]rune(word)) > width {
			lines = append(lines, line)
			line = indent + word
			continue
		}
		line += " " + word
	}
	return strings.Join(append(lines, line), "\n")
}
