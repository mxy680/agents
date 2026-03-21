package imessage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newContactsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contacts",
		Short:   "Manage contacts",
		Aliases: []string{"contact"},
	}

	cmd.AddCommand(newContactsListCmd(factory))
	cmd.AddCommand(newContactsGetCmd(factory))
	cmd.AddCommand(newContactsCreateCmd(factory))

	return cmd
}

func newContactsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all contacts",
		RunE:  makeRunContactsList(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunContactsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Get(ctx, "/contact", nil)
		if err != nil {
			return fmt.Errorf("listing contacts: %w", err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		var contacts []json.RawMessage
		if err := json.Unmarshal(data, &contacts); err != nil {
			return fmt.Errorf("parse contacts list: %w", err)
		}

		summaries := make([]ContactSummary, 0, len(contacts))
		for _, c := range contacts {
			summaries = append(summaries, toContactSummary(c))
		}

		lines := make([]string, 0, len(summaries))
		for _, s := range summaries {
			name := s.DisplayName
			if name == "" {
				name = s.FirstName + " " + s.LastName
			}
			contact := truncate(name, 30)
			phones := ""
			if len(s.Phones) > 0 {
				phones = s.Phones[0]
			}
			lines = append(lines, fmt.Sprintf("%-32s  %s", contact, phones))
		}
		return printResult(cmd, summaries, lines)
	}
}

func newContactsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Search for a contact by query",
		RunE:  makeRunContactsGet(factory),
	}
	cmd.Flags().String("query", "", "Search term (name, phone, or email)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunContactsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		reqBody := map[string]any{}
		if query != "" {
			reqBody["query"] = query
		}

		body, err := client.Post(ctx, "/contact/query", reqBody)
		if err != nil {
			return fmt.Errorf("querying contacts: %w", err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		var contacts []json.RawMessage
		if err := json.Unmarshal(data, &contacts); err != nil {
			return fmt.Errorf("parse contact query results: %w", err)
		}

		summaries := make([]ContactSummary, 0, len(contacts))
		for _, c := range contacts {
			summaries = append(summaries, toContactSummary(c))
		}

		lines := make([]string, 0, len(summaries))
		for _, s := range summaries {
			name := s.DisplayName
			if name == "" {
				name = s.FirstName + " " + s.LastName
			}
			phones := ""
			if len(s.Phones) > 0 {
				phones = s.Phones[0]
			}
			emails := ""
			if len(s.Emails) > 0 {
				emails = s.Emails[0]
			}
			line := fmt.Sprintf("%-32s  %s", truncate(name, 30), phones)
			if emails != "" {
				line += fmt.Sprintf("  %s", emails)
			}
			lines = append(lines, line)
		}
		return printResult(cmd, summaries, lines)
	}
}

func newContactsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new contact",
		RunE:  makeRunContactsCreate(factory),
	}
	cmd.Flags().String("data", "", "Contact data as JSON ({firstName, lastName, phoneNumbers, emails})")
	cmd.Flags().String("data-file", "", "Path to JSON file with contact data")
	cmd.Flags().Bool("dry-run", false, "Show what would be created without sending")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunContactsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		dataStr, _ := cmd.Flags().GetString("data")
		dataFile, _ := cmd.Flags().GetString("data-file")

		var reqBody map[string]any

		switch {
		case dataStr != "":
			if err := json.Unmarshal([]byte(dataStr), &reqBody); err != nil {
				return fmt.Errorf("parse --data JSON: %w", err)
			}
		case dataFile != "":
			fileBytes, err := os.ReadFile(dataFile)
			if err != nil {
				return fmt.Errorf("read --data-file %q: %w", dataFile, err)
			}
			if err := json.Unmarshal(fileBytes, &reqBody); err != nil {
				return fmt.Errorf("parse --data-file JSON: %w", err)
			}
		default:
			return fmt.Errorf("one of --data or --data-file is required")
		}

		if cli.IsDryRun(cmd) {
			result := dryRunResult("create", map[string]any{"contact": reqBody})
			return printResult(cmd, result, []string{
				fmt.Sprintf("[dry-run] Would create contact: %v", reqBody),
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Post(ctx, "/contact", reqBody)
		if err != nil {
			return fmt.Errorf("creating contact: %w", err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		summary := toContactSummary(data)
		name := summary.DisplayName
		if name == "" {
			name = summary.FirstName + " " + summary.LastName
		}
		return printResult(cmd, summary, []string{
			fmt.Sprintf("Created contact: %s (ID: %s)", name, summary.ID),
		})
	}
}
