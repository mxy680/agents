package gmail

import (
	"encoding/json"
	"strings"
	"testing"
)

// ---- settings delegates list ----

func TestSettingsDelegatesListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "delegates", "list", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var delegates []DelegateInfo
	if err := json.Unmarshal([]byte(output), &delegates); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if len(delegates) != 1 {
		t.Fatalf("expected 1 delegate, got %d", len(delegates))
	}
	if delegates[0].DelegateEmail != "delegate@example.com" {
		t.Errorf("expected delegateEmail=delegate@example.com, got %s", delegates[0].DelegateEmail)
	}
	if delegates[0].VerificationStatus != "accepted" {
		t.Errorf("expected verificationStatus=accepted, got %s", delegates[0].VerificationStatus)
	}
}

func TestSettingsDelegatesListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "delegates", "list"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
	if !strings.Contains(output, "delegate@example.com") {
		t.Error("expected output to contain delegate@example.com")
	}
}

// ---- settings delegates get ----

func TestSettingsDelegatesGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "delegates", "get", "--email=delegate@example.com", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var info DelegateInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if info.DelegateEmail != "delegate@example.com" {
		t.Errorf("expected delegateEmail=delegate@example.com, got %s", info.DelegateEmail)
	}
	if info.VerificationStatus != "accepted" {
		t.Errorf("expected verificationStatus=accepted, got %s", info.VerificationStatus)
	}
}

func TestSettingsDelegatesGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "delegates", "get", "--email=delegate@example.com"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

// ---- settings delegates create ----

func TestSettingsDelegatesCreateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"settings", "delegates", "create",
			"--email=delegate@example.com",
			"--confirm",
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["delegateEmail"] != "delegate@example.com" {
		t.Errorf("expected delegateEmail=delegate@example.com, got %s", result["delegateEmail"])
	}
	if result["status"] != "created" {
		t.Errorf("expected status=created, got %s", result["status"])
	}
}

func TestSettingsDelegatesCreateDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "delegates", "create", "--email=delegate@example.com", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

func TestSettingsDelegatesCreateRequiresConfirm(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	root.SetArgs([]string{"settings", "delegates", "create", "--email=delegate@example.com"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --confirm not provided")
	}
}

// ---- settings delegates delete ----

func TestSettingsDelegatesDeleteJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "delegates", "delete", "--email=delegate@example.com", "--confirm", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["delegateEmail"] != "delegate@example.com" {
		t.Errorf("expected delegateEmail=delegate@example.com, got %s", result["delegateEmail"])
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %s", result["status"])
	}
}

func TestSettingsDelegatesDeleteText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "delegates", "delete", "--email=delegate@example.com", "--confirm"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestSettingsDelegatesDeleteDryRun(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "delegates", "delete", "--email=delegate@example.com", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected dry-run output")
	}
}

func TestSettingsDelegatesDeleteRequiresConfirm(t *testing.T) {
	factory := newTestServiceFactory(newFullMockServer(t))
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	root.SetArgs([]string{"settings", "delegates", "delete", "--email=delegate@example.com"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --confirm not provided")
	}
}
