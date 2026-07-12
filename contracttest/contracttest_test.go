package contracttest

import (
	"testing"

	"github.com/DishanRajapaksha/industrial-cli-kit/exitcode"
)

func TestBaseline(t *testing.T) {
	run := func(args ...string) Result {
		return Result{Code: int(exitcode.Success), Stdout: "ok\n"}
	}
	Baseline(t, run)
}
