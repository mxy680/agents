package linkedin

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerCompanyResponse is the response envelope for a single company lookup.
// Used by unit tests (TestToCompanySummary) for mapping convenience.
type voyagerCompanyResponse struct {
	EntityURN     string `json:"entityUrn"`
	Name          string `json:"name"`
	IndustryName  string `json:"industryName"`
	StaffCount    int    `json:"staffCount"`
	FollowerCount int    `json:"followerCount"`
	Description   string `json:"description"`
}

// companyEntity represents the normalized company shape from included[].
type companyEntity struct {
	EntityURN     string `json:"entityUrn"`
	Name          string `json:"name"`
	UniversalName string `json:"universalName"`
	StaffCount    int    `json:"staffCount"`
	// Industries is an array; we use the first entry's localizedName.
	CompanyIndustries []struct {
		LocalizedName string `json:"localizedName"`
	} `json:"companyIndustries"`
}

// voyagerCompanySearchResponse is the response envelope for company search via dash/clusters.
type voyagerCompanySearchResponse struct {
	Elements []voyagerSearchCluster `json:"elements"`
}

type voyagerSearchCluster struct {
	Elements []voyagerSearchItem `json:"elements"`
}

type voyagerSearchItem struct {
	Item voyagerSearchItemWrap `json:"item"`
}

type voyagerSearchItemWrap struct {
	EntityResult voyagerEntityResult `json:"com.linkedin.voyager.search.SearchEntityResult"`
}

type voyagerEntityResult struct {
	EntityUrn string            `json:"entityUrn"`
	Title     voyagerTextObject `json:"title"`
	Subtitle  voyagerTextObject `json:"primarySubtitle"`
}

type voyagerTextObject struct {
	Text string `json:"text"`
}

// voyagerJobPostingsResponse is used for company jobs endpoint.
type voyagerJobPostingsResponse struct {
	Elements []voyagerJobPosting `json:"elements"`
	Paging   voyagerPaging       `json:"paging"`
}

type voyagerJobPosting struct {
	EntityURN          string `json:"entityUrn"`
	Title              string `json:"title"`
	CompanyName        string `json:"companyName"`
	FormattedLocation  string `json:"formattedLocation"`
	ListedAt           int64  `json:"listedAt"`
	WorkRemoteAllowed  bool   `json:"workRemoteAllowed"`
}

// toCompanySummary maps a raw company response to CompanySummary.
func toCompanySummary(raw voyagerCompanyResponse) CompanySummary {
	// Extract numeric ID from URN like "urn:li:fs_normalized_company:1234"
	id := raw.EntityURN
	if parts := strings.Split(raw.EntityURN, ":"); len(parts) > 0 {
		id = parts[len(parts)-1]
	}
	return CompanySummary{
		ID:            id,
		Name:          raw.Name,
		Industry:      raw.IndustryName,
		EmployeeCount: raw.StaffCount,
		FollowerCount: raw.FollowerCount,
	}
}

// toJobSummaryFromPosting maps a voyagerJobPosting to JobSummary.
func toJobSummaryFromPosting(raw voyagerJobPosting) JobSummary {
	id := raw.EntityURN
	if parts := strings.Split(raw.EntityURN, ":"); len(parts) > 0 {
		id = parts[len(parts)-1]
	}
	remote := ""
	if raw.WorkRemoteAllowed {
		remote = "remote"
	}
	return JobSummary{
		ID:       id,
		Title:    raw.Title,
		Company:  raw.CompanyName,
		Location: raw.FormattedLocation,
		PostedAt: raw.ListedAt,
		Remote:   remote,
	}
}

// newCompaniesCmd builds the "companies" subcommand group.
func newCompaniesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "companies",
		Short:   "Interact with LinkedIn companies",
		Aliases: []string{"company", "org"},
	}
	cmd.AddCommand(newCompaniesGetCmd(factory))
	cmd.AddCommand(newCompaniesSearchCmd(factory))
	cmd.AddCommand(newCompaniesEmployeesCmd(factory))
	cmd.AddCommand(newCompaniesJobsCmd(factory))
	return cmd
}

// newCompaniesGetCmd builds the "companies get" command.
func newCompaniesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a company by ID",
		RunE:  makeRunCompaniesGet(factory),
	}
	cmd.Flags().String("company-id", "", "Company ID or universal name slug (required)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("company-id")
	return cmd
}

func makeRunCompaniesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		companyID, _ := cmd.Flags().GetString("company-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("q", "universalName")
		params.Set("universalName", companyID)

		resp, err := client.Get(ctx, "/voyager/api/organization/companies", params)
		if err != nil {
			return fmt.Errorf("getting company %s: %w", companyID, err)
		}

		var normalized NormalizedResponse
		if err := client.DecodeJSON(resp, &normalized); err != nil {
			return fmt.Errorf("decoding company: %w", err)
		}

		rawJSON := FindIncluded(normalized.Included, "Company")
		if rawJSON == nil {
			// Fallback: try to parse as legacy flat response.
			// Re-decode data field as company entity.
			var legacy voyagerCompanyResponse
			if err := json.Unmarshal(normalized.Data, &legacy); err == nil && legacy.Name != "" {
				summary := toCompanySummary(legacy)
				return printCompanySummary(cmd, summary, legacy.Description)
			}
			return fmt.Errorf("company %q not found in response", companyID)
		}

		var entity companyEntity
		if err := json.Unmarshal(rawJSON, &entity); err != nil {
			return fmt.Errorf("parsing company entity: %w", err)
		}

		id := entity.EntityURN
		if parts := strings.Split(entity.EntityURN, ":"); len(parts) > 0 {
			id = parts[len(parts)-1]
		}
		industry := ""
		if len(entity.CompanyIndustries) > 0 {
			industry = entity.CompanyIndustries[0].LocalizedName
		}
		summary := CompanySummary{
			ID:            id,
			Name:          entity.Name,
			Industry:      industry,
			EmployeeCount: entity.StaffCount,
		}
		return printCompanySummary(cmd, summary, "")
	}
}

// newCompaniesSearchCmd builds the "companies search" command.
func newCompaniesSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for companies",
		RunE:  makeRunCompaniesSearch(factory),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	cmd.Flags().Int("limit", 10, "Maximum number of results to return")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("query")
	return cmd
}

func makeRunCompaniesSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("q", "all")
		params.Set("query", fmt.Sprintf("(keywords:%s,filterValues:List((id:resultType,values:List(COMPANIES))))", url.QueryEscape(query)))
		params.Set("count", fmt.Sprintf("%d", limit))

		resp, err := client.Get(ctx, "/voyager/api/search/dash/clusters", params)
		if err != nil {
			return fmt.Errorf("searching companies for %q: %w", query, err)
		}

		var raw voyagerCompanySearchResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding company search results: %w", err)
		}

		results := make([]SearchResult, 0)
		for _, cluster := range raw.Elements {
			for _, item := range cluster.Elements {
				er := item.Item.EntityResult
				results = append(results, SearchResult{
					URN:      er.EntityUrn,
					Title:    er.Title.Text,
					Subtitle: er.Subtitle.Text,
					Type:     "company",
				})
			}
		}
		return printSearchResults(cmd, results)
	}
}

// newCompaniesEmployeesCmd builds the "companies employees" command.
func newCompaniesEmployeesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "employees",
		Short: "List employees of a company",
		RunE:  makeRunCompaniesEmployees(factory),
	}
	cmd.Flags().String("company-id", "", "Company ID (required)")
	cmd.Flags().Int("limit", 10, "Maximum number of results to return")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("company-id")
	return cmd
}

func makeRunCompaniesEmployees(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		companyID, _ := cmd.Flags().GetString("company-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("q", "all")
		params.Set("query", fmt.Sprintf("(keywords:*,filterValues:List((id:currentCompany,values:List(%s))))", url.QueryEscape(companyID)))
		params.Set("count", fmt.Sprintf("%d", limit))
		if cursor != "" {
			params.Set("start", cursor)
		}

		resp, err := client.Get(ctx, "/voyager/api/search/dash/clusters", params)
		if err != nil {
			return fmt.Errorf("listing employees for company %s: %w", companyID, err)
		}

		var raw voyagerCompanySearchResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding employee results: %w", err)
		}

		results := make([]SearchResult, 0)
		for _, cluster := range raw.Elements {
			for _, item := range cluster.Elements {
				er := item.Item.EntityResult
				results = append(results, SearchResult{
					URN:      er.EntityUrn,
					Title:    er.Title.Text,
					Subtitle: er.Subtitle.Text,
					Type:     "person",
				})
			}
		}
		return printSearchResults(cmd, results)
	}
}

// newCompaniesJobsCmd builds the "companies jobs" command.
func newCompaniesJobsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "List job postings at a company",
		RunE:  makeRunCompaniesJobs(factory),
	}
	cmd.Flags().String("company-id", "", "Company universal name slug (required)")
	cmd.Flags().Int("limit", 10, "Maximum number of jobs to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("company-id")
	return cmd
}

func makeRunCompaniesJobs(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		companyID, _ := cmd.Flags().GetString("company-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("companyUniversalName", companyID)
		params.Set("count", fmt.Sprintf("%d", limit))
		if cursor != "" {
			params.Set("start", cursor)
		}

		resp, err := client.Get(ctx, "/voyager/api/jobs/jobPostings", params)
		if err != nil {
			return fmt.Errorf("listing jobs for company %s: %w", companyID, err)
		}

		var raw voyagerJobPostingsResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding company jobs: %w", err)
		}

		summaries := make([]JobSummary, 0, len(raw.Elements))
		for _, el := range raw.Elements {
			summaries = append(summaries, toJobSummaryFromPosting(el))
		}
		return printJobSummaries(cmd, summaries)
	}
}

// printCompanySummary outputs a company summary as JSON or formatted text.
func printCompanySummary(cmd *cobra.Command, c CompanySummary, description string) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(c)
	}
	lines := []string{
		fmt.Sprintf("ID:        %s", c.ID),
		fmt.Sprintf("Name:      %s", c.Name),
		fmt.Sprintf("Industry:  %s", c.Industry),
		fmt.Sprintf("Employees: %s", formatCount(c.EmployeeCount)),
		fmt.Sprintf("Followers: %s", formatCount(c.FollowerCount)),
	}
	if description != "" {
		lines = append(lines, fmt.Sprintf("About:     %s", truncate(description, 200)))
	}
	cli.PrintText(lines)
	return nil
}
