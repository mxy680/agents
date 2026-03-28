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

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, "create bookmark: "+name, map[string]any{"name": name, "url": bookmarkURL})
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			body := map[string]any{"name": name, "url": bookmarkURL}
			data, err := client.Post(ctx, "/users/self/bookmarks", body)
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
			fmt.Printf("Bookmark %d created: %s\n", bookmark.ID, bookmark.Name)
			return nil
		},
	}

	cmd.Flags().String("name", "", "Bookmark name (required)")
	cmd.Flags().String("url", "", "Bookmark URL")
	return cmd
}

func newBookmarksUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing bookmark",
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

			body := map[string]any{}
			if name, _ := cmd.Flags().GetString("name"); name != "" {
				body["name"] = name
			}
			if bookmarkURL, _ := cmd.Flags().GetString("url"); bookmarkURL != "" {
				body["url"] = bookmarkURL
			}

			data, err := client.Put(ctx, "/users/self/bookmarks/"+bookmarkID, body)
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
			fmt.Printf("Bookmark %s updated\n", bookmarkID)
			return nil
		},
	}

	cmd.Flags().String("bookmark-id", "", "Canvas bookmark ID (required)")
	cmd.Flags().String("name", "", "New name")
	cmd.Flags().String("url", "", "New URL")
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

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			if _, err := client.Delete(ctx, "/users/self/bookmarks/"+bookmarkID); err != nil {
				return err
			}

			fmt.Printf("Bookmark %s deleted\n", bookmarkID)
			return nil
		},
	}

	cmd.Flags().String("bookmark-id", "", "Canvas bookmark ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive action")
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

