package answers_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/XexusCode/hardenme/internal/answers"
	"github.com/XexusCode/hardenme/internal/check"
)

func TestRoundTripAndMode(t *testing.T) {
	s := answers.Store{Path: filepath.Join(t.TempDir(), "sub", "answers-ios.json")}
	empty, err := s.Load()
	if err != nil || len(empty) != 0 {
		t.Fatalf("missing file: %v %v", empty, err)
	}
	now := time.Date(2026, 11, 8, 10, 0, 0, 0, time.UTC)
	in := map[string]answers.Saved{
		"ios.a": {Answer: check.Yes, Time: now},
		"ios.b": {Answer: check.No, Time: now},
		"ios.c": {Answer: check.DontKnow, Time: now},
		"ios.d": {Answer: check.Quit, Time: now},
	}
	if err := s.Save(in); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(s.Path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode %v, want 0600", info.Mode().Perm())
	}
	got, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got["ios.b"].Answer != check.No || !got["ios.a"].Time.Equal(now) {
		t.Errorf("got %v", got)
	}
	if _, ok := got["ios.d"]; ok {
		t.Error("quit should not be stored")
	}
}

func TestOnlyYesIsReused(t *testing.T) {
	for answer, want := range map[check.Answer]bool{check.Yes: true, check.No: false, check.DontKnow: false} {
		if got := answers.Reusable(answers.Saved{Answer: answer}, true); got != want {
			t.Errorf("%s: %v, want %v", answer, got, want)
		}
	}
	if answers.Reusable(answers.Saved{Answer: check.Yes}, false) {
		t.Error("a missing answer is not reusable")
	}
}

func TestDefaultStoreHonoursXDG(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/x")
	s, err := answers.DefaultStore("android")
	if err != nil {
		t.Fatal(err)
	}
	if s.Path != "/x/hardenme/answers-android.json" {
		t.Errorf("path %s", s.Path)
	}
}
