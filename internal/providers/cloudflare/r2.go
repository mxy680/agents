package cloudflare

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newR2ListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List R2 buckets",
		RunE:  makeRunR2List(factory),
	}
	return cmd
}

func makeRunR2List(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath("/r2/buckets")
		if err != nil {
			return err
		}

		// R2 list response wraps buckets under a "buckets" key
		var resp struct {
			Buckets []map[string]any `json:"buckets"`
		}
		if err := client.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
			return fmt.Errorf("listing R2 buckets: %w", err)
		}

		buckets := make([]R2BucketSummary, 0, len(resp.Buckets))
		for _, b := range resp.Buckets {
			buckets = append(buckets, toR2BucketSummary(b))
		}

		return printR2Buckets(cmd, buckets)
	}
}

func newR2CreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an R2 bucket",
		RunE:  makeRunR2Create(factory),
	}
	cmd.Flags().String("name", "", "Bucket name (required)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunR2Create(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create R2 bucket %q", name), map[string]any{
				"action": "create",
				"name":   name,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path, err := client.accountPath(fmt.Sprintf("/r2/buckets/%s", name))
		if err != nil {
			return err
		}

		if _, err := client.do(ctx, http.MethodPut, path, map[string]any{}); err != nil {
			return fmt.Errorf("creating R2 bucket %q: %w", name, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "created", "name": name})
		}
		fmt.Printf("Created R2 bucket: %s\n", name)
		return nil
	}
}

func newR2DeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an R2 bucket (irreversible)",
		RunE:  makeRunR2Delete(factory),
	}
	cmd.Flags().String("name", "", "Bucket name (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunR2Delete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete R2 bucket %q", name), map[string]any{
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

		path, err := client.accountPath(fmt.Sprintf("/r2/buckets/%s", name))
		if err != nil {
			return err
		}

		if _, err := client.do(ctx, http.MethodDelete, path, nil); err != nil {
			return fmt.Errorf("deleting R2 bucket %q: %w", name, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "name": name})
		}
		fmt.Printf("Deleted R2 bucket: %s\n", name)
		return nil
	}
}
