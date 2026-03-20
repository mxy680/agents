package framer

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPublishCreate(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	publishCmd := newPublishCmd(factory)
	root.AddCommand(publishCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"publish", "create"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "deploy-1") {
		t.Errorf("expected output to contain 'deploy-1', got: %s", output)
	}
}

func TestPublishCreateDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	publishCmd := newPublishCmd(factory)
	root.AddCommand(publishCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"publish", "create", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestPublishDeploy(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	publishCmd := newPublishCmd(factory)
	root.AddCommand(publishCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"publish", "deploy", "--deployment-id", "deploy-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "deploy-1") {
		t.Errorf("expected output to contain 'deploy-1', got: %s", output)
	}
}

func TestPublishDeployDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	publishCmd := newPublishCmd(factory)
	root.AddCommand(publishCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"publish", "deploy", "--deployment-id", "deploy-1", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestPublishList(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	publishCmd := newPublishCmd(factory)
	root.AddCommand(publishCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"publish", "list"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "deploy-1") {
		t.Errorf("expected output to contain 'deploy-1', got: %s", output)
	}
	if !strings.Contains(output, "deploy-2") {
		t.Errorf("expected output to contain 'deploy-2', got: %s", output)
	}
}

func TestPublishListJSON(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	publishCmd := newPublishCmd(factory)
	root.AddCommand(publishCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"publish", "list", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var deployments []Deployment
	if err := json.Unmarshal([]byte(output), &deployments); err != nil {
		t.Fatalf("expected valid JSON, got: %s", output)
	}
	if len(deployments) != 2 {
		t.Errorf("expected 2 deployments, got: %d", len(deployments))
	}
}

func TestPublishInfo(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	publishCmd := newPublishCmd(factory)
	root.AddCommand(publishCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"publish", "info"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "https://example.com") {
		t.Errorf("expected output to contain 'https://example.com', got: %s", output)
	}
}
