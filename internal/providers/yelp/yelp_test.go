package yelp

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestProviderNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.clientFactory == nil {
		t.Fatal("clientFactory is nil")
	}
}

func TestProviderName(t *testing.T) {
	p := New()
	if got := p.Name(); got != "yelp" {
		t.Errorf("Name() = %q, want %q", got, "yelp")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	p := &Provider{clientFactory: newTestClientFactory(server)}
	root := &cobra.Command{Use: "root"}
	p.RegisterCommands(root)

	// Verify yelp command was registered
	yelpCmd, _, err := root.Find([]string{"yelp"})
	if err != nil {
		t.Fatalf("yelp command not found: %v", err)
	}
	if yelpCmd.Use != "yelp" {
		t.Errorf("yelp command Use = %q, want %q", yelpCmd.Use, "yelp")
	}

	// Verify expected subcommands exist
	expectedSubs := []string{"businesses", "reviews", "events", "categories", "autocomplete", "transactions"}
	for _, name := range expectedSubs {
		found := false
		for _, sub := range yelpCmd.Commands() {
			if sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found under yelp", name)
		}
	}
}
