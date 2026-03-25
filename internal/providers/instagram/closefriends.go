package instagram

import (
	"fmt"

	"github.com/spf13/cobra"
)

// closeFriendsListResponse is the response for GET /api/v1/friendships/besties/.
type closeFriendsListResponse struct {
	Users      []rawUser `json:"users"`
	NextMaxID  string    `json:"next_max_id"`
	BigList    bool      `json:"big_list"`
	Status     string    `json:"status"`
}

// newCloseFriendsCmd builds the `closefriends` subcommand group.
func newCloseFriendsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "closefriends",
		Short:   "View your close friends (besties) list",
		Aliases: []string{"cf", "besties"},
	}
	cmd.AddCommand(newCloseFriendsListCmd(factory))
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

		resp, err := client.MobileGet(ctx, "/api/v1/friendships/besties/", nil)
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


