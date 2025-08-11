package domain

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Runner struct {
	logger *slog.Logger
}

func NewRunner() *Runner {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "octave", "--version")
	err := cmd.Run()
	if err != nil {
		slog.Error("Could not run octave command, make sure it's installed and available in the PATH")
		os.Exit(1)
	}

	return &Runner{
		logger: slog.Default(),
	}
}

func (r *Runner) ExecuteScript(ctx context.Context, script string) (string, error) {
	r.logger.Debug("ExecuteScript started", "script_length", len(script))

	if script == "" {
		r.logger.Warn("ExecuteScript received empty script")
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
		r.logger.Error("ExecuteScript failed", "error", err, "result", result)
		return result, err
	}

	r.logger.Debug("ExecuteScript completed successfully", "result_length", len(result))
	return result, nil
}

func (r *Runner) GeneratePlot(ctx context.Context, script string, format string) ([]byte, error) {
	r.logger.Debug("GeneratePlot started", "script_length", len(script), "format", format)

	// Validate format
	format = strings.ToLower(format)
	if format != "png" && format != "svg" {
		r.logger.Warn("GeneratePlot received unsupported format", "format", format)
		return nil, fmt.Errorf("unsupported format: %s (must be png or svg)", format)
	}

	// Create temp dir
	tempDir, err := os.MkdirTemp("", "octave-plot-*")
	if err != nil {
		r.logger.Error("GeneratePlot failed to create temp dir", "error", err)
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			r.logger.Warn("GeneratePlot failed to clean up temp dir", "error", err, "temp_dir", tempDir)
		}
	}()

	// Setup plot command
	plotFile := filepath.Join(tempDir, "plot."+format)
	wrappedScript := fmt.Sprintf(`
graphics_toolkit("qt");
set(0, "defaultfigurevisible", "off");
%s
print("%s");
`, script, plotFile)

	r.logger.Debug("GeneratePlot executing script", "temp_dir", tempDir, "plot_file", plotFile)

	// Execute
	_, err = r.ExecuteScript(ctx, wrappedScript)
	if err != nil {
		r.logger.Error("GeneratePlot failed to execute script", "error", err)
		return nil, fmt.Errorf("plot generation failed: %w", err)
	}

	// Read plot file
	imgData, err := os.ReadFile(plotFile)
	if err != nil {
		r.logger.Error("GeneratePlot failed to read plot file", "error", err, "plot_file", plotFile)
		return nil, fmt.Errorf("failed to read plot file: %w", err)
	}

	r.logger.Debug("GeneratePlot completed successfully", "image_size", len(imgData))
	return imgData, nil
}
