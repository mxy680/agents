package gmail

import (
	"context"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
)

// SendAsInfo is the JSON-serializable representation of a Gmail send-as alias.
type SendAsInfo struct {
	SendAsEmail        string `json:"sendAsEmail"`
	DisplayName        string `json:"displayName"`
	ReplyToAddress     string `json:"replyToAddress,omitempty"`
	Signature          string `json:"signature,omitempty"`
	IsPrimary          bool   `json:"isPrimary"`
	IsDefault          bool   `json:"isDefault"`
	VerificationStatus string `json:"verificationStatus"`
}

// sendAsFromAPI converts a Gmail API SendAs to SendAsInfo.
func sendAsFromAPI(sa *api.SendAs) SendAsInfo {
	return SendAsInfo{
		SendAsEmail:        sa.SendAsEmail,
		DisplayName:        sa.DisplayName,
		ReplyToAddress:     sa.ReplyToAddress,
		Signature:          sa.Signature,
		IsPrimary:          sa.IsPrimary,
		IsDefault:          sa.IsDefault,
		VerificationStatus: sa.VerificationStatus,
	}
}

// --- settings send-as list ---

// newSettingsSendAsListCmd returns the `settings send-as list` command.
func newSettingsSendAsListCmd(factory ServiceFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all send-as aliases",
		RunE:  makeRunSettingsSendAsList(factory),
	}
}

func makeRunSettingsSendAsList(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Users.Settings.SendAs.List("me").Do()
		if err != nil {
			return fmt.Errorf("listing send-as aliases: %w", err)
		}

		aliases := make([]SendAsInfo, 0, len(resp.SendAs))
		for _, sa := range resp.SendAs {
			aliases = append(aliases, sendAsFromAPI(sa))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(aliases)
		}

		if len(aliases) == 0 {
			fmt.Println("No send-as aliases found.")
			return nil
		}

		lines := make([]string, 0, len(aliases)+1)
		lines = append(lines, fmt.Sprintf("%-40s  %-30s  %-10s  %s", "SEND-AS EMAIL", "DISPLAY NAME", "DEFAULT", "VERIFICATION"))
		for _, a := range aliases {
			isDefault := "no"
			if a.IsDefault {
				isDefault = "yes"
			}
			lines = append(lines, fmt.Sprintf("%-40s  %-30s  %-10s  %s",
				truncate(a.SendAsEmail, 40),
				truncate(a.DisplayName, 30),
				isDefault,
				a.VerificationStatus,
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- settings send-as get ---

// newSettingsSendAsGetCmd returns the `settings send-as get` command.
func newSettingsSendAsGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a send-as alias",
		RunE:  makeRunSettingsSendAsGet(factory),
	}
	cmd.Flags().String("email", "", "Send-as email address (required)")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func makeRunSettingsSendAsGet(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		email, _ := cmd.Flags().GetString("email")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		sa, err := svc.Users.Settings.SendAs.Get("me", email).Do()
		if err != nil {
			return fmt.Errorf("getting send-as alias %s: %w", email, err)
		}

		info := sendAsFromAPI(sa)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("Send-As Email:       %s", info.SendAsEmail),
			fmt.Sprintf("Display Name:        %s", info.DisplayName),
			fmt.Sprintf("Reply-To Address:    %s", info.ReplyToAddress),
			fmt.Sprintf("Signature:           %s", info.Signature),
			fmt.Sprintf("Is Primary:          %v", info.IsPrimary),
			fmt.Sprintf("Is Default:          %v", info.IsDefault),
			fmt.Sprintf("Verification Status: %s", info.VerificationStatus),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- settings send-as create ---

// newSettingsSendAsCreateCmd returns the `settings send-as create` command.
func newSettingsSendAsCreateCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a send-as alias",
		RunE:  makeRunSettingsSendAsCreate(factory),
	}
	cmd.Flags().String("email", "", "Send-as email address (required)")
	cmd.Flags().String("display-name", "", "Display name for the alias")
	cmd.Flags().String("reply-to", "", "Reply-to address")
	cmd.Flags().String("signature", "", "HTML signature")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func makeRunSettingsSendAsCreate(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		email, _ := cmd.Flags().GetString("email")
		displayName, _ := cmd.Flags().GetString("display-name")
		replyTo, _ := cmd.Flags().GetString("reply-to")
		signature, _ := cmd.Flags().GetString("signature")

		if cli.IsDryRun(cmd) {
			result := map[string]string{"sendAsEmail": email, "status": "created"}
			return dryRunResult(cmd, fmt.Sprintf("Would create send-as alias %s", email), result)
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		created, err := svc.Users.Settings.SendAs.Create("me", &api.SendAs{
			SendAsEmail:    email,
			DisplayName:    displayName,
			ReplyToAddress: replyTo,
			Signature:      signature,
		}).Do()
		if err != nil {
			return fmt.Errorf("creating send-as alias %s: %w", email, err)
		}

		result := map[string]string{"sendAsEmail": created.SendAsEmail, "status": "created"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Send-as alias %s created\n", created.SendAsEmail)
		return nil
	}
}

// --- settings send-as update ---

// newSettingsSendAsUpdateCmd returns the `settings send-as update` command.
func newSettingsSendAsUpdateCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Full update of a send-as alias",
		RunE:  makeRunSettingsSendAsUpdate(factory),
	}
	cmd.Flags().String("email", "", "Send-as email address (required)")
	cmd.Flags().String("display-name", "", "Display name for the alias")
	cmd.Flags().String("reply-to", "", "Reply-to address")
	cmd.Flags().String("signature", "", "HTML signature")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func makeRunSettingsSendAsUpdate(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		email, _ := cmd.Flags().GetString("email")
		displayName, _ := cmd.Flags().GetString("display-name")
		replyTo, _ := cmd.Flags().GetString("reply-to")
		signature, _ := cmd.Flags().GetString("signature")

		if cli.IsDryRun(cmd) {
			result := map[string]string{"sendAsEmail": email, "status": "updated"}
			return dryRunResult(cmd, fmt.Sprintf("Would update send-as alias %s", email), result)
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		updated, err := svc.Users.Settings.SendAs.Update("me", email, &api.SendAs{
			SendAsEmail:    email,
			DisplayName:    displayName,
			ReplyToAddress: replyTo,
			Signature:      signature,
		}).Do()
		if err != nil {
			return fmt.Errorf("updating send-as alias %s: %w", email, err)
		}

		result := map[string]string{"sendAsEmail": updated.SendAsEmail, "status": "updated"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Send-as alias %s updated\n", updated.SendAsEmail)
		return nil
	}
}

// --- settings send-as patch ---

// newSettingsSendAsPatchCmd returns the `settings send-as patch` command.
func newSettingsSendAsPatchCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Partial update of a send-as alias",
		RunE:  makeRunSettingsSendAsPatch(factory),
	}
	cmd.Flags().String("email", "", "Send-as email address (required)")
	cmd.Flags().String("display-name", "", "Display name for the alias")
	cmd.Flags().String("reply-to", "", "Reply-to address")
	cmd.Flags().String("signature", "", "HTML signature")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func makeRunSettingsSendAsPatch(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		email, _ := cmd.Flags().GetString("email")
		displayName, _ := cmd.Flags().GetString("display-name")
		replyTo, _ := cmd.Flags().GetString("reply-to")
		signature, _ := cmd.Flags().GetString("signature")

		if cli.IsDryRun(cmd) {
			result := map[string]string{"sendAsEmail": email, "status": "patched"}
			return dryRunResult(cmd, fmt.Sprintf("Would patch send-as alias %s", email), result)
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		patched, err := svc.Users.Settings.SendAs.Patch("me", email, &api.SendAs{
			DisplayName:    displayName,
			ReplyToAddress: replyTo,
			Signature:      signature,
		}).Do()
		if err != nil {
			return fmt.Errorf("patching send-as alias %s: %w", email, err)
		}

		result := map[string]string{"sendAsEmail": patched.SendAsEmail, "status": "patched"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Send-as alias %s patched\n", patched.SendAsEmail)
		return nil
	}
}

// --- settings send-as delete ---

// newSettingsSendAsDeleteCmd returns the `settings send-as delete` command.
func newSettingsSendAsDeleteCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a send-as alias (IRREVERSIBLE)",
		RunE:  makeRunSettingsSendAsDelete(factory),
	}
	cmd.Flags().String("email", "", "Send-as email address (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func makeRunSettingsSendAsDelete(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		email, _ := cmd.Flags().GetString("email")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would delete send-as alias "+email, map[string]string{"sendAsEmail": email, "status": "deleted"})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		err = svc.Users.Settings.SendAs.Delete("me", email).Do()
		if err != nil {
			return fmt.Errorf("deleting send-as alias %s: %w", email, err)
		}

		result := map[string]string{"sendAsEmail": email, "status": "deleted"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Send-as alias %s deleted\n", email)
		return nil
	}
}

// --- settings send-as verify ---

// newSettingsSendAsVerifyCmd returns the `settings send-as verify` command.
func newSettingsSendAsVerifyCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Send verification email for a send-as alias",
		RunE:  makeRunSettingsSendAsVerify(factory),
	}
	cmd.Flags().String("email", "", "Send-as email address (required)")
	_ = cmd.MarkFlagRequired("email")
	return cmd
}

func makeRunSettingsSendAsVerify(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		email, _ := cmd.Flags().GetString("email")

		if cli.IsDryRun(cmd) {
			result := map[string]string{"sendAsEmail": email, "status": "verification-sent"}
			return dryRunResult(cmd, fmt.Sprintf("Would send verification email to %s", email), result)
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		err = svc.Users.Settings.SendAs.Verify("me", email).Do()
		if err != nil {
			return fmt.Errorf("sending verification for send-as alias %s: %w", email, err)
		}

		result := map[string]string{"sendAsEmail": email, "status": "verification-sent"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Verification email sent to %s\n", email)
		return nil
	}
}
