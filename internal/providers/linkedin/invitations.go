package linkedin

import (
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerInvitationsResponse is the response envelope for invitation list endpoints.
type voyagerInvitationsResponse struct {
	Elements []struct {
		Invitation struct {
			InvitationID string `json:"invitationId"`
			SharedSecret string `json:"sharedSecret"`
			SentTime     int64  `json:"sentTime"`
			Message      string `json:"message"`
			InviterResolved struct {
				EntityURN string `json:"entityUrn"`
				FirstName string `json:"firstName"`
				LastName  string `json:"lastName"`
			} `json:"inviterResolved"`
		} `json:"invitation"`
	} `json:"elements"`
	Paging struct {
		Start int `json:"start"`
		Count int `json:"count"`
		Total int `json:"total"`
	} `json:"paging"`
}

// newInvitationsCmd builds the "invitations" subcommand group.
func newInvitationsCmd(factory ClientFactory) *cobra.Command {
	invitationsCmd := &cobra.Command{
		Use:     "invitations",
		Short:   "Manage LinkedIn invitations",
		Aliases: []string{"invite"},
	}
	invitationsCmd.AddCommand(newInvitationsListCmd(factory))
	invitationsCmd.AddCommand(newInvitationsSendCmd(factory))
	invitationsCmd.AddCommand(newInvitationsAcceptCmd(factory))
	invitationsCmd.AddCommand(newInvitationsRejectCmd(factory))
	invitationsCmd.AddCommand(newInvitationsWithdrawCmd(factory))
	return invitationsCmd
}

// newInvitationsListCmd builds the "invitations list" command.
func newInvitationsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List LinkedIn invitations",
		Long:  "List received or sent LinkedIn connection invitations.",
		RunE:  makeRunInvitationsList(factory),
	}
	cmd.Flags().String("direction", "received", "Direction: received or sent")
	cmd.Flags().Int("limit", 10, "Maximum number of invitations to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	return cmd
}

// newInvitationsSendCmd builds the "invitations send" command.
func newInvitationsSendCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a LinkedIn connection invitation",
		Long:  "Send a connection request to a LinkedIn member by their public profile ID.",
		RunE:  makeRunInvitationsSend(factory),
	}
	cmd.Flags().String("urn", "", "Public profile ID of the person to invite (URL slug)")
	cmd.Flags().String("message", "", "Optional personal note to include with the invitation")
	cmd.Flags().Bool("dry-run", false, "Preview the action without making changes")
	return cmd
}

// newInvitationsAcceptCmd builds the "invitations accept" command.
func newInvitationsAcceptCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accept",
		Short: "Accept a received LinkedIn invitation",
		RunE:  makeRunInvitationsAction(factory, "accept"),
	}
	cmd.Flags().String("invitation-id", "", "Invitation ID to accept")
	cmd.Flags().Bool("dry-run", false, "Preview the action without making changes")
	return cmd
}

// newInvitationsRejectCmd builds the "invitations reject" command.
func newInvitationsRejectCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reject",
		Short: "Reject a received LinkedIn invitation",
		RunE:  makeRunInvitationsAction(factory, "ignore"),
	}
	cmd.Flags().String("invitation-id", "", "Invitation ID to reject")
	cmd.Flags().Bool("dry-run", false, "Preview the action without making changes")
	return cmd
}

// newInvitationsWithdrawCmd builds the "invitations withdraw" command.
func newInvitationsWithdrawCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "Withdraw a sent LinkedIn invitation",
		RunE:  makeRunInvitationsWithdraw(factory),
	}
	cmd.Flags().String("invitation-id", "", "Invitation ID to withdraw")
	cmd.Flags().Bool("dry-run", false, "Preview the action without making changes")
	return cmd
}

func makeRunInvitationsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		direction, _ := cmd.Flags().GetString("direction")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		start := 0
		if cursor != "" {
			if _, err := fmt.Sscanf(cursor, "%d", &start); err != nil {
				return fmt.Errorf("invalid cursor %q: must be a numeric start offset", cursor)
			}
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var path string
		params := url.Values{
			"start": {fmt.Sprintf("%d", start)},
			"count": {fmt.Sprintf("%d", limit)},
		}

		switch direction {
		case "sent":
			path = "/voyager/api/relationships/sentInvitationViewsV2"
		default:
			path = "/voyager/api/relationships/invitationViews"
			params.Set("q", "receivedInvitation")
		}

		resp, err := client.Get(ctx, path, params)
		if err != nil {
			return fmt.Errorf("listing %s invitations: %w", direction, err)
		}

		var raw voyagerInvitationsResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding invitations: %w", err)
		}

		summaries := make([]InvitationSummary, 0, len(raw.Elements))
		for _, el := range raw.Elements {
			inv := el.Invitation
			summaries = append(summaries, InvitationSummary{
				ID:        inv.InvitationID,
				Direction: direction,
				FromURN:   inv.InviterResolved.EntityURN,
				FromName:  inv.InviterResolved.FirstName + " " + inv.InviterResolved.LastName,
				Message:   inv.Message,
				SentAt:    inv.SentTime,
			})
		}

		return printInvitationSummaries(cmd, summaries)
	}
}

func makeRunInvitationsSend(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		urn, _ := cmd.Flags().GetString("urn")
		if urn == "" {
			return fmt.Errorf("--urn is required")
		}
		message, _ := cmd.Flags().GetString("message")

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("Send invitation to: %s", urn), map[string]string{"urn": urn, "message": message})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"invitee": map[string]any{
				"com.linkedin.voyager.relationships.invitation.ProfileInvitee": map[string]string{
					"profileId": urn,
				},
			},
			"message": message,
		}
		resp, err := client.PostJSON(ctx, "/voyager/api/relationships/invitation", body)
		if err != nil {
			return fmt.Errorf("sending invitation to %s: %w", urn, err)
		}

		if err := client.DecodeJSON(resp, &struct{}{}); err != nil {
			return fmt.Errorf("sending invitation: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "sent", "urn": urn})
		}
		fmt.Printf("Invitation sent to: %s\n", urn)
		return nil
	}
}

// makeRunInvitationsAction handles accept and reject (ignore) for received invitations.
func makeRunInvitationsAction(factory ClientFactory, action string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		invitationID, _ := cmd.Flags().GetString("invitation-id")
		if invitationID == "" {
			return fmt.Errorf("--invitation-id is required")
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("%s invitation: %s", action, invitationID), map[string]string{"invitation_id": invitationID, "action": action})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/relationships/invitations/" + url.PathEscape(invitationID)
		body := map[string]string{"action": action}
		resp, err := client.PutJSON(ctx, path, body)
		if err != nil {
			return fmt.Errorf("%s invitation %s: %w", action, invitationID, err)
		}

		if err := client.DecodeJSON(resp, &struct{}{}); err != nil {
			return fmt.Errorf("%s invitation: %w", action, err)
		}

		displayAction := action
		if action == "ignore" {
			displayAction = "rejected"
		} else {
			displayAction = "accepted"
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": displayAction, "invitation_id": invitationID})
		}
		fmt.Printf("Invitation %s: %s\n", displayAction, invitationID)
		return nil
	}
}

func makeRunInvitationsWithdraw(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		invitationID, _ := cmd.Flags().GetString("invitation-id")
		if invitationID == "" {
			return fmt.Errorf("--invitation-id is required")
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("Withdraw invitation: %s", invitationID), map[string]string{"invitation_id": invitationID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/relationships/sentInvitationViewsV2/" + url.PathEscape(invitationID)
		resp, err := client.Delete(ctx, path)
		if err != nil {
			return fmt.Errorf("withdrawing invitation %s: %w", invitationID, err)
		}

		if err := client.DecodeJSON(resp, &struct{}{}); err != nil {
			return fmt.Errorf("withdrawing invitation: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "withdrawn", "invitation_id": invitationID})
		}
		fmt.Printf("Invitation withdrawn: %s\n", invitationID)
		return nil
	}
}

// printInvitationSummaries outputs invitation summaries as JSON or text.
func printInvitationSummaries(cmd *cobra.Command, invitations []InvitationSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(invitations)
	}
	if len(invitations) == 0 {
		fmt.Println("No invitations found.")
		return nil
	}
	lines := make([]string, 0, len(invitations)+1)
	lines = append(lines, fmt.Sprintf("%-12s  %-10s  %-30s  %-40s  %-16s", "ID", "DIRECTION", "FROM", "MESSAGE", "SENT AT"))
	for _, inv := range invitations {
		lines = append(lines, fmt.Sprintf("%-12s  %-10s  %-30s  %-40s  %-16s",
			truncate(inv.ID, 12),
			truncate(inv.Direction, 10),
			truncate(inv.FromName, 30),
			truncate(inv.Message, 40),
			formatTimestamp(inv.SentAt),
		))
	}
	cli.PrintText(lines)
	return nil
}
