package imessage

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newAttachmentsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "attachments",
		Short:   "Manage message attachments",
		Aliases: []string{"attach"},
	}

	cmd.AddCommand(newAttachmentsGetCmd(factory))
	cmd.AddCommand(newAttachmentsDownloadCmd(factory))
	cmd.AddCommand(newAttachmentsDownloadForceCmd(factory))
	cmd.AddCommand(newAttachmentsUploadCmd(factory))
	cmd.AddCommand(newAttachmentsLiveCmd(factory))
	cmd.AddCommand(newAttachmentsBlurhashCmd(factory))
	cmd.AddCommand(newAttachmentsCountCmd(factory))

	return cmd
}

func newAttachmentsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get attachment details by GUID",
		RunE:  makeRunAttachmentsGet(factory),
	}
	cmd.Flags().String("guid", "", "Attachment GUID (required)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("guid")
	return cmd
}

func makeRunAttachmentsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Get(ctx, fmt.Sprintf("/attachment/%s", guid), nil)
		if err != nil {
			return fmt.Errorf("getting attachment %s: %w", guid, err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		summary := toAttachmentSummary(data)
		return printResult(cmd, summary, []string{
			fmt.Sprintf("GUID:      %s", summary.GUID),
			fmt.Sprintf("File:      %s", summary.FileName),
			fmt.Sprintf("MIME Type: %s", summary.MIMEType),
			fmt.Sprintf("Size:      %d bytes", summary.TotalBytes),
			fmt.Sprintf("Outgoing:  %v", summary.IsOutgoing),
			fmt.Sprintf("Created:   %s", formatTimestamp(summary.DateCreated)),
		})
	}
}

func newAttachmentsDownloadCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download an attachment to a file",
		RunE:  makeRunAttachmentsDownload(factory),
	}
	cmd.Flags().String("guid", "", "Attachment GUID (required)")
	cmd.Flags().String("output", "", "Output file path (required)")
	_ = cmd.MarkFlagRequired("guid")
	_ = cmd.MarkFlagRequired("output")
	return cmd
}

func makeRunAttachmentsDownload(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")
		output, _ := cmd.Flags().GetString("output")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		f, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("create output file %q: %w", output, err)
		}
		defer f.Close()

		if err := client.Download(ctx, fmt.Sprintf("/attachment/%s/download", guid), f); err != nil {
			return fmt.Errorf("downloading attachment %s: %w", guid, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Downloaded attachment %s to %s\n", guid, output)
		return nil
	}
}

func newAttachmentsDownloadForceCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download-force",
		Short: "Force-download an attachment (re-fetch from Apple servers)",
		RunE:  makeRunAttachmentsDownloadForce(factory),
	}
	cmd.Flags().String("guid", "", "Attachment GUID (required)")
	cmd.Flags().String("output", "", "Output file path (required)")
	_ = cmd.MarkFlagRequired("guid")
	_ = cmd.MarkFlagRequired("output")
	return cmd
}

func makeRunAttachmentsDownloadForce(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")
		output, _ := cmd.Flags().GetString("output")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		f, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("create output file %q: %w", output, err)
		}
		defer f.Close()

		if err := client.Download(ctx, fmt.Sprintf("/attachment/%s/download/force", guid), f); err != nil {
			return fmt.Errorf("force-downloading attachment %s: %w", guid, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Force-downloaded attachment %s to %s\n", guid, output)
		return nil
	}
}

func newAttachmentsUploadCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a file as an attachment",
		RunE:  makeRunAttachmentsUpload(factory),
	}
	cmd.Flags().String("path", "", "Path to the file to upload (required)")
	cmd.Flags().Bool("dry-run", false, "Show what would be uploaded without sending")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func makeRunAttachmentsUpload(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		path, _ := cmd.Flags().GetString("path")

		if cli.IsDryRun(cmd) {
			result := dryRunResult("upload", map[string]any{"path": path})
			return printResult(cmd, result, []string{
				fmt.Sprintf("[dry-run] Would upload file: %s", path),
			})
		}

		// TODO: implement multipart file upload to /attachment/upload
		return fmt.Errorf("attachment upload not yet implemented (use --dry-run to preview)")
	}
}

func newAttachmentsLiveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "live",
		Short: "Download the Live Photo component of an attachment",
		RunE:  makeRunAttachmentsLive(factory),
	}
	cmd.Flags().String("guid", "", "Attachment GUID (required)")
	cmd.Flags().String("output", "", "Output file path (required)")
	_ = cmd.MarkFlagRequired("guid")
	_ = cmd.MarkFlagRequired("output")
	return cmd
}

func makeRunAttachmentsLive(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")
		output, _ := cmd.Flags().GetString("output")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		f, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("create output file %q: %w", output, err)
		}
		defer f.Close()

		if err := client.Download(ctx, fmt.Sprintf("/attachment/%s/live", guid), f); err != nil {
			return fmt.Errorf("downloading live photo for attachment %s: %w", guid, err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Downloaded live photo for attachment %s to %s\n", guid, output)
		return nil
	}
}

func newAttachmentsBlurhashCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blurhash",
		Short: "Get the blurhash preview for an attachment",
		RunE:  makeRunAttachmentsBlurhash(factory),
	}
	cmd.Flags().String("guid", "", "Attachment GUID (required)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("guid")
	return cmd
}

func makeRunAttachmentsBlurhash(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		guid, _ := cmd.Flags().GetString("guid")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Get(ctx, fmt.Sprintf("/attachment/%s/blurhash", guid), nil)
		if err != nil {
			return fmt.Errorf("getting blurhash for attachment %s: %w", guid, err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parse blurhash response: %w", err)
		}

		hash := getString(raw, "blurhash")
		return printResult(cmd, raw, []string{
			fmt.Sprintf("GUID:     %s", guid),
			fmt.Sprintf("Blurhash: %s", hash),
		})
	}
}

func newAttachmentsCountCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count",
		Short: "Get the total number of attachments",
		RunE:  makeRunAttachmentsCount(factory),
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunAttachmentsCount(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body, err := client.Get(ctx, "/attachment/count", nil)
		if err != nil {
			return fmt.Errorf("getting attachment count: %w", err)
		}

		data, err := ParseResponse(body)
		if err != nil {
			return err
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parse count response: %w", err)
		}

		count := getInt64(raw, "total")
		return printResult(cmd, raw, []string{
			fmt.Sprintf("Total attachments: %d", count),
		})
	}
}
