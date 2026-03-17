package instagram

import (
	"fmt"
	"net/url"
	"strings"

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

// highlightMutateResponse is a generic response for create/edit/delete.
type highlightMutateResponse struct {
	Reel   *rawHighlight `json:"reel,omitempty"`
	Status string        `json:"status"`
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
		Short:   "Manage story highlights",
		Aliases: []string{"highlight", "hl"},
	}
	cmd.AddCommand(newHighlightsListCmd(factory))
	cmd.AddCommand(newHighlightsGetCmd(factory))
	cmd.AddCommand(newHighlightsCreateCmd(factory))
	cmd.AddCommand(newHighlightsEditCmd(factory))
	cmd.AddCommand(newHighlightsDeleteCmd(factory))
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
			userID = client.session.DSUserID
		}

		resp, err := client.MobileGet(ctx, "/api/v1/highlights/"+userID+"/highlights_tray/", nil)
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

func newHighlightsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new story highlight",
		RunE:  makeRunHighlightsCreate(factory),
	}
	cmd.Flags().String("title", "", "Highlight title")
	_ = cmd.MarkFlagRequired("title")
	cmd.Flags().String("story-ids", "", "Comma-separated story media IDs to add")
	_ = cmd.MarkFlagRequired("story-ids")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunHighlightsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		title, _ := cmd.Flags().GetString("title")
		storyIDs, _ := cmd.Flags().GetString("story-ids")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("create highlight %q with stories %s", title, storyIDs),
				map[string]string{"title": title, "story_ids": storyIDs})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		ids := strings.Split(storyIDs, ",")
		reelIDs := make([]string, 0, len(ids))
		for _, id := range ids {
			id = strings.TrimSpace(id)
			if id != "" {
				reelIDs = append(reelIDs, id)
			}
		}

		body := url.Values{}
		body.Set("title", title)
		for _, id := range reelIDs {
			body.Add("reel_ids[]", id)
		}

		resp, err := client.MobilePost(ctx, "/api/v1/highlights/create_reel/", body)
		if err != nil {
			return fmt.Errorf("creating highlight: %w", err)
		}

		var result highlightMutateResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding create highlight response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Created highlight %q\n", title)
		return nil
	}
}

func newHighlightsEditCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a story highlight",
		RunE:  makeRunHighlightsEdit(factory),
	}
	cmd.Flags().String("highlight-id", "", "Highlight ID")
	_ = cmd.MarkFlagRequired("highlight-id")
	cmd.Flags().String("title", "", "New title (optional)")
	cmd.Flags().String("add-stories", "", "Comma-separated story IDs to add (optional)")
	cmd.Flags().String("remove-stories", "", "Comma-separated story IDs to remove (optional)")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunHighlightsEdit(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		highlightID, _ := cmd.Flags().GetString("highlight-id")
		title, _ := cmd.Flags().GetString("title")
		addStories, _ := cmd.Flags().GetString("add-stories")
		removeStories, _ := cmd.Flags().GetString("remove-stories")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("edit highlight %s", highlightID),
				map[string]string{
					"highlight_id":   highlightID,
					"title":          title,
					"add_stories":    addStories,
					"remove_stories": removeStories,
				})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		if title != "" {
			body.Set("title", title)
		}
		for _, id := range parseCSV(addStories) {
			body.Add("added_reel_ids[]", id)
		}
		for _, id := range parseCSV(removeStories) {
			body.Add("removed_reel_ids[]", id)
		}

		resp, err := client.MobilePost(ctx, "/api/v1/highlights/"+highlightID+"/edit_reel/", body)
		if err != nil {
			return fmt.Errorf("editing highlight %s: %w", highlightID, err)
		}

		var result highlightMutateResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding edit highlight response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Edited highlight %s\n", highlightID)
		return nil
	}
}

func newHighlightsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a story highlight",
		RunE:  makeRunHighlightsDelete(factory),
	}
	cmd.Flags().String("highlight-id", "", "Highlight ID")
	_ = cmd.MarkFlagRequired("highlight-id")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunHighlightsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		highlightID, _ := cmd.Flags().GetString("highlight-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("delete highlight %s", highlightID),
				map[string]string{"highlight_id": highlightID})
		}
		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/highlights/"+highlightID+"/delete_reel/", nil)
		if err != nil {
			return fmt.Errorf("deleting highlight %s: %w", highlightID, err)
		}

		var result highlightMutateResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding delete highlight response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Deleted highlight %s\n", highlightID)
		return nil
	}
}

// parseCSV splits a comma-separated string into trimmed, non-empty parts.
func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
