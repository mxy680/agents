package instagram

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// searchUsersResponse is the response for GET /api/v1/users/search/.
type searchUsersResponse struct {
	Users  []rawUser `json:"users"`
	Status string    `json:"status"`
}

// searchTagsResponse is the response for GET /api/v1/tags/search/.
type searchTagsResponse struct {
	Results []rawTag `json:"results"`
	Status  string   `json:"status"`
}

// rawTag is a minimal hashtag representation from search results.
type rawTag struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	MediaCount     int64  `json:"media_count"`
	FollowingCount int64  `json:"following_count"`
	IsFollowing    bool   `json:"following"`
}

// searchLocationsResponse is the response for GET /api/v1/location_search/.
type searchLocationsResponse struct {
	Venues []rawLocation `json:"venues"`
	Status string        `json:"status"`
}

// rawLocation is a minimal location from the search API.
type rawLocation struct {
	PK         int64   `json:"pk"`
	Name       string  `json:"name"`
	Address    string  `json:"address"`
	City       string  `json:"city"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	MediaCount int64   `json:"media_count"`
}

// topSearchResult represents an entry in the topsearch flat result.
type topSearchResult struct {
	Position int       `json:"position"`
	Type     string    `json:"type"`
	User     *rawUser  `json:"user,omitempty"`
	Hashtag  *rawTag   `json:"hashtag,omitempty"`
	Place    *rawPlace `json:"place,omitempty"`
}

// rawPlace is a place in topsearch results.
type rawPlace struct {
	Location rawLocation `json:"location"`
	Title    string      `json:"title"`
}

// topSearchResponse is the response for GET /api/v1/fbsearch/topsearch_flat/.
type topSearchResponse struct {
	RankedList []topSearchResult `json:"ranked_list"`
	Status     string            `json:"status"`
}

// clearSearchResponse is the response for POST /api/v1/fbsearch/clear_search_history/.
type clearSearchResponse struct {
	Status string `json:"status"`
}

// exploreResponse is the response for GET /api/v1/discover/explore/.
type exploreResponse struct {
	Items      []map[string]any `json:"items"`
	NextMaxID  string           `json:"next_max_id"`
	MoreAvail  bool             `json:"more_available"`
	Status     string           `json:"status"`
}

// newSearchCmd builds the `search` subcommand group.
func newSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "Search users, tags, locations, and explore",
		Aliases: []string{"find"},
	}
	cmd.AddCommand(newSearchUsersCmd(factory))
	cmd.AddCommand(newSearchTagsCmd(factory))
	cmd.AddCommand(newSearchLocationsCmd(factory))
	cmd.AddCommand(newSearchTopCmd(factory))
	cmd.AddCommand(newSearchClearCmd(factory))
	cmd.AddCommand(newSearchExploreCmd(factory))
	return cmd
}

func newSearchUsersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Search for users by username",
		RunE:  makeRunSearchUsers(factory),
	}
	cmd.Flags().String("query", "", "Search query")
	_ = cmd.MarkFlagRequired("query")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	return cmd
}

func makeRunSearchUsers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("q", query)
		params.Set("count", strconv.Itoa(limit))

		resp, err := client.Get(ctx, "/api/v1/users/search/", params)
		if err != nil {
			return fmt.Errorf("searching users: %w", err)
		}

		var result searchUsersResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding search users response: %w", err)
		}

		summaries := toUserSummaries(result.Users)
		return printUserSummaries(cmd, summaries)
	}
}

func newSearchTagsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Search for hashtags",
		RunE:  makeRunSearchTags(factory),
	}
	cmd.Flags().String("query", "", "Search query")
	_ = cmd.MarkFlagRequired("query")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	return cmd
}

func makeRunSearchTags(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("q", query)
		params.Set("count", strconv.Itoa(limit))

		resp, err := client.Get(ctx, "/api/v1/tags/search/", params)
		if err != nil {
			return fmt.Errorf("searching tags: %w", err)
		}

		var result searchTagsResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding search tags response: %w", err)
		}

		tags := result.Results
		if len(tags) > limit {
			tags = tags[:limit]
		}

		summaries := make([]TagSummary, 0, len(tags))
		for _, t := range tags {
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
			fmt.Println("No tags found.")
			return nil
		}

		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-30s  %-12s  %-12s  %-10s", "NAME", "POSTS", "FOLLOWERS", "FOLLOWING"))
		for _, t := range summaries {
			lines = append(lines, fmt.Sprintf("%-30s  %-12s  %-12s  %-10v",
				truncate(t.Name, 30),
				formatCount(t.MediaCount),
				formatCount(t.FollowingCount),
				t.IsFollowing,
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newSearchLocationsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "locations",
		Short: "Search for locations by name",
		RunE:  makeRunSearchLocations(factory),
	}
	cmd.Flags().String("query", "", "Search query")
	_ = cmd.MarkFlagRequired("query")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().Float64("lat", 0, "Latitude hint")
	cmd.Flags().Float64("lng", 0, "Longitude hint")
	return cmd
}

func makeRunSearchLocations(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")
		lat, _ := cmd.Flags().GetFloat64("lat")
		lng, _ := cmd.Flags().GetFloat64("lng")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("search_query", query)
		params.Set("count", strconv.Itoa(limit))
		if lat != 0 {
			params.Set("latitude", strconv.FormatFloat(lat, 'f', 6, 64))
		}
		if lng != 0 {
			params.Set("longitude", strconv.FormatFloat(lng, 'f', 6, 64))
		}

		resp, err := client.Get(ctx, "/api/v1/location_search/", params)
		if err != nil {
			return fmt.Errorf("searching locations: %w", err)
		}

		var result searchLocationsResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding search locations response: %w", err)
		}

		venues := result.Venues
		if len(venues) > limit {
			venues = venues[:limit]
		}

		summaries := make([]LocationSummary, 0, len(venues))
		for _, v := range venues {
			summaries = append(summaries, LocationSummary{
				PK:         v.PK,
				Name:       v.Name,
				Address:    v.Address,
				City:       v.City,
				Lat:        v.Lat,
				Lng:        v.Lng,
				MediaCount: v.MediaCount,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}

		if len(summaries) == 0 {
			fmt.Println("No locations found.")
			return nil
		}

		lines := make([]string, 0, len(summaries)+1)
		lines = append(lines, fmt.Sprintf("%-12s  %-30s  %-30s  %-12s", "ID", "NAME", "CITY", "POSTS"))
		for _, loc := range summaries {
			lines = append(lines, fmt.Sprintf("%-12d  %-30s  %-30s  %-12s",
				loc.PK,
				truncate(loc.Name, 30),
				truncate(loc.City, 30),
				formatCount(loc.MediaCount),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newSearchTopCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "top",
		Short: "Top search results across users, tags, and places",
		RunE:  makeRunSearchTop(factory),
	}
	cmd.Flags().String("query", "", "Search query")
	_ = cmd.MarkFlagRequired("query")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	return cmd
}

func makeRunSearchTop(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("query", query)
		params.Set("count", strconv.Itoa(limit))

		resp, err := client.Get(ctx, "/api/v1/fbsearch/topsearch_flat/", params)
		if err != nil {
			return fmt.Errorf("top search: %w", err)
		}

		var result topSearchResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding top search response: %w", err)
		}

		items := result.RankedList
		if len(items) > limit {
			items = items[:limit]
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(items)
		}

		if len(items) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		lines := make([]string, 0, len(items)+1)
		lines = append(lines, fmt.Sprintf("%-5s  %-10s  %-30s", "POS", "TYPE", "RESULT"))
		for _, item := range items {
			name := ""
			switch item.Type {
			case "user":
				if item.User != nil {
					name = item.User.Username
				}
			case "hashtag":
				if item.Hashtag != nil {
					name = "#" + item.Hashtag.Name
				}
			case "place":
				if item.Place != nil {
					name = item.Place.Title
				}
			}
			lines = append(lines, fmt.Sprintf("%-5d  %-10s  %-30s", item.Position, item.Type, truncate(name, 30)))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newSearchClearCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear search history",
		RunE:  makeRunSearchClear(factory),
	}
	cmd.Flags().Bool("dry-run", false, "Print what would be done without making changes")
	return cmd
}

func makeRunSearchClear(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "clear search history", map[string]string{})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Post(ctx, "/api/v1/fbsearch/clear_search_history/", nil)
		if err != nil {
			return fmt.Errorf("clearing search history: %w", err)
		}

		var result clearSearchResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding clear search response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Println("Search history cleared.")
		return nil
	}
}

func newSearchExploreCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explore",
		Short: "Browse the Explore/Discover feed",
		RunE:  makeRunSearchExplore(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of items")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

func makeRunSearchExplore(factory ClientFactory) func(*cobra.Command, []string) error {
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

		resp, err := client.Get(ctx, "/api/v1/discover/explore/", params)
		if err != nil {
			return fmt.Errorf("fetching explore feed: %w", err)
		}

		var result exploreResponse
		if err := client.DecodeJSON(resp, &result); err != nil {
			return fmt.Errorf("decoding explore response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		fmt.Printf("Explore items: %d\n", len(result.Items))
		if result.MoreAvail && result.NextMaxID != "" {
			fmt.Printf("Next cursor: %s\n", result.NextMaxID)
		}
		return nil
	}
}
