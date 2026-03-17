package instagram

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// currentUserResponse is the response for GET /api/v1/accounts/current_user/.
type currentUserResponse struct {
	User   rawUserSettings `json:"user"`
	Status string          `json:"status"`
}

// rawUserSettings holds the current user account fields.
type rawUserSettings struct {
	PK                string `json:"pk"`
	Username          string `json:"username"`
	FullName          string `json:"full_name"`
	IsPrivate         bool   `json:"is_private"`
	IsVerified        bool   `json:"is_verified"`
	Email             string `json:"email"`
	PhoneNumber       string `json:"phone_number"`
	Biography         string `json:"biography"`
	ExternalURL       string `json:"external_url"`
	ProfilePicURL     string `json:"profile_pic_url"`
}

// accountActionResponse is a generic response for set-private/set-public.
type accountActionResponse struct {
	Status string `json:"status"`
}

// loginActivityResponse is the response for GET /api/v1/session/login_activity/.
type loginActivityResponse struct {
	LoginActivity []map[string]any `json:"login_activity"`
	Status        string           `json:"status"`
}

// newSettingsCmd builds the `settings` subcommand group.
func newSettingsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "settings",
		Short:   "View and manage account settings",
		Aliases: []string{"setting", "account"},
	}
	cmd.AddCommand(newSettingsGetCmd(factory))
	cmd.AddCommand(newSettingsSetPrivateCmd(factory))
	cmd.AddCommand(newSettingsSetPublicCmd(factory))
	cmd.AddCommand(newSettingsLoginActivityCmd(factory))
	return cmd
}

func newSettingsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get current account settings",
		RunE:  makeRunSettingsGet(factory),
	}
	return cmd
}

func makeRunSettingsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobileGet(ctx, "/api/v1/accounts/current_user/", nil)
		if err != nil {
			return fmt.Errorf("getting account settings: %w", err)
		}

		var result currentUserResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding current user response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result.User)
		}

		lines := []string{
			fmt.Sprintf("Username:    %s", result.User.Username),
			fmt.Sprintf("Full Name:   %s", result.User.FullName),
			fmt.Sprintf("Email:       %s", result.User.Email),
			fmt.Sprintf("Phone:       %s", result.User.PhoneNumber),
			fmt.Sprintf("Private:     %v", result.User.IsPrivate),
			fmt.Sprintf("Verified:    %v", result.User.IsVerified),
			fmt.Sprintf("Bio:         %s", truncate(result.User.Biography, 60)),
			fmt.Sprintf("Website:     %s", result.User.ExternalURL),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newSettingsSetPrivateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-private",
		Short: "Make your account private",
		RunE:  makeRunSettingsSetPrivate(factory),
	}
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunSettingsSetPrivate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "set account to private", map[string]string{})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/accounts/set_private/", nil)
		if err != nil {
			return fmt.Errorf("setting account private: %w", err)
		}

		var result accountActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding set private response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Println("Account set to private.")
		return nil
	}
}

func newSettingsSetPublicCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-public",
		Short: "Make your account public",
		RunE:  makeRunSettingsSetPublic(factory),
	}
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunSettingsSetPublic(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "set account to public", map[string]string{})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/accounts/set_public/", nil)
		if err != nil {
			return fmt.Errorf("setting account public: %w", err)
		}

		var result accountActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding set public response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Println("Account set to public.")
		return nil
	}
}

func newSettingsLoginActivityCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login-activity",
		Short: "Get login activity history",
		RunE:  makeRunSettingsLoginActivity(factory),
	}
	return cmd
}

func makeRunSettingsLoginActivity(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobileGet(ctx, "/api/v1/session/login_activity/", nil)
		if err != nil {
			return fmt.Errorf("getting login activity: %w", err)
		}

		var result loginActivityResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding login activity: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		fmt.Printf("Login activity: %d entries\n", len(result.LoginActivity))
		return nil
	}
}
