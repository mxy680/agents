package main

import (
	"os"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/emdash-projects/agents/internal/providers/gmail"
)

func main() {
	// Register providers
	gmailProvider := gmail.New()
	gmailProvider.RegisterCommands(cli.RootCmd())

	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
