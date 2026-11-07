package render_test

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/XexusCode/hardenme/checks"
	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/render"
	"github.com/XexusCode/hardenme/internal/runner"
	"github.com/XexusCode/hardenme/internal/scan"
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
	return r
}

func TestTextGolden(t *testing.T) {
	for _, lang := range []check.Lang{check.EN, check.ES} {
		t.Run(string(lang), func(t *testing.T) {
			var buf bytes.Buffer
			if err := render.Render(&buf, sampleReport(t), render.Text, lang); err != nil {
				t.Fatal(err)
			}
			golden := filepath.Join("testdata", "report."+string(lang)+".golden")
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
		})
	}
}

func TestUnsupportedFormat(t *testing.T) {
	if _, err := render.ParseFormat("pdf"); err == nil {
		t.Fatal("expected an error")
	}
}
