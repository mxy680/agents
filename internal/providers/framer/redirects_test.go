package framer

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRedirectsList(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	redirectsCmd := &cobra.Command{Use: "redirects"}
	redirectsCmd.AddCommand(newRedirectsListCmd(factory))
	root.AddCommand(redirectsCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"redirects", "list"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "/old") {
		t.Errorf("expected '/old' in output, got: %s", output)
	}
	if !strings.Contains(output, "/new") {
		t.Errorf("expected '/new' in output, got: %s", output)
	}
}

func TestRedirectsAdd(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	redirectsCmd := &cobra.Command{Use: "redirects"}
	redirectsCmd.AddCommand(newRedirectsAddCmd(factory))
	root.AddCommand(redirectsCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"redirects", "add", "--redirects", `[{"from":"/test","to":"/dest"}]`})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "redir-new") {
		t.Errorf("expected 'redir-new' in output, got: %s", output)
	}
}

func TestRedirectsAddDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	redirectsCmd := &cobra.Command{Use: "redirects"}
	redirectsCmd.AddCommand(newRedirectsAddCmd(factory))
	root.AddCommand(redirectsCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"redirects", "add", "--redirects", `[{"from":"/test","to":"/dest"}]`, "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestRedirectsRemoveWithConfirm(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	redirectsCmd := &cobra.Command{Use: "redirects"}
	redirectsCmd.AddCommand(newRedirectsRemoveCmd(factory))
	root.AddCommand(redirectsCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"redirects", "remove", "--ids", "redir-1", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Removed") {
		t.Errorf("expected 'Removed' in output, got: %s", output)
	}
}

func TestRedirectsSetOrder(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	redirectsCmd := &cobra.Command{Use: "redirects"}
	redirectsCmd.AddCommand(newRedirectsSetOrderCmd(factory))
	root.AddCommand(redirectsCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"redirects", "set-order", "--ids", "redir-2,redir-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Reordered") {
		t.Errorf("expected 'Reordered' in output, got: %s", output)
	}
}

func TestRedirectsRemoveWithoutConfirm(t *testing.T) {
	bridgeCalled := false
	factory := mockBridge(func(method string, params map[string]any) (any, error) {
		if method == "removeRedirects" {
			bridgeCalled = true
		}
		return defaultHandler()(method, params)
	})
	root := newTestRootCmd()
	redirectsCmd := &cobra.Command{Use: "redirects"}
	redirectsCmd.AddCommand(newRedirectsRemoveCmd(factory))
	root.AddCommand(redirectsCmd)

	root.SetArgs([]string{"redirects", "remove", "--ids", "redir-1"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	if bridgeCalled {
		t.Error("expected bridge NOT to be called without --confirm")
	}
}
