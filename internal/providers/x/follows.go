package x

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// GraphQL query hashes for follower/following operations.
const (
	hashFollowers         = "gC_lyAxZOptAMLCJX5UhWw"
	hashFollowing         = "2vUj-_Ek-UmBVDNtd8OnQA"
	hashVerifiedFollowers = "VmIlPJNEDVQ29HfzIhV4mw"
	hashFollowersYouKnow  = "f2tbuGNjfOE8mNUO5itMew"
)

// newFollowsCmd builds the "follows" subcommand group.
func newFollowsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "follows",
		Short:   "Manage follows and followers",
		Aliases: []string{"follow"},
	}
	cmd.AddCommand(newFollowsFollowersCmd(factory))
	cmd.AddCommand(newFollowsFollowingCmd(factory))
	cmd.AddCommand(newFollowsVerifiedFollowersCmd(factory))
	cmd.AddCommand(newFollowsFollowersYouKnowCmd(factory))
	cmd.AddCommand(newFollowsFollowCmd(factory))
	cmd.AddCommand(newFollowsUnfollowCmd(factory))
	return cmd
}

// newFollowsFollowersCmd builds the "follows followers" command.
func newFollowsFollowersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "followers",
		Short: "Get followers of a user",
		RunE:  makeRunFollowsFollowers(factory),
	}
	cmd.Flags().String("user-id", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newFollowsFollowingCmd builds the "follows following" command.
func newFollowsFollowingCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "following",
		Short: "Get users that a user is following",
		RunE:  makeRunFollowsFollowing(factory),
	}
	cmd.Flags().String("user-id", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newFollowsVerifiedFollowersCmd builds the "follows verified-followers" command.
func newFollowsVerifiedFollowersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verified-followers",
		Short: "Get blue-verified followers of a user",
		RunE:  makeRunFollowsVerifiedFollowers(factory),
	}
	cmd.Flags().String("user-id", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newFollowsFollowersYouKnowCmd builds the "follows followers-you-know" command.
func newFollowsFollowersYouKnowCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "followers-you-know",
		Short: "Get followers of a user that you also follow",
		RunE:  makeRunFollowsFollowersYouKnow(factory),
	}
	cmd.Flags().String("user-id", "", "User ID (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newFollowsFollowCmd builds the "follows follow" command.
func newFollowsFollowCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "follow",
		Short: "Follow a user",
		RunE:  makeRunFollowsFollow(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to follow (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be sent without following")
	return cmd
}

// newFollowsUnfollowCmd builds the "follows unfollow" command.
func newFollowsUnfollowCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfollow",
		Short: "Unfollow a user",
		RunE:  makeRunFollowsUnfollow(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to unfollow (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be sent without unfollowing")
	return cmd
}

// --- RunE implementations ---

func makeRunFollowsFollowers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		return runFollowsUserTimeline(cmd, factory, hashFollowers, "Followers")
	}
}

func makeRunFollowsFollowing(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		return runFollowsUserTimeline(cmd, factory, hashFollowing, "Following")
	}
}

func makeRunFollowsVerifiedFollowers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		return runFollowsUserTimeline(cmd, factory, hashVerifiedFollowers, "BlueVerifiedFollowers")
	}
}

func makeRunFollowsFollowersYouKnow(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		return runFollowsUserTimeline(cmd, factory, hashFollowersYouKnow, "FollowersYouKnow")
	}
}

// runFollowsUserTimeline is the shared implementation for follower/following list commands.
func runFollowsUserTimeline(cmd *cobra.Command, factory ClientFactory, queryHash, operationName string) error {
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

	data, err := client.GraphQL(ctx, queryHash, operationName, vars, DefaultFeatures)
	if err != nil {
		return fmt.Errorf("fetching %s for user %s: %w", operationName, userID, err)
	}

	return printUserListResult(cmd, data)
}

func makeRunFollowsFollow(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("follow user %s", userID), map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("user_id", userID)

		resp, err := client.Post(ctx, "/i/api/1.1/friendships/create.json", body)
		if err != nil {
			return fmt.Errorf("following user %s: %w", userID, err)
		}

		var result json.RawMessage
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("following user %s: %w", userID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "followed", "user_id": userID})
		}
		fmt.Printf("Now following user: %s\n", userID)
		return nil
	}
}

func makeRunFollowsUnfollow(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("unfollow user %s", userID), map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("user_id", userID)

		resp, err := client.Post(ctx, "/i/api/1.1/friendships/destroy.json", body)
		if err != nil {
			return fmt.Errorf("unfollowing user %s: %w", userID, err)
		}

		var result json.RawMessage
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("unfollowing user %s: %w", userID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "unfollowed", "user_id": userID})
		}
		fmt.Printf("Unfollowed user: %s\n", userID)
		return nil
	}
}
