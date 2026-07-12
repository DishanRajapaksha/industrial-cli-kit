// Package contracttest provides reusable checks for the shared CLI contract.
package contracttest

import (
	"strings"

	"github.com/DishanRajapaksha/industrial-cli-kit/exitcode"
)

// Result is the observable outcome of a CLI invocation.
type Result struct {
	Code   int
	Stdout string
	Stderr string
}

// Run invokes a CLI without starting a subprocess.
type Run func(args ...string) Result

// TestingT is the subset of testing.TB used by this package.
type TestingT interface {
	Helper()
	Errorf(format string, args ...any)
}

// Baseline verifies the help and version aliases required by contract v1.
func Baseline(t TestingT, run Run) {
	t.Helper()
	for _, args := range [][]string{{"help"}, {"--help"}, {"-h"}, {"version"}, {"--version"}, {"-v"}} {
		result := run(args...)
		if result.Code != int(exitcode.Success) {
			t.Errorf("%s exit code = %d, want 0; stderr=%s", strings.Join(args, " "), result.Code, result.Stderr)
		}
		if strings.TrimSpace(result.Stdout) == "" {
			t.Errorf("%s produced no stdout", strings.Join(args, " "))
		}
	}
}

// RequireExitCode reports an unexpected process exit status.
func RequireExitCode(t TestingT, result Result, want exitcode.Code) {
	t.Helper()
	if result.Code != int(want) {
		t.Errorf("exit code = %d, want %d; stdout=%s stderr=%s", result.Code, want, result.Stdout, result.Stderr)
	}
}
