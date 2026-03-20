package framer

import (
	"strings"
	"testing"
)

func TestProviderNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	if p.Name() != "framer" {
		t.Errorf("expected name 'framer', got: %s", p.Name())
	}
	if p.BridgeClientFactory == nil {
		t.Error("expected non-nil BridgeClientFactory")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := &Provider{BridgeClientFactory: mockBridge(defaultHandler())}
	root := newTestRootCmd()
	p.RegisterCommands(root)

	// Find the framer subcommand
	framerCmd, _, err := root.Find([]string{"framer"})
	if err != nil || framerCmd == nil || framerCmd.Use != "framer" {
		t.Fatal("expected 'framer' command to be registered")
	}

	// Collect all subcommand names
	subNames := make(map[string]bool)
	for _, sub := range framerCmd.Commands() {
		subNames[sub.Use] = true
	}

	expected := []string{
		"project",
		"publish",
		"agent",
		"screenshot",
		"plugin-data",
		"collections",
		"managed-collections",
		"redirects",
		"code",
		"images",
		"files",
		"svg",
		"nodes",
		"styles",
		"fonts",
		"locales",
	}

	for _, name := range expected {
		if !subNames[name] {
			t.Errorf("expected subcommand %q to be registered", name)
		}
	}
}

func TestProviderAlias(t *testing.T) {
	p := &Provider{BridgeClientFactory: mockBridge(defaultHandler())}
	root := newTestRootCmd()
	p.RegisterCommands(root)

	framerCmd, _, _ := root.Find([]string{"framer"})
	if framerCmd == nil {
		t.Fatal("framer command not found")
	}

	aliases := framerCmd.Aliases
	found := false
	for _, a := range aliases {
		if a == "fr" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected alias 'fr' on framer command, got: %v", aliases)
	}
}

func TestProviderName(t *testing.T) {
	p := New()
	if !strings.EqualFold(p.Name(), "framer") {
		t.Errorf("expected provider name 'framer', got: %s", p.Name())
	}
}
