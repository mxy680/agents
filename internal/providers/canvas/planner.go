package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newPlannerCmd returns the parent "planner" command with all subcommands attached.
func newPlannerCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "planner",
		Short:   "Manage Canvas planner items, notes, and overrides",
		Aliases: []string{"plan"},
	}

	cmd.AddCommand(newPlannerListCmd(factory))
	cmd.AddCommand(newPlannerNotesCmd(factory))
	cmd.AddCommand(newPlannerCreateNoteCmd(factory))
	cmd.AddCommand(newPlannerUpdateNoteCmd(factory))
	cmd.AddCommand(newPlannerDeleteNoteCmd(factory))
	cmd.AddCommand(newPlannerOverridesCmd(factory))
	cmd.AddCommand(newPlannerOverrideCmd(factory))

	return cmd
}

func newPlannerListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List planner items for the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			startDate, _ := cmd.Flags().GetString("start-date")
			endDate, _ := cmd.Flags().GetString("end-date")
			contextCodes, _ := cmd.Flags().GetStringSlice("context-codes")
			limit, _ := cmd.Flags().GetInt("limit")

			params := url.Values{}
			if startDate != "" {
				params.Set("start_date", startDate)
			}
			if endDate != "" {
				params.Set("end_date", endDate)
			}
			for _, code := range contextCodes {
				params.Add("context_codes[]", code)
			}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/planner/items", params)
			if err != nil {
				return err
			}

			var items []map[string]any
			if err := json.Unmarshal(data, &items); err != nil {
				return fmt.Errorf("parse planner items: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(items)
			}

			if len(items) == 0 {
				fmt.Println("No planner items found.")
				return nil
			}
			for _, item := range items {
				plannable, _ := item["plannable"].(map[string]any)
				title := ""
				if plannable != nil {
					if t, ok := plannable["title"].(string); ok {
						title = t
					} else if t, ok := plannable["name"].(string); ok {
						title = t
					}
				}
				plannableType, _ := item["plannable_type"].(string)
				plannableDate, _ := item["plannable_date"].(string)
				fmt.Printf("%-20s  %-12s  %s\n", plannableDate, plannableType, truncate(title, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("start-date", "", "Start date (RFC3339) for filtering items")
	cmd.Flags().String("end-date", "", "End date (RFC3339) for filtering items")
	cmd.Flags().StringSlice("context-codes", nil, "Context codes to filter by (e.g. course_123)")
	cmd.Flags().Int("limit", 0, "Maximum number of items to return")
	return cmd
}

func newPlannerNotesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notes",
		Short: "List planner notes for the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			startDate, _ := cmd.Flags().GetString("start-date")
			endDate, _ := cmd.Flags().GetString("end-date")

			params := url.Values{}
			if startDate != "" {
				params.Set("start_date", startDate)
			}
			if endDate != "" {
				params.Set("end_date", endDate)
			}

			data, err := client.Get(ctx, "/planner/notes", params)
			if err != nil {
				return err
			}

			var notes []PlannerNoteSummary
			if err := json.Unmarshal(data, &notes); err != nil {
				return fmt.Errorf("parse planner notes: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(notes)
			}

			if len(notes) == 0 {
				fmt.Println("No planner notes found.")
				return nil
			}
			for _, n := range notes {
				fmt.Printf("%-6d  %-20s  %s\n", n.ID, n.TodoDate, truncate(n.Title, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("start-date", "", "Start date (RFC3339) for filtering notes")
	cmd.Flags().String("end-date", "", "End date (RFC3339) for filtering notes")
	return cmd
}

func newPlannerCreateNoteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-note",
		Short: "Create a new planner note",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			title, _ := cmd.Flags().GetString("title")
			if title == "" {
				return fmt.Errorf("--title is required")
			}

			details, _ := cmd.Flags().GetString("details")
			courseID, _ := cmd.Flags().GetInt("course-id")
			todoDate, _ := cmd.Flags().GetString("todo-date")

			body := map[string]any{"title": title}
			if details != "" {
				body["details"] = details
			}
			if courseID > 0 {
				body["course_id"] = courseID
			}
			if todoDate != "" {
				body["todo_date"] = todoDate
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("create planner note %q", title), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/planner/notes", body)
			if err != nil {
				return err
			}

			var note PlannerNoteSummary
			if err := json.Unmarshal(data, &note); err != nil {
				return fmt.Errorf("parse created planner note: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(note)
			}
			fmt.Printf("Planner note created: %d — %s\n", note.ID, note.Title)
			return nil
		},
	}

	cmd.Flags().String("title", "", "Note title (required)")
	cmd.Flags().String("details", "", "Note details/body")
	cmd.Flags().Int("course-id", 0, "Associate note with a course ID")
	cmd.Flags().String("todo-date", "", "Due date for the note (RFC3339)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newPlannerUpdateNoteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-note",
		Short: "Update an existing planner note",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			noteID, _ := cmd.Flags().GetString("note-id")
			if noteID == "" {
				return fmt.Errorf("--note-id is required")
			}

			body := map[string]any{}
			if cmd.Flags().Changed("title") {
				v, _ := cmd.Flags().GetString("title")
				body["title"] = v
			}
			if cmd.Flags().Changed("details") {
				v, _ := cmd.Flags().GetString("details")
				body["details"] = v
			}
			if cmd.Flags().Changed("todo-date") {
				v, _ := cmd.Flags().GetString("todo-date")
				body["todo_date"] = v
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("update planner note %s", noteID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Put(ctx, "/planner/notes/"+noteID, body)
			if err != nil {
				return err
			}

			var note PlannerNoteSummary
			if err := json.Unmarshal(data, &note); err != nil {
				return fmt.Errorf("parse updated planner note: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(note)
			}
			fmt.Printf("Planner note %d updated.\n", note.ID)
			return nil
		},
	}

	cmd.Flags().String("note-id", "", "Canvas planner note ID (required)")
	cmd.Flags().String("title", "", "New note title")
	cmd.Flags().String("details", "", "New note details/body")
	cmd.Flags().String("todo-date", "", "New due date (RFC3339)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newPlannerDeleteNoteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-note",
		Short: "Delete a planner note",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			noteID, _ := cmd.Flags().GetString("note-id")
			if noteID == "" {
				return fmt.Errorf("--note-id is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the planner note"); err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("delete planner note %s", noteID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Delete(ctx, "/planner/notes/"+noteID)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"deleted": true, "note_id": noteID})
			}
			fmt.Printf("Planner note %s deleted.\n", noteID)
			return nil
		},
	}

	cmd.Flags().String("note-id", "", "Canvas planner note ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newPlannerOverridesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "overrides",
		Short: "List planner overrides for the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Get(ctx, "/planner/overrides", nil)
			if err != nil {
				return err
			}

			var overrides []map[string]any
			if err := json.Unmarshal(data, &overrides); err != nil {
				return fmt.Errorf("parse planner overrides: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(overrides)
			}

			if len(overrides) == 0 {
				fmt.Println("No planner overrides found.")
				return nil
			}
			for _, o := range overrides {
				id, _ := o["id"]
				plannableType, _ := o["plannable_type"].(string)
				plannableID, _ := o["plannable_id"]
				markedComplete, _ := o["marked_complete"].(bool)
				fmt.Printf("%-6v  %-20s  plannable_id:%-6v  complete:%v\n",
					id, plannableType, plannableID, markedComplete)
			}
			return nil
		},
	}

	return cmd
}

func newPlannerOverrideCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "override",
		Short: "Create or update a planner override to mark an item complete",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			plannableType, _ := cmd.Flags().GetString("plannable-type")
			plannableID, _ := cmd.Flags().GetString("plannable-id")
			markedComplete, _ := cmd.Flags().GetBool("marked-complete")

			if plannableType == "" {
				return fmt.Errorf("--plannable-type is required")
			}
			if plannableID == "" {
				return fmt.Errorf("--plannable-id is required")
			}

			body := map[string]any{
				"plannable_type":  plannableType,
				"plannable_id":    plannableID,
				"marked_complete": markedComplete,
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("override %s %s marked_complete=%v", plannableType, plannableID, markedComplete), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/planner/overrides", body)
			if err != nil {
				return err
			}

			var override map[string]any
			if err := json.Unmarshal(data, &override); err != nil {
				return fmt.Errorf("parse planner override: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(override)
			}

			id, _ := override["id"]
			fmt.Printf("Planner override created: %v\n", id)
			return nil
		},
	}

	cmd.Flags().String("plannable-type", "", "Type of plannable item (e.g. Assignment, DiscussionTopic) (required)")
	cmd.Flags().String("plannable-id", "", "ID of the plannable item (required)")
	cmd.Flags().Bool("marked-complete", false, "Mark the item as complete")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}
