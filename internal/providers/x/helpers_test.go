package x

import (
	"encoding/json"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		max   int
		want  string
	}{
		{name: "short string unchanged", input: "hello", max: 10, want: "hello"},
		{name: "exact length unchanged", input: "hello", max: 5, want: "hello"},
		{name: "truncated with ellipsis", input: "hello world", max: 8, want: "hello..."},
		{name: "empty string", input: "", max: 5, want: ""},
		{name: "unicode handled correctly", input: "héllo wörld", max: 8, want: "héllo..."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := truncate(tc.input, tc.max)
			if got != tc.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tc.input, tc.max, got, tc.want)
			}
		})
	}
}

func TestParseTweetResult(t *testing.T) {
	raw := mockTweetResult("123456789", "Hello X world!", "999", "Test User", "testuser")

	tweet, err := parseTweetResult(raw)
	if err != nil {
		t.Fatalf("parseTweetResult error: %v", err)
	}

	if tweet.ID != "123456789" {
		t.Errorf("expected ID=123456789, got %s", tweet.ID)
	}
	if tweet.Text != "Hello X world!" {
		t.Errorf("expected text='Hello X world!', got %s", tweet.Text)
	}
	if tweet.AuthorID != "999" {
		t.Errorf("expected author_id=999, got %s", tweet.AuthorID)
	}
	if tweet.AuthorName != "Test User" {
		t.Errorf("expected author_name='Test User', got %s", tweet.AuthorName)
	}
	if tweet.AuthorUsername != "testuser" {
		t.Errorf("expected author_username=testuser, got %s", tweet.AuthorUsername)
	}
	if tweet.LikeCount != 42 {
		t.Errorf("expected like_count=42, got %d", tweet.LikeCount)
	}
	if tweet.RetweetCount != 7 {
		t.Errorf("expected retweet_count=7, got %d", tweet.RetweetCount)
	}
	if tweet.ViewCount != 1500 {
		t.Errorf("expected view_count=1500, got %d", tweet.ViewCount)
	}
	if tweet.IsRetweet {
		t.Error("expected is_retweet=false for non-retweet")
	}
}

func TestParseTweetResult_Invalid(t *testing.T) {
	_, err := parseTweetResult(json.RawMessage(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParseUserResult(t *testing.T) {
	raw := mockUserResult("999", "testuser", "Test User")

	user, err := parseUserResult(raw)
	if err != nil {
		t.Fatalf("parseUserResult error: %v", err)
	}

	if user.ID != "999" {
		t.Errorf("expected ID=999, got %s", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("expected username=testuser, got %s", user.Username)
	}
	if user.Name != "Test User" {
		t.Errorf("expected name='Test User', got %s", user.Name)
	}
	if user.FollowersCount != 1000 {
		t.Errorf("expected followers_count=1000, got %d", user.FollowersCount)
	}
	if user.FollowingCount != 500 {
		t.Errorf("expected following_count=500, got %d", user.FollowingCount)
	}
	if !user.Verified {
		// is_blue_verified=true in mock
		t.Error("expected verified=true from is_blue_verified")
	}
}

func TestParseUserResult_Invalid(t *testing.T) {
	_, err := parseUserResult(json.RawMessage(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestExtractTimelineEntries(t *testing.T) {
	tweetRaw := mockTweetResult("123456789", "Hello X world!", "999", "Test User", "testuser")
	timelineResp := mockTimelineResponse(tweetRaw, "next-cursor-value")

	// mockTimelineResponse wraps in {"data": {...}} — extract the data field.
	var outer struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(timelineResp, &outer); err != nil {
		t.Fatalf("unmarshal outer: %v", err)
	}

	entries, cursor, err := extractTimelineEntries(outer.Data)
	if err != nil {
		t.Fatalf("extractTimelineEntries error: %v", err)
	}

	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
	if cursor != "next-cursor-value" {
		t.Errorf("expected cursor='next-cursor-value', got %q", cursor)
	}
}

func TestExtractTimelineEntries_Empty(t *testing.T) {
	// Timeline with no entries.
	raw, _ := json.Marshal(map[string]any{
		"timeline": map[string]any{
			"instructions": []any{
				map[string]any{
					"type":    "TimelineAddEntries",
					"entries": []any{},
				},
			},
		},
	})

	entries, cursor, err := extractTimelineEntries(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
	if cursor != "" {
		t.Errorf("expected empty cursor, got %q", cursor)
	}
}

func TestExtractTimelineEntries_NoInstructions(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{
		"some_other_key": "value",
	})

	_, _, err := extractTimelineEntries(raw)
	if err == nil {
		t.Error("expected error when instructions not found, got nil")
	}
}

func TestRateLimitError(t *testing.T) {
	e := &RateLimitError{}
	if e.Error() != "x rate limit exceeded" {
		t.Errorf("unexpected error string: %s", e.Error())
	}
}

func TestAccountLockedError(t *testing.T) {
	e := &AccountLockedError{Message: "suspicious activity"}
	if !containsAny(e.Error(), "account locked", "suspicious") {
		t.Errorf("unexpected error string: %s", e.Error())
	}
}
