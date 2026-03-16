package gmail

import (
	"context"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
)

// DelegateInfo is the JSON-serializable representation of a Gmail delegate.
type DelegateInfo struct {
	DelegateEmail      string `json:"delegateEmail"`
	VerificationStatus string `json:"verificationStatus"`
}

// delegateFromAPI converts a Gmail API Delegate to DelegateInfo.
func delegateFromAPI(d *api.Delegate) DelegateInfo {
	return DelegateInfo{
		DelegateEmail:      d.DelegateEmail,
		VerificationStatus: d.VerificationStatus,
	}
}

// --- settings delegates list ---

// newSettingsDelegatesListCmd returns the `settings delegates list` command.
func newSettingsDelegatesListCmd(factory ServiceFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all delegates",
		RunE:  makeRunSettingsDelegatesList(factory),
	}
}

func makeRunSettingsDelegatesList(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Users.Settings.Delegates.List("me").Do()
		if err != nil {
			return fmt.Errorf("listing delegates: %w", err)
		}

		delegates := make([]DelegateInfo, 0, len(resp.Delegates))
		for _, d := range resp.Delegates {
			delegates = append(delegates, delegateFromAPI(d))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(delegates)
		}

		if len(delegates) == 0 {
			fmt.Println("No delegates found.")
			return nil
		}

		lines := make([]string, 0, len(delegates)+1)
		lines = append(lines, fmt.Sprintf("%-40s  %s", "DELEGATE EMAIL", "VERIFICATION STATUS"))
		for _, d := range delegates {
			lines = append(lines, fmt.Sprintf("%-40s  %s", truncate(d.DelegateEmail, 40), d.VerificationStatus))
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- settings delegates get ---

// newSettingsDelegatesGetCmd returns the `settings delegates get` command.
func newSettingsDelegatesGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a delegate",
		RunE:  makeRunSettingsDelegatesGet(factory),
	}
	cmd.Flags().String("email", "", "Delegate email address (required)")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func makeRunSettingsDelegatesGet(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		email, _ := cmd.Flags().GetString("email")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		d, err := svc.Users.Settings.Delegates.Get("me", email).Do()
		if err != nil {
			return fmt.Errorf("getting delegate %s: %w", email, err)
		}

		info := delegateFromAPI(d)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("Delegate Email:      %s", info.DelegateEmail),
			fmt.Sprintf("Verification Status: %s", info.VerificationStatus),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- settings delegates create ---

// newSettingsDelegatesCreateCmd returns the `settings delegates create` command.
func newSettingsDelegatesCreateCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a delegate (grants account access — IRREVERSIBLE)",
		RunE:  makeRunSettingsDelegatesCreate(factory),
	}
	cmd.Flags().String("email", "", "Delegate email address (required)")
	cmd.Flags().Bool("confirm", false, "Confirm granting account access to this delegate")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func makeRunSettingsDelegatesCreate(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		email, _ := cmd.Flags().GetString("email")

		if cli.IsDryRun(cmd) {
			result := map[string]string{
				"delegateEmail":      email,
				"verificationStatus": "pending",
				"status":             "created",
			}
			return dryRunResult(cmd, fmt.Sprintf("Would create delegate %s (grants account access)", email), result)
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		created, err := svc.Users.Settings.Delegates.Create("me", &api.Delegate{
			DelegateEmail: email,
		}).Do()
		if err != nil {
			return fmt.Errorf("creating delegate %s: %w", email, err)
		}

		result := map[string]string{
			"delegateEmail":      created.DelegateEmail,
			"verificationStatus": created.VerificationStatus,
			"status":             "created",
		}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Delegate %s created\n", created.DelegateEmail)
		return nil
	}
}

// --- settings delegates delete ---

// newSettingsDelegatesDeleteCmd returns the `settings delegates delete` command.
func newSettingsDelegatesDeleteCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a delegate (IRREVERSIBLE)",
		RunE:  makeRunSettingsDelegatesDelete(factory),
	}
	cmd.Flags().String("email", "", "Delegate email address (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func makeRunSettingsDelegatesDelete(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		email, _ := cmd.Flags().GetString("email")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would delete delegate "+email, map[string]string{"delegateEmail": email, "status": "deleted"})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		err = svc.Users.Settings.Delegates.Delete("me", email).Do()
		if err != nil {
			return fmt.Errorf("deleting delegate %s: %w", email, err)
		}

		result := map[string]string{"delegateEmail": email, "status": "deleted"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Delegate %s deleted\n", email)
		return nil
	}
}
