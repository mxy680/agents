package canvas

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newAssignmentGroupsCmd returns the parent "assignment-groups" command with all subcommands attached.
func newAssignmentGroupsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "assignment-groups",
		Short:   "Manage Canvas assignment groups",
		Aliases: []string{"assign-group", "ag"},
	}

	cmd.AddCommand(newAssignmentGroupsListCmd(factory))
	cmd.AddCommand(newAssignmentGroupsGetCmd(factory))
	cmd.AddCommand(newAssignmentGroupsCreateCmd(factory))
	cmd.AddCommand(newAssignmentGroupsUpdateCmd(factory))
	cmd.AddCommand(newAssignmentGroupsDeleteCmd(factory))

	return cmd
}

func newAssignmentGroupsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List assignment groups for a course",
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

			data, err := client.Get(ctx, "/courses/"+courseID+"/assignment_groups", nil)
			if err != nil {
				return err
			}

			var groups []AssignmentGroupSummary
			if err := json.Unmarshal(data, &groups); err != nil {
				return fmt.Errorf("parse assignment groups: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(groups)
			}

			if len(groups) == 0 {
				fmt.Println("No assignment groups found.")
				return nil
			}
			for _, g := range groups {
				fmt.Printf("%-6d  pos=%-3d  weight=%-5.1f  %s\n", g.ID, g.Position, g.GroupWeight, truncate(g.Name, 50))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	return cmd
}

func newAssignmentGroupsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific assignment group",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			groupID, _ := cmd.Flags().GetString("group-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if groupID == "" {
				return fmt.Errorf("--group-id is required")
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/assignment_groups/"+groupID, nil)
			if err != nil {
				return err
			}

			var group AssignmentGroupSummary
			if err := json.Unmarshal(data, &group); err != nil {
				return fmt.Errorf("parse assignment group: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(group)
			}

			fmt.Printf("ID:           %d\n", group.ID)
			fmt.Printf("Name:         %s\n", group.Name)
			fmt.Printf("Position:     %d\n", group.Position)
			fmt.Printf("Group Weight: %.2f\n", group.GroupWeight)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("group-id", "", "Canvas assignment group ID (required)")
	return cmd
}

func newAssignmentGroupsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new assignment group",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			body := map[string]any{"name": name}
			if cmd.Flags().Changed("position") {
				v, _ := cmd.Flags().GetInt("position")
				body["position"] = v
			}
			if cmd.Flags().Changed("weight") {
				v, _ := cmd.Flags().GetFloat64("weight")
				body["group_weight"] = v
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("create assignment group %q in course %s", name, courseID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/courses/"+courseID+"/assignment_groups", body)
			if err != nil {
				return err
			}

			var group AssignmentGroupSummary
			if err := json.Unmarshal(data, &group); err != nil {
				return fmt.Errorf("parse created assignment group: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(group)
			}
			fmt.Printf("Assignment group created: %d — %s\n", group.ID, group.Name)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("name", "", "Assignment group name (required)")
	cmd.Flags().Int("position", 0, "Display position of the group")
	cmd.Flags().Float64("weight", 0, "Weight of the group for grade calculations")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newAssignmentGroupsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing assignment group",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			groupID, _ := cmd.Flags().GetString("group-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if groupID == "" {
				return fmt.Errorf("--group-id is required")
			}

			body := map[string]any{}
			if cmd.Flags().Changed("name") {
				v, _ := cmd.Flags().GetString("name")
				body["name"] = v
			}
			if cmd.Flags().Changed("position") {
				v, _ := cmd.Flags().GetInt("position")
				body["position"] = v
			}
			if cmd.Flags().Changed("weight") {
				v, _ := cmd.Flags().GetFloat64("weight")
				body["group_weight"] = v
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("update assignment group %s in course %s", groupID, courseID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Put(ctx, "/courses/"+courseID+"/assignment_groups/"+groupID, body)
			if err != nil {
				return err
			}

			var group AssignmentGroupSummary
			if err := json.Unmarshal(data, &group); err != nil {
				return fmt.Errorf("parse updated assignment group: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(group)
			}
			fmt.Printf("Assignment group %d updated.\n", group.ID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("group-id", "", "Canvas assignment group ID (required)")
	cmd.Flags().String("name", "", "New name for the group")
	cmd.Flags().Int("position", 0, "New display position")
	cmd.Flags().Float64("weight", 0, "New group weight for grade calculations")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newAssignmentGroupsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an assignment group",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			groupID, _ := cmd.Flags().GetString("group-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if groupID == "" {
				return fmt.Errorf("--group-id is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the assignment group"); err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("delete assignment group %s in course %s", groupID, courseID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Delete(ctx, "/courses/"+courseID+"/assignment_groups/"+groupID)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"deleted": true, "group_id": groupID})
			}
			fmt.Printf("Assignment group %s deleted.\n", groupID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("group-id", "", "Canvas assignment group ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}
