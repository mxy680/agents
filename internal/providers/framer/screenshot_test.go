package framer

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestScreenshotTake(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	screenshotCmd := newScreenshotCmd(factory)
	root.AddCommand(screenshotCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"screenshot", "take", "--node-id", "node-abc"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	// Without --output the command prints JSON with image/url data
	if !strings.Contains(output, "framer.com") {
		t.Errorf("expected 'framer.com' in output, got: %s", output)
	}
}

func TestScreenshotTakeWithOutput(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	screenshotCmd := newScreenshotCmd(factory)
	root.AddCommand(screenshotCmd)

	// The mock returns base64 "iVBORw0KGgo=" which decodes to a short PNG header
	outputPath := filepath.Join(t.TempDir(), "shot.png")

	output := captureStdout(t, func() {
		root.SetArgs([]string{"screenshot", "take", "--node-id", "node-abc", "--output", outputPath})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Screenshot saved") {
		t.Errorf("expected 'Screenshot saved' in output, got: %s", output)
	}
}

func TestScreenshotExportSVGWithOutput(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	screenshotCmd := newScreenshotCmd(factory)
	root.AddCommand(screenshotCmd)

	outputPath := filepath.Join(t.TempDir(), "export.svg")

	output := captureStdout(t, func() {
		root.SetArgs([]string{"screenshot", "export-svg", "--node-id", "node-abc", "--output", outputPath})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "SVG saved") {
		t.Errorf("expected 'SVG saved' in output, got: %s", output)
	}
}

func TestScreenshotExportSVG(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	screenshotCmd := newScreenshotCmd(factory)
	root.AddCommand(screenshotCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"screenshot", "export-svg", "--node-id", "node-abc"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "<svg") {
		t.Errorf("expected SVG content in output, got: %s", output)
	}
}
