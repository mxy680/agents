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

