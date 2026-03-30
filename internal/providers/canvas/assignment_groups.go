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

func newAssignmentGroupsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new assignment group in a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			name, _ := cmd.Flags().GetString("name")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			body := map[string]any{"name": name}
			if weight, _ := cmd.Flags().GetFloat64("weight"); weight > 0 {
				body["group_weight"] = weight
			}
			if position, _ := cmd.Flags().GetInt("position"); position > 0 {
				body["position"] = position
			}

			data, err := client.Post(ctx, "/courses/"+courseID+"/assignment_groups", body)
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
			fmt.Printf("Assignment group %d created: %s\n", group.ID, group.Name)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("name", "", "Assignment group name (required)")
	cmd.Flags().Float64("weight", 0, "Group weight percentage")
	cmd.Flags().Int("position", 0, "Position in list")
	return cmd
}

func newAssignmentGroupsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing assignment group",
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

			body := map[string]any{}
			if name, _ := cmd.Flags().GetString("name"); name != "" {
				body["name"] = name
			}
			if weight, _ := cmd.Flags().GetFloat64("weight"); weight > 0 {
				body["group_weight"] = weight
			}

			data, err := client.Put(ctx, "/courses/"+courseID+"/assignment_groups/"+groupID, body)
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
			fmt.Printf("Assignment group %s updated\n", groupID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("group-id", "", "Canvas assignment group ID (required)")
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().Float64("weight", 0, "New group weight")
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

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			if _, err := client.Delete(ctx, "/courses/"+courseID+"/assignment_groups/"+groupID); err != nil {
				return err
			}

			fmt.Printf("Assignment group %s deleted\n", groupID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("group-id", "", "Canvas assignment group ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive action")
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
