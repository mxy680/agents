package docs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	docsapi "google.golang.org/api/docs/v1"
)

// newDocumentsCreateCmd returns the `documents create` command.
func newDocumentsCreateCmd(factory DocsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new document",
		Long:  "Create a new blank Google Docs document with the given title.",
		RunE:  makeRunDocumentsCreate(factory),
	}
	cmd.Flags().String("title", "", "Title for the new document (required)")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func makeRunDocumentsCreate(factory DocsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		title, _ := cmd.Flags().GetString("title")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create document %q", title), map[string]string{
				"status": "dry-run",
				"title":  title,
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Documents.Create(&docsapi.Document{Title: title}).Do()
		if err != nil {
			return fmt.Errorf("creating document: %w", err)
		}

		result := DocumentSummary{
			ID:    resp.DocumentId,
			Title: resp.Title,
			URL:   docURL(resp.DocumentId),
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		cli.PrintText([]string{
			fmt.Sprintf("Created document %q", result.Title),
			fmt.Sprintf("ID:  %s", result.ID),
			fmt.Sprintf("URL: %s", result.URL),
		})
		return nil
	}
}

// newDocumentsGetCmd returns the `documents get` command.
func newDocumentsGetCmd(factory DocsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get document content",
		Long:  "Get a Google Docs document including its title and body text.",
		RunE:  makeRunDocumentsGet(factory),
	}
	cmd.Flags().String("document-id", "", "Document ID (required)")
	_ = cmd.MarkFlagRequired("document-id")
	return cmd
}

func makeRunDocumentsGet(factory DocsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		documentID, _ := cmd.Flags().GetString("document-id")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Documents.Get(documentID).Do()
		if err != nil {
			return fmt.Errorf("getting document: %w", err)
		}

		bodyText := extractBodyText(resp)

		detail := DocumentDetail{
			ID:    resp.DocumentId,
			Title: resp.Title,
			URL:   docURL(resp.DocumentId),
			Body:  bodyText,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("Title: %s", detail.Title),
			fmt.Sprintf("ID:    %s", detail.ID),
			fmt.Sprintf("URL:   %s", detail.URL),
			"",
			"Body:",
			detail.Body,
		}
		cli.PrintText(lines)
		return nil
	}
}

// extractBodyText extracts plain text content from the document body.
func extractBodyText(doc *docsapi.Document) string {
	if doc.Body == nil {
		return ""
	}
	var sb strings.Builder
	for _, elem := range doc.Body.Content {
		if elem.Paragraph != nil {
			for _, pe := range elem.Paragraph.Elements {
				if pe.TextRun != nil {
					sb.WriteString(pe.TextRun.Content)
				}
			}
		}
	}
	return sb.String()
}

// newDocumentsAppendCmd returns the `documents append` command.
func newDocumentsAppendCmd(factory DocsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "append",
		Short: "Append text to a document",
		Long:  "Append text to the end of a Google Docs document. Use \\n in --text for newlines.",
		RunE:  makeRunDocumentsAppend(factory),
	}
	cmd.Flags().String("document-id", "", "Document ID (required)")
	cmd.Flags().String("text", "", "Text to append (use \\n for newlines)")
	cmd.Flags().String("text-file", "", "Path to file containing text to append")
	_ = cmd.MarkFlagRequired("document-id")
	return cmd
}

func makeRunDocumentsAppend(factory DocsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		documentID, _ := cmd.Flags().GetString("document-id")
		textFlag, _ := cmd.Flags().GetString("text")
		textFile, _ := cmd.Flags().GetString("text-file")

		text, err := resolveText(textFlag, textFile)
		if err != nil {
			return err
		}
		if text == "" {
			return fmt.Errorf("--text or --text-file is required")
		}

		// Replace literal \n sequences with actual newlines.
		text = strings.ReplaceAll(text, `\n`, "\n")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would append text to document %s", documentID), map[string]string{
				"documentId": documentID,
				"status":     "dry-run",
				"text":       truncate(text, 80),
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		// Fetch the document to find the end index.
		doc, err := svc.Documents.Get(documentID).Do()
		if err != nil {
			return fmt.Errorf("getting document to find end index: %w", err)
		}

		endIndex := documentEndIndex(doc)

		req := &docsapi.BatchUpdateDocumentRequest{
			Requests: []*docsapi.Request{
				{
					InsertText: &docsapi.InsertTextRequest{
						Location: &docsapi.Location{Index: endIndex},
						Text:     text,
					},
				},
			},
		}

		resp, err := svc.Documents.BatchUpdate(documentID, req).Do()
		if err != nil {
			return fmt.Errorf("appending text: %w", err)
		}

		result := BatchUpdateResult{
			DocumentID: resp.DocumentId,
			Replies:    len(resp.Replies),
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		fmt.Printf("Appended text to document %s (%d request(s) applied)\n", result.DocumentID, result.Replies)
		return nil
	}
}

// documentEndIndex returns the index just before the final newline at the end
// of the document body, which is the safe insertion point for appending text.
func documentEndIndex(doc *docsapi.Document) int64 {
	if doc.Body == nil || len(doc.Body.Content) == 0 {
		return 1
	}
	// The last structural element's EndIndex is the document end.
	// We insert at EndIndex - 1 to stay before the trailing paragraph marker.
	last := doc.Body.Content[len(doc.Body.Content)-1]
	if last.EndIndex > 1 {
		return last.EndIndex - 1
	}
	return 1
}

// newDocumentsBatchUpdateCmd returns the `documents batch-update` command.
func newDocumentsBatchUpdateCmd(factory DocsServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-update",
		Short: "Apply batch update requests to a document",
		Long:  "Apply a JSON array of Google Docs API request objects to a document.",
		RunE:  makeRunDocumentsBatchUpdate(factory),
	}
	cmd.Flags().String("document-id", "", "Document ID (required)")
	cmd.Flags().String("requests", "", "JSON array of request objects")
	cmd.Flags().String("requests-file", "", "Path to file containing JSON array of request objects")
	_ = cmd.MarkFlagRequired("document-id")
	return cmd
}

func makeRunDocumentsBatchUpdate(factory DocsServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		documentID, _ := cmd.Flags().GetString("document-id")
		requestsFlag, _ := cmd.Flags().GetString("requests")
		requestsFile, _ := cmd.Flags().GetString("requests-file")

		raw, err := resolveText(requestsFlag, requestsFile)
		if err != nil {
			return err
		}
		if raw == "" {
			return fmt.Errorf("--requests or --requests-file is required")
		}

		var requests []*docsapi.Request
		if err := json.Unmarshal([]byte(raw), &requests); err != nil {
			return fmt.Errorf("invalid JSON for --requests: %w (expected array of request objects)", err)
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would apply %d request(s) to document %s", len(requests), documentID), map[string]any{
				"documentId": documentID,
				"status":     "dry-run",
				"count":      len(requests),
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Documents.BatchUpdate(documentID, &docsapi.BatchUpdateDocumentRequest{
			Requests: requests,
		}).Do()
		if err != nil {
			return fmt.Errorf("batch update: %w", err)
		}

		result := BatchUpdateResult{
			DocumentID: resp.DocumentId,
			Replies:    len(resp.Replies),
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		fmt.Printf("Applied %d request(s) to document %s\n", result.Replies, result.DocumentID)
		return nil
	}
}

// resolveText returns either the inline flag value or the contents of a file.
func resolveText(flagValue, filePath string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("reading file %s: %w", filePath, err)
		}
		return string(data), nil
	}
	return "", nil
}
