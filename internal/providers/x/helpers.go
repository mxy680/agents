package x

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// TweetSummary is a condensed X tweet representation.
type TweetSummary struct {
	ID             string `json:"id"`
	Text           string `json:"text"`
	AuthorID       string `json:"author_id"`
	AuthorName     string `json:"author_name,omitempty"`
	AuthorUsername string `json:"author_username,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	LikeCount      int    `json:"like_count"`
	RetweetCount   int    `json:"retweet_count"`
	ReplyCount     int    `json:"reply_count"`
	QuoteCount     int    `json:"quote_count"`
	ViewCount      int    `json:"view_count"`
	IsRetweet      bool   `json:"is_retweet,omitempty"`
}

// UserSummary is a condensed X user representation.
type UserSummary struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Username        string `json:"username"`
	Description     string `json:"description,omitempty"`
	Location        string `json:"location,omitempty"`
	Verified        bool   `json:"verified"`
	FollowersCount  int    `json:"followers_count"`
	FollowingCount  int    `json:"following_count"`
	TweetCount      int    `json:"tweet_count"`
	ProfileImageURL string `json:"profile_image_url,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
}

// truncate shortens s to at most max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// confirmDestructive returns an error if --confirm flag is absent or false.
func confirmDestructive(cmd *cobra.Command, msg string) error {
	confirmed, _ := cmd.Flags().GetBool("confirm")
	if !confirmed {
		return fmt.Errorf("%s; re-run with --confirm to proceed", msg)
	}
	return nil
}

// dryRunResult prints a dry-run preview and returns nil.
func dryRunResult(cmd *cobra.Command, action string, data any) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(data)
	}
	fmt.Printf("[DRY RUN] %s\n", action)
	return nil
}

// xTweetLegacy is the nested "legacy" object inside a tweet GraphQL result.
type xTweetLegacy struct {
	FullText          string `json:"full_text"`
	FavoriteCount     int    `json:"favorite_count"`
	RetweetCount      int    `json:"retweet_count"`
	ReplyCount        int    `json:"reply_count"`
	QuoteCount        int    `json:"quote_count"`
	CreatedAt         string `json:"created_at"`
	RetweetedStatusID string `json:"retweeted_status_id_str"`
}

// xUserLegacy is the nested "legacy" object inside a user GraphQL result.
type xUserLegacy struct {
	ScreenName            string `json:"screen_name"`
	Name                  string `json:"name"`
	Description           string `json:"description"`
	Location              string `json:"location"`
	Verified              bool   `json:"verified"`
	FollowersCount        int    `json:"followers_count"`
	FriendsCount          int    `json:"friends_count"`
	StatusesCount         int    `json:"statuses_count"`
	ProfileImageURLHTTPS  string `json:"profile_image_url_https"`
	CreatedAt             string `json:"created_at"`
}

// parseTweetResult parses X's GraphQL tweet_results nested format into TweetSummary.
// The input raw is expected to be the full tweetResult object, e.g.:
//
//	{ "result": { "__typename": "Tweet", "rest_id": "123", "legacy": {...}, "core": {...} } }
func parseTweetResult(raw json.RawMessage) (*TweetSummary, error) {
	var wrapper struct {
		Result struct {
			TypeName string        `json:"__typename"`
			RestID   string        `json:"rest_id"`
			Legacy   xTweetLegacy  `json:"legacy"`
			Core     struct {
				UserResults struct {
					Result struct {
						RestID string      `json:"rest_id"`
						Legacy xUserLegacy `json:"legacy"`
					} `json:"result"`
				} `json:"user_results"`
			} `json:"core"`
			Views struct {
				Count string `json:"count"`
			} `json:"views"`
		} `json:"result"`
	}

	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, fmt.Errorf("parse tweet result: %w", err)
	}

	r := wrapper.Result
	viewCount := 0
	if r.Views.Count != "" {
		fmt.Sscanf(r.Views.Count, "%d", &viewCount) //nolint:errcheck
	}

	tweet := &TweetSummary{
		ID:             r.RestID,
		Text:           r.Legacy.FullText,
		AuthorID:       r.Core.UserResults.Result.RestID,
		AuthorName:     r.Core.UserResults.Result.Legacy.Name,
		AuthorUsername: r.Core.UserResults.Result.Legacy.ScreenName,
		CreatedAt:      r.Legacy.CreatedAt,
		LikeCount:      r.Legacy.FavoriteCount,
		RetweetCount:   r.Legacy.RetweetCount,
		ReplyCount:     r.Legacy.ReplyCount,
		QuoteCount:     r.Legacy.QuoteCount,
		ViewCount:      viewCount,
		IsRetweet:      r.Legacy.RetweetedStatusID != "",
	}
	return tweet, nil
}

// parseUserResult parses X's GraphQL user_results nested format into UserSummary.
// The input raw is expected to be the result object directly, e.g.:
//
//	{ "__typename": "User", "rest_id": "123", "legacy": {...}, "is_blue_verified": true }
func parseUserResult(raw json.RawMessage) (*UserSummary, error) {
	var result struct {
		TypeName      string      `json:"__typename"`
		RestID        string      `json:"rest_id"`
		Legacy        xUserLegacy `json:"legacy"`
		IsBlueVerified bool       `json:"is_blue_verified"`
	}

	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse user result: %w", err)
	}

	user := &UserSummary{
		ID:              result.RestID,
		Name:            result.Legacy.Name,
		Username:        result.Legacy.ScreenName,
		Description:     result.Legacy.Description,
		Location:        result.Legacy.Location,
		Verified:        result.Legacy.Verified || result.IsBlueVerified,
		FollowersCount:  result.Legacy.FollowersCount,
		FollowingCount:  result.Legacy.FriendsCount,
		TweetCount:      result.Legacy.StatusesCount,
		ProfileImageURL: result.Legacy.ProfileImageURLHTTPS,
		CreatedAt:       result.Legacy.CreatedAt,
	}
	return user, nil
}

// extractTimelineEntries extracts tweet entries and the bottom cursor value
// from a GraphQL timeline response. The raw input is the full data payload.
// It traverses instructions → TimelineAddEntries → entries, collecting tweet
// entries and extracting the Bottom cursor.
func extractTimelineEntries(raw json.RawMessage) ([]json.RawMessage, string, error) {
	// The timeline data can be nested under various top-level keys.
	// We unmarshal into a generic map and search for "instructions".
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return nil, "", fmt.Errorf("parse timeline data: %w", err)
	}

	// Find "instructions" array by walking the nested structure.
	instructionsRaw, err := findInstructions(top)
	if err != nil {
		return nil, "", err
	}

	var instructions []struct {
		Type    string            `json:"type"`
		Entries []json.RawMessage `json:"entries"`
	}
	if err := json.Unmarshal(instructionsRaw, &instructions); err != nil {
		return nil, "", fmt.Errorf("parse timeline instructions: %w", err)
	}

	var tweetEntries []json.RawMessage
	cursor := ""

	for _, instr := range instructions {
		if instr.Type != "TimelineAddEntries" {
			continue
		}
		for _, entryRaw := range instr.Entries {
			var entry struct {
				EntryID string `json:"entryId"`
				Content struct {
					EntryType  string `json:"entryType"`
					Value      string `json:"value"`
					CursorType string `json:"cursorType"`
					ItemContent struct {
						ItemType     string          `json:"itemType"`
						TweetResults json.RawMessage `json:"tweet_results"`
					} `json:"itemContent"`
				} `json:"content"`
			}
			if err := json.Unmarshal(entryRaw, &entry); err != nil {
				continue
			}

			switch entry.Content.EntryType {
			case "TimelineTimelineCursor":
				if entry.Content.CursorType == "Bottom" {
					cursor = entry.Content.Value
				}
			case "TimelineTimelineItem":
				if entry.Content.ItemContent.ItemType == "TimelineTweet" &&
					entry.Content.ItemContent.TweetResults != nil {
					tweetEntries = append(tweetEntries, entry.Content.ItemContent.TweetResults)
				}
			}
		}
	}

	return tweetEntries, cursor, nil
}

// findInstructions walks a nested map looking for an "instructions" key inside
// a "timeline" object (or directly at the top level). Depth is bounded to
// prevent stack overflow from malformed responses.
func findInstructions(data map[string]json.RawMessage) (json.RawMessage, error) {
	return findInstructionsDepth(data, 4)
}

func findInstructionsDepth(data map[string]json.RawMessage, maxDepth int) (json.RawMessage, error) {
	// Try direct key "instructions".
	if raw, ok := data["instructions"]; ok {
		return raw, nil
	}

	// Try nested under "timeline".
	if timelineRaw, ok := data["timeline"]; ok {
		var timeline map[string]json.RawMessage
		if err := json.Unmarshal(timelineRaw, &timeline); err == nil {
			if raw, ok := timeline["instructions"]; ok {
				return raw, nil
			}
		}
	}

	if maxDepth <= 0 {
		return nil, fmt.Errorf("timeline instructions not found in response")
	}

	// Walk one level deeper into any nested object.
	for _, v := range data {
		var nested map[string]json.RawMessage
		if err := json.Unmarshal(v, &nested); err != nil {
			continue
		}
		if found, err := findInstructionsDepth(nested, maxDepth-1); err == nil {
			return found, nil
		}
	}

	return nil, fmt.Errorf("timeline instructions not found in response")
}

// printTweetSummaries outputs tweet summaries as JSON or text.
func printTweetSummaries(cmd *cobra.Command, tweets []TweetSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(tweets)
	}
	if len(tweets) == 0 {
		fmt.Println("No tweets found.")
		return nil
	}
	lines := make([]string, 0, len(tweets)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-20s  %-60s  %-8s  %-8s", "ID", "AUTHOR", "TEXT", "LIKES", "RTS"))
	for _, t := range tweets {
		author := t.AuthorUsername
		if author == "" {
			author = t.AuthorID
		}
		lines = append(lines, fmt.Sprintf("%-20s  %-20s  %-60s  %-8d  %-8d",
			truncate(t.ID, 20),
			truncate(author, 20),
			truncate(t.Text, 60),
			t.LikeCount,
			t.RetweetCount,
		))
	}
	cli.PrintText(lines)
	return nil
}

// printUserSummaries outputs user summaries as JSON or text.
func printUserSummaries(cmd *cobra.Command, users []UserSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(users)
	}
	if len(users) == 0 {
		fmt.Println("No users found.")
		return nil
	}
	lines := make([]string, 0, len(users)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-25s  %-20s  %-10s  %-10s", "ID", "NAME", "USERNAME", "FOLLOWERS", "FOLLOWING"))
	for _, u := range users {
		lines = append(lines, fmt.Sprintf("%-20s  %-25s  %-20s  %-10d  %-10d",
			truncate(u.ID, 20),
			truncate(u.Name, 25),
			truncate(u.Username, 20),
			u.FollowersCount,
			u.FollowingCount,
		))
	}
	cli.PrintText(lines)
	return nil
}

// voyagerPaging is a re-export alias for pagination state (unused here but kept
// for structural consistency with other providers).
type voyagerPaging struct {
	Start int `json:"start"`
	Count int `json:"count"`
	Total int `json:"total"`
}

// containsAny returns true if s contains any of the substrings.
func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

// DMConversationSummary is a condensed representation of an X DM conversation.
type DMConversationSummary struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Participants []string `json:"participants"`
}

// DMMessageSummary is a condensed representation of an X direct message.
type DMMessageSummary struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
	Text           string `json:"text"`
	Timestamp      string `json:"timestamp"`
}

// ListSummary is a condensed representation of an X list.
type ListSummary struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	MemberCount     int    `json:"member_count"`
	SubscriberCount int    `json:"subscriber_count"`
	Private         bool   `json:"private"`
	CreatedAt       string `json:"created_at,omitempty"`
	OwnerName       string `json:"owner_name,omitempty"`
	OwnerUsername   string `json:"owner_username,omitempty"`
}

// parseListResult parses X's GraphQL list result format into a ListSummary.
// The input raw is the value of the "list" key in the data response.
func parseListResult(raw json.RawMessage) (*ListSummary, error) {
	var result struct {
		IDStr           string `json:"id_str"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		MemberCount     int    `json:"member_count"`
		SubscriberCount int    `json:"subscriber_count"`
		Mode            string `json:"mode"`
		CreatedAt       int64  `json:"created_at"`
		UserResults     struct {
			Result struct {
				Legacy xUserLegacy `json:"legacy"`
			} `json:"result"`
		} `json:"user_results"`
	}

	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("parse list result: %w", err)
	}

	createdAt := ""
	if result.CreatedAt > 0 {
		createdAt = fmt.Sprintf("%d", result.CreatedAt)
	}

	return &ListSummary{
		ID:              result.IDStr,
		Name:            result.Name,
		Description:     result.Description,
		MemberCount:     result.MemberCount,
		SubscriberCount: result.SubscriberCount,
		Private:         result.Mode == "Private",
		CreatedAt:       createdAt,
		OwnerName:       result.UserResults.Result.Legacy.Name,
		OwnerUsername:   result.UserResults.Result.Legacy.ScreenName,
	}, nil
}
