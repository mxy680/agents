package instagram

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// feedUserResponse is the response envelope for GET /api/v1/feed/user/{user_id}/.
type feedUserResponse struct {
	Items         []rawMediaItem `json:"items"`
	NextMaxID     string         `json:"next_max_id"`
	MoreAvailable bool           `json:"more_available"`
	Status        string         `json:"status"`
}

// mediaInfoResponse is the response envelope for GET /api/v1/media/{id}/info/.
type mediaInfoResponse struct {
	Items  []rawMediaItem `json:"items"`
	Status string         `json:"status"`
}

// rawMediaItem is the raw media object from the Instagram API.
type rawMediaItem struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	MediaType int    `json:"media_type"`
	Caption   struct {
		Text string `json:"text"`
	} `json:"caption"`
	TakenAt       int64 `json:"taken_at"`
	LikeCount     int64 `json:"like_count"`
	CommentCount  int64 `json:"comment_count"`
	ImageVersions struct {
		Candidates []struct {
			URL string `json:"url"`
		} `json:"candidates"`
	} `json:"image_versions2"`
	// For reels/clips
	PlayCount int64 `json:"play_count"`
}

// toMediaSummary converts a rawMediaItem to MediaSummary.
func toMediaSummary(item rawMediaItem) MediaSummary {
	thumbnailURL := ""
	if len(item.ImageVersions.Candidates) > 0 {
		thumbnailURL = item.ImageVersions.Candidates[0].URL
	}
	return MediaSummary{
		ID:           item.ID,
		Shortcode:    item.Code,
		MediaType:    item.MediaType,
		Caption:      item.Caption.Text,
		Timestamp:    item.TakenAt,
		LikeCount:    item.LikeCount,
		CommentCount: item.CommentCount,
		ThumbnailURL: thumbnailURL,
	}
}

// mediaDeleteResponse is the response for POST /api/v1/media/{id}/delete/.
type mediaDeleteResponse struct {
	DidDelete bool   `json:"did_delete"`
	Status    string `json:"status"`
}

// mediaActionResponse is a generic response for archive/save/unsave operations.
type mediaActionResponse struct {
	Status string `json:"status"`
}

// mediaLikersResponse is the response for GET /api/v1/media/{id}/likers/.
type mediaLikersResponse struct {
	Users  []rawUser `json:"users"`
	Status string    `json:"status"`
}

// rawUser is a minimal user representation from liker/follower lists.
type rawUser struct {
	PK            string `json:"pk"`
	Username      string `json:"username"`
	FullName      string `json:"full_name"`
	ProfilePicURL string `json:"profile_pic_url"`
	IsPrivate     bool   `json:"is_private"`
	IsVerified    bool   `json:"is_verified"`
}

// newMediaCmd builds the `media` subcommand group.
func newMediaCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "media",
		Short:   "View and manage posts/media",
		Aliases: []string{"post", "posts"},
	}
	cmd.AddCommand(newMediaListCmd(factory))
	cmd.AddCommand(newMediaGetCmd(factory))
	cmd.AddCommand(newMediaDeleteCmd(factory))
	cmd.AddCommand(newMediaArchiveCmd(factory))
	cmd.AddCommand(newMediaUnarchiveCmd(factory))
	cmd.AddCommand(newMediaLikersCmd(factory))
	cmd.AddCommand(newMediaSaveCmd(factory))
	cmd.AddCommand(newMediaUnsaveCmd(factory))
	return cmd
}

func newMediaListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List user's posts",
		Long:  "List media posts for a user. Defaults to the authenticated user.",
		RunE:  makeRunMediaList(factory),
	}
	cmd.Flags().String("user-id", "", "User ID to list media for (defaults to own user)")
	cmd.Flags().Int("limit", 20, "Maximum number of posts to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (next_max_id from previous response)")
	return cmd
}

func makeRunMediaList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

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

		resp, err := client.MobileGet(ctx, "/api/v1/feed/user/"+url.PathEscape(userID)+"/", params)
		if err != nil {
			return fmt.Errorf("listing media for user %s: %w", userID, err)
		}

		var feed feedUserResponse
		if err := client.DecodeJSON(resp, &feed); err != nil {
			return fmt.Errorf("decoding media list: %w", err)
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

func newMediaGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a single post/media item",
		Long:  "Retrieve full details for a media post by ID.",
		RunE:  makeRunMediaGet(factory),
	}
	cmd.Flags().String("media-id", "", "Media/post ID")
	_ = cmd.MarkFlagRequired("media-id")
	return cmd
}

func makeRunMediaGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobileGet(ctx, "/api/v1/media/"+url.PathEscape(mediaID)+"/info/", nil)
		if err != nil {
			return fmt.Errorf("getting media %s: %w", mediaID, err)
		}

		var info mediaInfoResponse
		if err := client.DecodeJSON(resp, &info); err != nil {
			return fmt.Errorf("decoding media info: %w", err)
		}

		if len(info.Items) == 0 {
			return fmt.Errorf("media %s not found", mediaID)
		}

		summary := toMediaSummary(info.Items[0])
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summary)
		}

		lines := []string{
			fmt.Sprintf("ID:        %s", summary.ID),
			fmt.Sprintf("Shortcode: %s", summary.Shortcode),
			fmt.Sprintf("Type:      %d", summary.MediaType),
			fmt.Sprintf("Caption:   %s", truncate(summary.Caption, 80)),
			fmt.Sprintf("Date:      %s", formatTimestamp(summary.Timestamp)),
			fmt.Sprintf("Likes:     %s", formatCount(summary.LikeCount)),
			fmt.Sprintf("Comments:  %s", formatCount(summary.CommentCount)),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newMediaDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a post",
		Long:  "Permanently delete one of your own posts.",
		RunE:  makeRunMediaDelete(factory),
	}
	cmd.Flags().String("media-id", "", "Media/post ID to delete")
	_ = cmd.MarkFlagRequired("media-id")
	cmd.Flags().Bool("confirm", false, "Confirm deletion (required)")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunMediaDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("delete media %s", mediaID), map[string]string{"media_id": mediaID})
		}
		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/media/"+url.PathEscape(mediaID)+"/delete/", nil)
		if err != nil {
			return fmt.Errorf("deleting media %s: %w", mediaID, err)
		}

		var result mediaDeleteResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding delete response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Deleted media %s\n", mediaID)
		return nil
	}
}

func newMediaArchiveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "archive",
		Short: "Archive a post (make only visible to you)",
		RunE:  makeRunMediaArchive(factory),
	}
	cmd.Flags().String("media-id", "", "Media/post ID to archive")
	_ = cmd.MarkFlagRequired("media-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunMediaArchive(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("archive media %s", mediaID), map[string]string{"media_id": mediaID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/media/"+url.PathEscape(mediaID)+"/only_me/", nil)
		if err != nil {
			return fmt.Errorf("archiving media %s: %w", mediaID, err)
		}

		var result mediaActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding archive response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Archived media %s\n", mediaID)
		return nil
	}
}

func newMediaUnarchiveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unarchive",
		Short: "Unarchive a post (make visible to everyone)",
		RunE:  makeRunMediaUnarchive(factory),
	}
	cmd.Flags().String("media-id", "", "Media/post ID to unarchive")
	_ = cmd.MarkFlagRequired("media-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunMediaUnarchive(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("unarchive media %s", mediaID), map[string]string{"media_id": mediaID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/media/"+url.PathEscape(mediaID)+"/undo_only_me/", nil)
		if err != nil {
			return fmt.Errorf("unarchiving media %s: %w", mediaID, err)
		}

		var result mediaActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding unarchive response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Unarchived media %s\n", mediaID)
		return nil
	}
}

func newMediaLikersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "likers",
		Short: "List users who liked a post",
		RunE:  makeRunMediaLikers(factory),
	}
	cmd.Flags().String("media-id", "", "Media/post ID")
	_ = cmd.MarkFlagRequired("media-id")
	cmd.Flags().Int("limit", 50, "Maximum number of likers to return")
	return cmd
}

func makeRunMediaLikers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")

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

		limit, _ := cmd.Flags().GetInt("limit")
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

func newMediaSaveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save/bookmark a post",
		RunE:  makeRunMediaSave(factory),
	}
	cmd.Flags().String("media-id", "", "Media/post ID to save")
	_ = cmd.MarkFlagRequired("media-id")
	cmd.Flags().String("collection-id", "", "Collection ID to save into (optional)")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunMediaSave(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")
		collectionID, _ := cmd.Flags().GetString("collection-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("save media %s", mediaID), map[string]string{"media_id": mediaID, "collection_id": collectionID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := url.Values{}
		if collectionID != "" {
			body.Set("collection_ids", collectionID)
		}

		resp, err := client.MobilePost(ctx, "/api/v1/media/"+url.PathEscape(mediaID)+"/save/", body)
		if err != nil {
			return fmt.Errorf("saving media %s: %w", mediaID, err)
		}

		var result mediaActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding save response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Saved media %s\n", mediaID)
		return nil
	}
}

func newMediaUnsaveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsave",
		Short: "Remove a post from saved/bookmarks",
		RunE:  makeRunMediaUnsave(factory),
	}
	cmd.Flags().String("media-id", "", "Media/post ID to unsave")
	_ = cmd.MarkFlagRequired("media-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunMediaUnsave(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		mediaID, _ := cmd.Flags().GetString("media-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("unsave media %s", mediaID), map[string]string{"media_id": mediaID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.MobilePost(ctx, "/api/v1/media/"+url.PathEscape(mediaID)+"/unsave/", nil)
		if err != nil {
			return fmt.Errorf("unsaving media %s: %w", mediaID, err)
		}

		var result mediaActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding unsave response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Unsaved media %s\n", mediaID)
		return nil
	}
}
