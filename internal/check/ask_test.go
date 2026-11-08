package check_test

import (
	"context"
	"testing"

	"github.com/XexusCode/hardenme/internal/check"
)

const askText = `text:
  en: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d], question: "Is it on?", where: "Settings"}
  es: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d], question: "¿Está activado?", where: "Ajustes"}
references: [https://example.com]
`

func loadOne(t *testing.T, platform, body string) (check.Check, error) {
	t.Helper()
	fsys := memFS(map[string]string{platform + "/x.yaml": body})
	all, err := check.Load(fsys, platform)
	if err != nil {
		return check.Check{}, err
	}
	return all[0], nil
}

func TestAskProbeMapsAnswers(t *testing.T) {
	c, err := loadOne(t, "ios", "id: ios.x\nplatform: ios\nrelevance: {default: essential}\nprobe: {ask: true}\n"+askText)
	if err != nil {
		t.Fatal(err)
	}
	tests := map[check.Answer]check.Status{
		check.Yes:      check.Pass,
		check.No:       check.Fail,
		check.DontKnow: check.Manual,
		check.Quit:     check.Skipped,
	}
	for answer, want := range tests {
		var asked check.Check
		res := c.Run(context.Background(), check.Env{Ask: func(c check.Check) check.Answer { asked = c; return answer }})
		if res.Status != want {
			t.Errorf("%s: status %s, want %s", answer, res.Status, want)
		}
		if asked.ID != "ios.x" {
			t.Errorf("%s: Ask got %q", answer, asked.ID)
		}
	}
	if res := c.Run(context.Background(), check.Env{}); res.Status != check.Manual {
		t.Errorf("without Ask: status %s, want manual", res.Status)
	}
	if res := c.Run(context.Background(), check.Env{Ask: func(check.Check) check.Answer { return check.Yes }}); res.Credit != 1 {
		t.Errorf("yes should give full credit, got %v", res.Credit)
	}
}

func TestLoadRejectsBrokenAskChecks(t *testing.T) {
	noQuestion := `text:
  en: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d]}
  es: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d]}
references: [https://example.com]
`
	tests := map[string]struct{ platform, body string }{
		"ask without question":  {"ios", "id: ios.x\nplatform: ios\nrelevance: {default: essential}\nprobe: {ask: true}\n" + noQuestion},
		"question without ask":  {"macos", "id: macos.x\nplatform: macos\nrelevance: {default: essential}\nprobe: {manual: true}\n" + askText},
		"command on a phone":    {"android", "id: android.x\nplatform: android\nrelevance: {default: essential}\nprobe: {command: [x], rules: [{exit: 0, status: pass}]}\n" + noQuestion},
		"ask and manual":        {"ios", "id: ios.x\nplatform: ios\nrelevance: {default: essential}\nprobe: {ask: true, manual: true}\n" + askText},
		"missing neutral title": {"ios", "id: ios.x\nplatform: ios\nrelevance: {default: essential}\nprobe: {ask: true}\ntext:\n  en: {title_pass: a, title_fail: b, why: c, fix_steps: [d], question: q}\n  es: {title_pass: a, title_fail: b, why: c, fix_steps: [d], question: q}\nreferences: [https://example.com]\n"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := loadOne(t, tc.platform, tc.body); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
