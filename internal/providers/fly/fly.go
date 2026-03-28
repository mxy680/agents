package fly

import (
	"github.com/spf13/cobra"
)

// Provider implements the Fly.io integration.
type Provider struct {
	// ClientFactory creates the Fly.io API client. Defaults to NewClient.
	// Override in tests to inject a mock client pointing at a test server.
	ClientFactory ClientFactory
}

// New creates a new Fly.io provider using the real Fly.io API.
func New() *Provider {
	return &Provider{
		ClientFactory: NewClient,
	}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "fly"
}

// RegisterCommands adds all Fly.io subcommands to the parent command.
func (p *Provider) RegisterCommands(parent *cobra.Command) {
	flyCmd := &cobra.Command{
		Use:     "fly",
		Short:   "Interact with Fly.io",
		Long:    "Manage apps, machines, volumes, certificates, secrets, and regions via the Fly.io API.",
		Aliases: []string{"f"},
	}

	appsCmd := &cobra.Command{
		Use:     "apps",
		Short:   "Manage Fly.io apps",
		Aliases: []string{"app", "a"},
	}
	appsCmd.AddCommand(newAppsListCmd(p.ClientFactory))
	appsCmd.AddCommand(newAppsGetCmd(p.ClientFactory))
	appsCmd.AddCommand(newAppsCreateCmd(p.ClientFactory))
	appsCmd.AddCommand(newAppsDeleteCmd(p.ClientFactory))
	flyCmd.AddCommand(appsCmd)

	machinesCmd := &cobra.Command{
		Use:     "machines",
		Short:   "Manage Fly.io machines",
		Aliases: []string{"machine", "m"},
	}
	machinesCmd.AddCommand(newMachinesListCmd(p.ClientFactory))
	machinesCmd.AddCommand(newMachinesGetCmd(p.ClientFactory))
	machinesCmd.AddCommand(newMachinesCreateCmd(p.ClientFactory))
	machinesCmd.AddCommand(newMachinesUpdateCmd(p.ClientFactory))
	machinesCmd.AddCommand(newMachinesDeleteCmd(p.ClientFactory))
	machinesCmd.AddCommand(newMachinesStartCmd(p.ClientFactory))
	machinesCmd.AddCommand(newMachinesStopCmd(p.ClientFactory))
	machinesCmd.AddCommand(newMachinesWaitCmd(p.ClientFactory))
	flyCmd.AddCommand(machinesCmd)

	volumesCmd := &cobra.Command{
		Use:     "volumes",
		Short:   "Manage Fly.io volumes",
		Aliases: []string{"vol", "v"},
	}
	volumesCmd.AddCommand(newVolumesListCmd(p.ClientFactory))
	volumesCmd.AddCommand(newVolumesGetCmd(p.ClientFactory))
	volumesCmd.AddCommand(newVolumesCreateCmd(p.ClientFactory))
	volumesCmd.AddCommand(newVolumesExtendCmd(p.ClientFactory))
	volumesCmd.AddCommand(newVolumesDeleteCmd(p.ClientFactory))
	volumesCmd.AddCommand(newVolumesSnapshotsCmd(p.ClientFactory))
	flyCmd.AddCommand(volumesCmd)

	certsCmd := &cobra.Command{
		Use:     "certs",
		Short:   "Manage Fly.io TLS certificates",
		Aliases: []string{"cert"},
	}
	certsCmd.AddCommand(newCertsListCmd(p.ClientFactory))
	certsCmd.AddCommand(newCertsGetCmd(p.ClientFactory))
	certsCmd.AddCommand(newCertsAddCmd(p.ClientFactory))
	certsCmd.AddCommand(newCertsCheckCmd(p.ClientFactory))
	certsCmd.AddCommand(newCertsRemoveCmd(p.ClientFactory))
	flyCmd.AddCommand(certsCmd)

	secretsCmd := &cobra.Command{
		Use:     "secrets",
		Short:   "Manage Fly.io app secrets",
		Aliases: []string{"sec", "s"},
	}
	secretsCmd.AddCommand(newSecretsListCmd(p.ClientFactory))
	secretsCmd.AddCommand(newSecretsSetCmd(p.ClientFactory))
	secretsCmd.AddCommand(newSecretsUnsetCmd(p.ClientFactory))
	flyCmd.AddCommand(secretsCmd)

	regionsCmd := &cobra.Command{
		Use:   "regions",
		Short: "List Fly.io regions",
	}
	regionsCmd.AddCommand(newRegionsListCmd(p.ClientFactory))
	flyCmd.AddCommand(regionsCmd)

	parent.AddCommand(flyCmd)
}
