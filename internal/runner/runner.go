// Package runner executes system commands behind an interface so that
// checks can be tested with captured outputs instead of a real machine.
package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Output is what a command produced.
type Output struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Combined returns stdout followed by stderr. Several macOS tools
// (sysadminctl, defaults) print their answer on stderr.
func (o Output) Combined() string {
	switch {
	case o.Stderr == "":
		return o.Stdout
	case o.Stdout == "":
		return o.Stderr
	default:
		return o.Stdout + "\n" + o.Stderr
	}
}

// ErrNotFound is returned when the command does not exist on this system.
var ErrNotFound = errors.New("command not found")

// Runner runs a command. A non-zero exit code is NOT an error: it is reported
// in Output.ExitCode. Errors are reserved for commands that could not run.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) (Output, error)
}

// Exec runs real commands with a per-command timeout.
type Exec struct {
	Timeout time.Duration
}

// Run implements Runner.
func (e Exec) Run(ctx context.Context, name string, args ...string) (Output, error) {
	timeout := e.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := Output{Stdout: stdout.String(), Stderr: stderr.String()}

	var exitErr *exec.ExitError
	switch {
	case err == nil:
		return out, nil
	case errors.As(err, &exitErr):
		out.ExitCode = exitErr.ExitCode()
		return out, nil
	case errors.Is(err, exec.ErrNotFound):
		return out, fmt.Errorf("%s: %w", name, ErrNotFound)
	case ctx.Err() != nil:
		return out, fmt.Errorf("%s: timed out after %s", name, timeout)
	default:
		return out, err
	}
}

// Key builds the lookup key used by Fake for a command line.
func Key(name string, args ...string) string {
	return strings.TrimSpace(name + " " + strings.Join(args, " "))
}

// Fake returns canned outputs keyed by the full command line.
// Commands without a canned output behave as "not found".
type Fake struct {
	Outputs map[string]Output
	Errors  map[string]error
	Calls   []string
}

// Run implements Runner.
func (f *Fake) Run(_ context.Context, name string, args ...string) (Output, error) {
	key := Key(name, args...)
	f.Calls = append(f.Calls, key)
	if err, ok := f.Errors[key]; ok {
		return Output{}, err
	}
	if out, ok := f.Outputs[key]; ok {
		return out, nil
	}
	return Output{}, fmt.Errorf("%s: %w", name, ErrNotFound)
}
