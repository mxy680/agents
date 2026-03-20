package framer

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newProjectCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Project information commands",
	}
	cmd.AddCommand(
		newProjectInfoCmd(factory),
		newProjectUserCmd(factory),
		newProjectChangedPathsCmd(factory),
		newProjectContributorsCmd(factory),
	)
	return cmd
}

func newProjectInfoCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get project info",
		RunE:  makeRunProjectInfo(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunProjectInfo(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getProjectInfo", nil)
		if err != nil {
			return fmt.Errorf("get project info: %w", err)
		}

		var info ProjectInfo
		if err := json.Unmarshal(result, &info); err != nil {
			return fmt.Errorf("parse project info: %w", err)
		}

		return cli.PrintResult(cmd, info, []string{
			fmt.Sprintf("Name: %s", info.Name),
			fmt.Sprintf("ID:   %s", info.ID),
		})
	}
}

func newProjectUserCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Get current user",
		RunE:  makeRunProjectUser(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunProjectUser(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getCurrentUser", nil)
		if err != nil {
			return fmt.Errorf("get current user: %w", err)
		}

		var user User
		if err := json.Unmarshal(result, &user); err != nil {
			return fmt.Errorf("parse user: %w", err)
		}

		return cli.PrintResult(cmd, user, []string{
			fmt.Sprintf("ID:   %s", user.ID),
			fmt.Sprintf("Name: %s", user.Name),
		})
	}
}

func newProjectChangedPathsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "changed-paths",
		Short: "Get changed paths in the project",
		RunE:  makeRunProjectChangedPaths(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunProjectChangedPaths(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getChangedPaths", nil)
		if err != nil {
			return fmt.Errorf("get changed paths: %w", err)
		}

		var paths ChangedPaths
		if err := json.Unmarshal(result, &paths); err != nil {
			return fmt.Errorf("parse changed paths: %w", err)
		}

		lines := []string{
			fmt.Sprintf("Added (%d):", len(paths.Added)),
		}
		for _, p := range paths.Added {
			lines = append(lines, "  + "+p)
		}
		lines = append(lines, fmt.Sprintf("Removed (%d):", len(paths.Removed)))
		for _, p := range paths.Removed {
			lines = append(lines, "  - "+p)
		}
		lines = append(lines, fmt.Sprintf("Modified (%d):", len(paths.Modified)))
		for _, p := range paths.Modified {
			lines = append(lines, "  ~ "+p)
		}

		return cli.PrintResult(cmd, paths, lines)
	}
}

func newProjectContributorsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contributors",
		Short: "Get change contributors",
		RunE:  makeRunProjectContributors(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().String("from-version", "", "Start version for contributor range")
	cmd.Flags().String("to-version", "", "End version for contributor range")
	return cmd
}

func makeRunProjectContributors(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		params := map[string]any{}
		if v, _ := cmd.Flags().GetString("from-version"); v != "" {
			params["fromVersion"] = v
		}
		if v, _ := cmd.Flags().GetString("to-version"); v != "" {
			params["toVersion"] = v
		}
		if len(params) == 0 {
			params = nil
		}

		result, err := client.Call("getChangeContributors", params)
		if err != nil {
			return fmt.Errorf("get contributors: %w", err)
		}

		var contributors []string
		if err := json.Unmarshal(result, &contributors); err != nil {
			return fmt.Errorf("parse contributors: %w", err)
		}

		lines := make([]string, 0, len(contributors)+1)
		lines = append(lines, fmt.Sprintf("Contributors (%d):", len(contributors)))
		for _, c := range contributors {
			lines = append(lines, "  "+c)
		}

		return cli.PrintResult(cmd, contributors, lines)
	}
}
