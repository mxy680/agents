package framer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newFilesCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "files",
		Short: "File upload commands",
	}
	cmd.AddCommand(
		newFilesUploadCmd(factory),
		newFilesUploadBatchCmd(factory),
	)
	return cmd
}

func newFilesUploadCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a file",
		RunE:  makeRunFilesUpload(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().String("path", "", "Local path to the file")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func makeRunFilesUpload(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		path, _ := cmd.Flags().GetString("path")

		params, err := buildFileParams(path)
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("uploadFile", map[string]any{
			"file": params,
		})
		if err != nil {
			return fmt.Errorf("upload file: %w", err)
		}

		var raw json.RawMessage = result
		return cli.PrintResult(cmd, raw, []string{
			fmt.Sprintf("Uploaded: %s", filepath.Base(path)),
		})
	}
}

func newFilesUploadBatchCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload-batch",
		Short: "Upload multiple files",
		RunE:  makeRunFilesUploadBatch(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().String("paths", "", "Comma-separated local paths to files")
	_ = cmd.MarkFlagRequired("paths")
	return cmd
}

func makeRunFilesUploadBatch(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		pathsVal, _ := cmd.Flags().GetString("paths")
		paths := parseStringList(pathsVal)
		if len(paths) == 0 {
			return fmt.Errorf("--paths must specify at least one path")
		}

		files := make([]map[string]any, 0, len(paths))
		for _, p := range paths {
			params, err := buildFileParams(p)
			if err != nil {
				return err
			}
			files = append(files, params)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("uploadFiles", map[string]any{
			"files": files,
		})
		if err != nil {
			return fmt.Errorf("upload files: %w", err)
		}

		var raw json.RawMessage = result
		lines := []string{fmt.Sprintf("Uploaded %d file(s)", len(paths))}
		return cli.PrintResult(cmd, raw, lines)
	}
}

// buildFileParams reads a file, base64-encodes it, and returns the file parameter map.
func buildFileParams(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", path, err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return map[string]any{
		"data": encoded,
		"name": filepath.Base(path),
		"type": mimeTypeFromExt(filepath.Ext(path)),
	}, nil
}
