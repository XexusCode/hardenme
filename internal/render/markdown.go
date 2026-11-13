package render

import (
	"io"
	"regexp"
	"strings"

	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/scan"
)

// markdownWriter writes the same report as the text format, as Markdown
// that can be pasted into a message, an issue or a note.
type markdownWriter struct{ textWriter }

// renderMarkdown writes through a buffer so blank lines can be tidied:
// never more than one in a row.
func renderMarkdown(w io.Writer, r scan.Report, lang check.Lang) error {
	var buf strings.Builder
	m := &markdownWriter{textWriter{w: &buf, lang: lang}}
	m.report(r)
	out := strings.TrimSpace(blankLines.ReplaceAllString(buf.String(), "\n\n")) + "\n"
	_, err := io.WriteString(w, out)
	return err
}

var blankLines = regexp.MustCompile(`\n{3,}`)

func (m *markdownWriter) report(r scan.Report) {
	title := "hardenme · " + platformName(r.Platform)
	if r.OSVersion != "" {
		title += " " + r.OSVersion
	}
	m.printf("# %s\n\n", title)
	m.printf("**%s:** %s\n\n", msg(m.lang, "profile"), describeProfile(r.Profile, r.Platform, m.lang))

	groups := group(r.Entries)
	if score := r.Score(); score.Value < 0 {
		m.printf("**%s**\n\n", msg(m.lang, "no_score"))
	} else {
		m.printf("**%s: %d/100 (%s, %s)**\n\n", msg(m.lang, "score"), score.Value, score.Grade, msg(m.lang, "label_"+score.Grade))
	}
	m.printf("%s\n\n", summaryLine(groups, m.lang))
	m.printf("_%s_\n", msg(m.lang, "scoring_note"))

	risks, safe := splitStories(r.Stories)
	for _, part := range []struct {
		key     string
		stories []scan.Story
	}{{"sec_stories", risks}, {"sec_safe", safe}} {
		if len(part.stories) == 0 {
			continue
		}
		m.printf("\n## %s\n\n", msg(m.lang, storyKey(part.key, r.Platform)))
		for _, st := range part.stories {
			m.printf("- %s\n", mdEscape(storyText(st, m.lang)))
		}
	}

	for _, s := range sections {
		m.section(s.key, s.mark, groups[s.group], s.detailed)
	}
}

func (m *markdownWriter) section(key, mark string, entries []scan.Entry, detailed bool) {
	if len(entries) == 0 {
		return
	}
	m.printf("\n## %s (%d)\n\n", msg(m.lang, key), len(entries))
	for _, e := range entries {
		text := e.Check.TextFor(m.lang)
		title := mdEscape(entryTitle(e, m.lang))
		if !detailed {
			line := "- " + mark + " " + title
			if mark == "-" {
				line = "- " + title
			}
			if e.Result.Status == check.Unknown && e.Result.Current != "" {
				line += " (`" + scrub(e.Result.Current) + "`)"
			}
			m.printf("%s\n", line)
			continue
		}
		m.printf("### %s %s\n\n", mark, title)
		if e.Result.Current != "" && e.Result.Status != check.Manual {
			m.printf("**%s:** `%s`\n\n", msg(m.lang, "now"), scrub(e.Result.Current))
		}
		m.printf("**%s:** %s\n\n", msg(m.lang, "why"), mdEscape(strings.Join(strings.Fields(text.Why), " ")))
		m.printf("**%s:**\n\n", msg(m.lang, "how"))
		for i, step := range text.FixSteps {
			m.printf("%d. %s\n", i+1, mdEscape(step))
		}
		if text.DeepLink != "" {
			m.printf("\n%s: `open \"%s\"`\n", msg(m.lang, "open"), text.DeepLink)
		}
		if text.Command != "" {
			m.printf("\n%s: `%s`\n", msg(m.lang, "command"), text.Command)
		}
		m.printf("\n")
	}
}

// mdEscape keeps check text from turning into Markdown by accident.
func mdEscape(s string) string {
	r := strings.NewReplacer(`\`, `\\`, "*", `\*`, "_", `\_`, "`", "\\`", "[", `\[`, "]", `\]`, "<", `\<`)
	return r.Replace(s)
}
