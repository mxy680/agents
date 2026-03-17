package instagram

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// tagInfoResponse is the response for GET /api/v1/tags/{tag_name}/info/.
type tagInfoResponse struct {
	Tag    rawTagDetail `json:"tag"`
	Status string       `json:"status"`
}

// rawTagDetail is the full tag object from the info endpoint.
type rawTagDetail struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	MediaCount     int64  `json:"media_count"`
	FollowingCount int64  `json:"following_count"`
	IsFollowing    bool   `json:"following"`
}

// tagSectionsResponse is the response for GET /api/v1/tags/{tag_name}/sections/.
type tagSectionsResponse struct {
	Sections      []rawTagSection `json:"sections"`
	NextMaxID     string          `json:"next_max_id"`
	MoreAvailable bool            `json:"more_available"`
	Status        string          `json:"status"`
}

// rawTagSection represents a section (row of items) in the tag feed.
type rawTagSection struct {
	FeedType string         `json:"feed_type"`
	LayoutContent struct {
		Medias []struct {
			Media rawMediaItem `json:"media"`
		} `json:"medias"`
	} `json:"layout_content"`
}

// tagActionResponse is a generic response for follow/unfollow.
type tagActionResponse struct {
	Result string `json:"result"`
	Status string `json:"status"`
}

// followingTagsResponse is the response for GET /api/v1/users/self/following_tag_list/.
type followingTagsResponse struct {
	Tags   []rawTag `json:"tags"`
	Status string   `json:"status"`
}

// relatedTagsResponse is the response for GET /api/v1/tags/{tag_name}/related/.
type relatedTagsResponse struct {
	Related []struct {
		Tag    rawTag `json:"tag"`
		Type   string `json:"type"`
	} `json:"related"`
	Status string `json:"status"`
}

// newTagsCmd builds the `tags` subcommand group.
func newTagsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tags",
		Short:   "Browse and follow hashtags",
		Aliases: []string{"tag", "hashtag"},
	}
	cmd.AddCommand(newTagsGetCmd(factory))
	cmd.AddCommand(newTagsFeedCmd(factory))
	cmd.AddCommand(newTagsFollowCmd(factory))
	cmd.AddCommand(newTagsUnfollowCmd(factory))
	cmd.AddCommand(newTagsFollowingCmd(factory))
	cmd.AddCommand(newTagsRelatedCmd(factory))
	return cmd
}

func newTagsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get hashtag info",
		RunE:  makeRunTagsGet(factory),
	}
	cmd.Flags().String("name", "", "Hashtag name (without #)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunTagsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Get(ctx, "/api/v1/tags/"+name+"/info/", nil)
		if err != nil {
			return fmt.Errorf("getting tag info for #%s: %w", name, err)
		}

		var result tagInfoResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding tag info: %w", err)
		}

		tag := TagSummary{
			ID:             result.Tag.ID,
			Name:           result.Tag.Name,
			MediaCount:     result.Tag.MediaCount,
			FollowingCount: result.Tag.FollowingCount,
			IsFollowing:    result.Tag.IsFollowing,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(tag)
		}

		lines := []string{
			fmt.Sprintf("Name:      #%s", tag.Name),
			fmt.Sprintf("ID:        %s", tag.ID),
			fmt.Sprintf("Posts:     %s", formatCount(tag.MediaCount)),
			fmt.Sprintf("Followers: %s", formatCount(tag.FollowingCount)),
			fmt.Sprintf("Following: %v", tag.IsFollowing),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newTagsFeedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feed",
		Short: "Get posts for a hashtag",
		RunE:  makeRunTagsFeed(factory),
	}
	cmd.Flags().String("name", "", "Hashtag name (without #)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().String("tab", "top", "Feed tab: top or recent")
	cmd.Flags().Int("limit", 20, "Maximum number of items")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func makeRunTagsFeed(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		tab, _ := cmd.Flags().GetString("tab")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("tab", tab)
		params.Set("count", strconv.Itoa(limit))
		if cursor != "" {
			params.Set("max_id", cursor)
		}

		resp, err := client.Get(ctx, "/api/v1/tags/"+name+"/sections/", params)
		if err != nil {
			return fmt.Errorf("getting tag feed for #%s: %w", name, err)
		}

		var result tagSectionsResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding tag feed: %w", err)
		}

		var summaries []MediaSummary
		for _, section := range result.Sections {
			for _, m := range section.LayoutContent.Medias {
				summaries = append(summaries, toMediaSummary(m.Media))
			}
		}

		if err := printMediaSummaries(cmd, summaries); err != nil {
			return err
		}
		if result.MoreAvailable && result.NextMaxID != "" {
			fmt.Printf("Next cursor: %s\n", result.NextMaxID)
		}
		return nil
	}
}

func newTagsFollowCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "follow",
		Short: "Follow a hashtag",
		RunE:  makeRunTagsFollow(factory),
	}
	cmd.Flags().String("name", "", "Hashtag name (without #)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunTagsFollow(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("follow tag #%s", name), map[string]string{"tag": name})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Post(ctx, "/api/v1/tags/follow/"+name+"/", nil)
		if err != nil {
			return fmt.Errorf("following tag #%s: %w", name, err)
		}

		var result tagActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding tag follow response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Followed tag #%s\n", name)
		return nil
	}
}

func newTagsUnfollowCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfollow",
		Short: "Unfollow a hashtag",
		RunE:  makeRunTagsUnfollow(factory),
	}
	cmd.Flags().String("name", "", "Hashtag name (without #)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunTagsUnfollow(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("unfollow tag #%s", name), map[string]string{"tag": name})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Post(ctx, "/api/v1/tags/unfollow/"+name+"/", nil)
		if err != nil {
			return fmt.Errorf("unfollowing tag #%s: %w", name, err)
		}

		var result tagActionResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding tag unfollow response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Unfollowed tag #%s\n", name)
		return nil
	}
}

func newTagsFollowingCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "following",
		Short: "List hashtags you follow",
		RunE:  makeRunTagsFollowing(factory),
	}
	return cmd
}

func makeRunTagsFollowing(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Get(ctx, "/api/v1/users/self/following_tag_list/", nil)
		if err != nil {
			return fmt.Errorf("listing following tags: %w", err)
		}

		var result followingTagsResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding following tags response: %w", err)
		}

		summaries := make([]TagSummary, 0, len(result.Tags))
		for _, t := range result.Tags {
			summaries = append(summaries, TagSummary{
				ID:             t.ID,
				Name:           t.Name,
				MediaCount:     t.MediaCount,
				FollowingCount: t.FollowingCount,
				IsFollowing:    t.IsFollowing,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}

		if len(summaries) == 0 {
			fmt.Println("Not following any tags.")
			return nil
		}

		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-30s  %-12s", "NAME", "POSTS"))
		for _, t := range summaries {
			lines = append(lines, fmt.Sprintf("%-30s  %-12s", truncate(t.Name, 30), formatCount(t.MediaCount)))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newTagsRelatedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "related",
		Short: "Get related hashtags",
		RunE:  makeRunTagsRelated(factory),
	}
	cmd.Flags().String("name", "", "Hashtag name (without #)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunTagsRelated(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Get(ctx, "/api/v1/tags/"+name+"/related/", nil)
		if err != nil {
			return fmt.Errorf("getting related tags for #%s: %w", name, err)
		}

		var result relatedTagsResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding related tags response: %w", err)
		}

		summaries := make([]TagSummary, 0, len(result.Related))
		for _, r := range result.Related {
			summaries = append(summaries, TagSummary{
				ID:             r.Tag.ID,
				Name:           r.Tag.Name,
				MediaCount:     r.Tag.MediaCount,
				FollowingCount: r.Tag.FollowingCount,
				IsFollowing:    r.Tag.IsFollowing,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}

		if len(summaries) == 0 {
			fmt.Println("No related tags found.")
			return nil
		}

		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-30s  %-12s", "NAME", "POSTS"))
		for _, t := range summaries {
			lines = append(lines, fmt.Sprintf("%-30s  %-12s", truncate(t.Name, 30), formatCount(t.MediaCount)))
		}
		cli.PrintText(lines)
		return nil
	}
}
