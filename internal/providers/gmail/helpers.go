package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
)

// EmailSummary is the JSON-serializable summary of an email.
type EmailSummary struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	Subject string `json:"subject"`
	Snippet string `json:"snippet"`
	Date    string `json:"date"`
}

// EmailDetail is the JSON-serializable full email content.
type EmailDetail struct {
	ID      string `json:"id"`
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Date    string `json:"date"`
	Body    string `json:"body"`
}

// SendResult is the JSON-serializable result of sending an email.
type SendResult struct {
	ID       string `json:"id"`
	ThreadID string `json:"threadId"`
	Status   string `json:"status"`
}

// truncate shortens s to at most max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
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

// decodeBody base64url-decodes the body data of a message part.
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

// parseSinceDuration parses a duration string like "24h" or "7d".
func parseSinceDuration(s string) (time.Duration, error) {
	// Handle day shorthand
	if len(s) > 0 && s[len(s)-1] == 'd' {
		s = s[:len(s)-1] + "h"
		d, err := time.ParseDuration(s)
		if err != nil {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}
		return d * 24, nil
	}
	return time.ParseDuration(s)
}

// fetchSummaries retrieves metadata for a list of messages and returns EmailSummary values.
func fetchSummaries(_ context.Context, svc *api.Service, msgs []*api.Message) ([]EmailSummary, error) {
	summaries := make([]EmailSummary, 0, len(msgs))
	for _, m := range msgs {
		msg, err := svc.Users.Messages.Get("me", m.Id).Format("metadata").MetadataHeaders("From", "Subject", "Date").Do()
		if err != nil {
			return nil, fmt.Errorf("getting message %s: %w", m.Id, err)
		}
		headers := extractHeaders(msg.Payload.Headers, "From", "Subject", "Date")
		summaries = append(summaries, EmailSummary{
			ID:      msg.Id,
			Snippet: msg.Snippet,
			From:    headers["From"],
			Subject: headers["Subject"],
			Date:    headers["Date"],
		})
	}
	return summaries, nil
}

// extractDetail builds an EmailDetail from a full Gmail message.
func extractDetail(msg *api.Message) EmailDetail {
	headers := extractHeaders(msg.Payload.Headers, "From", "To", "Subject", "Date")
	return EmailDetail{
		ID:      msg.Id,
		From:    headers["From"],
		To:      headers["To"],
		Subject: headers["Subject"],
		Date:    headers["Date"],
		Body:    extractBody(msg.Payload),
	}
}

// composeMessage builds a raw RFC 2822 message string ready for Gmail's send API.
// It optionally fetches threading headers when replyTo is non-empty.
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

// printSummaries outputs summaries as JSON or a formatted text table.
func printSummaries(cmd *cobra.Command, summaries []EmailSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(summaries)
	}

	if len(summaries) == 0 {
		fmt.Println("No messages found.")
		return nil
	}

	lines := make([]string, 0, len(summaries)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-40s  %s", "FROM", "SUBJECT", "DATE"))
	for _, s := range summaries {
		from := truncate(s.From, 20)
		subject := truncate(s.Subject, 40)
		lines = append(lines, fmt.Sprintf("%-20s  %-40s  %s", from, subject, s.Date))
	}
	cli.PrintText(lines)
	return nil
}

// extractHeaders returns a map of header name → value for the requested names.
// Names not present in headers will have an empty string value.
func extractHeaders(headers []*api.MessagePartHeader, names ...string) map[string]string {
	result := make(map[string]string, len(names))
	for _, h := range headers {
		for _, name := range names {
			if h.Name == name {
				result[name] = h.Value
			}
		}
	}
	return result
}

// confirmDestructive returns an error if the --confirm flag is absent or false.
// Use this for commands that perform irreversible actions.
func confirmDestructive(cmd *cobra.Command) error {
	confirmed, _ := cmd.Flags().GetBool("confirm")
	if !confirmed {
		return fmt.Errorf("this action is irreversible; re-run with --confirm to proceed")
	}
	return nil
}

// dryRunResult prints a standardised dry-run response and returns nil.
// description is a human-readable summary of what would have happened.
// data is serialised as JSON when --json is set, otherwise description is printed.
func dryRunResult(cmd *cobra.Command, description string, data any) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(data)
	}
	fmt.Printf("[DRY RUN] %s\n", description)
	return nil
}
