package gmail

import (
	"context"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
)

// ForwardingAddressInfo is the JSON-serializable representation of a forwarding address.
type ForwardingAddressInfo struct {
	ForwardingEmail    string `json:"forwardingEmail"`
	VerificationStatus string `json:"verificationStatus"`
}

// forwardingAddressFromAPI converts a Gmail API ForwardingAddress to ForwardingAddressInfo.
func forwardingAddressFromAPI(fa *api.ForwardingAddress) ForwardingAddressInfo {
	return ForwardingAddressInfo{
		ForwardingEmail:    fa.ForwardingEmail,
		VerificationStatus: fa.VerificationStatus,
	}
}

// --- settings forwarding-addresses list ---

// newSettingsForwardingListCmd returns the `settings forwarding-addresses list` command.
func newSettingsForwardingListCmd(factory ServiceFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all forwarding addresses",
		RunE:  makeRunSettingsForwardingList(factory),
	}
}

func makeRunSettingsForwardingList(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Users.Settings.ForwardingAddresses.List("me").Do()
		if err != nil {
			return fmt.Errorf("listing forwarding addresses: %w", err)
		}

		addresses := make([]ForwardingAddressInfo, 0, len(resp.ForwardingAddresses))
		for _, fa := range resp.ForwardingAddresses {
			addresses = append(addresses, forwardingAddressFromAPI(fa))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(addresses)
		}

		if len(addresses) == 0 {
			fmt.Println("No forwarding addresses found.")
			return nil
		}

		lines := make([]string, 0, len(addresses)+1)
		lines = append(lines, fmt.Sprintf("%-40s  %s", "FORWARDING EMAIL", "VERIFICATION STATUS"))
		for _, a := range addresses {
			lines = append(lines, fmt.Sprintf("%-40s  %s", truncate(a.ForwardingEmail, 40), a.VerificationStatus))
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- settings forwarding-addresses get ---

// newSettingsForwardingGetCmd returns the `settings forwarding-addresses get` command.
func newSettingsForwardingGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a forwarding address",
		RunE:  makeRunSettingsForwardingGet(factory),
	}
	cmd.Flags().String("email", "", "Forwarding email address (required)")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func makeRunSettingsForwardingGet(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		email, _ := cmd.Flags().GetString("email")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		fa, err := svc.Users.Settings.ForwardingAddresses.Get("me", email).Do()
		if err != nil {
			return fmt.Errorf("getting forwarding address %s: %w", email, err)
		}

		info := forwardingAddressFromAPI(fa)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("Forwarding Email:    %s", info.ForwardingEmail),
			fmt.Sprintf("Verification Status: %s", info.VerificationStatus),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- settings forwarding-addresses create ---

// newSettingsForwardingCreateCmd returns the `settings forwarding-addresses create` command.
func newSettingsForwardingCreateCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a forwarding address (sends verification email)",
		RunE:  makeRunSettingsForwardingCreate(factory),
	}
	cmd.Flags().String("email", "", "Forwarding email address (required)")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func makeRunSettingsForwardingCreate(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		email, _ := cmd.Flags().GetString("email")

		if cli.IsDryRun(cmd) {
			info := map[string]string{
				"forwardingEmail":    email,
				"verificationStatus": "pending",
				"status":             "created",
			}
			return dryRunResult(cmd, fmt.Sprintf("Would create forwarding address %s (verification email would be sent)", email), info)
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		created, err := svc.Users.Settings.ForwardingAddresses.Create("me", &api.ForwardingAddress{
			ForwardingEmail: email,
		}).Do()
		if err != nil {
			return fmt.Errorf("creating forwarding address %s: %w", email, err)
		}

		result := map[string]string{
			"forwardingEmail":    created.ForwardingEmail,
			"verificationStatus": created.VerificationStatus,
			"status":             "created",
		}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Forwarding address %s created (verification email sent)\n", created.ForwardingEmail)
		return nil
	}
}

// --- settings forwarding-addresses delete ---

// newSettingsForwardingDeleteCmd returns the `settings forwarding-addresses delete` command.
func newSettingsForwardingDeleteCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a forwarding address (IRREVERSIBLE)",
		RunE:  makeRunSettingsForwardingDelete(factory),
	}
	cmd.Flags().String("email", "", "Forwarding email address (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func makeRunSettingsForwardingDelete(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		email, _ := cmd.Flags().GetString("email")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would delete forwarding address "+email, map[string]string{"forwardingEmail": email, "status": "deleted"})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		err = svc.Users.Settings.ForwardingAddresses.Delete("me", email).Do()
		if err != nil {
			return fmt.Errorf("deleting forwarding address %s: %w", email, err)
		}

		result := map[string]string{"forwardingEmail": email, "status": "deleted"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Forwarding address %s deleted\n", email)
		return nil
	}
}
