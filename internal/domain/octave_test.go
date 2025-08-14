package domain_test

import (
	"context"
	"errors"
	"strings"
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

func TestValidateScript_MaliciousCases(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		script  string
		wantErr string
	}{
		{
			name:    "command substitution with backticks",
			script:  "`rm -rf /`",
			wantErr: "invalid script: script contains command substitution patterns",
		},
		{
			name:    "command substitution with $()",
			script:  "$(rm -rf /)",
			wantErr: "invalid script: script contains command substitution patterns",
		},
		{
			name:    "dangerous function system",
			script:  "system('rm -rf /')",
			wantErr: "invalid script: script contains potentially dangerous function: system(",
		},
		{
			name:    "dangerous function exec",
			script:  "exec('shutdown now')",
			wantErr: "invalid script: script contains potentially dangerous function: exec(",
		},
		{
			name:    "dangerous function popen",
			script:  "popen('cat /etc/passwd')",
			wantErr: "invalid script: script contains potentially dangerous function: popen(",
		},
		{
			name:    "dangerous pattern ; rm",
			script:  "x=1; rm -rf /",
			wantErr: "invalid script: script contains potentially dangerous pattern: ; rm ",
		},
		{
			name:    "dangerous pattern | sh",
			script:  "echo 'malicious' | sh",
			wantErr: "invalid script: script contains potentially dangerous pattern: | sh",
		},
		{
			name:    "bypass attempt - mixed case",
			script:  "SyStEm('ls')",
			wantErr: "invalid script: script contains potentially dangerous function: system(",
		},
		{
			name:    "bypass attempt - whitespace",
			script:  "system\t('ls')",
			wantErr: "invalid script: script contains potentially dangerous function: system\t(",
		},
		{
			name:    "bypass attempt - comment injection",
			script:  "# system('safe'); \n system('rm -rf /')",
			wantErr: "invalid script: script contains potentially dangerous function: system(",
		},
		{
			name:    "safe script should pass",
			script:  "x = 1 + 1;",
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := &mocks.MockRunner{
				ExecuteScriptFunc: func(ctx context.Context, script string) (string, error) {
					if tt.wantErr != "" {
						return "", errors.New(tt.wantErr)
					}
					return "validation passed", nil
				},
			}

			_, err := mockRunner.ExecuteScript(ctx, tt.script)
			if tt.wantErr == "" && err != nil {
				t.Errorf("Expected no error for safe script but got: %v", err)
			}
			if tt.wantErr != "" {
				if err == nil {
					t.Error("Expected error for malicious script but got none")
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Expected error containing '%s', got: '%s'", tt.wantErr, err.Error())
				}
			}
		})
	}
}
