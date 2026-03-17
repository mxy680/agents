package sheets

import (
	"testing"
)

func TestProviderNew(t *testing.T) {
	p := New()
	if p.Name() != "sheets" {
		t.Errorf("expected name=sheets, got %s", p.Name())
	}
	if p.SheetsServiceFactory == nil {
		t.Error("expected SheetsServiceFactory to be set")
	}
	if p.DriveServiceFactory == nil {
		t.Error("expected DriveServiceFactory to be set")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	p := &Provider{
		SheetsServiceFactory: newTestSheetsServiceFactory(server),
		DriveServiceFactory:  newTestDriveServiceFactory(server),
	}
	root := newTestRootCmd()
	p.RegisterCommands(root)

	// Verify the sheets subcommand is registered.
	sheetsCmd, _, err := root.Find([]string{"sheets"})
	if err != nil || sheetsCmd == nil {
		t.Fatal("expected sheets command to be registered")
	}

	// Verify subcommand groups.
	for _, group := range []string{"spreadsheets", "values", "tabs"} {
		cmd, _, err := root.Find([]string{"sheets", group})
		if err != nil || cmd == nil {
			t.Errorf("expected sheets %s command to be registered", group)
		}
	}

	// Verify values subcommands.
	valuesCmd, _, _ := root.Find([]string{"sheets", "values"})
	subCmds := map[string]bool{}
	for _, c := range valuesCmd.Commands() {
		subCmds[c.Use] = true
	}
	for _, expected := range []string{"get", "update", "append", "clear", "batch-get", "batch-update"} {
		if !subCmds[expected] {
			t.Errorf("expected subcommand %q to be registered under values", expected)
		}
	}
}
