package gmail

import (
	"context"
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
)

// FilterCriteria is the JSON-serializable representation of filter criteria.
type FilterCriteria struct {
	From    string `json:"from,omitempty"`
	To      string `json:"to,omitempty"`
	Subject string `json:"subject,omitempty"`
	Query   string `json:"query,omitempty"`
}

// FilterAction is the JSON-serializable representation of filter actions.
type FilterAction struct {
	AddLabelIds    []string `json:"addLabelIds,omitempty"`
	RemoveLabelIds []string `json:"removeLabelIds,omitempty"`
	Forward        string   `json:"forward,omitempty"`
}

// FilterInfo is the JSON-serializable representation of a Gmail filter.
type FilterInfo struct {
	ID       string         `json:"id"`
	Criteria FilterCriteria `json:"criteria"`
	Action   FilterAction   `json:"action"`
}

// filterFromAPI converts a Gmail API Filter to FilterInfo.
func filterFromAPI(f *api.Filter) FilterInfo {
	info := FilterInfo{ID: f.Id}
	if f.Criteria != nil {
		info.Criteria = FilterCriteria{
			From:    f.Criteria.From,
			To:      f.Criteria.To,
			Subject: f.Criteria.Subject,
			Query:   f.Criteria.Query,
		}
	}
	if f.Action != nil {
		info.Action = FilterAction{
			AddLabelIds:    f.Action.AddLabelIds,
			RemoveLabelIds: f.Action.RemoveLabelIds,
			Forward:        f.Action.Forward,
		}
	}
	return info
}

// criteriaSummary returns a short human-readable summary of filter criteria.
func criteriaSummary(c FilterCriteria) string {
	var parts []string
	if c.From != "" {
		parts = append(parts, "from:"+c.From)
	}
	if c.To != "" {
		parts = append(parts, "to:"+c.To)
	}
	if c.Subject != "" {
		parts = append(parts, "subject:"+c.Subject)
	}
	if c.Query != "" {
		parts = append(parts, "query:"+c.Query)
	}
	if len(parts) == 0 {
		return "(none)"
	}
	return strings.Join(parts, " ")
}

// actionSummary returns a short human-readable summary of filter actions.
func actionSummary(a FilterAction) string {
	var parts []string
	if len(a.AddLabelIds) > 0 {
		parts = append(parts, "add:"+strings.Join(a.AddLabelIds, ","))
	}
	if len(a.RemoveLabelIds) > 0 {
		parts = append(parts, "remove:"+strings.Join(a.RemoveLabelIds, ","))
	}
	if a.Forward != "" {
		parts = append(parts, "fwd:"+a.Forward)
	}
	if len(parts) == 0 {
		return "(none)"
	}
	return strings.Join(parts, " ")
}

// --- settings filters list ---

// newSettingsFiltersListCmd returns the `settings filters list` command.
func newSettingsFiltersListCmd(factory ServiceFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all email filters",
		RunE:  makeRunSettingsFiltersList(factory),
	}
}

func makeRunSettingsFiltersList(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Users.Settings.Filters.List("me").Do()
		if err != nil {
			return fmt.Errorf("listing filters: %w", err)
		}

		filters := make([]FilterInfo, 0, len(resp.Filter))
		for _, f := range resp.Filter {
			filters = append(filters, filterFromAPI(f))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(filters)
		}

		if len(filters) == 0 {
			fmt.Println("No filters found.")
			return nil
		}

		lines := make([]string, 0, len(filters)+1)
		lines = append(lines, fmt.Sprintf("%-20s  %-40s  %s", "ID", "CRITERIA", "ACTION"))
		for _, f := range filters {
			criteria := truncate(criteriaSummary(f.Criteria), 40)
			action := actionSummary(f.Action)
			lines = append(lines, fmt.Sprintf("%-20s  %-40s  %s", truncate(f.ID, 20), criteria, action))
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- settings filters get ---

// newSettingsFiltersGetCmd returns the `settings filters get` command.
func newSettingsFiltersGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a filter by ID",
		RunE:  makeRunSettingsFiltersGet(factory),
	}
	cmd.Flags().String("id", "", "Filter ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunSettingsFiltersGet(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		f, err := svc.Users.Settings.Filters.Get("me", id).Do()
		if err != nil {
			return fmt.Errorf("getting filter %s: %w", id, err)
		}

		info := filterFromAPI(f)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("ID:              %s", info.ID),
			fmt.Sprintf("Criteria From:   %s", info.Criteria.From),
			fmt.Sprintf("Criteria To:     %s", info.Criteria.To),
			fmt.Sprintf("Criteria Subject:%s", info.Criteria.Subject),
			fmt.Sprintf("Criteria Query:  %s", info.Criteria.Query),
			fmt.Sprintf("Action Add:      %s", strings.Join(info.Action.AddLabelIds, ", ")),
			fmt.Sprintf("Action Remove:   %s", strings.Join(info.Action.RemoveLabelIds, ", ")),
			fmt.Sprintf("Action Forward:  %s", info.Action.Forward),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- settings filters create ---

// newSettingsFiltersCreateCmd returns the `settings filters create` command.
func newSettingsFiltersCreateCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new email filter",
		RunE:  makeRunSettingsFiltersCreate(factory),
	}
	cmd.Flags().String("from", "", "Criteria: sender address")
	cmd.Flags().String("to", "", "Criteria: recipient address")
	cmd.Flags().String("subject", "", "Criteria: subject text")
	cmd.Flags().String("query", "", "Criteria: search query")
	cmd.Flags().StringSlice("add-label", nil, "Action: label IDs to add (repeatable)")
	cmd.Flags().StringSlice("remove-label", nil, "Action: label IDs to remove (repeatable)")
	cmd.Flags().String("forward-to", "", "Action: forward to this address")
	return cmd
}

func makeRunSettingsFiltersCreate(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		from, _ := cmd.Flags().GetString("from")
		to, _ := cmd.Flags().GetString("to")
		subject, _ := cmd.Flags().GetString("subject")
		query, _ := cmd.Flags().GetString("query")
		addLabels, _ := cmd.Flags().GetStringSlice("add-label")
		removeLabels, _ := cmd.Flags().GetStringSlice("remove-label")
		forwardTo, _ := cmd.Flags().GetString("forward-to")

		filter := &api.Filter{
			Criteria: &api.FilterCriteria{
				From:    from,
				To:      to,
				Subject: subject,
				Query:   query,
			},
			Action: &api.FilterAction{
				AddLabelIds:    addLabels,
				RemoveLabelIds: removeLabels,
				Forward:        forwardTo,
			},
		}

		if cli.IsDryRun(cmd) {
			info := FilterInfo{
				ID: "",
				Criteria: FilterCriteria{
					From:    from,
					To:      to,
					Subject: subject,
					Query:   query,
				},
				Action: FilterAction{
					AddLabelIds:    addLabels,
					RemoveLabelIds: removeLabels,
					Forward:        forwardTo,
				},
			}
			return dryRunResult(cmd, "Would create filter", map[string]any{
				"id":       "",
				"criteria": info.Criteria,
				"action":   info.Action,
				"status":   "created",
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		created, err := svc.Users.Settings.Filters.Create("me", filter).Do()
		if err != nil {
			return fmt.Errorf("creating filter: %w", err)
		}

		result := map[string]string{"id": created.Id, "status": "created"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Filter created (id: %s)\n", created.Id)
		return nil
	}
}

// --- settings filters delete ---

// newSettingsFiltersDeleteCmd returns the `settings filters delete` command.
func newSettingsFiltersDeleteCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an email filter (IRREVERSIBLE)",
		RunE:  makeRunSettingsFiltersDelete(factory),
	}
	cmd.Flags().String("id", "", "Filter ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunSettingsFiltersDelete(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would delete filter "+id, map[string]string{"id": id, "status": "deleted"})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		err = svc.Users.Settings.Filters.Delete("me", id).Do()
		if err != nil {
			return fmt.Errorf("deleting filter %s: %w", id, err)
		}

		result := map[string]string{"id": id, "status": "deleted"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Filter %s deleted\n", id)
		return nil
	}
}
