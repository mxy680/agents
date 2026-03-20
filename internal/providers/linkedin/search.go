package linkedin

import (
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerSearchResponse is the response envelope for GET /voyager/api/search/dash/clusters.
type voyagerSearchResponse struct {
	Elements []struct {
		Items []struct {
			Item struct {
				EntityResult struct {
					EntityURN       string `json:"entityUrn"`
					Title           struct{ Text string `json:"text"` } `json:"title"`
					PrimarySubtitle struct{ Text string `json:"text"` } `json:"primarySubtitle"`
				} `json:"entityResult"`
				// Alternate key used by some LinkedIn response shapes
				SearchEntityResult struct {
					EntityURN       string `json:"entityUrn"`
					Title           struct{ Text string `json:"text"` } `json:"title"`
					PrimarySubtitle struct{ Text string `json:"text"` } `json:"primarySubtitle"`
				} `json:"com.linkedin.voyager.search.SearchEntityResult"`
			} `json:"item"`
		} `json:"items"`
		// Outer cluster may also have an "elements" list (some shapes flatten items here)
		Elements []struct {
			Item struct {
				EntityResult struct {
					EntityURN       string `json:"entityUrn"`
					Title           struct{ Text string `json:"text"` } `json:"title"`
					PrimarySubtitle struct{ Text string `json:"text"` } `json:"primarySubtitle"`
				} `json:"entityResult"`
				SearchEntityResult struct {
					EntityURN       string `json:"entityUrn"`
					Title           struct{ Text string `json:"text"` } `json:"title"`
					PrimarySubtitle struct{ Text string `json:"text"` } `json:"primarySubtitle"`
				} `json:"com.linkedin.voyager.search.SearchEntityResult"`
			} `json:"item"`
		} `json:"elements"`
	} `json:"elements"`
	Paging struct {
		Start int `json:"start"`
		Count int `json:"count"`
		Total int `json:"total"`
	} `json:"paging"`
}

// newSearchCmd builds the "search" (alias: find) subcommand group.
func newSearchCmd(factory ClientFactory) *cobra.Command {
	searchCmd := &cobra.Command{
		Use:     "search",
		Short:   "Search LinkedIn for people, companies, jobs, posts, and groups",
		Aliases: []string{"find"},
	}
	searchCmd.AddCommand(newSearchPeopleCmd(factory))
	searchCmd.AddCommand(newSearchCompaniesCmd(factory))
	searchCmd.AddCommand(newSearchJobsCmd(factory))
	searchCmd.AddCommand(newSearchPostsCmd(factory))
	searchCmd.AddCommand(newSearchGroupsCmd(factory))
	return searchCmd
}

// newSearchPeopleCmd builds the "search people" command.
func newSearchPeopleCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "people",
		Short: "Search for LinkedIn members",
		Long:  "Search LinkedIn members by keyword with optional filters.",
		RunE:  makeRunSearch(factory, "person"),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	cmd.Flags().String("network", "", "Network depth filter: F (1st), S (2nd), O (out-of-network)")
	cmd.Flags().String("company", "", "Filter by company ID")
	cmd.Flags().String("location", "", "Filter by location")
	cmd.Flags().String("title", "", "Filter by title")
	cmd.Flags().String("industry", "", "Filter by industry")
	cmd.Flags().Int("limit", 10, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	return cmd
}

// newSearchCompaniesCmd builds the "search companies" command.
func newSearchCompaniesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "companies",
		Short: "Search for LinkedIn companies",
		Long:  "Search LinkedIn companies by keyword with optional filters.",
		RunE:  makeRunSearch(factory, "company"),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	cmd.Flags().String("industry", "", "Filter by industry")
	cmd.Flags().String("size", "", "Filter by company size range")
	cmd.Flags().Int("limit", 10, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	return cmd
}

// newSearchJobsCmd builds the "search jobs" command.
func newSearchJobsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Search for LinkedIn job postings",
		Long:  "Search LinkedIn job postings by keyword with optional location filter.",
		RunE:  makeRunSearch(factory, "job"),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	cmd.Flags().String("location", "", "Filter by location")
	cmd.Flags().Int("limit", 10, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	return cmd
}

// newSearchPostsCmd builds the "search posts" command.
func newSearchPostsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "posts",
		Short: "Search for LinkedIn posts",
		Long:  "Search LinkedIn posts by keyword with optional author filter.",
		RunE:  makeRunSearch(factory, "post"),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	cmd.Flags().String("author", "", "Filter by author URN")
	cmd.Flags().Int("limit", 10, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	return cmd
}

// newSearchGroupsCmd builds the "search groups" command.
func newSearchGroupsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "Search for LinkedIn groups",
		Long:  "Search LinkedIn groups by keyword.",
		RunE:  makeRunSearch(factory, "group"),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	cmd.Flags().Int("limit", 10, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	return cmd
}

// searchTypeToOrigin maps result types to the LinkedIn search origin query parameter.
var searchTypeToOrigin = map[string]string{
	"person":  "FACETED_SEARCH",
	"company": "FACETED_SEARCH",
	"job":     "JOB_SEARCH",
	"post":    "FACETED_SEARCH",
	"group":   "FACETED_SEARCH",
}

// searchTypeToQuery maps result types to the LinkedIn q parameter.
var searchTypeToQuery = map[string]string{
	"person":  "SEARCH_HITS",
	"company": "SEARCH_HITS",
	"job":     "SEARCH_HITS",
	"post":    "SEARCH_HITS",
	"group":   "SEARCH_HITS",
}

// makeRunSearch builds a RunE function for the given search type.
func makeRunSearch(factory ClientFactory, resultType string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		if query == "" {
			return fmt.Errorf("--query is required")
		}

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
			"q":      {searchTypeToQuery[resultType]},
			"origin": {searchTypeToOrigin[resultType]},
			"query":  {fmt.Sprintf(`(keywords:%s)`, query)},
			"start":  {fmt.Sprintf("%d", start)},
			"count":  {fmt.Sprintf("%d", limit)},
		}

		// Add optional filters when present
		if loc, _ := cmd.Flags().GetString("location"); loc != "" {
			params.Set("filters", fmt.Sprintf("List((filter:geoUrn,values:List((text:%s))))", loc))
		}
		if network, _ := cmd.Flags().GetString("network"); network != "" {
			params.Set("filters", fmt.Sprintf("List((filter:network,values:List((value:%s))))", network))
		}
		if author, _ := cmd.Flags().GetString("author"); author != "" {
			params.Set("filters", fmt.Sprintf("List((filter:authorUrn,values:List((value:%s))))", author))
		}

		resp, err := client.Get(ctx, "/voyager/api/search/dash/clusters", params)
		if err != nil {
			return fmt.Errorf("searching %ss: %w", resultType, err)
		}

		var raw voyagerSearchResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding search results: %w", err)
		}

		results := extractSearchResults(raw, resultType)
		return printSearchResults(cmd, results)
	}
}

// extractSearchResults flattens the nested search response into a SearchResult slice.
func extractSearchResults(raw voyagerSearchResponse, resultType string) []SearchResult {
	results := make([]SearchResult, 0)
	for _, cluster := range raw.Elements {
		// Handle shape 1: cluster.items
		for _, item := range cluster.Items {
			if r := toSearchResult(item.Item.EntityResult.EntityURN,
				item.Item.EntityResult.Title.Text,
				item.Item.EntityResult.PrimarySubtitle.Text,
				resultType); r != nil {
				results = append(results, *r)
				continue
			}
			if r := toSearchResult(item.Item.SearchEntityResult.EntityURN,
				item.Item.SearchEntityResult.Title.Text,
				item.Item.SearchEntityResult.PrimarySubtitle.Text,
				resultType); r != nil {
				results = append(results, *r)
			}
		}
		// Handle shape 2: cluster.elements
		for _, el := range cluster.Elements {
			if r := toSearchResult(el.Item.EntityResult.EntityURN,
				el.Item.EntityResult.Title.Text,
				el.Item.EntityResult.PrimarySubtitle.Text,
				resultType); r != nil {
				results = append(results, *r)
				continue
			}
			if r := toSearchResult(el.Item.SearchEntityResult.EntityURN,
				el.Item.SearchEntityResult.Title.Text,
				el.Item.SearchEntityResult.PrimarySubtitle.Text,
				resultType); r != nil {
				results = append(results, *r)
			}
		}
	}
	return results
}

// toSearchResult returns a SearchResult if urn is non-empty, otherwise nil.
func toSearchResult(urn, title, subtitle, resultType string) *SearchResult {
	if urn == "" {
		return nil
	}
	return &SearchResult{
		URN:      urn,
		Title:    title,
		Subtitle: subtitle,
		Type:     resultType,
	}
}

// printSearchResults outputs search results as JSON or text.
func printSearchResults(cmd *cobra.Command, results []SearchResult) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(results)
	}
	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}
	lines := make([]string, 0, len(results)+1)
	lines = append(lines, fmt.Sprintf("%-45s  %-30s  %-40s  %-10s", "URN", "TITLE", "SUBTITLE", "TYPE"))
	for _, r := range results {
		lines = append(lines, fmt.Sprintf("%-45s  %-30s  %-40s  %-10s",
			truncate(r.URN, 45),
			truncate(r.Title, 30),
			truncate(r.Subtitle, 40),
			r.Type,
		))
	}
	cli.PrintText(lines)
	return nil
}
