package linkedin

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// searchClustersQueryID is the current known queryId for the search clusters GraphQL endpoint.
const searchClustersQueryID = "voyagerSearchDashClusters.05111e1b90ee7fea15bebe9f9410ced9"

// searchEntityResult maps the EntityResultViewModel shape returned in GraphQL included[].
type searchEntityResult struct {
	Title struct {
		Text string `json:"text"`
	} `json:"title"`
	PrimarySubtitle struct {
		Text string `json:"text"`
	} `json:"primarySubtitle"`
	TrackingURN string `json:"trackingUrn"`
	EntityURN   string `json:"entityUrn"`
}

// searchTypeToResultType maps our result type names to LinkedIn GraphQL result type values.
var searchTypeToResultType = map[string]string{
	"person":  "PEOPLE",
	"company": "COMPANIES",
	"job":     "JOBS",
	"post":    "CONTENT",
	"group":   "GROUPS",
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

// makeRunSearch builds a RunE function for the given search type.
func makeRunSearch(factory ClientFactory, resultType string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		if query == "" {
			return fmt.Errorf("--query is required")
		}

		_, _ = cmd.Flags().GetInt("limit") // limit flag kept for API compatibility; not used in GraphQL variables
		cursor, _ := cmd.Flags().GetString("cursor")

		start := 0
		if cursor != "" {
			if _, err := fmt.Sscanf(cursor, "%d", &start); err != nil {
				return fmt.Errorf("invalid cursor %q: must be a numeric start offset", cursor)
			}
		}

		linkedinType := searchTypeToResultType[resultType]

		// Build the Rest-li tuple variables string for the GraphQL request.
		variables := fmt.Sprintf(
			"(start:%d,origin:GLOBAL_SEARCH_HEADER,query:(keywords:%s,flagshipSearchIntent:SEARCH_SRP,queryParameters:List((key:resultType,value:List(%s))),includeFiltersInResponse:false))",
			start, query, linkedinType,
		)

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.GetGraphQL(ctx, searchClustersQueryID, variables)
		if err != nil {
			return fmt.Errorf("searching %ss: %w", resultType, err)
		}

		var normalized NormalizedResponse
		if err := client.DecodeJSON(resp, &normalized); err != nil {
			return fmt.Errorf("decoding search results: %w", err)
		}

		results := extractGraphQLSearchResults(normalized.Included, resultType)
		return printSearchResults(cmd, results)
	}
}

// extractGraphQLSearchResults finds all EntityResultViewModel entities in the included array
// and maps them to SearchResult values.
func extractGraphQLSearchResults(included []json.RawMessage, resultType string) []SearchResult {
	rawEntities := FindAllIncluded(included, "EntityResultViewModel")
	results := make([]SearchResult, 0, len(rawEntities))
	for _, raw := range rawEntities {
		var entity searchEntityResult
		if err := json.Unmarshal(raw, &entity); err != nil {
			continue
		}
		// Prefer entityUrn; fall back to trackingUrn.
		urn := entity.EntityURN
		if urn == "" {
			urn = entity.TrackingURN
		}
		if urn == "" {
			continue
		}
		results = append(results, SearchResult{
			URN:      urn,
			Title:    entity.Title.Text,
			Subtitle: entity.PrimarySubtitle.Text,
			Type:     resultType,
		})
	}
	return results
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
