// Package profile describes who uses a device and how, so that each check
// can be weighted for that person instead of for everyone.
package profile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/XexusCode/hardenme/internal/runner"
)

// Risk is the user's threat level.
type Risk string

const (
	RiskNormal   Risk = "normal"
	RiskElevated Risk = "elevated"
)

// ParseRisk validates a risk level.
func ParseRisk(s string) (Risk, error) {
	switch Risk(strings.ToLower(strings.TrimSpace(s))) {
	case RiskNormal:
		return RiskNormal, nil
	case RiskElevated:
		return RiskElevated, nil
	}
	return "", fmt.Errorf("unknown risk level %q (use normal or elevated)", s)
}

// Profile is how the device is used.
type Profile struct {
	Portable      bool `json:"portable"`
	LeavesHome    bool `json:"leaves_home"`
	Shared        bool `json:"shared"`
	SensitiveData bool `json:"sensitive_data"`
	Risk          Risk `json:"risk"`
}

// Default returns a cautious profile for a device whose portability is known.
func Default(portable bool) Profile {
	return Profile{Portable: portable, LeavesHome: portable, Risk: RiskNormal}
}

// Store persists the profile as JSON.
type Store struct {
	Path string
}

// DefaultStore stores the profile under the user's config directory,
// e.g. ~/.config/hardenme/profile.json (or the OS equivalent).
func DefaultStore() (Store, error) {
	dir, err := configDir()
	if err != nil {
		return Store{}, err
	}
	return Store{Path: filepath.Join(dir, "hardenme", "profile.json")}, nil
}

// PhoneStore stores the phone's profile next to the computer's, in
// profile-phone.json: a phone is used differently from the computer.
func PhoneStore() (Store, error) {
	dir, err := configDir()
	if err != nil {
		return Store{}, err
	}
	return Store{Path: filepath.Join(dir, "hardenme", "profile-phone.json")}, nil
}

// DefaultPhone is the cautious starting profile for a phone: it is portable
// and goes outside.
func DefaultPhone() Profile {
	return Profile{Portable: true, LeavesHome: true, Risk: RiskNormal}
}

func configDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config"), nil
}

// ErrNoProfile means the user has not saved a profile yet.
var ErrNoProfile = errors.New("no saved profile")

// Load reads the saved profile.
func (s Store) Load() (Profile, error) {
	data, err := os.ReadFile(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return Profile{}, ErrNoProfile
	}
	if err != nil {
		return Profile{}, err
	}
	var p Profile
	if err := json.Unmarshal(data, &p); err != nil {
		return Profile{}, fmt.Errorf("reading %s: %w", s.Path, err)
	}
	if p.Risk == "" {
		p.Risk = RiskNormal
	}
	if _, err := ParseRisk(string(p.Risk)); err != nil {
		return Profile{}, fmt.Errorf("reading %s: %w", s.Path, err)
	}
	return p, nil
}

// Save writes the profile, creating the directory if needed.
func (s Store) Save(p Profile) error {
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, append(data, '\n'), 0o600)
}

// DetectPortable guesses whether the machine is a laptop by looking for a
// battery. ok is false when it could not tell.
func DetectPortable(ctx context.Context, r runner.Runner, goos string) (portable, ok bool) {
	switch goos {
	case "darwin":
		out, err := r.Run(ctx, "pmset", "-g", "batt")
		if err != nil || out.ExitCode != 0 {
			return false, false
		}
		return strings.Contains(out.Stdout, "InternalBattery"), true
	default:
		return false, false
	}
}
