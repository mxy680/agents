package framer

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestFontsList(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	fontsCmd := &cobra.Command{Use: "fonts"}
	fontsCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	fontsCmd.AddCommand(newFontsListCmd(factory))
	root.AddCommand(fontsCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"fonts", "list"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Inter") {
		t.Errorf("expected 'Inter' in output, got: %s", output)
	}
}

func TestFontsGet(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	fontsCmd := &cobra.Command{Use: "fonts"}
	fontsCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	fontsCmd.AddCommand(newFontsGetCmd(factory))
	root.AddCommand(fontsCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"fonts", "get", "--family", "Inter"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Inter") {
		t.Errorf("expected 'Inter' in output, got: %s", output)
	}
	if !strings.Contains(output, "normal") {
		t.Errorf("expected 'normal' in output, got: %s", output)
	}
}

func TestFontsGetNotFound(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	fontsCmd := &cobra.Command{Use: "fonts"}
	fontsCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	fontsCmd.AddCommand(newFontsGetCmd(factory))
	root.AddCommand(fontsCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"fonts", "get", "--family", "NotFound"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "not found") {
		t.Errorf("expected 'not found' in output, got: %s", output)
	}
}
