// Package story joins failing checks into plain-language risks ("theft
// stories") shown at the top of the report.
package story

import (
	"errors"
	"fmt"
	"io/fs"

	"gopkg.in/yaml.v3"

	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/scan"
)

// Story is one risk and the checks that make it true.
type Story struct {
	ID string `yaml:"id"`
	// All must fail or warn.
	All []string `yaml:"all"`
	// When set, at least one of Any must fail or warn too.
	Any []string `yaml:"any"`
	// Pass makes this a reassuring story ("they could not..."): it shows
	// when every check listed passes. It cannot be mixed with All or Any.
	Pass []string              `yaml:"pass"`
	Text map[check.Lang]string `yaml:"text"`
}

// Safe reports whether the story reassures instead of warning.
func (s Story) Safe() bool { return len(s.Pass) > 0 }

// Checks lists every check the story depends on.
func (s Story) Checks() []string {
	out := append([]string{}, s.All...)
	out = append(out, s.Any...)
	return append(out, s.Pass...)
}

// Load reads stories/<platform>.yaml from fsys. A platform without stories
// has none.
func Load(fsys fs.FS, platform string) ([]Story, error) {
	data, err := fs.ReadFile(fsys, "stories/"+platform+".yaml")
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var file struct {
		Stories []Story `yaml:"stories"`
	}
	if err := yaml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("stories/%s.yaml: %w", platform, err)
	}
	seen := map[string]bool{}
	for _, s := range file.Stories {
		switch {
		case s.ID == "":
			return nil, fmt.Errorf("stories/%s.yaml: a story has no id", platform)
		case seen[s.ID]:
			return nil, fmt.Errorf("stories/%s.yaml: duplicate story %q", platform, s.ID)
		case len(s.All) == 0 && len(s.Any) == 0 && len(s.Pass) == 0:
			return nil, fmt.Errorf("stories/%s.yaml: %s needs all, any or pass", platform, s.ID)
		case len(s.Pass) > 0 && len(s.All)+len(s.Any) > 0:
			return nil, fmt.Errorf("stories/%s.yaml: %s mixes pass with all or any", platform, s.ID)
		case s.Text[check.EN] == "" || s.Text[check.ES] == "":
			return nil, fmt.Errorf("stories/%s.yaml: %s needs en and es text", platform, s.ID)
		}
		seen[s.ID] = true
	}
	return file.Stories, nil
}

// Add fills r.Stories with the stories the report makes true: the risks its
// failures add up to, and the reassurances its passes earn.
func Add(r *scan.Report, stories []Story) {
	failing, passing := map[string]bool{}, map[string]bool{}
	for _, e := range r.Entries {
		if e.Relevance == check.NotApplicable {
			continue
		}
		switch e.Result.Status {
		case check.Fail, check.Warn:
			failing[e.Check.ID] = true
		case check.Pass:
			passing[e.Check.ID] = true
		}
	}
	r.Stories = nil
	for _, s := range stories {
		if s.Safe() {
			if allIn(s.Pass, passing) {
				r.Stories = append(r.Stories, scan.Story{ID: s.ID, Text: s.Text, Safe: true})
			}
			continue
		}
		if fires(s, failing) {
			r.Stories = append(r.Stories, scan.Story{ID: s.ID, Text: s.Text})
		}
	}
}

func allIn(ids []string, set map[string]bool) bool {
	for _, id := range ids {
		if !set[id] {
			return false
		}
	}
	return true
}

func fires(s Story, failing map[string]bool) bool {
	for _, id := range s.All {
		if !failing[id] {
			return false
		}
	}
	if len(s.Any) == 0 {
		return true
	}
	for _, id := range s.Any {
		if failing[id] {
			return true
		}
	}
	return false
}
