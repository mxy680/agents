package framer

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCodeList(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	codeCmd := &cobra.Command{Use: "code"}
	codeCmd.AddCommand(newCodeListCmd(factory))
	root.AddCommand(codeCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"code", "list"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "analytics.tsx") {
		t.Errorf("expected 'analytics.tsx' in output, got: %s", output)
	}
	if !strings.Contains(output, "utils.ts") {
		t.Errorf("expected 'utils.ts' in output, got: %s", output)
	}
}

func TestCodeGet(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	codeCmd := &cobra.Command{Use: "code"}
	codeCmd.AddCommand(newCodeGetCmd(factory))
	root.AddCommand(codeCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"code", "get", "--id", "code-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "analytics.tsx") {
		t.Errorf("expected 'analytics.tsx' in output, got: %s", output)
	}
}

func TestCodeCreate(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	codeCmd := &cobra.Command{Use: "code"}
	codeCmd.AddCommand(newCodeCreateCmd(factory))
	root.AddCommand(codeCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"code", "create", "--name", "new.ts", "--code", "export const x = 1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "code-new") {
		t.Errorf("expected 'code-new' in output, got: %s", output)
	}
}

func TestCodeCreateDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	codeCmd := &cobra.Command{Use: "code"}
	codeCmd.AddCommand(newCodeCreateCmd(factory))
	root.AddCommand(codeCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"code", "create", "--name", "new.ts", "--code", "export const x = 1", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestCodeTypecheck(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	codeCmd := &cobra.Command{Use: "code"}
	codeCmd.AddCommand(newCodeTypecheckCmd(factory))
	root.AddCommand(codeCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"code", "typecheck", "--name", "test.ts", "--content", "const x: string = 1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Diagnostics") {
		t.Errorf("expected 'Diagnostics' in output, got: %s", output)
	}
}

func TestCodeCustomSet(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	codeCmd := &cobra.Command{Use: "code"}
	codeCmd.AddCommand(newCodeCustomSetCmd(factory))
	root.AddCommand(codeCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"code", "custom-set", "--html", "<script>alert(1)</script>", "--location", "headEnd"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "headEnd") {
		t.Errorf("expected 'headEnd' in output, got: %s", output)
	}
}

func TestCodeCustomSetDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	codeCmd := &cobra.Command{Use: "code"}
	codeCmd.AddCommand(newCodeCustomSetCmd(factory))
	root.AddCommand(codeCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"code", "custom-set", "--html", "<script>alert(1)</script>", "--location", "headEnd", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestCodeCustomGet(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	codeCmd := &cobra.Command{Use: "code"}
	codeCmd.AddCommand(newCodeCustomGetCmd(factory))
	root.AddCommand(codeCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"code", "custom-get"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "HeadEnd") {
		t.Errorf("expected 'HeadEnd' in output, got: %s", output)
	}
	if !strings.Contains(output, "console.log") {
		t.Errorf("expected 'console.log' in output, got: %s", output)
	}
}
