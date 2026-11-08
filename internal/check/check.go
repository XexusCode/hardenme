// Package check defines security checks as data, decides how much each one
// matters for a given profile, and runs their probes.
package check

import (
	"fmt"
	"regexp"

	"github.com/XexusCode/hardenme/internal/profile"
)

// Lang is a supported language.
type Lang string

const (
	EN Lang = "en"
	ES Lang = "es"
)

// ParseLang validates a language code.
func ParseLang(s string) (Lang, error) {
	switch Lang(s) {
	case EN, ES:
		return Lang(s), nil
	}
	return "", fmt.Errorf("unsupported language %q (use en or es)", s)
}

// Status is the outcome of a check.
type Status string

const (
	Pass    Status = "pass"
	Fail    Status = "fail"
	Warn    Status = "warn"
	Unknown Status = "unknown" // the probe ran but the state could not be read
	Manual  Status = "manual"  // the user has to check it themselves
	Skipped Status = "skipped" // not run (e.g. needs admin rights)
)

// Scored reports whether a status counts towards the score.
func (s Status) Scored() bool { return s == Pass || s == Fail || s == Warn }

// Relevance is how much a check matters for a profile.
type Relevance string

const (
	NotApplicable Relevance = "not_applicable"
	Optional      Relevance = "optional"
	Recommended   Relevance = "recommended"
	Essential     Relevance = "essential"
)

var relevanceRank = map[Relevance]int{NotApplicable: 0, Optional: 1, Recommended: 2, Essential: 3}

// Weight is the score multiplier for a relevance level.
func (r Relevance) Weight() float64 { return float64(relevanceRank[r]) }

func (r Relevance) valid() bool { _, ok := relevanceRank[r]; return ok }

// Condition matches profile fields. Unset fields match anything.
type Condition struct {
	Portable      *bool         `yaml:"portable" json:"portable,omitempty"`
	LeavesHome    *bool         `yaml:"leaves_home" json:"leaves_home,omitempty"`
	Shared        *bool         `yaml:"shared" json:"shared,omitempty"`
	SensitiveData *bool         `yaml:"sensitive_data" json:"sensitive_data,omitempty"`
	Risk          *profile.Risk `yaml:"risk" json:"risk,omitempty"`
}

func (c Condition) matches(p profile.Profile) bool {
	eq := func(want *bool, got bool) bool { return want == nil || *want == got }
	return eq(c.Portable, p.Portable) &&
		eq(c.LeavesHome, p.LeavesHome) &&
		eq(c.Shared, p.Shared) &&
		eq(c.SensitiveData, p.SensitiveData) &&
		(c.Risk == nil || *c.Risk == p.Risk)
}

// RelevanceRule raises a check's relevance when the profile matches.
type RelevanceRule struct {
	When  Condition `yaml:"when" json:"when"`
	Level Relevance `yaml:"level" json:"level"`
}

// RelevanceSpec is the default relevance plus the rules that raise it.
type RelevanceSpec struct {
	Default Relevance       `yaml:"default" json:"default"`
	Rules   []RelevanceRule `yaml:"rules" json:"rules,omitempty"`
}

// Rule maps a command's output to a status. All set fields must match;
// the first matching rule wins.
type Rule struct {
	Match  string   `yaml:"match" json:"match,omitempty"`
	Exit   *int     `yaml:"exit" json:"exit,omitempty"`
	Status Status   `yaml:"status" json:"status"`
	Credit *float64 `yaml:"credit" json:"credit,omitempty"`

	re *regexp.Regexp
}

// Probe says how to read the state. Exactly one of Command, Builtin or
// Manual is set.
type Probe struct {
	Command []string `yaml:"command" json:"command,omitempty"`
	Rules   []Rule   `yaml:"rules" json:"rules,omitempty"`
	// Default is the status when the command ran but no rule matched.
	// It is Unknown unless set.
	Default Status `yaml:"default" json:"default,omitempty"`
	Builtin string `yaml:"builtin" json:"builtin,omitempty"`
	Manual  bool   `yaml:"manual" json:"manual,omitempty"`
}

// Text is everything shown to the user about a check, in one language.
type Text struct {
	// Title names the setting without saying whether it is right, for
	// results that could not be read.
	Title     string   `yaml:"title" json:"title"`
	TitlePass string   `yaml:"title_pass" json:"title_pass"`
	TitleFail string   `yaml:"title_fail" json:"title_fail"`
	Why       string   `yaml:"why" json:"why"`
	FixSteps  []string `yaml:"fix_steps" json:"fix_steps"`
	DeepLink  string   `yaml:"deep_link" json:"deep_link,omitempty"`
	Command   string   `yaml:"command" json:"command,omitempty"`
	// FixWarning is shown before hardenme applies the fix, e.g. "keep it on
	// if you connect to this Mac from another computer".
	FixWarning string `yaml:"fix_warning" json:"fix_warning,omitempty"`
}

// FixKind says how hardenme can fix a check.
type FixKind string

const (
	FixAuto   FixKind = "auto"   // hardenme applies it
	FixGuided FixKind = "guided" // hardenme opens the setting and waits
	FixInfo   FixKind = "info"   // only the written steps
)

// FixStep is one change. Commands are argv arrays with absolute paths and
// never go through a shell. Undo may contain {previous}, replaced by the
// value captured with Read and ReadMatch just before applying.
type FixStep struct {
	Admin bool     `yaml:"admin" json:"admin,omitempty"`
	Read  []string `yaml:"read" json:"read,omitempty"`
	// ReadMatch is a regexp with one group that extracts the previous value
	// from Read's output. When it does not match, the step does not apply
	// to this device and is skipped.
	ReadMatch string `yaml:"read_match" json:"read_match,omitempty"`
	// PreviousMap translates the value read into the form Undo needs, e.g.
	// "1" -> "true". A value missing from the map makes the step unsafe to
	// undo, so it is skipped.
	PreviousMap map[string]string `yaml:"previous_map" json:"previous_map,omitempty"`
	Apply       []string          `yaml:"apply" json:"apply"`
	Undo        []string          `yaml:"undo" json:"undo"`
	// UndoIfMissing is the undo when Read fails (the setting did not exist,
	// so macOS was using its default).
	UndoIfMissing []string `yaml:"undo_if_missing" json:"undo_if_missing,omitempty"`

	re *regexp.Regexp
}

// Matcher returns the compiled ReadMatch, or nil.
func (s FixStep) Matcher() *regexp.Regexp { return s.re }

// Fix is how hardenme applies a check's fix automatically.
type Fix struct {
	Steps []FixStep `yaml:"steps" json:"steps"`
}

// Check is one security check, loaded from checks/<platform>/<name>.yaml.
type Check struct {
	ID         string        `yaml:"id" json:"id"`
	Platform   string        `yaml:"platform" json:"platform"`
	Category   string        `yaml:"category" json:"category"`
	Severity   string        `yaml:"severity" json:"severity"`
	Relevance  RelevanceSpec `yaml:"relevance" json:"relevance"`
	Probe      Probe         `yaml:"probe" json:"probe"`
	Text       map[Lang]Text `yaml:"text" json:"text"`
	References []string      `yaml:"references" json:"references"`
	Credits    []string      `yaml:"credits" json:"credits,omitempty"`
	Fix        *Fix          `yaml:"fix" json:"fix,omitempty"`
}

// FixKind is auto when the check has fix steps, guided when it has a link
// to the setting, and info otherwise.
func (c Check) FixKind() FixKind {
	switch {
	case c.Fix != nil:
		return FixAuto
	case c.TextFor(EN).DeepLink != "":
		return FixGuided
	default:
		return FixInfo
	}
}

// RelevanceFor returns how much the check matters for p: the highest level
// among the default and every rule whose condition matches.
func (c Check) RelevanceFor(p profile.Profile) Relevance {
	level := c.Relevance.Default
	for _, rule := range c.Relevance.Rules {
		if rule.When.matches(p) && relevanceRank[rule.Level] > relevanceRank[level] {
			level = rule.Level
		}
	}
	return level
}

// TextFor returns the text in lang, falling back to English.
func (c Check) TextFor(lang Lang) Text {
	if t, ok := c.Text[lang]; ok {
		return t
	}
	return c.Text[EN]
}

// Result is the outcome of running one check.
type Result struct {
	CheckID  string  `json:"check_id"`
	Status   Status  `json:"status"`
	Current  string  `json:"current,omitempty"`
	Evidence string  `json:"evidence,omitempty"`
	Credit   float64 `json:"-"` // share of the points earned, 0..1
}
