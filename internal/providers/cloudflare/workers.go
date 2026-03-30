package cloudflare

import (
	"fmt"
	"net/http"
	"os"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newWorkersListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Worker scripts",
		RunE:  makeRunWorkersList(factory),
	}
	return cmd
}

func makeRunWorkersList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath("/workers/scripts")
		if err != nil {
			return err
		}

		var resp []map[string]any
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return fmt.Errorf("listing workers: %w", err)
		}

		workers := make([]WorkerSummary, 0, len(resp))
		for _, w := range resp {
			workers = append(workers, toWorkerSummary(w))
		}

		return printWorkers(cmd, workers)
	}
}

func newWorkersGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a Worker script",
		RunE:  makeRunWorkersGet(factory),
	}
	cmd.Flags().String("name", "", "Worker script name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunWorkersGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath(fmt.Sprintf("/workers/scripts/%s", name))
		if err != nil {
			return err
		}

		var data map[string]any
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("getting worker %q: %w", name, err)
		}

		worker := toWorkerSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(worker)
		}

		lines := []string{
			fmt.Sprintf("ID:          %s", worker.ID),
			fmt.Sprintf("ETAG:        %s", worker.ETAG),
			fmt.Sprintf("Created:     %s", worker.CreatedOn),
			fmt.Sprintf("Modified:    %s", worker.ModifiedOn),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newWorkersDeployCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a Worker script from a file",
		RunE:  makeRunWorkersDeploy(factory),
	}
	cmd.Flags().String("name", "", "Worker script name (required)")
	cmd.Flags().String("file", "", "Path to the JavaScript file to upload (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func makeRunWorkersDeploy(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		filePath, _ := cmd.Flags().GetString("file")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would deploy worker %q from %q", name, filePath), map[string]any{
				"action": "deploy",
				"name":   name,
				"file":   filePath,
			})
		}

		scriptData, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("reading script file %q: %w", filePath, err)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath(fmt.Sprintf("/workers/scripts/%s", name))
		if err != nil {
			return err
		}

		body := rawBody{data: scriptData, contentType: "application/javascript"}
		var data map[string]any
		if err := client.doJSON(ctx, http.MethodPut, path, body, &data); err != nil {
			return fmt.Errorf("deploying worker %q: %w", name, err)
		}

		worker := toWorkerSummary(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(worker)
		}
		fmt.Printf("Deployed worker: %s\n", worker.ID)
		return nil
	}
}

func newWorkersDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Worker script (irreversible)",
		RunE:  makeRunWorkersDelete(factory),
	}
	cmd.Flags().String("name", "", "Worker script name (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunWorkersDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete worker %q", name), map[string]any{
				"action": "delete",
				"name":   name,
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

		path, err := client.accountPath(fmt.Sprintf("/workers/scripts/%s", name))
		if err != nil {
			return err
		}

		if _, err := client.do(ctx, http.MethodDelete, path, nil); err != nil {
			return fmt.Errorf("deleting worker %q: %w", name, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "name": name})
		}
		fmt.Printf("Deleted worker: %s\n", name)
		return nil
	}
}
