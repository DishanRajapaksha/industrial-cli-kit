// Package safety standardizes confirmation for mutating CLI operations.
package safety

import "fmt"

// Mode describes whether a mutating operation may be transmitted.
type Mode int

const (
	DryRun Mode = iota
	Execute
)

// Resolve returns DryRun by default. --yes enables Execute; --dry-run makes
// the intent explicit. Supplying both is rejected.
func Resolve(yes bool, dryRun bool) (Mode, error) {
	if yes && dryRun {
		return DryRun, fmt.Errorf("--yes and --dry-run cannot be used together")
	}
	if yes {
		return Execute, nil
	}
	return DryRun, nil
}

func (m Mode) String() string {
	if m == Execute {
		return "execute"
	}
	return "dry-run"
}
