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

// collectionMutateResponse is a generic response for create/edit/delete.
type collectionMutateResponse struct {
	Collection *rawCollection `json:"collection,omitempty"`
	Status     string         `json:"status"`
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
		Short:   "Manage saved media collections",
		Aliases: []string{"collection", "saved"},
	}
	cmd.AddCommand(newCollectionsListCmd(factory))
	cmd.AddCommand(newCollectionsGetCmd(factory))
	cmd.AddCommand(newCollectionsCreateCmd(factory))
	cmd.AddCommand(newCollectionsEditCmd(factory))
	cmd.AddCommand(newCollectionsDeleteCmd(factory))
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
		params.Set("collection_types", "ALL_MEDIA_AUTO_COLLECTION,MEDIA")
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

		resp, err := client.MobileGet(ctx, "/api/v1/feed/collection/"+collectionID+"/", params)
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

func newCollectionsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new collection",
		RunE:  makeRunCollectionsCreate(factory),
	}
	cmd.Flags().String("name", "", "Collection name")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunCollectionsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("create collection %q", name), map[string]string{"name": name})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("name", name)
		body.Set("module_name", "feed_contextual_post")

		resp, err := client.MobilePost(ctx, "/api/v1/collections/create/", body)
		if err != nil {
			return fmt.Errorf("creating collection: %w", err)
		}

		var result collectionMutateResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding create collection response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Created collection %q\n", name)
		return nil
	}
}

func newCollectionsEditCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Rename a collection",
		RunE:  makeRunCollectionsEdit(factory),
	}
	cmd.Flags().String("collection-id", "", "Collection ID")
	_ = cmd.MarkFlagRequired("collection-id")
	cmd.Flags().String("name", "", "New collection name")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunCollectionsEdit(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		collectionID, _ := cmd.Flags().GetString("collection-id")
		name, _ := cmd.Flags().GetString("name")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("rename collection %s to %q", collectionID, name),
				map[string]string{"collection_id": collectionID, "name": name})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("name", name)

		resp, err := client.MobilePost(ctx, "/api/v1/collections/"+collectionID+"/edit/", body)
		if err != nil {
			return fmt.Errorf("editing collection %s: %w", collectionID, err)
		}

		var result collectionMutateResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding edit collection response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Renamed collection %s to %q\n", collectionID, name)
		return nil
	}
}

func newCollectionsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a collection",
		RunE:  makeRunCollectionsDelete(factory),
	}
	cmd.Flags().String("collection-id", "", "Collection ID")
	_ = cmd.MarkFlagRequired("collection-id")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunCollectionsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		collectionID, _ := cmd.Flags().GetString("collection-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("delete collection %s", collectionID),
				map[string]string{"collection_id": collectionID})
		}
		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/collections/"+collectionID+"/delete/", nil)
		if err != nil {
			return fmt.Errorf("deleting collection %s: %w", collectionID, err)
		}

		var result collectionMutateResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding delete collection response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Deleted collection %s\n", collectionID)
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
