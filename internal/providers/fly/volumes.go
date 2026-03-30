package fly

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newVolumesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List volumes in an app",
		RunE:  makeRunVolumesList(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	_ = cmd.MarkFlagRequired("app")
	return cmd
}

func makeRunVolumesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var volumes []VolumeSummary
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/apps/%s/volumes", app), nil, &volumes); err != nil {
			return fmt.Errorf("listing volumes for app %q: %w", app, err)
		}

		return printVolumeSummaries(cmd, volumes)
	}
}

func newVolumesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get volume details",
		RunE:  makeRunVolumesGet(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("volume", "", "Volume ID (required)")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("volume")
	return cmd
}

func makeRunVolumesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		volume, _ := cmd.Flags().GetString("volume")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data VolumeDetail
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/apps/%s/volumes/%s", app, volume), nil, &data); err != nil {
			return fmt.Errorf("getting volume %q: %w", volume, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		lines := []string{
			fmt.Sprintf("ID:                 %s", data.ID),
			fmt.Sprintf("Name:               %s", data.Name),
			fmt.Sprintf("State:              %s", data.State),
			fmt.Sprintf("Region:             %s", data.Region),
			fmt.Sprintf("Size (GB):          %d", data.SizeGB),
			fmt.Sprintf("Encrypted:          %v", data.Encrypted),
			fmt.Sprintf("Attached Machine:   %s", data.AttachedMachineID),
			fmt.Sprintf("Created At:         %s", data.CreatedAt),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newVolumesCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new volume",
		RunE:  makeRunVolumesCreate(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("name", "", "Volume name (required)")
	cmd.Flags().String("region", "", "Region to create volume in (required)")
	cmd.Flags().Int("size", 1, "Volume size in GB")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("region")
	return cmd
}

func makeRunVolumesCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		size, _ := cmd.Flags().GetInt("size")

		body := map[string]any{
			"name":    name,
			"region":  region,
			"size_gb": size,
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create volume %q in app %q region %q (%dGB)", name, app, region, size), body)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data VolumeDetail
		if err := client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v1/apps/%s/volumes", app), body, &data); err != nil {
			return fmt.Errorf("creating volume %q in app %q: %w", name, app, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}
		fmt.Printf("Created volume: %s (ID: %s, region: %s, size: %dGB)\n", data.Name, data.ID, data.Region, data.SizeGB)
		return nil
	}
}

func newVolumesExtendCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extend",
		Short: "Extend a volume's size",
		RunE:  makeRunVolumesExtend(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("volume", "", "Volume ID (required)")
	cmd.Flags().Int("size", 0, "New size in GB (required)")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("volume")
	_ = cmd.MarkFlagRequired("size")
	return cmd
}

func makeRunVolumesExtend(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		volume, _ := cmd.Flags().GetString("volume")
		size, _ := cmd.Flags().GetInt("size")

		body := map[string]any{"size_gb": size}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would extend volume %q to %dGB", volume, size), body)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data VolumeDetail
		if err := client.doJSON(ctx, http.MethodPut, fmt.Sprintf("/v1/apps/%s/volumes/%s/extend", app, volume), body, &data); err != nil {
			return fmt.Errorf("extending volume %q: %w", volume, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}
		fmt.Printf("Extended volume %s to %dGB\n", data.ID, data.SizeGB)
		return nil
	}
}

func newVolumesDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a volume (irreversible)",
		RunE:  makeRunVolumesDelete(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("volume", "", "Volume ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("volume")
	return cmd
}

func makeRunVolumesDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		volume, _ := cmd.Flags().GetString("volume")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete volume %q in app %q", volume, app), map[string]any{
				"action": "delete",
				"app":    app,
				"volume": volume,
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

		if _, err := client.do(ctx, http.MethodDelete, fmt.Sprintf("/v1/apps/%s/volumes/%s", app, volume), nil); err != nil {
			return fmt.Errorf("deleting volume %q: %w", volume, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "volume": volume})
		}
		fmt.Printf("Deleted volume: %s\n", volume)
		return nil
	}
}

func newVolumesSnapshotsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshots",
		Short: "List snapshots for a volume",
		RunE:  makeRunVolumesSnapshots(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("volume", "", "Volume ID (required)")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("volume")
	return cmd
}

func makeRunVolumesSnapshots(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		volume, _ := cmd.Flags().GetString("volume")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var snapshots []VolumeSnapshot
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/apps/%s/volumes/%s/snapshots", app, volume), nil, &snapshots); err != nil {
			return fmt.Errorf("listing snapshots for volume %q: %w", volume, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(snapshots)
		}
		if len(snapshots) == 0 {
			fmt.Println("No snapshots found.")
			return nil
		}
		lines := make([]string, 0, len(snapshots)+1)
		lines = append(lines, fmt.Sprintf("%-28s  %-12s  %-12s  %s", "ID", "STATUS", "SIZE", "CREATED"))
		for _, s := range snapshots {
			lines = append(lines, fmt.Sprintf("%-28s  %-12s  %-12d  %s",
				truncate(s.ID, 28), s.Status, s.SizeBytes, s.CreatedAt))
		}
		cli.PrintText(lines)
		return nil
	}
}
