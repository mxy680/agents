package github

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func toLabelInfo(data map[string]any) LabelInfo {
	return LabelInfo{
		ID:          jsonInt64(data["id"]),
		Name:        jsonString(data["name"]),
		Color:       jsonString(data["color"]),
		Description: jsonString(data["description"]),
	}
}

func newLabelsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List labels in a repository",
		RunE:  makeRunLabelsList(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().Int("limit", 20, "Maximum number of labels to return")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunLabelsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/labels?per_page=%s", owner, repo, strconv.Itoa(limit))

		var raw []map[string]any
		if _, err := doGitHub(client, "GET", path, nil, &raw); err != nil {
			return fmt.Errorf("listing labels for %s/%s: %w", owner, repo, err)
		}

		labels := make([]LabelInfo, 0, len(raw))
		for _, item := range raw {
			labels = append(labels, toLabelInfo(item))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(labels)
		}

		if len(labels) == 0 {
			fmt.Println("No labels found.")
			return nil
		}

		lines := make([]string, 0, len(labels)+1)
		lines = append(lines, fmt.Sprintf("%-12s  %-30s  %-8s  %s", "ID", "NAME", "COLOR", "DESCRIPTION"))
		for _, l := range labels {
			lines = append(lines, fmt.Sprintf("%-12d  %-30s  %-8s  %s", l.ID, truncate(l.Name, 30), l.Color, l.Description))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newLabelsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a label by name",
		RunE:  makeRunLabelsGet(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("name", "", "Label name (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunLabelsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		name, _ := cmd.Flags().GetString("name")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/labels/%s", owner, repo, url.PathEscape(name))

		var raw map[string]any
		if _, err := doGitHub(client, "GET", path, nil, &raw); err != nil {
			return fmt.Errorf("getting label %q in %s/%s: %w", name, owner, repo, err)
		}

		label := toLabelInfo(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(label)
		}

		lines := []string{
			fmt.Sprintf("ID:           %d", label.ID),
			fmt.Sprintf("Name:         %s", label.Name),
			fmt.Sprintf("Color:        %s", label.Color),
			fmt.Sprintf("Description:  %s", label.Description),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newLabelsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a label in a repository",
		RunE:  makeRunLabelsCreate(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("name", "", "Label name (required)")
	cmd.Flags().String("color", "", "Hex color without # (e.g. ff0000)")
	cmd.Flags().String("description", "", "Label description")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunLabelsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		name, _ := cmd.Flags().GetString("name")
		color, _ := cmd.Flags().GetString("color")
		description, _ := cmd.Flags().GetString("description")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		reqBody := map[string]any{
			"name":        name,
			"color":       color,
			"description": description,
		}

		path := fmt.Sprintf("/repos/%s/%s/labels", owner, repo)

		var raw map[string]any
		if _, err := doGitHub(client, "POST", path, reqBody, &raw); err != nil {
			return fmt.Errorf("creating label %q in %s/%s: %w", name, owner, repo, err)
		}

		label := toLabelInfo(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(label)
		}

		lines := []string{
			fmt.Sprintf("ID:           %d", label.ID),
			fmt.Sprintf("Name:         %s", label.Name),
			fmt.Sprintf("Color:        %s", label.Color),
			fmt.Sprintf("Description:  %s", label.Description),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newLabelsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing label",
		RunE:  makeRunLabelsUpdate(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("name", "", "Current label name (required)")
	cmd.Flags().String("new-name", "", "New label name")
	cmd.Flags().String("color", "", "New hex color without #")
	cmd.Flags().String("description", "", "New label description")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunLabelsUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		name, _ := cmd.Flags().GetString("name")

		reqBody := map[string]any{}
		if cmd.Flags().Changed("new-name") {
			newName, _ := cmd.Flags().GetString("new-name")
			reqBody["new_name"] = newName
		}
		if cmd.Flags().Changed("color") {
			color, _ := cmd.Flags().GetString("color")
			reqBody["color"] = color
		}
		if cmd.Flags().Changed("description") {
			description, _ := cmd.Flags().GetString("description")
			reqBody["description"] = description
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/labels/%s", owner, repo, url.PathEscape(name))

		var raw map[string]any
		if _, err := doGitHub(client, "PATCH", path, reqBody, &raw); err != nil {
			return fmt.Errorf("updating label %q in %s/%s: %w", name, owner, repo, err)
		}

		label := toLabelInfo(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(label)
		}

		lines := []string{
			fmt.Sprintf("ID:           %d", label.ID),
			fmt.Sprintf("Name:         %s", label.Name),
			fmt.Sprintf("Color:        %s", label.Color),
			fmt.Sprintf("Description:  %s", label.Description),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newLabelsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a label from a repository",
		RunE:  makeRunLabelsDelete(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("name", "", "Label name (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunLabelsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		name, _ := cmd.Flags().GetString("name")

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/labels/%s", owner, repo, url.PathEscape(name))

		if _, err := doGitHub(client, "DELETE", path, nil, nil); err != nil {
			return fmt.Errorf("deleting label %q in %s/%s: %w", name, owner, repo, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "name": name})
		}
		fmt.Printf("Deleted label: %s\n", name)
		return nil
	}
}
