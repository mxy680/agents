package linear

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newUsersListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all users",
		RunE:  makeRunUsersList(factory),
	}
	return cmd
}

func makeRunUsersList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query {
  users {
    nodes {
      id
      name
      email
      displayName
      active
    }
  }
}`

		var resp struct {
			Users struct {
				Nodes []struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					Email       string `json:"email"`
					DisplayName string `json:"displayName"`
					Active      bool   `json:"active"`
				} `json:"nodes"`
			} `json:"users"`
		}

		if err := client.graphQL(ctx, q, nil, &resp); err != nil {
			return fmt.Errorf("listing users: %w", err)
		}

		users := make([]UserSummary, 0, len(resp.Users.Nodes))
		for _, n := range resp.Users.Nodes {
			users = append(users, UserSummary{
				ID:          n.ID,
				Name:        n.Name,
				Email:       n.Email,
				DisplayName: n.DisplayName,
				Active:      n.Active,
			})
		}

		return printUserSummaries(cmd, users)
	}
}

func newUsersGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get user details",
		RunE:  makeRunUsersGet(factory),
	}
	cmd.Flags().String("id", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunUsersGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query($id: String!) {
  user(id: $id) {
    id
    name
    email
    displayName
    active
  }
}`

		var resp struct {
			User struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Email       string `json:"email"`
				DisplayName string `json:"displayName"`
				Active      bool   `json:"active"`
			} `json:"user"`
		}

		if err := client.graphQL(ctx, q, map[string]any{"id": id}, &resp); err != nil {
			return fmt.Errorf("getting user %q: %w", id, err)
		}

		u := UserSummary{
			ID:          resp.User.ID,
			Name:        resp.User.Name,
			Email:       resp.User.Email,
			DisplayName: resp.User.DisplayName,
			Active:      resp.User.Active,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(u)
		}

		lines := []string{
			fmt.Sprintf("ID:           %s", u.ID),
			fmt.Sprintf("Name:         %s", u.Name),
			fmt.Sprintf("Display Name: %s", u.DisplayName),
			fmt.Sprintf("Email:        %s", u.Email),
			fmt.Sprintf("Active:       %v", u.Active),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newUsersMeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "me",
		Short: "Get the authenticated user (viewer)",
		RunE:  makeRunUsersMe(factory),
	}
	return cmd
}

func makeRunUsersMe(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		const q = `
query {
  viewer {
    id
    name
    email
    displayName
  }
}`

		var resp struct {
			Viewer struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Email       string `json:"email"`
				DisplayName string `json:"displayName"`
			} `json:"viewer"`
		}

		if err := client.graphQL(ctx, q, nil, &resp); err != nil {
			return fmt.Errorf("getting viewer: %w", err)
		}

		u := UserSummary{
			ID:          resp.Viewer.ID,
			Name:        resp.Viewer.Name,
			Email:       resp.Viewer.Email,
			DisplayName: resp.Viewer.DisplayName,
			Active:      true,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(u)
		}

		lines := []string{
			fmt.Sprintf("ID:           %s", u.ID),
			fmt.Sprintf("Name:         %s", u.Name),
			fmt.Sprintf("Display Name: %s", u.DisplayName),
			fmt.Sprintf("Email:        %s", u.Email),
		}
		cli.PrintText(lines)
		return nil
	}
}
