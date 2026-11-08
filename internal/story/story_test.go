package story_test

import (
	"testing"
	"testing/fstest"

	"github.com/XexusCode/hardenme/checks"
	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/scan"
	"github.com/XexusCode/hardenme/internal/story"
)

func entry(id string, rel check.Relevance, status check.Status) scan.Entry {
	return scan.Entry{Check: check.Check{ID: id}, Relevance: rel, Result: check.Result{Status: status}}
}

func ids(r scan.Report) []string {
	var out []string
	for _, s := range r.Stories {
		out = append(out, s.ID)
	}
	return out
}

func TestStoriesFire(t *testing.T) {
	stories := []story.Story{
		{ID: "all", All: []string{"a", "b"}},
		{ID: "any", Any: []string{"c", "d"}},
		{ID: "both", All: []string{"a"}, Any: []string{"c", "e"}},
	}
	tests := []struct {
		name    string
		entries []scan.Entry
		want    []string
	}{
		{"nothing failing", []scan.Entry{entry("a", check.Essential, check.Pass)}, nil},
		{"all needs every check", []scan.Entry{entry("a", check.Essential, check.Fail), entry("b", check.Essential, check.Pass)}, nil},
		{"warn counts", []scan.Entry{entry("a", check.Essential, check.Fail), entry("b", check.Optional, check.Warn)}, []string{"all"}},
		{"any needs one", []scan.Entry{entry("d", check.Recommended, check.Fail)}, []string{"any"}},
		{"all and any", []scan.Entry{entry("a", check.Essential, check.Fail), entry("e", check.Essential, check.Fail)}, []string{"both"}},
		{"unknown and manual never fire", []scan.Entry{entry("c", check.Essential, check.Unknown), entry("d", check.Essential, check.Manual)}, nil},
		{"not applicable never fires", []scan.Entry{entry("c", check.NotApplicable, check.Fail)}, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := scan.Report{Entries: tc.entries}
			story.Add(&r, stories)
			if got := ids(r); len(got) != len(tc.want) || (len(got) > 0 && got[0] != tc.want[0]) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLoadRejectsBrokenStories(t *testing.T) {
	tests := map[string]string{
		"no id":        "stories: [{all: [a], text: {en: x, es: y}}]",
		"no checks":    "stories: [{id: s, text: {en: x, es: y}}]",
		"no spanish":   "stories: [{id: s, all: [a], text: {en: x}}]",
		"duplicate id": "stories: [{id: s, all: [a], text: {en: x, es: y}}, {id: s, all: [b], text: {en: x, es: y}}]",
	}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			fsys := fstest.MapFS{"stories/macos.yaml": {Data: []byte(body)}}
			if _, err := story.Load(fsys, "macos"); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
	if s, err := story.Load(fstest.MapFS{}, "windows"); err != nil || s != nil {
		t.Errorf("missing file: %v %v", s, err)
	}
}

// Every story in the catalog refers to real checks of its platform, so a
// renamed check cannot silently switch a story off.
func TestCatalogStoriesUseRealChecks(t *testing.T) {
	for _, platform := range []string{"macos", "ios", "android"} {
		stories, err := story.Load(checks.FS, platform)
		if err != nil {
			t.Fatal(err)
		}
		if len(stories) == 0 {
			continue
		}
		all, err := check.Load(checks.FS, platform)
		if err != nil {
			t.Fatal(err)
		}
		known := map[string]bool{}
		for _, c := range all {
			known[c.ID] = true
		}
		for _, s := range stories {
			for _, id := range append(append([]string{}, s.All...), s.Any...) {
				if !known[id] {
					t.Errorf("%s: unknown check %q", s.ID, id)
				}
			}
		}
	}
	if s, _ := story.Load(checks.FS, "macos"); len(s) < 3 {
		t.Errorf("expected at least 3 macOS stories, got %d", len(s))
	}
}
