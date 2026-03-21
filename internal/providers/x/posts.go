package x

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// GraphQL query hashes for tweet/post operations.
const (
	hashTweetResultByRestId      = "Xl5pC_lBk_gcO2ItU39DQw"
	hashTweetResultsByRestIds    = "PTN9HhBAlpoCTHfspDgqLA"
	hashCreateTweet              = "SiM_cAu83R0wnrpmKQQSEw"
	hashDeleteTweet              = "VaenaVgh5q5ih7kvyVjgtg"
	hashSearchTimeline           = "flaR-PUMshxFWZWPNpq4zA"
	hashHomeTimeline             = "-X_hcgQzmHGl29-UXxz4sw"
	hashHomeLatestTimeline       = "U0cdisy7QFIoTfu3-Okw0A"
	hashUserTweets               = "QWF3SzpHmykQHsQMixG0cg"
	hashUserTweetsAndReplies     = "vMkJyzx1wdmvOeeNG0n6Wg"
	hashRetweeters               = "X-XEqG5qHQSAwmvy00xfyQ"
	hashFavoriters               = "LLkw5EcVutJL6y-2gkz22A"
	hashSimilarPosts             = "EToazR74i0rJyZYalfVEAQ"
)

// newPostsCmd builds the "posts" subcommand group.
func newPostsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "posts",
		Short:   "Manage tweets and posts",
		Aliases: []string{"post", "tweet"},
	}
	cmd.AddCommand(newPostsGetCmd(factory))
	cmd.AddCommand(newPostsLookupCmd(factory))
	cmd.AddCommand(newPostsCreateCmd(factory))
	cmd.AddCommand(newPostsDeleteCmd(factory))
	cmd.AddCommand(newPostsSearchCmd(factory))
	cmd.AddCommand(newPostsTimelineCmd(factory))
	cmd.AddCommand(newPostsLatestTimelineCmd(factory))
	cmd.AddCommand(newPostsUserTweetsCmd(factory))
	cmd.AddCommand(newPostsUserRepliesCmd(factory))
	cmd.AddCommand(newPostsRetweetersCmd(factory))
	cmd.AddCommand(newPostsFavoritersCmd(factory))
	cmd.AddCommand(newPostsSimilarCmd(factory))
	return cmd
}

// newPostsGetCmd builds the "posts get" command.
func newPostsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a tweet by ID",
		RunE:  makeRunPostsGet(factory),
	}
	cmd.Flags().String("tweet-id", "", "Tweet ID (required)")
	_ = cmd.MarkFlagRequired("tweet-id")
	return cmd
}

// newPostsLookupCmd builds the "posts lookup" command.
func newPostsLookupCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lookup",
		Short: "Look up multiple tweets by ID",
		RunE:  makeRunPostsLookup(factory),
	}
	cmd.Flags().StringSlice("tweet-ids", nil, "Comma-separated tweet IDs (required)")
	_ = cmd.MarkFlagRequired("tweet-ids")
	return cmd
}

// newPostsCreateCmd builds the "posts create" command.
func newPostsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new tweet",
		RunE:  makeRunPostsCreate(factory),
	}
	cmd.Flags().String("text", "", "Tweet text content (required)")
	_ = cmd.MarkFlagRequired("text")
	cmd.Flags().String("reply-to", "", "Tweet ID to reply to")
	cmd.Flags().String("quote-url", "", "URL of tweet to quote-tweet")
	cmd.Flags().Bool("dry-run", false, "Print what would be sent without creating the tweet")
	return cmd
}

// newPostsDeleteCmd builds the "posts delete" command.
func newPostsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a tweet",
		RunE:  makeRunPostsDelete(factory),
	}
	cmd.Flags().String("tweet-id", "", "Tweet ID to delete (required)")
	_ = cmd.MarkFlagRequired("tweet-id")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Print what would be deleted without deleting")
	return cmd
}

// newPostsSearchCmd builds the "posts search" command.
func newPostsSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for tweets",
		RunE:  makeRunPostsSearch(factory),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	_ = cmd.MarkFlagRequired("query")
	cmd.Flags().String("product", "Latest", "Search product: Top, Latest, People, Photos, Videos")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newPostsTimelineCmd builds the "posts timeline" command.
func newPostsTimelineCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "timeline",
		Short: "Get the home timeline",
		RunE:  makeRunPostsTimeline(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of tweets")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newPostsLatestTimelineCmd builds the "posts latest-timeline" command.
func newPostsLatestTimelineCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latest-timeline",
		Short: "Get the home latest (chronological) timeline",
		RunE:  makeRunPostsLatestTimeline(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of tweets")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newPostsUserTweetsCmd builds the "posts user-tweets" command.
func newPostsUserTweetsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-tweets",
		Short: "Get tweets from a user",
		RunE:  makeRunPostsUserTweets(factory),
	}
	cmd.Flags().String("user-id", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Int("limit", 20, "Maximum number of tweets")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newPostsUserRepliesCmd builds the "posts user-replies" command.
func newPostsUserRepliesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-replies",
		Short: "Get tweets and replies from a user",
		RunE:  makeRunPostsUserReplies(factory),
	}
	cmd.Flags().String("user-id", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Int("limit", 20, "Maximum number of tweets")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newPostsRetweetersCmd builds the "posts retweeters" command.
func newPostsRetweetersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retweeters",
		Short: "Get users who retweeted a tweet",
		RunE:  makeRunPostsRetweeters(factory),
	}
	cmd.Flags().String("tweet-id", "", "Tweet ID (required)")
	_ = cmd.MarkFlagRequired("tweet-id")
	cmd.Flags().Int("limit", 20, "Maximum number of users")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newPostsFavoritersCmd builds the "posts favoriters" command.
func newPostsFavoritersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "favoriters",
		Short: "Get users who liked a tweet",
		RunE:  makeRunPostsFavoriters(factory),
	}
	cmd.Flags().String("tweet-id", "", "Tweet ID (required)")
	_ = cmd.MarkFlagRequired("tweet-id")
	cmd.Flags().Int("limit", 20, "Maximum number of users")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newPostsSimilarCmd builds the "posts similar" command.
func newPostsSimilarCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "similar",
		Short: "Get posts similar to a tweet",
		RunE:  makeRunPostsSimilar(factory),
	}
	cmd.Flags().String("tweet-id", "", "Tweet ID (required)")
	_ = cmd.MarkFlagRequired("tweet-id")
	return cmd
}

// --- RunE implementations ---

func makeRunPostsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tweetID, _ := cmd.Flags().GetString("tweet-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"tweetId":                              tweetID,
			"referrer":                             "tweet",
			"includePromotedContent":               false,
			"withCommunity":                        true,
			"withQuickPromoteEligibilityTweetFields": true,
			"withBirdwatchNotes":                   true,
			"withVoice":                            true,
			"withV2Timeline":                       true,
		}

		data, err := client.GraphQL(ctx, hashTweetResultByRestId, "TweetResultByRestId", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("getting tweet %s: %w", tweetID, err)
		}

		var payload struct {
			TweetResult json.RawMessage `json:"tweetResult"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("parse tweet response: %w", err)
		}

		tweet, err := parseTweetResult(payload.TweetResult)
		if err != nil {
			return fmt.Errorf("parse tweet: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(tweet)
		}

		lines := []string{
			fmt.Sprintf("ID:       %s", tweet.ID),
			fmt.Sprintf("Author:   %s (@%s)", tweet.AuthorName, tweet.AuthorUsername),
			fmt.Sprintf("Created:  %s", tweet.CreatedAt),
			fmt.Sprintf("Likes:    %d", tweet.LikeCount),
			fmt.Sprintf("Retweets: %d", tweet.RetweetCount),
			fmt.Sprintf("Replies:  %d", tweet.ReplyCount),
			fmt.Sprintf("Views:    %d", tweet.ViewCount),
			fmt.Sprintf("Text:     %s", truncate(tweet.Text, 300)),
		}
		cli.PrintText(lines)
		return nil
	}
}

func makeRunPostsLookup(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tweetIDs, _ := cmd.Flags().GetStringSlice("tweet-ids")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"ids":                    tweetIDs,
			"includePromotedContent": false,
		}

		data, err := client.GraphQL(ctx, hashTweetResultsByRestIds, "TweetResultsByRestIds", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("looking up tweets: %w", err)
		}

		var payload struct {
			TweetResults []json.RawMessage `json:"tweetResults"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("parse lookup response: %w", err)
		}

		tweets := make([]TweetSummary, 0, len(payload.TweetResults))
		for _, raw := range payload.TweetResults {
			tweet, err := parseTweetResult(raw)
			if err != nil {
				continue
			}
			tweets = append(tweets, *tweet)
		}

		return printTweetSummaries(cmd, tweets)
	}
}

func makeRunPostsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		text, _ := cmd.Flags().GetString("text")
		replyTo, _ := cmd.Flags().GetString("reply-to")
		quoteURL, _ := cmd.Flags().GetString("quote-url")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"tweet_text":              text,
			"dark_request":            false,
			"media":                   map[string]any{"media_entities": []any{}, "possibly_sensitive": false},
			"semantic_annotation_ids": []any{},
		}

		if replyTo != "" {
			vars["reply"] = map[string]any{
				"in_reply_to_tweet_id":  replyTo,
				"exclude_reply_user_ids": []any{},
			}
		}
		if quoteURL != "" {
			vars["attachment_url"] = quoteURL
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("create tweet with text %q", truncate(text, 60)), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := client.GraphQLPost(ctx, hashCreateTweet, "CreateTweet", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("creating tweet: %w", err)
		}

		var payload struct {
			CreateTweet struct {
				TweetResults json.RawMessage `json:"tweet_results"`
			} `json:"create_tweet"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("parse create tweet response: %w", err)
		}

		tweet, err := parseTweetResult(payload.CreateTweet.TweetResults)
		if err != nil {
			return fmt.Errorf("parse created tweet: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(tweet)
		}
		fmt.Printf("Tweet created: %s\n", tweet.ID)
		return nil
	}
}

func makeRunPostsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tweetID, _ := cmd.Flags().GetString("tweet-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("delete tweet %s", tweetID), map[string]string{"tweet_id": tweetID})
		}

		if err := confirmDestructive(cmd, "this action is irreversible"); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"tweet_id":     tweetID,
			"dark_request": false,
		}

		_, err = client.GraphQLPost(ctx, hashDeleteTweet, "DeleteTweet", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("deleting tweet %s: %w", tweetID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "tweet_id": tweetID})
		}
		fmt.Printf("Tweet deleted: %s\n", tweetID)
		return nil
	}
}

func makeRunPostsSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		product, _ := cmd.Flags().GetString("product")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"rawQuery":              query,
			"count":                 limit,
			"product":               product,
			"querySource":           "typed_query",
			"includePromotedContent": false,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashSearchTimeline, "SearchTimeline", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("searching tweets: %w", err)
		}

		rawEntries, nextCursor, err := extractTimelineEntries(data)
		if err != nil {
			return fmt.Errorf("extract search results: %w", err)
		}

		tweets := make([]TweetSummary, 0, len(rawEntries))
		for _, raw := range rawEntries {
			tweet, err := parseTweetResult(raw)
			if err != nil {
				continue
			}
			tweets = append(tweets, *tweet)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{
				"tweets":      tweets,
				"next_cursor": nextCursor,
			})
		}
		if nextCursor != "" {
			fmt.Printf("Next cursor: %s\n", nextCursor)
		}
		return printTweetSummaries(cmd, tweets)
	}
}

func makeRunPostsTimeline(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"count":                  limit,
			"includePromotedContent": false,
			"latestControlAvailable": true,
			"requestContext":         "launch",
			"withCommunity":          true,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashHomeTimeline, "HomeTimeline", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching home timeline: %w", err)
		}

		return printTimelineResult(cmd, data)
	}
}

func makeRunPostsLatestTimeline(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"count":                  limit,
			"includePromotedContent": false,
			"latestControlAvailable": true,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashHomeLatestTimeline, "HomeLatestTimeline", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching latest timeline: %w", err)
		}

		return printTimelineResult(cmd, data)
	}
}

func makeRunPostsUserTweets(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"userId":                     userID,
			"count":                      limit,
			"includePromotedContent":     false,
			"withQuickPromoteEligibilityTweetFields": true,
			"withVoice":                  true,
			"withV2Timeline":             true,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashUserTweets, "UserTweets", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching user tweets: %w", err)
		}

		return printTimelineResult(cmd, data)
	}
}

func makeRunPostsUserReplies(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"userId":                 userID,
			"count":                  limit,
			"includePromotedContent": false,
			"withCommunity":          true,
			"withVoice":              true,
			"withV2Timeline":         true,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashUserTweetsAndReplies, "UserTweetsAndReplies", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching user tweets and replies: %w", err)
		}

		return printTimelineResult(cmd, data)
	}
}

func makeRunPostsRetweeters(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tweetID, _ := cmd.Flags().GetString("tweet-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"tweetId": tweetID,
			"count":   limit,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashRetweeters, "Retweeters", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching retweeters for %s: %w", tweetID, err)
		}

		return printUserListResult(cmd, data)
	}
}

func makeRunPostsFavoriters(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tweetID, _ := cmd.Flags().GetString("tweet-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"tweetId": tweetID,
			"count":   limit,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashFavoriters, "Favoriters", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching favoriters for %s: %w", tweetID, err)
		}

		return printUserListResult(cmd, data)
	}
}

func makeRunPostsSimilar(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tweetID, _ := cmd.Flags().GetString("tweet-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"tweet_id": tweetID,
		}

		data, err := client.GraphQL(ctx, hashSimilarPosts, "SimilarPosts", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching similar posts for %s: %w", tweetID, err)
		}

		return printTimelineResult(cmd, data)
	}
}

// printTimelineResult extracts and prints tweets from a timeline GraphQL response.
func printTimelineResult(cmd *cobra.Command, data json.RawMessage) error {
	rawEntries, nextCursor, err := extractTimelineEntries(data)
	if err != nil {
		return fmt.Errorf("extract timeline entries: %w", err)
	}

	tweets := make([]TweetSummary, 0, len(rawEntries))
	for _, raw := range rawEntries {
		tweet, err := parseTweetResult(raw)
		if err != nil {
			continue
		}
		tweets = append(tweets, *tweet)
	}

	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(map[string]any{
			"tweets":      tweets,
			"next_cursor": nextCursor,
		})
	}
	if nextCursor != "" {
		fmt.Printf("Next cursor: %s\n", nextCursor)
	}
	return printTweetSummaries(cmd, tweets)
}

// printUserListResult extracts and prints users from a user-list GraphQL response
// (used for retweeters, favoriters).
func printUserListResult(cmd *cobra.Command, data json.RawMessage) error {
	// User list timelines follow the same instructions pattern but contain
	// user entries instead of tweet entries.
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return fmt.Errorf("parse user list data: %w", err)
	}

	instructionsRaw, err := findInstructions(top)
	if err != nil {
		return fmt.Errorf("find instructions: %w", err)
	}

	var instructions []struct {
		Type    string            `json:"type"`
		Entries []json.RawMessage `json:"entries"`
	}
	if err := json.Unmarshal(instructionsRaw, &instructions); err != nil {
		return fmt.Errorf("parse instructions: %w", err)
	}

	var users []UserSummary
	nextCursor := ""

	for _, instr := range instructions {
		if instr.Type != "TimelineAddEntries" {
			continue
		}
		for _, entryRaw := range instr.Entries {
			var entry struct {
				Content struct {
					EntryType  string `json:"entryType"`
					Value      string `json:"value"`
					CursorType string `json:"cursorType"`
					ItemContent struct {
						ItemType    string          `json:"itemType"`
						UserResults json.RawMessage `json:"user_results"`
					} `json:"itemContent"`
				} `json:"content"`
			}
			if err := json.Unmarshal(entryRaw, &entry); err != nil {
				continue
			}

			switch entry.Content.EntryType {
			case "TimelineTimelineCursor":
				if entry.Content.CursorType == "Bottom" {
					nextCursor = entry.Content.Value
				}
			case "TimelineTimelineItem":
				if entry.Content.ItemContent.ItemType == "TimelineUser" &&
					entry.Content.ItemContent.UserResults != nil {
					var wrapper struct {
						Result json.RawMessage `json:"result"`
					}
					if err := json.Unmarshal(entry.Content.ItemContent.UserResults, &wrapper); err != nil {
						continue
					}
					user, err := parseUserResult(wrapper.Result)
					if err != nil {
						continue
					}
					users = append(users, *user)
				}
			}
		}
	}

	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(map[string]any{
			"users":       users,
			"next_cursor": nextCursor,
		})
	}
	if nextCursor != "" {
		fmt.Printf("Next cursor: %s\n", nextCursor)
	}
	return printUserSummaries(cmd, users)
}
