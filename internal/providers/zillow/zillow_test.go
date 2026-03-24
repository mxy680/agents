package zillow

import (
	"testing"

	"github.com/spf13/cobra"
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
	if got := p.Name(); got != "zillow" {
		t.Errorf("Name() = %q, want %q", got, "zillow")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := &Provider{
		ClientFactory: func(_ interface{}) ClientFactory {
			return nil
		}(nil),
	}
	// Use a real factory so RegisterCommands doesn't panic
	p.ClientFactory = DefaultClientFactory()

	root := &cobra.Command{Use: "root"}
	p.RegisterCommands(root)

	// Verify zillow command was registered
	zillowCmd, _, err := root.Find([]string{"zillow"})
	if err != nil {
		t.Fatalf("zillow command not found: %v", err)
	}
	if zillowCmd.Use != "zillow" {
		t.Errorf("zillow command Use = %q, want %q", zillowCmd.Use, "zillow")
	}

	// Verify alias
	zwCmd, _, err := root.Find([]string{"zw"})
	if err != nil {
		t.Fatalf("zw alias not found: %v", err)
	}
	if zwCmd != zillowCmd {
		t.Error("zw alias does not point to zillow command")
	}

	// Verify subcommands exist
	expectedSubs := []string{
		"properties", "search", "mortgage", "rentals",
	}
	for _, name := range expectedSubs {
		found := false
		for _, sub := range zillowCmd.Commands() {
			if sub.Use == name || sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found under zillow", name)
		}
	}
}
