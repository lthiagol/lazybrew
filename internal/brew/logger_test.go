package brew

import (
	"testing"
)

func TestSetDebug(t *testing.T) {
	SetDebug(true)
	l := Logger()
	if l == nil {
		t.Fatal("Logger should not be nil after SetDebug(true)")
	}

	SetDebug(false)
	l2 := Logger()
	if l2 == nil {
		t.Fatal("Logger should not be nil after SetDebug(false)")
	}
}

func TestBrewNotFoundError(t *testing.T) {
	err := &BrewNotFoundError{Searched: []string{"/opt/homebrew/bin/brew"}}
	msg := err.Error()
	if msg == "" {
		t.Error("error message should not be empty")
	}
}

func TestBrewExitError(t *testing.T) {
	err := &BrewExitError{Command: "install", ExitCode: 1, Stderr: "not found"}
	msg := err.Error()
	if msg == "" {
		t.Error("error message should not be empty")
	}
}

func TestJSONParseError(t *testing.T) {
	inner := &BrewNotFoundError{}
	err := &JSONParseError{Command: "info", Cause: inner, RawOutput: []byte("bad json")}
	if err.Unwrap() != inner {
		t.Error("Unwrap should return the cause")
	}
	if err.Error() == "" {
		t.Error("error message should not be empty")
	}
}

func TestTimeoutError(t *testing.T) {
	err := &TimeoutError{Command: "search"}
	msg := err.Error()
	if msg == "" {
		t.Error("error message should not be empty")
	}
}
