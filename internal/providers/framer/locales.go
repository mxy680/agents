package framer

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/emdash-projects/agents/internal/cli"
)

func newLocalesCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "locales",
		Short: "Manage Framer localization",
	}
	cmd.PersistentFlags().Bool("json", false, "Output as JSON")

	cmd.AddCommand(
		newLocalesListCmd(factory),
		newLocalesDefaultCmd(factory),
		newLocalesCreateCmd(factory),
		newLocalesLanguagesCmd(factory),
		newLocalesRegionsCmd(factory),
		newLocalesGroupsCmd(factory),
		newLocalesSetDataCmd(factory),
	)

	return cmd
}

func newLocalesListCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all locales",
		RunE:  makeRunLocalesList(factory),
	}
	return cmd
}

func makeRunLocalesList(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getLocales", nil)
		if err != nil {
			return fmt.Errorf("list locales: %w", err)
		}

		var locales []Locale
		if err := json.Unmarshal(result, &locales); err != nil {
			return fmt.Errorf("parse locales: %w", err)
		}

		lines := make([]string, 0, len(locales))
		for _, l := range locales {
			lines = append(lines, fmt.Sprintf("%s\t%s\t%s", l.ID, l.Code, l.Name))
		}

		return cli.PrintResult(cmd, locales, lines)
	}
}

func newLocalesDefaultCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "default",
		Short: "Get the default locale",
		RunE:  makeRunLocalesDefault(factory),
	}
	return cmd
}

func makeRunLocalesDefault(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getDefaultLocale", nil)
		if err != nil {
			return fmt.Errorf("get default locale: %w", err)
		}

		var locale Locale
		if err := json.Unmarshal(result, &locale); err != nil {
			return fmt.Errorf("parse locale: %w", err)
		}

		return cli.PrintResult(cmd, locale, []string{
			fmt.Sprintf("ID:   %s", locale.ID),
			fmt.Sprintf("Code: %s", locale.Code),
			fmt.Sprintf("Name: %s", locale.Name),
			fmt.Sprintf("Slug: %s", locale.Slug),
		})
	}
}

func newLocalesCreateCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new locale",
		RunE:  makeRunLocalesCreate(factory),
	}
	cmd.Flags().String("language", "", "Language code (required)")
	_ = cmd.MarkFlagRequired("language")
	cmd.Flags().String("region", "", "Region code (optional)")
	cmd.Flags().Bool("dry-run", false, "Preview without creating")
	return cmd
}

func makeRunLocalesCreate(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		language, _ := cmd.Flags().GetString("language")
		region, _ := cmd.Flags().GetString("region")

		input := map[string]any{"language": language}
		if region != "" {
			input["region"] = region
		}

		params := map[string]any{"input": input}

		if isDry, err := dryRunResult(cmd, fmt.Sprintf("create locale %s", language), params); isDry {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("createLocale", params)
		if err != nil {
			return fmt.Errorf("create locale: %w", err)
		}

		var locale Locale
		if err := json.Unmarshal(result, &locale); err != nil {
			return fmt.Errorf("parse locale: %w", err)
		}

		return cli.PrintResult(cmd, locale, []string{
			fmt.Sprintf("Created locale: %s (%s)", locale.Name, locale.Code),
		})
	}
}

// localeLanguage holds a language code and name.
type localeLanguage struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func newLocalesLanguagesCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "languages",
		Short: "List available locale languages",
		RunE:  makeRunLocalesLanguages(factory),
	}
	return cmd
}

func makeRunLocalesLanguages(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getLocaleLanguages", nil)
		if err != nil {
			return fmt.Errorf("list locale languages: %w", err)
		}

		var langs []localeLanguage
		if err := json.Unmarshal(result, &langs); err != nil {
			return fmt.Errorf("parse languages: %w", err)
		}

		lines := make([]string, 0, len(langs))
		for _, l := range langs {
			lines = append(lines, fmt.Sprintf("%s\t%s", l.Code, l.Name))
		}

		return cli.PrintResult(cmd, langs, lines)
	}
}

// localeRegion holds a region code, name, and whether it is commonly used.
type localeRegion struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsCommon bool   `json:"isCommon"`
}

func newLocalesRegionsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regions",
		Short: "List available regions for a language",
		RunE:  makeRunLocalesRegions(factory),
	}
	cmd.Flags().String("language", "", "Language code (required)")
	_ = cmd.MarkFlagRequired("language")
	return cmd
}

func makeRunLocalesRegions(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		language, _ := cmd.Flags().GetString("language")
		result, err := client.Call("getLocaleRegions", map[string]any{"language": language})
		if err != nil {
			return fmt.Errorf("list locale regions: %w", err)
		}

		var regions []localeRegion
		if err := json.Unmarshal(result, &regions); err != nil {
			return fmt.Errorf("parse regions: %w", err)
		}

		lines := make([]string, 0, len(regions))
		for _, r := range regions {
			common := ""
			if r.IsCommon {
				common = " (common)"
			}
			lines = append(lines, fmt.Sprintf("%s\t%s%s", r.Code, r.Name, common))
		}

		return cli.PrintResult(cmd, regions, lines)
	}
}

func newLocalesGroupsCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "Get localization groups",
		RunE:  makeRunLocalesGroups(factory),
	}
	return cmd
}

func makeRunLocalesGroups(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getLocalizationGroups", nil)
		if err != nil {
			return fmt.Errorf("get localization groups: %w", err)
		}

		// Result is a complex structure — return raw JSON as text summary
		return cli.PrintResult(cmd, result, []string{string(result)})
	}
}

func newLocalesSetDataCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-data",
		Short: "Set localization data",
		RunE:  makeRunLocalesSetData(factory),
	}
	cmd.Flags().String("data", "", "Localization data as JSON")
	cmd.Flags().String("data-file", "", "Path to JSON file with localization data")
	cmd.Flags().Bool("dry-run", false, "Preview without updating")
	return cmd
}

func makeRunLocalesSetData(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		dataStr, _ := cmd.Flags().GetString("data")
		dataFile, _ := cmd.Flags().GetString("data-file")

		data, err := parseJSONFlagOrFile(dataStr, dataFile)
		if err != nil {
			return fmt.Errorf("parse data: %w", err)
		}

		params := map[string]any{"data": data}

		if isDry, err := dryRunResult(cmd, "set localization data", params); isDry {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("setLocalizationData", params)
		if err != nil {
			return fmt.Errorf("set localization data: %w", err)
		}

		return cli.PrintResult(cmd, result, []string{"Localization data updated"})
	}
}
