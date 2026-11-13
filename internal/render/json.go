package render

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/scan"
)

// JSONSchemaVersion changes only when a field is removed or changes
// meaning. New fields can be added without changing it.
const JSONSchemaVersion = 1

// JSONReport is the --format json output.
type JSONReport struct {
	SchemaVersion int             `json:"schema_version"`
	Tool          string          `json:"tool"`
	Language      check.Lang      `json:"language"`
	Platform      string          `json:"platform"`
	OSVersion     string          `json:"os_version,omitempty"`
	GeneratedAt   time.Time       `json:"generated_at"`
	Profile       profile.Profile `json:"profile"`
	// Score is null when nothing could be scored.
	Score   *JSONScore  `json:"score"`
	Stories []JSONStory `json:"stories"`
	Entries []JSONEntry `json:"entries"`
}

// JSONStory is a plain-language risk.
type JSONStory struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	// Safe is true for a reassurance ("they could not...") and false for
	// a risk.
	Safe bool `json:"safe"`
}

// JSONScore is the weighted score.
type JSONScore struct {
	Value int    `json:"value"`
	Grade string `json:"grade"`
}

// JSONEntry is one check in the report.
type JSONEntry struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Status    check.Status    `json:"status"`
	Relevance check.Relevance `json:"relevance"`
	Current   string          `json:"current,omitempty"`
	FixKind   check.FixKind   `json:"fix_kind"`
}

// BuildJSON turns a report into its JSON form.
func BuildJSON(r scan.Report, lang check.Lang) JSONReport {
	out := JSONReport{
		SchemaVersion: JSONSchemaVersion,
		Tool:          "hardenme",
		Language:      lang,
		Platform:      r.Platform,
		OSVersion:     r.OSVersion,
		GeneratedAt:   r.StartedAt.UTC().Truncate(time.Second),
		Profile:       r.Profile,
		Stories:       []JSONStory{},
		Entries:       []JSONEntry{},
	}
	for _, st := range r.Stories {
		out.Stories = append(out.Stories, JSONStory{ID: st.ID, Text: storyText(st, lang), Safe: st.Safe})
	}
	if s := r.Score(); s.Value >= 0 {
		out.Score = &JSONScore{Value: s.Value, Grade: s.Grade}
	}
	for _, e := range r.Entries {
		out.Entries = append(out.Entries, JSONEntry{
			ID:        e.Check.ID,
			Title:     entryTitle(e, lang),
			Status:    e.Result.Status,
			Relevance: e.Relevance,
			Current:   scrub(e.Result.Current),
			FixKind:   e.Check.FixKind(),
		})
	}
	return out
}

func renderJSON(w io.Writer, r scan.Report, lang check.Lang) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(BuildJSON(r, lang))
}

// scrub replaces the home folder with ~ so a shared report does not carry
// the account name.
func scrub(s string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" || home == "/" {
		return s
	}
	return strings.ReplaceAll(s, home, "~")
}
