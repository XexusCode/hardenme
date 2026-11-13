package render_test

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/XexusCode/hardenme/checks"
	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/render"
	"github.com/XexusCode/hardenme/internal/runner"
	"github.com/XexusCode/hardenme/internal/scan"
	"github.com/XexusCode/hardenme/internal/story"
)

var update = flag.Bool("update", false, "rewrite the golden files")

// sampleReport scans a few real checks against outputs captured on macOS 26.2.
func sampleReport(t *testing.T) scan.Report {
	t.Helper()
	all, err := check.Load(checks.FS, "macos")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"macos.filevault": true, "macos.gatekeeper": true, "macos.firewall_stealth": true,
		"macos.admin_account": true, "macos.find_my": true, "macos.lockdown_mode": true, "macos.sip": true,
	}
	var selected []check.Check
	for _, c := range all {
		if want[c.ID] {
			selected = append(selected, c)
		}
	}
	fake := &runner.Fake{Outputs: map[string]runner.Output{
		runner.Key("fdesetup", "status"): {Stdout: "FileVault is Off.\n"},
		runner.Key("spctl", "--status"):  {Stdout: "assessments enabled\n"},
		runner.Key("/usr/libexec/ApplicationFirewall/socketfilterfw", "--getstealthmode"): {Stdout: "Firewall stealth mode is off\n"},
		runner.Key("id", "-Gn"):                       {Stdout: "staff everyone admin\n"},
		runner.Key("nvram", "fmm-mobileme-token-FMM"): {Stderr: "nvram: Error getting variable", ExitCode: 1},
	}}
	p := profile.Profile{Portable: true, LeavesHome: true, Risk: profile.RiskNormal}
	r := scan.Scan(context.Background(), "macos", selected, p, fake)
	r.OSVersion = "26.2"
	r.StartedAt = time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC)
	stories, err := story.Load(checks.FS, "macos")
	if err != nil {
		t.Fatal(err)
	}
	story.Add(&r, stories)
	return r
}

func TestGolden(t *testing.T) {
	ext := map[render.Format]string{render.Text: ".golden", render.JSON: ".json.golden", render.Markdown: ".md.golden"}
	for _, f := range []render.Format{render.Text, render.JSON, render.Markdown} {
		for _, lang := range []check.Lang{check.EN, check.ES} {
			t.Run(string(f)+"/"+string(lang), func(t *testing.T) {
				golden(t, f, lang, filepath.Join("testdata", "report."+string(lang)+ext[f]))
			})
		}
	}
}

func golden(t *testing.T, f render.Format, lang check.Lang, golden string) {
	t.Helper()
	{
		{
			var buf bytes.Buffer
			if err := render.Render(&buf, sampleReport(t), f, lang); err != nil {
				t.Fatal(err)
			}
			if *update {
				if err := os.WriteFile(golden, buf.Bytes(), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("%v (run go test ./internal/render -update)", err)
			}
			if !bytes.Equal(buf.Bytes(), want) {
				t.Fatalf("output differs from %s (run with -update after checking it):\n%s", golden, buf.String())
			}
		}
	}
}

func TestJSONRoundTrips(t *testing.T) {
	var buf bytes.Buffer
	if err := render.Render(&buf, sampleReport(t), render.JSON, check.EN); err != nil {
		t.Fatal(err)
	}
	var got render.JSONReport
	dec := json.NewDecoder(&buf)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.SchemaVersion != 1 || got.Score == nil || got.Score.Value != 39 || len(got.Entries) != 7 {
		t.Fatalf("unexpected report: %+v", got)
	}
}

func TestReportsHideTheHomeFolder(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home folder")
	}
	r := sampleReport(t)
	r.Entries[0].Result.Current = "cannot read " + home + "/Library/x"
	for _, f := range []render.Format{render.JSON, render.Markdown} {
		var buf bytes.Buffer
		if err := render.Render(&buf, r, f, check.EN); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(buf.String(), home) || !strings.Contains(buf.String(), "~/Library/x") {
			t.Errorf("%s: home folder not replaced", f)
		}
	}
}

func TestUnsupportedFormat(t *testing.T) {
	if _, err := render.ParseFormat("pdf"); err == nil {
		t.Fatal("expected an error")
	}
	if f, err := render.ParseFormat("md"); err != nil || f != render.Markdown {
		t.Fatalf("md: %v %v", f, err)
	}
}

func TestSafeStoriesHaveTheirOwnSection(t *testing.T) {
	r := sampleReport(t)
	r.Stories = append(r.Stories, scan.Story{ID: "x.safe", Safe: true, Text: map[check.Lang]string{check.EN: "They could not do this."}})
	for _, tc := range []struct {
		platform, risks, safe string
	}{
		{"macos", "WHAT COULD HAPPEN TODAY", "WHAT YOU ARE ALREADY PROTECTED FROM"},
		{"ios", "IF THIS PHONE IS STOLEN TODAY", "WHAT A THIEF COULD NOT DO"},
	} {
		r.Platform = tc.platform
		var buf bytes.Buffer
		if err := render.Render(&buf, r, render.Text, check.EN); err != nil {
			t.Fatal(err)
		}
		out := buf.String()
		risks, safe := strings.Index(out, tc.risks), strings.Index(out, tc.safe)
		if risks < 0 || safe < risks || !strings.Contains(out[safe:], "✓ They could not do this.") {
			t.Errorf("%s: sections missing or out of order:\n%s", tc.platform, out[:600])
		}
		if strings.Contains(out[risks:safe], "They could not do this.") {
			t.Errorf("%s: the reassurance is listed as a risk", tc.platform)
		}
	}
}

func TestFindingsAreQuestionsOutsideTheScore(t *testing.T) {
	r := sampleReport(t)
	before := r.Score()
	r.FindingsChecked = true
	for _, f := range []render.Format{render.Text, render.Markdown} {
		var buf bytes.Buffer
		if err := render.Render(&buf, r, f, check.ES); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(buf.String(), "No he visto nada raro") || !strings.Contains(buf.String(), "no es un antivirus") {
			t.Errorf("%s: an empty result must say it is no guarantee:\n%s", f, buf.String()[:500])
		}
	}
	r.Findings = []scan.Finding{{Kind: "remote_tool", Items: []string{"AnyDesk"}}}
	if r.Score() != before {
		t.Error("findings changed the score")
	}
	for _, f := range []render.Format{render.Text, render.Markdown, render.JSON} {
		var buf bytes.Buffer
		if err := render.Render(&buf, r, f, check.ES); err != nil {
			t.Fatal(err)
		}
		out := buf.String()
		if !strings.Contains(out, "¿Instalaste tú estas apps de acceso remoto?") || !strings.Contains(out, "AnyDesk") {
			t.Errorf("%s: finding missing:\n%s", f, out[:500])
		}
	}
}
