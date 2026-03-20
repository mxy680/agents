package framer

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestStylesColorsList(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	stylesCmd := &cobra.Command{Use: "styles"}
	stylesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	colorsCmd := &cobra.Command{Use: "colors"}
	colorsCmd.AddCommand(newStylesColorsListCmd(factory))
	stylesCmd.AddCommand(colorsCmd)
	root.AddCommand(stylesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"styles", "colors", "list"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Primary") {
		t.Errorf("expected 'Primary' in output, got: %s", output)
	}
	if !strings.Contains(output, "#0066FF") {
		t.Errorf("expected '#0066FF' in output, got: %s", output)
	}
}

func TestStylesColorsGet(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	stylesCmd := &cobra.Command{Use: "styles"}
	stylesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	colorsCmd := &cobra.Command{Use: "colors"}
	colorsCmd.AddCommand(newStylesColorsGetCmd(factory))
	stylesCmd.AddCommand(colorsCmd)
	root.AddCommand(stylesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"styles", "colors", "get", "--id", "cs-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Primary") {
		t.Errorf("expected 'Primary' in output, got: %s", output)
	}
	if !strings.Contains(output, "#0066FF") {
		t.Errorf("expected '#0066FF' in output, got: %s", output)
	}
}

func TestStylesColorsCreate(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	stylesCmd := &cobra.Command{Use: "styles"}
	stylesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	colorsCmd := &cobra.Command{Use: "colors"}
	colorsCmd.AddCommand(newStylesColorsCreateCmd(factory))
	stylesCmd.AddCommand(colorsCmd)
	root.AddCommand(stylesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"styles", "colors", "create", "--attributes", `{"name":"Red","light":"#FF0000"}`})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "cs-new") {
		t.Errorf("expected 'cs-new' in output, got: %s", output)
	}
}

func TestStylesColorsCreateDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	stylesCmd := &cobra.Command{Use: "styles"}
	stylesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	colorsCmd := &cobra.Command{Use: "colors"}
	colorsCmd.AddCommand(newStylesColorsCreateCmd(factory))
	stylesCmd.AddCommand(colorsCmd)
	root.AddCommand(stylesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"styles", "colors", "create", "--attributes", `{"name":"Red","light":"#FF0000"}`, "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestStylesTextList(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	stylesCmd := &cobra.Command{Use: "styles"}
	stylesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	textCmd := &cobra.Command{Use: "text"}
	textCmd.AddCommand(newStylesTextListCmd(factory))
	stylesCmd.AddCommand(textCmd)
	root.AddCommand(stylesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"styles", "text", "list"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Heading 1") {
		t.Errorf("expected 'Heading 1' in output, got: %s", output)
	}
	if !strings.Contains(output, "Inter") {
		t.Errorf("expected 'Inter' in output, got: %s", output)
	}
}

func TestStylesTextCreate(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	stylesCmd := &cobra.Command{Use: "styles"}
	stylesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	textCmd := &cobra.Command{Use: "text"}
	textCmd.AddCommand(newStylesTextCreateCmd(factory))
	stylesCmd.AddCommand(textCmd)
	root.AddCommand(stylesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"styles", "text", "create", "--attributes", `{"name":"Caption","font":"Inter"}`})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "ts-new") {
		t.Errorf("expected 'ts-new' in output, got: %s", output)
	}
}

func TestStylesTextGet(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	stylesCmd := &cobra.Command{Use: "styles"}
	stylesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	textCmd := &cobra.Command{Use: "text"}
	textCmd.AddCommand(newStylesTextGetCmd(factory))
	stylesCmd.AddCommand(textCmd)
	root.AddCommand(stylesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"styles", "text", "get", "--id", "ts-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Heading 1") {
		t.Errorf("expected 'Heading 1' in output, got: %s", output)
	}
	if !strings.Contains(output, "32") {
		t.Errorf("expected font size '32' in output, got: %s", output)
	}
}
