package check_test

import (
	"encoding/json"
	"io/fs"
	"path"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"

	"github.com/XexusCode/hardenme/checks"
	"github.com/XexusCode/hardenme/internal/check"
)

func TestEveryCheckMatchesTheSchema(t *testing.T) {
	schemaData, err := fs.ReadFile(checks.FS, "schema.json")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := jsonschema.UnmarshalJSON(strings.NewReader(string(schemaData)))
	if err != nil {
		t.Fatal(err)
	}
	c := jsonschema.NewCompiler()
	if err := c.AddResource("schema.json", doc); err != nil {
		t.Fatal(err)
	}
	schema, err := c.Compile("schema.json")
	if err != nil {
		t.Fatal(err)
	}

	files, _ := fs.Glob(checks.FS, "*/*.yaml")
	if len(files) == 0 {
		t.Fatal("no check files found")
	}
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			data, _ := fs.ReadFile(checks.FS, file)
			var raw any
			if err := yaml.Unmarshal(data, &raw); err != nil {
				t.Fatal(err)
			}
			// Round-trip through JSON so the validator sees JSON types.
			js, err := json.Marshal(raw)
			if err != nil {
				t.Fatal(err)
			}
			inst, err := jsonschema.UnmarshalJSON(strings.NewReader(string(js)))
			if err != nil {
				t.Fatal(err)
			}
			if err := schema.Validate(inst); err != nil {
				t.Fatalf("%v", err)
			}
			// The file name is the ID without the platform prefix.
			var c struct{ ID string }
			_ = yaml.Unmarshal(data, &c)
			want := path.Dir(file) + "." + strings.TrimSuffix(path.Base(file), ".yaml")
			if c.ID != want {
				t.Errorf("id %q should be %q to match its file name", c.ID, want)
			}
		})
	}
}

func TestMacOSCatalogLoads(t *testing.T) {
	all, err := check.Load(checks.FS, "macos")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) < 15 {
		t.Fatalf("expected at least 15 macOS checks, got %d", len(all))
	}
	for _, c := range all {
		for _, lang := range []check.Lang{check.EN, check.ES} {
			text := c.TextFor(lang)
			if text.TitlePass == text.TitleFail {
				t.Errorf("%s (%s): pass and fail titles must differ", c.ID, lang)
			}
			if strings.ContainsAny(text.Why+strings.Join(text.FixSteps, ""), "—–") {
				t.Errorf("%s (%s): use ':' or ',' instead of long dashes", c.ID, lang)
			}
		}
	}
}

func TestLoadRejectsBrokenChecks(t *testing.T) {
	tests := map[string]string{
		"no probe": `
id: macos.x
platform: macos
relevance: {default: essential}
probe: {}
text: {en: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d]}, es: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d]}}
references: [https://example.com]`,
		"bad regex": `
id: macos.x
platform: macos
relevance: {default: essential}
probe: {command: [x], rules: [{match: "(", status: pass}]}
text: {en: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d]}, es: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d]}}
references: [https://example.com]`,
		"unknown builtin": `
id: macos.x
platform: macos
relevance: {default: essential}
probe: {builtin: nope}
text: {en: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d]}, es: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d]}}
references: [https://example.com]`,
		"missing spanish": `
id: macos.x
platform: macos
relevance: {default: essential}
probe: {manual: true}
text: {en: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d]}}
references: [https://example.com]`,
		"no references": `
id: macos.x
platform: macos
relevance: {default: essential}
probe: {manual: true}
text: {en: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d]}, es: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d]}}`,
	}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			fsys := memFS(map[string]string{"macos/x.yaml": body})
			if _, err := check.Load(fsys, "macos"); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestLoadRejectsUnsafeFixes(t *testing.T) {
	head := `
id: macos.x
platform: macos
relevance: {default: essential}
probe: {manual: true}
text: {en: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d]}, es: {title: t, title_pass: a, title_fail: b, why: c, fix_steps: [d]}}
references: [https://example.com]
`
	tests := map[string]string{
		"relative command":         `fix: {steps: [{apply: [pmset, x], undo: [/usr/bin/pmset, y]}]}`,
		"no undo":                  `fix: {steps: [{apply: [/usr/bin/pmset, x]}]}`,
		"shell characters":         `fix: {steps: [{apply: [/bin/sh, "x; rm -rf /"], undo: [/bin/true]}]}`,
		"dollar":                   `fix: {steps: [{apply: [/bin/echo, "$HOME"], undo: [/bin/true]}]}`,
		"previous without read":    `fix: {steps: [{apply: [/bin/a], undo: [/bin/b, "{previous}"]}]}`,
		"previous in apply":        `fix: {steps: [{read: [/bin/r], apply: [/bin/a, "{previous}"], undo: [/bin/b]}]}`,
		"read_match without group": `fix: {steps: [{read: [/bin/r], read_match: "x", apply: [/bin/a], undo: [/bin/b]}]}`,
		"no steps":                 `fix: {steps: []}`,
	}
	for name, fix := range tests {
		t.Run(name, func(t *testing.T) {
			fsys := memFS(map[string]string{"macos/x.yaml": head + fix})
			if _, err := check.Load(fsys, "macos"); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
