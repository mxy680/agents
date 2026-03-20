package linkedin

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerCompanyResponse is the response envelope for a single company lookup.
type voyagerCompanyResponse struct {
	EntityURN     string `json:"entityUrn"`
	Name          string `json:"name"`
	IndustryName  string `json:"industryName"`
	StaffCount    int    `json:"staffCount"`
	FollowerCount int    `json:"followerCount"`
	Description   string `json:"description"`
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
	cmd.AddCommand(newCompaniesFollowCmd(factory))
	cmd.AddCommand(newCompaniesUnfollowCmd(factory))
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

		var raw voyagerCompanyResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding company: %w", err)
		}

		summary := toCompanySummary(raw)
		return printCompanySummary(cmd, summary, raw.Description)
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

// newCompaniesFollowCmd builds the "companies follow" command.
func newCompaniesFollowCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "follow",
		Short: "Follow a company",
		RunE:  makeRunCompaniesFollow(factory),
	}
	cmd.Flags().String("company-id", "", "Company ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without following")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("company-id")
	return cmd
}

func makeRunCompaniesFollow(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		companyID, _ := cmd.Flags().GetString("company-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would follow company %s", companyID), map[string]any{
				"action":     "follow",
				"company_id": companyID,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"urn": fmt.Sprintf("urn:li:company:%s", companyID),
		}
		resp, err := client.PostJSON(ctx, "/voyager/api/feed/follows", body)
		if err != nil {
			return fmt.Errorf("following company %s: %w", companyID, err)
		}
		resp.Body.Close()

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"followed": true, "company_id": companyID})
		}
		fmt.Printf("Now following company %s\n", companyID)
		return nil
	}
}

// newCompaniesUnfollowCmd builds the "companies unfollow" command.
func newCompaniesUnfollowCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfollow",
		Short: "Unfollow a company",
		RunE:  makeRunCompaniesUnfollow(factory),
	}
	cmd.Flags().String("company-id", "", "Company ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without unfollowing")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("company-id")
	return cmd
}

func makeRunCompaniesUnfollow(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		companyID, _ := cmd.Flags().GetString("company-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would unfollow company %s", companyID), map[string]any{
				"action":     "unfollow",
				"company_id": companyID,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/voyager/api/feed/follows/urn:li:company:%s", url.PathEscape(companyID))
		resp, err := client.Delete(ctx, path)
		if err != nil {
			return fmt.Errorf("unfollowing company %s: %w", companyID, err)
		}
		resp.Body.Close()

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"unfollowed": true, "company_id": companyID})
		}
		fmt.Printf("Unfollowed company %s\n", companyID)
		return nil
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
	lines = append(lines, fmt.Sprintf("%-50s  %-40s  %-30s", "URN", "TITLE", "SUBTITLE"))
	for _, r := range results {
		lines = append(lines, fmt.Sprintf("%-50s  %-40s  %-30s",
			truncate(r.URN, 50),
			truncate(r.Title, 40),
			truncate(r.Subtitle, 30),
		))
	}
	cli.PrintText(lines)
	return nil
}
