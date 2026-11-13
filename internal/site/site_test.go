package site_test

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/XexusCode/hardenme/checks"
	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/site"
)

var dataRE = regexp.MustCompile(`(?s)<script id="data" type="application/json">(.*?)</script>`)

func build(t *testing.T) ([]byte, site.Data) {
	t.Helper()
	html, err := site.Build(checks.FS)
	if err != nil {
		t.Fatal(err)
	}
	m := dataRE.FindSubmatch(html)
	if m == nil {
		t.Fatal("no data block in the page")
	}
	var data site.Data
	if err := json.Unmarshal(bytes.ReplaceAll(m[1], []byte(`<\/`), []byte("</")), &data); err != nil {
		t.Fatalf("the embedded data is not valid JSON: %v", err)
	}
	return html, data
}

func TestPageCarriesBothPhonesInBothLanguages(t *testing.T) {
	_, data := build(t)
	for _, platform := range site.Platforms {
		p := data.Platforms[platform]
		if len(p.Checks) < 10 || len(p.Stories) < 5 {
			t.Fatalf("%s: %d checks, %d stories", platform, len(p.Checks), len(p.Stories))
		}
		for _, c := range p.Checks {
			for _, lang := range []check.Lang{check.EN, check.ES} {
				text := c.Text[lang]
				if text.Question == "" || text.Where == "" || text.Why == "" || len(text.Steps) == 0 || text.TitleFail == "" {
					t.Errorf("%s (%s): incomplete text", c.ID, lang)
				}
				if !strings.HasPrefix(c.Links[lang], "https://") {
					t.Errorf("%s (%s): no official link", c.ID, lang)
				}
			}
		}
	}
	for key := range data.Strings[check.EN] {
		if data.Strings[check.ES][key] == "" {
			t.Errorf("string %q has no Spanish text", key)
		}
	}
}

func TestSelfTestCoversScoresGradesAndStories(t *testing.T) {
	_, data := build(t)
	if len(data.SelfTest) < 40 {
		t.Fatalf("only %d self-test cases", len(data.SelfTest))
	}
	grades, noScore, withStories := map[string]bool{}, false, false
	for _, c := range data.SelfTest {
		grades[c.Grade] = true
		noScore = noScore || c.Score == -1
		withStories = withStories || len(c.Stories) > 0
	}
	if !grades["A"] || !grades["F"] || !noScore || !withStories {
		t.Errorf("cases do not cover enough: grades %v, no-score %v, stories %v", grades, noScore, withStories)
	}
}

func TestPageIsSelfContained(t *testing.T) {
	html, _ := build(t)
	page := dataRE.ReplaceAll(html, nil)
	if bytes.Contains(html, []byte("__HARDENME_DATA__")) {
		t.Error("the data placeholder was not replaced")
	}
	for _, bad := range []string{"<script src", "<link", "fetch(", "XMLHttpRequest", "innerHTML", "eval(", "sendBeacon", "<img"} {
		if bytes.Contains(page, []byte(bad)) {
			t.Errorf("the page contains %q", bad)
		}
	}
	if bytes.Contains(page, []byte("kids.flat()")) {
		t.Error("lists of elements must be flattened at every depth, or they print as [object ...]")
	}
	if !bytes.Contains(page, []byte("default-src 'none'")) {
		t.Error("no restrictive Content-Security-Policy")
	}
}
