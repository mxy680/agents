package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
)

// SendResult is the JSON-serializable result of sending an email.
type SendResult struct {
	ID       string `json:"id"`
	ThreadID string `json:"threadId"`
	Status   string `json:"status"`
}

func newSendCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send an email via Gmail",
		RunE:  makeRunSend(factory),
	}
	cmd.Flags().String("to", "", "Recipient email address (required)")
	cmd.Flags().String("subject", "", "Email subject (required)")
	cmd.Flags().String("body", "", "Email body text")
	cmd.Flags().String("body-file", "", "Path to file containing email body")
	cmd.Flags().String("cc", "", "CC recipient email address")
	cmd.Flags().String("reply-to", "", "Message ID to reply to")
	_ = cmd.MarkFlagRequired("to")
	_ = cmd.MarkFlagRequired("subject")
	return cmd
}

func makeRunSend(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		to, _ := cmd.Flags().GetString("to")
		subject, _ := cmd.Flags().GetString("subject")
		body, _ := cmd.Flags().GetString("body")
		bodyFile, _ := cmd.Flags().GetString("body-file")
		cc, _ := cmd.Flags().GetString("cc")
		replyTo, _ := cmd.Flags().GetString("reply-to")

		if body == "" && bodyFile == "" {
			return fmt.Errorf("either --body or --body-file is required")
		}

		if bodyFile != "" {
			data, err := os.ReadFile(bodyFile)
			if err != nil {
				return fmt.Errorf("reading body file: %w", err)
			}
			body = string(data)
		}

		raw, err := composeMessage(ctx, factory, to, subject, body, cc, replyTo)
		if err != nil {
			return err
		}

		if cli.IsDryRun(cmd) {
			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]string{
					"status":  "dry-run",
					"raw":     raw,
					"to":      to,
					"subject": subject,
				})
			}
			fmt.Println("[DRY RUN] Would send:")
			fmt.Println(raw)
			return nil
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		msg := &api.Message{
			Raw: base64.URLEncoding.EncodeToString([]byte(raw)),
		}

		sent, err := svc.Users.Messages.Send("me", msg).Do()
		if err != nil {
			return fmt.Errorf("sending message: %w", err)
		}

		result := SendResult{
			ID:       sent.Id,
			ThreadID: sent.ThreadId,
			Status:   "sent",
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		fmt.Printf("Email sent to %s (id: %s)\n", to, sent.Id)
		return nil
	}
}

func composeMessage(ctx context.Context, factory ServiceFactory, to, subject, body, cc, replyTo string) (string, error) {
	var headers []string
	headers = append(headers, fmt.Sprintf("To: %s", to))
	headers = append(headers, fmt.Sprintf("Subject: %s", subject))
	headers = append(headers, "MIME-Version: 1.0")
	headers = append(headers, "Content-Type: text/plain; charset=\"UTF-8\"")

	if cc != "" {
		headers = append(headers, fmt.Sprintf("Cc: %s", cc))
	}

	if replyTo != "" {
		// Fetch the original message to get headers for threading
		svc, err := factory(ctx)
		if err != nil {
			return "", err
		}
		original, err := svc.Users.Messages.Get("me", replyTo).Format("metadata").MetadataHeaders("Message-ID").Do()
		if err != nil {
			return "", fmt.Errorf("getting reply-to message %s: %w", replyTo, err)
		}
		for _, h := range original.Payload.Headers {
			if h.Name == "Message-ID" || h.Name == "Message-Id" {
				headers = append(headers, fmt.Sprintf("In-Reply-To: %s", h.Value))
				headers = append(headers, fmt.Sprintf("References: %s", h.Value))
				break
			}
		}
	}

	raw := strings.Join(headers, "\r\n") + "\r\n\r\n" + body
	return raw, nil
}
