package fix

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// Change is one applied step and the command that reverts it.
type Change struct {
	CheckID string   `json:"check_id"`
	Admin   bool     `json:"admin"`
	Undo    []string `json:"undo"`
}

// Batch is every change made by one `hardenme fix` run.
type Batch struct {
	Time    time.Time `json:"time"`
	Changes []Change  `json:"changes"`
}

// Log stores the batches, oldest first, in a JSON file.
type Log struct{ Path string }

// DefaultLog is ~/.local/state/hardenme/changes.json, or the same file
// under $XDG_STATE_HOME.
func DefaultLog() (Log, error) {
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Log{}, err
		}
		dir = filepath.Join(home, ".local", "state")
	}
	return Log{Path: filepath.Join(dir, "hardenme", "changes.json")}, nil
}

func (l Log) load() ([]Batch, error) {
	data, err := os.ReadFile(l.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var batches []Batch
	if err := json.Unmarshal(data, &batches); err != nil {
		return nil, err
	}
	return batches, nil
}

func (l Log) save(batches []Batch) error {
	if err := os.MkdirAll(filepath.Dir(l.Path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(batches, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(l.Path, append(data, '\n'), 0o600)
}

// Append adds a batch. Empty batches are not stored.
func (l Log) Append(b Batch) error {
	if len(b.Changes) == 0 {
		return nil
	}
	batches, err := l.load()
	if err != nil {
		return err
	}
	return l.save(append(batches, b))
}

// Last returns the newest batch.
func (l Log) Last() (Batch, bool, error) {
	batches, err := l.load()
	if err != nil || len(batches) == 0 {
		return Batch{}, false, err
	}
	return batches[len(batches)-1], true, nil
}

// DropLast removes the newest batch, after it has been undone.
func (l Log) DropLast() error {
	batches, err := l.load()
	if err != nil || len(batches) == 0 {
		return err
	}
	return l.save(batches[:len(batches)-1])
}
