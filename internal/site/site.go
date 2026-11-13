// Package site builds the phone checklist as one static web page, from the
// same check and story files the CLI uses. The page works on the phone
// itself: nothing is installed and nothing leaves the browser.
//
// The page scores answers with its own JavaScript, so the build embeds test
// cases computed by the Go engine. Opening the page with "#selftest" runs
// them, which is how a mismatch between the two is caught.
package site

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/scan"
	"github.com/XexusCode/hardenme/internal/story"
)

//go:embed page.html
var page []byte

// Platforms are the phones the page covers.
var Platforms = []string{"ios", "android"}

// placeholder is replaced by the JSON data in page.html.
const placeholder = `{"__HARDENME_DATA__":true}`

// Data is everything the page needs.
type Data struct {
	Strings   map[check.Lang]map[string]string `json:"strings"`
	Platforms map[string]Platform              `json:"platforms"`
	SelfTest  []Case                           `json:"selftest"`
}

// Platform is one phone's questions and stories.
type Platform struct {
	Checks  []Check `json:"checks"`
	Stories []Story `json:"stories"`
}

// Check is one question.
type Check struct {
	ID        string              `json:"id"`
	Relevance check.RelevanceSpec `json:"relevance"`
	Text      map[check.Lang]Text `json:"text"`
	// Links are the official pages behind the check, by language.
	Links map[check.Lang]string `json:"links"`
}

// Text is a check's wording in one language.
type Text struct {
	Title     string   `json:"title"`
	TitlePass string   `json:"title_pass"`
	TitleFail string   `json:"title_fail"`
	Question  string   `json:"question"`
	Where     string   `json:"where"`
	Why       string   `json:"why"`
	Steps     []string `json:"steps"`
}

// Story is a risk or a reassurance and the checks behind it.
type Story struct {
	ID   string                `json:"id"`
	All  []string              `json:"all,omitempty"`
	Any  []string              `json:"any,omitempty"`
	Pass []string              `json:"pass,omitempty"`
	Text map[check.Lang]string `json:"text"`
}

// Case is one self-test: for these answers, the Go engine got this result.
type Case struct {
	Platform string                  `json:"platform"`
	Profile  profile.Profile         `json:"profile"`
	Answers  map[string]check.Answer `json:"answers"`
	Score    int                     `json:"score"`
	Grade    string                  `json:"grade"`
	Stories  []string                `json:"stories"`
}

// Build returns the page with the catalog's data inside it.
func Build(catalog fs.FS) ([]byte, error) {
	data := Data{Strings: strs, Platforms: map[string]Platform{}}
	for _, platform := range Platforms {
		checks, err := check.Load(catalog, platform)
		if err != nil {
			return nil, err
		}
		stories, err := story.Load(catalog, platform)
		if err != nil {
			return nil, err
		}
		var p Platform
		for _, c := range checks {
			if !c.Probe.Ask {
				continue
			}
			out := Check{ID: c.ID, Relevance: c.Relevance, Text: map[check.Lang]Text{}, Links: links(c.References)}
			for _, lang := range []check.Lang{check.EN, check.ES} {
				t := c.TextFor(lang)
				out.Text[lang] = Text{
					Title: t.Title, TitlePass: t.TitlePass, TitleFail: t.TitleFail,
					Question: t.Question, Where: t.Where, Why: t.Why, Steps: t.FixSteps,
				}
			}
			p.Checks = append(p.Checks, out)
		}
		for _, s := range stories {
			p.Stories = append(p.Stories, Story{ID: s.ID, All: s.All, Any: s.Any, Pass: s.Pass, Text: s.Text})
		}
		data.Platforms[platform] = p
		data.SelfTest = append(data.SelfTest, cases(platform, checks, stories)...)
	}
	js, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	// Keep the JSON from ending the script element it sits in.
	js = bytes.ReplaceAll(js, []byte("</"), []byte(`<\/`))
	if !bytes.Contains(page, []byte(placeholder)) {
		return nil, fmt.Errorf("page.html has no data placeholder")
	}
	return bytes.Replace(page, []byte(placeholder), js, 1), nil
}

// links picks one official page per language: the Spanish one when a
// reference is marked as such, the first one otherwise.
func links(refs []string) map[check.Lang]string {
	out := map[check.Lang]string{}
	if len(refs) == 0 {
		return out
	}
	out[check.EN], out[check.ES] = refs[0], refs[0]
	for _, ref := range refs {
		if bytes.Contains([]byte(ref), []byte("/es-es/")) || bytes.HasSuffix([]byte(ref), []byte("hl=es")) {
			out[check.ES] = ref
			break
		}
	}
	return out
}

// cases runs the real engine over a spread of profiles and answers.
func cases(platform string, checks []check.Check, stories []story.Story) []Case {
	profiles := []profile.Profile{
		profile.DefaultPhone(),
		{Portable: true, Risk: profile.RiskNormal},
		{Portable: true, LeavesHome: true, Shared: true, SensitiveData: true, Risk: profile.RiskNormal},
		{Portable: true, LeavesHome: true, SensitiveData: true, Risk: profile.RiskElevated},
	}
	patterns := []func(i int) check.Answer{
		func(int) check.Answer { return check.Yes },
		func(int) check.Answer { return check.No },
		func(int) check.Answer { return check.DontKnow },
		func(i int) check.Answer { return []check.Answer{check.Yes, check.No}[i%2] },
		func(i int) check.Answer { return []check.Answer{check.No, check.Yes, check.DontKnow}[i%3] },
		func(i int) check.Answer {
			return []check.Answer{check.Yes, check.Yes, check.No, check.DontKnow, check.Yes}[i%5]
		},
	}
	var out []Case
	for _, p := range profiles {
		for _, pattern := range patterns {
			answers := map[string]check.Answer{}
			for i, c := range checks {
				answers[c.ID] = pattern(i)
			}
			report := scan.ScanEnv(context.Background(), platform, checks, check.Env{
				Profile: p,
				Ask:     func(c check.Check) check.Answer { return answers[c.ID] },
			})
			story.Add(&report, stories)
			c := Case{Platform: platform, Profile: p, Answers: answers, Stories: []string{}}
			score := report.Score()
			c.Score, c.Grade = score.Value, score.Grade
			for _, s := range report.Stories {
				c.Stories = append(c.Stories, s.ID)
			}
			out = append(out, c)
		}
	}
	return out
}
