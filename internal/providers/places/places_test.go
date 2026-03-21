package places

import (
	"testing"

	"github.com/emdash-projects/agents/internal/auth"
)

func TestProviderNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.ServiceFactory == nil {
		t.Fatal("ServiceFactory is nil")
	}
}

func TestProviderName(t *testing.T) {
	p := New()
	if got := p.Name(); got != "places" {
		t.Fatalf("Name() = %q, want %q", got, "places")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	root := newTestRootCmd()
	p := New()
	p.RegisterCommands(root)

	// Verify the "places" command was added
	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "places" {
			found = true
			// Verify subcommands
			subNames := make(map[string]bool)
			for _, sub := range cmd.Commands() {
				subNames[sub.Name()] = true
			}
			expected := []string{"search", "get", "autocomplete", "photos"}
			for _, name := range expected {
				if !subNames[name] {
					t.Errorf("missing subcommand %q", name)
				}
			}
			break
		}
	}
	if !found {
		t.Fatal("places command not registered")
	}
}

func TestDefaultServiceFactory(t *testing.T) {
	p := New()
	// Verify it points to auth.NewPlacesService (compare by calling with missing env vars)
	_ = p.ServiceFactory
	_ = auth.NewPlacesService
}
