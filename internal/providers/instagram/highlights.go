package instagram

import (
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// highlightsTrayResponse is the response for GET /api/v1/highlights/{user_id}/highlights_tray/.
type highlightsTrayResponse struct {
	Tray   []rawHighlight `json:"tray"`
	Status string         `json:"status"`
}

// rawHighlight is the raw highlight/reel object from the highlights tray.
type rawHighlight struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	MediaCount int    `json:"media_count"`
	CreatedAt  int64  `json:"created_at"`
	CoverMedia *struct {
		CroppedImageVersion struct {
			URL string `json:"url"`
		} `json:"cropped_image_version"`
	} `json:"cover_media"`
}

// highlightMediaResponse is the response for GET /api/v1/feed/reels_media/.
type highlightMediaResponse struct {
	Reels  map[string]any `json:"reels"`
	Status string         `json:"status"`
}

// toHighlightSummary converts a rawHighlight to HighlightSummary.
func toHighlightSummary(h rawHighlight) HighlightSummary {
	coverURL := ""
	if h.CoverMedia != nil {
		coverURL = h.CoverMedia.CroppedImageVersion.URL
	}
	return HighlightSummary{
		ID:         h.ID,
		Title:      h.Title,
		MediaCount: h.MediaCount,
		CoverURL:   coverURL,
		CreatedAt:  h.CreatedAt,
	}
}

// newHighlightsCmd builds the `highlights` subcommand group.
func newHighlightsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "highlights",
		Short:   "View story highlights",
		Aliases: []string{"highlight", "hl"},
	}
	cmd.AddCommand(newHighlightsListCmd(factory))
	cmd.AddCommand(newHighlightsGetCmd(factory))
	return cmd
}

func newHighlightsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List story highlights for a user",
		RunE:  makeRunHighlightsList(factory),
	}
	cmd.Flags().String("user-id", "", "User ID (defaults to own user)")
	return cmd
}

func makeRunHighlightsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if userID == "" {
			userID = client.SelfUserID()
		}

		resp, err := client.MobileGet(ctx, "/api/v1/highlights/"+url.PathEscape(userID)+"/highlights_tray/", nil)
		if err != nil {
			return fmt.Errorf("listing highlights for user %s: %w", userID, err)
		}

		var result highlightsTrayResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding highlights tray: %w", err)
		}

		summaries := make([]HighlightSummary, 0, len(result.Tray))
		for _, h := range result.Tray {
			summaries = append(summaries, toHighlightSummary(h))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}

		if len(summaries) == 0 {
			fmt.Println("No highlights found.")
			return nil
		}

		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-30s  %-25s  %-8s  %-12s", "ID", "TITLE", "ITEMS", "CREATED"))
		for _, h := range summaries {
			lines = append(lines, fmt.Sprintf("%-30s  %-25s  %-8d  %-12s",
				truncate(h.ID, 30),
				truncate(h.Title, 25),
				h.MediaCount,
				formatTimestamp(h.CreatedAt),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newHighlightsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get media items in a highlight",
		RunE:  makeRunHighlightsGet(factory),
	}
	cmd.Flags().String("highlight-id", "", "Highlight ID")
	_ = cmd.MarkFlagRequired("highlight-id")
	return cmd
}

func makeRunHighlightsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		highlightID, _ := cmd.Flags().GetString("highlight-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("reel_ids", "highlight:"+highlightID)

		resp, err := client.MobileGet(ctx, "/api/v1/feed/reels_media/", params)
		if err != nil {
			return fmt.Errorf("getting highlight %s: %w", highlightID, err)
		}

		var result highlightMediaResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding highlight media: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Highlight %s retrieved.\n", highlightID)
		return nil
	}
}
