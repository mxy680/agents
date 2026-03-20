package framer

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newSVGCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "svg",
		Short: "SVG and vector set commands",
	}
	cmd.AddCommand(
		newSVGAddCmd(factory),
		newSVGVectorSetsCmd(factory),
	)
	return cmd
}

func newSVGAddCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an SVG to the project",
		RunE:  makeRunSVGAdd(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("dry-run", false, "Preview without making changes")
	cmd.Flags().String("svg", "", "SVG content")
	cmd.Flags().String("svg-file", "", "Path to file containing SVG content")
	return cmd
}

func makeRunSVGAdd(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		svgVal, _ := cmd.Flags().GetString("svg")
		svgFile, _ := cmd.Flags().GetString("svg-file")

		svgContent, err := readFileOrFlag(svgVal, svgFile)
		if err != nil {
			return fmt.Errorf("--svg: %w", err)
		}
		if svgContent == "" {
			return fmt.Errorf("--svg or --svg-file is required")
		}

		params := map[string]any{
			"svg": svgContent,
		}

		if isDry, result := dryRunResult(cmd, "add SVG", params); isDry {
			return result
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("addSVG", params)
		if err != nil {
			return fmt.Errorf("add SVG: %w", err)
		}

		var raw json.RawMessage = result
		return cli.PrintResult(cmd, raw, []string{"SVG added"})
	}
}

func newSVGVectorSetsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vector-sets",
		Short: "Get vector sets",
		RunE:  makeRunSVGVectorSets(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunSVGVectorSets(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getVectorSets", nil)
		if err != nil {
			return fmt.Errorf("get vector sets: %w", err)
		}

		var raw json.RawMessage = result
		return cli.PrintResult(cmd, raw, []string{string(result)})
	}
}
