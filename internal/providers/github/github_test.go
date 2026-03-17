package github

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestProviderNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
	if p.ClientFactory == nil {
		t.Fatal("expected non-nil ClientFactory")
	}
}

func TestProviderName(t *testing.T) {
	p := New()
	if got := p.Name(); got != "github" {
		t.Errorf("Name() = %q, want %q", got, "github")
	}
}

func TestProviderRegisterCommands(t *testing.T) {
	p := New()
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")

	p.RegisterCommands(root)

	// Verify github command was added
	githubCmd, _, err := root.Find([]string{"github"})
	if err != nil {
		t.Fatalf("github command not found: %v", err)
	}

	// Verify all resource group subcommands exist
	expectedSubcommands := []string{
		"repos", "issues", "pulls", "runs", "releases",
		"gists", "search", "git", "orgs", "teams", "labels", "branches",
	}
	for _, name := range expectedSubcommands {
		found := false
		for _, cmd := range githubCmd.Commands() {
			if cmd.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found under github", name)
		}
	}
}

func TestProviderAliases(t *testing.T) {
	p := New()
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")

	p.RegisterCommands(root)

	githubCmd, _, _ := root.Find([]string{"github"})

	aliasTests := []struct {
		name    string
		aliases []string
	}{
		{"repos", []string{"repo"}},
		{"issues", []string{"issue"}},
		{"pulls", []string{"pull", "pr"}},
		{"runs", []string{"run"}},
		{"releases", []string{"release"}},
		{"gists", []string{"gist"}},
		{"orgs", []string{"org"}},
		{"teams", []string{"team"}},
		{"labels", []string{"label"}},
		{"branches", []string{"branch"}},
	}
	for _, tt := range aliasTests {
		for _, cmd := range githubCmd.Commands() {
			if cmd.Name() == tt.name {
				for _, alias := range tt.aliases {
					found := false
					for _, a := range cmd.Aliases {
						if a == alias {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected alias %q for %q", alias, tt.name)
					}
				}
			}
		}
	}
}
