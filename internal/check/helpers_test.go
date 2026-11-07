package check_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/XexusCode/hardenme/checks"
	"github.com/XexusCode/hardenme/internal/check"
	"github.com/XexusCode/hardenme/internal/profile"
	"github.com/XexusCode/hardenme/internal/runner"
)

// memFS builds an in-memory file system from file contents.
func memFS(files map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for name, body := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(body)}
	}
	return fsys
}

// out builds a runner output.
func out(stdout string, exit int) runner.Output { return runner.Output{Stdout: stdout, ExitCode: exit} }

// errOut builds a runner output printed on stderr.
func errOut(stderr string, exit int) runner.Output {
	return runner.Output{Stderr: stderr, ExitCode: exit}
}

// macCheck returns a real macOS check from the embedded catalog.
func macCheck(t *testing.T, id string) check.Check {
	t.Helper()
	all, err := check.Load(checks.FS, "macos")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range all {
		if c.ID == id {
			return c
		}
	}
	t.Fatalf("check %s not found", id)
	return check.Check{}
}

// runWith runs a check against canned outputs.
func runWith(t *testing.T, c check.Check, p profile.Profile, outputs map[string]runner.Output) check.Result {
	t.Helper()
	fake := &runner.Fake{Outputs: outputs}
	return c.Run(context.Background(), check.Env{Runner: fake, Profile: p})
}
