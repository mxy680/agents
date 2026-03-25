package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newDiscussionsCmd returns the parent "discussions" command with all subcommands attached.
func newDiscussionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "discussions",
		Short:   "Manage Canvas discussion topics",
		Aliases: []string{"discuss", "disc"},
	}

	cmd.AddCommand(newDiscussionsListCmd(factory))
	cmd.AddCommand(newDiscussionsGetCmd(factory))
	cmd.AddCommand(newDiscussionsEntriesCmd(factory))

	return cmd
}

func newDiscussionsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List discussion topics for a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}

			scope, _ := cmd.Flags().GetString("scope")
			orderBy, _ := cmd.Flags().GetString("order-by")
			search, _ := cmd.Flags().GetString("search")
			limit, _ := cmd.Flags().GetInt("limit")

			params := url.Values{}
			if scope != "" {
				params.Set("scope", scope)
			}
			if orderBy != "" {
				params.Set("order_by", orderBy)
			}
			if search != "" {
				params.Set("search_term", search)
			}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/discussion_topics", params)
			if err != nil {
				return err
			}

			var topics []DiscussionSummary
			if err := json.Unmarshal(data, &topics); err != nil {
				return fmt.Errorf("parse discussion topics: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(topics)
			}

			if len(topics) == 0 {
				fmt.Println("No discussion topics found.")
				return nil
			}
			for _, t := range topics {
				pinned := ""
				if t.Pinned {
					pinned = " [pinned]"
				}
				locked := ""
				if t.Locked {
					locked = " [locked]"
				}
				fmt.Printf("%-6d  %s%s%s\n", t.ID, truncate(t.Title, 60), pinned, locked)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("scope", "", "Filter by scope: locked, unlocked, pinned, unpinned")
	cmd.Flags().String("order-by", "", "Sort order: position, recent_activity, title")
	cmd.Flags().String("search", "", "Search term to filter topics")
	cmd.Flags().Int("limit", 0, "Maximum number of topics to return")
	return cmd
}

func newDiscussionsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific discussion topic",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			topicID, _ := cmd.Flags().GetString("topic-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if topicID == "" {
				return fmt.Errorf("--topic-id is required")
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/discussion_topics/"+topicID, nil)
			if err != nil {
				return err
			}

			var topic DiscussionSummary
			if err := json.Unmarshal(data, &topic); err != nil {
				return fmt.Errorf("parse discussion topic: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(topic)
			}

			fmt.Printf("ID:        %d\n", topic.ID)
			fmt.Printf("Title:     %s\n", topic.Title)
			fmt.Printf("Type:      %s\n", topic.DiscussionType)
			fmt.Printf("Published: %v\n", topic.Published)
			fmt.Printf("Pinned:    %v\n", topic.Pinned)
			fmt.Printf("Locked:    %v\n", topic.Locked)
			if topic.UserName != "" {
				fmt.Printf("Author:    %s\n", topic.UserName)
			}
			if topic.PostedAt != "" {
				fmt.Printf("Posted:    %s\n", topic.PostedAt)
			}
			if topic.LastReplyAt != "" {
				fmt.Printf("Last Reply:%s\n", topic.LastReplyAt)
			}
			if topic.Message != "" {
				fmt.Printf("Message:   %s\n", truncate(topic.Message, 200))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("topic-id", "", "Canvas discussion topic ID (required)")
	return cmd
}

// DiscussionEntry is a single entry in a discussion topic.
type DiscussionEntry struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id,omitempty"`
	UserName  string `json:"user_name,omitempty"`
	Message   string `json:"message,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func newDiscussionsEntriesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entries",
		Short: "List entries in a discussion topic",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			topicID, _ := cmd.Flags().GetString("topic-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if topicID == "" {
				return fmt.Errorf("--topic-id is required")
			}

			limit, _ := cmd.Flags().GetInt("limit")
			params := url.Values{}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			path := "/courses/" + courseID + "/discussion_topics/" + topicID + "/entries"
			data, err := client.Get(ctx, path, params)
			if err != nil {
				return err
			}

			var entries []DiscussionEntry
			if err := json.Unmarshal(data, &entries); err != nil {
				return fmt.Errorf("parse discussion entries: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(entries)
			}

			if len(entries) == 0 {
				fmt.Println("No entries found.")
				return nil
			}
			for _, e := range entries {
				fmt.Printf("%-6d  %-20s  %s\n", e.ID, e.UserName, truncate(e.Message, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("topic-id", "", "Canvas discussion topic ID (required)")
	cmd.Flags().Int("limit", 0, "Maximum number of entries to return")
	return cmd
}
