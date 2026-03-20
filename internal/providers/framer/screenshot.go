package framer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newScreenshotCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "screenshot",
		Short: "Screenshot and SVG export commands",
	}
	cmd.AddCommand(
		newScreenshotTakeCmd(factory),
		newScreenshotExportSVGCmd(factory),
	)
	return cmd
}

func newScreenshotTakeCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "take",
		Short: "Take a screenshot of a node",
		RunE:  makeRunScreenshotTake(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().String("node-id", "", "Node ID to screenshot (required)")
	cmd.Flags().String("format", "png", "Image format: png or jpeg")
	cmd.Flags().Float64("scale", 1.0, "Scale factor for the screenshot")
	cmd.Flags().String("output", "", "File path to write the image to")
	_ = cmd.MarkFlagRequired("node-id")
	return cmd
}

func makeRunScreenshotTake(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		nodeID, _ := cmd.Flags().GetString("node-id")
		format, _ := cmd.Flags().GetString("format")
		scale, _ := cmd.Flags().GetFloat64("scale")
		outputPath, _ := cmd.Flags().GetString("output")

		params := map[string]any{
			"nodeId": nodeID,
			"format": format,
			"scale":  scale,
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("screenshot", params)
		if err != nil {
			return fmt.Errorf("take screenshot: %w", err)
		}

		// If --output is set, the result is expected to contain binary image data.
		if outputPath != "" {
			var screenshotResult ScreenshotResult
			if err := json.Unmarshal(result, &screenshotResult); err != nil {
				return fmt.Errorf("parse screenshot result: %w", err)
			}
			if len(screenshotResult.Image) > 0 {
				if err := os.WriteFile(outputPath, screenshotResult.Image, 0644); err != nil {
					return fmt.Errorf("write screenshot to %s: %w", outputPath, err)
				}
				return cli.PrintResult(cmd, map[string]string{"output": outputPath}, []string{
					fmt.Sprintf("Screenshot saved to %s", outputPath),
				})
			}
			return fmt.Errorf("screenshot result contained no image data")
		}

		// No output file: print the URL or base64 via JSON.
		var raw json.RawMessage
		if err := json.Unmarshal(result, &raw); err != nil {
			return fmt.Errorf("parse screenshot result: %w", err)
		}
		return cli.PrintJSON(raw)
	}
}

func newScreenshotExportSVGCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-svg",
		Short: "Export a node as SVG",
		RunE:  makeRunScreenshotExportSVG(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().String("node-id", "", "Node ID to export (required)")
	cmd.Flags().String("output", "", "File path to write the SVG to")
	_ = cmd.MarkFlagRequired("node-id")
	return cmd
}

func makeRunScreenshotExportSVG(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		nodeID, _ := cmd.Flags().GetString("node-id")
		outputPath, _ := cmd.Flags().GetString("output")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("exportSVG", map[string]any{
			"nodeId": nodeID,
		})
		if err != nil {
			return fmt.Errorf("export SVG: %w", err)
		}

		var svgContent string
		if err := json.Unmarshal(result, &svgContent); err != nil {
			return fmt.Errorf("parse SVG result: %w", err)
		}

		if outputPath != "" {
			if err := os.WriteFile(outputPath, []byte(svgContent), 0644); err != nil {
				return fmt.Errorf("write SVG to %s: %w", outputPath, err)
			}
			return cli.PrintResult(cmd, map[string]string{"output": outputPath}, []string{
				fmt.Sprintf("SVG saved to %s", outputPath),
			})
		}

		return cli.PrintResult(cmd, svgContent, []string{svgContent})
	}
}
