package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newAnnouncementsCmd returns the parent "announcements" command with all subcommands attached.
func newAnnouncementsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "announcements",
		Short:   "Manage Canvas announcements",
		Aliases: []string{"announce", "ann"},
	}

	cmd.AddCommand(newAnnouncementsListCmd(factory))
	cmd.AddCommand(newAnnouncementsGetCmd(factory))
	cmd.AddCommand(newAnnouncementsCreateCmd(factory))
	cmd.AddCommand(newAnnouncementsUpdateCmd(factory))
	cmd.AddCommand(newAnnouncementsDeleteCmd(factory))

	return cmd
}

func newAnnouncementsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List announcements for one or more courses",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseIDs, _ := cmd.Flags().GetString("course-ids")
			if courseIDs == "" {
				return fmt.Errorf("--course-ids is required")
			}

			startDate, _ := cmd.Flags().GetString("start-date")
			endDate, _ := cmd.Flags().GetString("end-date")
			activeOnly, _ := cmd.Flags().GetBool("active-only")
			limit, _ := cmd.Flags().GetInt("limit")

			// Build context_codes[] params manually since url.Values encodes
			// context_codes%5B%5D which Canvas also accepts, but we use Add to
			// produce repeated keys for each course ID.
			params := url.Values{}
			for _, id := range strings.Split(courseIDs, ",") {
				id = strings.TrimSpace(id)
				if id != "" {
					params.Add("context_codes[]", "course_"+id)
				}
			}
			if startDate != "" {
				params.Set("start_date", startDate)
			}
			if endDate != "" {
				params.Set("end_date", endDate)
			}
			if activeOnly {
				params.Set("active_only", "true")
			}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/announcements", params)
			if err != nil {
				return err
			}

			var announcements []AnnouncementSummary
			if err := json.Unmarshal(data, &announcements); err != nil {
				return fmt.Errorf("parse announcements: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(announcements)
			}

			if len(announcements) == 0 {
				fmt.Println("No announcements found.")
				return nil
			}
			for _, a := range announcements {
				posted := a.PostedAt
				if posted == "" {
					posted = "—"
				}
				fmt.Printf("%-6d  %-25s  %s\n", a.ID, posted, truncate(a.Title, 50))
			}
			return nil
		},
	}

	cmd.Flags().String("course-ids", "", "Comma-separated Canvas course IDs (required)")
	cmd.Flags().String("start-date", "", "Return announcements posted after this date (RFC3339)")
	cmd.Flags().String("end-date", "", "Return announcements posted before this date (RFC3339)")
	cmd.Flags().Bool("active-only", false, "Only return active (non-deleted) announcements")
	cmd.Flags().Int("limit", 0, "Maximum number of announcements to return")
	return cmd
}

func newAnnouncementsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific announcement",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			announcementID, _ := cmd.Flags().GetString("announcement-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if announcementID == "" {
				return fmt.Errorf("--announcement-id is required")
			}

			// Announcements are discussion topics with is_announcement=true.
			data, err := client.Get(ctx, "/courses/"+courseID+"/discussion_topics/"+announcementID, nil)
			if err != nil {
				return err
			}

			var announcement AnnouncementSummary
			if err := json.Unmarshal(data, &announcement); err != nil {
				return fmt.Errorf("parse announcement: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(announcement)
			}

			fmt.Printf("ID:        %d\n", announcement.ID)
			fmt.Printf("Title:     %s\n", announcement.Title)
			fmt.Printf("Published: %v\n", announcement.Published)
			if announcement.UserName != "" {
				fmt.Printf("Author:    %s\n", announcement.UserName)
			}
			if announcement.PostedAt != "" {
				fmt.Printf("Posted:    %s\n", announcement.PostedAt)
			}
			if announcement.Message != "" {
				fmt.Printf("Message:   %s\n", truncate(announcement.Message, 200))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("announcement-id", "", "Canvas announcement (discussion topic) ID (required)")
	return cmd
}

func newAnnouncementsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new announcement",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			title, _ := cmd.Flags().GetString("title")
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			message, _ := cmd.Flags().GetString("message")
			if message == "" {
				return fmt.Errorf("--message is required")
			}

			published, _ := cmd.Flags().GetBool("published")

			body := map[string]any{
				"title":           title,
				"message":         message,
				"is_announcement": true,
				"published":       published,
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("create announcement %q in course %s", title, courseID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/courses/"+courseID+"/discussion_topics", body)
			if err != nil {
				return err
			}

			var announcement AnnouncementSummary
			if err := json.Unmarshal(data, &announcement); err != nil {
				return fmt.Errorf("parse created announcement: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(announcement)
			}
			fmt.Printf("Announcement created: %d — %s\n", announcement.ID, announcement.Title)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("title", "", "Announcement title (required)")
	cmd.Flags().String("message", "", "Announcement body (required)")
	cmd.Flags().Bool("published", false, "Publish the announcement immediately")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newAnnouncementsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing announcement",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			announcementID, _ := cmd.Flags().GetString("announcement-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if announcementID == "" {
				return fmt.Errorf("--announcement-id is required")
			}

			body := map[string]any{}
			if cmd.Flags().Changed("title") {
				v, _ := cmd.Flags().GetString("title")
				body["title"] = v
			}
			if cmd.Flags().Changed("message") {
				v, _ := cmd.Flags().GetString("message")
				body["message"] = v
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("update announcement %s in course %s", announcementID, courseID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Put(ctx, "/courses/"+courseID+"/discussion_topics/"+announcementID, body)
			if err != nil {
				return err
			}

			var announcement AnnouncementSummary
			if err := json.Unmarshal(data, &announcement); err != nil {
				return fmt.Errorf("parse updated announcement: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(announcement)
			}
			fmt.Printf("Announcement %d updated.\n", announcement.ID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("announcement-id", "", "Canvas announcement (discussion topic) ID (required)")
	cmd.Flags().String("title", "", "New title")
	cmd.Flags().String("message", "", "New body/message")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newAnnouncementsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an announcement",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			announcementID, _ := cmd.Flags().GetString("announcement-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if announcementID == "" {
				return fmt.Errorf("--announcement-id is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the announcement"); err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("delete announcement %s in course %s", announcementID, courseID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Delete(ctx, "/courses/"+courseID+"/discussion_topics/"+announcementID)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"deleted": true, "announcement_id": announcementID})
			}
			fmt.Printf("Announcement %s deleted.\n", announcementID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("announcement-id", "", "Canvas announcement (discussion topic) ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}
