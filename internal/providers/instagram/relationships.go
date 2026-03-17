package instagram

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// friendshipsListResponse is the response for followers/following endpoints.
type friendshipsListResponse struct {
	Users      []rawUser `json:"users"`
	NextMaxID  string    `json:"next_max_id"`
	BigList    bool      `json:"big_list"`
	Status     string    `json:"status"`
}

// friendshipActionResponse is a generic response for follow/unfollow/block actions.
type friendshipActionResponse struct {
	Friendship friendshipStatus `json:"friendship_status"`
	Status     string           `json:"status"`
}

// friendshipStatus represents the relationship status between users.
type friendshipStatus struct {
	Following       bool `json:"following"`
	FollowedBy      bool `json:"followed_by"`
	Blocking        bool `json:"blocking"`
	Muting          bool `json:"muting"`
	IsPrivate       bool `json:"is_private"`
	IncomingRequest bool `json:"incoming_request"`
	OutgoingRequest bool `json:"outgoing_request"`
	IsRestricted    bool `json:"is_restricted"`
}

// blockedUsersResponse is the response for GET /api/v1/users/blocked_list/.
type blockedUsersResponse struct {
	BlockedList []rawUser `json:"blocked_list"`
	Status      string    `json:"status"`
}

// RelationshipStatusResult is the output shape for the status command.
type RelationshipStatusResult struct {
	UserID          string `json:"user_id"`
	Following       bool   `json:"following"`
	FollowedBy      bool   `json:"followed_by"`
	Blocking        bool   `json:"blocking"`
	Muting          bool   `json:"muting"`
	IsPrivate       bool   `json:"is_private"`
	IncomingRequest bool   `json:"incoming_request"`
	OutgoingRequest bool   `json:"outgoing_request"`
	IsRestricted    bool   `json:"is_restricted"`
}

// newRelationshipsCmd builds the `relationships` subcommand group.
func newRelationshipsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "relationships",
		Short:   "Manage followers, following, and blocks",
		Aliases: []string{"rel", "friendship"},
	}
	cmd.AddCommand(newRelationshipsFollowersCmd(factory))
	cmd.AddCommand(newRelationshipsFollowingCmd(factory))
	cmd.AddCommand(newRelationshipsFollowCmd(factory))
	cmd.AddCommand(newRelationshipsUnfollowCmd(factory))
	cmd.AddCommand(newRelationshipsRemoveFollowerCmd(factory))
	cmd.AddCommand(newRelationshipsBlockCmd(factory))
	cmd.AddCommand(newRelationshipsUnblockCmd(factory))
	cmd.AddCommand(newRelationshipsBlockedCmd(factory))
	cmd.AddCommand(newRelationshipsMuteCmd(factory))
	cmd.AddCommand(newRelationshipsUnmuteCmd(factory))
	cmd.AddCommand(newRelationshipsRestrictCmd(factory))
	cmd.AddCommand(newRelationshipsUnrestrictCmd(factory))
	cmd.AddCommand(newRelationshipsStatusCmd(factory))
	return cmd
}

func newRelationshipsFollowersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "followers",
		Short: "List followers of a user",
		RunE:  makeRunRelationshipsFollowers(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to list followers for (defaults to own user)")
	cmd.Flags().Int("limit", 50, "Maximum number of followers to return")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	cmd.Flags().String("query", "", "Filter followers by username prefix")
	return cmd
}

func makeRunRelationshipsFollowers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")
		query, _ := cmd.Flags().GetString("query")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if userID == "" {
			userID = client.session.DSUserID
		}

		params := url.Values{}
		params.Set("count", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("max_id", cursor)
		}
		if query != "" {
			params.Set("query", query)
		}

		resp, err := client.Get(ctx, "/api/v1/friendships/"+userID+"/followers/", params)
		if err != nil {
			return fmt.Errorf("listing followers for user %s: %w", userID, err)
		}

		var result friendshipsListResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding followers response: %w", err)
		}

		summaries := toUserSummaries(result.Users)
		if err := printUserSummaries(cmd, summaries); err != nil {
			return err
		}
		if result.BigList && result.NextMaxID != "" {
			fmt.Printf("Next cursor: %s\n", result.NextMaxID)
		}
		return nil
	}
}

func newRelationshipsFollowingCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "following",
		Short: "List users a user is following",
		RunE:  makeRunRelationshipsFollowing(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to list following for (defaults to own user)")
	cmd.Flags().Int("limit", 50, "Maximum number of users to return")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	cmd.Flags().String("query", "", "Filter following by username prefix")
	return cmd
}

func makeRunRelationshipsFollowing(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")
		query, _ := cmd.Flags().GetString("query")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if userID == "" {
			userID = client.session.DSUserID
		}

		params := url.Values{}
		params.Set("count", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("max_id", cursor)
		}
		if query != "" {
			params.Set("query", query)
		}

		resp, err := client.Get(ctx, "/api/v1/friendships/"+userID+"/following/", params)
		if err != nil {
			return fmt.Errorf("listing following for user %s: %w", userID, err)
		}

		var result friendshipsListResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding following response: %w", err)
		}

		summaries := toUserSummaries(result.Users)
		if err := printUserSummaries(cmd, summaries); err != nil {
			return err
		}
		if result.BigList && result.NextMaxID != "" {
			fmt.Printf("Next cursor: %s\n", result.NextMaxID)
		}
		return nil
	}
}

func newRelationshipsFollowCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "follow",
		Short: "Follow a user",
		RunE:  makeRunRelationshipsFollow(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to follow")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunRelationshipsFollow(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("follow user %s", userID),
				map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Post(ctx, "/api/v1/friendships/create/"+userID+"/", nil)
		if err != nil {
			return fmt.Errorf("following user %s: %w", userID, err)
		}

		var result friendshipActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding follow response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Followed user %s\n", userID)
		return nil
	}
}

func newRelationshipsUnfollowCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfollow",
		Short: "Unfollow a user",
		RunE:  makeRunRelationshipsUnfollow(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to unfollow")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunRelationshipsUnfollow(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("unfollow user %s", userID),
				map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Post(ctx, "/api/v1/friendships/destroy/"+userID+"/", nil)
		if err != nil {
			return fmt.Errorf("unfollowing user %s: %w", userID, err)
		}

		var result friendshipActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding unfollow response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Unfollowed user %s\n", userID)
		return nil
	}
}

func newRelationshipsRemoveFollowerCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-follower",
		Short: "Remove a follower from your followers list",
		RunE:  makeRunRelationshipsRemoveFollower(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to remove as a follower")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunRelationshipsRemoveFollower(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("remove follower %s", userID),
				map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Post(ctx, "/api/v1/friendships/remove_follower/"+userID+"/", nil)
		if err != nil {
			return fmt.Errorf("removing follower %s: %w", userID, err)
		}

		var result friendshipActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding remove follower response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Removed follower %s\n", userID)
		return nil
	}
}

func newRelationshipsBlockCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block",
		Short: "Block a user",
		RunE:  makeRunRelationshipsBlock(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to block")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunRelationshipsBlock(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("block user %s", userID),
				map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Post(ctx, "/api/v1/friendships/block/"+userID+"/", nil)
		if err != nil {
			return fmt.Errorf("blocking user %s: %w", userID, err)
		}

		var result friendshipActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding block response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Blocked user %s\n", userID)
		return nil
	}
}

func newRelationshipsUnblockCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unblock",
		Short: "Unblock a user",
		RunE:  makeRunRelationshipsUnblock(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to unblock")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunRelationshipsUnblock(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("unblock user %s", userID),
				map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Post(ctx, "/api/v1/friendships/unblock/"+userID+"/", nil)
		if err != nil {
			return fmt.Errorf("unblocking user %s: %w", userID, err)
		}

		var result friendshipActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding unblock response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Unblocked user %s\n", userID)
		return nil
	}
}

func newRelationshipsBlockedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blocked",
		Short: "List blocked users",
		RunE:  makeRunRelationshipsBlocked(factory),
	}
	cmd.Flags().Int("limit", 50, "Maximum number of blocked users to return")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func makeRunRelationshipsBlocked(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("count", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("max_id", cursor)
		}

		resp, err := client.Get(ctx, "/api/v1/users/blocked_list/", params)
		if err != nil {
			return fmt.Errorf("listing blocked users: %w", err)
		}

		var result blockedUsersResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding blocked users response: %w", err)
		}

		summaries := toUserSummaries(result.BlockedList)
		return printUserSummaries(cmd, summaries)
	}
}

func newRelationshipsMuteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mute",
		Short: "Mute posts and/or stories from a user",
		RunE:  makeRunRelationshipsMute(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to mute")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("stories", false, "Mute stories")
	cmd.Flags().Bool("posts", false, "Mute posts")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunRelationshipsMute(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		muteStories, _ := cmd.Flags().GetBool("stories")
		mutePosts, _ := cmd.Flags().GetBool("posts")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("mute user %s (stories=%v posts=%v)", userID, muteStories, mutePosts),
				map[string]any{"user_id": userID, "mute_stories": muteStories, "mute_posts": mutePosts})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("target_posts_author_id", userID)
		if muteStories {
			body.Set("story_mute_state", "unmuted_content_muted")
		}
		if mutePosts {
			body.Set("post_mute_state", "unmuted_content_muted")
		}

		resp, err := client.Post(ctx, "/api/v1/friendships/mute_posts_or_story_from_follow/", body)
		if err != nil {
			return fmt.Errorf("muting user %s: %w", userID, err)
		}

		var result friendshipActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding mute response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Muted user %s\n", userID)
		return nil
	}
}

func newRelationshipsUnmuteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unmute",
		Short: "Unmute posts and/or stories from a user",
		RunE:  makeRunRelationshipsUnmute(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to unmute")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("stories", false, "Unmute stories")
	cmd.Flags().Bool("posts", false, "Unmute posts")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunRelationshipsUnmute(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		unmuteStories, _ := cmd.Flags().GetBool("stories")
		unmutePosts, _ := cmd.Flags().GetBool("posts")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("unmute user %s (stories=%v posts=%v)", userID, unmuteStories, unmutePosts),
				map[string]any{"user_id": userID, "unmute_stories": unmuteStories, "unmute_posts": unmutePosts})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("target_posts_author_id", userID)
		if unmuteStories {
			body.Set("story_mute_state", "unmuted")
		}
		if unmutePosts {
			body.Set("post_mute_state", "unmuted")
		}

		resp, err := client.Post(ctx, "/api/v1/friendships/unmute_posts_or_story_from_follow/", body)
		if err != nil {
			return fmt.Errorf("unmuting user %s: %w", userID, err)
		}

		var result friendshipActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding unmute response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Unmuted user %s\n", userID)
		return nil
	}
}

func newRelationshipsRestrictCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restrict",
		Short: "Restrict a user",
		RunE:  makeRunRelationshipsRestrict(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to restrict")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunRelationshipsRestrict(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("restrict user %s", userID),
				map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("user_ids", userID)

		resp, err := client.Post(ctx, "/api/v1/restrict_action/restrict/", body)
		if err != nil {
			return fmt.Errorf("restricting user %s: %w", userID, err)
		}

		var result friendshipActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding restrict response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Restricted user %s\n", userID)
		return nil
	}
}

func newRelationshipsUnrestrictCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unrestrict",
		Short: "Unrestrict a user",
		RunE:  makeRunRelationshipsUnrestrict(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to unrestrict")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunRelationshipsUnrestrict(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("unrestrict user %s", userID),
				map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("user_ids", userID)

		resp, err := client.Post(ctx, "/api/v1/restrict_action/unrestrict/", body)
		if err != nil {
			return fmt.Errorf("unrestricting user %s: %w", userID, err)
		}

		var result friendshipActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding unrestrict response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Unrestricted user %s\n", userID)
		return nil
	}
}

func newRelationshipsStatusCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show friendship status with a user",
		RunE:  makeRunRelationshipsStatus(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to check relationship status with")
	_ = cmd.MarkFlagRequired("user-id")
	return cmd
}

func makeRunRelationshipsStatus(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Get(ctx, "/api/v1/friendships/show/"+userID+"/", nil)
		if err != nil {
			return fmt.Errorf("getting friendship status with user %s: %w", userID, err)
		}

		var result friendshipActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding friendship status response: %w", err)
		}

		status := RelationshipStatusResult{
			UserID:          userID,
			Following:       result.Friendship.Following,
			FollowedBy:      result.Friendship.FollowedBy,
			Blocking:        result.Friendship.Blocking,
			Muting:          result.Friendship.Muting,
			IsPrivate:       result.Friendship.IsPrivate,
			IncomingRequest: result.Friendship.IncomingRequest,
			OutgoingRequest: result.Friendship.OutgoingRequest,
			IsRestricted:    result.Friendship.IsRestricted,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(status)
		}

		lines := []string{
			fmt.Sprintf("User ID:          %s", status.UserID),
			fmt.Sprintf("Following:        %v", status.Following),
			fmt.Sprintf("Followed by:      %v", status.FollowedBy),
			fmt.Sprintf("Blocking:         %v", status.Blocking),
			fmt.Sprintf("Muting:           %v", status.Muting),
			fmt.Sprintf("Private:          %v", status.IsPrivate),
			fmt.Sprintf("Incoming request: %v", status.IncomingRequest),
			fmt.Sprintf("Outgoing request: %v", status.OutgoingRequest),
			fmt.Sprintf("Restricted:       %v", status.IsRestricted),
		}
		cli.PrintText(lines)
		return nil
	}
}

// toUserSummaries converts a slice of rawUser to a slice of UserSummary.
func toUserSummaries(users []rawUser) []UserSummary {
	summaries := make([]UserSummary, 0, len(users))
	for _, u := range users {
		summaries = append(summaries, UserSummary{
			ID:            u.PK,
			Username:      u.Username,
			FullName:      u.FullName,
			ProfilePicURL: u.ProfilePicURL,
			IsPrivate:     u.IsPrivate,
			IsVerified:    u.IsVerified,
		})
	}
	return summaries
}
