package framer

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newCodeCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code",
		Short: "Code file management commands",
	}
	cmd.AddCommand(
		newCodeListCmd(factory),
		newCodeGetCmd(factory),
		newCodeCreateCmd(factory),
		newCodeTypecheckCmd(factory),
		newCodeCustomGetCmd(factory),
		newCodeCustomSetCmd(factory),
	)
	return cmd
}

func newCodeListCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List code files",
		RunE:  makeRunCodeList(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunCodeList(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getCodeFiles", nil)
		if err != nil {
			return fmt.Errorf("list code files: %w", err)
		}

		var files []CodeFile
		if err := json.Unmarshal(result, &files); err != nil {
			return fmt.Errorf("parse code files: %w", err)
		}

		lines := make([]string, 0, len(files))
		for _, f := range files {
			lines = append(lines, fmt.Sprintf("%s  %s", f.ID, f.Name))
		}
		return cli.PrintResult(cmd, files, lines)
	}
}

func newCodeGetCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a code file by ID",
		RunE:  makeRunCodeGet(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().String("id", "", "Code file ID")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunCodeGet(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		id, _ := cmd.Flags().GetString("id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getCodeFile", map[string]any{
			"id": id,
		})
		if err != nil {
			return fmt.Errorf("get code file: %w", err)
		}

		// result may be null if not found
		if string(result) == "null" {
			return fmt.Errorf("code file %q not found", id)
		}

		var file CodeFile
		if err := json.Unmarshal(result, &file); err != nil {
			return fmt.Errorf("parse code file: %w", err)
		}

		return cli.PrintResult(cmd, file, []string{
			fmt.Sprintf("ID:   %s", file.ID),
			fmt.Sprintf("Name: %s", file.Name),
		})
	}
}

func newCodeCreateCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new code file",
		RunE:  makeRunCodeCreate(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("dry-run", false, "Preview without making changes")
	cmd.Flags().String("name", "", "Name for the new code file")
	cmd.Flags().String("code", "", "Source code content")
	cmd.Flags().String("code-file", "", "Path to file containing source code")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunCodeCreate(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		codeVal, _ := cmd.Flags().GetString("code")
		codeFile, _ := cmd.Flags().GetString("code-file")

		code, err := readFileOrFlag(codeVal, codeFile)
		if err != nil {
			return fmt.Errorf("--code: %w", err)
		}
		if code == "" {
			return fmt.Errorf("--code or --code-file is required")
		}

		params := map[string]any{
			"name": name,
			"code": code,
		}

		if isDry, result := dryRunResult(cmd, "create code file", params); isDry {
			return result
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("createCodeFile", params)
		if err != nil {
			return fmt.Errorf("create code file: %w", err)
		}

		var file CodeFile
		if err := json.Unmarshal(result, &file); err != nil {
			return fmt.Errorf("parse created code file: %w", err)
		}

		return cli.PrintResult(cmd, file, []string{
			fmt.Sprintf("Created code file: %s (%s)", file.Name, file.ID),
		})
	}
}

func newCodeTypecheckCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "typecheck",
		Short: "Typecheck code content",
		RunE:  makeRunCodeTypecheck(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().String("name", "", "Name of the code file to typecheck")
	cmd.Flags().String("content", "", "Code content to typecheck")
	cmd.Flags().String("content-file", "", "Path to file containing code to typecheck")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunCodeTypecheck(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		contentVal, _ := cmd.Flags().GetString("content")
		contentFile, _ := cmd.Flags().GetString("content-file")

		content, err := readFileOrFlag(contentVal, contentFile)
		if err != nil {
			return fmt.Errorf("--content: %w", err)
		}
		if content == "" {
			return fmt.Errorf("--content or --content-file is required")
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("typecheckCode", map[string]any{
			"name":    name,
			"content": content,
		})
		if err != nil {
			return fmt.Errorf("typecheck code: %w", err)
		}

		var diagnostics []json.RawMessage
		if err := json.Unmarshal(result, &diagnostics); err != nil {
			return fmt.Errorf("parse diagnostics: %w", err)
		}

		lines := []string{fmt.Sprintf("Diagnostics (%d):", len(diagnostics))}
		for i, d := range diagnostics {
			lines = append(lines, fmt.Sprintf("  [%d] %s", i+1, truncate(string(d), 120)))
		}
		return cli.PrintResult(cmd, diagnostics, lines)
	}
}

func newCodeCustomGetCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "custom-get",
		Short: "Get project-level custom code",
		RunE:  makeRunCodeCustomGet(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunCodeCustomGet(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("getCustomCode", nil)
		if err != nil {
			return fmt.Errorf("get custom code: %w", err)
		}

		var cc CustomCode
		if err := json.Unmarshal(result, &cc); err != nil {
			return fmt.Errorf("parse custom code: %w", err)
		}

		lines := []string{
			fmt.Sprintf("HeadStart: %s", truncate(cc.HeadStart, 80)),
			fmt.Sprintf("HeadEnd:   %s", truncate(cc.HeadEnd, 80)),
			fmt.Sprintf("BodyStart: %s", truncate(cc.BodyStart, 80)),
			fmt.Sprintf("BodyEnd:   %s", truncate(cc.BodyEnd, 80)),
		}
		return cli.PrintResult(cmd, cc, lines)
	}
}

func newCodeCustomSetCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "custom-set",
		Short: "Set project-level custom code",
		RunE:  makeRunCodeCustomSet(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().Bool("dry-run", false, "Preview without making changes")
	cmd.Flags().String("html", "", "HTML/JS snippet to inject")
	cmd.Flags().String("location", "", "Injection location: headStart|headEnd|bodyStart|bodyEnd")
	_ = cmd.MarkFlagRequired("html")
	_ = cmd.MarkFlagRequired("location")
	return cmd
}

func makeRunCodeCustomSet(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		html, _ := cmd.Flags().GetString("html")
		location, _ := cmd.Flags().GetString("location")

		validLocations := map[string]bool{
			"headStart": true,
			"headEnd":   true,
			"bodyStart": true,
			"bodyEnd":   true,
		}
		if !validLocations[location] {
			return fmt.Errorf("--location must be one of: headStart, headEnd, bodyStart, bodyEnd")
		}

		params := map[string]any{
			"html":     html,
			"location": location,
		}

		if isDry, result := dryRunResult(cmd, "set custom code", params); isDry {
			return result
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("setCustomCode", params)
		if err != nil {
			return fmt.Errorf("set custom code: %w", err)
		}

		var cc CustomCode
		if err := json.Unmarshal(result, &cc); err != nil {
			return fmt.Errorf("parse custom code result: %w", err)
		}

		return cli.PrintResult(cmd, cc, []string{
			fmt.Sprintf("Custom code set at %s", location),
		})
	}
}
