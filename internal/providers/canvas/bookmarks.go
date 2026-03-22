package canvas

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newBookmarksCmd returns the parent "bookmarks" command with all subcommands attached.
func newBookmarksCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bookmarks",
		Short:   "Manage Canvas bookmarks",
		Aliases: []string{"bookmark", "bm"},
	}

	cmd.AddCommand(newBookmarksListCmd(factory))
	cmd.AddCommand(newBookmarksGetCmd(factory))
	cmd.AddCommand(newBookmarksCreateCmd(factory))
	cmd.AddCommand(newBookmarksUpdateCmd(factory))
	cmd.AddCommand(newBookmarksDeleteCmd(factory))

	return cmd
}

func newBookmarksListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List bookmarks for the current user",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Get(ctx, "/users/self/bookmarks", nil)
			if err != nil {
				return err
			}

			var bookmarks []BookmarkSummary
			if err := json.Unmarshal(data, &bookmarks); err != nil {
				return fmt.Errorf("parse bookmarks: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(bookmarks)
			}

			if len(bookmarks) == 0 {
				fmt.Println("No bookmarks found.")
				return nil
			}
			for _, b := range bookmarks {
				fmt.Printf("%-6d  pos:%-4d  %-40s  %s\n", b.ID, b.Position, truncate(b.Name, 40), truncate(b.URL, 60))
			}
			return nil
		},
	}

	return cmd
}

func newBookmarksGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific bookmark",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			bookmarkID, _ := cmd.Flags().GetString("bookmark-id")
			if bookmarkID == "" {
				return fmt.Errorf("--bookmark-id is required")
			}

			data, err := client.Get(ctx, "/users/self/bookmarks/"+bookmarkID, nil)
			if err != nil {
				return err
			}

			var bookmark BookmarkSummary
			if err := json.Unmarshal(data, &bookmark); err != nil {
				return fmt.Errorf("parse bookmark: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(bookmark)
			}

			fmt.Printf("ID:       %d\n", bookmark.ID)
			fmt.Printf("Name:     %s\n", bookmark.Name)
			fmt.Printf("URL:      %s\n", bookmark.URL)
			if bookmark.Position > 0 {
				fmt.Printf("Position: %d\n", bookmark.Position)
			}
			return nil
		},
	}

	cmd.Flags().String("bookmark-id", "", "Canvas bookmark ID (required)")
	return cmd
}

func newBookmarksCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new bookmark",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			name, _ := cmd.Flags().GetString("name")
			bookmarkURL, _ := cmd.Flags().GetString("url")
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if bookmarkURL == "" {
				return fmt.Errorf("--url is required")
			}

			position, _ := cmd.Flags().GetInt("position")

			body := map[string]any{
				"name": name,
				"url":  bookmarkURL,
			}
			if position > 0 {
				body["position"] = position
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("create bookmark %q", name), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/users/self/bookmarks", body)
			if err != nil {
				return err
			}

			var bookmark BookmarkSummary
			if err := json.Unmarshal(data, &bookmark); err != nil {
				return fmt.Errorf("parse created bookmark: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(bookmark)
			}
			fmt.Printf("Bookmark created: %d — %s\n", bookmark.ID, bookmark.Name)
			return nil
		},
	}

	cmd.Flags().String("name", "", "Bookmark name (required)")
	cmd.Flags().String("url", "", "Bookmark URL (required)")
	cmd.Flags().Int("position", 0, "Position ordering for the bookmark")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newBookmarksUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing bookmark",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			bookmarkID, _ := cmd.Flags().GetString("bookmark-id")
			if bookmarkID == "" {
				return fmt.Errorf("--bookmark-id is required")
			}

			body := map[string]any{}
			if cmd.Flags().Changed("name") {
				v, _ := cmd.Flags().GetString("name")
				body["name"] = v
			}
			if cmd.Flags().Changed("url") {
				v, _ := cmd.Flags().GetString("url")
				body["url"] = v
			}
			if cmd.Flags().Changed("position") {
				v, _ := cmd.Flags().GetInt("position")
				body["position"] = v
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("update bookmark %s", bookmarkID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Put(ctx, "/users/self/bookmarks/"+bookmarkID, body)
			if err != nil {
				return err
			}

			var bookmark BookmarkSummary
			if err := json.Unmarshal(data, &bookmark); err != nil {
				return fmt.Errorf("parse updated bookmark: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(bookmark)
			}
			fmt.Printf("Bookmark %d updated.\n", bookmark.ID)
			return nil
		},
	}

	cmd.Flags().String("bookmark-id", "", "Canvas bookmark ID (required)")
	cmd.Flags().String("name", "", "New bookmark name")
	cmd.Flags().String("url", "", "New bookmark URL")
	cmd.Flags().Int("position", 0, "New position ordering")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newBookmarksDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a bookmark",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			bookmarkID, _ := cmd.Flags().GetString("bookmark-id")
			if bookmarkID == "" {
				return fmt.Errorf("--bookmark-id is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the bookmark"); err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("delete bookmark %s", bookmarkID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Delete(ctx, "/users/self/bookmarks/"+bookmarkID)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"deleted": true, "bookmark_id": bookmarkID})
			}
			fmt.Printf("Bookmark %s deleted.\n", bookmarkID)
			return nil
		},
	}

	cmd.Flags().String("bookmark-id", "", "Canvas bookmark ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}
