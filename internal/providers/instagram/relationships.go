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
		Short:   "View followers, following, and blocks",
		Aliases: []string{"rel", "friendship"},
	}
	cmd.AddCommand(newRelationshipsFollowersCmd(factory))
	cmd.AddCommand(newRelationshipsFollowingCmd(factory))
	cmd.AddCommand(newRelationshipsBlockedCmd(factory))
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
			userID = client.SelfUserID()
		}

		params := url.Values{}
		params.Set("count", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("max_id", cursor)
		}
		if query != "" {
			params.Set("query", query)
		}

		resp, err := client.MobileGet(ctx, "/api/v1/friendships/"+url.PathEscape(userID)+"/followers/", params)
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
			userID = client.SelfUserID()
		}

		params := url.Values{}
		params.Set("count", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("max_id", cursor)
		}
		if query != "" {
			params.Set("query", query)
		}

		resp, err := client.MobileGet(ctx, "/api/v1/friendships/"+url.PathEscape(userID)+"/following/", params)
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

		resp, err := client.MobileGet(ctx, "/api/v1/users/blocked_list/", params)
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

		resp, err := client.MobileGet(ctx, "/api/v1/friendships/show/"+url.PathEscape(userID)+"/", nil)
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
