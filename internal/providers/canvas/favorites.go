package canvas

import (
	"encoding/json"
	"fmt"
	"strconv"

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

func newFavoritesAddCourseCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-course",
		Short: "Add a course to the current user's favorites",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetInt("course-id")
			if courseID == 0 {
				return fmt.Errorf("--course-id is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("add course %d to favorites", courseID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/users/self/favorites/courses/"+strconv.Itoa(courseID), nil)
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
			fmt.Printf("Course %d added to favorites.\n", courseID)
			return nil
		},
	}

	cmd.Flags().Int("course-id", 0, "Canvas course ID to favorite (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newFavoritesRemoveCourseCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-course",
		Short: "Remove a course from the current user's favorites",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetInt("course-id")
			if courseID == 0 {
				return fmt.Errorf("--course-id is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("remove course %d from favorites", courseID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Delete(ctx, "/users/self/favorites/courses/"+strconv.Itoa(courseID))
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"removed": true, "course_id": courseID})
			}
			fmt.Printf("Course %d removed from favorites.\n", courseID)
			return nil
		},
	}

	cmd.Flags().Int("course-id", 0, "Canvas course ID to unfavorite (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newFavoritesAddGroupCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-group",
		Short: "Add a group to the current user's favorites",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			groupID, _ := cmd.Flags().GetInt("group-id")
			if groupID == 0 {
				return fmt.Errorf("--group-id is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("add group %d to favorites", groupID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/users/self/favorites/groups/"+strconv.Itoa(groupID), nil)
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
			fmt.Printf("Group %d added to favorites.\n", groupID)
			return nil
		},
	}

	cmd.Flags().Int("group-id", 0, "Canvas group ID to favorite (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newFavoritesRemoveGroupCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-group",
		Short: "Remove a group from the current user's favorites",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			groupID, _ := cmd.Flags().GetInt("group-id")
			if groupID == 0 {
				return fmt.Errorf("--group-id is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("remove group %d from favorites", groupID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Delete(ctx, "/users/self/favorites/groups/"+strconv.Itoa(groupID))
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"removed": true, "group_id": groupID})
			}
			fmt.Printf("Group %d removed from favorites.\n", groupID)
			return nil
		},
	}

	cmd.Flags().Int("group-id", 0, "Canvas group ID to unfavorite (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}
