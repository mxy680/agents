package instagram

import (
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// closeFriendsListResponse is the response for GET /api/v1/friendships/besties/.
type closeFriendsListResponse struct {
	Users      []rawUser `json:"users"`
	NextMaxID  string    `json:"next_max_id"`
	BigList    bool      `json:"big_list"`
	Status     string    `json:"status"`
}

// setBestiesResponse is the response for POST /api/v1/friendships/set_besties/.
type setBestiesResponse struct {
	Status string `json:"status"`
}

// newCloseFriendsCmd builds the `closefriends` subcommand group.
func newCloseFriendsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "closefriends",
		Short:   "Manage your close friends (besties) list",
		Aliases: []string{"cf", "besties"},
	}
	cmd.AddCommand(newCloseFriendsListCmd(factory))
	cmd.AddCommand(newCloseFriendsAddCmd(factory))
	cmd.AddCommand(newCloseFriendsRemoveCmd(factory))
	return cmd
}

func newCloseFriendsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your close friends",
		RunE:  makeRunCloseFriendsList(factory),
	}
	return cmd
}

func makeRunCloseFriendsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Get(ctx, "/api/v1/friendships/besties/", nil)
		if err != nil {
			return fmt.Errorf("listing close friends: %w", err)
		}

		var result closeFriendsListResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding close friends response: %w", err)
		}

		summaries := toUserSummaries(result.Users)
		return printUserSummaries(cmd, summaries)
	}
}

func newCloseFriendsAddCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a user to your close friends",
		RunE:  makeRunCloseFriendsAdd(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to add")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunCloseFriendsAdd(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("add user %s to close friends", userID),
				map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("add", userID)

		resp, err := client.Post(ctx, "/api/v1/friendships/set_besties/", body)
		if err != nil {
			return fmt.Errorf("adding user %s to close friends: %w", userID, err)
		}

		var result setBestiesResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding set besties response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Added user %s to close friends\n", userID)
		return nil
	}
}

func newCloseFriendsRemoveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a user from your close friends",
		RunE:  makeRunCloseFriendsRemove(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to remove")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunCloseFriendsRemove(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("remove user %s from close friends", userID),
				map[string]string{"user_id": userID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		body.Set("remove", userID)

		resp, err := client.Post(ctx, "/api/v1/friendships/set_besties/", body)
		if err != nil {
			return fmt.Errorf("removing user %s from close friends: %w", userID, err)
		}

		var result setBestiesResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding set besties response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Removed user %s from close friends\n", userID)
		return nil
	}
}
