package framer

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newAgentCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Agent context and change application commands",
	}
	cmd.AddCommand(
		newAgentSystemPromptCmd(factory),
		newAgentContextCmd(factory),
		newAgentReadCmd(factory),
		newAgentApplyCmd(factory),
	)
	return cmd
}

func newAgentSystemPromptCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system-prompt",
		Short: "Get the agent system prompt",
		RunE:  makeRunAgentSystemPrompt(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunAgentSystemPrompt(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getAgentSystemPrompt", nil)
		if err != nil {
			return fmt.Errorf("get agent system prompt: %w", err)
		}

		var prompt string
		if err := json.Unmarshal(result, &prompt); err != nil {
			return fmt.Errorf("parse system prompt: %w", err)
		}

		// In text mode, print the raw prompt text directly.
		return cli.PrintResult(cmd, prompt, []string{prompt})
	}
}

func newAgentContextCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Get the agent context",
		RunE:  makeRunAgentContext(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunAgentContext(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getAgentContext", nil)
		if err != nil {
			return fmt.Errorf("get agent context: %w", err)
		}

		var raw json.RawMessage
		if err := json.Unmarshal(result, &raw); err != nil {
			return fmt.Errorf("parse agent context: %w", err)
		}

		// Always print as JSON since context is complex structured data.
		return cli.PrintJSON(raw)
	}
}

func newAgentReadCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read project data for agent queries",
		RunE:  makeRunAgentRead(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().String("queries", "", "JSON array of queries (required)")
	_ = cmd.MarkFlagRequired("queries")
	return cmd
}

func makeRunAgentRead(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		queriesStr, _ := cmd.Flags().GetString("queries")
		queries, err := parseJSONFlag(queriesStr)
		if err != nil {
			return fmt.Errorf("parse --queries: %w", err)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("readProjectForAgent", map[string]any{
			"queries": queries,
		})
		if err != nil {
			return fmt.Errorf("read project for agent: %w", err)
		}

		var raw json.RawMessage
		if err := json.Unmarshal(result, &raw); err != nil {
			return fmt.Errorf("parse read result: %w", err)
		}

		return cli.PrintJSON(raw)
	}
}

func newAgentApplyCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply agent DSL changes to the project",
		RunE:  makeRunAgentApply(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("dry-run", false, "Preview without applying changes")
	cmd.Flags().String("dsl", "", "DSL JSON string")
	cmd.Flags().String("dsl-file", "", "Path to DSL JSON file")
	return cmd
}

func makeRunAgentApply(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		dslStr, _ := cmd.Flags().GetString("dsl")
		dslFile, _ := cmd.Flags().GetString("dsl-file")

		dsl, err := parseJSONFlagOrFile(dslStr, dslFile)
		if err != nil {
			return fmt.Errorf("parse DSL: %w", err)
		}

		if ok, err := dryRunResult(cmd, "apply agent changes", dsl); ok {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("applyAgentChanges", map[string]any{
			"dsl": dsl,
		})
		if err != nil {
			return fmt.Errorf("apply agent changes: %w", err)
		}

		var raw json.RawMessage
		if err := json.Unmarshal(result, &raw); err != nil {
			return fmt.Errorf("parse apply result: %w", err)
		}

		return cli.PrintJSON(raw)
	}
}
