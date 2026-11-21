// Package brewtest holds the test double for brew.Runner. It has no imports
// from the brew package, so brew's own tests can use it too.
package brewtest

import (
	"context"
	"encoding/json"
	"errors"
)

// ErrTest is the error tests inject where any failure will do.
var ErrTest = errors.New("test error")

// MockRunner is a brew.Runner. Each test sets the behaviour it needs through
// the Fn fields; a nil Fn falls back to an empty result.
type MockRunner struct {
	ExecuteFn       func(ctx context.Context, args ...string) ([]byte, error)
	ExecuteJSONFn   func(ctx context.Context, result any, args ...string) error
	ExecuteStreamFn func(ctx context.Context, args ...string) (<-chan string, <-chan error)
	BrewPathFn      func() string
}

func (m *MockRunner) Execute(ctx context.Context, args ...string) ([]byte, error) {
	if m.ExecuteFn != nil {
		return m.ExecuteFn(ctx, args...)
	}
	return []byte{}, nil
}

func (m *MockRunner) ExecuteJSON(ctx context.Context, result any, args ...string) error {
	if m.ExecuteJSONFn != nil {
		return m.ExecuteJSONFn(ctx, result, args...)
	}
	stdout, err := m.Execute(ctx, args...)
	if err != nil {
		return err
	}
	return json.Unmarshal(stdout, result)
}

func (m *MockRunner) ExecuteStream(ctx context.Context, args ...string) (<-chan string, <-chan error) {
	if m.ExecuteStreamFn != nil {
		return m.ExecuteStreamFn(ctx, args...)
	}
	ch := make(chan string, 1)
	errCh := make(chan error, 1)
	close(ch)
	close(errCh)
	return ch, errCh
}

func (m *MockRunner) BrewPath() string {
	if m.BrewPathFn != nil {
		return m.BrewPathFn()
	}
	return "/fake/brew"
}

func NewMockRunner() *MockRunner {
	return &MockRunner{}
}
