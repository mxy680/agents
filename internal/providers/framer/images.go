package framer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newImagesCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "images",
		Short: "Image upload commands",
	}
	cmd.AddCommand(
		newImagesUploadCmd(factory),
		newImagesUploadBatchCmd(factory),
	)
	return cmd
}

func newImagesUploadCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload an image",
		RunE:  makeRunImagesUpload(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().String("path", "", "Local path to the image file")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func makeRunImagesUpload(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		path, _ := cmd.Flags().GetString("path")

		params, err := buildImageParams(path)
		if err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("uploadImage", map[string]any{
			"image": params,
		})
		if err != nil {
			return fmt.Errorf("upload image: %w", err)
		}

		var raw json.RawMessage = result
		return cli.PrintResult(cmd, raw, []string{
			fmt.Sprintf("Uploaded: %s", filepath.Base(path)),
		})
	}
}

func newImagesUploadBatchCmd(factory BridgeClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload-batch",
		Short: "Upload multiple images",
		RunE:  makeRunImagesUploadBatch(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	cmd.Flags().String("paths", "", "Comma-separated local paths to image files")
	_ = cmd.MarkFlagRequired("paths")
	return cmd
}

func makeRunImagesUploadBatch(factory BridgeClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		pathsVal, _ := cmd.Flags().GetString("paths")
		paths := parseStringList(pathsVal)
		if len(paths) == 0 {
			return fmt.Errorf("--paths must specify at least one path")
		}

		images := make([]map[string]any, 0, len(paths))
		for _, p := range paths {
			params, err := buildImageParams(p)
			if err != nil {
				return err
			}
			images = append(images, params)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create framer client: %w", err)
		}
		defer client.Close()

		result, err := client.Call("uploadImages", map[string]any{
			"images": images,
		})
		if err != nil {
			return fmt.Errorf("upload images: %w", err)
		}

		var raw json.RawMessage = result
		lines := []string{fmt.Sprintf("Uploaded %d image(s)", len(paths))}
		return cli.PrintResult(cmd, raw, lines)
	}
}

// buildImageParams reads a file, base64-encodes it, and returns the image parameter map.
func buildImageParams(path string) (map[string]any, error) {
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

// mimeTypeFromExt returns the MIME type for a file extension.
func mimeTypeFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".pdf":
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}
