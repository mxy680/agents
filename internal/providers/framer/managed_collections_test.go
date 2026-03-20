package framer

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestManagedCollectionsList(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	mcolCmd := &cobra.Command{Use: "managed-collections"}
	mcolCmd.AddCommand(newManagedCollectionsListCmd(factory))
	root.AddCommand(mcolCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"managed-collections", "list"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Managed Posts") {
		t.Errorf("expected 'Managed Posts' in output, got: %s", output)
	}
}

func TestManagedCollectionsCreate(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	mcolCmd := &cobra.Command{Use: "managed-collections"}
	mcolCmd.AddCommand(newManagedCollectionsCreateCmd(factory))
	root.AddCommand(mcolCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"managed-collections", "create", "--name", "New Managed"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "mcol-new") {
		t.Errorf("expected 'mcol-new' in output, got: %s", output)
	}
}

func TestManagedCollectionsCreateDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	mcolCmd := &cobra.Command{Use: "managed-collections"}
	mcolCmd.AddCommand(newManagedCollectionsCreateCmd(factory))
	root.AddCommand(mcolCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"managed-collections", "create", "--name", "New Managed", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestManagedCollectionsFields(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	mcolCmd := &cobra.Command{Use: "managed-collections"}
	mcolCmd.AddCommand(newManagedCollectionsFieldsCmd(factory))
	root.AddCommand(mcolCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"managed-collections", "fields", "--id", "mcol-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Title") {
		t.Errorf("expected 'Title' in output, got: %s", output)
	}
}

func TestManagedCollectionsSetFields(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	mcolCmd := &cobra.Command{Use: "managed-collections"}
	mcolCmd.AddCommand(newManagedCollectionsSetFieldsCmd(factory))
	root.AddCommand(mcolCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"managed-collections", "set-fields", "--id", "mcol-1", "--fields", `[{"name":"Title","type":"string"}]`})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Title") {
		t.Errorf("expected 'Title' in output, got: %s", output)
	}
}

func TestManagedCollectionsSetFieldsDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	mcolCmd := &cobra.Command{Use: "managed-collections"}
	mcolCmd.AddCommand(newManagedCollectionsSetFieldsCmd(factory))
	root.AddCommand(mcolCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"managed-collections", "set-fields", "--id", "mcol-1", "--fields", `[{"name":"Title","type":"string"}]`, "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestManagedCollectionsAddItems(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	mcolCmd := &cobra.Command{Use: "managed-collections"}
	mcolCmd.AddCommand(newManagedCollectionsAddItemsCmd(factory))
	root.AddCommand(mcolCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"managed-collections", "add-items", "--id", "mcol-1", "--items", `[{"id":"item-3"}]`})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "item-new") {
		t.Errorf("expected 'item-new' in output, got: %s", output)
	}
}

func TestManagedCollectionsAddItemsDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	mcolCmd := &cobra.Command{Use: "managed-collections"}
	mcolCmd.AddCommand(newManagedCollectionsAddItemsCmd(factory))
	root.AddCommand(mcolCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"managed-collections", "add-items", "--id", "mcol-1", "--items", `[{"id":"item-3"}]`, "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestManagedCollectionsRemoveItems(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	mcolCmd := &cobra.Command{Use: "managed-collections"}
	mcolCmd.AddCommand(newManagedCollectionsRemoveItemsCmd(factory))
	root.AddCommand(mcolCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"managed-collections", "remove-items", "--id", "mcol-1", "--item-ids", "item-1", "--confirm"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Removed") {
		t.Errorf("expected 'Removed' in output, got: %s", output)
	}
}

func TestManagedCollectionsSetItemOrder(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	mcolCmd := &cobra.Command{Use: "managed-collections"}
	mcolCmd.AddCommand(newManagedCollectionsSetItemOrderCmd(factory))
	root.AddCommand(mcolCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"managed-collections", "set-item-order", "--id", "mcol-1", "--item-ids", "item-2,item-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "Updated item order") {
		t.Errorf("expected 'Updated item order' in output, got: %s", output)
	}
}

func TestManagedCollectionsItems(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	mcolCmd := &cobra.Command{Use: "managed-collections"}
	mcolCmd.AddCommand(newManagedCollectionsItemsCmd(factory))
	root.AddCommand(mcolCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"managed-collections", "items", "--id", "mcol-1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "item-1") {
		t.Errorf("expected 'item-1' in output, got: %s", output)
	}
}
