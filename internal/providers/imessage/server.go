package imessage

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newServerCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "BlueBubbles server management",
	}

	cmd.AddCommand(newServerInfoCmd(factory))
	cmd.AddCommand(newServerLogsCmd(factory))
	cmd.AddCommand(newServerRestartCmd(factory))
	cmd.AddCommand(newServerUpdateCheckCmd(factory))
	cmd.AddCommand(newServerUpdateInstallCmd(factory))
	cmd.AddCommand(newServerAlertsCmd(factory))
	cmd.AddCommand(newServerAlertsReadCmd(factory))
	cmd.AddCommand(newServerStatsCmd(factory))

	return cmd
}

func newServerInfoCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get server information",
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunServerInfo(factory)
	return cmd
}

func makeRunServerInfo(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Get(cmd.Context(), "server/info", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		var info ServerInfo
		if err := json.Unmarshal(data, &info); err != nil {
			return printResult(cmd, data, []string{string(data)})
		}

		privateAPI := "disabled"
		if info.PrivateAPI {
			privateAPI = "enabled"
		}
		lines := []string{
			fmt.Sprintf("Server Version: %s", info.ServerVersion),
			fmt.Sprintf("OS Version:     %s", info.OSVersion),
			fmt.Sprintf("Mac Model:      %s", info.MacModel),
			fmt.Sprintf("Private API:    %s", privateAPI),
			fmt.Sprintf("Proxy Service:  %s", info.ProxyService),
		}
		return printResult(cmd, info, lines)
	}
}

func newServerLogsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Get server logs",
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunServerLogs(factory)
	return cmd
}

func makeRunServerLogs(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Get(cmd.Context(), "server/logs", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		return printResult(cmd, data, []string{string(data)})
	}
}

func newServerRestartCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the BlueBubbles server",
	}
	cmd.Flags().Bool("soft", true, "Perform a soft restart (default true; use --soft=false for hard restart)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without executing")
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunServerRestart(factory)
	return cmd
}

func makeRunServerRestart(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		soft, _ := cmd.Flags().GetBool("soft")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		restartType := "hard"
		if soft {
			restartType = "soft"
		}

		if dryRun {
			result := dryRunResult("server restart", map[string]any{"type": restartType})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would perform %s restart.", restartType)})
		}

		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Get(cmd.Context(), fmt.Sprintf("server/restart/%s", restartType), nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		return printResult(cmd, data, []string{fmt.Sprintf("Server %s restart initiated.", restartType)})
	}
}

func newServerUpdateCheckCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-check",
		Short: "Check for server updates",
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunServerUpdateCheck(factory)
	return cmd
}

func makeRunServerUpdateCheck(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Get(cmd.Context(), "server/update/check", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		var m map[string]any
		if err := json.Unmarshal(data, &m); err != nil {
			return printResult(cmd, data, []string{string(data)})
		}

		available := getBool(m, "available")
		version := getString(m, "version")
		if available {
			return printResult(cmd, m, []string{fmt.Sprintf("Update available: %s", version)})
		}
		return printResult(cmd, m, []string{"No update available."})
	}
}

func newServerUpdateInstallCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-install",
		Short: "Install a server update",
	}
	cmd.Flags().Bool("dry-run", false, "Preview the action without executing")
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunServerUpdateInstall(factory)
	return cmd
}

func makeRunServerUpdateInstall(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			result := dryRunResult("server update-install", nil)
			return printResult(cmd, result, []string{"[dry-run] Would install server update."})
		}

		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Post(cmd.Context(), "server/update/install", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		return printResult(cmd, data, []string{"Server update installation initiated."})
	}
}

func newServerAlertsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alerts",
		Short: "Get server alerts",
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunServerAlerts(factory)
	return cmd
}

func makeRunServerAlerts(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Get(cmd.Context(), "server/alert", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		var alerts []map[string]any
		if err := json.Unmarshal(data, &alerts); err != nil {
			return printResult(cmd, data, []string{string(data)})
		}

		lines := make([]string, 0, len(alerts))
		for _, a := range alerts {
			lines = append(lines, fmt.Sprintf("[%s] %s: %s",
				getString(a, "type"),
				getString(a, "name"),
				truncate(getString(a, "value"), 80),
			))
		}
		if len(lines) == 0 {
			lines = []string{"No alerts."}
		}

		return printResult(cmd, alerts, lines)
	}
}

func newServerAlertsReadCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alerts-read",
		Short: "Mark server alerts as read",
	}
	cmd.Flags().Bool("dry-run", false, "Preview the action without executing")
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunServerAlertsRead(factory)
	return cmd
}

func makeRunServerAlertsRead(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			result := dryRunResult("server alerts-read", nil)
			return printResult(cmd, result, []string{"[dry-run] Would mark all alerts as read."})
		}

		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Post(cmd.Context(), "server/alert/read", nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		return printResult(cmd, data, []string{"Server alerts marked as read."})
	}
}

func newServerStatsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Get server statistics",
	}
	cmd.Flags().String("type", "totals", "Statistics type: totals, media, or media-by-chat")
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.RunE = makeRunServerStats(factory)
	return cmd
}

func makeRunServerStats(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		statsType, _ := cmd.Flags().GetString("type")

		var path string
		switch statsType {
		case "totals", "":
			path = "server/statistics/totals"
		case "media":
			path = "server/statistics/media"
		case "media-by-chat":
			path = "server/statistics/media/chat"
		default:
			return fmt.Errorf("unknown stats type %q: must be totals, media, or media-by-chat", statsType)
		}

		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Get(cmd.Context(), path, nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		return printResult(cmd, data, []string{string(data)})
	}
}
