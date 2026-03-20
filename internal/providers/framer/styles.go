package framer

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/emdash-projects/agents/internal/cli"
)

func newStylesCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "styles",
		Short: "Manage Framer design styles",
	}
	cmd.PersistentFlags().Bool("json", false, "Output as JSON")

	colorsCmd := &cobra.Command{
		Use:   "colors",
		Short: "Manage color styles",
	}
	colorsCmd.AddCommand(
		newStylesColorsListCmd(factory),
		newStylesColorsGetCmd(factory),
		newStylesColorsCreateCmd(factory),
	)

	textCmd := &cobra.Command{
		Use:   "text",
		Short: "Manage text styles",
	}
	textCmd.AddCommand(
		newStylesTextListCmd(factory),
		newStylesTextGetCmd(factory),
		newStylesTextCreateCmd(factory),
	)

	cmd.AddCommand(colorsCmd, textCmd)

	return cmd
}

func newStylesColorsListCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all color styles",
		RunE:  makeRunStylesColorsList(factory),
	}
	return cmd
}

func makeRunStylesColorsList(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getColorStyles", nil)
		if err != nil {
			return fmt.Errorf("list color styles: %w", err)
		}

		var styles []ColorStyle
		if err := json.Unmarshal(result, &styles); err != nil {
			return fmt.Errorf("parse color styles: %w", err)
		}

		lines := make([]string, 0, len(styles))
		for _, s := range styles {
			lines = append(lines, fmt.Sprintf("%s\t%s\t%s", s.ID, s.Name, s.Light))
		}

		return cli.PrintResult(cmd, styles, lines)
	}
}

func newStylesColorsGetCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a color style by ID",
		RunE:  makeRunStylesColorsGet(factory),
	}
	cmd.Flags().String("id", "", "Color style ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunStylesColorsGet(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		id, _ := cmd.Flags().GetString("id")
		result, err := client.Call("getColorStyle", map[string]any{"id": id})
		if err != nil {
			return fmt.Errorf("get color style: %w", err)
		}

		var style ColorStyle
		if err := json.Unmarshal(result, &style); err != nil {
			return fmt.Errorf("parse color style: %w", err)
		}

		return cli.PrintResult(cmd, style, []string{
			fmt.Sprintf("ID:    %s", style.ID),
			fmt.Sprintf("Name:  %s", style.Name),
			fmt.Sprintf("Light: %s", style.Light),
			fmt.Sprintf("Dark:  %s", style.Dark),
		})
	}
}

func newStylesColorsCreateCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a color style",
		RunE:  makeRunStylesColorsCreate(factory),
	}
	cmd.Flags().String("attributes", "", "Color style attributes as JSON (required)")
	_ = cmd.MarkFlagRequired("attributes")
	cmd.Flags().Bool("dry-run", false, "Preview without creating")
	return cmd
}

func makeRunStylesColorsCreate(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		attrsStr, _ := cmd.Flags().GetString("attributes")

		attrs, err := parseJSONFlag(attrsStr)
		if err != nil {
			return fmt.Errorf("parse attributes: %w", err)
		}

		params := map[string]any{"attributes": attrs}

		if isDry, err := dryRunResult(cmd, "create color style", params); isDry {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("createColorStyle", params)
		if err != nil {
			return fmt.Errorf("create color style: %w", err)
		}

		var style ColorStyle
		if err := json.Unmarshal(result, &style); err != nil {
			return fmt.Errorf("parse color style: %w", err)
		}

		return cli.PrintResult(cmd, style, []string{
			fmt.Sprintf("Created color style: %s (%s)", style.Name, style.ID),
		})
	}
}

func newStylesTextListCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all text styles",
		RunE:  makeRunStylesTextList(factory),
	}
	return cmd
}

func makeRunStylesTextList(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getTextStyles", nil)
		if err != nil {
			return fmt.Errorf("list text styles: %w", err)
		}

		var styles []TextStyle
		if err := json.Unmarshal(result, &styles); err != nil {
			return fmt.Errorf("parse text styles: %w", err)
		}

		lines := make([]string, 0, len(styles))
		for _, s := range styles {
			lines = append(lines, fmt.Sprintf("%s\t%s\t%s", s.ID, s.Name, s.Font))
		}

		return cli.PrintResult(cmd, styles, lines)
	}
}

func newStylesTextGetCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a text style by ID",
		RunE:  makeRunStylesTextGet(factory),
	}
	cmd.Flags().String("id", "", "Text style ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunStylesTextGet(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		id, _ := cmd.Flags().GetString("id")
		result, err := client.Call("getTextStyle", map[string]any{"id": id})
		if err != nil {
			return fmt.Errorf("get text style: %w", err)
		}

		var style TextStyle
		if err := json.Unmarshal(result, &style); err != nil {
			return fmt.Errorf("parse text style: %w", err)
		}

		return cli.PrintResult(cmd, style, []string{
			fmt.Sprintf("ID:       %s", style.ID),
			fmt.Sprintf("Name:     %s", style.Name),
			fmt.Sprintf("Font:     %s", style.Font),
			fmt.Sprintf("FontSize: %g", style.FontSize),
		})
	}
}

func newStylesTextCreateCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a text style",
		RunE:  makeRunStylesTextCreate(factory),
	}
	cmd.Flags().String("attributes", "", "Text style attributes as JSON (required)")
	_ = cmd.MarkFlagRequired("attributes")
	cmd.Flags().Bool("dry-run", false, "Preview without creating")
	return cmd
}

func makeRunStylesTextCreate(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		attrsStr, _ := cmd.Flags().GetString("attributes")

		attrs, err := parseJSONFlag(attrsStr)
		if err != nil {
			return fmt.Errorf("parse attributes: %w", err)
		}

		params := map[string]any{"attributes": attrs}

		if isDry, err := dryRunResult(cmd, "create text style", params); isDry {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("createTextStyle", params)
		if err != nil {
			return fmt.Errorf("create text style: %w", err)
		}

		var style TextStyle
		if err := json.Unmarshal(result, &style); err != nil {
			return fmt.Errorf("parse text style: %w", err)
		}

		return cli.PrintResult(cmd, style, []string{
			fmt.Sprintf("Created text style: %s (%s)", style.Name, style.ID),
		})
	}
}
