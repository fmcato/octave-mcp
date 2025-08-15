package domain

import (
	"context"
)

// RunnerInterface defines the interface for executing Octave scripts
type RunnerInterface interface {
	ExecuteScript(ctx context.Context, script string) (string, error)
	GeneratePlot(ctx context.Context, script string, format string) ([]byte, error)
	GetVersion() string
}
