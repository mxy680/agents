package instagram

import (
	"fmt"
	"time"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// UserSummary is a condensed user representation.
type UserSummary struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	FullName      string `json:"full_name"`
	ProfilePicURL string `json:"profile_pic_url"`
	IsPrivate     bool   `json:"is_private"`
	IsVerified    bool   `json:"is_verified"`
}

// UserDetail extends UserSummary with full profile info.
type UserDetail struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	FullName        string `json:"full_name"`
	ProfilePicURL   string `json:"profile_pic_url"`
	IsPrivate       bool   `json:"is_private"`
	IsVerified      bool   `json:"is_verified"`
	Biography       string `json:"biography"`
	ExternalURL     string `json:"external_url,omitempty"`
	FollowerCount   int64  `json:"follower_count"`
	FollowingCount  int64  `json:"following_count"`
	MediaCount      int64  `json:"media_count"`
	IsBusiness      bool   `json:"is_business"`
	Category        string `json:"category,omitempty"`
	AccountType     int    `json:"account_type"`
	IsProfessional  bool   `json:"is_professional_account"`
	TotalClipsCount int64  `json:"total_clips_count"`
	HasProfilePic   bool   `json:"has_profile_pic"`
}

// ProfileFormData is the edit profile form structure from /api/v1/accounts/edit/web_form_data/.
type ProfileFormData struct {
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	Email            string `json:"email"`
	Username         string `json:"username"`
	PhoneNumber      string `json:"phone_number"`
	Gender           int    `json:"gender"`
	Biography        string `json:"biography"`
	ExternalURL      string `json:"external_url"`
	IsEmailConfirmed bool   `json:"is_email_confirmed"`
	IsPhoneConfirmed bool   `json:"is_phone_confirmed"`
	BusinessAccount  bool   `json:"business_account"`
}

// MediaSummary is a condensed media/post representation.
type MediaSummary struct {
	ID           string `json:"id"`
	Shortcode    string `json:"shortcode,omitempty"`
	MediaType    int    `json:"media_type"`
	Caption      string `json:"caption,omitempty"`
	Timestamp    int64  `json:"timestamp"`
	LikeCount    int64  `json:"like_count"`
	CommentCount int64  `json:"comment_count"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// StorySummary is a condensed story representation.
type StorySummary struct {
	ID          string `json:"id"`
	MediaType   int    `json:"media_type"`
	Timestamp   int64  `json:"taken_at"`
	ExpiresAt   int64  `json:"expiring_at"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// ReelSummary is a condensed reel (clip) representation.
type ReelSummary struct {
	ID           string `json:"id"`
	Shortcode    string `json:"shortcode,omitempty"`
	Caption      string `json:"caption,omitempty"`
	Timestamp    int64  `json:"timestamp"`
	LikeCount    int64  `json:"like_count"`
	CommentCount int64  `json:"comment_count"`
	PlayCount    int64  `json:"play_count"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// DirectThreadSummary is a condensed direct message thread.
type DirectThreadSummary struct {
	ThreadID      string `json:"thread_id"`
	ThreadTitle   string `json:"thread_title"`
	LastActivity  int64  `json:"last_activity_at"`
	IsGroup       bool   `json:"is_group"`
	UnseenCount   int    `json:"unseen_count"`
}

// CommentSummary is a condensed comment representation.
type CommentSummary struct {
	PK          string `json:"pk"`
	Text        string `json:"text"`
	Timestamp   int64  `json:"created_at"`
	LikeCount   int64  `json:"comment_like_count"`
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
}

// ActivityItem represents a single notification/activity entry.
type ActivityItem struct {
	PK          string `json:"pk"`
	Type        int    `json:"type"`
	Timestamp   int64  `json:"timestamp"`
	Text        string `json:"text"`
	ProfileID   string `json:"profile_id"`
	ProfileName string `json:"profile_name"`
}

// CollectionSummary is a condensed saved-media collection.
type CollectionSummary struct {
	CollectionID   string `json:"collection_id"`
	CollectionName string `json:"collection_name"`
	CollectionType string `json:"collection_type"`
	MediaCount     int64  `json:"media_count"`
	CoverMediaURL  string `json:"cover_media_url,omitempty"`
}

// TagSummary is a condensed hashtag representation.
type TagSummary struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	MediaCount     int64  `json:"media_count"`
	FollowingCount int64  `json:"following_count"`
	IsFollowing    bool   `json:"is_following"`
}

// LocationSummary is a condensed location/place representation.
type LocationSummary struct {
	PK        int64   `json:"pk"`
	Name      string  `json:"name"`
	Address   string  `json:"address,omitempty"`
	City      string  `json:"city,omitempty"`
	Lat       float64 `json:"lat,omitempty"`
	Lng       float64 `json:"lng,omitempty"`
	MediaCount int64  `json:"media_count"`
}

// LiveBroadcast is a condensed live video broadcast representation.
type LiveBroadcast struct {
	ID             string `json:"id"`
	BroadcastStatus string `json:"broadcast_status"`
	CoverFrameURL  string `json:"cover_frame_url,omitempty"`
	ViewerCount    int64  `json:"viewer_count"`
	StartedAt      int64  `json:"published_time"`
}

// HighlightSummary is a condensed story highlight representation.
type HighlightSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	MediaCount  int    `json:"media_count"`
	CoverURL    string `json:"cover_url,omitempty"`
	CreatedAt   int64  `json:"created_at"`
}

// truncate shortens s to at most max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// formatTimestamp converts a Unix epoch to a human-readable date/time string.
func formatTimestamp(epoch int64) string {
	if epoch == 0 {
		return "-"
	}
	return time.Unix(epoch, 0).Format("2006-01-02 15:04")
}

// formatCount formats a large integer as a short human-readable string (e.g. "1.2K", "3.4M").
func formatCount(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

// confirmDestructive returns an error if the --confirm flag is absent or false.
func confirmDestructive(cmd *cobra.Command) error {
	confirmed, _ := cmd.Flags().GetBool("confirm")
	if !confirmed {
		return fmt.Errorf("this action is irreversible; re-run with --confirm to proceed")
	}
	return nil
}

// dryRunResult prints a standardised dry-run response and returns nil.
func dryRunResult(cmd *cobra.Command, description string, data any) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(data)
	}
	fmt.Printf("[DRY RUN] %s\n", description)
	return nil
}

// printUserSummaries outputs user summaries as JSON or a formatted text table.
func printUserSummaries(cmd *cobra.Command, users []UserSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(users)
	}

	if len(users) == 0 {
		fmt.Println("No users found.")
		return nil
	}

	lines := make([]string, 0, len(users)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-30s  %-10s  %-10s", "USERNAME", "FULL NAME", "PRIVATE", "VERIFIED"))
	for _, u := range users {
		username := truncate(u.Username, 20)
		fullName := truncate(u.FullName, 30)
		lines = append(lines, fmt.Sprintf("%-20s  %-30s  %-10v  %-10v", username, fullName, u.IsPrivate, u.IsVerified))
	}
	cli.PrintText(lines)
	return nil
}

// printMediaSummaries outputs media summaries as JSON or a formatted text table.
func printMediaSummaries(cmd *cobra.Command, media []MediaSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(media)
	}

	if len(media) == 0 {
		fmt.Println("No media found.")
		return nil
	}

	lines := make([]string, 0, len(media)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-40s  %-12s  %-10s  %-10s", "ID", "CAPTION", "DATE", "LIKES", "COMMENTS"))
	for _, m := range media {
		caption := truncate(m.Caption, 40)
		lines = append(lines, fmt.Sprintf("%-20s  %-40s  %-12s  %-10s  %-10s",
			truncate(m.ID, 20),
			caption,
			formatTimestamp(m.Timestamp),
			formatCount(m.LikeCount),
			formatCount(m.CommentCount),
		))
	}
	cli.PrintText(lines)
	return nil
}
