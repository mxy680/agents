package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newRestCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rest",
		Short: "PostgREST configuration",
	}
	cmd.AddCommand(newRestGetCmd(factory), newRestUpdateCmd(factory))
	return cmd
}

// --- get ---

func newRestGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get PostgREST configuration for a project",
		RunE:  makeRunRestGet(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunRestGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/postgrest", ref), nil)
		if err != nil {
			return fmt.Errorf("getting PostgREST config: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(data, &raw); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var pretty bytes.Buffer
		if err := json.Indent(&pretty, data, "", "  "); err != nil {
			fmt.Println(string(data))
			return nil
		}
		fmt.Println(pretty.String())
		return nil
	}
}

// --- update ---

func newRestUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update PostgREST configuration for a project",
		RunE:  makeRunRestUpdate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("config", "", "Configuration JSON")
	cmd.Flags().String("config-file", "", "Path to configuration JSON file")
	cmd.Flags().Bool("dry-run", false, "Print what would happen without executing")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunRestUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		configStr, _ := cmd.Flags().GetString("config")
		configFile, _ := cmd.Flags().GetString("config-file")

		var bodyReader io.Reader
		if configFile != "" {
			fileData, err := os.ReadFile(configFile)
			if err != nil {
				return fmt.Errorf("read config file: %w", err)
			}
			bodyReader = bytes.NewReader(fileData)
		} else if configStr != "" {
			bodyReader = strings.NewReader(configStr)
		} else {
			return fmt.Errorf("either --config or --config-file is required")
		}

		if dryRunResult(cmd, fmt.Sprintf("Would update PostgREST config for project %s", ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodPatch, fmt.Sprintf("/projects/%s/postgrest", ref), bodyReader)
		if err != nil {
			return fmt.Errorf("updating PostgREST config: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(data, &raw); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var pretty bytes.Buffer
		if err := json.Indent(&pretty, data, "", "  "); err != nil {
			fmt.Println(string(data))
			return nil
		}
		fmt.Println(pretty.String())
		return nil
	}
}
