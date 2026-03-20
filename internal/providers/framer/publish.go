package framer

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newPublishCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "Publish and deployment commands",
	}
	cmd.AddCommand(
		newPublishCreateCmd(factory),
		newPublishDeployCmd(factory),
		newPublishListCmd(factory),
		newPublishInfoCmd(factory),
	)
	return cmd
}

func newPublishCreateCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Publish the project",
		RunE:  makeRunPublishCreate(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("dry-run", false, "Preview without publishing")
	return cmd
}

func makeRunPublishCreate(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		if ok, err := dryRunResult(cmd, "publish project", nil); ok {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("publish", nil)
		if err != nil {
			return fmt.Errorf("publish: %w", err)
		}

		var pubResult PublishResult
		if err := json.Unmarshal(result, &pubResult); err != nil {
			return fmt.Errorf("parse publish result: %w", err)
		}

		lines := []string{}
		if pubResult.DeploymentID != "" {
			lines = append(lines, fmt.Sprintf("Deployment ID: %s", pubResult.DeploymentID))
		}
		if pubResult.URL != "" {
			lines = append(lines, fmt.Sprintf("URL: %s", pubResult.URL))
		}
		if len(lines) == 0 {
			lines = append(lines, "Published successfully")
		}

		return cli.PrintResult(cmd, pubResult, lines)
	}
}

func newPublishDeployCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a specific deployment to domains",
		RunE:  makeRunPublishDeploy(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("dry-run", false, "Preview without deploying")
	cmd.Flags().String("deployment-id", "", "Deployment ID to deploy (required)")
	cmd.Flags().String("domains", "", "Comma-separated list of domains")
	_ = cmd.MarkFlagRequired("deployment-id")
	return cmd
}

func makeRunPublishDeploy(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		deploymentID, _ := cmd.Flags().GetString("deployment-id")
		domainsStr, _ := cmd.Flags().GetString("domains")

		params := map[string]any{
			"deploymentId": deploymentID,
		}
		if domainsStr != "" {
			params["domains"] = parseStringList(domainsStr)
		}

		if ok, err := dryRunResult(cmd, fmt.Sprintf("deploy %s", deploymentID), params); ok {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("deploy", params)
		if err != nil {
			return fmt.Errorf("deploy: %w", err)
		}

		var deployment Deployment
		if err := json.Unmarshal(result, &deployment); err != nil {
			return fmt.Errorf("parse deployment: %w", err)
		}

		lines := []string{
			fmt.Sprintf("ID:  %s", deployment.ID),
		}
		if deployment.URL != "" {
			lines = append(lines, fmt.Sprintf("URL: %s", deployment.URL))
		}

		return cli.PrintResult(cmd, deployment, lines)
	}
}

func newPublishListCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deployments",
		RunE:  makeRunPublishList(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunPublishList(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getDeployments", nil)
		if err != nil {
			return fmt.Errorf("get deployments: %w", err)
		}

		var deployments []Deployment
		if err := json.Unmarshal(result, &deployments); err != nil {
			return fmt.Errorf("parse deployments: %w", err)
		}

		lines := []string{fmt.Sprintf("Deployments (%d):", len(deployments))}
		for _, d := range deployments {
			line := fmt.Sprintf("  %s", d.ID)
			if d.URL != "" {
				line += fmt.Sprintf("  %s", d.URL)
			}
			if d.CreatedAt != "" {
				line += fmt.Sprintf("  (%s)", d.CreatedAt)
			}
			lines = append(lines, line)
		}

		return cli.PrintResult(cmd, deployments, lines)
	}
}

func newPublishInfoCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get publish info",
		RunE:  makeRunPublishInfo(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunPublishInfo(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getPublishInfo", nil)
		if err != nil {
			return fmt.Errorf("get publish info: %w", err)
		}

		var info PublishInfo
		if err := json.Unmarshal(result, &info); err != nil {
			return fmt.Errorf("parse publish info: %w", err)
		}

		lines := []string{}
		if info.URL != "" {
			lines = append(lines, fmt.Sprintf("URL:            %s", info.URL))
		}
		if info.LastPublished != "" {
			lines = append(lines, fmt.Sprintf("Last Published: %s", info.LastPublished))
		}
		if len(lines) == 0 {
			lines = append(lines, "(no publish info)")
		}

		return cli.PrintResult(cmd, info, lines)
	}
}
