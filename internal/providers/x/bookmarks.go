package x

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// GraphQL query hashes for bookmark operations.
const (
	hashBookmarks              = "qToeLeMs43Q8cr7tRYXmaQ"
	hashCreateBookmark         = "aoDbu3RHznuiSkQ9aNM67Q"
	hashBookmarkToFolder       = "4KHZvvNbHNf07bsgnL9gWA"
	hashDeleteBookmark         = "Wlmlj2-xzyS1GN3a6cj-mQ"
	hashBookmarksAllDelete     = "skiACZKC1GDYli-M8RzEPQ"
	hashBookmarkFoldersSlice   = "i78YDd0Tza-dV4SYs58kRg"
	hashBookmarkFolderTimeline = "8HoabOvl7jl9IC1Aixj-vg"
	hashCreateBookmarkFolder   = "6Xxqpq8TM_CREYiuof_h5w"
	hashEditBookmarkFolder     = "a6kPp1cS1Dgbsjhapz1PNw"
	hashDeleteBookmarkFolder   = "2UTTsO-6zs93XqlEUZPsSg"
)

// BookmarkFolder is a condensed representation of an X bookmark folder.
type BookmarkFolder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// newBookmarksCmd builds the "bookmarks" subcommand group.
func newBookmarksCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bookmarks",
		Short:   "Manage bookmarks and bookmark folders",
		Aliases: []string{"bookmark", "bm"},
	}
	cmd.AddCommand(newBookmarksListCmd(factory))
	cmd.AddCommand(newBookmarksAddCmd(factory))
	cmd.AddCommand(newBookmarksRemoveCmd(factory))
	cmd.AddCommand(newBookmarksClearCmd(factory))
	cmd.AddCommand(newBookmarksFoldersCmd(factory))
	cmd.AddCommand(newBookmarksFolderTweetsCmd(factory))
	cmd.AddCommand(newBookmarksCreateFolderCmd(factory))
	cmd.AddCommand(newBookmarksEditFolderCmd(factory))
	cmd.AddCommand(newBookmarksDeleteFolderCmd(factory))
	return cmd
}

// newBookmarksListCmd builds the "bookmarks list" command.
func newBookmarksListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List bookmarked tweets",
		RunE:  makeRunBookmarksList(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newBookmarksAddCmd builds the "bookmarks add" command.
func newBookmarksAddCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Bookmark a tweet",
		RunE:  makeRunBookmarksAdd(factory),
	}
	cmd.Flags().String("tweet-id", "", "Tweet ID to bookmark (required)")
	_ = cmd.MarkFlagRequired("tweet-id")
	cmd.Flags().String("folder-id", "", "Folder ID to add bookmark to (optional)")
	cmd.Flags().Bool("dry-run", false, "Print what would be bookmarked without bookmarking")
	return cmd
}

// newBookmarksRemoveCmd builds the "bookmarks remove" command.
func newBookmarksRemoveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a bookmark",
		RunE:  makeRunBookmarksRemove(factory),
	}
	cmd.Flags().String("tweet-id", "", "Tweet ID to remove from bookmarks (required)")
	_ = cmd.MarkFlagRequired("tweet-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be removed without removing")
	return cmd
}

// newBookmarksClearCmd builds the "bookmarks clear" command.
func newBookmarksClearCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Delete all bookmarks",
		RunE:  makeRunBookmarksClear(factory),
	}
	cmd.Flags().Bool("confirm", false, "Confirm deletion of all bookmarks")
	cmd.Flags().Bool("dry-run", false, "Print what would be cleared without clearing")
	return cmd
}

// newBookmarksFoldersCmd builds the "bookmarks folders" command.
func newBookmarksFoldersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "folders",
		Short: "List bookmark folders",
		RunE:  makeRunBookmarksFolders(factory),
	}
	return cmd
}

// newBookmarksFolderTweetsCmd builds the "bookmarks folder-tweets" command.
func newBookmarksFolderTweetsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "folder-tweets",
		Short: "List tweets in a bookmark folder",
		RunE:  makeRunBookmarksFolderTweets(factory),
	}
	cmd.Flags().String("folder-id", "", "Folder ID (required)")
	_ = cmd.MarkFlagRequired("folder-id")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newBookmarksCreateFolderCmd builds the "bookmarks create-folder" command.
func newBookmarksCreateFolderCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-folder",
		Short: "Create a bookmark folder",
		RunE:  makeRunBookmarksCreateFolder(factory),
	}
	cmd.Flags().String("name", "", "Folder name (required)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().Bool("dry-run", false, "Print what would be created without creating")
	return cmd
}

// newBookmarksEditFolderCmd builds the "bookmarks edit-folder" command.
func newBookmarksEditFolderCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-folder",
		Short: "Edit a bookmark folder",
		RunE:  makeRunBookmarksEditFolder(factory),
	}
	cmd.Flags().String("folder-id", "", "Folder ID (required)")
	_ = cmd.MarkFlagRequired("folder-id")
	cmd.Flags().String("name", "", "New folder name (required)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().Bool("dry-run", false, "Print what would be edited without editing")
	return cmd
}

// newBookmarksDeleteFolderCmd builds the "bookmarks delete-folder" command.
func newBookmarksDeleteFolderCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-folder",
		Short: "Delete a bookmark folder",
		RunE:  makeRunBookmarksDeleteFolder(factory),
	}
	cmd.Flags().String("folder-id", "", "Folder ID (required)")
	_ = cmd.MarkFlagRequired("folder-id")
	cmd.Flags().Bool("confirm", false, "Confirm deletion of the folder")
	cmd.Flags().Bool("dry-run", false, "Print what would be deleted without deleting")
	return cmd
}

// --- RunE implementations ---

func makeRunBookmarksList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"count":                  limit,
			"includePromotedContent": false,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashBookmarks, "Bookmarks", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching bookmarks: %w", err)
		}

		return printTimelineResult(cmd, data)
	}
}

func makeRunBookmarksAdd(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tweetID, _ := cmd.Flags().GetString("tweet-id")
		folderID, _ := cmd.Flags().GetString("folder-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Choose hash based on whether a folder is specified.
		queryHash := hashCreateBookmark
		operationName := "CreateBookmark"
		vars := map[string]any{
			"tweet_id": tweetID,
		}
		if folderID != "" {
			queryHash = hashBookmarkToFolder
			operationName = "BookmarkToFolder"
			vars["bookmark_collection_id"] = folderID
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("bookmark tweet %s", tweetID), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, queryHash, operationName, vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("bookmarking tweet %s: %w", tweetID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "bookmarked", "tweet_id": tweetID})
		}
		fmt.Printf("Tweet bookmarked: %s\n", tweetID)
		return nil
	}
}

func makeRunBookmarksRemove(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tweetID, _ := cmd.Flags().GetString("tweet-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"tweet_id": tweetID,
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("remove bookmark for tweet %s", tweetID), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, hashDeleteBookmark, "DeleteBookmark", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("removing bookmark for tweet %s: %w", tweetID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "removed", "tweet_id": tweetID})
		}
		fmt.Printf("Bookmark removed: %s\n", tweetID)
		return nil
	}
}

func makeRunBookmarksClear(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, "clear all bookmarks", map[string]string{})
		}

		if err := confirmDestructive(cmd, "this will delete all bookmarks"); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, hashBookmarksAllDelete, "BookmarksAllDelete", map[string]any{}, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("clearing bookmarks: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "cleared"})
		}
		fmt.Println("All bookmarks cleared.")
		return nil
	}
}

func makeRunBookmarksFolders(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := client.GraphQL(ctx, hashBookmarkFoldersSlice, "BookmarkFoldersSlice", map[string]any{}, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching bookmark folders: %w", err)
		}

		folders, err := parseBookmarkFolders(data)
		if err != nil {
			return fmt.Errorf("parse bookmark folders: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(folders)
		}

		if len(folders) == 0 {
			fmt.Println("No bookmark folders found.")
			return nil
		}
		lines := make([]string, 0, len(folders)+1)
		lines = append(lines, fmt.Sprintf("%-30s  %-30s", "ID", "NAME"))
		for _, f := range folders {
			lines = append(lines, fmt.Sprintf("%-30s  %-30s",
				truncate(f.ID, 30),
				truncate(f.Name, 30),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

func makeRunBookmarksFolderTweets(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		folderID, _ := cmd.Flags().GetString("folder-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"bookmark_collection_id": folderID,
			"count":                  limit,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashBookmarkFolderTimeline, "BookmarkFolderTimeline", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching folder tweets for %s: %w", folderID, err)
		}

		return printTimelineResult(cmd, data)
	}
}

func makeRunBookmarksCreateFolder(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"name": name,
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("create bookmark folder %q", name), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := client.GraphQLPost(ctx, hashCreateBookmarkFolder, "CreateBookmarkFolder", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("creating bookmark folder %q: %w", name, err)
		}

		folder, err := parseCreatedBookmarkFolder(data)
		if err != nil {
			// Return partial success if parsing fails.
			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]string{"status": "created", "name": name})
			}
			fmt.Printf("Bookmark folder created: %s\n", name)
			return nil
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(folder)
		}
		fmt.Printf("Bookmark folder created: %s (ID: %s)\n", folder.Name, folder.ID)
		return nil
	}
}

func makeRunBookmarksEditFolder(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		folderID, _ := cmd.Flags().GetString("folder-id")
		name, _ := cmd.Flags().GetString("name")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"bookmark_collection_id": folderID,
			"name":                   name,
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("edit bookmark folder %s to name %q", folderID, name), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, hashEditBookmarkFolder, "EditBookmarkFolder", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("editing bookmark folder %s: %w", folderID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "updated", "folder_id": folderID, "name": name})
		}
		fmt.Printf("Bookmark folder updated: %s\n", folderID)
		return nil
	}
}

func makeRunBookmarksDeleteFolder(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		folderID, _ := cmd.Flags().GetString("folder-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("delete bookmark folder %s", folderID), map[string]string{"bookmark_collection_id": folderID})
		}

		if err := confirmDestructive(cmd, "this will delete the bookmark folder"); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"bookmark_collection_id": folderID,
		}

		_, err = client.GraphQLPost(ctx, hashDeleteBookmarkFolder, "DeleteBookmarkFolder", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("deleting bookmark folder %s: %w", folderID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "folder_id": folderID})
		}
		fmt.Printf("Bookmark folder deleted: %s\n", folderID)
		return nil
	}
}

// parseBookmarkFolders extracts bookmark folders from the GraphQL response data.
func parseBookmarkFolders(data json.RawMessage) ([]BookmarkFolder, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, fmt.Errorf("parse folders data: %w", err)
	}

	// Walk for a "bookmark_collections_slice" or similar key.
	for _, v := range top {
		var nested struct {
			Items []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"items"`
		}
		if err := json.Unmarshal(v, &nested); err == nil && len(nested.Items) > 0 {
			folders := make([]BookmarkFolder, 0, len(nested.Items))
			for _, item := range nested.Items {
				folders = append(folders, BookmarkFolder{ID: item.ID, Name: item.Name})
			}
			return folders, nil
		}

		// Try as an array of folder objects directly.
		var items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(v, &items); err == nil && len(items) > 0 {
			folders := make([]BookmarkFolder, 0, len(items))
			for _, item := range items {
				if item.ID != "" {
					folders = append(folders, BookmarkFolder{ID: item.ID, Name: item.Name})
				}
			}
			if len(folders) > 0 {
				return folders, nil
			}
		}
	}

	return []BookmarkFolder{}, nil
}

// parseCreatedBookmarkFolder extracts the created folder from the GraphQL response.
func parseCreatedBookmarkFolder(data json.RawMessage) (*BookmarkFolder, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, fmt.Errorf("parse created folder data: %w", err)
	}

	for _, v := range top {
		var folder struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(v, &folder); err == nil && folder.ID != "" {
			return &BookmarkFolder{ID: folder.ID, Name: folder.Name}, nil
		}
	}

	return nil, fmt.Errorf("created folder not found in response")
}
