package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fmcato/octave-mcp/internal/domain/mocks"
)

func TestGeneratePlot(t *testing.T) {
	ctx := context.Background()

	t.Run("Valid PNG format", func(t *testing.T) {
		mockRunner := &mocks.MockRunner{
			GeneratePlotFunc: func(ctx context.Context, script string, format string) ([]byte, error) {
				return []byte("mock png data"), nil
			},
		}

		_, err := mockRunner.GeneratePlot(ctx, "plot([1,2,3]);", "png")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Valid SVG format", func(t *testing.T) {
		mockRunner := &mocks.MockRunner{
			GeneratePlotFunc: func(ctx context.Context, script string, format string) ([]byte, error) {
				return []byte("<svg>mock svg data</svg>"), nil
			},
		}

		_, err := mockRunner.GeneratePlot(ctx, "plot([1,2,3]);", "svg")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	})

	t.Run("Invalid format", func(t *testing.T) {
		mockRunner := &mocks.MockRunner{
			GeneratePlotFunc: func(ctx context.Context, script string, format string) ([]byte, error) {
				return nil, errors.New("unsupported format: jpg (must be png or svg)")
			},
		}

		_, err := mockRunner.GeneratePlot(ctx, "plot([1,2,3]);", "jpg")
		if err == nil {
			t.Fatal("Expected error for invalid format")
		}
		expected := "unsupported format: jpg (must be png or svg)"
		if err.Error() != expected {
			t.Errorf("Expected error: %s, got: %s", expected, err.Error())
		}
	})

	t.Run("Empty script", func(t *testing.T) {
		mockRunner := &mocks.MockRunner{
			GeneratePlotFunc: func(ctx context.Context, script string, format string) ([]byte, error) {
				return nil, errors.New("script cannot be empty")
			},
		}

		_, err := mockRunner.GeneratePlot(ctx, "", "png")
		if err == nil {
			t.Fatal("Expected error for empty script")
		}
		expected := "script cannot be empty"
		if err.Error() != expected {
			t.Errorf("Expected error: %s, got: %s", expected, err.Error())
		}
	})
}

func TestExecuteScript(t *testing.T) {
	ctx := context.Background()

	t.Run("Valid script", func(t *testing.T) {
		mockRunner := &mocks.MockRunner{
			ExecuteScriptFunc: func(ctx context.Context, script string) (string, error) {
				return "ans =  6", nil
			},
		}

		result, err := mockRunner.ExecuteScript(ctx, "x = 2 + 4")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if result != "ans =  6" {
			t.Errorf("Expected 'ans =  6', got: %s", result)
		}
	})

	t.Run("Invalid script", func(t *testing.T) {
		mockRunner := &mocks.MockRunner{
			ExecuteScriptFunc: func(ctx context.Context, script string) (string, error) {
				return "error: some error", errors.New("execution failed")
			},
		}

		_, err := mockRunner.ExecuteScript(ctx, "invalid script")
		if err == nil {
			t.Fatal("Expected error for invalid script")
		}
	})

	t.Run("Empty script", func(t *testing.T) {
		mockRunner := &mocks.MockRunner{
			ExecuteScriptFunc: func(ctx context.Context, script string) (string, error) {
				return "", errors.New("script cannot be empty")
			},
		}

		_, err := mockRunner.ExecuteScript(ctx, "")
		if err == nil {
			t.Fatal("Expected error for empty script")
		}
		expected := "script cannot be empty"
		if err.Error() != expected {
			t.Errorf("Expected error: %s, got: %s", expected, err.Error())
		}
	})
}
