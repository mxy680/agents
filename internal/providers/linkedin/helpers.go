package linkedin

import (
	"errors"
	"fmt"
	"time"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// ProfileSummary is a condensed LinkedIn profile representation.
type ProfileSummary struct {
	URN      string `json:"urn"`
	PublicID string `json:"public_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Headline  string `json:"headline,omitempty"`
	Location  string `json:"location,omitempty"`
}

// ProfileDetail extends ProfileSummary with full profile info.
type ProfileDetail struct {
	URN             string `json:"urn"`
	PublicID        string `json:"public_id"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Headline        string `json:"headline,omitempty"`
	Summary         string `json:"summary,omitempty"`
	Location        string `json:"location,omitempty"`
	Industry        string `json:"industry,omitempty"`
	ProfilePicURL   string `json:"profile_pic_url,omitempty"`
	ConnectionCount int    `json:"connection_count"`
	FollowerCount   int    `json:"follower_count"`
}

// ConnectionSummary is a condensed connection representation.
type ConnectionSummary struct {
	URN       string `json:"urn"`
	PublicID  string `json:"public_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Headline  string `json:"headline,omitempty"`
	CreatedAt int64  `json:"created_at"`
}

// PostSummary is a condensed LinkedIn post representation.
type PostSummary struct {
	URN          string `json:"urn"`
	AuthorURN    string `json:"author_urn"`
	AuthorName   string `json:"author_name,omitempty"`
	Text         string `json:"text,omitempty"`
	Timestamp    int64  `json:"timestamp"`
	LikeCount    int    `json:"like_count"`
	CommentCount int    `json:"comment_count"`
	ShareCount   int    `json:"share_count"`
}

// CommentSummary is a condensed comment representation.
type CommentSummary struct {
	URN        string `json:"urn"`
	Text       string `json:"text"`
	AuthorURN  string `json:"author_urn"`
	AuthorName string `json:"author_name,omitempty"`
	Timestamp  int64  `json:"timestamp"`
	LikeCount  int    `json:"like_count"`
}

// ConversationSummary is a condensed messaging conversation.
type ConversationSummary struct {
	ID              string   `json:"id"`
	Title           string   `json:"title,omitempty"`
	LastActivityAt  int64    `json:"last_activity_at"`
	UnreadCount     int      `json:"unread_count"`
	ParticipantURNs []string `json:"participant_urns"`
}

// MessageSummary is a condensed message representation.
type MessageSummary struct {
	ID         string `json:"id"`
	SenderURN  string `json:"sender_urn"`
	SenderName string `json:"sender_name,omitempty"`
	Text       string `json:"text"`
	Timestamp  int64  `json:"timestamp"`
}

// CompanySummary is a condensed company representation.
type CompanySummary struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Industry      string `json:"industry,omitempty"`
	EmployeeCount int    `json:"employee_count"`
	FollowerCount int    `json:"follower_count"`
}

// JobSummary is a condensed job listing representation.
type JobSummary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Company  string `json:"company"`
	Location string `json:"location,omitempty"`
	PostedAt int64  `json:"posted_at"`
	Remote   string `json:"remote,omitempty"`
}

// InvitationSummary is a condensed invitation representation.
type InvitationSummary struct {
	ID        string `json:"id"`
	Direction string `json:"direction"` // "received" or "sent"
	FromURN   string `json:"from_urn"`
	FromName  string `json:"from_name,omitempty"`
	Message   string `json:"message,omitempty"`
	SentAt    int64  `json:"sent_at"`
}

// NotificationSummary is a condensed notification representation.
type NotificationSummary struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
	IsRead    bool   `json:"is_read"`
}

// SearchResult is a generic search result.
type SearchResult struct {
	URN      string `json:"urn"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
	Type     string `json:"type"` // "person", "company", "job", "post", "group"
}

// GroupSummary is a condensed group representation.
type GroupSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	MemberCount int    `json:"member_count"`
	Description string `json:"description,omitempty"`
}

// EventSummary is a condensed event representation.
type EventSummary struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	StartsAt int64  `json:"starts_at"`
	Location string `json:"location,omitempty"`
}

// SkillSummary is a condensed skill representation.
type SkillSummary struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	EndorsementCount int    `json:"endorsement_count"`
}

// AnalyticsProfileViews holds profile view analytics.
type AnalyticsProfileViews struct {
	TotalViews int    `json:"total_views"`
	TimePeriod string `json:"time_period"`
}

// errEndpointDeprecated is returned for endpoints LinkedIn has migrated to SSR,
// making them unavailable via the Voyager API.
var errEndpointDeprecated = errors.New("this endpoint has been deprecated by LinkedIn and is no longer available via API")

// truncate shortens s to at most max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// formatTimestamp converts a Unix epoch (milliseconds) to a human-readable string.
// LinkedIn timestamps are in milliseconds.
func formatTimestamp(epochMs int64) string {
	if epochMs == 0 {
		return "-"
	}
	return time.Unix(epochMs/1000, 0).Format("2006-01-02 15:04")
}

// formatCount formats a large integer as a short human-readable string.
func formatCount(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

// confirmDestructive returns an error if --confirm flag is absent or false.
func confirmDestructive(cmd *cobra.Command) error {
	confirmed, _ := cmd.Flags().GetBool("confirm")
	if !confirmed {
		return fmt.Errorf("this action is irreversible; re-run with --confirm to proceed")
	}
	return nil
}

// dryRunResult prints a dry-run response and returns nil.
func dryRunResult(cmd *cobra.Command, description string, data any) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(data)
	}
	fmt.Printf("[DRY RUN] %s\n", description)
	return nil
}

// printProfileSummaries outputs profile summaries as JSON or text.
func printProfileSummaries(cmd *cobra.Command, profiles []ProfileSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(profiles)
	}
	if len(profiles) == 0 {
		fmt.Println("No profiles found.")
		return nil
	}
	lines := make([]string, 0, len(profiles)+1)
	lines = append(lines, fmt.Sprintf("%-25s  %-20s  %-20s  %-40s", "PUBLIC ID", "FIRST NAME", "LAST NAME", "HEADLINE"))
	for _, p := range profiles {
		lines = append(lines, fmt.Sprintf("%-25s  %-20s  %-20s  %-40s",
			truncate(p.PublicID, 25),
			truncate(p.FirstName, 20),
			truncate(p.LastName, 20),
			truncate(p.Headline, 40),
		))
	}
	cli.PrintText(lines)
	return nil
}

// printPostSummaries outputs post summaries as JSON or text.
func printPostSummaries(cmd *cobra.Command, posts []PostSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(posts)
	}
	if len(posts) == 0 {
		fmt.Println("No posts found.")
		return nil
	}
	lines := make([]string, 0, len(posts)+1)
	lines = append(lines, fmt.Sprintf("%-40s  %-50s  %-12s  %-8s  %-8s", "URN", "TEXT", "DATE", "LIKES", "COMMENTS"))
	for _, p := range posts {
		lines = append(lines, fmt.Sprintf("%-40s  %-50s  %-12s  %-8s  %-8s",
			truncate(p.URN, 40),
			truncate(p.Text, 50),
			formatTimestamp(p.Timestamp),
			formatCount(p.LikeCount),
			formatCount(p.CommentCount),
		))
	}
	cli.PrintText(lines)
	return nil
}
