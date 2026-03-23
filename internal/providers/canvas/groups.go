package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newGroupsCmd returns the parent "groups" command with all subcommands attached.
func newGroupsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "groups",
		Short:   "Manage Canvas groups",
		Aliases: []string{"group", "grp"},
	}

	cmd.AddCommand(newGroupsListCmd(factory))
	cmd.AddCommand(newGroupsGetCmd(factory))
	cmd.AddCommand(newGroupsCreateCmd(factory))
	cmd.AddCommand(newGroupsUpdateCmd(factory))
	cmd.AddCommand(newGroupsDeleteCmd(factory))
	cmd.AddCommand(newGroupsMembersCmd(factory))
	cmd.AddCommand(newGroupsCategoriesCmd(factory))

	return cmd
}

func newGroupsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List groups for the current user or a specific course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			contextType, _ := cmd.Flags().GetString("context-type")
			contextID, _ := cmd.Flags().GetString("context-id")
			limit, _ := cmd.Flags().GetInt("limit")

			params := url.Values{}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			var path string
			if contextType == "Course" && contextID != "" {
				path = "/courses/" + contextID + "/groups"
			} else {
				path = "/users/self/groups"
				if contextType != "" {
					params.Set("context_type", contextType)
				}
			}

			data, err := client.Get(ctx, path, params)
			if err != nil {
				return err
			}

			var groups []GroupSummary
			if err := json.Unmarshal(data, &groups); err != nil {
				return fmt.Errorf("parse groups: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(groups)
			}

			if len(groups) == 0 {
				fmt.Println("No groups found.")
				return nil
			}
			for _, g := range groups {
				fmt.Printf("%-6d  %-10s  %s\n", g.ID, g.JoinLevel, truncate(g.Name, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("context-type", "", "Filter by context type: Account or Course")
	cmd.Flags().String("context-id", "", "Context ID (course ID when context-type=Course)")
	cmd.Flags().Int("limit", 0, "Maximum number of groups to return")
	return cmd
}

func newGroupsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific group",
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

			data, err := client.Get(ctx, "/groups/"+groupID, nil)
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

			fmt.Printf("ID:           %d\n", group.ID)
			fmt.Printf("Name:         %s\n", group.Name)
			fmt.Printf("Join Level:   %s\n", group.JoinLevel)
			if group.ContextType != "" {
				fmt.Printf("Context:      %s\n", group.ContextType)
			}
			if group.MembersCount > 0 {
				fmt.Printf("Members:      %d\n", group.MembersCount)
			}
			if group.Description != "" {
				fmt.Printf("Description:  %s\n", truncate(group.Description, 200))
			}
			return nil
		},
	}

	cmd.Flags().String("group-id", "", "Canvas group ID (required)")
	return cmd
}

func newGroupsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new group",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			groupCategoryID, _ := cmd.Flags().GetString("group-category-id")
			if groupCategoryID == "" {
				return fmt.Errorf("--group-category-id is required")
			}

			description, _ := cmd.Flags().GetString("description")
			joinLevel, _ := cmd.Flags().GetString("join-level")

			body := map[string]any{
				"name":              name,
				"group_category_id": groupCategoryID,
			}
			if description != "" {
				body["description"] = description
			}
			if joinLevel != "" {
				body["join_level"] = joinLevel
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("create group %q", name), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/groups", body)
			if err != nil {
				return err
			}

			var group GroupSummary
			if err := json.Unmarshal(data, &group); err != nil {
				return fmt.Errorf("parse created group: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(group)
			}
			fmt.Printf("Group created: %d — %s\n", group.ID, group.Name)
			return nil
		},
	}

	cmd.Flags().String("name", "", "Group name (required)")
	cmd.Flags().String("group-category-id", "", "Group category ID (required)")
	cmd.Flags().String("description", "", "Group description")
	cmd.Flags().String("join-level", "", "Join level: parent_context_auto_join, parent_context_request, or invitation_only")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newGroupsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing group",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			groupID, _ := cmd.Flags().GetString("group-id")
			if groupID == "" {
				return fmt.Errorf("--group-id is required")
			}

			body := map[string]any{}
			if cmd.Flags().Changed("name") {
				v, _ := cmd.Flags().GetString("name")
				body["name"] = v
			}
			if cmd.Flags().Changed("description") {
				v, _ := cmd.Flags().GetString("description")
				body["description"] = v
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("update group %s", groupID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Put(ctx, "/groups/"+groupID, body)
			if err != nil {
				return err
			}

			var group GroupSummary
			if err := json.Unmarshal(data, &group); err != nil {
				return fmt.Errorf("parse updated group: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(group)
			}
			fmt.Printf("Group %d updated.\n", group.ID)
			return nil
		},
	}

	cmd.Flags().String("group-id", "", "Canvas group ID (required)")
	cmd.Flags().String("name", "", "New group name")
	cmd.Flags().String("description", "", "New group description")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newGroupsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a group",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			groupID, _ := cmd.Flags().GetString("group-id")
			if groupID == "" {
				return fmt.Errorf("--group-id is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the group"); err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("delete group %s", groupID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Delete(ctx, "/groups/"+groupID)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"deleted": true, "group_id": groupID})
			}
			fmt.Printf("Group %s deleted.\n", groupID)
			return nil
		},
	}

	cmd.Flags().String("group-id", "", "Canvas group ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

// GroupMembership is a single group membership entry.
type GroupMembership struct {
	ID      int    `json:"id"`
	GroupID int    `json:"group_id,omitempty"`
	UserID  int    `json:"user_id,omitempty"`
	State   string `json:"workflow_state,omitempty"`
}

func newGroupsMembersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members",
		Short: "List members of a group",
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

			limit, _ := cmd.Flags().GetInt("limit")
			params := url.Values{}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/groups/"+groupID+"/memberships", params)
			if err != nil {
				return err
			}

			var memberships []GroupMembership
			if err := json.Unmarshal(data, &memberships); err != nil {
				return fmt.Errorf("parse group members: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(memberships)
			}

			if len(memberships) == 0 {
				fmt.Println("No members found.")
				return nil
			}
			for _, m := range memberships {
				fmt.Printf("member_id:%-6d  state:%s\n", m.UserID, m.State)
			}
			return nil
		},
	}

	cmd.Flags().String("group-id", "", "Canvas group ID (required)")
	cmd.Flags().Int("limit", 0, "Maximum number of members to return")
	return cmd
}

// GroupCategory is a Canvas group category.
type GroupCategory struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Role            string `json:"role,omitempty"`
	SelfSignup      string `json:"self_signup,omitempty"`
	GroupsCount     int    `json:"groups_count,omitempty"`
	UnassignedCount int    `json:"unassigned_users_count,omitempty"`
}

func newGroupsCategoriesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "categories",
		Short: "List group categories for a course",
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

			data, err := client.Get(ctx, "/courses/"+courseID+"/group_categories", nil)
			if err != nil {
				return err
			}

			var categories []GroupCategory
			if err := json.Unmarshal(data, &categories); err != nil {
				return fmt.Errorf("parse group categories: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(categories)
			}

			if len(categories) == 0 {
				fmt.Println("No group categories found.")
				return nil
			}
			for _, c := range categories {
				fmt.Printf("%-6d  groups:%-4d  %s\n", c.ID, c.GroupsCount, c.Name)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	return cmd
}
