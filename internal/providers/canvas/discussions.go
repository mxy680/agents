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
	cmd.AddCommand(newDiscussionsCreateCmd(factory))
	cmd.AddCommand(newDiscussionsUpdateCmd(factory))
	cmd.AddCommand(newDiscussionsDeleteCmd(factory))
	cmd.AddCommand(newDiscussionsReplyCmd(factory))
	cmd.AddCommand(newDiscussionsMarkReadCmd(factory))

	return cmd
}

func newDiscussionsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new discussion topic in a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			title, _ := cmd.Flags().GetString("title")
			message, _ := cmd.Flags().GetString("message")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if title == "" {
				return fmt.Errorf("--title is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, "create discussion topic: "+title, map[string]any{
					"course_id": courseID, "title": title, "message": message,
				})
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			body := map[string]any{"title": title, "message": message}
			data, err := client.Post(ctx, "/courses/"+courseID+"/discussion_topics", body)
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
			fmt.Printf("Discussion topic %d created: %s\n", topic.ID, topic.Title)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("title", "", "Topic title (required)")
	cmd.Flags().String("message", "", "Topic message/body")
	return cmd
}

func newDiscussionsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a discussion topic",
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

			body := map[string]any{}
			if title, _ := cmd.Flags().GetString("title"); title != "" {
				body["title"] = title
			}
			if message, _ := cmd.Flags().GetString("message"); message != "" {
				body["message"] = message
			}
			if pinned, _ := cmd.Flags().GetBool("pinned"); pinned {
				body["pinned"] = true
			}
			if locked, _ := cmd.Flags().GetBool("locked"); locked {
				body["locked"] = true
			}

			data, err := client.Put(ctx, "/courses/"+courseID+"/discussion_topics/"+topicID, body)
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
			fmt.Printf("Discussion topic %s updated\n", topicID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("topic-id", "", "Canvas discussion topic ID (required)")
	cmd.Flags().String("title", "", "New title")
	cmd.Flags().String("message", "", "New message")
	cmd.Flags().Bool("pinned", false, "Pin the topic")
	cmd.Flags().Bool("locked", false, "Lock the topic")
	return cmd
}

func newDiscussionsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a discussion topic",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			topicID, _ := cmd.Flags().GetString("topic-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if topicID == "" {
				return fmt.Errorf("--topic-id is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the discussion topic"); err != nil {
				return err
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			if _, err := client.Delete(ctx, "/courses/"+courseID+"/discussion_topics/"+topicID); err != nil {
				return err
			}

			fmt.Printf("Discussion topic %s deleted\n", topicID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("topic-id", "", "Canvas discussion topic ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive action")
	return cmd
}

func newDiscussionsReplyCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reply",
		Short: "Post a reply/entry to a discussion topic",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			topicID, _ := cmd.Flags().GetString("topic-id")
			message, _ := cmd.Flags().GetString("message")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if topicID == "" {
				return fmt.Errorf("--topic-id is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, "reply to discussion topic "+topicID, map[string]any{
					"message": message,
				})
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			body := map[string]any{"message": message}
			path := "/courses/" + courseID + "/discussion_topics/" + topicID + "/entries"
			data, err := client.Post(ctx, path, body)
			if err != nil {
				return err
			}

			var entry DiscussionEntry
			if err := json.Unmarshal(data, &entry); err != nil {
				return fmt.Errorf("parse entry: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(entry)
			}
			fmt.Printf("Reply %d posted to discussion topic %s\n", entry.ID, topicID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("topic-id", "", "Canvas discussion topic ID (required)")
	cmd.Flags().String("message", "", "Reply message")
	return cmd
}

func newDiscussionsMarkReadCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-read",
		Short: "Mark all entries in a discussion topic as read",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			topicID, _ := cmd.Flags().GetString("topic-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if topicID == "" {
				return fmt.Errorf("--topic-id is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, "mark discussion topic "+topicID+" as read", nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			path := "/courses/" + courseID + "/discussion_topics/" + topicID + "/read_all"
			if _, err := client.Put(ctx, path, nil); err != nil {
				return err
			}

			fmt.Printf("Discussion topic %s marked as read\n", topicID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("topic-id", "", "Canvas discussion topic ID (required)")
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
