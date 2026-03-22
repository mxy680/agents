package x

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// GraphQL query hashes for user operations.
const (
	hashUserByScreenName          = "NimuplG1OB7Fd2btCLdBOw"
	hashUserByRestId              = "tD8zKvQzwY3kdx5yz6YmOw"
	hashUserHighlightsTweets      = "tHFm_XZc_NNi-CfUThwbNw"
	hashUserMedia                 = "2tLOJWwGuCTytDrGBg8VwQ"
	hashUserCreatorSubscriptions  = "Wsm5ZTCYtg2eH7mXAXPIgw"
	// hashSearchTimeline is shared with posts.go — used with product=People here.
)

// newUsersCmd builds the "users" subcommand group.
func newUsersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "users",
		Short:   "Look up X users",
		Aliases: []string{"user"},
	}
	cmd.AddCommand(newUsersGetCmd(factory))
	cmd.AddCommand(newUsersGetByIDCmd(factory))
	cmd.AddCommand(newUsersSearchCmd(factory))
	cmd.AddCommand(newUsersHighlightsCmd(factory))
	cmd.AddCommand(newUsersMediaCmd(factory))
	cmd.AddCommand(newUsersSubscriptionsCmd(factory))
	return cmd
}

// newUsersGetCmd builds the "users get" command.
func newUsersGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a user by username (screen name)",
		RunE:  makeRunUsersGet(factory),
	}
	cmd.Flags().String("username", "", "X username / screen name (required)")
	_ = cmd.MarkFlagRequired("username")
	return cmd
}

// newUsersGetByIDCmd builds the "users get-by-id" command.
func newUsersGetByIDCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-by-id",
		Short: "Get a user by numeric user ID",
		RunE:  makeRunUsersGetByID(factory),
	}
	cmd.Flags().String("user-id", "", "Numeric user ID (required)")
	_ = cmd.MarkFlagRequired("user-id")
	return cmd
}

// newUsersSearchCmd builds the "users search" command.
func newUsersSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for users",
		RunE:  makeRunUsersSearch(factory),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	_ = cmd.MarkFlagRequired("query")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newUsersHighlightsCmd builds the "users highlights" command.
func newUsersHighlightsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "highlights",
		Short: "Get highlighted tweets from a user",
		RunE:  makeRunUsersHighlights(factory),
	}
	cmd.Flags().String("user-id", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Int("limit", 20, "Maximum number of tweets")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newUsersMediaCmd builds the "users media" command.
func newUsersMediaCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "media",
		Short: "Get media tweets from a user",
		RunE:  makeRunUsersMedia(factory),
	}
	cmd.Flags().String("user-id", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Int("limit", 20, "Maximum number of tweets")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newUsersSubscriptionsCmd builds the "users subscriptions" command.
func newUsersSubscriptionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscriptions",
		Short: "Get creator subscriptions for a user",
		RunE:  makeRunUsersSubscriptions(factory),
	}
	cmd.Flags().String("user-id", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// --- RunE implementations ---

func makeRunUsersGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		username, _ := cmd.Flags().GetString("username")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"screen_name":              username,
			"withSafetyModeUserFields": true,
		}

		data, err := client.GraphQL(ctx, hashUserByScreenName, "UserByScreenName", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("getting user @%s: %w", username, err)
		}

		user, err := parseUserByScreenNameResponse(data)
		if err != nil {
			return fmt.Errorf("parse user response: %w", err)
		}

		return printSingleUser(cmd, user)
	}
}

func makeRunUsersGetByID(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"userId":                   userID,
			"withSafetyModeUserFields": true,
		}

		data, err := client.GraphQL(ctx, hashUserByRestId, "UserByRestId", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("getting user by ID %s: %w", userID, err)
		}

		user, err := parseUserByScreenNameResponse(data)
		if err != nil {
			return fmt.Errorf("parse user response: %w", err)
		}

		return printSingleUser(cmd, user)
	}
}

func makeRunUsersSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"rawQuery":               query,
			"count":                  limit,
			"product":                "People",
			"querySource":            "typed_query",
			"includePromotedContent": false,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashSearchTimeline, "SearchTimeline", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("searching users: %w", err)
		}

		return printUserListResult(cmd, data)
	}
}

func makeRunUsersHighlights(factory ClientFactory) func(*cobra.Command, []string) error {
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
			"withVoice":              true,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashUserHighlightsTweets, "UserHighlightsTweets", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching highlights for user %s: %w", userID, err)
		}

		return printTimelineResult(cmd, data)
	}
}

func makeRunUsersMedia(factory ClientFactory) func(*cobra.Command, []string) error {
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
			"withClientEventToken":   false,
			"withBirdwatchNotes":     false,
			"withVoice":              true,
			"withV2Timeline":         true,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashUserMedia, "UserMedia", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching media for user %s: %w", userID, err)
		}

		return printTimelineResult(cmd, data)
	}
}

func makeRunUsersSubscriptions(factory ClientFactory) func(*cobra.Command, []string) error {
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
			"userId": userID,
			"count":  limit,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashUserCreatorSubscriptions, "UserCreatorSubscriptions", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching subscriptions for user %s: %w", userID, err)
		}

		return printUserListResult(cmd, data)
	}
}

// parseUserByScreenNameResponse extracts a UserSummary from a UserByScreenName or
// UserByRestId response, where the result is nested under data.user.result.
func parseUserByScreenNameResponse(data json.RawMessage) (*UserSummary, error) {
	var payload struct {
		User struct {
			Result json.RawMessage `json:"result"`
		} `json:"user"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("parse user payload: %w", err)
	}
	return parseUserResult(payload.User.Result)
}

// printSingleUser outputs a single user as JSON or formatted text.
func printSingleUser(cmd *cobra.Command, user *UserSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(user)
	}

	lines := []string{
		fmt.Sprintf("ID:          %s", user.ID),
		fmt.Sprintf("Name:        %s", user.Name),
		fmt.Sprintf("Username:    @%s", user.Username),
		fmt.Sprintf("Verified:    %v", user.Verified),
		fmt.Sprintf("Followers:   %d", user.FollowersCount),
		fmt.Sprintf("Following:   %d", user.FollowingCount),
		fmt.Sprintf("Tweets:      %d", user.TweetCount),
		fmt.Sprintf("Created:     %s", user.CreatedAt),
	}
	if user.Location != "" {
		lines = append(lines, fmt.Sprintf("Location:    %s", user.Location))
	}
	if user.Description != "" {
		lines = append(lines, fmt.Sprintf("Bio:         %s", truncate(user.Description, 200)))
	}
	cli.PrintText(lines)
	return nil
}
