package domain_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fmcato/octave-mcp/internal/domain"
)

func TestGeneratePlot(t *testing.T) {
	runner := domain.NewRunner()
	ctx := context.Background()

	t.Run("PNG output", func(t *testing.T) {
		testPlotGeneration(t, runner, ctx, "png")
	})

	t.Run("SVG output", func(t *testing.T) {
		testPlotGeneration(t, runner, ctx, "svg")
	})

	t.Run("Invalid format", func(t *testing.T) {
		_, err := runner.GeneratePlot(ctx, "plot([1,2,3]);", "jpg")
		if err == nil {
			t.Fatal("Expected error for invalid format")
		}
		expected := "unsupported format: jpg (must be png or svg)"
		if err.Error() != expected {
			t.Errorf("Expected error: %s, got: %s", expected, err.Error())
		}
	})

	t.Run("Invalid script", func(t *testing.T) {
		_, err := runner.GeneratePlot(ctx, "invalid octave script", "png")
		if err == nil {
			t.Fatal("Expected error for invalid script")
		}
		if !strings.Contains(err.Error(), "plot generation failed") {
			t.Errorf("Expected plot generation error, got: %s", err.Error())
		}
	})

	t.Run("Temp file cleanup", func(t *testing.T) {
		// Count existing temp directories
		beforeDirs, _ := filepath.Glob("/tmp/octave-plot-*")
		beforeCount := len(beforeDirs)

		// Generate plot
		_, err := runner.GeneratePlot(ctx, "plot([1,2,3]);", "png")
		if err != nil {
			t.Fatal(err)
		}

		// Count temp directories after
		afterDirs, _ := filepath.Glob("/tmp/octave-plot-*")
		afterCount := len(afterDirs)

		// Should have same number or less (cleanup might be async)
		if afterCount > beforeCount {
			t.Errorf("Temp dir not cleaned up: before %d, after %d", beforeCount, afterCount)
		}
	})

	t.Run("Script with post-plot commands", func(t *testing.T) {
		script := `plot([1,2,3,4]);
xlabel('X Axis');
ylabel('Y Axis');
title('Test Plot');
legend('Data');
grid on;`
		imgData, err := runner.GeneratePlot(ctx, script, "png")
		if err != nil {
			t.Fatal(err)
		}

		validateImgData(t, imgData, "png")
	})
}

func testPlotGeneration(t *testing.T, runner *domain.Runner, ctx context.Context, format string) {
	imgData, err := runner.GeneratePlot(ctx, "plot([1,2,3,4]);", format)
	if err != nil {
		t.Fatal(err)
	}

	validateImgData(t, imgData, format)
}

func validateImgData(t *testing.T, imgData []byte, format string) {
	// Verify image size (1MB max)
	if len(imgData) > 1024*1024 {
		t.Errorf("%s image size exceeds 1MB: %d bytes", format, len(imgData))
	}

	// Verify non-zero size
	if len(imgData) == 0 {
		t.Errorf("%s image data is empty", format)
	}

	// Verify basic header
	if format == "png" && string(imgData[1:4]) != "PNG" {
		t.Errorf("Invalid PNG header: %x", imgData[:4])
	}
	if format == "svg" && !strings.Contains(string(imgData[:100]), "<svg") {
		t.Errorf("Invalid SVG header: %s", string(imgData[:100]))
	}
}
