package instagram

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// collectionsListResponse is the response for GET /api/v1/collections/list/.
type collectionsListResponse struct {
	Items         []rawCollection `json:"items"`
	NextMaxID     string          `json:"next_max_id"`
	MoreAvailable bool            `json:"more_available"`
	Status        string          `json:"status"`
}

// rawCollection is the raw collection object from the Instagram API.
type rawCollection struct {
	CollectionID   string          `json:"collection_id"`
	CollectionName string          `json:"collection_name"`
	CollectionType string          `json:"collection_type"`
	MediaCount     int64           `json:"media_count"`
	CoverMedia     *rawCoverMedia  `json:"cover_media,omitempty"`
}

// rawCoverMedia holds the cover media info for a collection.
type rawCoverMedia struct {
	CroppedImageVersion struct {
		URL string `json:"url"`
	} `json:"cropped_image_version"`
}

// collectionFeedResponse is the response for GET /api/v1/feed/collection/{id}/.
type collectionFeedResponse struct {
	Items         []rawMediaItem `json:"items"`
	NextMaxID     string         `json:"next_max_id"`
	MoreAvailable bool           `json:"more_available"`
	Status        string         `json:"status"`
}

// savedFeedResponse is the response for GET /api/v1/feed/saved/posts/.
type savedFeedResponse struct {
	Items         []rawSavedItem `json:"items"`
	NextMaxID     string         `json:"next_max_id"`
	MoreAvailable bool           `json:"more_available"`
	Status        string         `json:"status"`
}

// rawSavedItem wraps a media item inside the saved posts feed.
type rawSavedItem struct {
	Media rawMediaItem `json:"media"`
}

// toCollectionSummary converts a rawCollection to CollectionSummary.
func toCollectionSummary(c rawCollection) CollectionSummary {
	coverURL := ""
	if c.CoverMedia != nil {
		coverURL = c.CoverMedia.CroppedImageVersion.URL
	}
	return CollectionSummary{
		CollectionID:   c.CollectionID,
		CollectionName: c.CollectionName,
		CollectionType: c.CollectionType,
		MediaCount:     c.MediaCount,
		CoverMediaURL:  coverURL,
	}
}

// newCollectionsCmd builds the `collections` subcommand group.
func newCollectionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "collections",
		Short:   "View saved media collections",
		Aliases: []string{"collection", "saved"},
	}
	cmd.AddCommand(newCollectionsListCmd(factory))
	cmd.AddCommand(newCollectionsGetCmd(factory))
	cmd.AddCommand(newCollectionsSavedCmd(factory))
	return cmd
}

func newCollectionsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your saved collections",
		RunE:  makeRunCollectionsList(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of collections to return")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func makeRunCollectionsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		// Include all collection types to avoid 500 errors from the real API.
		params.Set("collection_types", `["ALL_MEDIA_AUTO_COLLECTION","MEDIA","PRODUCT_AUTO_COLLECTION"]`)
		params.Set("count", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("max_id", cursor)
		}

		resp, err := client.MobileGet(ctx, "/api/v1/collections/list/", params)
		if err != nil {
			return fmt.Errorf("listing collections: %w", err)
		}

		var result collectionsListResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding collections list: %w", err)
		}

		summaries := make([]CollectionSummary, 0, len(result.Items))
		for _, item := range result.Items {
			summaries = append(summaries, toCollectionSummary(item))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}

		if len(summaries) == 0 {
			fmt.Println("No collections found.")
			return nil
		}

		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-20s  %-30s  %-15s  %-10s", "ID", "NAME", "TYPE", "ITEMS"))
		for _, c := range summaries {
			lines = append(lines, fmt.Sprintf("%-20s  %-30s  %-15s  %-10s",
				truncate(c.CollectionID, 20),
				truncate(c.CollectionName, 30),
				truncate(c.CollectionType, 15),
				formatCount(c.MediaCount),
			))
		}
		cli.PrintText(lines)
		if result.MoreAvailable && result.NextMaxID != "" {
			fmt.Printf("Next cursor: %s\n", result.NextMaxID)
		}
		return nil
	}
}

func newCollectionsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get media items in a collection",
		RunE:  makeRunCollectionsGet(factory),
	}
	cmd.Flags().String("collection-id", "", "Collection ID")
	_ = cmd.MarkFlagRequired("collection-id")
	cmd.Flags().Int("limit", 20, "Maximum number of items to return")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func makeRunCollectionsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		collectionID, _ := cmd.Flags().GetString("collection-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("count", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("max_id", cursor)
		}

		resp, err := client.MobileGet(ctx, "/api/v1/feed/collection/"+url.PathEscape(collectionID)+"/", params)
		if err != nil {
			return fmt.Errorf("getting collection %s: %w", collectionID, err)
		}

		var result collectionFeedResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding collection feed: %w", err)
		}

		summaries := make([]MediaSummary, 0, len(result.Items))
		for _, item := range result.Items {
			summaries = append(summaries, toMediaSummary(item))
		}

		if err := printMediaSummaries(cmd, summaries); err != nil {
			return err
		}
		if result.MoreAvailable && result.NextMaxID != "" {
			fmt.Printf("Next cursor: %s\n", result.NextMaxID)
		}
		return nil
	}
}

func newCollectionsSavedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "saved",
		Short: "List all saved posts (All Posts collection)",
		RunE:  makeRunCollectionsSaved(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of items to return")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func makeRunCollectionsSaved(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("count", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("max_id", cursor)
		}

		resp, err := client.MobileGet(ctx, "/api/v1/feed/saved/posts/", params)
		if err != nil {
			return fmt.Errorf("listing saved posts: %w", err)
		}

		var result savedFeedResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding saved posts response: %w", err)
		}

		summaries := make([]MediaSummary, 0, len(result.Items))
		for _, item := range result.Items {
			summaries = append(summaries, toMediaSummary(item.Media))
		}

		if err := printMediaSummaries(cmd, summaries); err != nil {
			return err
		}
		if result.MoreAvailable && result.NextMaxID != "" {
			fmt.Printf("Next cursor: %s\n", result.NextMaxID)
		}
		return nil
	}
}
