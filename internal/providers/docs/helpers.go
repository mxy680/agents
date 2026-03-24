package docs

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// DocumentSummary is the JSON-serializable summary of a document.
type DocumentSummary struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// DocumentDetail is the JSON-serializable detail of a document including body text.
type DocumentDetail struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Body  string `json:"body"`
}

// BatchUpdateResult is the JSON-serializable result of a batch update.
type BatchUpdateResult struct {
	DocumentID string `json:"documentId"`
	Replies    int    `json:"replies"`
}

// docURL returns the canonical edit URL for a document ID.
func docURL(id string) string {
	return fmt.Sprintf("https://docs.google.com/document/d/%s/edit", id)
}

// dryRunResult prints a standardised dry-run response and returns nil.
func dryRunResult(cmd *cobra.Command, description string, data any) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(data)
	}
	fmt.Printf("[DRY RUN] %s\n", description)
	return nil
}

// truncate shortens s to max characters, appending "..." if truncated.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
