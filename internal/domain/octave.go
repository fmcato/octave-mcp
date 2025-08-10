package domain

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Runner struct{}

func NewRunner() *Runner {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "octave", "--version")
	err := cmd.Run()
	if err != nil {
		log.Fatal("Could not run octave command, make sure it's installed and available in the PATH")
	}

	return &Runner{}
}

func (r *Runner) ExecuteScript(ctx context.Context, script string) (string, error) {
	if script == "" {
		return "", fmt.Errorf("script cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "octave", "--eval", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := strings.TrimSpace(stdout.String())

	if err != nil {
		result = stderr.String() + "\n" + result
		return result, err
	}

	return result, nil
}

func (r *Runner) GeneratePlot(ctx context.Context, script string, format string) ([]byte, error) {
	// Validate format
	format = strings.ToLower(format)
	if format != "png" && format != "svg" {
		return nil, fmt.Errorf("unsupported format: %s (must be png or svg)", format)
	}

	// Create temp dir
	tempDir, err := os.MkdirTemp("", "octave-plot-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup plot command
	plotFile := filepath.Join(tempDir, "plot."+format)
	wrappedScript := fmt.Sprintf(`
graphics_toolkit("qt");
set(0, "defaultfigurevisible", "off");
%s
print("%s");
`, script, plotFile)

	// Execute
	_, err = r.ExecuteScript(ctx, wrappedScript)
	if err != nil {
		return nil, fmt.Errorf("plot generation failed: %w", err)
	}

	// Read plot file
	imgData, err := os.ReadFile(plotFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read plot file: %w", err)
	}

	return imgData, nil
}
