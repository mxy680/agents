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
			// sentInvitationViewsV2 returned 400; use sentInvitationViews instead.
			path = "/voyager/api/relationships/sentInvitationViews"
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

// newInvitationsSendCmd builds the "invitations send" command.
func newInvitationsSendCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a connection invitation",
		RunE:  makeRunInvitationsSend(factory),
	}
	cmd.Flags().String("urn", "", "Profile URN or public identifier (required)")
	cmd.Flags().String("message", "", "Optional invitation message")
	cmd.Flags().Bool("dry-run", false, "Preview action without executing it")
	_ = cmd.MarkFlagRequired("urn")
	return cmd
}

func makeRunInvitationsSend(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		urn, _ := cmd.Flags().GetString("urn")
		message, _ := cmd.Flags().GetString("message")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("send invitation to %s", urn), map[string]string{"urn": urn, "message": message})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"emberInvitee": map[string]any{
				"com.linkedin.voyager.growth.invitation.InviteToConnect": map[string]any{
					"memberUrn": urn,
				},
			},
			"message": message,
		}
		_, err = client.PostJSON(ctx, "/voyager/api/relationships/invitation", body)
		if err != nil {
			return fmt.Errorf("sending invitation to %s: %w", urn, err)
		}

		fmt.Printf("Invitation sent to %s\n", urn)
		return nil
	}
}

// newInvitationsAcceptCmd builds the "invitations accept" command.
func newInvitationsAcceptCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accept",
		Short: "Accept a received invitation",
		RunE:  makeRunInvitationsAccept(factory),
	}
	cmd.Flags().String("invitation-id", "", "Invitation ID (required)")
	cmd.Flags().String("shared-secret", "", "Invitation shared secret")
	cmd.Flags().Bool("dry-run", false, "Preview action without executing it")
	_ = cmd.MarkFlagRequired("invitation-id")
	return cmd
}

func makeRunInvitationsAccept(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		invitationID, _ := cmd.Flags().GetString("invitation-id")
		sharedSecret, _ := cmd.Flags().GetString("shared-secret")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("accept invitation %s", invitationID),
				map[string]string{"status": "accepted", "invitation_id": invitationID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"invitationId":       invitationID,
			"sharedSecret":       sharedSecret,
			"isGenericInvitation": false,
		}
		path := "/voyager/api/relationships/invitations/" + url.PathEscape(invitationID)
		_, err = client.PutJSON(ctx, path, body)
		if err != nil {
			return fmt.Errorf("accepting invitation %s: %w", invitationID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "accepted", "invitation_id": invitationID})
		}
		fmt.Printf("Invitation %s accepted\n", invitationID)
		return nil
	}
}

// newInvitationsRejectCmd builds the "invitations reject" command.
func newInvitationsRejectCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reject",
		Short: "Reject a received invitation",
		RunE:  makeRunInvitationsReject(factory),
	}
	cmd.Flags().String("invitation-id", "", "Invitation ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview action without executing it")
	_ = cmd.MarkFlagRequired("invitation-id")
	return cmd
}

func makeRunInvitationsReject(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		invitationID, _ := cmd.Flags().GetString("invitation-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("reject invitation %s", invitationID),
				map[string]string{"status": "rejected", "invitation_id": invitationID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"invitationId":       invitationID,
			"isGenericInvitation": false,
		}
		path := "/voyager/api/relationships/invitations/" + url.PathEscape(invitationID)
		_, err = client.PutJSON(ctx, path, body)
		if err != nil {
			return fmt.Errorf("rejecting invitation %s: %w", invitationID, err)
		}

		fmt.Printf("Invitation %s rejected\n", invitationID)
		return nil
	}
}

// newInvitationsWithdrawCmd builds the "invitations withdraw" command.
func newInvitationsWithdrawCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "Withdraw a sent invitation",
		RunE:  makeRunInvitationsWithdraw(factory),
	}
	cmd.Flags().String("invitation-id", "", "Invitation ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview action without executing it")
	_ = cmd.MarkFlagRequired("invitation-id")
	return cmd
}

func makeRunInvitationsWithdraw(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		invitationID, _ := cmd.Flags().GetString("invitation-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("withdraw invitation %s", invitationID),
				map[string]string{"status": "withdrawn", "invitation_id": invitationID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/relationships/sentInvitationViewsV2/" + url.PathEscape(invitationID)
		_, err = client.Delete(ctx, path)
		if err != nil {
			return fmt.Errorf("withdrawing invitation %s: %w", invitationID, err)
		}

		fmt.Printf("Invitation %s withdrawn\n", invitationID)
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
