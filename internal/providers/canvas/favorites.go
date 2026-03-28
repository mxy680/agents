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
	cmd.AddCommand(newFavoritesAddCourseCmd(factory))
	cmd.AddCommand(newFavoritesRemoveCourseCmd(factory))
	cmd.AddCommand(newFavoritesAddGroupCmd(factory))
	cmd.AddCommand(newFavoritesRemoveGroupCmd(factory))

	return cmd
}

func newFavoritesAddCourseCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-course",
		Short: "Add a course to favorites",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, "add course "+courseID+" to favorites", map[string]any{"course_id": courseID})
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/users/self/favorites/courses/"+courseID, nil)
			if err != nil {
				return err
			}

			var course CourseSummary
			if err := json.Unmarshal(data, &course); err != nil {
				return fmt.Errorf("parse course: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(course)
			}
			fmt.Printf("Course %s added to favorites\n", courseID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	return cmd
}

func newFavoritesRemoveCourseCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-course",
		Short: "Remove a course from favorites",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}

			if _, err := client.Delete(ctx, "/users/self/favorites/courses/"+courseID); err != nil {
				return err
			}

			result := map[string]any{"course_id": courseID, "removed": true}
			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(result)
			}
			fmt.Printf("Course %s removed from favorites\n", courseID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	return cmd
}

func newFavoritesAddGroupCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-group",
		Short: "Add a group to favorites",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			groupID, _ := cmd.Flags().GetString("group-id")
			if groupID == "" {
				return fmt.Errorf("--group-id is required")
			}

			data, err := client.Post(ctx, "/users/self/favorites/groups/"+groupID, nil)
			if err != nil {
				return err
			}

			var group GroupSummary
			if err := json.Unmarshal(data, &group); err != nil {
				return fmt.Errorf("parse group: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(group)
			}
			fmt.Printf("Group %s added to favorites\n", groupID)
			return nil
		},
	}

	cmd.Flags().String("group-id", "", "Canvas group ID (required)")
	return cmd
}

func newFavoritesRemoveGroupCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-group",
		Short: "Remove a group from favorites",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			groupID, _ := cmd.Flags().GetString("group-id")
			if groupID == "" {
				return fmt.Errorf("--group-id is required")
			}

			if _, err := client.Delete(ctx, "/users/self/favorites/groups/"+groupID); err != nil {
				return err
			}

			result := map[string]any{"group_id": groupID, "removed": true}
			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(result)
			}
			fmt.Printf("Group %s removed from favorites\n", groupID)
			return nil
		},
	}

	cmd.Flags().String("group-id", "", "Canvas group ID (required)")
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

