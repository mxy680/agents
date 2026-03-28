package vercel

import (
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// LogEvent is the JSON-serializable representation of a deployment log event.
type LogEvent struct {
	ID          string `json:"id,omitempty"`
	Text        string `json:"text,omitempty"`
	Type        string `json:"type,omitempty"`
	Source      string `json:"source,omitempty"`
	DeploymentID string `json:"deploymentId,omitempty"`
	Date        int64  `json:"date,omitempty"`
}

func toLogEvent(data map[string]any) LogEvent {
	return LogEvent{
		ID:           jsonString(data["id"]),
		Text:         jsonString(data["text"]),
		Type:         jsonString(data["type"]),
		Source:       jsonString(data["source"]),
		DeploymentID: jsonString(data["deploymentId"]),
		Date:         jsonInt64(data["date"]),
	}
}

func newLogsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get deployment log events",
		RunE:  makeRunLogsGet(factory),
	}
	cmd.Flags().String("deployment-id", "", "Deployment ID (required)")
	_ = cmd.MarkFlagRequired("deployment-id")
	return cmd
}

func makeRunLogsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		deploymentID, _ := cmd.Flags().GetString("deployment-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var events []map[string]any
		if err := client.doJSON(ctx, http.MethodGet, fmt.Sprintf("/v2/deployments/%s/events", deploymentID), nil, &events); err != nil {
			return fmt.Errorf("getting logs for deployment %q: %w", deploymentID, err)
		}

		logEvents := make([]LogEvent, 0, len(events))
		for _, e := range events {
			logEvents = append(logEvents, toLogEvent(e))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(logEvents)
		}

		if len(logEvents) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No log events found.")
			return nil
		}

		for _, e := range logEvents {
			if e.Text != "" {
				fmt.Fprintln(cmd.OutOrStdout(), e.Text)
			}
		}
		return nil
	}
}
