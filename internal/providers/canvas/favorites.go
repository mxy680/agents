package canvas

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newFavoritesCmd returns the parent "favorites" command with all subcommands attached.
func newFavoritesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "favorites",
		Short:   "Manage Canvas favorite courses and groups",
		Aliases: []string{"fav"},
	}

	cmd.AddCommand(newFavoritesCoursesCmd(factory))
	cmd.AddCommand(newFavoritesGroupsCmd(factory))

	return cmd
}

func newFavoritesCoursesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "courses",
		Short: "List favorite courses for the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Get(ctx, "/users/self/favorites/courses", nil)
			if err != nil {
				return err
			}

			var courses []CourseSummary
			if err := json.Unmarshal(data, &courses); err != nil {
				return fmt.Errorf("parse favorite courses: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(courses)
			}

			if len(courses) == 0 {
				fmt.Println("No favorite courses found.")
				return nil
			}
			for _, c := range courses {
				fmt.Printf("%-6d  %-12s  %s\n", c.ID, c.CourseCode, truncate(c.Name, 60))
			}
			return nil
		},
	}

	return cmd
}

func newFavoritesGroupsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "List favorite groups for the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Get(ctx, "/users/self/favorites/groups", nil)
			if err != nil {
				return err
			}

			var groups []GroupSummary
			if err := json.Unmarshal(data, &groups); err != nil {
				return fmt.Errorf("parse favorite groups: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(groups)
			}

			if len(groups) == 0 {
				fmt.Println("No favorite groups found.")
				return nil
			}
			for _, g := range groups {
				fmt.Printf("%-6d  %-10s  %s\n", g.ID, g.JoinLevel, truncate(g.Name, 60))
			}
			return nil
		},
	}

	return cmd
}

