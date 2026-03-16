package gmail

import (
	"encoding/json"
	"strings"
	"testing"
)

// helpers to build a root + settings command against the mock server.
func newSettingsTestRoot(t *testing.T) (*testEnv, func()) {
	t.Helper()
	srv := newFullMockServer(t)
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))
	return &testEnv{root: root}, srv.Close
}

// testEnv wraps a root command for cleaner test setup.
type testEnv struct {
	root interface {
		SetArgs([]string)
		Execute() error
	}
}

// --- settings get-vacation ---

func TestSettings_GetVacation_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get-vacation", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("get-vacation failed: %v", execErr)
	}

	var info VacationInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if !info.EnableAutoReply {
		t.Error("expected enableAutoReply=true")
	}
	if info.ResponseSubject != "Out of office" {
		t.Errorf("expected responseSubject='Out of office', got %q", info.ResponseSubject)
	}
}

func TestSettings_GetVacation_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get-vacation"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("get-vacation text failed: %v", execErr)
	}
	if !strings.Contains(output, "Enable Auto Reply") {
		t.Errorf("expected 'Enable Auto Reply' in output, got: %s", output)
	}
	if !strings.Contains(output, "Out of office") {
		t.Errorf("expected 'Out of office' in output, got: %s", output)
	}
}

// --- settings set-vacation ---

func TestSettings_SetVacation_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"settings", "set-vacation",
			"--enable-auto-reply=true",
			"--subject=Out of office",
			"--body=I am away.",
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-vacation failed: %v", execErr)
	}

	var info VacationInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
}

func TestSettings_SetVacation_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "set-vacation", "--enable-auto-reply=true"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-vacation text failed: %v", execErr)
	}
	if !strings.Contains(output, "Vacation responder updated") {
		t.Errorf("expected 'Vacation responder updated', got: %s", output)
	}
}

func TestSettings_SetVacation_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"settings", "set-vacation",
			"--enable-auto-reply=true",
			"--subject=Test",
			"--dry-run",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-vacation dry-run failed: %v", execErr)
	}
	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", output)
	}
}

// --- settings get-auto-forwarding ---

func TestSettings_GetAutoForwarding_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get-auto-forwarding", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("get-auto-forwarding failed: %v", execErr)
	}

	var info AutoForwardingInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if !info.Enabled {
		t.Error("expected enabled=true")
	}
	if info.EmailAddress != "forward@example.com" {
		t.Errorf("expected emailAddress='forward@example.com', got %q", info.EmailAddress)
	}
}

func TestSettings_GetAutoForwarding_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get-auto-forwarding"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("get-auto-forwarding text failed: %v", execErr)
	}
	if !strings.Contains(output, "forward@example.com") {
		t.Errorf("expected email address in output, got: %s", output)
	}
}

// --- settings set-auto-forwarding ---

func TestSettings_SetAutoForwarding_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"settings", "set-auto-forwarding",
			"--enabled=true",
			"--email=forward@example.com",
			"--disposition=leaveInInbox",
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-auto-forwarding failed: %v", execErr)
	}

	var info AutoForwardingInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
}

func TestSettings_SetAutoForwarding_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "set-auto-forwarding", "--enabled=true"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-auto-forwarding text failed: %v", execErr)
	}
	if !strings.Contains(output, "Auto-forwarding updated") {
		t.Errorf("expected 'Auto-forwarding updated', got: %s", output)
	}
}

func TestSettings_SetAutoForwarding_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "set-auto-forwarding", "--enabled=true", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-auto-forwarding dry-run failed: %v", execErr)
	}
	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", output)
	}
}

// --- settings get-imap ---

func TestSettings_GetImap_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get-imap", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("get-imap failed: %v", execErr)
	}

	var info ImapInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if !info.Enabled {
		t.Error("expected enabled=true")
	}
	if info.ExpungeBehavior != "archive" {
		t.Errorf("expected expungeBehavior='archive', got %q", info.ExpungeBehavior)
	}
}

func TestSettings_GetImap_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get-imap"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("get-imap text failed: %v", execErr)
	}
	if !strings.Contains(output, "Enabled") {
		t.Errorf("expected 'Enabled' in output, got: %s", output)
	}
}

// --- settings set-imap ---

func TestSettings_SetImap_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"settings", "set-imap",
			"--enabled=true",
			"--auto-expunge=true",
			"--expunge-behavior=archive",
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-imap failed: %v", execErr)
	}

	var info ImapInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
}

func TestSettings_SetImap_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "set-imap", "--enabled=true"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-imap text failed: %v", execErr)
	}
	if !strings.Contains(output, "IMAP settings updated") {
		t.Errorf("expected 'IMAP settings updated', got: %s", output)
	}
}

func TestSettings_SetImap_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "set-imap", "--enabled=true", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-imap dry-run failed: %v", execErr)
	}
	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", output)
	}
}

// --- settings get-pop ---

func TestSettings_GetPop_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get-pop", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("get-pop failed: %v", execErr)
	}

	var info PopInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if info.AccessWindow != "allMail" {
		t.Errorf("expected accessWindow='allMail', got %q", info.AccessWindow)
	}
}

func TestSettings_GetPop_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get-pop"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("get-pop text failed: %v", execErr)
	}
	if !strings.Contains(output, "allMail") {
		t.Errorf("expected 'allMail' in output, got: %s", output)
	}
}

// --- settings set-pop ---

func TestSettings_SetPop_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"settings", "set-pop",
			"--access-window=allMail",
			"--disposition=leaveInInbox",
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-pop failed: %v", execErr)
	}

	var info PopInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if info.AccessWindow != "allMail" {
		t.Errorf("expected accessWindow='allMail', got %q", info.AccessWindow)
	}
}

func TestSettings_SetPop_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "set-pop", "--access-window=allMail"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-pop text failed: %v", execErr)
	}
	if !strings.Contains(output, "POP settings updated") {
		t.Errorf("expected 'POP settings updated', got: %s", output)
	}
}

func TestSettings_SetPop_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "set-pop", "--access-window=disabled", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-pop dry-run failed: %v", execErr)
	}
	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", output)
	}
}

// --- settings get-language ---

func TestSettings_GetLanguage_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get-language", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("get-language failed: %v", execErr)
	}

	var info LanguageInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if info.DisplayLanguage != "en" {
		t.Errorf("expected displayLanguage='en', got %q", info.DisplayLanguage)
	}
}

func TestSettings_GetLanguage_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get-language"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("get-language text failed: %v", execErr)
	}
	if !strings.Contains(output, "Display language: en") {
		t.Errorf("expected 'Display language: en' in output, got: %s", output)
	}
}

// --- settings set-language ---

func TestSettings_SetLanguage_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "set-language", "--display-language=en", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-language failed: %v", execErr)
	}

	var info LanguageInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if info.DisplayLanguage != "en" {
		t.Errorf("expected displayLanguage='en', got %q", info.DisplayLanguage)
	}
}

func TestSettings_SetLanguage_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "set-language", "--display-language=fr"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-language text failed: %v", execErr)
	}
	if !strings.Contains(output, "Language updated to") {
		t.Errorf("expected 'Language updated to' in output, got: %s", output)
	}
}

func TestSettings_SetLanguage_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()
	factory := newTestServiceFactory(srv)
	root := newTestRootCmd()
	root.AddCommand(buildTestSettingsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "set-language", "--display-language=de", "--dry-run"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("set-language dry-run failed: %v", execErr)
	}
	if !strings.Contains(output, "[DRY RUN]") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", output)
	}
}
