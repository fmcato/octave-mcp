package octave

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
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
