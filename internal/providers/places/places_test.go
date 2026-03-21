package places

import (
	"testing"
)

func TestProviderNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.ScraperFunc == nil {
		t.Fatal("ScraperFunc is nil")
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
	p := &Provider{ScraperFunc: mockScraper(nil, nil)}
	p.RegisterCommands(root)

	found := false
	for _, cmd := range root.Commands() {
		if cmd.Name() == "places" {
			found = true
			subNames := make(map[string]bool)
			for _, sub := range cmd.Commands() {
				subNames[sub.Name()] = true
			}
			expected := []string{"search", "lookup"}
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

func TestProviderAliases(t *testing.T) {
	root := newTestRootCmd()
	p := &Provider{ScraperFunc: mockScraper(nil, nil)}
	p.RegisterCommands(root)

	for _, cmd := range root.Commands() {
		if cmd.Name() == "places" {
			hasAlias := false
			for _, a := range cmd.Aliases {
				if a == "place" {
					hasAlias = true
				}
			}
			if !hasAlias {
				t.Error("places command missing 'place' alias")
			}

			for _, sub := range cmd.Commands() {
				if sub.Name() == "search" {
					hasFindAlias := false
					for _, a := range sub.Aliases {
						if a == "find" {
							hasFindAlias = true
						}
					}
					if !hasFindAlias {
						t.Error("search command missing 'find' alias")
					}
				}
			}
			break
		}
	}
}

func TestDefaultScraperBinary(t *testing.T) {
	// Should return something (either from env, PATH, or the default name)
	bin := defaultScraperBinary()
	if bin == "" {
		t.Error("defaultScraperBinary returned empty string")
	}
}
