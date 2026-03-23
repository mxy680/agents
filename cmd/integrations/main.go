package main

import (
	"os"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/emdash-projects/agents/internal/providers/calendar"
	"github.com/emdash-projects/agents/internal/providers/canvas"
	"github.com/emdash-projects/agents/internal/providers/drive"
	"github.com/emdash-projects/agents/internal/providers/framer"
	githubprovider "github.com/emdash-projects/agents/internal/providers/github"
	"github.com/emdash-projects/agents/internal/providers/gmail"
	"github.com/emdash-projects/agents/internal/providers/imessage"
	"github.com/emdash-projects/agents/internal/providers/instagram"
	"github.com/emdash-projects/agents/internal/providers/linkedin"
	"github.com/emdash-projects/agents/internal/providers/places"
	"github.com/emdash-projects/agents/internal/providers/sheets"
	supabaseprovider "github.com/emdash-projects/agents/internal/providers/supabase"
	xprovider "github.com/emdash-projects/agents/internal/providers/x"
	"github.com/emdash-projects/agents/internal/providers/zillow"
)

func main() {
	// Register providers
	gmailProvider := gmail.New()
	gmailProvider.RegisterCommands(cli.RootCmd())

	sheetsProvider := sheets.New()
	sheetsProvider.RegisterCommands(cli.RootCmd())

	calendarProvider := calendar.New()
	calendarProvider.RegisterCommands(cli.RootCmd())

	driveProvider := drive.New()
	driveProvider.RegisterCommands(cli.RootCmd())

	instagramProvider := instagram.New()
	instagramProvider.RegisterCommands(cli.RootCmd())

	githubProvider := githubprovider.New()
	githubProvider.RegisterCommands(cli.RootCmd())

	linkedinProvider := linkedin.New()
	linkedinProvider.RegisterCommands(cli.RootCmd())

	framerProvider := framer.New()
	framerProvider.RegisterCommands(cli.RootCmd())

	placesProvider := places.New()
	placesProvider.RegisterCommands(cli.RootCmd())

	supabaseProvider := supabaseprovider.New()
	supabaseProvider.RegisterCommands(cli.RootCmd())

	xProvider := xprovider.New()
	xProvider.RegisterCommands(cli.RootCmd())

	imessageProvider := imessage.New()
	imessageProvider.RegisterCommands(cli.RootCmd())

	canvasProvider := canvas.New()
	canvasProvider.RegisterCommands(cli.RootCmd())

	zillowProvider := zillow.New()
	zillowProvider.RegisterCommands(cli.RootCmd())

	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
