package imessage

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func newHandlesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "handles",
		Short:   "Manage message handles (phone numbers/emails)",
		Aliases: []string{"handle"},
	}

	cmd.AddCommand(newHandlesListCmd(factory))
	cmd.AddCommand(newHandlesGetCmd(factory))
	cmd.AddCommand(newHandlesCountCmd(factory))
	cmd.AddCommand(newHandlesFocusCmd(factory))
	cmd.AddCommand(newHandlesAvailabilityCmd(factory))

	return cmd
}

func newHandlesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List handles",
		RunE:  makeRunHandlesList(factory),
	}
	cmd.Flags().Int("limit", 25, "Maximum number of handles to return")
	cmd.Flags().Int("offset", 0, "Pagination offset")
	cmd.Flags().String("query", "", "Filter handles by address")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunHandlesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		query, _ := cmd.Flags().GetString("query")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		reqBody := map[string]any{
			"limit":  limit,
			"offset": offset,
		}
		if query != "" {
			reqBody["query"] = query
		}

		body, err := client.Post(ctx, "/handle/query", reqBody)
		if err != nil {
			return fmt.Errorf("listing handles: %w", err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		var handles []json.RawMessage
		if err := json.Unmarshal(data, &handles); err != nil {
			return fmt.Errorf("parse handles list: %w", err)
		}

		summaries := make([]HandleSummary, 0, len(handles))
		for _, h := range handles {
			summaries = append(summaries, toHandleSummary(h))
		}

		lines := make([]string, 0, len(summaries))
		for _, s := range summaries {
			lines = append(lines, fmt.Sprintf("%-40s  %s", s.Address, s.Service))
		}
		return printResult(cmd, summaries, lines)
	}
}

func newHandlesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get handle details by GUID",
		RunE:  makeRunHandlesGet(factory),
	}
	cmd.Flags().String("guid", "", "Handle GUID (e.g. iMessage;-;+1234567890) (required)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("guid")
	return cmd
}

func makeRunHandlesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Get(ctx, fmt.Sprintf("/handle/%s", guid), nil)
		if err != nil {
			return fmt.Errorf("getting handle %s: %w", guid, err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		summary := toHandleSummary(data)
		return printResult(cmd, summary, []string{
			fmt.Sprintf("Address:  %s", summary.Address),
			fmt.Sprintf("Service:  %s", summary.Service),
			fmt.Sprintf("Country:  %s", summary.Country),
			fmt.Sprintf("Uncanon:  %s", summary.UncanonID),
		})
	}
}

func newHandlesCountCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count",
		Short: "Get the total number of handles",
		RunE:  makeRunHandlesCount(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunHandlesCount(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Get(ctx, "/handle/count", nil)
		if err != nil {
			return fmt.Errorf("getting handle count: %w", err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parse count response: %w", err)
		}

		count := getInt64(raw, "total")
		return printResult(cmd, raw, []string{
			fmt.Sprintf("Total handles: %d", count),
		})
	}
}

func newHandlesFocusCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "focus",
		Short: "Get focus status for a handle",
		RunE:  makeRunHandlesFocus(factory),
	}
	cmd.Flags().String("guid", "", "Handle GUID (required)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("guid")
	return cmd
}

func makeRunHandlesFocus(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Get(ctx, fmt.Sprintf("/handle/%s/focus", guid), nil)
		if err != nil {
			return fmt.Errorf("getting focus status for handle %s: %w", guid, err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parse focus response: %w", err)
		}

		focusStatus := getString(raw, "focusStatus")
		return printResult(cmd, raw, []string{
			fmt.Sprintf("Handle:       %s", guid),
			fmt.Sprintf("Focus Status: %s", focusStatus),
		})
	}
}

func newHandlesAvailabilityCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "availability",
		Short: "Check if an address is available on iMessage or FaceTime",
		RunE:  makeRunHandlesAvailability(factory),
	}
	cmd.Flags().String("address", "", "Phone number or email address to check (required)")
	cmd.Flags().String("service", "imessage", "Service to check: imessage or facetime")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("address")
	return cmd
}

func makeRunHandlesAvailability(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		address, _ := cmd.Flags().GetString("address")
		service, _ := cmd.Flags().GetString("service")

		if service != "imessage" && service != "facetime" {
			return fmt.Errorf("--service must be \"imessage\" or \"facetime\", got %q", service)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := make(map[string][]string)
		params["address"] = []string{address}

		body, err := client.Get(ctx, fmt.Sprintf("/handle/availability/%s", service), params)
		if err != nil {
			return fmt.Errorf("checking %s availability for %s: %w", service, address, err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parse availability response: %w", err)
		}

		available := getBool(raw, "available")
		availStr := "no"
		if available {
			availStr = "yes"
		}
		return printResult(cmd, raw, []string{
			fmt.Sprintf("Address:   %s", address),
			fmt.Sprintf("Service:   %s", service),
			fmt.Sprintf("Available: %s", availStr),
		})
	}
}
