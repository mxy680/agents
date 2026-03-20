package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newDatabaseCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Database management",
	}
	cmd.AddCommand(
		newDBMigrationsCmd(factory),
		newDBTypesCmd(factory),
		newDBSSLCmd(factory),
		newDBJITCmd(factory),
	)
	return cmd
}

func newDBSSLCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ssl-enforcement",
		Aliases: []string{"ssl"},
		Short:   "SSL enforcement settings",
	}
	cmd.AddCommand(newDBSSLGetCmd(factory), newDBSSLUpdateCmd(factory))
	return cmd
}

func newDBJITCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jit-access",
		Aliases: []string{"jit"},
		Short:   "JIT access settings",
	}
	cmd.AddCommand(newDBJITGetCmd(factory), newDBJITUpdateCmd(factory))
	return cmd
}

// --- migrations ---

func newDBMigrationsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrations",
		Short: "List database migrations",
		RunE:  makeRunDBMigrations(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunDBMigrations(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/database/migrations", ref), nil)
		if err != nil {
			return fmt.Errorf("listing migrations for project %q: %w", ref, err)
		}

		var data []map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing migrations response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		if len(data) == 0 {
			fmt.Println("No migrations found.")
			return nil
		}
		lines := make([]string, 0, len(data)+1)
		lines = append(lines, fmt.Sprintf("%-30s  %-20s  %s", "VERSION", "NAME", "STATEMENTS"))
		for _, m := range data {
			version, _ := m["version"].(string)
			name, _ := m["name"].(string)
			stmts, _ := m["statements"].([]any)
			lines = append(lines, fmt.Sprintf("%-30s  %-20s  %d",
				truncate(version, 30), truncate(name, 20), len(stmts)))
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- types ---

func newDBTypesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "types",
		Short: "Get TypeScript type definitions for the project",
		RunE:  makeRunDBTypes(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("lang", "typescript", "Language for type definitions (default: typescript)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunDBTypes(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/types/typescript", ref), nil)
		if err != nil {
			return fmt.Errorf("getting types for project %q: %w", ref, err)
		}

		// The API returns raw type definitions as text (may be JSON or plain text).
		// Try to parse as JSON; if it fails, print raw.
		if cli.IsJSONOutput(cmd) {
			var parsed any
			if jsonErr := json.Unmarshal(raw, &parsed); jsonErr == nil {
				return cli.PrintJSON(parsed)
			}
			return cli.PrintJSON(map[string]string{"types": string(raw)})
		}

		fmt.Print(string(raw))
		return nil
	}
}

// --- ssl-enforcement get ---

func newDBSSLGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get SSL enforcement settings",
		RunE:  makeRunDBSSLGet(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunDBSSLGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/ssl-enforcement", ref), nil)
		if err != nil {
			return fmt.Errorf("getting SSL enforcement for project %q: %w", ref, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing SSL enforcement response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		enforced, _ := data["enforced"].(bool)
		lines := []string{
			fmt.Sprintf("Enforced: %v", enforced),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- ssl-enforcement update ---

func newDBSSLUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update SSL enforcement settings",
		RunE:  makeRunDBSSLUpdate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().Bool("enabled", false, "Enable SSL enforcement")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("enabled")
	return cmd
}

func makeRunDBSSLUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		enabled, _ := cmd.Flags().GetBool("enabled")

		if dryRunResult(cmd, fmt.Sprintf("Would set SSL enforcement to %v for project %q", enabled, ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		bodyMap := map[string]any{"enforced": enabled}
		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return fmt.Errorf("encoding request body: %w", err)
		}

		raw, err := doSupabase(client, http.MethodPut, fmt.Sprintf("/projects/%s/ssl-enforcement", ref), bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("updating SSL enforcement for project %q: %w", ref, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing SSL enforcement response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		enforced, _ := data["enforced"].(bool)
		fmt.Printf("SSL enforcement updated: enforced=%v\n", enforced)
		return nil
	}
}

// --- jit-access get ---

func newDBJITGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get JIT access settings",
		RunE:  makeRunDBJITGet(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunDBJITGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/jit-access", ref), nil)
		if err != nil {
			return fmt.Errorf("getting JIT access for project %q: %w", ref, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing JIT access response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		pretty, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(pretty))
		return nil
	}
}

// --- jit-access update ---

func newDBJITUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update JIT access settings",
		RunE:  makeRunDBJITUpdate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("config", "", "JIT access config as JSON")
	cmd.Flags().String("config-file", "", "Path to JIT access config JSON file")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunDBJITUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		configStr, _ := cmd.Flags().GetString("config")
		configFile, _ := cmd.Flags().GetString("config-file")

		if dryRunResult(cmd, fmt.Sprintf("Would update JIT access config for project %q", ref)) {
			return nil
		}

		var bodyReader *strings.Reader
		if configFile != "" {
			data, err := os.ReadFile(configFile)
			if err != nil {
				return fmt.Errorf("read config file: %w", err)
			}
			bodyReader = strings.NewReader(string(data))
		} else if configStr != "" {
			bodyReader = strings.NewReader(configStr)
		} else {
			return fmt.Errorf("either --config or --config-file is required")
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		raw, err := doSupabase(client, http.MethodPut, fmt.Sprintf("/projects/%s/jit-access", ref), bodyReader)
		if err != nil {
			return fmt.Errorf("updating JIT access for project %q: %w", ref, err)
		}

		var data map[string]any
		if err := json.Unmarshal(raw, &data); err != nil {
			return fmt.Errorf("parsing JIT access response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(data)
		}

		pretty, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(pretty))
		return nil
	}
}
