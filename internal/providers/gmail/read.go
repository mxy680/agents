package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
)

// EmailDetail is the JSON-serializable full email content.
type EmailDetail struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Date    string `json:"date"`
	Body    string `json:"body"`
}

func newReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read a full email by message ID",
		RunE:  runRead,
	}
	cmd.Flags().String("id", "", "Message ID to read (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func runRead(cmd *cobra.Command, _ []string) error {
	ctx := context.Background()
	msgID, _ := cmd.Flags().GetString("id")

	svc, err := auth.NewGmailService(ctx)
	if err != nil {
		return err
	}

	msg, err := svc.Users.Messages.Get("me", msgID).Format("full").Do()
	if err != nil {
		return fmt.Errorf("getting message %s: %w", msgID, err)
	}

	detail := extractDetail(msg)

	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(detail)
	}

	lines := []string{
		fmt.Sprintf("From:    %s", detail.From),
		fmt.Sprintf("To:      %s", detail.To),
		fmt.Sprintf("Subject: %s", detail.Subject),
		fmt.Sprintf("Date:    %s", detail.Date),
		"",
		detail.Body,
	}
	cli.PrintText(lines)
	return nil
}

func extractDetail(msg *api.Message) EmailDetail {
	detail := EmailDetail{ID: msg.Id}

	for _, h := range msg.Payload.Headers {
		switch h.Name {
		case "From":
			detail.From = h.Value
		case "To":
			detail.To = h.Value
		case "Subject":
			detail.Subject = h.Value
		case "Date":
			detail.Date = h.Value
		}
	}

	detail.Body = extractBody(msg.Payload)
	return detail
}

// extractBody walks MIME parts to find the text body.
// Prefers text/plain, falls back to text/html with tags stripped.
func extractBody(payload *api.MessagePart) string {
	if payload == nil {
		return ""
	}

	// Single-part message
	if len(payload.Parts) == 0 {
		return decodeBody(payload)
	}

	// Multi-part: look for text/plain first, then text/html
	var htmlBody string
	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" {
			if body := decodeBody(part); body != "" {
				return body
			}
		}
		if part.MimeType == "text/html" {
			htmlBody = decodeBody(part)
		}
		// Recurse into nested multipart
		if strings.HasPrefix(part.MimeType, "multipart/") {
			if body := extractBody(part); body != "" {
				return body
			}
		}
	}

	if htmlBody != "" {
		return stripHTMLTags(htmlBody)
	}
	return ""
}

func decodeBody(part *api.MessagePart) string {
	if part.Body == nil || part.Body.Data == "" {
		return ""
	}
	data, err := base64.URLEncoding.DecodeString(part.Body.Data)
	if err != nil {
		return ""
	}
	return string(data)
}

// stripHTMLTags does a basic removal of HTML tags.
func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result.WriteRune(r)
		}
	}
	return result.String()
}
