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

const (
	Text     Format = "text"
	JSON     Format = "json"
	Markdown Format = "markdown"
)

// ParseFormat validates an output format.
func ParseFormat(s string) (Format, error) {
	switch Format(s) {
	case Text, JSON, Markdown:
		return Format(s), nil
	case "md":
		return Markdown, nil
	}
	return "", fmt.Errorf("unsupported format %q (use text, json or markdown)", s)
}

const width = 78

// Render writes the report in the given format and language.
func Render(w io.Writer, r scan.Report, f Format, lang check.Lang) error {
	switch f {
	case Text:
		t := &textWriter{w: w, lang: lang}
		t.report(r)
		return t.err
	case JSON:
		return renderJSON(w, r, lang)
	case Markdown:
		return renderMarkdown(w, r, lang)
	}
	return fmt.Errorf("unsupported format %q", f)
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
	t.printf("%s: %s\n\n", msg(t.lang, "profile"), describeProfile(r.Profile, r.Platform, t.lang))

	groups := group(r.Entries)
	score := r.Score()
	if score.Value < 0 {
		t.printf("%s\n", msg(t.lang, "no_score"))
	} else {
		t.printf("%s: %d/100 (%s, %s)\n", msg(t.lang, "score"), score.Value, score.Grade, msg(t.lang, "label_"+score.Grade))
	}
	t.printf("%s\n", summaryLine(groups, t.lang))
	t.printf("%s\n", wrap(msg(t.lang, "scoring_note"), width, ""))

	if len(r.Stories) > 0 {
		t.printf("\n%s\n", msg(t.lang, "sec_stories"))
		for _, st := range r.Stories {
			t.printf("→ %s\n", strings.TrimPrefix(wrap(storyText(st, t.lang), width, "  "), "  "))
		}
	}

	for _, s := range sections {
		t.section(s.key, s.mark, groups[s.group], s.detailed)
	}

	if isPhone(r.Platform) {
		t.printf("\n%s\n", wrap(fmt.Sprintf(msg(t.lang, "next_phone"), r.Platform), width, ""))
		t.printf("%s\n", fmt.Sprintf(msg(t.lang, "edit_profile_phone"), r.Platform))
		return
	}
	t.printf("\n%s\n", msg(t.lang, "edit_profile"))
}

func isPhone(platform string) bool { return platform == "ios" || platform == "android" }

func storyText(s scan.Story, lang check.Lang) string {
	if text, ok := s.Text[lang]; ok {
		return text
	}
	return s.Text[check.EN]
}

// summaryLine is "2 to fix, 1 suggestion, 1 to check yourself".
func summaryLine(groups map[string][]scan.Entry, lang check.Lang) string {
	toFix := len(groups[string(check.Essential)]) + len(groups[string(check.Recommended)])
	suggestions := len(groups[string(check.Optional)])
	return fmt.Sprintf(msg(lang, "summary"), toFix, suggestions, plural(lang, "suggestion", suggestions), len(groups["manual"]))
}

// sections are the report sections in order: key, mark, group and whether
// each entry is shown with its explanation.
var sections = []struct {
	key, mark, group string
	detailed         bool
}{
	{"sec_essential", "✗", string(check.Essential), true},
	{"sec_recommended", "✗", string(check.Recommended), true},
	{"sec_optional", "·", string(check.Optional), true},
	{"sec_manual", "?", "manual", true},
	{"sec_unknown", "!", "unknown", false},
	{"sec_pass", "✓", "pass", false},
	{"sec_na", "-", "na", false},
}

// entryTitle says what was found: the pass or fix title, or the neutral
// name when the setting was not read or does not apply.
func entryTitle(e scan.Entry, lang check.Lang) string {
	text := e.Check.TextFor(lang)
	switch {
	case e.Relevance == check.NotApplicable:
		return text.Title
	case e.Result.Status == check.Pass:
		return text.TitlePass
	case e.Result.Status == check.Unknown, e.Result.Status == check.Skipped:
		return text.Title
	}
	return text.TitleFail
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
		title := entryTitle(e, t.lang)
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

func describeProfile(p profile.Profile, platform string, lang check.Lang) string {
	parts := []string{msg(lang, "desktop")}
	switch {
	case isPhone(platform):
		parts[0] = msg(lang, "phone")
	case p.Portable:
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
