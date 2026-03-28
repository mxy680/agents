package linkedin

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerJobPostingDetailResponse is the response for a single job posting.
type voyagerJobPostingDetailResponse struct {
	EntityURN         string `json:"entityUrn"`
	Title             string `json:"title"`
	CompanyName       string `json:"companyName"`
	FormattedLocation string `json:"formattedLocation"`
	ListedAt          int64  `json:"listedAt"`
	WorkRemoteAllowed bool   `json:"workRemoteAllowed"`
	Description       struct {
		Text string `json:"text"`
	} `json:"description"`
}

// voyagerSavedJobsResponse is the response for listing saved jobs.
type voyagerSavedJobsResponse struct {
	Elements []voyagerSavedJobElement `json:"elements"`
	Paging   voyagerPaging            `json:"paging"`
}

type voyagerSavedJobElement struct {
	EntityURN  string             `json:"entityUrn"`
	JobPosting voyagerJobPosting  `json:"jobPosting"`
}

// voyagerJobSearchResponse is the response for job search via dash/clusters.
type voyagerJobSearchResponse struct {
	Elements []voyagerJobSearchCluster `json:"elements"`
}

type voyagerJobSearchCluster struct {
	Elements []voyagerJobSearchItem `json:"elements"`
}

type voyagerJobSearchItem struct {
	Item voyagerJobSearchItemWrap `json:"item"`
}

type voyagerJobSearchItemWrap struct {
	JobPosting voyagerJobPosting `json:"com.linkedin.voyager.jobs.JobPosting"`
}

// voyagerRecommendedJobsResponse is the response for recommended jobs.
type voyagerRecommendedJobsResponse struct {
	Elements []voyagerJobPosting `json:"elements"`
	Paging   voyagerPaging       `json:"paging"`
}

// toJobSummaryFromDetail maps a voyagerJobPostingDetailResponse to JobSummary.
func toJobSummaryFromDetail(raw voyagerJobPostingDetailResponse) JobSummary {
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

// newJobsCmd builds the "jobs" subcommand group.
func newJobsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jobs",
		Short:   "Interact with LinkedIn job listings",
		Aliases: []string{"job"},
	}
	cmd.AddCommand(newJobsSearchCmd(factory))
	cmd.AddCommand(newJobsGetCmd(factory))
	cmd.AddCommand(newJobsSavedCmd(factory))
	cmd.AddCommand(newJobsRecommendedCmd(factory))
	cmd.AddCommand(newJobsSaveCmd(factory))
	cmd.AddCommand(newJobsUnsaveCmd(factory))
	return cmd
}

// newJobsSearchCmd builds the "jobs search" command.
func newJobsSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for job listings",
		RunE:  makeRunJobsSearch(factory),
	}
	cmd.Flags().String("query", "", "Search keywords (required)")
	cmd.Flags().String("location", "", "Location filter")
	cmd.Flags().String("experience", "", "Experience level: ENTRY_LEVEL, ASSOCIATE, MID_SENIOR_LEVEL, DIRECTOR, EXECUTIVE")
	cmd.Flags().String("type", "", "Job type: FULL_TIME, PART_TIME, CONTRACT, TEMPORARY, INTERNSHIP")
	cmd.Flags().String("remote", "", "Remote preference: ON_SITE, REMOTE, HYBRID")
	cmd.Flags().Int("limit", 10, "Maximum number of results to return")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("query")
	return cmd
}

func makeRunJobsSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		location, _ := cmd.Flags().GetString("location")
		experience, _ := cmd.Flags().GetString("experience")
		jobType, _ := cmd.Flags().GetString("type")
		remote, _ := cmd.Flags().GetString("remote")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		filterParts := []string{fmt.Sprintf("keywords:%s", url.QueryEscape(query))}
		if location != "" {
			filterParts = append(filterParts, fmt.Sprintf("locationFallback:%s", url.QueryEscape(location)))
		}

		filterValues := []string{}
		if experience != "" {
			filterValues = append(filterValues, fmt.Sprintf("(id:experience,values:List(%s))", url.QueryEscape(experience)))
		}
		if jobType != "" {
			filterValues = append(filterValues, fmt.Sprintf("(id:jobType,values:List(%s))", url.QueryEscape(jobType)))
		}
		if remote != "" {
			filterValues = append(filterValues, fmt.Sprintf("(id:workplaceType,values:List(%s))", url.QueryEscape(remote)))
		}

		queryStr := "(" + strings.Join(filterParts, ",")
		if len(filterValues) > 0 {
			queryStr += ",filterValues:List(" + strings.Join(filterValues, ",") + ")"
		}
		queryStr += ")"

		params := url.Values{}
		params.Set("q", "all")
		params.Set("query", queryStr)
		params.Set("count", fmt.Sprintf("%d", limit))
		if cursor != "" {
			params.Set("start", cursor)
		}

		resp, err := client.Get(ctx, "/voyager/api/search/dash/clusters", params)
		if err != nil {
			return fmt.Errorf("searching jobs for %q: %w", query, err)
		}

		var raw voyagerJobSearchResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding job search results: %w", err)
		}

		summaries := make([]JobSummary, 0)
		for _, cluster := range raw.Elements {
			for _, item := range cluster.Elements {
				jp := item.Item.JobPosting
				summaries = append(summaries, toJobSummaryFromPosting(jp))
			}
		}
		return printJobSummaries(cmd, summaries)
	}
}

// newJobsGetCmd builds the "jobs get" command.
func newJobsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a job posting by ID",
		RunE:  makeRunJobsGet(factory),
	}
	cmd.Flags().String("job-id", "", "Job posting ID (required)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("job-id")
	return cmd
}

func makeRunJobsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		jobID, _ := cmd.Flags().GetString("job-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/jobs/jobPostings/" + url.PathEscape(jobID)
		resp, err := client.Get(ctx, path, nil)
		if err != nil {
			return fmt.Errorf("getting job %s: %w", jobID, err)
		}

		var raw voyagerJobPostingDetailResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding job posting: %w", err)
		}

		summary := toJobSummaryFromDetail(raw)
		return printJobDetail(cmd, summary, raw.Description.Text)
	}
}

// newJobsSavedCmd builds the "jobs saved" command.
func newJobsSavedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "saved",
		Short: "List saved job postings",
		RunE:  makeRunJobsSaved(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of saved jobs to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunJobsSaved(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

// newJobsRecommendedCmd builds the "jobs recommended" command.
func newJobsRecommendedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recommended",
		Short: "List recommended job postings",
		RunE:  makeRunJobsRecommended(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of jobs to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func makeRunJobsRecommended(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

// printJobSummaries outputs job summaries as JSON or text.
func printJobSummaries(cmd *cobra.Command, jobs []JobSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(jobs)
	}
	if len(jobs) == 0 {
		fmt.Println("No jobs found.")
		return nil
	}
	lines := make([]string, 0, len(jobs)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-40s  %-20s  %-20s  %-10s", "ID", "TITLE", "COMPANY", "LOCATION", "POSTED"))
	for _, j := range jobs {
		lines = append(lines, fmt.Sprintf("%-20s  %-40s  %-20s  %-20s  %-10s",
			truncate(j.ID, 20),
			truncate(j.Title, 40),
			truncate(j.Company, 20),
			truncate(j.Location, 20),
			formatTimestamp(j.PostedAt),
		))
	}
	cli.PrintText(lines)
	return nil
}

// printJobDetail outputs a single job as JSON or formatted text block.
func printJobDetail(cmd *cobra.Command, j JobSummary, description string) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(j)
	}
	lines := []string{
		fmt.Sprintf("ID:       %s", j.ID),
		fmt.Sprintf("Title:    %s", j.Title),
		fmt.Sprintf("Company:  %s", j.Company),
		fmt.Sprintf("Location: %s", j.Location),
		fmt.Sprintf("Posted:   %s", formatTimestamp(j.PostedAt)),
		fmt.Sprintf("Remote:   %s", j.Remote),
	}
	if description != "" {
		lines = append(lines, fmt.Sprintf("Description:\n%s", truncate(description, 500)))
	}
	cli.PrintText(lines)
	return nil
}

// newJobsSaveCmd builds the "jobs save" command.
func newJobsSaveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save a job posting",
		RunE:  makeRunJobsSave(factory),
	}
	cmd.Flags().String("job-id", "", "Job posting ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview action without executing it")
	_ = cmd.MarkFlagRequired("job-id")
	return cmd
}

func makeRunJobsSave(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		jobID, _ := cmd.Flags().GetString("job-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("save job %s", jobID), map[string]string{"job_id": jobID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"jobPostingUrn": fmt.Sprintf("urn:li:fs_normalized_jobPosting:%s", jobID),
		}
		_, err = client.PostJSON(ctx, "/voyager/api/jobs/savedJobs", body)
		if err != nil {
			return fmt.Errorf("saving job %s: %w", jobID, err)
		}

		fmt.Printf("Saved job %s\n", jobID)
		return nil
	}
}

// newJobsUnsaveCmd builds the "jobs unsave" command.
func newJobsUnsaveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unsave",
		Short: "Remove a saved job posting",
		RunE:  makeRunJobsUnsave(factory),
	}
	cmd.Flags().String("job-id", "", "Saved job ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview action without executing it")
	_ = cmd.MarkFlagRequired("job-id")
	return cmd
}

func makeRunJobsUnsave(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		jobID, _ := cmd.Flags().GetString("job-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("unsave job %s", jobID), map[string]string{"job_id": jobID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/jobs/savedJobs/" + url.PathEscape(jobID)
		_, err = client.Delete(ctx, path)
		if err != nil {
			return fmt.Errorf("unsaving job %s: %w", jobID, err)
		}

		fmt.Printf("Removed saved job %s\n", jobID)
		return nil
	}
}
