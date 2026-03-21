package x

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// MediaUploadResult holds the result of a media upload operation.
type MediaUploadResult struct {
	MediaID       string `json:"media_id"`
	MediaIDString string `json:"media_id_string"`
	State         string `json:"state,omitempty"`
}

// newMediaCmd builds the "media" subcommand group.
func newMediaCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "media",
		Short:   "Upload and manage media",
		Aliases: []string{"upload"},
	}
	cmd.AddCommand(newMediaUploadCmd(factory))
	cmd.AddCommand(newMediaStatusCmd(factory))
	cmd.AddCommand(newMediaSetAltTextCmd(factory))
	return cmd
}

func newMediaUploadCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a media file",
		RunE:  makeRunMediaUpload(factory),
	}
	cmd.Flags().String("path", "", "Path to the media file (required)")
	_ = cmd.MarkFlagRequired("path")
	cmd.Flags().String("alt-text", "", "Alt text / accessibility description for the media")
	return cmd
}

func newMediaStatusCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check media upload status",
		RunE:  makeRunMediaStatus(factory),
	}
	cmd.Flags().String("media-id", "", "Media ID (required)")
	_ = cmd.MarkFlagRequired("media-id")
	return cmd
}

func newMediaSetAltTextCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-alt-text",
		Short: "Set alt text for an uploaded media",
		RunE:  makeRunMediaSetAltText(factory),
	}
	cmd.Flags().String("media-id", "", "Media ID (required)")
	_ = cmd.MarkFlagRequired("media-id")
	cmd.Flags().String("alt-text", "", "Alt text description (required)")
	_ = cmd.MarkFlagRequired("alt-text")
	cmd.Flags().Bool("dry-run", false, "Print what would be set without setting")
	return cmd
}

// --- RunE implementations ---

func makeRunMediaUpload(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		filePath, _ := cmd.Flags().GetString("path")
		altText, _ := cmd.Flags().GetString("alt-text")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read file %s: %w", filePath, err)
		}

		totalBytes := len(fileData)
		mimeType := detectMIMEType(filePath, fileData)
		mediaCategory := mimeToMediaCategory(mimeType)

		// INIT phase.
		initForm := url.Values{}
		initForm.Set("command", "INIT")
		initForm.Set("total_bytes", strconv.Itoa(totalBytes))
		initForm.Set("media_type", mimeType)
		initForm.Set("media_category", mediaCategory)

		var initResult struct {
			MediaID       int64  `json:"media_id"`
			MediaIDString string `json:"media_id_string"`
		}
		if err := client.doUploadForm(ctx, "/i/media/upload.json", initForm, &initResult); err != nil {
			return fmt.Errorf("media upload INIT: %w", err)
		}

		mediaIDStr := initResult.MediaIDString
		if mediaIDStr == "" {
			mediaIDStr = strconv.FormatInt(initResult.MediaID, 10)
		}

		// APPEND phase (single chunk).
		if err := client.appendMediaChunk(ctx, mediaIDStr, fileData); err != nil {
			return fmt.Errorf("media upload APPEND: %w", err)
		}

		// FINALIZE phase.
		finalForm := url.Values{}
		finalForm.Set("command", "FINALIZE")
		finalForm.Set("media_id", mediaIDStr)

		var finalResult struct {
			MediaID        int64  `json:"media_id"`
			MediaIDString  string `json:"media_id_string"`
			ProcessingInfo *struct {
				State string `json:"state"`
			} `json:"processing_info"`
		}
		if err := client.doUploadForm(ctx, "/i/media/upload.json", finalForm, &finalResult); err != nil {
			return fmt.Errorf("media upload FINALIZE: %w", err)
		}

		state := "succeeded"
		if finalResult.ProcessingInfo != nil && finalResult.ProcessingInfo.State != "" {
			state = finalResult.ProcessingInfo.State
		}

		result := MediaUploadResult{
			MediaIDString: mediaIDStr,
			State:         state,
		}
		if finalResult.MediaIDString != "" {
			result.MediaIDString = finalResult.MediaIDString
		}
		result.MediaID = result.MediaIDString

		// Optionally set alt text.
		if altText != "" {
			if setErr := client.setMediaAltText(ctx, result.MediaIDString, altText); setErr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to set alt text: %v\n", setErr)
			}
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Media uploaded: %s (state: %s)\n", result.MediaIDString, result.State)
		return nil
	}
}

func makeRunMediaStatus(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("command", "STATUS")
		params.Set("media_id", mediaID)

		var statusResult struct {
			MediaID        int64  `json:"media_id"`
			MediaIDString  string `json:"media_id_string"`
			ProcessingInfo *struct {
				State           string `json:"state"`
				CheckAfterSecs  int    `json:"check_after_secs"`
				ProgressPercent int    `json:"progress_percent"`
			} `json:"processing_info"`
		}

		fullURL := client.uploadURL + "/i/media/upload.json?" + params.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
		if err != nil {
			return fmt.Errorf("build media status request: %w", err)
		}
		client.applyHeaders(req)

		resp, err := client.http.Do(req)
		if err != nil {
			return fmt.Errorf("media status: %w", err)
		}
		client.captureResponseHeaders(resp)

		if err := client.DecodeJSON(resp, &statusResult); err != nil {
			return fmt.Errorf("decode status response: %w", err)
		}

		state := "unknown"
		if statusResult.ProcessingInfo != nil {
			state = statusResult.ProcessingInfo.State
		}

		result := MediaUploadResult{
			MediaIDString: statusResult.MediaIDString,
			State:         state,
		}
		if result.MediaIDString == "" {
			result.MediaIDString = mediaID
		}
		result.MediaID = result.MediaIDString

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Media %s: state=%s\n", result.MediaIDString, result.State)
		return nil
	}
}

func makeRunMediaSetAltText(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")
		altText, _ := cmd.Flags().GetString("alt-text")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("set alt text for media %s", mediaID), map[string]string{
				"media_id": mediaID,
				"alt_text": altText,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if err := client.setMediaAltText(ctx, mediaID, altText); err != nil {
			return fmt.Errorf("setting alt text for media %s: %w", mediaID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "updated", "media_id": mediaID})
		}
		fmt.Printf("Alt text set for media: %s\n", mediaID)
		return nil
	}
}

// --- Client helper methods ---

// doUploadForm sends a POST form request to upload.x.com and decodes the JSON response into target.
func (c *Client) doUploadForm(ctx context.Context, path string, form url.Values, target any) error {
	fullURL := c.uploadURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build upload request: %w", err)
	}
	c.applyHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("upload request: %w", err)
	}
	c.captureResponseHeaders(resp)

	return c.DecodeJSON(resp, target)
}

// appendMediaChunk sends the APPEND phase request for a media upload.
func (c *Client) appendMediaChunk(ctx context.Context, mediaID string, data []byte) error {
	fullURL := c.uploadURL + "/i/media/upload.json"

	var buf bytes.Buffer
	boundary := "xmediaboundary"
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString("Content-Disposition: form-data; name=\"command\"\r\n\r\n")
	buf.WriteString("APPEND\r\n")
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString("Content-Disposition: form-data; name=\"media_id\"\r\n\r\n")
	buf.WriteString(mediaID + "\r\n")
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString("Content-Disposition: form-data; name=\"segment_index\"\r\n\r\n")
	buf.WriteString("0\r\n")
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString("Content-Disposition: form-data; name=\"media\"; filename=\"media\"\r\n")
	buf.WriteString("Content-Type: application/octet-stream\r\n\r\n")
	buf.Write(data)
	buf.WriteString("\r\n--" + boundary + "--\r\n")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, &buf)
	if err != nil {
		return fmt.Errorf("build append request: %w", err)
	}
	c.applyHeaders(req)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("append request: %w", err)
	}
	c.captureResponseHeaders(resp)
	resp.Body.Close()

	// APPEND returns 204 No Content on success, or 2xx.
	if resp.StatusCode >= 400 {
		return fmt.Errorf("media APPEND failed (http=%d)", resp.StatusCode)
	}
	return nil
}

// setMediaAltText calls api.x.com to set alt text metadata on an uploaded media.
func (c *Client) setMediaAltText(ctx context.Context, mediaID, altText string) error {
	body := map[string]any{
		"media_id": mediaID,
		"alt_text": map[string]string{
			"text": altText,
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal alt text body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/1.1/media/metadata/create.json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("build alt text request: %w", err)
	}
	c.applyHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("alt text request: %w", err)
	}
	c.captureResponseHeaders(resp)
	resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("set alt text failed (http=%d)", resp.StatusCode)
	}
	return nil
}

// detectMIMEType guesses the MIME type from file extension then content sniffing.
func detectMIMEType(filePath string, data []byte) string {
	ext := filepath.Ext(filePath)
	if ext != "" {
		if t := mime.TypeByExtension(ext); t != "" {
			return t
		}
	}
	return http.DetectContentType(data)
}

// mimeToMediaCategory maps a MIME type to X's media_category value.
func mimeToMediaCategory(mimeType string) string {
	switch {
	case mimeType == "image/gif":
		return "tweet_gif"
	case strings.HasPrefix(mimeType, "video/"):
		return "tweet_video"
	default:
		return "tweet_image"
	}
}
