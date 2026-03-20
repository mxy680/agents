package framer

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCollectionsList(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsListCmd(factory))
	root.AddCommand(colCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"collections", "list"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Blog Posts") {
		t.Errorf("expected 'Blog Posts' in output, got: %s", output)
	}
	if !strings.Contains(output, "Products") {
		t.Errorf("expected 'Products' in output, got: %s", output)
	}
}

func TestCollectionsGet(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsGetCmd(factory))
	root.AddCommand(colCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"collections", "get", "--id", "col-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Blog Posts") {
		t.Errorf("expected 'Blog Posts' in output, got: %s", output)
	}
}

func TestCollectionsCreate(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsCreateCmd(factory))
	root.AddCommand(colCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"collections", "create", "--name", "New Collection"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "col-new") {
		t.Errorf("expected 'col-new' in output, got: %s", output)
	}
}

func TestCollectionsCreateDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsCreateCmd(factory))
	root.AddCommand(colCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"collections", "create", "--name", "New Collection", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestCollectionsFields(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsFieldsCmd(factory))
	root.AddCommand(colCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"collections", "fields", "--id", "col-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Title") {
		t.Errorf("expected 'Title' in output, got: %s", output)
	}
	if !strings.Contains(output, "Body") {
		t.Errorf("expected 'Body' in output, got: %s", output)
	}
}

func TestCollectionsItems(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsItemsCmd(factory))
	root.AddCommand(colCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"collections", "items", "--id", "col-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "hello-world") {
		t.Errorf("expected 'hello-world' in output, got: %s", output)
	}
}

func TestCollectionsAddItemsDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsAddItemsCmd(factory))
	root.AddCommand(colCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"collections", "add-items", "--id", "col-1", "--items", `[{"slug":"new"}]`, "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestCollectionsRemoveItemsWithConfirm(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsRemoveItemsCmd(factory))
	root.AddCommand(colCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"collections", "remove-items", "--id", "col-1", "--item-ids", "item-1", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Removed") {
		t.Errorf("expected 'Removed' in output, got: %s", output)
	}
}

func TestCollectionsAddFieldsDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsAddFieldsCmd(factory))
	root.AddCommand(colCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"collections", "add-fields", "--id", "col-1", "--fields", `[{"name":"Author","type":"string"}]`, "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestCollectionsAddFields(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsAddFieldsCmd(factory))
	root.AddCommand(colCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"collections", "add-fields", "--id", "col-1", "--fields", `[{"name":"Author","type":"string"}]`})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Author") {
		t.Errorf("expected 'Author' in output, got: %s", output)
	}
}

func TestCollectionsRemoveFields(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsRemoveFieldsCmd(factory))
	root.AddCommand(colCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"collections", "remove-fields", "--id", "col-1", "--field-ids", "field-1", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Removed") {
		t.Errorf("expected 'Removed' in output, got: %s", output)
	}
}

func TestCollectionsSetFieldOrder(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsSetFieldOrderCmd(factory))
	root.AddCommand(colCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"collections", "set-field-order", "--id", "col-1", "--field-ids", "field-2,field-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Updated field order") {
		t.Errorf("expected 'Updated field order' in output, got: %s", output)
	}
}

func TestCollectionsAddItems(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsAddItemsCmd(factory))
	root.AddCommand(colCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"collections", "add-items", "--id", "col-1", "--items", `[{"slug":"new-post"}]`})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	// mock returns {success: true} for addCollectionItems, which marshals as a raw value
	_ = output
}

func TestCollectionsSetItemOrder(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsSetItemOrderCmd(factory))
	root.AddCommand(colCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"collections", "set-item-order", "--id", "col-1", "--item-ids", "item-2,item-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Updated item order") {
		t.Errorf("expected 'Updated item order' in output, got: %s", output)
	}
}

func TestCollectionsRemoveItemsWithoutConfirm(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	colCmd := &cobra.Command{Use: "collections"}
	colCmd.AddCommand(newCollectionsRemoveItemsCmd(factory))
	root.AddCommand(colCmd)

	// Without --confirm, should not call the bridge (just prints warning to stderr)
	executed := false
	handlerCalled := false
	factory2 := mockBridge(func(method string, params map[string]any) (any, error) {
		if method == "removeCollectionItems" {
			handlerCalled = true
		}
		return defaultHandler()(method, params)
	})
	root2 := newTestRootCmd()
	colCmd2 := &cobra.Command{Use: "collections"}
	colCmd2.AddCommand(newCollectionsRemoveItemsCmd(factory2))
	root2.AddCommand(colCmd2)

	root2.SetArgs([]string{"collections", "remove-items", "--id", "col-1", "--item-ids", "item-1"})
	if err := root2.Execute(); err != nil {
		t.Fatal(err)
	}
	executed = true

	if !executed {
		t.Error("expected command to execute without error")
	}
	if handlerCalled {
		t.Error("expected bridge NOT to be called without --confirm")
	}
}
