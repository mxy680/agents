package framer

import (
	"strings"
	"testing"
)

func TestSVGAdd(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	svgCmd := newSVGCmd(factory)
	root.AddCommand(svgCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"svg", "add", "--svg", `<svg xmlns="http://www.w3.org/2000/svg"/>`})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "SVG added") {
		t.Errorf("expected 'SVG added' in output, got: %s", output)
	}
}

func TestSVGAddDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	svgCmd := newSVGCmd(factory)
	root.AddCommand(svgCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"svg", "add", "--svg", `<svg/>`, "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestSVGVectorSets(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	svgCmd := newSVGCmd(factory)
	root.AddCommand(svgCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"svg", "vector-sets"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Icons") {
		t.Errorf("expected 'Icons' in output, got: %s", output)
	}
}
