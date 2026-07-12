// Package exitcode defines the process-exit contract shared by industrial CLIs.
package exitcode

import "errors"

// Code is a stable process exit status.
type Code int

const (
	Success         Code = 0
	General         Code = 1
	Config          Code = 2
	Connection      Code = 3
	Request         Code = 4
	AuthSecurity    Code = 5
	ResourceMissing Code = 6
	Rejected        Code = 7
	Timeout         Code = 8
	Output          Code = 9
)

// Coder is implemented by errors that carry a specific exit status.
type Coder interface {
	ExitCode() Code
}

// Error associates an error with a shared process exit status.
type Error struct {
	Code Code
	Err  error
}

func (e *Error) Error() string { return e.Err.Error() }

func (e *Error) Unwrap() error { return e.Err }

func (e *Error) ExitCode() Code { return e.Code }

// Wrap returns an error that maps to code. A nil error remains nil.
func Wrap(code Code, err error) error {
	if err == nil {
		return nil
	}
	return &Error{Code: code, Err: err}
}

// From returns the first shared exit status found in err's wrapping chain.
// Unknown errors are general failures.
func From(err error) Code {
	if err == nil {
		return Success
	}
	var coded Coder
	if errors.As(err, &coded) {
		return coded.ExitCode()
	}
	return General
}
