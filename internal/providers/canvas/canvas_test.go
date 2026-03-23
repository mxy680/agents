package canvas

import (
	"context"
	"testing"
)

func TestProviderNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.ClientFactory == nil {
		t.Fatal("ClientFactory is nil")
	}
}

func TestProviderName(t *testing.T) {
	p := New()
	if got := p.Name(); got != "canvas" {
		t.Errorf("Name() = %q, want %q", got, "canvas")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := &Provider{
		ClientFactory: func(_ context.Context) (*Client, error) {
			return nil, nil
		},
	}
	root := newTestRootCmd()
	p.RegisterCommands(root)

	// Verify "canvas" subcommand exists
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "canvas" {
			found = true
			// Verify it has subcommands
			if len(cmd.Commands()) == 0 {
				t.Error("canvas command has no subcommands")
			}
			break
		}
	}
	if !found {
		t.Error("canvas command not found on root")
	}
}
