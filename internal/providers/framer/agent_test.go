package framer

import (
	"strings"
	"testing"
)

func TestAgentSystemPrompt(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	agentCmd := newAgentCmd(factory)
	root.AddCommand(agentCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"agent", "system-prompt"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Framer design agent") {
		t.Errorf("expected 'Framer design agent' in output, got: %s", output)
	}
}

func TestAgentContext(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	agentCmd := newAgentCmd(factory)
	root.AddCommand(agentCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"agent", "context"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Inter") {
		t.Errorf("expected 'Inter' in output, got: %s", output)
	}
}

func TestAgentRead(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	agentCmd := newAgentCmd(factory)
	root.AddCommand(agentCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"agent", "read", "--queries", `["pages"]`})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "pages") {
		t.Errorf("expected 'pages' in output, got: %s", output)
	}
}

func TestAgentApply(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	agentCmd := newAgentCmd(factory)
	root.AddCommand(agentCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"agent", "apply", "--dsl", `{"commands":[]}`})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "applied") {
		t.Errorf("expected 'applied' in output, got: %s", output)
	}
}

func TestAgentApplyDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	agentCmd := newAgentCmd(factory)
	root.AddCommand(agentCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"agent", "apply", "--dsl", `{"commands":[]}`, "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}
