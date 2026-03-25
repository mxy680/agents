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
