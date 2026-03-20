package supabase

import (
	"context"
	"net/http"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
)

// ClientFactory creates an authenticated HTTP client for the Supabase Management API.
type ClientFactory func(ctx context.Context) (*http.Client, error)

// Provider implements the Supabase Management API integration.
type Provider struct {
	// ClientFactory creates the Supabase HTTP client. Defaults to auth.NewSupabaseClient.
	// Override in tests to inject a mock client pointing at a test server.
	ClientFactory ClientFactory
}

// New creates a new Supabase provider using the real Supabase Management API.
func New() *Provider {
	return &Provider{
		ClientFactory: auth.NewSupabaseClient,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "supabase"
}

// RegisterCommands adds all Supabase subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	rootCmd := &cobra.Command{
		Use:     "supabase",
		Aliases: []string{"sb"},
		Short:   "Interact with Supabase Management API",
		Long:    "Manage Supabase projects, organizations, branches, secrets, and more via the Supabase Management API.",
	}

	// Resource subcommands will be added here incrementally

	parent.AddCommand(rootCmd)
}
