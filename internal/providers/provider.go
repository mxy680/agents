package providers

import "github.com/spf13/cobra"

// Provider defines the interface for service integrations.
type Provider interface {
	// Name returns the provider's identifier (used as the subcommand name).
	Name() string
	// RegisterCommands adds the provider's subcommands to the parent command.
	RegisterCommands(parent *cobra.Command)
}
