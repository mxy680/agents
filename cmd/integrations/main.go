package main

import (
	"os"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/emdash-projects/agents/internal/providers/calendar"
	"github.com/emdash-projects/agents/internal/providers/canvas"
	"github.com/emdash-projects/agents/internal/providers/docs"
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
	"github.com/emdash-projects/agents/internal/providers/census"
	"github.com/emdash-projects/agents/internal/providers/citibike"
	"github.com/emdash-projects/agents/internal/providers/dof"
	"github.com/emdash-projects/agents/internal/providers/hmda"
	vercelprovider "github.com/emdash-projects/agents/internal/providers/vercel"
	cloudflareprovider "github.com/emdash-projects/agents/internal/providers/cloudflare"
	linearprovider "github.com/emdash-projects/agents/internal/providers/linear"
	flyprovider "github.com/emdash-projects/agents/internal/providers/fly"
	gcpprovider "github.com/emdash-projects/agents/internal/providers/gcp"
	gcpconsoleprovider "github.com/emdash-projects/agents/internal/providers/gcpconsole"
	"github.com/emdash-projects/agents/internal/providers/nydos"
	"github.com/emdash-projects/agents/internal/providers/nysla"
	"github.com/emdash-projects/agents/internal/providers/nyscef"
	"github.com/emdash-projects/agents/internal/providers/obituaries"
	"github.com/emdash-projects/agents/internal/providers/trends"
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

	docsProvider := docs.New()
	docsProvider.RegisterCommands(cli.RootCmd())

	zillowProvider := zillow.New()
	zillowProvider.RegisterCommands(cli.RootCmd())

	nyscefProvider := nyscef.New()
	nyscefProvider.RegisterCommands(cli.RootCmd())

	censusProvider := census.New()
	censusProvider.RegisterCommands(cli.RootCmd())

	citibikeProvider := citibike.New()
	citibikeProvider.RegisterCommands(cli.RootCmd())

	hmdaProvider := hmda.New()
	hmdaProvider.RegisterCommands(cli.RootCmd())

	trendsProvider := trends.New()
	trendsProvider.RegisterCommands(cli.RootCmd())

	obituariesProvider := obituaries.New()
	obituariesProvider.RegisterCommands(cli.RootCmd())

	nyslaProvider := nysla.New()
	nyslaProvider.RegisterCommands(cli.RootCmd())

	nydosProvider := nydos.New()
	nydosProvider.RegisterCommands(cli.RootCmd())

	dofProvider := dof.New()
	dofProvider.RegisterCommands(cli.RootCmd())

	vercelProvider := vercelprovider.New()
	vercelProvider.RegisterCommands(cli.RootCmd())

	cloudflareProvider := cloudflareprovider.New()
	cloudflareProvider.RegisterCommands(cli.RootCmd())

	linearProvider := linearprovider.New()
	linearProvider.RegisterCommands(cli.RootCmd())

	flyProvider := flyprovider.New()
	flyProvider.RegisterCommands(cli.RootCmd())

	gcpProvider := gcpprovider.New()
	gcpProvider.RegisterCommands(cli.RootCmd())

	gcpConsoleProvider := gcpconsoleprovider.New()
	gcpConsoleProvider.RegisterCommands(cli.RootCmd())

	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
