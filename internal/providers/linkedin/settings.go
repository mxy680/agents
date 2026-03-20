package linkedin

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerProfileSettingsResponse is the response for GET /voyager/api/identity/profileSettings.
type voyagerProfileSettingsResponse struct {
	ProfileVisibility   string `json:"profileVisibility"`
	MessagingPreference string `json:"messagingPreference"`
	ActiveStatus        bool   `json:"activeStatus"`
}

// voyagerPrivacySettingsResponse is the response for GET /voyager/api/identity/privacySettings.
type voyagerPrivacySettingsResponse struct {
	ProfileVisibility      string `json:"profileVisibility"`
	ConnectionsVisibility  string `json:"connectionsVisibility"`
	LastNameVisibility     string `json:"lastNameVisibility"`
	ProfilePhotoVisibility string `json:"profilePhotoVisibility"`
}

// SettingsInfo holds general profile settings.
type SettingsInfo struct {
	ProfileVisibility   string `json:"profile_visibility"`
	MessagingPreference string `json:"messaging_preference"`
	ActiveStatus        bool   `json:"active_status"`
}

// PrivacySettings holds privacy-related settings.
type PrivacySettings struct {
	ProfileVisibility      string `json:"profile_visibility"`
	ConnectionsVisibility  string `json:"connections_visibility"`
	LastNameVisibility     string `json:"last_name_visibility"`
	ProfilePhotoVisibility string `json:"profile_photo_visibility"`
}

// newSettingsCmd builds the "settings" subcommand group.
func newSettingsCmd(factory ClientFactory) *cobra.Command {
	settingsCmd := &cobra.Command{
		Use:     "settings",
		Short:   "Manage LinkedIn account settings",
		Long:    "View and update your LinkedIn profile and privacy settings.",
		Aliases: []string{"setting"},
	}
	settingsCmd.AddCommand(newSettingsGetCmd(factory))
	settingsCmd.AddCommand(newSettingsPrivacyCmd(factory))
	settingsCmd.AddCommand(newSettingsVisibilityCmd(factory))
	return settingsCmd
}

// newSettingsGetCmd builds the "settings get" command.
func newSettingsGetCmd(factory ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get profile settings",
		Long:  "Retrieve current LinkedIn profile settings.",
		RunE:  makeRunSettingsGet(factory),
	}
}

// newSettingsPrivacyCmd builds the "settings privacy" command.
func newSettingsPrivacyCmd(factory ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "privacy",
		Short: "Get privacy settings",
		Long:  "Retrieve current LinkedIn privacy settings.",
		RunE:  makeRunSettingsPrivacy(factory),
	}
}

// newSettingsVisibilityCmd builds the "settings visibility" command.
func newSettingsVisibilityCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "visibility",
		Short: "Update a profile setting",
		Long:  "Update a named profile setting field to a new value.",
		RunE:  makeRunSettingsVisibility(factory),
	}
	cmd.Flags().String("field", "", "Setting field name to update (e.g. profileVisibility)")
	_ = cmd.MarkFlagRequired("field")
	cmd.Flags().String("value", "", "New value for the setting field")
	_ = cmd.MarkFlagRequired("value")
	cmd.Flags().Bool("dry-run", false, "Print what would be sent without updating")
	return cmd
}

func makeRunSettingsGet(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

func makeRunSettingsPrivacy(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

func makeRunSettingsVisibility(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		field, _ := cmd.Flags().GetString("field")
		value, _ := cmd.Flags().GetString("value")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		body := map[string]string{field: value}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("update setting %q to %q", field, value), body)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.PostJSON(ctx, "/voyager/api/identity/profileSettings", body)
		if err != nil {
			return fmt.Errorf("updating setting %q: %w", field, err)
		}
		resp.Body.Close()

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "updated", "field": field, "value": value})
		}
		fmt.Printf("Setting %q updated to %q\n", field, value)
		return nil
	}
}
