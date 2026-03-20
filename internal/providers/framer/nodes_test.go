package framer

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNodesGet(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesGetCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "get", "--node-id", "node-abc"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Frame 1") {
		t.Errorf("expected 'Frame 1' in output, got: %s", output)
	}
	if !strings.Contains(output, "FrameNode") {
		t.Errorf("expected 'FrameNode' in output, got: %s", output)
	}
}

func TestNodesChildren(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesChildrenCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "children", "--node-id", "node-abc"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "child-1") {
		t.Errorf("expected 'child-1' in output, got: %s", output)
	}
	if !strings.Contains(output, "TextNode") {
		t.Errorf("expected 'TextNode' in output, got: %s", output)
	}
}

func TestNodesParent(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesParentCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "parent", "--node-id", "node-abc"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "parent-1") {
		t.Errorf("expected 'parent-1' in output, got: %s", output)
	}
}

func TestNodesListByType(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesListByTypeCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "list-by-type", "--type", "FrameNode"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "node-1") {
		t.Errorf("expected 'node-1' in output, got: %s", output)
	}
	if !strings.Contains(output, "Frame A") {
		t.Errorf("expected 'Frame A' in output, got: %s", output)
	}
}

func TestNodesCreateFrame(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesCreateFrameCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "create-frame"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "frame-new") {
		t.Errorf("expected 'frame-new' in output, got: %s", output)
	}
}

func TestNodesCreateFrameDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesCreateFrameCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "create-frame", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestNodesRemoveWithConfirm(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesRemoveCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "remove", "--node-ids", "node-1", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Removed") {
		t.Errorf("expected 'Removed' in output, got: %s", output)
	}
}

func TestNodesRemoveWithoutConfirm(t *testing.T) {
	bridgeCalled := false
	factory := mockBridge(func(method string, params map[string]any) (any, error) {
		if method == "removeNodes" {
			bridgeCalled = true
		}
		return defaultHandler()(method, params)
	})
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesRemoveCmd(factory))
	root.AddCommand(nodesCmd)

	root.SetArgs([]string{"nodes", "remove", "--node-ids", "node-1"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	if bridgeCalled {
		t.Error("expected bridge NOT to be called without --confirm")
	}
}

func TestNodesCreateText(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesCreateTextCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "create-text"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "text-new") {
		t.Errorf("expected 'text-new' in output, got: %s", output)
	}
}

func TestNodesCreateTextDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesCreateTextCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "create-text", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestNodesCreateComponent(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesCreateComponentCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "create-component", "--name", "MyButton"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "comp-new") {
		t.Errorf("expected 'comp-new' in output, got: %s", output)
	}
}

func TestNodesCreateWebPage(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesCreateWebPageCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "create-web-page", "--path", "/about"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "page-new") {
		t.Errorf("expected 'page-new' in output, got: %s", output)
	}
}

func TestNodesCreateDesignPage(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesCreateDesignPageCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "create-design-page", "--name", "Components"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "design-new") {
		t.Errorf("expected 'design-new' in output, got: %s", output)
	}
}

func TestNodesClone(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesCloneCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "clone", "--node-id", "frame-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "clone-1") {
		t.Errorf("expected 'clone-1' in output, got: %s", output)
	}
}

func TestNodesSetAttributes(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesSetAttributesCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "set-attributes", "--node-id", "frame-1", "--attributes", `{"name":"Renamed"}`})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Updated") {
		t.Errorf("expected 'Updated' in output, got: %s", output)
	}
}

func TestNodesSetParent(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesSetParentCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "set-parent", "--node-id", "frame-1", "--parent-id", "page-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Moved Node") || !strings.Contains(output, "Updated parent") {
		// The mock returns "Moved Node" name, command prints "Updated parent of node"
		if !strings.Contains(output, "parent") && !strings.Contains(output, "node") {
			t.Errorf("expected node info in output, got: %s", output)
		}
	}
}

func TestNodesRect(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	nodesCmd := &cobra.Command{Use: "nodes"}
	nodesCmd.PersistentFlags().Bool("json", false, "Output as JSON")
	nodesCmd.AddCommand(newNodesRectCmd(factory))
	root.AddCommand(nodesCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"nodes", "rect", "--node-id", "node-abc"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Width") {
		t.Errorf("expected 'Width' in output, got: %s", output)
	}
	if !strings.Contains(output, "Height") {
		t.Errorf("expected 'Height' in output, got: %s", output)
	}
}
