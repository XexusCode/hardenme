package profile_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/runner"
)

func TestStoreRoundTrip(t *testing.T) {
	store := profile.Store{Path: filepath.Join(t.TempDir(), "hardenme", "profile.json")}
	want := profile.Profile{Portable: true, LeavesHome: true, Shared: true, SensitiveData: true, Risk: profile.RiskElevated}
	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	info, err := os.Stat(store.Path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("profile file should be private, got %v", info.Mode().Perm())
	}
}

func TestLoadWithoutProfile(t *testing.T) {
	store := profile.Store{Path: filepath.Join(t.TempDir(), "missing.json")}
	if _, err := store.Load(); !errors.Is(err, profile.ErrNoProfile) {
		t.Fatalf("got %v, want ErrNoProfile", err)
	}
}

func TestLoadRejectsUnknownRisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")
	if err := os.WriteFile(path, []byte(`{"risk":"extreme"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := (profile.Store{Path: path}).Load(); err == nil {
		t.Fatal("expected an error")
	}
}

func TestLoadDefaultsMissingRiskToNormal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "profile.json")
	if err := os.WriteFile(path, []byte(`{"portable":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	p, err := (profile.Store{Path: path}).Load()
	if err != nil {
		t.Fatal(err)
	}
	if p.Risk != profile.RiskNormal {
		t.Fatalf("risk = %q, want normal", p.Risk)
	}
}

func TestDefaultIsCautiousForLaptops(t *testing.T) {
	if p := profile.Default(true); !p.LeavesHome {
		t.Error("a laptop should be assumed to leave home")
	}
	if p := profile.Default(false); p.LeavesHome {
		t.Error("a desktop should be assumed to stay home")
	}
}

func TestDetectPortable(t *testing.T) {
	key := runner.Key("pmset", "-g", "batt")
	tests := []struct {
		name         string
		outputs      map[string]runner.Output
		goos         string
		wantPortable bool
		wantOK       bool
	}{
		// Captured on a MacBook running macOS 26.2.
		{"macbook", map[string]runner.Output{key: {Stdout: "Now drawing from 'AC Power'\n -InternalBattery-0 (id=7077987)\t100%; charged; 0:00 remaining present: true\n"}}, "darwin", true, true},
		{"mac mini", map[string]runner.Output{key: {Stdout: "Now drawing from 'AC Power'\n"}}, "darwin", false, true},
		{"pmset missing", nil, "darwin", false, false},
		{"other os", nil, "linux", false, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &runner.Fake{Outputs: tc.outputs}
			portable, ok := profile.DetectPortable(context.Background(), fake, tc.goos)
			if portable != tc.wantPortable || ok != tc.wantOK {
				t.Fatalf("got (%v, %v), want (%v, %v)", portable, ok, tc.wantPortable, tc.wantOK)
			}
		})
	}
}
