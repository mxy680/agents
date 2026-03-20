package framer

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/emdash-projects/agents/internal/cli"
)

func newFontsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fonts",
		Short: "Manage Framer fonts",
	}
	cmd.PersistentFlags().Bool("json", false, "Output as JSON")

	cmd.AddCommand(
		newFontsListCmd(factory),
		newFontsGetCmd(factory),
	)

	return cmd
}

func newFontsListCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all fonts",
		RunE:  makeRunFontsList(factory),
	}
	return cmd
}

func makeRunFontsList(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getFonts", nil)
		if err != nil {
			return fmt.Errorf("list fonts: %w", err)
		}

		var fonts []Font
		if err := json.Unmarshal(result, &fonts); err != nil {
			return fmt.Errorf("parse fonts: %w", err)
		}

		lines := make([]string, 0, len(fonts))
		for _, f := range fonts {
			lines = append(lines, fmt.Sprintf("%s\t%s\t%d", f.Family, f.Style, f.Weight))
		}

		return cli.PrintResult(cmd, fonts, lines)
	}
}

func newFontsGetCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a font by family name",
		RunE:  makeRunFontsGet(factory),
	}
	cmd.Flags().String("family", "", "Font family name (required)")
	_ = cmd.MarkFlagRequired("family")
	return cmd
}

func makeRunFontsGet(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		family, _ := cmd.Flags().GetString("family")
		result, err := client.Call("getFont", map[string]any{"family": family})
		if err != nil {
			return fmt.Errorf("get font: %w", err)
		}

		// Result may be null if font not found
		if string(result) == "null" || len(result) == 0 {
			return cli.PrintResult(cmd, nil, []string{"Font not found"})
		}

		var font Font
		if err := json.Unmarshal(result, &font); err != nil {
			return fmt.Errorf("parse font: %w", err)
		}

		return cli.PrintResult(cmd, font, []string{
			fmt.Sprintf("Family: %s", font.Family),
			fmt.Sprintf("Style:  %s", font.Style),
			fmt.Sprintf("Weight: %d", font.Weight),
		})
	}
}
