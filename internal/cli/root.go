package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:              "integrations",
	Short:            "CLI for AI agents to interact with external services",
	Long:             "A CLI binary that AI agents call inside Docker containers to interact with external services like Gmail, Calendar, and more.",
	PersistentPreRunE: resolveCredentials,
}

func init() {
	rootCmd.PersistentFlags().Bool("json", false, "Output in JSON format")
	rootCmd.PersistentFlags().Bool("dry-run", false, "Preview actions without executing them")
}

// RootCmd returns the root cobra command for provider registration.
func RootCmd() *cobra.Command {
	return rootCmd
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
