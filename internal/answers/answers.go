// Package answers remembers what a person answered in a phone checklist, so
// the next run asks again only what was not right.
package answers

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/XexusCode/hardenme/internal/check"
)

// Saved is one answer and when it was given.
type Saved struct {
	Answer check.Answer `json:"answer"`
	Time   time.Time    `json:"time"`
}

// Store keeps the answers of one platform in a JSON file.
type Store struct{ Path string }

// DefaultStore is ~/.local/state/hardenme/answers-<platform>.json, or the
// same file under $XDG_STATE_HOME.
func DefaultStore(platform string) (Store, error) {
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Store{}, err
		}
		dir = filepath.Join(home, ".local", "state")
	}
	return Store{Path: filepath.Join(dir, "hardenme", "answers-"+platform+".json")}, nil
}

// Load returns the saved answers by check ID. A missing file is empty.
func (s Store) Load() (map[string]Saved, error) {
	data, err := os.ReadFile(s.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]Saved{}, nil
	}
	if err != nil {
		return nil, err
	}
	saved := map[string]Saved{}
	if err := json.Unmarshal(data, &saved); err != nil {
		return nil, err
	}
	return saved, nil
}

// Save writes the answers. Only yes, no and dont_know are stored.
func (s Store) Save(saved map[string]Saved) error {
	keep := map[string]Saved{}
	for id, a := range saved {
		switch a.Answer {
		case check.Yes, check.No, check.DontKnow:
			keep[id] = a
		}
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(keep, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, append(data, '\n'), 0o600)
}

// Reusable reports whether a saved answer can stand without asking again:
// only a "yes" does. A "no" or "I don't know" is asked again, because the
// person may have changed the setting since.
func Reusable(a Saved, ok bool) bool { return ok && a.Answer == check.Yes }
