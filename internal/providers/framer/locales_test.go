package framer

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestLocalesList(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	localesCmd := &cobra.Command{Use: "locales"}
	localesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	localesCmd.AddCommand(newLocalesListCmd(factory))
	root.AddCommand(localesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"locales", "list"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "en-US") {
		t.Errorf("expected 'en-US' in output, got: %s", output)
	}
	if !strings.Contains(output, "French") {
		t.Errorf("expected 'French' in output, got: %s", output)
	}
}

func TestLocalesDefault(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	localesCmd := &cobra.Command{Use: "locales"}
	localesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	localesCmd.AddCommand(newLocalesDefaultCmd(factory))
	root.AddCommand(localesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"locales", "default"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "en-US") {
		t.Errorf("expected 'en-US' in output, got: %s", output)
	}
	if !strings.Contains(output, "English") {
		t.Errorf("expected 'English' in output, got: %s", output)
	}
}

func TestLocalesCreate(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	localesCmd := &cobra.Command{Use: "locales"}
	localesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	localesCmd.AddCommand(newLocalesCreateCmd(factory))
	root.AddCommand(localesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"locales", "create", "--language", "de"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "German") {
		t.Errorf("expected 'German' in output, got: %s", output)
	}
}

func TestLocalesCreateDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	localesCmd := &cobra.Command{Use: "locales"}
	localesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	localesCmd.AddCommand(newLocalesCreateCmd(factory))
	root.AddCommand(localesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"locales", "create", "--language", "de", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestLocalesRegions(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	localesCmd := &cobra.Command{Use: "locales"}
	localesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	localesCmd.AddCommand(newLocalesRegionsCmd(factory))
	root.AddCommand(localesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"locales", "regions", "--language", "en"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "United States") {
		t.Errorf("expected 'United States' in output, got: %s", output)
	}
}

func TestLocalesGroups(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	localesCmd := &cobra.Command{Use: "locales"}
	localesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	localesCmd.AddCommand(newLocalesGroupsCmd(factory))
	root.AddCommand(localesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"locales", "groups"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Default") {
		t.Errorf("expected 'Default' in output, got: %s", output)
	}
}

func TestLocalesSetData(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	localesCmd := &cobra.Command{Use: "locales"}
	localesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	localesCmd.AddCommand(newLocalesSetDataCmd(factory))
	root.AddCommand(localesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"locales", "set-data", "--data", `{"en":{"hello":"Hello"}}`})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "updated") {
		t.Errorf("expected 'updated' in output, got: %s", output)
	}
}

func TestLocalesSetDataDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	localesCmd := &cobra.Command{Use: "locales"}
	localesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	localesCmd.AddCommand(newLocalesSetDataCmd(factory))
	root.AddCommand(localesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"locales", "set-data", "--data", `{"en":{"hello":"Hello"}}`, "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestLocalesLanguages(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	localesCmd := &cobra.Command{Use: "locales"}
	localesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	localesCmd.AddCommand(newLocalesLanguagesCmd(factory))
	root.AddCommand(localesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"locales", "languages"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "English") {
		t.Errorf("expected 'English' in output, got: %s", output)
	}
	if !strings.Contains(output, "French") {
		t.Errorf("expected 'French' in output, got: %s", output)
	}
}
