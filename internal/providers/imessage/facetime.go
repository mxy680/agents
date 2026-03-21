package imessage

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newFaceTimeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "facetime",
		Short:   "Manage FaceTime calls",
		Aliases: []string{"ft"},
	}

	cmd.AddCommand(newFaceTimeCallCmd(factory))
	cmd.AddCommand(newFaceTimeAnswerCmd(factory))
	cmd.AddCommand(newFaceTimeLeaveCmd(factory))

	return cmd
}

func newFaceTimeCallCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call",
		Short: "Start a FaceTime call",
	}
	cmd.Flags().String("addresses", "", "Comma-separated addresses to call (required)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without executing")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("addresses")
	cmd.RunE = makeRunFaceTimeCall(factory)
	return cmd
}

func makeRunFaceTimeCall(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		addressesRaw, _ := cmd.Flags().GetString("addresses")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		addresses := splitCSV(addressesRaw)

		if dryRun {
			result := dryRunResult("facetime call", map[string]any{
				"addresses": addresses,
			})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would call FaceTime with: %s", addressesRaw)})
		}

		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		body := map[string]any{"handle": addresses}
		resp, err := client.Post(cmd.Context(), "facetime/session", body)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		return printResult(cmd, data, []string{fmt.Sprintf("FaceTime call initiated with: %s", strings.Join(addresses, ", "))})
	}
}

func newFaceTimeAnswerCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "answer",
		Short: "Answer a FaceTime call",
	}
	cmd.Flags().String("call-uuid", "", "UUID of the call to answer (required)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without executing")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("call-uuid")
	cmd.RunE = makeRunFaceTimeAnswer(factory)
	return cmd
}

func makeRunFaceTimeAnswer(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		callUUID, _ := cmd.Flags().GetString("call-uuid")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			result := dryRunResult("facetime answer", map[string]any{"call_uuid": callUUID})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would answer FaceTime call: %s", callUUID)})
		}

		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Post(cmd.Context(), fmt.Sprintf("facetime/answer/%s", callUUID), nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		return printResult(cmd, data, []string{fmt.Sprintf("Answered FaceTime call: %s", callUUID)})
	}
}

func newFaceTimeLeaveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leave",
		Short: "Leave a FaceTime call",
	}
	cmd.Flags().String("call-uuid", "", "UUID of the call to leave (required)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without executing")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("call-uuid")
	cmd.RunE = makeRunFaceTimeLeave(factory)
	return cmd
}

func makeRunFaceTimeLeave(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		callUUID, _ := cmd.Flags().GetString("call-uuid")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			result := dryRunResult("facetime leave", map[string]any{"call_uuid": callUUID})
			return printResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would leave FaceTime call: %s", callUUID)})
		}

		client, err := factory(cmd.Context())
		if err != nil {
			return err
		}

		resp, err := client.Post(cmd.Context(), fmt.Sprintf("facetime/leave/%s", callUUID), nil)
		if err != nil {
			return err
		}

		data, err := ParseResponse(resp)
		if err != nil {
			return err
		}

		return printResult(cmd, data, []string{fmt.Sprintf("Left FaceTime call: %s", callUUID)})
	}
}

// splitCSV splits a comma-separated string and trims whitespace from each element.
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
