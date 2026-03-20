package linkedin

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerGroupsResponse is the response envelope for GET /voyager/api/groups/groups.
type voyagerGroupsResponse struct {
	Elements []voyagerGroupElement `json:"elements"`
	Paging   voyagerPaging         `json:"paging"`
}

type voyagerGroupElement struct {
	EntityURN   string `json:"entityUrn"`
	Name        string `json:"name"`
	MemberCount int    `json:"memberCount"`
	Description string `json:"description"`
}

// voyagerGroupMembersResponse is the response for GET /voyager/api/groups/groups/{id}/members.
type voyagerGroupMembersResponse struct {
	Elements []struct {
		EntityURN string `json:"entityUrn"`
		MiniProfile struct {
			FirstName        string `json:"firstName"`
			LastName         string `json:"lastName"`
			PublicIdentifier string `json:"publicIdentifier"`
			Occupation       string `json:"occupation"`
		} `json:"miniProfile"`
	} `json:"elements"`
	Paging voyagerPaging `json:"paging"`
}

// voyagerGroupPostsResponse is the response for GET /voyager/api/groups/groups/{id}/posts.
type voyagerGroupPostsResponse struct {
	Elements []struct {
		UpdateMetadata struct {
			URN string `json:"urn"`
		} `json:"updateMetadata"`
		Actor struct {
			Name struct {
				Text string `json:"text"`
			} `json:"name"`
			URN string `json:"urn"`
		} `json:"actor"`
		Commentary struct {
			Text struct {
				Text string `json:"text"`
			} `json:"text"`
		} `json:"commentary"`
		SocialDetail struct {
			TotalSocialActivityCounts struct {
				NumLikes    int `json:"numLikes"`
				NumComments int `json:"numComments"`
				NumShares   int `json:"numShares"`
			} `json:"totalSocialActivityCounts"`
		} `json:"socialDetail"`
		CreatedAt int64 `json:"createdAt"`
	} `json:"elements"`
	Paging voyagerPaging `json:"paging"`
}

// toGroupSummary maps a voyagerGroupElement to GroupSummary.
func toGroupSummary(el voyagerGroupElement) GroupSummary {
	id := el.EntityURN
	if parts := strings.Split(el.EntityURN, ":"); len(parts) > 0 {
		id = parts[len(parts)-1]
	}
	return GroupSummary{
		ID:          id,
		Name:        el.Name,
		MemberCount: el.MemberCount,
		Description: el.Description,
	}
}

// newGroupsCmd builds the "groups" subcommand group.
func newGroupsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "groups",
		Short:   "Interact with LinkedIn groups",
		Aliases: []string{"group"},
	}
	cmd.AddCommand(newGroupsListCmd(factory))
	cmd.AddCommand(newGroupsGetCmd(factory))
	cmd.AddCommand(newGroupsMembersCmd(factory))
	cmd.AddCommand(newGroupsPostsCmd(factory))
	cmd.AddCommand(newGroupsJoinCmd(factory))
	cmd.AddCommand(newGroupsLeaveCmd(factory))
	return cmd
}

// newGroupsListCmd builds the "groups list" command.
func newGroupsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List groups you are a member of",
		RunE:  makeRunGroupsList(factory),
	}
	cmd.Flags().Int("limit", 10, "Maximum number of groups to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	return cmd
}

// newGroupsGetCmd builds the "groups get" command.
func newGroupsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a group by ID",
		RunE:  makeRunGroupsGet(factory),
	}
	cmd.Flags().String("group-id", "", "Group ID (required)")
	_ = cmd.MarkFlagRequired("group-id")
	return cmd
}

// newGroupsMembersCmd builds the "groups members" command.
func newGroupsMembersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members",
		Short: "List members of a group",
		RunE:  makeRunGroupsMembers(factory),
	}
	cmd.Flags().String("group-id", "", "Group ID (required)")
	cmd.Flags().Int("limit", 10, "Maximum number of members to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	_ = cmd.MarkFlagRequired("group-id")
	return cmd
}

// newGroupsPostsCmd builds the "groups posts" command.
func newGroupsPostsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "posts",
		Short: "List posts in a group",
		RunE:  makeRunGroupsPosts(factory),
	}
	cmd.Flags().String("group-id", "", "Group ID (required)")
	cmd.Flags().Int("limit", 10, "Maximum number of posts to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	_ = cmd.MarkFlagRequired("group-id")
	return cmd
}

// newGroupsJoinCmd builds the "groups join" command.
func newGroupsJoinCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join",
		Short: "Join a group",
		RunE:  makeRunGroupsJoin(factory),
	}
	cmd.Flags().String("group-id", "", "Group ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without joining")
	_ = cmd.MarkFlagRequired("group-id")
	return cmd
}

// newGroupsLeaveCmd builds the "groups leave" command.
func newGroupsLeaveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leave",
		Short: "Leave a group",
		RunE:  makeRunGroupsLeave(factory),
	}
	cmd.Flags().String("group-id", "", "Group ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm leaving (required for destructive action)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without leaving")
	_ = cmd.MarkFlagRequired("group-id")
	return cmd
}

func makeRunGroupsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		start := 0
		if cursor != "" {
			if _, err := fmt.Sscanf(cursor, "%d", &start); err != nil {
				return fmt.Errorf("invalid cursor %q: must be a numeric start offset", cursor)
			}
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{
			"q":     {"memberGroups"},
			"start": {fmt.Sprintf("%d", start)},
			"count": {fmt.Sprintf("%d", limit)},
		}
		resp, err := client.Get(ctx, "/voyager/api/groups/groups", params)
		if err != nil {
			return fmt.Errorf("listing groups: %w", err)
		}

		var raw voyagerGroupsResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding groups: %w", err)
		}

		summaries := make([]GroupSummary, 0, len(raw.Elements))
		for _, el := range raw.Elements {
			summaries = append(summaries, toGroupSummary(el))
		}
		return printGroupSummaries(cmd, summaries)
	}
}

func makeRunGroupsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		groupID, _ := cmd.Flags().GetString("group-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/groups/groups/" + url.PathEscape(groupID)
		resp, err := client.Get(ctx, path, nil)
		if err != nil {
			return fmt.Errorf("getting group %s: %w", groupID, err)
		}

		var raw voyagerGroupElement
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding group: %w", err)
		}

		summary := toGroupSummary(raw)
		return printGroupDetail(cmd, summary)
	}
}

func makeRunGroupsMembers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		groupID, _ := cmd.Flags().GetString("group-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		start := 0
		if cursor != "" {
			if _, err := fmt.Sscanf(cursor, "%d", &start); err != nil {
				return fmt.Errorf("invalid cursor %q: must be a numeric start offset", cursor)
			}
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/groups/groups/" + url.PathEscape(groupID) + "/members"
		params := url.Values{
			"start": {fmt.Sprintf("%d", start)},
			"count": {fmt.Sprintf("%d", limit)},
		}
		resp, err := client.Get(ctx, path, params)
		if err != nil {
			return fmt.Errorf("listing group members: %w", err)
		}

		var raw voyagerGroupMembersResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding group members: %w", err)
		}

		type memberEntry struct {
			URN      string `json:"urn"`
			PublicID string `json:"public_id"`
			Name     string `json:"name"`
			Headline string `json:"headline,omitempty"`
		}
		members := make([]memberEntry, 0, len(raw.Elements))
		for _, el := range raw.Elements {
			members = append(members, memberEntry{
				URN:      el.EntityURN,
				PublicID: el.MiniProfile.PublicIdentifier,
				Name:     el.MiniProfile.FirstName + " " + el.MiniProfile.LastName,
				Headline: el.MiniProfile.Occupation,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(members)
		}
		if len(members) == 0 {
			fmt.Println("No members found.")
			return nil
		}
		lines := make([]string, 0, len(members)+1)
		lines = append(lines, fmt.Sprintf("%-25s  %-40s  %-40s", "PUBLIC ID", "NAME", "HEADLINE"))
		for _, m := range members {
			lines = append(lines, fmt.Sprintf("%-25s  %-40s  %-40s",
				truncate(m.PublicID, 25),
				truncate(m.Name, 40),
				truncate(m.Headline, 40),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

func makeRunGroupsPosts(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		groupID, _ := cmd.Flags().GetString("group-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		start := 0
		if cursor != "" {
			if _, err := fmt.Sscanf(cursor, "%d", &start); err != nil {
				return fmt.Errorf("invalid cursor %q: must be a numeric start offset", cursor)
			}
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/groups/groups/" + url.PathEscape(groupID) + "/posts"
		params := url.Values{
			"start": {fmt.Sprintf("%d", start)},
			"count": {fmt.Sprintf("%d", limit)},
		}
		resp, err := client.Get(ctx, path, params)
		if err != nil {
			return fmt.Errorf("listing group posts: %w", err)
		}

		var raw voyagerGroupPostsResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding group posts: %w", err)
		}

		posts := make([]PostSummary, 0, len(raw.Elements))
		for _, el := range raw.Elements {
			posts = append(posts, PostSummary{
				URN:          el.UpdateMetadata.URN,
				AuthorURN:    el.Actor.URN,
				AuthorName:   el.Actor.Name.Text,
				Text:         el.Commentary.Text.Text,
				Timestamp:    el.CreatedAt,
				LikeCount:    el.SocialDetail.TotalSocialActivityCounts.NumLikes,
				CommentCount: el.SocialDetail.TotalSocialActivityCounts.NumComments,
				ShareCount:   el.SocialDetail.TotalSocialActivityCounts.NumShares,
			})
		}
		return printPostSummaries(cmd, posts)
	}
}

func makeRunGroupsJoin(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		groupID, _ := cmd.Flags().GetString("group-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would join group %s", groupID), map[string]any{
				"action":   "join",
				"group_id": groupID,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/groups/groups/" + url.PathEscape(groupID) + "/members"
		resp, err := client.PostJSON(ctx, path, map[string]any{})
		if err != nil {
			return fmt.Errorf("joining group %s: %w", groupID, err)
		}
		resp.Body.Close()

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"joined": true, "group_id": groupID})
		}
		fmt.Printf("Joined group %s\n", groupID)
		return nil
	}
}

func makeRunGroupsLeave(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		groupID, _ := cmd.Flags().GetString("group-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would leave group %s", groupID), map[string]any{
				"action":   "leave",
				"group_id": groupID,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		// Use the current user's URN as the memberId; the API accepts "me" as a sentinel.
		path := "/voyager/api/groups/groups/" + url.PathEscape(groupID) + "/members/me"
		resp, err := client.Delete(ctx, path)
		if err != nil {
			return fmt.Errorf("leaving group %s: %w", groupID, err)
		}
		resp.Body.Close()

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"left": true, "group_id": groupID})
		}
		fmt.Printf("Left group %s\n", groupID)
		return nil
	}
}

// printGroupSummaries outputs group summaries as JSON or text.
func printGroupSummaries(cmd *cobra.Command, groups []GroupSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(groups)
	}
	if len(groups) == 0 {
		fmt.Println("No groups found.")
		return nil
	}
	lines := make([]string, 0, len(groups)+1)
	lines = append(lines, fmt.Sprintf("%-15s  %-35s  %-10s  %-50s", "ID", "NAME", "MEMBERS", "DESCRIPTION"))
	for _, g := range groups {
		lines = append(lines, fmt.Sprintf("%-15s  %-35s  %-10s  %-50s",
			truncate(g.ID, 15),
			truncate(g.Name, 35),
			formatCount(g.MemberCount),
			truncate(g.Description, 50),
		))
	}
	cli.PrintText(lines)
	return nil
}

// printGroupDetail outputs a single group as JSON or formatted text block.
func printGroupDetail(cmd *cobra.Command, g GroupSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(g)
	}
	lines := []string{
		fmt.Sprintf("ID:          %s", g.ID),
		fmt.Sprintf("Name:        %s", g.Name),
		fmt.Sprintf("Members:     %s", formatCount(g.MemberCount)),
		fmt.Sprintf("Description: %s", g.Description),
	}
	cli.PrintText(lines)
	return nil
}
