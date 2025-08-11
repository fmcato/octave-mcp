package mocks

import (
	"context"
)

// MockRunner implements domain.RunnerInterface for testing
type MockRunner struct {
	ExecuteScriptFunc func(ctx context.Context, script string) (string, error)
	GeneratePlotFunc  func(ctx context.Context, script string, format string) ([]byte, error)
}

// ExecuteScript calls the mock function if set, otherwise returns empty string and nil error
func (m *MockRunner) ExecuteScript(ctx context.Context, script string) (string, error) {
	if m.ExecuteScriptFunc != nil {
		return m.ExecuteScriptFunc(ctx, script)
	}
	return "", nil
}

// GeneratePlot calls the mock function if set, otherwise returns empty byte slice and nil error
func (m *MockRunner) GeneratePlot(ctx context.Context, script string, format string) ([]byte, error) {
	if m.GeneratePlotFunc != nil {
		return m.GeneratePlotFunc(ctx, script, format)
	}
	return []byte{}, nil
}
