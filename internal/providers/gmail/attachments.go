package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// AttachmentInfo is the JSON-serializable metadata for a Gmail message attachment.
type AttachmentInfo struct {
	AttachmentID string `json:"attachmentId"`
	Size         int64  `json:"size"`
	Filename     string `json:"filename,omitempty"`
}

// newAttachmentsGetCmd returns the `attachments get` command.
func newAttachmentsGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Download a message attachment",
		RunE:  makeRunAttachmentsGet(factory),
	}
	cmd.Flags().String("message-id", "", "Message ID containing the attachment (required)")
	cmd.Flags().String("attachment-id", "", "Attachment ID to download (required)")
	cmd.Flags().String("output", "", "Write attachment bytes to this file path (optional; defaults to stdout)")
	_ = cmd.MarkFlagRequired("message-id")
	_ = cmd.MarkFlagRequired("attachment-id")
	return cmd
}

func makeRunAttachmentsGet(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		messageID, _ := cmd.Flags().GetString("message-id")
		attachmentID, _ := cmd.Flags().GetString("attachment-id")
		outputPath, _ := cmd.Flags().GetString("output")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		att, err := svc.Users.Messages.Attachments.Get("me", messageID, attachmentID).Do()
		if err != nil {
			return fmt.Errorf("getting attachment %s from message %s: %w", attachmentID, messageID, err)
		}

		// The API returns base64url-encoded data.
		data, err := base64.URLEncoding.DecodeString(att.Data)
		if err != nil {
			// Some responses use standard base64 — fall back.
			data, err = base64.StdEncoding.DecodeString(att.Data)
			if err != nil {
				return fmt.Errorf("decoding attachment data: %w", err)
			}
		}

		if cli.IsJSONOutput(cmd) {
			info := AttachmentInfo{
				AttachmentID: att.AttachmentId,
				Size:         att.Size,
			}
			return cli.PrintJSON(info)
		}

		if outputPath != "" {
			if err := os.WriteFile(outputPath, data, 0o644); err != nil {
				return fmt.Errorf("writing attachment to %s: %w", outputPath, err)
			}
			fmt.Fprintf(os.Stderr, "Attachment written to %s (%d bytes)\n", outputPath, len(data))
			return nil
		}

		// Text mode with no --output: write raw bytes to stdout for piping.
		_, err = os.Stdout.Write(data)
		return err
	}
}
