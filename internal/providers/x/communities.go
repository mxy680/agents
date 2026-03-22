package x

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// GraphQL query hashes for community operations.
const (
	hashCommunityQuery              = "lUBKrilodgg9Nikaw3cIiA"
	hashCommunitiesSearchQuery      = "daVUkhfHn7-Z8llpYVKJSw"
	hashCommunityTweetsTimeline     = "mhwSsmub4JZgHcs0dtsjrw"
	hashCommunityMediaTimeline      = "Ht5K2ckaZYAOuRFmFfbHig"
	hashMembersSliceTimeline        = "KDAssJ5lafCy-asH4wm1dw"
	hashModeratorsSliceTimeline     = "9KI_r8e-tgp3--N5SZYVjg"
	hashCommunitiesMainPageTimeline = "4-4iuIdaLPpmxKnA3mr2LA"
	hashJoinCommunity               = "xZQLbDwbI585YTG0QIpokw"
	hashLeaveCommunity              = "OoS6Kd4-noNLXPZYHtygeA"
	hashRequestToJoinCommunity      = "XwWChphD_6g7JnsFus2f2Q"
	hashCommunityTweetSearch        = "5341rmzzvdjqfmPKfoHUBw"
)

// CommunitySummary is a condensed representation of an X community.
type CommunitySummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MemberCount int    `json:"member_count"`
	CreatedAt   string `json:"created_at,omitempty"`
}

// newCommunitiesCmd builds the "communities" subcommand group.
func newCommunitiesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "communities",
		Short:   "Interact with X communities",
		Aliases: []string{"community"},
	}
	cmd.AddCommand(newCommunitiesGetCmd(factory))
	cmd.AddCommand(newCommunitiesSearchCmd(factory))
	cmd.AddCommand(newCommunityTweetsCmd(factory))
	cmd.AddCommand(newCommunityMediaCmd(factory))
	cmd.AddCommand(newCommunityMembersCmd(factory))
	cmd.AddCommand(newCommunityModeratorsCmd(factory))
	cmd.AddCommand(newCommunitiesTimelineCmd(factory))
	cmd.AddCommand(newCommunityJoinCmd(factory))
	cmd.AddCommand(newCommunityLeaveCmd(factory))
	cmd.AddCommand(newCommunityRequestJoinCmd(factory))
	cmd.AddCommand(newCommunitySearchTweetsCmd(factory))
	return cmd
}

func newCommunitiesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a community by ID",
		RunE:  makeRunCommunitiesGet(factory),
	}
	cmd.Flags().String("community-id", "", "Community ID (required)")
	_ = cmd.MarkFlagRequired("community-id")
	return cmd
}

func newCommunitiesSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for communities",
		RunE:  makeRunCommunitiesSearch(factory),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	_ = cmd.MarkFlagRequired("query")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	return cmd
}

func newCommunityTweetsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tweets",
		Short: "List tweets in a community",
		RunE:  makeRunCommunityTweets(factory),
	}
	cmd.Flags().String("community-id", "", "Community ID (required)")
	_ = cmd.MarkFlagRequired("community-id")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func newCommunityMediaCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "media",
		Short: "List media in a community",
		RunE:  makeRunCommunityMedia(factory),
	}
	cmd.Flags().String("community-id", "", "Community ID (required)")
	_ = cmd.MarkFlagRequired("community-id")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func newCommunityMembersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members",
		Short: "List members of a community",
		RunE:  makeRunCommunityMembers(factory),
	}
	cmd.Flags().String("community-id", "", "Community ID (required)")
	_ = cmd.MarkFlagRequired("community-id")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func newCommunityModeratorsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "moderators",
		Short: "List moderators of a community",
		RunE:  makeRunCommunityModerators(factory),
	}
	cmd.Flags().String("community-id", "", "Community ID (required)")
	_ = cmd.MarkFlagRequired("community-id")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func newCommunitiesTimelineCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "timeline",
		Short: "Get the communities main page timeline",
		RunE:  makeRunCommunitiesTimeline(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func newCommunityJoinCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join",
		Short: "Join a community",
		RunE:  makeRunCommunityJoin(factory),
	}
	cmd.Flags().String("community-id", "", "Community ID (required)")
	_ = cmd.MarkFlagRequired("community-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without joining")
	return cmd
}

func newCommunityLeaveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leave",
		Short: "Leave a community",
		RunE:  makeRunCommunityLeave(factory),
	}
	cmd.Flags().String("community-id", "", "Community ID (required)")
	_ = cmd.MarkFlagRequired("community-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without leaving")
	return cmd
}

func newCommunityRequestJoinCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-join",
		Short: "Request to join a private community",
		RunE:  makeRunCommunityRequestJoin(factory),
	}
	cmd.Flags().String("community-id", "", "Community ID (required)")
	_ = cmd.MarkFlagRequired("community-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without requesting")
	return cmd
}

func newCommunitySearchTweetsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search-tweets",
		Short: "Search tweets within a community",
		RunE:  makeRunCommunitySearchTweets(factory),
	}
	cmd.Flags().String("community-id", "", "Community ID (required)")
	_ = cmd.MarkFlagRequired("community-id")
	cmd.Flags().String("query", "", "Search query (required)")
	_ = cmd.MarkFlagRequired("query")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	return cmd
}

// --- RunE implementations ---

func makeRunCommunitiesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		communityID, _ := cmd.Flags().GetString("community-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"communityId": communityID,
			"withDMConversationId": true,
		}

		data, err := client.GraphQL(ctx, hashCommunityQuery, "CommunityQuery", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("getting community %s: %w", communityID, err)
		}

		community, err := parseCommunityResult(data)
		if err != nil {
			// Fall back to raw JSON output.
			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(data)
			}
			return fmt.Errorf("parse community: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(community)
		}

		lines := []string{
			fmt.Sprintf("ID:      %s", community.ID),
			fmt.Sprintf("Name:    %s", community.Name),
			fmt.Sprintf("Members: %d", community.MemberCount),
		}
		if community.Description != "" {
			lines = append(lines, fmt.Sprintf("Desc:    %s", truncate(community.Description, 120)))
		}
		cli.PrintText(lines)
		return nil
	}
}

func makeRunCommunitiesSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"query": query,
			"count": limit,
		}

		data, err := client.GraphQL(ctx, hashCommunitiesSearchQuery, "CommunitiesSearchQuery", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("searching communities %q: %w", query, err)
		}

		communities, err := parseCommunityList(data)
		if err != nil || len(communities) == 0 {
			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(data)
			}
			fmt.Println("No communities found.")
			return nil
		}

		return printCommunitySummaries(cmd, communities)
	}
}

func makeRunCommunityTweets(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		communityID, _ := cmd.Flags().GetString("community-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"communityId": communityID,
			"count":       limit,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashCommunityTweetsTimeline, "CommunityTweetsTimeline", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching community tweets for %s: %w", communityID, err)
		}

		return printTimelineResult(cmd, data)
	}
}

func makeRunCommunityMedia(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		communityID, _ := cmd.Flags().GetString("community-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"communityId": communityID,
			"count":       limit,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashCommunityMediaTimeline, "CommunityMediaTimeline", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching community media for %s: %w", communityID, err)
		}

		return printTimelineResult(cmd, data)
	}
}

func makeRunCommunityMembers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		communityID, _ := cmd.Flags().GetString("community-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"communityId": communityID,
			"count":       limit,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashMembersSliceTimeline, "membersSliceTimeline_Query", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching community members for %s: %w", communityID, err)
		}

		return printUserListResult(cmd, data)
	}
}

func makeRunCommunityModerators(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		communityID, _ := cmd.Flags().GetString("community-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"communityId": communityID,
			"count":       limit,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashModeratorsSliceTimeline, "moderatorsSliceTimeline_Query", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching community moderators for %s: %w", communityID, err)
		}

		return printUserListResult(cmd, data)
	}
}

func makeRunCommunitiesTimeline(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"count": limit,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashCommunitiesMainPageTimeline, "CommunitiesMainPageTimeline", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching communities timeline: %w", err)
		}

		return printTimelineResult(cmd, data)
	}
}

func makeRunCommunityJoin(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		communityID, _ := cmd.Flags().GetString("community-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("join community %s", communityID), map[string]string{"community_id": communityID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"communityId": communityID,
		}

		_, err = client.GraphQLPost(ctx, hashJoinCommunity, "JoinCommunity", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("joining community %s: %w", communityID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "joined", "community_id": communityID})
		}
		fmt.Printf("Joined community: %s\n", communityID)
		return nil
	}
}

func makeRunCommunityLeave(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		communityID, _ := cmd.Flags().GetString("community-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("leave community %s", communityID), map[string]string{"community_id": communityID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"communityId": communityID,
		}

		_, err = client.GraphQLPost(ctx, hashLeaveCommunity, "LeaveCommunity", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("leaving community %s: %w", communityID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "left", "community_id": communityID})
		}
		fmt.Printf("Left community: %s\n", communityID)
		return nil
	}
}

func makeRunCommunityRequestJoin(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		communityID, _ := cmd.Flags().GetString("community-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("request to join community %s", communityID), map[string]string{"community_id": communityID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"communityId": communityID,
		}

		_, err = client.GraphQLPost(ctx, hashRequestToJoinCommunity, "RequestToJoinCommunity", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("requesting to join community %s: %w", communityID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "requested", "community_id": communityID})
		}
		fmt.Printf("Join request sent for community: %s\n", communityID)
		return nil
	}
}

func makeRunCommunitySearchTweets(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		communityID, _ := cmd.Flags().GetString("community-id")
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"communityId": communityID,
			"rawQuery":    query,
			"count":       limit,
		}

		data, err := client.GraphQL(ctx, hashCommunityTweetSearch, "CommunityTweetSearchModuleQuery", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("searching community tweets for %s: %w", communityID, err)
		}

		return printTimelineResult(cmd, data)
	}
}

// parseCommunityResult parses a community from a GraphQL response.
func parseCommunityResult(data json.RawMessage) (*CommunitySummary, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, fmt.Errorf("parse community data: %w", err)
	}

	// Walk looking for a community object with id_str or rest_id.
	for _, v := range top {
		var community struct {
			IDStr       string `json:"id_str"`
			RestID      string `json:"rest_id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			MemberCount int    `json:"member_count"`
			CreatedAt   string `json:"created_at"`
		}
		if err := json.Unmarshal(v, &community); err != nil {
			continue
		}
		id := community.IDStr
		if id == "" {
			id = community.RestID
		}
		if id != "" {
			return &CommunitySummary{
				ID:          id,
				Name:        community.Name,
				Description: community.Description,
				MemberCount: community.MemberCount,
				CreatedAt:   community.CreatedAt,
			}, nil
		}
	}

	return nil, fmt.Errorf("community not found in response")
}

// parseCommunityList extracts a list of communities from a search response.
func parseCommunityList(data json.RawMessage) ([]CommunitySummary, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, fmt.Errorf("parse community list: %w", err)
	}

	var communities []CommunitySummary

	for _, v := range top {
		var items []struct {
			IDStr       string `json:"id_str"`
			RestID      string `json:"rest_id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			MemberCount int    `json:"member_count"`
			CreatedAt   string `json:"created_at"`
		}
		if err := json.Unmarshal(v, &items); err == nil && len(items) > 0 {
			for _, item := range items {
				id := item.IDStr
				if id == "" {
					id = item.RestID
				}
				if id != "" {
					communities = append(communities, CommunitySummary{
						ID:          id,
						Name:        item.Name,
						Description: item.Description,
						MemberCount: item.MemberCount,
						CreatedAt:   item.CreatedAt,
					})
				}
			}
			if len(communities) > 0 {
				return communities, nil
			}
		}
	}

	return communities, nil
}

// printCommunitySummaries outputs community summaries as JSON or text.
func printCommunitySummaries(cmd *cobra.Command, communities []CommunitySummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(communities)
	}
	if len(communities) == 0 {
		fmt.Println("No communities found.")
		return nil
	}
	lines := make([]string, 0, len(communities)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-30s  %-10s", "ID", "NAME", "MEMBERS"))
	for _, c := range communities {
		lines = append(lines, fmt.Sprintf("%-20s  %-30s  %-10d",
			truncate(c.ID, 20),
			truncate(c.Name, 30),
			c.MemberCount,
		))
	}
	cli.PrintText(lines)
	return nil
}
