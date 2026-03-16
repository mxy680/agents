package gmail

import (
	"encoding/json"
	"testing"
)

// ---- labels list ----

func TestLabelsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "list", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var labels []LabelInfo
	if err := json.Unmarshal([]byte(output), &labels); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if len(labels) != 3 {
		t.Fatalf("expected 3 labels, got %d", len(labels))
	}
	if labels[0].ID != "INBOX" {
		t.Errorf("expected first label ID=INBOX, got %s", labels[0].ID)
	}
	if labels[2].Name != "TestLabel" {
		t.Errorf("expected third label Name=TestLabel, got %s", labels[2].Name)
	}
}

func TestLabelsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "list"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

// ---- labels get ----

func TestLabelsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "get", "--id=Label_1", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var info LabelInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if info.ID != "Label_1" {
		t.Errorf("expected ID=Label_1, got %s", info.ID)
	}
	if info.Name != "TestLabel" {
		t.Errorf("expected Name=TestLabel, got %s", info.Name)
	}
	if info.LabelListVisibility != "labelShow" {
		t.Errorf("expected LabelListVisibility=labelShow, got %s", info.LabelListVisibility)
	}
}

func TestLabelsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "get", "--id=Label_1"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

// ---- labels create ----

func TestLabelsCreateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "create", "--name=NewLabel", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "Label_created1" {
		t.Errorf("expected id=Label_created1, got %s", result["id"])
	}
	if result["status"] != "created" {
		t.Errorf("expected status=created, got %s", result["status"])
	}
}

func TestLabelsCreateText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "create", "--name=NewLabel"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestLabelsCreateDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "create", "--name=NewLabel", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- labels update ----

func TestLabelsUpdateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "update", "--id=Label_1", "--name=UpdatedLabel", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "Label_1" {
		t.Errorf("expected id=Label_1, got %s", result["id"])
	}
	if result["status"] != "updated" {
		t.Errorf("expected status=updated, got %s", result["status"])
	}
}

func TestLabelsUpdateText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "update", "--id=Label_1", "--name=UpdatedLabel"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestLabelsUpdateDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "update", "--id=Label_1", "--name=UpdatedLabel", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- labels patch ----

func TestLabelsPatchJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "patch", "--id=Label_1", "--name=PatchedLabel", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "Label_1" {
		t.Errorf("expected id=Label_1, got %s", result["id"])
	}
	if result["status"] != "patched" {
		t.Errorf("expected status=patched, got %s", result["status"])
	}
}

func TestLabelsPatchText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "patch", "--id=Label_1", "--name=PatchedLabel"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestLabelsPatchDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "patch", "--id=Label_1", "--name=PatchedLabel", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

// ---- labels delete ----

func TestLabelsDeleteJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "delete", "--id=Label_1", "--confirm", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["id"] != "Label_1" {
		t.Errorf("expected id=Label_1, got %s", result["id"])
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %s", result["status"])
	}
}

func TestLabelsDeleteText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "delete", "--id=Label_1", "--confirm"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestLabelsDeleteDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "delete", "--id=Label_1", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

func TestLabelsDeleteRequiresConfirm(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestLabelsCmd(factory))

	root.SetArgs([]string{"labels", "delete", "--id=Label_1"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --confirm not provided")
	}
}
