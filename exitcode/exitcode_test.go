package exitcode

import (
	"errors"
	"fmt"
	"testing"
)

func TestContractValues(t *testing.T) {
	got := []Code{Success, General, Config, Connection, Request, AuthSecurity, ResourceMissing, Rejected, Timeout, Output}
	for want, code := range got {
		if int(code) != want {
			t.Fatalf("code %d = %d", want, code)
		}
	}
}

func TestFromWrappedError(t *testing.T) {
	err := fmt.Errorf("operation: %w", Wrap(Timeout, errors.New("deadline exceeded")))
	if got := From(err); got != Timeout {
		t.Fatalf("From() = %d, want %d", got, Timeout)
	}
}
