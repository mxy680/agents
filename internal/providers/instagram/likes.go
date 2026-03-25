package instagram

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

// likedFeedResponse is the response for GET /api/v1/feed/liked/.
type likedFeedResponse struct {
	Items         []rawMediaItem `json:"items"`
	NextMaxID     string         `json:"next_max_id"`
	MoreAvailable bool           `json:"more_available"`
	Status        string         `json:"status"`
}

// newLikesCmd builds the `likes` subcommand group.
func newLikesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "likes",
		Short:   "View post likes",
		Aliases: []string{"like"},
	}
	cmd.AddCommand(newLikesListCmd(factory))
	cmd.AddCommand(newLikesLikedCmd(factory))
	return cmd
}

func newLikesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users who liked a post",
		RunE:  makeRunLikesList(factory),
	}
	cmd.Flags().String("media-id", "", "Media/post ID")
	_ = cmd.MarkFlagRequired("media-id")
	cmd.Flags().Int("limit", 50, "Maximum number of likers to return")
	return cmd
}

func makeRunLikesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobileGet(ctx, "/api/v1/media/"+url.PathEscape(mediaID)+"/likers/", nil)
		if err != nil {
			return fmt.Errorf("getting likers for media %s: %w", mediaID, err)
		}

		var result mediaLikersResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding likers response: %w", err)
		}

		users := result.Users
		if len(users) > limit {
			users = users[:limit]
		}

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
		return printUserSummaries(cmd, summaries)
	}
}

func newLikesLikedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "liked",
		Short: "List posts the authenticated user has liked",
		RunE:  makeRunLikesLiked(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of posts to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (next_max_id from previous response)")
	return cmd
}

func makeRunLikesLiked(factory ClientFactory) func(*cobra.Command, []string) error {
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

		resp, err := client.MobileGet(ctx, "/api/v1/feed/liked/", params)
		if err != nil {
			return fmt.Errorf("getting liked posts: %w", err)
		}

		var feed likedFeedResponse
		if err := client.DecodeJSON(resp, &feed); err != nil {
			return fmt.Errorf("decoding liked feed: %w", err)
		}

		summaries := make([]MediaSummary, 0, len(feed.Items))
		for _, item := range feed.Items {
			summaries = append(summaries, toMediaSummary(item))
		}

		if err := printMediaSummaries(cmd, summaries); err != nil {
			return err
		}
		if feed.MoreAvailable && feed.NextMaxID != "" {
			fmt.Printf("Next cursor: %s\n", feed.NextMaxID)
		}
		return nil
	}
}
