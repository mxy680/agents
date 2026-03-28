package linear

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newProjectsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE:  makeRunProjectsList(factory),
	}
	cmd.Flags().Int("limit", 25, "Maximum number of projects to return")
	return cmd
}

func makeRunProjectsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query($first: Int!) {
  projects(first: $first) {
    nodes {
      id
      name
      state
      progress
      teams { nodes { name } }
    }
  }
}`

		var resp struct {
			Projects struct {
				Nodes []struct {
					ID       string  `json:"id"`
					Name     string  `json:"name"`
					State    string  `json:"state"`
					Progress float64 `json:"progress"`
					Teams    struct {
						Nodes []struct {
							Name string `json:"name"`
						} `json:"nodes"`
					} `json:"teams"`
				} `json:"nodes"`
			} `json:"projects"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"first": limit}, &resp); err != nil {
			return fmt.Errorf("listing projects: %w", err)
		}

		summaries := make([]ProjectSummary, 0, len(resp.Projects.Nodes))
		for _, n := range resp.Projects.Nodes {
			s := ProjectSummary{
				ID:       n.ID,
				Name:     n.Name,
				State:    n.State,
				Progress: n.Progress,
			}
			for _, t := range n.Teams.Nodes {
				s.Teams = append(s.Teams, t.Name)
			}
			summaries = append(summaries, s)
		}

		return printProjectSummaries(cmd, summaries)
	}
}

func newProjectsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get project details",
		RunE:  makeRunProjectsGet(factory),
	}
	cmd.Flags().String("id", "", "Project ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunProjectsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query($id: String!) {
  project(id: $id) {
    id
    name
    description
    state
    startDate
    targetDate
    progress
  }
}`

		var resp struct {
			Project struct {
				ID          string  `json:"id"`
				Name        string  `json:"name"`
				Description string  `json:"description"`
				State       string  `json:"state"`
				StartDate   string  `json:"startDate"`
				TargetDate  string  `json:"targetDate"`
				Progress    float64 `json:"progress"`
			} `json:"project"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"id": id}, &resp); err != nil {
			return fmt.Errorf("getting project %q: %w", id, err)
		}

		n := resp.Project
		detail := ProjectDetail{
			ID:          n.ID,
			Name:        n.Name,
			Description: n.Description,
			State:       n.State,
			StartDate:   n.StartDate,
			TargetDate:  n.TargetDate,
			Progress:    n.Progress,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:          %s", detail.ID),
			fmt.Sprintf("Name:        %s", detail.Name),
			fmt.Sprintf("State:       %s", detail.State),
			fmt.Sprintf("Progress:    %.0f%%", detail.Progress*100),
			fmt.Sprintf("Start Date:  %s", detail.StartDate),
			fmt.Sprintf("Target Date: %s", detail.TargetDate),
			fmt.Sprintf("Description: %s", detail.Description),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newProjectsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		RunE:  makeRunProjectsCreate(factory),
	}
	cmd.Flags().String("name", "", "Project name (required)")
	cmd.Flags().String("team", "", "Team ID (required)")
	cmd.Flags().String("description", "", "Project description")
	cmd.Flags().Bool("dry-run", false, "Print what would be created without making changes")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("team")
	return cmd
}

func makeRunProjectsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		teamID, _ := cmd.Flags().GetString("team")
		description, _ := cmd.Flags().GetString("description")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create project %q", name), map[string]any{
				"action":      "create",
				"name":        name,
				"teamId":      teamID,
				"description": description,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
mutation($input: ProjectCreateInput!) {
  projectCreate(input: $input) {
    project {
      id
      name
    }
  }
}`

		input := map[string]any{
			"name":    name,
			"teamIds": []string{teamID},
		}
		if description != "" {
			input["description"] = description
		}

		var resp struct {
			ProjectCreate struct {
				Project struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"project"`
			} `json:"projectCreate"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"input": input}, &resp); err != nil {
			return fmt.Errorf("creating project: %w", err)
		}

		project := resp.ProjectCreate.Project
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(project)
		}
		fmt.Printf("Created project: %s (ID: %s)\n", project.Name, project.ID)
		return nil
	}
}

func newProjectsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a project",
		RunE:  makeRunProjectsUpdate(factory),
	}
	cmd.Flags().String("id", "", "Project ID (required)")
	cmd.Flags().String("name", "", "New project name")
	cmd.Flags().String("description", "", "New description")
	cmd.Flags().String("state", "", "New state")
	cmd.Flags().Bool("dry-run", false, "Print what would be updated without making changes")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunProjectsUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		state, _ := cmd.Flags().GetString("state")

		input := map[string]any{}
		if name != "" {
			input["name"] = name
		}
		if description != "" {
			input["description"] = description
		}
		if state != "" {
			input["state"] = state
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update project %q", id), input)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
mutation($id: String!, $input: ProjectUpdateInput!) {
  projectUpdate(id: $id, input: $input) {
    project {
      id
      name
    }
  }
}`

		var resp struct {
			ProjectUpdate struct {
				Project struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"project"`
			} `json:"projectUpdate"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"id": id, "input": input}, &resp); err != nil {
			return fmt.Errorf("updating project %q: %w", id, err)
		}

		project := resp.ProjectUpdate.Project
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(project)
		}
		fmt.Printf("Updated project: %s (ID: %s)\n", project.Name, project.ID)
		return nil
	}
}

func newProjectsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a project (irreversible)",
		RunE:  makeRunProjectsDelete(factory),
	}
	cmd.Flags().String("id", "", "Project ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	cmd.Flags().Bool("dry-run", false, "Print what would be deleted without making changes")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunProjectsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete project %q", id), map[string]any{
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
  projectDelete(id: $id) {
    success
  }
}`

		var resp struct {
			ProjectDelete struct {
				Success bool `json:"success"`
			} `json:"projectDelete"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"id": id}, &resp); err != nil {
			return fmt.Errorf("deleting project %q: %w", id, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"success": resp.ProjectDelete.Success, "id": id})
		}
		fmt.Printf("Deleted project: %s\n", id)
		return nil
	}
}
