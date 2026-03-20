package linkedin

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerFeedUpdatesResponse is the response envelope for profile post lists.
type voyagerFeedUpdatesResponse struct {
	Elements []voyagerFeedElement `json:"elements"`
	Paging   voyagerPaging        `json:"paging"`
}

// voyagerFeedElement represents a single post element in feed responses.
type voyagerFeedElement struct {
	UpdateMetadata *struct {
		URN string `json:"urn"`
	} `json:"updateMetadata"`
	Actor *struct {
		Name struct {
			Text string `json:"text"`
		} `json:"name"`
		URN string `json:"urn"`
	} `json:"actor"`
	Commentary *struct {
		Text struct {
			Text string `json:"text"`
		} `json:"text"`
	} `json:"commentary"`
	SocialDetail *struct {
		TotalSocialActivityCounts struct {
			NumLikes    int `json:"numLikes"`
			NumComments int `json:"numComments"`
			NumShares   int `json:"numShares"`
		} `json:"totalSocialActivityCounts"`
	} `json:"socialDetail"`
	CreatedAt int64 `json:"createdAt"`
}

// voyagerSinglePostResponse is the response envelope for GET /voyager/api/feed/updates/{urn}.
type voyagerSinglePostResponse struct {
	URN   string `json:"entityUrn"`
	Actor *struct {
		Name struct {
			Text string `json:"text"`
		} `json:"name"`
		URN string `json:"urn"`
	} `json:"actor"`
	Commentary *struct {
		Text struct {
			Text string `json:"text"`
		} `json:"text"`
	} `json:"commentary"`
	SocialDetail *struct {
		TotalSocialActivityCounts struct {
			NumLikes    int `json:"numLikes"`
			NumComments int `json:"numComments"`
			NumShares   int `json:"numShares"`
		} `json:"totalSocialActivityCounts"`
	} `json:"socialDetail"`
	CreatedAt int64 `json:"createdAt"`
}

// voyagerReactionsResponse is the response envelope for listing reactions on a post.
type voyagerReactionsResponse struct {
	Elements []struct {
		ReactionType string `json:"reactionType"`
		Actor        *struct {
			Name struct {
				Text string `json:"text"`
			} `json:"name"`
			URN string `json:"urn"`
		} `json:"actor"`
	} `json:"elements"`
	Paging voyagerPaging `json:"paging"`
}

// ReactionSummary is a condensed reaction representation.
type ReactionSummary struct {
	ReactionType string `json:"reaction_type"`
	ActorURN     string `json:"actor_urn"`
	ActorName    string `json:"actor_name,omitempty"`
}

// newPostsCmd builds the "posts" subcommand group.
func newPostsCmd(factory ClientFactory) *cobra.Command {
	postsCmd := &cobra.Command{
		Use:     "posts",
		Short:   "Manage LinkedIn posts",
		Aliases: []string{"post"},
	}
	postsCmd.AddCommand(newPostsListCmd(factory))
	postsCmd.AddCommand(newPostsGetCmd(factory))
	postsCmd.AddCommand(newPostsCreateCmd(factory))
	postsCmd.AddCommand(newPostsDeleteCmd(factory))
	postsCmd.AddCommand(newPostsReactionsCmd(factory))
	postsCmd.AddCommand(newPostsReactCmd(factory))
	return postsCmd
}

// newPostsListCmd builds the "posts list" command.
func newPostsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your LinkedIn posts",
		Long:  "List posts for the authenticated user or a specified user.",
		RunE:  makeRunPostsList(factory),
	}
	cmd.Flags().String("username", "", "LinkedIn public profile ID to list posts for (defaults to self)")
	cmd.Flags().Int("limit", 10, "Maximum number of posts to return")
	cmd.Flags().String("cursor", "0", "Pagination start offset")
	return cmd
}

// newPostsGetCmd builds the "posts get" command.
func newPostsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a single LinkedIn post",
		Long:  "Retrieve a single post by its activity URN.",
		RunE:  makeRunPostsGet(factory),
	}
	cmd.Flags().String("post-urn", "", "Activity URN of the post (e.g. urn:li:activity:1234)")
	_ = cmd.MarkFlagRequired("post-urn")
	return cmd
}

// newPostsCreateCmd builds the "posts create" command.
func newPostsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new LinkedIn post",
		Long:  "Share a text post on LinkedIn.",
		RunE:  makeRunPostsCreate(factory),
	}
	cmd.Flags().String("text", "", "Post text content")
	_ = cmd.MarkFlagRequired("text")
	cmd.Flags().String("visibility", "public", "Post visibility: public or connections")
	cmd.Flags().Bool("dry-run", false, "Print what would be sent without creating the post")
	return cmd
}

// newPostsDeleteCmd builds the "posts delete" command.
func newPostsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a LinkedIn post",
		Long:  "Delete a post by its activity URN.",
		RunE:  makeRunPostsDelete(factory),
	}
	cmd.Flags().String("post-urn", "", "Activity URN of the post (e.g. urn:li:activity:1234)")
	_ = cmd.MarkFlagRequired("post-urn")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Print what would be deleted without deleting")
	return cmd
}

// newPostsReactionsCmd builds the "posts reactions" command.
func newPostsReactionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reactions",
		Short: "List reactions on a LinkedIn post",
		Long:  "List all reactions on a post by its activity URN.",
		RunE:  makeRunPostsReactions(factory),
	}
	cmd.Flags().String("post-urn", "", "Activity URN of the post (e.g. urn:li:activity:1234)")
	_ = cmd.MarkFlagRequired("post-urn")
	cmd.Flags().Int("limit", 10, "Maximum number of reactions to return")
	return cmd
}

// newPostsReactCmd builds the "posts react" command.
func newPostsReactCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "react",
		Short: "React to a LinkedIn post",
		Long:  "Add a reaction to a post by its activity URN.",
		RunE:  makeRunPostsReact(factory),
	}
	cmd.Flags().String("post-urn", "", "Activity URN of the post (e.g. urn:li:activity:1234)")
	_ = cmd.MarkFlagRequired("post-urn")
	cmd.Flags().String("type", "LIKE", "Reaction type: LIKE, CELEBRATE, SUPPORT, LOVE, INSIGHTFUL, FUNNY")
	_ = cmd.MarkFlagRequired("type")
	cmd.Flags().Bool("dry-run", false, "Print what would be reacted without sending")
	return cmd
}

func makeRunPostsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		username, _ := cmd.Flags().GetString("username")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		// Resolve the profile ID to a fsd_profile URN.
		// The posts endpoint requires fsd_profile, not fs_miniProfile.
		profileURN := ""
		if username != "" {
			// Use the provided username directly as the fsd_profile ID.
			profileURN = "urn:li:fsd_profile:" + username
		} else {
			// Fetch the current user's miniProfile URN from /me, then
			// convert to fsd_profile by extracting the profile ID.
			meResp, err := client.Get(ctx, "/voyager/api/me", nil)
			if err != nil {
				return fmt.Errorf("getting current user: %w", err)
			}
			var normalized NormalizedResponse
			if err := client.DecodeJSON(meResp, &normalized); err != nil {
				return fmt.Errorf("decoding current user: %w", err)
			}
			rawMP := FindIncluded(normalized.Included, "MiniProfile")
			if rawMP != nil {
				var mp miniProfileEntity
				if jsonErr := json.Unmarshal(rawMP, &mp); jsonErr == nil && mp.EntityURN != "" {
					// Extract profile ID from urn:li:fs_miniProfile:<id>
					parts := strings.Split(mp.EntityURN, ":")
					if len(parts) > 0 {
						profileID := parts[len(parts)-1]
						profileURN = "urn:li:fsd_profile:" + profileID
					}
				}
			}
		}

		params := url.Values{}
		params.Set("q", "memberShareFeed")
		params.Set("moduleKey", "member-shares:phone")
		params.Set("count", fmt.Sprintf("%d", limit))
		params.Set("start", cursor)
		if profileURN != "" {
			params.Set("profileUrn", profileURN)
		}

		resp, err := client.Get(ctx, "/voyager/api/identity/profileUpdatesV2", params)
		if err != nil {
			return fmt.Errorf("listing posts: %w", err)
		}

		var raw voyagerFeedUpdatesResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding posts: %w", err)
		}

		posts := make([]PostSummary, 0, len(raw.Elements))
		for _, el := range raw.Elements {
			posts = append(posts, feedElementToPostSummary(el))
		}
		return printPostSummaries(cmd, posts)
	}
}

func makeRunPostsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		postURN, _ := cmd.Flags().GetString("post-urn")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/feed/updates/" + url.PathEscape(postURN)
		resp, err := client.Get(ctx, path, nil)
		if err != nil {
			return fmt.Errorf("getting post %s: %w", postURN, err)
		}

		var raw voyagerSinglePostResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding post: %w", err)
		}

		post := singlePostResponseToSummary(postURN, raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(post)
		}

		lines := []string{
			fmt.Sprintf("URN:      %s", post.URN),
			fmt.Sprintf("Author:   %s (%s)", post.AuthorName, post.AuthorURN),
			fmt.Sprintf("Date:     %s", formatTimestamp(post.Timestamp)),
			fmt.Sprintf("Likes:    %s", formatCount(post.LikeCount)),
			fmt.Sprintf("Comments: %s", formatCount(post.CommentCount)),
			fmt.Sprintf("Shares:   %s", formatCount(post.ShareCount)),
		}
		if post.Text != "" {
			lines = append(lines, fmt.Sprintf("Text:     %s", truncate(post.Text, 300)))
		}
		cli.PrintText(lines)
		return nil
	}
}

func makeRunPostsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		text, _ := cmd.Flags().GetString("text")
		visibility, _ := cmd.Flags().GetString("visibility")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		visibilityCode := "PUBLIC"
		if visibility == "connections" {
			visibilityCode = "CONNECTIONS"
		}

		body := map[string]any{
			"commentary": text,
			"visibility": visibilityCode,
			"distribution": map[string]any{
				"feedDistribution": "MAIN_FEED",
			},
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("create post with text %q (visibility=%s)", truncate(text, 60), visibilityCode), body)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.PostJSON(ctx, "/voyager/api/contentCreation/normalizedContent", body)
		if err != nil {
			return fmt.Errorf("creating post: %w", err)
		}

		var result struct {
			Value struct {
				URN string `json:"entityUrn"`
			} `json:"value"`
		}
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding create response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"urn": result.Value.URN})
		}
		fmt.Printf("Post created: %s\n", result.Value.URN)
		return nil
	}
}

func makeRunPostsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		postURN, _ := cmd.Flags().GetString("post-urn")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("delete post %s", postURN), map[string]string{"post_urn": postURN})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/feed/updates/" + url.PathEscape(postURN)
		resp, err := client.Delete(ctx, path)
		if err != nil {
			return fmt.Errorf("deleting post %s: %w", postURN, err)
		}
		resp.Body.Close()

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "post_urn": postURN})
		}
		fmt.Printf("Post deleted: %s\n", postURN)
		return nil
	}
}

func makeRunPostsReactions(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		postURN, _ := cmd.Flags().GetString("post-urn")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("count", fmt.Sprintf("%d", limit))
		params.Set("start", "0")

		path := "/voyager/api/socialActions/" + url.PathEscape(postURN) + "/reactions"
		resp, err := client.Get(ctx, path, params)
		if err != nil {
			return fmt.Errorf("listing reactions for %s: %w", postURN, err)
		}

		var raw voyagerReactionsResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding reactions: %w", err)
		}

		reactions := make([]ReactionSummary, 0, len(raw.Elements))
		for _, el := range raw.Elements {
			r := ReactionSummary{ReactionType: el.ReactionType}
			if el.Actor != nil {
				r.ActorURN = el.Actor.URN
				r.ActorName = el.Actor.Name.Text
			}
			reactions = append(reactions, r)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(reactions)
		}
		if len(reactions) == 0 {
			fmt.Println("No reactions found.")
			return nil
		}
		lines := make([]string, 0, len(reactions)+1)
		lines = append(lines, fmt.Sprintf("%-15s  %-40s  %-30s", "TYPE", "ACTOR URN", "ACTOR NAME"))
		for _, r := range reactions {
			lines = append(lines, fmt.Sprintf("%-15s  %-40s  %-30s",
				r.ReactionType,
				truncate(r.ActorURN, 40),
				truncate(r.ActorName, 30),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

func makeRunPostsReact(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		postURN, _ := cmd.Flags().GetString("post-urn")
		reactionType, _ := cmd.Flags().GetString("type")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		body := map[string]string{"reactionType": reactionType}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("react to post %s with %s", postURN, reactionType), body)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/socialActions/" + url.PathEscape(postURN) + "/reactions"
		resp, err := client.PostJSON(ctx, path, body)
		if err != nil {
			return fmt.Errorf("reacting to post %s: %w", postURN, err)
		}
		resp.Body.Close()

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "reacted", "reaction_type": reactionType, "post_urn": postURN})
		}
		fmt.Printf("Reacted to post %s with %s\n", postURN, reactionType)
		return nil
	}
}

// feedElementToPostSummary converts a voyagerFeedElement to PostSummary.
func feedElementToPostSummary(el voyagerFeedElement) PostSummary {
	post := PostSummary{Timestamp: el.CreatedAt}
	if el.UpdateMetadata != nil {
		post.URN = el.UpdateMetadata.URN
	}
	if el.Actor != nil {
		post.AuthorURN = el.Actor.URN
		post.AuthorName = el.Actor.Name.Text
	}
	if el.Commentary != nil {
		post.Text = el.Commentary.Text.Text
	}
	if el.SocialDetail != nil {
		counts := el.SocialDetail.TotalSocialActivityCounts
		post.LikeCount = counts.NumLikes
		post.CommentCount = counts.NumComments
		post.ShareCount = counts.NumShares
	}
	return post
}

// singlePostResponseToSummary converts a voyagerSinglePostResponse to PostSummary.
func singlePostResponseToSummary(urn string, raw voyagerSinglePostResponse) PostSummary {
	post := PostSummary{URN: urn}
	if raw.URN != "" {
		post.URN = raw.URN
	}
	if raw.Actor != nil {
		post.AuthorURN = raw.Actor.URN
		post.AuthorName = raw.Actor.Name.Text
	}
	if raw.Commentary != nil {
		post.Text = raw.Commentary.Text.Text
	}
	if raw.SocialDetail != nil {
		counts := raw.SocialDetail.TotalSocialActivityCounts
		post.LikeCount = counts.NumLikes
		post.CommentCount = counts.NumComments
		post.ShareCount = counts.NumShares
	}
	post.Timestamp = raw.CreatedAt
	return post
}
