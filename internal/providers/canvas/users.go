package canvas

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newUsersCmd returns the parent "users" command with all subcommands attached.
func newUsersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "users",
		Short:   "Manage Canvas users",
		Aliases: []string{"user"},
	}

	cmd.AddCommand(newUsersMeCmd(factory))
	cmd.AddCommand(newUsersTodoCmd(factory))
	cmd.AddCommand(newUsersUpcomingCmd(factory))
	cmd.AddCommand(newUsersMissingCmd(factory))

	return cmd
}

func newUsersMeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Get the current user's profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Get(ctx, "/users/self", nil)
			if err != nil {
				return err
			}

			var user UserSummary
			if err := json.Unmarshal(data, &user); err != nil {
				return fmt.Errorf("parse user: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(user)
			}

			fmt.Printf("ID:        %d\n", user.ID)
			fmt.Printf("Name:      %s\n", user.Name)
			if user.Email != "" {
				fmt.Printf("Email:     %s\n", user.Email)
			}
			if user.LoginID != "" {
				fmt.Printf("Login:     %s\n", user.LoginID)
			}
			return nil
		},
	}

	return cmd
}

func newUsersTodoCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "todo",
		Short: "List the current user's to-do items",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Get(ctx, "/users/self/todo", nil)
			if err != nil {
				return err
			}

			var items []map[string]any
			if err := json.Unmarshal(data, &items); err != nil {
				return fmt.Errorf("parse todo items: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(items)
			}

			if len(items) == 0 {
				fmt.Println("No to-do items found.")
				return nil
			}
			for _, item := range items {
				itemType, _ := item["type"].(string)
				contextName, _ := item["context_name"].(string)
				fmt.Printf("type:%-12s  course:%s\n", itemType, contextName)
			}
			return nil
		},
	}

	return cmd
}

func newUsersUpcomingCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upcoming",
		Short: "List upcoming events for the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Get(ctx, "/users/self/upcoming_events", nil)
			if err != nil {
				return err
			}

			var events []CalendarEventSummary
			if err := json.Unmarshal(data, &events); err != nil {
				return fmt.Errorf("parse upcoming events: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(events)
			}

			if len(events) == 0 {
				fmt.Println("No upcoming events found.")
				return nil
			}
			for _, e := range events {
				fmt.Printf("%-6d  %-25s  %s\n", e.ID, e.StartAt, e.Title)
			}
			return nil
		},
	}

	return cmd
}

func newUsersMissingCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "missing",
		Short: "List missing submissions for the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Get(ctx, "/users/self/missing_submissions", nil)
			if err != nil {
				return err
			}

			var assignments []AssignmentSummary
			if err := json.Unmarshal(data, &assignments); err != nil {
				return fmt.Errorf("parse missing submissions: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(assignments)
			}

			if len(assignments) == 0 {
				fmt.Println("No missing submissions found.")
				return nil
			}
			for _, a := range assignments {
				fmt.Printf("course:%-6d  due:%-25s  %s\n", a.CourseID, a.DueAt, a.Name)
			}
			return nil
		},
	}

	return cmd
}
