package scan_test

import (
	"context"
	"testing"

	"github.com/XexusCode/hardenme/checks"
	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/runner"
	"github.com/XexusCode/hardenme/internal/scan"
)

func entry(rel check.Relevance, status check.Status, credit float64) scan.Entry {
	return scan.Entry{Relevance: rel, Result: check.Result{Status: status, Credit: credit}}
}

func TestScoreWeightsByRelevance(t *testing.T) {
	tests := []struct {
		name    string
		entries []scan.Entry
		want    int
		grade   string
	}{
		{"everything passes", []scan.Entry{entry(check.Essential, check.Pass, 1), entry(check.Optional, check.Pass, 1)}, 100, "A"},
		// essential fail (0/3) + recommended pass (2/2) + optional pass (1/1) = 3/6
		{"one essential failure weighs three times", []scan.Entry{entry(check.Essential, check.Fail, 0), entry(check.Recommended, check.Pass, 1), entry(check.Optional, check.Pass, 1)}, 50, "F"},
		// essential pass (3/3) + optional fail (0/1) = 3/4
		{"an optional failure weighs little", []scan.Entry{entry(check.Essential, check.Pass, 1), entry(check.Optional, check.Fail, 0)}, 75, "C"},
		// essential warn at half credit (1.5/3) + essential pass (3/3) = 4.5/6
		{"partial credit", []scan.Entry{entry(check.Essential, check.Warn, 0.5), entry(check.Essential, check.Pass, 1)}, 75, "C"},
		{"unknown, manual and skipped do not count", []scan.Entry{
			entry(check.Essential, check.Pass, 1),
			entry(check.Essential, check.Unknown, 0),
			entry(check.Essential, check.Manual, 0),
			entry(check.Essential, check.Skipped, 0),
		}, 100, "A"},
		{"not applicable does not count", []scan.Entry{entry(check.Essential, check.Pass, 1), entry(check.NotApplicable, check.Fail, 0)}, 100, "A"},
		{"nothing scorable", []scan.Entry{entry(check.Essential, check.Manual, 0)}, -1, "-"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := scan.Report{Entries: tc.entries}.Score()
			if got.Value != tc.want || got.Grade != tc.grade {
				t.Fatalf("got %d (%s), want %d (%s)", got.Value, got.Grade, tc.want, tc.grade)
			}
		})
	}
}

func TestGradeBoundaries(t *testing.T) {
	for v, want := range map[int]string{100: "A", 90: "A", 89: "B", 80: "B", 79: "C", 70: "C", 69: "D", 60: "D", 59: "F", 0: "F"} {
		if got := scan.Grade(v); got != want {
			t.Errorf("Grade(%d) = %s, want %s", v, got, want)
		}
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		name    string
		entries []scan.Entry
		want    int
	}{
		{"all good", []scan.Entry{entry(check.Essential, check.Pass, 1), entry(check.Essential, check.Manual, 0)}, 0},
		{"essential failure", []scan.Entry{entry(check.Essential, check.Fail, 0)}, 1},
		{"recommended failure", []scan.Entry{entry(check.Recommended, check.Fail, 0)}, 1},
		{"only optional failure", []scan.Entry{entry(check.Optional, check.Fail, 0)}, 2},
		{"only a warning", []scan.Entry{entry(check.Essential, check.Warn, 0.5)}, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := (scan.Report{Entries: tc.entries}).ExitCode(); got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

// The same Mac without FileVault scores lower for someone who carries it
// around than for someone whose desktop never leaves home.
func TestDiskEncryptionMattersMoreOnALaptop(t *testing.T) {
	all, err := check.Load(checks.FS, "macos")
	if err != nil {
		t.Fatal(err)
	}
	var selected []check.Check
	for _, c := range all {
		if c.ID == "macos.filevault" || c.ID == "macos.gatekeeper" {
			selected = append(selected, c)
		}
	}
	fake := &runner.Fake{Outputs: map[string]runner.Output{
		runner.Key("fdesetup", "status"): {Stdout: "FileVault is Off.\n"},
		runner.Key("spctl", "--status"):  {Stdout: "assessments enabled\n"},
	}}
	laptop := scan.Scan(context.Background(), "macos", selected, profile.Profile{Portable: true, LeavesHome: true, Risk: profile.RiskNormal}, fake)
	desktop := scan.Scan(context.Background(), "macos", selected, profile.Profile{Risk: profile.RiskNormal}, fake)

	// laptop: filevault essential fail (0/3) + gatekeeper essential pass (3/3) = 50
	// desktop: filevault optional fail (0/1) + gatekeeper essential pass (3/3) = 75
	if got := laptop.Score().Value; got != 50 {
		t.Errorf("laptop score = %d, want 50", got)
	}
	if got := desktop.Score().Value; got != 75 {
		t.Errorf("desktop score = %d, want 75", got)
	}
	if laptop.ExitCode() != 1 || desktop.ExitCode() != 2 {
		t.Errorf("exit codes = %d/%d, want 1/2", laptop.ExitCode(), desktop.ExitCode())
	}
}

func TestNotApplicableChecksAreNotRun(t *testing.T) {
	all, err := check.Load(checks.FS, "macos")
	if err != nil {
		t.Fatal(err)
	}
	var lockdown []check.Check
	for _, c := range all {
		if c.ID == "macos.lockdown_mode" {
			lockdown = append(lockdown, c)
		}
	}
	fake := &runner.Fake{}
	r := scan.Scan(context.Background(), "macos", lockdown, profile.Profile{Risk: profile.RiskNormal}, fake)
	if len(fake.Calls) != 0 {
		t.Fatalf("ran %v for a check that does not apply", fake.Calls)
	}
	if r.Entries[0].Relevance != check.NotApplicable {
		t.Fatalf("relevance = %s", r.Entries[0].Relevance)
	}
}
