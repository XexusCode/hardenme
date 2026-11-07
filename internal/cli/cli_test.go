package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/profile"
)

func TestAskProfile(t *testing.T) {
	in := strings.NewReader("\nn\ns\nyes\nmaybe\nno\n")
	var out bytes.Buffer
	got, err := askProfile(profile.Default(true), check.ES, in, &out)
	if err != nil {
		t.Fatal(err)
	}
	// Enter keeps the detected default; "maybe" is asked again.
	want := profile.Profile{Portable: true, LeavesHome: false, Shared: true, SensitiveData: true, Risk: profile.RiskNormal}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if !strings.Contains(out.String(), "¿Este dispositivo es un portátil? [S/n]") {
		t.Errorf("questions should be in Spanish with the default shown:\n%s", out.String())
	}
}

func TestAskProfileKeepsDefaultsOnEOF(t *testing.T) {
	start := profile.Profile{Portable: true, LeavesHome: true, Risk: profile.RiskElevated}
	got, err := askProfile(start, check.EN, strings.NewReader(""), &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if got != start {
		t.Fatalf("got %+v, want %+v", got, start)
	}
}

func TestFilterChecksAcceptsShortIDs(t *testing.T) {
	all := []check.Check{{ID: "macos.filevault"}, {ID: "macos.firewall"}, {ID: "macos.firewall_stealth"}}
	got, err := filterChecks(all, "filevault, macos.firewall")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "macos.filevault" || got[1].ID != "macos.firewall" {
		t.Fatalf("got %v", got)
	}
	if _, err := filterChecks(all, "nope"); err == nil {
		t.Fatal("expected an error for an unknown check")
	}
}

func TestPickLang(t *testing.T) {
	env := func(m map[string]string) func(string) string {
		return func(k string) string { return m[k] }
	}
	cases := []struct {
		name   string
		env    map[string]string
		system string
		want   string
	}{
		{"LANG Spanish", map[string]string{"LANG": "es_ES.UTF-8"}, "en-US", "es"},
		{"LC_ALL wins", map[string]string{"LC_ALL": "en_US.UTF-8", "LANG": "es_ES.UTF-8"}, "es-ES", "en"},
		{"no env, macOS Spanish", nil, "es-ES", "es"},
		{"C locale falls back", map[string]string{"LANG": "C"}, "es-419", "es"},
		{"nothing", nil, "", "en"},
	}
	for _, c := range cases {
		if got := pickLang(env(c.env), func() string { return c.system }); got != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}

func TestFirstAppleLanguage(t *testing.T) {
	cases := map[string]string{
		"(\n    \"es-ES\",\n    en\n)\n": "es-ES",
		"(\n    \"en-GB\"\n)\n":          "en-GB",
		"(\n    en\n)":                   "en",
		"":                               "",
	}
	for in, want := range cases {
		if got := firstAppleLanguage(in); got != want {
			t.Errorf("firstAppleLanguage(%q) = %q, want %q", in, got, want)
		}
	}
}
