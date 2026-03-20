package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newSnippetsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "snippets",
		Aliases: []string{"snippet"},
		Short:   "SQL snippets",
	}
	cmd.AddCommand(newSnippetsListCmd(factory), newSnippetsGetCmd(factory))
	return cmd
}

// SnippetSummary is a lightweight representation of a SQL snippet.
type SnippetSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Content     string `json:"content,omitempty"`
}

// toSnippetSummary converts a raw API response map to a SnippetSummary.
func toSnippetSummary(data map[string]any) SnippetSummary {
	id, _ := data["id"].(string)
	name, _ := data["name"].(string)
	description, _ := data["description"].(string)
	content, _ := data["content"].(string)
	return SnippetSummary{
		ID:          id,
		Name:        name,
		Description: description,
		Content:     content,
	}
}

// printSnippetSummaries outputs snippet summaries as JSON or a formatted text table.
func printSnippetSummaries(cmd *cobra.Command, snippets []SnippetSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(snippets)
	}
	if len(snippets) == 0 {
		fmt.Println("No snippets found.")
		return nil
	}
	lines := make([]string, 0, len(snippets)+1)
	lines = append(lines, fmt.Sprintf("%-36s  %-40s  %s", "ID", "NAME", "DESCRIPTION"))
	for _, s := range snippets {
		lines = append(lines, fmt.Sprintf("%-36s  %-40s  %s",
			truncate(s.ID, 36), truncate(s.Name, 40), truncate(s.Description, 60)))
	}
	cli.PrintText(lines)
	return nil
}

// --- list ---

func newSnippetsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List SQL snippets",
		RunE:  makeRunSnippetsList(factory),
	}
	return cmd
}

func makeRunSnippetsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, "/snippets", nil)
		if err != nil {
			return fmt.Errorf("listing snippets: %w", err)
		}

		var raw []map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		snippets := make([]SnippetSummary, 0, len(raw))
		for _, r := range raw {
			snippets = append(snippets, toSnippetSummary(r))
		}
		return printSnippetSummaries(cmd, snippets)
	}
}

// --- get ---

func newSnippetsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a SQL snippet by ID",
		RunE:  makeRunSnippetsGet(factory),
	}
	cmd.Flags().String("snippet-id", "", "Snippet ID (required)")
	_ = cmd.MarkFlagRequired("snippet-id")
	return cmd
}

func makeRunSnippetsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		snippetID, _ := cmd.Flags().GetString("snippet-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/snippets/%s", snippetID), nil)
		if err != nil {
			return fmt.Errorf("getting snippet %s: %w", snippetID, err)
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		s := toSnippetSummary(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(s)
		}

		lines := []string{
			fmt.Sprintf("ID:          %s", s.ID),
			fmt.Sprintf("Name:        %s", s.Name),
			fmt.Sprintf("Description: %s", s.Description),
		}
		if s.Content != "" {
			lines = append(lines, fmt.Sprintf("Content:\n%s", s.Content))
		}
		cli.PrintText(lines)

		if s.Content != "" {
			var pretty bytes.Buffer
			if jsonErr := json.Indent(&pretty, []byte(s.Content), "", "  "); jsonErr != nil {
				// Content is not JSON — already printed above
			}
		}
		return nil
	}
}
