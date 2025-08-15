package domain

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	defaultExecTimeoutSeconds = 10
	defaultConcurrencyLimit   = 10
	defaultScriptLenLimit     = 10000
)

type Runner struct {
	logger *slog.Logger
	// semaphore to limit concurrent executions
	semaphore chan struct{}
	version   string
}

// Ensure Runner implements RunnerInterface
var _ RunnerInterface = (*Runner)(nil)

func NewRunner() *Runner {
	ctx := context.Background()
	// Configure timeout for version check (default: 10 seconds)
	versionCheckTimeout := defaultExecTimeoutSeconds
	if timeoutStr := os.Getenv("OCTAVE_SCRIPT_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil && timeout > 0 {
			versionCheckTimeout = timeout
		} else {
			slog.Warn("Invalid OCTAVE_SCRIPT_TIMEOUT, using default", "value", timeoutStr)
		}
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(versionCheckTimeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "octave-cli", "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		slog.Error("Could not run octave command", "error", err)
		os.Exit(1)
	}

	// Extract version
	versionRe := regexp.MustCompile(`version (\d+\.\d+\.\d+)`)
	matches := versionRe.FindStringSubmatch(out.String())
	if len(matches) < 2 {
		slog.Error("Could not parse octave version", "output", out.String())
		os.Exit(1)
	}
	version := matches[1]

	// Configure concurrency limit (default: 10)
	concurrencyLimit := defaultConcurrencyLimit
	if limitStr := os.Getenv("OCTAVE_CONCURRENCY_LIMIT"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			concurrencyLimit = limit
		} else {
			slog.Warn("Invalid OCTAVE_CONCURRENCY_LIMIT, using default", "value", limitStr)
		}
	}

	return &Runner{
		logger: slog.Default(),

		semaphore: make(chan struct{}, concurrencyLimit),
		version:   version,
	}
}

func (r *Runner) ExecuteScript(ctx context.Context, script string) (string, error) {
	// Acquire semaphore to limit concurrent executions
	select {
	case r.semaphore <- struct{}{}:
		// Acquired semaphore
	case <-ctx.Done():
		// Context cancelled while waiting for semaphore
		return "", ctx.Err()
	}
	// Release semaphore when function returns
	defer func() {
		<-r.semaphore
	}()

	r.logger.Debug("ExecuteScript started", "script_length", len(script))

	if script == "" {
		r.logger.Warn("ExecuteScript received empty script")
		return "", fmt.Errorf("script cannot be empty")
	}

	// Validate script for command injection attempts
	if err := validateScript(script); err != nil {
		r.logger.Warn("ExecuteScript received invalid script", "error", err)
		return "", fmt.Errorf("invalid script: %w", err)
	}

	// Sanitize script
	sanitizedScript := sanitizeScript(script)

	// Configure script execution timeout (default: 10 seconds)
	scriptTimeout := defaultExecTimeoutSeconds
	if timeoutStr := os.Getenv("OCTAVE_SCRIPT_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil && timeout > 0 {
			scriptTimeout = timeout
		} else {
			r.logger.Warn("Invalid OCTAVE_SCRIPT_TIMEOUT, using default", "value", timeoutStr)
		}
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(scriptTimeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "octave-cli", "--silent", "--no-window-system", "--eval", sanitizedScript)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := strings.TrimSpace(stdout.String())

	// Filter the output to prevent data leaks
	result = filterOutput(result)

	if err != nil {
		// Also filter stderr output
		stderrOutput := filterOutput(stderr.String())
		result = stderrOutput + "\n" + result
		r.logger.Error("ExecuteScript failed", "error", err, "result", result)
		return result, err
	}

	r.logger.Debug("ExecuteScript completed successfully", "result_length", len(result))
	return result, nil
}

// GetVersion returns the Octave version
func (r *Runner) GetVersion() string {
	return r.version
}

// filterOutput removes potentially sensitive information from the output
func filterOutput(output string) string {
	// Remove file paths that might contain sensitive information
	// This is a simple example, in practice you might want to use more sophisticated filtering
	output = regexp.MustCompile(`/[^:\s]*`).ReplaceAllString(output, "/[REDACTED]")

	// Remove environment variable-like strings
	output = regexp.MustCompile(`[A-Z_][A-Z0-9_]*=[^:\s]*`).ReplaceAllString(output, "[REDACTED]")

	// Remove IP addresses
	output = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`).ReplaceAllString(output, "[IP_ADDRESS]")

	// Remove email addresses
	output = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`).ReplaceAllString(output, "[EMAIL]")

	return output
}

func (r *Runner) GeneratePlot(ctx context.Context, script string, format string) ([]byte, error) {
	// Acquire semaphore to limit concurrent executions
	select {
	case r.semaphore <- struct{}{}:
		// Acquired semaphore
	case <-ctx.Done():
		// Context cancelled while waiting for semaphore
		return nil, ctx.Err()
	}
	// Release semaphore when function returns
	defer func() {
		<-r.semaphore
	}()

	r.logger.Debug("GeneratePlot started", "script_length", len(script), "format", format)

	// Validate format
	format = strings.ToLower(format)
	if format != "png" && format != "svg" {
		r.logger.Warn("GeneratePlot received unsupported format", "format", format)
		return nil, fmt.Errorf("unsupported format: %s (must be png or svg)", format)
	}

	// Validate script for command injection attempts
	if err := validateScript(script); err != nil {
		r.logger.Warn("GeneratePlot received invalid script", "error", err)
		return nil, fmt.Errorf("invalid script: %w", err)
	}

	// Create temp dir
	tempDir, err := os.MkdirTemp("", "octave-plot-*")
	if err != nil {
		r.logger.Error("GeneratePlot failed to create temp dir", "error", err)
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Set restrictive permissions on the temp directory
	if err := os.Chmod(tempDir, 0700); err != nil {
		// Clean up the temp directory before returning error
		os.RemoveAll(tempDir)
		r.logger.Error("GeneratePlot failed to set permissions on temp dir", "error", err)
		return nil, fmt.Errorf("failed to set permissions on temp dir: %w", err)
	}

	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			r.logger.Warn("GeneratePlot failed to clean up temp dir", "error", err, "temp_dir", tempDir)
		}
	}()

	// Setup plot command
	plotFile := filepath.Join(tempDir, "plot."+format)
	wrappedScript := fmt.Sprintf(`
graphics_toolkit("gnuplot");
set(0, "defaultfigurevisible", "off");
%s
print("%s");
`, sanitizeScript(script), plotFile)

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
	// Note: We don't filter imgData as it's binary image data, not text output
	return imgData, nil
}

// validateScript checks if the script contains any potentially dangerous patterns
// that could lead to command injection or other security issues in GNU Octave
// TODO add test cases with examples of actual malicious scripts that would work in GNU Octave
func validateScript(script string) error {
	// Check for command substitution patterns
	if strings.Contains(script, "$(") || strings.Contains(script, "`") {
		return fmt.Errorf("script contains command substitution patterns")
	}

	// Check for shell command execution patterns in Octave
	dangerousFunctions := []string{
		"system(", "exec(", "popen(", // Direct system command execution
		"eval(", "evalin(", // Code execution functions
		"urlread(", "urlwrite(", // Network functions that could be used for data exfiltration
		"load(", "save(", // File I/O functions that could be misused
		"unix(", "dos(", // Platform-specific command execution
		"waitpid(", "fork(", // Process control functions
	}

	for _, function := range dangerousFunctions {
		if strings.Contains(script, function) {
			return fmt.Errorf("script contains potentially dangerous function: %s", function)
		}
	}

	// Check for dangerous shell redirection operators that could be used maliciously
	dangerousPatterns := []string{
		"; rm ",  // Preventing rm commands
		"; del ", // Windows delete
		"| sh",   // Piping to shell
		"| bash", // Piping to bash
		"`",      // Command substitution
		"&&",     // Command chaining
		"||",     // Command chaining
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(script, pattern) {
			return fmt.Errorf("script contains potentially dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// sanitizeScript removes or escapes potentially harmful content from the script
func sanitizeScript(script string) string {
	// Remove null bytes which can be used to terminate strings prematurely
	script = strings.ReplaceAll(script, "\x00", "")

	// Configure script length limit (default: 10000 characters)
	scriptLengthLimit := defaultScriptLenLimit
	if limitStr := os.Getenv("OCTAVE_SCRIPT_LENGTH_LIMIT"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			scriptLengthLimit = limit
		} else {
			slog.Warn("Invalid OCTAVE_SCRIPT_LENGTH_LIMIT, using default", "value", limitStr)
		}
	}
	// Limit script length to prevent resource exhaustion
	if len(script) > scriptLengthLimit {
		script = script[:scriptLengthLimit]
	}

	return script
}
