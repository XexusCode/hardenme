// Package scan runs a platform's checks for a profile and scores the result.
package scan

import (
	"context"
	"math"
	"time"

	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/runner"
)

// Entry is one check with how much it matters and what was found.
type Entry struct {
	Check     check.Check
	Relevance check.Relevance
	Result    check.Result
}

// Report is the outcome of a scan.
type Report struct {
	Platform  string
	OSVersion string
	Profile   profile.Profile
	StartedAt time.Time
	Entries   []Entry
	// Stories are the plain-language risks the failures add up to.
	Stories []Story
}

// Story is a plain-language risk, e.g. "if your Mac is stolen, the thief
// could read your files", in each language.
type Story struct {
	ID   string
	Text map[check.Lang]string
}

// Scan runs every check that applies to the profile. Checks that do not
// apply are kept in the report (as not applicable) but never run.
func Scan(ctx context.Context, platform string, checks []check.Check, p profile.Profile, r runner.Runner) Report {
	return ScanEnv(ctx, platform, checks, check.Env{Runner: r, Profile: p})
}

// ScanEnv is Scan with a full environment, e.g. one that can ask the
// person (phone checklists). The profile is env.Profile.
func ScanEnv(ctx context.Context, platform string, checks []check.Check, env check.Env) Report {
	p := env.Profile
	report := Report{Platform: platform, Profile: p, StartedAt: time.Now()}
	for _, c := range checks {
		rel := c.RelevanceFor(p)
		entry := Entry{Check: c, Relevance: rel}
		if rel == check.NotApplicable {
			entry.Result = check.Result{CheckID: c.ID, Status: check.Skipped}
		} else {
			entry.Result = c.Run(ctx, env)
		}
		report.Entries = append(report.Entries, entry)
	}
	return report
}

// Score is the weighted result: Value is 0-100, or -1 when nothing could be
// scored.
type Score struct {
	Value int
	Grade string
}

// Score weights each scored check by its relevance (essential 3,
// recommended 2, optional 1, not applicable 0). Checks that could not be
// read, must be checked by hand or were skipped do not count.
func (r Report) Score() Score {
	var earned, max float64
	for _, e := range r.Entries {
		if !e.Result.Status.Scored() {
			continue
		}
		w := e.Relevance.Weight()
		earned += e.Result.Credit * w
		max += w
	}
	if max == 0 {
		return Score{Value: -1, Grade: "-"}
	}
	v := int(math.Round(earned / max * 100))
	return Score{Value: v, Grade: Grade(v)}
}

// Grade turns a 0-100 score into a letter.
func Grade(v int) string {
	switch {
	case v >= 90:
		return "A"
	case v >= 80:
		return "B"
	case v >= 70:
		return "C"
	case v >= 60:
		return "D"
	default:
		return "F"
	}
}

// ExitCode is 1 when an essential or recommended check failed, 2 when there
// are only warnings or optional failures, and 0 otherwise.
func (r Report) ExitCode() int {
	code := 0
	for _, e := range r.Entries {
		important := e.Relevance == check.Essential || e.Relevance == check.Recommended
		switch {
		case e.Result.Status == check.Fail && important:
			return 1
		case e.Result.Status == check.Fail, e.Result.Status == check.Warn:
			code = 2
		}
	}
	return code
}
