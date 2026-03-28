package fly

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newMachinesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List machines in an app",
		RunE:  makeRunMachinesList(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	_ = cmd.MarkFlagRequired("app")
	return cmd
}

func makeRunMachinesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var machines []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			State  string `json:"state"`
			Region string `json:"region"`
			Config *struct {
				Image string `json:"image"`
			} `json:"config"`
		}
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/apps/%s/machines", app), nil, &machines); err != nil {
			return fmt.Errorf("listing machines for app %q: %w", app, err)
		}

		summaries := make([]MachineSummary, 0, len(machines))
		for _, m := range machines {
			img := ""
			if m.Config != nil {
				img = m.Config.Image
			}
			summaries = append(summaries, MachineSummary{
				ID:     m.ID,
				Name:   m.Name,
				State:  m.State,
				Region: m.Region,
				Image:  img,
			})
		}

		return printMachineSummaries(cmd, summaries)
	}
}

func newMachinesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get machine details",
		RunE:  makeRunMachinesGet(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("machine", "", "Machine ID (required)")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("machine")
	return cmd
}

func makeRunMachinesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		machine, _ := cmd.Flags().GetString("machine")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data MachineDetail
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v1/apps/%s/machines/%s", app, machine), nil, &data); err != nil {
			return fmt.Errorf("getting machine %q: %w", machine, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		image := ""
		if data.Config != nil {
			image = data.Config.Image
		}
		lines := []string{
			fmt.Sprintf("ID:          %s", data.ID),
			fmt.Sprintf("Name:        %s", data.Name),
			fmt.Sprintf("State:       %s", data.State),
			fmt.Sprintf("Region:      %s", data.Region),
			fmt.Sprintf("Instance ID: %s", data.InstanceID),
			fmt.Sprintf("Private IP:  %s", data.PrivateIP),
			fmt.Sprintf("Image:       %s", image),
			fmt.Sprintf("Created At:  %s", data.CreatedAt),
			fmt.Sprintf("Updated At:  %s", data.UpdatedAt),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newMachinesCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new machine",
		RunE:  makeRunMachinesCreate(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("image", "", "Docker image to run (required)")
	cmd.Flags().String("region", "", "Region to deploy in (optional)")
	cmd.Flags().String("name", "", "Machine name (optional)")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("image")
	return cmd
}

func makeRunMachinesCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		image, _ := cmd.Flags().GetString("image")
		region, _ := cmd.Flags().GetString("region")
		name, _ := cmd.Flags().GetString("name")

		body := map[string]any{
			"config": map[string]any{"image": image},
		}
		if region != "" {
			body["region"] = region
		}
		if name != "" {
			body["name"] = name
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create machine in app %q with image %q", app, image), body)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data MachineDetail
		if err := client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v1/apps/%s/machines", app), body, &data); err != nil {
			return fmt.Errorf("creating machine in app %q: %w", app, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}
		fmt.Printf("Created machine: %s (ID: %s, state: %s)\n", data.Name, data.ID, data.State)
		return nil
	}
}

func newMachinesUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a machine's configuration",
		RunE:  makeRunMachinesUpdate(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("machine", "", "Machine ID (required)")
	cmd.Flags().String("image", "", "New Docker image (required)")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("machine")
	_ = cmd.MarkFlagRequired("image")
	return cmd
}

func makeRunMachinesUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		machine, _ := cmd.Flags().GetString("machine")
		image, _ := cmd.Flags().GetString("image")

		body := map[string]any{
			"config": map[string]any{"image": image},
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update machine %q in app %q", machine, app), body)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data MachineDetail
		if err := client.doJSON(ctx, http.MethodPost, fmt.Sprintf("/v1/apps/%s/machines/%s", app, machine), body, &data); err != nil {
			return fmt.Errorf("updating machine %q: %w", machine, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}
		fmt.Printf("Updated machine: %s (ID: %s, state: %s)\n", data.Name, data.ID, data.State)
		return nil
	}
}

func newMachinesDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a machine (irreversible)",
		RunE:  makeRunMachinesDelete(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("machine", "", "Machine ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("machine")
	return cmd
}

func makeRunMachinesDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		machine, _ := cmd.Flags().GetString("machine")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete machine %q in app %q", machine, app), map[string]any{
				"action":  "delete",
				"app":     app,
				"machine": machine,
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

		if _, err := client.do(ctx, http.MethodDelete, fmt.Sprintf("/v1/apps/%s/machines/%s", app, machine), nil); err != nil {
			return fmt.Errorf("deleting machine %q: %w", machine, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "machine": machine})
		}
		fmt.Printf("Deleted machine: %s\n", machine)
		return nil
	}
}

func newMachinesStartCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a stopped machine",
		RunE:  makeRunMachinesStart(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("machine", "", "Machine ID (required)")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("machine")
	return cmd
}

func makeRunMachinesStart(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		machine, _ := cmd.Flags().GetString("machine")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if _, err := client.do(ctx, http.MethodPost, fmt.Sprintf("/v1/apps/%s/machines/%s/start", app, machine), nil); err != nil {
			return fmt.Errorf("starting machine %q: %w", machine, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "started", "machine": machine})
		}
		fmt.Printf("Started machine: %s\n", machine)
		return nil
	}
}

func newMachinesStopCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop a running machine",
		RunE:  makeRunMachinesStop(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("machine", "", "Machine ID (required)")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("machine")
	return cmd
}

func makeRunMachinesStop(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		machine, _ := cmd.Flags().GetString("machine")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if _, err := client.do(ctx, http.MethodPost, fmt.Sprintf("/v1/apps/%s/machines/%s/stop", app, machine), nil); err != nil {
			return fmt.Errorf("stopping machine %q: %w", machine, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "stopped", "machine": machine})
		}
		fmt.Printf("Stopped machine: %s\n", machine)
		return nil
	}
}

func newMachinesWaitCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Wait for a machine to reach a given state",
		RunE:  makeRunMachinesWait(factory),
	}
	cmd.Flags().String("app", "", "App name (required)")
	cmd.Flags().String("machine", "", "Machine ID (required)")
	cmd.Flags().String("state", "started", "Target state to wait for (e.g. started, stopped, destroyed)")
	cmd.Flags().Int("timeout", 30, "Timeout in seconds")
	_ = cmd.MarkFlagRequired("app")
	_ = cmd.MarkFlagRequired("machine")
	return cmd
}

func makeRunMachinesWait(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		app, _ := cmd.Flags().GetString("app")
		machine, _ := cmd.Flags().GetString("machine")
		state, _ := cmd.Flags().GetString("state")
		timeout, _ := cmd.Flags().GetInt("timeout")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/v1/apps/%s/machines/%s/wait?state=%s&timeout=%d", app, machine, state, timeout)
		if _, err := client.do(ctx, http.MethodGet, path, nil); err != nil {
			return fmt.Errorf("waiting for machine %q to reach state %q: %w", machine, state, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"machine": machine, "state": state})
		}
		fmt.Printf("Machine %s reached state: %s\n", machine, state)
		return nil
	}
}
