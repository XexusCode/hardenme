package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/XexusCode/hardenme/internal/answers"
	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/profile"
)

func phoneCheck(id, level, question string) string {
	return `id: ios.` + id + `
platform: ios
category: theft
severity: high
relevance:
  default: ` + level + `
probe: {ask: true}
text:
  en: {title: ` + id + `, title_pass: ` + id + ` ok, title_fail: fix ` + id + `, why: w, fix_steps: [s], question: "EN ` + question + `?", where: "Settings"}
  es: {title: ` + id + `, title_pass: ` + id + ` bien, title_fail: arregla ` + id + `, why: w, fix_steps: [s], question: "¿` + question + `?", where: "Ajustes"}
references: [https://support.apple.com/x]
`
}

type checklistEnv struct {
	deps    checklistDeps
	answers answers.Store
	profile profile.Store
}

func newChecklistEnv(t *testing.T) checklistEnv {
	dir := t.TempDir()
	e := checklistEnv{
		answers: answers.Store{Path: filepath.Join(dir, "answers-ios.json")},
		profile: profile.Store{Path: filepath.Join(dir, "profile-phone.json")},
	}
	e.deps = checklistDeps{
		catalog: fstest.MapFS{
			"ios/aaa.yaml": {Data: []byte(phoneCheck("aaa", "essential", "Uno"))},
			"ios/bbb.yaml": {Data: []byte(phoneCheck("bbb", "recommended", "Dos"))},
			"ios/abc.yaml": {Data: []byte(phoneCheck("abc", "optional", "Tres"))},
		},
		answers:  func(string) (answers.Store, error) { return e.answers, nil },
		profile:  func() (profile.Store, error) { return e.profile, nil },
		terminal: func(io.Reader) bool { return true },
		now:      func() time.Time { return time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC) },
	}
	return e
}

func (e checklistEnv) run(t *testing.T, input string, args ...string) (string, error) {
	t.Helper()
	cmd := newChecklistWith(&options{lang: "es"}, e.deps)
	var out bytes.Buffer
	cmd.SetIn(strings.NewReader(input))
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(append([]string{"ios"}, args...))
	err := cmd.ExecuteContext(context.Background())
	return out.String(), err
}

func exitCode(err error) int {
	var exit ExitError
	if errors.As(err, &exit) {
		return exit.Code
	}
	if err != nil {
		return 3
	}
	return 0
}

func TestChecklistFirstRunThenOnlyAsksWhatWasNotRight(t *testing.T) {
	e := newChecklistEnv(t)
	// Profile: leaves home (Enter), not shared, no sensitive data, normal risk.
	// Then: aaa yes, bbb no, ccc ? (with one answer it does not understand).
	out, err := e.run(t, "\nn\nn\nn\ns\nn\nquizá\n?\n")
	if exitCode(err) != 1 {
		t.Fatalf("exit %v, want 1 (a recommended check failed)\n%s", err, out)
	}
	for _, want := range []string{
		"Vamos a revisar tu iPhone",
		"Primero, unas preguntas",
		"1/3 ¿Uno?",
		"Dónde: Ajustes",
		"3/3 ¿Tres?",
		"Contesta s, n, ? o q",
		"Perfil: móvil, sale de casa",
		"arregla bbb",
		"REVÍSALO TÚ (1)",
		"YA ESTÁ BIEN (1)",
		"hardenme checklist ios --profile",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
	if _, err := e.profile.Load(); err != nil {
		t.Errorf("phone profile not saved: %v", err)
	}
	saved, _ := e.answers.Load()
	if saved["ios.aaa"].Answer != check.Yes || saved["ios.bbb"].Answer != check.No || saved["ios.abc"].Answer != check.DontKnow {
		t.Errorf("saved answers: %v", saved)
	}

	// Second run: no profile questions, aaa is not asked again.
	out, err = e.run(t, "s\ns\n")
	if exitCode(err) != 0 {
		t.Fatalf("exit %v, want 0\n%s", err, out)
	}
	if strings.Contains(out, "Primero, unas preguntas") || strings.Contains(out, "¿Uno?") {
		t.Errorf("asked again what was already right:\n%s", out)
	}
	for _, want := range []string{"Me salto lo que ya estaba bien", "1/2 ¿Dos?", "2/2 ¿Tres?", "100/100"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}

	// --again asks everything.
	out, _ = e.run(t, "s\ns\ns\n", "--again")
	if !strings.Contains(out, "1/3 ¿Uno?") {
		t.Errorf("--again did not ask everything:\n%s", out)
	}
}

func TestChecklistQuitHalfway(t *testing.T) {
	e := newChecklistEnv(t)
	if err := e.profile.Save(profile.DefaultPhone()); err != nil {
		t.Fatal(err)
	}
	out, _ := e.run(t, "s\nq\n")
	if strings.Contains(out, "¿Tres?") {
		t.Errorf("kept asking after quit:\n%s", out)
	}
	if !strings.Contains(out, "Lo has dejado a medias") || !strings.Contains(out, "NO SE HA PODIDO COMPROBAR (2)") {
		t.Errorf("quit not reported:\n%s", out)
	}
	saved, _ := e.answers.Load()
	if len(saved) != 1 || saved["ios.aaa"].Answer != check.Yes {
		t.Errorf("saved %v, want only aaa=yes", saved)
	}
}

func TestChecklistNeedsATerminalAndAPhone(t *testing.T) {
	e := newChecklistEnv(t)
	e.deps.terminal = func(io.Reader) bool { return false }
	if _, err := e.run(t, ""); err == nil || !strings.Contains(err.Error(), "Terminal") {
		t.Errorf("without a terminal: %v", err)
	}
	e = newChecklistEnv(t)
	cmd := newChecklistWith(&options{lang: "en"}, e.deps)
	cmd.SetArgs([]string{"windows"})
	cmd.SetOut(io.Discard)
	if err := cmd.Execute(); err == nil {
		t.Error("accepted an unknown phone")
	}
}

func TestParseAnswer(t *testing.T) {
	tests := map[string]check.Answer{
		"s": check.Yes, " Sí ": check.Yes, "YES": check.Yes,
		"n": check.No, "No": check.No,
		"?": check.DontKnow, "no sé": check.DontKnow, "No  lo  se": check.DontKnow, "idk": check.DontKnow,
		"q": check.Quit, "salir": check.Quit,
	}
	for in, want := range tests {
		if got, ok := parseAnswer(in); !ok || got != want {
			t.Errorf("%q: %v %v, want %v", in, got, ok, want)
		}
	}
	for _, in := range []string{"", "quizá", "1"} {
		if _, ok := parseAnswer(in); ok {
			t.Errorf("%q should not be understood", in)
		}
	}
}
