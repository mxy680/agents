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

func newAuthCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage Supabase Auth configuration",
	}
	cmd.AddCommand(
		newAuthGetCmd(factory),
		newAuthUpdateCmd(factory),
		newAuthSigningKeysCmd(factory),
		newAuthThirdPartyCmd(factory),
	)
	return cmd
}

// readConfigBody reads config JSON from --config or --config-file flags.
func readConfigBody(cmd *cobra.Command) (*bytes.Reader, error) {
	configStr, _ := cmd.Flags().GetString("config")
	configFile, _ := cmd.Flags().GetString("config-file")

	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
		return bytes.NewReader(data), nil
	}
	if configStr != "" {
		return bytes.NewReader([]byte(configStr)), nil
	}
	return nil, fmt.Errorf("either --config or --config-file is required")
}

// printAuthConfigJSON pretty-prints raw auth config JSON.
func printAuthConfigJSON(cmd *cobra.Command, data []byte) error {
	if cli.IsJSONOutput(cmd) {
		// Print raw JSON as-is for --json mode
		var v any
		if err := json.Unmarshal(data, &v); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
		return cli.PrintJSON(v)
	}
	// Pretty-print for text mode
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}
	pretty, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("formatting response: %w", err)
	}
	fmt.Println(string(pretty))
	return nil
}

// --- auth get ---

func newAuthGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get the Auth configuration for a project",
		RunE:  makeRunAuthGet(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunAuthGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/config/auth", ref), nil)
		if err != nil {
			return fmt.Errorf("getting auth config: %w", err)
		}

		return printAuthConfigJSON(cmd, data)
	}
}

// --- auth update ---

func newAuthUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the Auth configuration for a project",
		RunE:  makeRunAuthUpdate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("config", "", "Auth config JSON")
	cmd.Flags().String("config-file", "", "Path to auth config JSON file")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunAuthUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		if cli.IsDryRun(cmd) {
			if cli.IsJSONOutput(cmd) {
				_ = cli.PrintJSON(map[string]string{"dryRun": fmt.Sprintf("Would update auth config for project %s", ref)})
			} else {
				fmt.Printf("[DRY RUN] Would update auth config for project %s\n", ref)
			}
			return nil
		}

		bodyReader, err := readConfigBody(cmd)
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodPatch, fmt.Sprintf("/projects/%s/config/auth", ref), bodyReader)
		if err != nil {
			return fmt.Errorf("updating auth config: %w", err)
		}

		return printAuthConfigJSON(cmd, data)
	}
}

// --- signing-keys ---

func newAuthSigningKeysCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "signing-keys",
		Aliases: []string{"sk"},
		Short:   "Manage auth signing keys",
	}
	cmd.AddCommand(
		newSigningKeysListCmd(factory),
		newSigningKeysGetCmd(factory),
		newSigningKeysCreateCmd(factory),
		newSigningKeysUpdateCmd(factory),
		newSigningKeysDeleteCmd(factory),
	)
	return cmd
}

func newSigningKeysListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List auth signing keys for a project",
		RunE:  makeRunSigningKeysList(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunSigningKeysList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/config/auth/signing-keys", ref), nil)
		if err != nil {
			return fmt.Errorf("listing signing keys: %w", err)
		}

		return printAuthConfigJSON(cmd, data)
	}
}

func newSigningKeysGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a specific auth signing key",
		RunE:  makeRunSigningKeysGet(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("key-id", "", "Signing key ID (required)")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("key-id")
	return cmd
}

func makeRunSigningKeysGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		keyID, _ := cmd.Flags().GetString("key-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/config/auth/signing-keys/%s", ref, keyID), nil)
		if err != nil {
			return fmt.Errorf("getting signing key %s: %w", keyID, err)
		}

		return printAuthConfigJSON(cmd, data)
	}
}

func newSigningKeysCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new auth signing key",
		RunE:  makeRunSigningKeysCreate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunSigningKeysCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		if dryRunResult(cmd, fmt.Sprintf("Would create a new signing key in project %s", ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		// POST with empty body
		data, err := doSupabase(client, http.MethodPost, fmt.Sprintf("/projects/%s/config/auth/signing-keys", ref), strings.NewReader("{}"))
		if err != nil {
			return fmt.Errorf("creating signing key: %w", err)
		}

		return printAuthConfigJSON(cmd, data)
	}
}

func newSigningKeysUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an auth signing key",
		RunE:  makeRunSigningKeysUpdate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("key-id", "", "Signing key ID (required)")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("key-id")
	return cmd
}

func makeRunSigningKeysUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		keyID, _ := cmd.Flags().GetString("key-id")

		if dryRunResult(cmd, fmt.Sprintf("Would update signing key %s in project %s", keyID, ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodPatch, fmt.Sprintf("/projects/%s/config/auth/signing-keys/%s", ref, keyID), strings.NewReader("{}"))
		if err != nil {
			return fmt.Errorf("updating signing key %s: %w", keyID, err)
		}

		return printAuthConfigJSON(cmd, data)
	}
}

func newSigningKeysDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an auth signing key (irreversible)",
		RunE:  makeRunSigningKeysDelete(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("key-id", "", "Signing key ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("key-id")
	return cmd
}

func makeRunSigningKeysDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		keyID, _ := cmd.Flags().GetString("key-id")

		if dryRunResult(cmd, fmt.Sprintf("Would permanently delete signing key %s from project %s", keyID, ref)) {
			return nil
		}

		if err := confirmDestructive(cmd, fmt.Sprintf("deleting signing key %s is irreversible", keyID)); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if _, err := doSupabase(client, http.MethodDelete, fmt.Sprintf("/projects/%s/config/auth/signing-keys/%s", ref, keyID), nil); err != nil {
			return fmt.Errorf("deleting signing key %s: %w", keyID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "keyId": keyID})
		}
		fmt.Printf("Deleted signing key: %s\n", keyID)
		return nil
	}
}

// --- third-party ---

func newAuthThirdPartyCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "third-party",
		Aliases: []string{"tpa"},
		Short:   "Manage third-party auth providers",
	}
	cmd.AddCommand(
		newThirdPartyListCmd(factory),
		newThirdPartyGetCmd(factory),
		newThirdPartyCreateCmd(factory),
		newThirdPartyDeleteCmd(factory),
	)
	return cmd
}

func newThirdPartyListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List third-party auth providers for a project",
		RunE:  makeRunThirdPartyList(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunThirdPartyList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/config/auth/third-party-auth", ref), nil)
		if err != nil {
			return fmt.Errorf("listing third-party auth providers: %w", err)
		}

		return printAuthConfigJSON(cmd, data)
	}
}

func newThirdPartyGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a specific third-party auth provider",
		RunE:  makeRunThirdPartyGet(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("tpa-id", "", "Third-party auth provider ID (required)")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("tpa-id")
	return cmd
}

func makeRunThirdPartyGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		tpaID, _ := cmd.Flags().GetString("tpa-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/config/auth/third-party-auth/%s", ref, tpaID), nil)
		if err != nil {
			return fmt.Errorf("getting third-party auth provider %s: %w", tpaID, err)
		}

		return printAuthConfigJSON(cmd, data)
	}
}

func newThirdPartyCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a third-party auth provider",
		RunE:  makeRunThirdPartyCreate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("config", "", "Provider config JSON")
	cmd.Flags().String("config-file", "", "Path to provider config JSON file")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunThirdPartyCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		if cli.IsDryRun(cmd) {
			if cli.IsJSONOutput(cmd) {
				_ = cli.PrintJSON(map[string]string{"dryRun": fmt.Sprintf("Would create third-party auth provider in project %s", ref)})
			} else {
				fmt.Printf("[DRY RUN] Would create third-party auth provider in project %s\n", ref)
			}
			return nil
		}

		bodyReader, err := readConfigBody(cmd)
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodPost, fmt.Sprintf("/projects/%s/config/auth/third-party-auth", ref), bodyReader)
		if err != nil {
			return fmt.Errorf("creating third-party auth provider: %w", err)
		}

		return printAuthConfigJSON(cmd, data)
	}
}

func newThirdPartyDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a third-party auth provider (irreversible)",
		RunE:  makeRunThirdPartyDelete(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("tpa-id", "", "Third-party auth provider ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("tpa-id")
	return cmd
}

func makeRunThirdPartyDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		tpaID, _ := cmd.Flags().GetString("tpa-id")

		if dryRunResult(cmd, fmt.Sprintf("Would permanently delete third-party auth provider %s from project %s", tpaID, ref)) {
			return nil
		}

		if err := confirmDestructive(cmd, fmt.Sprintf("deleting third-party auth provider %s is irreversible", tpaID)); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if _, err := doSupabase(client, http.MethodDelete, fmt.Sprintf("/projects/%s/config/auth/third-party-auth/%s", ref, tpaID), nil); err != nil {
			return fmt.Errorf("deleting third-party auth provider %s: %w", tpaID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "tpaId": tpaID})
		}
		fmt.Printf("Deleted third-party auth provider: %s\n", tpaID)
		return nil
	}
}
