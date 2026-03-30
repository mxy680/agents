package linear

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newIssuesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues",
		RunE:  makeRunIssuesList(factory),
	}
	cmd.Flags().String("team", "", "Team ID to filter issues (required)")
	cmd.Flags().Int("limit", 25, "Maximum number of issues to return")
	_ = cmd.MarkFlagRequired("team")
	return cmd
}

func makeRunIssuesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		teamID, _ := cmd.Flags().GetString("team")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query($teamId: ID!, $first: Int!) {
  issues(filter: {team: {id: {eq: $teamId}}}, first: $first) {
    nodes {
      id
      identifier
      title
      state { name }
      priority
      assignee { name }
      createdAt
    }
  }
}`

		var resp struct {
			Issues struct {
				Nodes []struct {
					ID         string `json:"id"`
					Identifier string `json:"identifier"`
					Title      string `json:"title"`
					State      *struct {
						Name string `json:"name"`
					} `json:"state"`
					Priority int `json:"priority"`
					Assignee *struct {
						Name string `json:"name"`
					} `json:"assignee"`
					CreatedAt string `json:"createdAt"`
				} `json:"nodes"`
			} `json:"issues"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"teamId": teamID, "first": limit}, &resp); err != nil {
			return fmt.Errorf("listing issues: %w", err)
		}

		summaries := make([]IssueSummary, 0, len(resp.Issues.Nodes))
		for _, n := range resp.Issues.Nodes {
			s := IssueSummary{
				ID:         n.ID,
				Identifier: n.Identifier,
				Title:      n.Title,
				Priority:   n.Priority,
				CreatedAt:  n.CreatedAt,
			}
			if n.State != nil {
				s.State = n.State.Name
			}
			if n.Assignee != nil {
				s.Assignee = n.Assignee.Name
			}
			summaries = append(summaries, s)
		}

		return printIssueSummaries(cmd, summaries)
	}
}

func newIssuesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get issue details",
		RunE:  makeRunIssuesGet(factory),
	}
	cmd.Flags().String("id", "", "Issue ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunIssuesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query($id: String!) {
  issue(id: $id) {
    id
    identifier
    title
    description
    state { name }
    priority
    assignee { name }
    team { name }
    labels { nodes { name } }
    createdAt
    updatedAt
  }
}`

		var resp struct {
			Issue struct {
				ID          string `json:"id"`
				Identifier  string `json:"identifier"`
				Title       string `json:"title"`
				Description string `json:"description"`
				State       *struct {
					Name string `json:"name"`
				} `json:"state"`
				Priority int `json:"priority"`
				Assignee *struct {
					Name string `json:"name"`
				} `json:"assignee"`
				Team *struct {
					Name string `json:"name"`
				} `json:"team"`
				Labels struct {
					Nodes []struct {
						Name string `json:"name"`
					} `json:"nodes"`
				} `json:"labels"`
				CreatedAt string `json:"createdAt"`
				UpdatedAt string `json:"updatedAt"`
			} `json:"issue"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"id": id}, &resp); err != nil {
			return fmt.Errorf("getting issue %q: %w", id, err)
		}

		n := resp.Issue
		detail := IssueDetail{
			ID:          n.ID,
			Identifier:  n.Identifier,
			Title:       n.Title,
			Description: n.Description,
			Priority:    n.Priority,
			CreatedAt:   n.CreatedAt,
			UpdatedAt:   n.UpdatedAt,
		}
		if n.State != nil {
			detail.State = n.State.Name
		}
		if n.Assignee != nil {
			detail.Assignee = n.Assignee.Name
		}
		if n.Team != nil {
			detail.Team = n.Team.Name
		}
		for _, lbl := range n.Labels.Nodes {
			detail.Labels = append(detail.Labels, lbl.Name)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:          %s", detail.ID),
			fmt.Sprintf("Identifier:  %s", detail.Identifier),
			fmt.Sprintf("Title:       %s", detail.Title),
			fmt.Sprintf("State:       %s", detail.State),
			fmt.Sprintf("Priority:    %s", priorityLabel(detail.Priority)),
			fmt.Sprintf("Assignee:    %s", detail.Assignee),
			fmt.Sprintf("Team:        %s", detail.Team),
			fmt.Sprintf("Labels:      %v", detail.Labels),
			fmt.Sprintf("Created:     %s", detail.CreatedAt),
			fmt.Sprintf("Updated:     %s", detail.UpdatedAt),
			fmt.Sprintf("Description: %s", detail.Description),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newIssuesCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue",
		RunE:  makeRunIssuesCreate(factory),
	}
	cmd.Flags().String("team", "", "Team ID (required)")
	cmd.Flags().String("title", "", "Issue title (required)")
	cmd.Flags().String("description", "", "Issue description")
	cmd.Flags().Int("priority", 0, "Priority: 0=none, 1=urgent, 2=high, 3=medium, 4=low")
	cmd.Flags().Bool("dry-run", false, "Print what would be created without making changes")
	_ = cmd.MarkFlagRequired("team")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func makeRunIssuesCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		teamID, _ := cmd.Flags().GetString("team")
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		priority, _ := cmd.Flags().GetInt("priority")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create issue %q in team %s", title, teamID), map[string]any{
				"action":      "create",
				"teamId":      teamID,
				"title":       title,
				"description": description,
				"priority":    priority,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
mutation($input: IssueCreateInput!) {
  issueCreate(input: $input) {
    issue {
      id
      identifier
      title
    }
  }
}`

		input := map[string]any{
			"title":  title,
			"teamId": teamID,
		}
		if description != "" {
			input["description"] = description
		}
		if priority != 0 {
			input["priority"] = priority
		}

		var resp struct {
			IssueCreate struct {
				Issue struct {
					ID         string `json:"id"`
					Identifier string `json:"identifier"`
					Title      string `json:"title"`
				} `json:"issue"`
			} `json:"issueCreate"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"input": input}, &resp); err != nil {
			return fmt.Errorf("creating issue: %w", err)
		}

		issue := resp.IssueCreate.Issue
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(issue)
		}
		fmt.Printf("Created issue: %s — %s (ID: %s)\n", issue.Identifier, issue.Title, issue.ID)
		return nil
	}
}

func newIssuesUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an issue",
		RunE:  makeRunIssuesUpdate(factory),
	}
	cmd.Flags().String("id", "", "Issue ID (required)")
	cmd.Flags().String("title", "", "New title")
	cmd.Flags().String("state", "", "State ID")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().Int("priority", -1, "Priority: 0=none, 1=urgent, 2=high, 3=medium, 4=low")
	cmd.Flags().Bool("dry-run", false, "Print what would be updated without making changes")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunIssuesUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")
		title, _ := cmd.Flags().GetString("title")
		stateID, _ := cmd.Flags().GetString("state")
		description, _ := cmd.Flags().GetString("description")
		priority, _ := cmd.Flags().GetInt("priority")

		input := map[string]any{}
		if title != "" {
			input["title"] = title
		}
		if stateID != "" {
			input["stateId"] = stateID
		}
		if description != "" {
			input["description"] = description
		}
		if priority >= 0 {
			input["priority"] = priority
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update issue %q", id), input)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
mutation($id: String!, $input: IssueUpdateInput!) {
  issueUpdate(id: $id, input: $input) {
    issue {
      id
      identifier
      title
      state { name }
    }
  }
}`

		var resp struct {
			IssueUpdate struct {
				Issue struct {
					ID         string `json:"id"`
					Identifier string `json:"identifier"`
					Title      string `json:"title"`
					State      *struct {
						Name string `json:"name"`
					} `json:"state"`
				} `json:"issue"`
			} `json:"issueUpdate"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"id": id, "input": input}, &resp); err != nil {
			return fmt.Errorf("updating issue %q: %w", id, err)
		}

		issue := resp.IssueUpdate.Issue
		if cli.IsJSONOutput(cmd) {
			stateName := ""
			if issue.State != nil {
				stateName = issue.State.Name
			}
			return cli.PrintJSON(map[string]any{
				"id":         issue.ID,
				"identifier": issue.Identifier,
				"title":      issue.Title,
				"state":      stateName,
			})
		}
		fmt.Printf("Updated issue: %s — %s\n", issue.Identifier, issue.Title)
		return nil
	}
}

func newIssuesDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an issue (irreversible)",
		RunE:  makeRunIssuesDelete(factory),
	}
	cmd.Flags().String("id", "", "Issue ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	cmd.Flags().Bool("dry-run", false, "Print what would be deleted without making changes")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunIssuesDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete issue %q", id), map[string]any{
				"action": "delete",
				"id":     id,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
mutation($id: String!) {
  issueDelete(id: $id) {
    success
  }
}`

		var resp struct {
			IssueDelete struct {
				Success bool `json:"success"`
			} `json:"issueDelete"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"id": id}, &resp); err != nil {
			return fmt.Errorf("deleting issue %q: %w", id, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"success": resp.IssueDelete.Success, "id": id})
		}
		fmt.Printf("Deleted issue: %s\n", id)
		return nil
	}
}
