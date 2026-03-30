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

// newSettingsVisibilityCmd builds the "settings visibility" command.
func newSettingsVisibilityCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "visibility",
		Short: "Update a profile visibility setting",
		RunE:  makeRunSettingsVisibility(factory),
	}
	cmd.Flags().String("field", "", "Setting field name (required)")
	cmd.Flags().String("value", "", "New value for the setting (required)")
	cmd.Flags().Bool("dry-run", false, "Preview action without executing it")
	_ = cmd.MarkFlagRequired("field")
	_ = cmd.MarkFlagRequired("value")
	return cmd
}

func makeRunSettingsVisibility(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		field, _ := cmd.Flags().GetString("field")
		value, _ := cmd.Flags().GetString("value")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("set %s to %s", field, value),
				map[string]string{"field": field, "value": value})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{field: value}
		_, err = client.PostJSON(ctx, "/voyager/api/identity/profileSettings", body)
		if err != nil {
			return fmt.Errorf("updating setting %s: %w", field, err)
		}

		fmt.Printf("Updated %s to %s\n", field, value)
		return nil
	}
}

