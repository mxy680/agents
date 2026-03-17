package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// --- API client helper ---

// GitHubError represents an error response from the GitHub API.
type GitHubError struct {
	StatusCode int
	Message    string `json:"message"`
	DocURL     string `json:"documentation_url,omitempty"`
}

func (e *GitHubError) Error() string {
	return fmt.Sprintf("GitHub API error (HTTP %d): %s", e.StatusCode, e.Message)
}

// doGitHub performs an HTTP request against the GitHub API.
// method is the HTTP method, path is the API path (e.g. "/repos/owner/repo").
// body is optional and will be JSON-encoded if non-nil.
// result is optional and will be JSON-decoded from the response if non-nil.
// Returns the *http.Response for access to headers (e.g. Link for pagination).
func doGitHub(client *http.Client, method, path string, body any, result any) (*http.Response, error) {
	baseURL := auth.GitHubBaseURL()
	url := baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encoding request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	// Check rate limit headers
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		if n, err := strconv.Atoi(remaining); err == nil && n < 10 {
			resetTime := resp.Header.Get("X-RateLimit-Reset")
			fmt.Fprintf(os.Stderr, "WARNING: GitHub API rate limit low (%d remaining, resets at %s)\n", n, resetTime)
		}
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		ghErr := &GitHubError{StatusCode: resp.StatusCode}
		data, _ := io.ReadAll(resp.Body)
		json.Unmarshal(data, ghErr)
		if ghErr.Message == "" {
			ghErr.Message = string(data)
		}
		return resp, ghErr
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return resp, fmt.Errorf("decoding response: %w", err)
		}
	}

	return resp, nil
}

// parseLinkHeader extracts rel links from a GitHub Link header.
// Returns a map of rel -> URL, e.g. {"next": "https://api.github.com/...?page=2"}.
func parseLinkHeader(header string) map[string]string {
	links := make(map[string]string)
	if header == "" {
		return links
	}
	for _, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		sections := strings.SplitN(part, ";", 2)
		if len(sections) < 2 {
			continue
		}
		urlPart := strings.TrimSpace(sections[0])
		relPart := strings.TrimSpace(sections[1])

		// Extract URL from <...>
		if !strings.HasPrefix(urlPart, "<") || !strings.HasSuffix(urlPart, ">") {
			continue
		}
		url := urlPart[1 : len(urlPart)-1]

		// Extract rel from rel="..."
		if strings.HasPrefix(relPart, `rel="`) && strings.HasSuffix(relPart, `"`) {
			rel := relPart[5 : len(relPart)-1]
			links[rel] = url
		}
	}
	return links
}

// hasNextPage checks if a response has a next page based on the Link header.
func hasNextPage(resp *http.Response) bool {
	if resp == nil {
		return false
	}
	links := parseLinkHeader(resp.Header.Get("Link"))
	_, ok := links["next"]
	return ok
}

// --- Shared types ---

// RepoSummary is the JSON-serializable summary of a GitHub repository.
type RepoSummary struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"fullName"`
	Owner       string `json:"owner"`
	Private     bool   `json:"private"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
	UpdatedAt   string `json:"updatedAt,omitempty"`
}

// RepoDetail is the JSON-serializable full repository metadata.
type RepoDetail struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"fullName"`
	Owner         string `json:"owner"`
	Private       bool   `json:"private"`
	Description   string `json:"description,omitempty"`
	URL           string `json:"url"`
	CloneURL      string `json:"cloneUrl,omitempty"`
	DefaultBranch string `json:"defaultBranch,omitempty"`
	Language      string `json:"language,omitempty"`
	Stars         int    `json:"stars"`
	Forks         int    `json:"forks"`
	OpenIssues    int    `json:"openIssues"`
	CreatedAt     string `json:"createdAt,omitempty"`
	UpdatedAt     string `json:"updatedAt,omitempty"`
}

// IssueSummary is the JSON-serializable summary of a GitHub issue.
type IssueSummary struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	State     string   `json:"state"`
	User      string   `json:"user"`
	Labels    []string `json:"labels,omitempty"`
	CreatedAt string   `json:"createdAt,omitempty"`
	UpdatedAt string   `json:"updatedAt,omitempty"`
}

// IssueDetail is the JSON-serializable full issue metadata.
type IssueDetail struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	State     string   `json:"state"`
	User      string   `json:"user"`
	Body      string   `json:"body,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	Assignees []string `json:"assignees,omitempty"`
	URL       string   `json:"url"`
	CreatedAt string   `json:"createdAt,omitempty"`
	UpdatedAt string   `json:"updatedAt,omitempty"`
	ClosedAt  string   `json:"closedAt,omitempty"`
	Comments  int      `json:"comments"`
}

// IssueCommentInfo represents a comment on an issue.
type IssueCommentInfo struct {
	ID        int64  `json:"id"`
	User      string `json:"user"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt,omitempty"`
}

// PullSummary is the JSON-serializable summary of a pull request.
type PullSummary struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	State     string `json:"state"`
	User      string `json:"user"`
	Head      string `json:"head"`
	Base      string `json:"base"`
	Draft     bool   `json:"draft,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// PullDetail is the JSON-serializable full pull request metadata.
type PullDetail struct {
	Number    int      `json:"number"`
	Title     string   `json:"title"`
	State     string   `json:"state"`
	User      string   `json:"user"`
	Body      string   `json:"body,omitempty"`
	Head      string   `json:"head"`
	Base      string   `json:"base"`
	Draft     bool     `json:"draft,omitempty"`
	Mergeable *bool    `json:"mergeable,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	URL       string   `json:"url"`
	Additions int      `json:"additions"`
	Deletions int      `json:"deletions"`
	Commits   int      `json:"commits"`
	CreatedAt string   `json:"createdAt,omitempty"`
	UpdatedAt string   `json:"updatedAt,omitempty"`
	MergedAt  string   `json:"mergedAt,omitempty"`
}

// RunSummary is the JSON-serializable summary of a workflow run.
type RunSummary struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion,omitempty"`
	Branch     string `json:"branch,omitempty"`
	Event      string `json:"event,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

// RunDetail is the JSON-serializable full workflow run metadata.
type RunDetail struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion,omitempty"`
	Branch       string `json:"branch,omitempty"`
	Event        string `json:"event,omitempty"`
	WorkflowID   int64  `json:"workflowId"`
	RunNumber    int    `json:"runNumber"`
	RunAttempt   int    `json:"runAttempt"`
	URL          string `json:"url"`
	CreatedAt    string `json:"createdAt,omitempty"`
	UpdatedAt    string `json:"updatedAt,omitempty"`
	RunStartedAt string `json:"runStartedAt,omitempty"`
}

// WorkflowSummary is the JSON-serializable summary of a workflow.
type WorkflowSummary struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Path  string `json:"path"`
	State string `json:"state"`
}

// ReleaseSummary is the JSON-serializable summary of a release.
type ReleaseSummary struct {
	ID         int64  `json:"id"`
	TagName    string `json:"tagName"`
	Name       string `json:"name,omitempty"`
	Draft      bool   `json:"draft,omitempty"`
	Prerelease bool   `json:"prerelease,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

// ReleaseDetail is the JSON-serializable full release metadata.
type ReleaseDetail struct {
	ID          int64  `json:"id"`
	TagName     string `json:"tagName"`
	Name        string `json:"name,omitempty"`
	Body        string `json:"body,omitempty"`
	Draft       bool   `json:"draft,omitempty"`
	Prerelease  bool   `json:"prerelease,omitempty"`
	Target      string `json:"target,omitempty"`
	URL         string `json:"url,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`
	PublishedAt string `json:"publishedAt,omitempty"`
}

// GistSummary is the JSON-serializable summary of a gist.
type GistSummary struct {
	ID          string   `json:"id"`
	Description string   `json:"description,omitempty"`
	Public      bool     `json:"public"`
	Files       []string `json:"files"`
	CreatedAt   string   `json:"createdAt,omitempty"`
	UpdatedAt   string   `json:"updatedAt,omitempty"`
}

// GistDetail is the JSON-serializable full gist metadata.
type GistDetail struct {
	ID          string              `json:"id"`
	Description string              `json:"description,omitempty"`
	Public      bool                `json:"public"`
	Files       map[string]GistFile `json:"files"`
	URL         string              `json:"url,omitempty"`
	CreatedAt   string              `json:"createdAt,omitempty"`
	UpdatedAt   string              `json:"updatedAt,omitempty"`
}

// GistFile represents a file within a gist.
type GistFile struct {
	Filename string `json:"filename"`
	Language string `json:"language,omitempty"`
	Size     int    `json:"size"`
	Content  string `json:"content,omitempty"`
}

// SearchResult is the JSON-serializable wrapper for search results.
type SearchResult struct {
	TotalCount int `json:"totalCount"`
	Items      any `json:"items"`
}

// RefInfo is the JSON-serializable representation of a Git reference.
type RefInfo struct {
	Ref    string `json:"ref"`
	NodeID string `json:"nodeId,omitempty"`
	URL    string `json:"url,omitempty"`
	Object struct {
		Type string `json:"type"`
		SHA  string `json:"sha"`
	} `json:"object"`
}

// GitCommitInfo is the JSON-serializable representation of a Git commit object.
type GitCommitInfo struct {
	SHA     string `json:"sha"`
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
	Author  struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Date  string `json:"date"`
	} `json:"author"`
	Tree struct {
		SHA string `json:"sha"`
	} `json:"tree"`
	Parents []struct {
		SHA string `json:"sha"`
	} `json:"parents,omitempty"`
}

// GitTreeInfo is the JSON-serializable representation of a Git tree.
type GitTreeInfo struct {
	SHA       string         `json:"sha"`
	URL       string         `json:"url,omitempty"`
	Truncated bool           `json:"truncated,omitempty"`
	Tree      []TreeEntryInfo `json:"tree"`
}

// TreeEntryInfo is a single entry in a Git tree.
type TreeEntryInfo struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	Size int    `json:"size,omitempty"`
	SHA  string `json:"sha"`
}

// GitBlobInfo is the JSON-serializable representation of a Git blob.
type GitBlobInfo struct {
	SHA      string `json:"sha"`
	Size     int    `json:"size"`
	URL      string `json:"url,omitempty"`
	Content  string `json:"content,omitempty"`
	Encoding string `json:"encoding,omitempty"`
}

// GitTagInfo is the JSON-serializable representation of a Git tag object.
type GitTagInfo struct {
	Tag     string `json:"tag"`
	SHA     string `json:"sha"`
	Message string `json:"message"`
	URL     string `json:"url,omitempty"`
	Tagger  struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Date  string `json:"date"`
	} `json:"tagger"`
	Object struct {
		Type string `json:"type"`
		SHA  string `json:"sha"`
	} `json:"object"`
}

// OrgSummary is the JSON-serializable summary of an organization.
type OrgSummary struct {
	Login       string `json:"login"`
	ID          int64  `json:"id"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
}

// OrgDetail is the JSON-serializable full organization metadata.
type OrgDetail struct {
	Login       string `json:"login"`
	ID          int64  `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	Blog        string `json:"blog,omitempty"`
	Location    string `json:"location,omitempty"`
	Email       string `json:"email,omitempty"`
	PublicRepos int    `json:"publicRepos"`
	CreatedAt   string `json:"createdAt,omitempty"`
}

// MemberInfo is the JSON-serializable representation of an org member.
type MemberInfo struct {
	Login     string `json:"login"`
	ID        int64  `json:"id"`
	AvatarURL string `json:"avatarUrl,omitempty"`
	Role      string `json:"role,omitempty"`
}

// TeamSummary is the JSON-serializable summary of a team.
type TeamSummary struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	Permission  string `json:"permission,omitempty"`
}

// LabelInfo is the JSON-serializable representation of a label.
type LabelInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description,omitempty"`
}

// BranchSummary is the JSON-serializable summary of a branch.
type BranchSummary struct {
	Name      string `json:"name"`
	SHA       string `json:"sha"`
	Protected bool   `json:"protected"`
}

// BranchProtectionInfo is the JSON-serializable representation of branch protection.
type BranchProtectionInfo struct {
	URL                     string `json:"url,omitempty"`
	EnforceAdmins           bool   `json:"enforceAdmins"`
	RequiredPullReviewCount int    `json:"requiredPullReviewCount,omitempty"`
	RequireCodeOwnerReviews bool   `json:"requireCodeOwnerReviews,omitempty"`
	RequiredStatusChecks    []string `json:"requiredStatusChecks,omitempty"`
}

// --- Conversion helpers ---

// toRepoSummary converts a GitHub API repo response to a RepoSummary.
func toRepoSummary(data map[string]any) RepoSummary {
	owner := ""
	if o, ok := data["owner"].(map[string]any); ok {
		owner, _ = o["login"].(string)
	}
	return RepoSummary{
		ID:          jsonInt64(data["id"]),
		Name:        jsonString(data["name"]),
		FullName:    jsonString(data["full_name"]),
		Owner:       owner,
		Private:     jsonBool(data["private"]),
		Description: jsonString(data["description"]),
		URL:         jsonString(data["html_url"]),
		UpdatedAt:   jsonString(data["updated_at"]),
	}
}

// toRepoDetail converts a GitHub API repo response to a RepoDetail.
func toRepoDetail(data map[string]any) RepoDetail {
	owner := ""
	if o, ok := data["owner"].(map[string]any); ok {
		owner, _ = o["login"].(string)
	}
	return RepoDetail{
		ID:            jsonInt64(data["id"]),
		Name:          jsonString(data["name"]),
		FullName:      jsonString(data["full_name"]),
		Owner:         owner,
		Private:       jsonBool(data["private"]),
		Description:   jsonString(data["description"]),
		URL:           jsonString(data["html_url"]),
		CloneURL:      jsonString(data["clone_url"]),
		DefaultBranch: jsonString(data["default_branch"]),
		Language:      jsonString(data["language"]),
		Stars:         jsonInt(data["stargazers_count"]),
		Forks:         jsonInt(data["forks_count"]),
		OpenIssues:    jsonInt(data["open_issues_count"]),
		CreatedAt:     jsonString(data["created_at"]),
		UpdatedAt:     jsonString(data["updated_at"]),
	}
}

// toIssueSummary converts a GitHub API issue response to an IssueSummary.
func toIssueSummary(data map[string]any) IssueSummary {
	return IssueSummary{
		Number:    jsonInt(data["number"]),
		Title:     jsonString(data["title"]),
		State:     jsonString(data["state"]),
		User:      jsonNestedString(data["user"], "login"),
		Labels:    jsonStringSliceFromLabels(data["labels"]),
		CreatedAt: jsonString(data["created_at"]),
		UpdatedAt: jsonString(data["updated_at"]),
	}
}

// toIssueDetail converts a GitHub API issue response to an IssueDetail.
func toIssueDetail(data map[string]any) IssueDetail {
	return IssueDetail{
		Number:    jsonInt(data["number"]),
		Title:     jsonString(data["title"]),
		State:     jsonString(data["state"]),
		User:      jsonNestedString(data["user"], "login"),
		Body:      jsonString(data["body"]),
		Labels:    jsonStringSliceFromLabels(data["labels"]),
		Assignees: jsonStringSliceFromUsers(data["assignees"]),
		URL:       jsonString(data["html_url"]),
		CreatedAt: jsonString(data["created_at"]),
		UpdatedAt: jsonString(data["updated_at"]),
		ClosedAt:  jsonString(data["closed_at"]),
		Comments:  jsonInt(data["comments"]),
	}
}

// toPullSummary converts a GitHub API PR response to a PullSummary.
func toPullSummary(data map[string]any) PullSummary {
	return PullSummary{
		Number:    jsonInt(data["number"]),
		Title:     jsonString(data["title"]),
		State:     jsonString(data["state"]),
		User:      jsonNestedString(data["user"], "login"),
		Head:      jsonNestedString(data["head"], "ref"),
		Base:      jsonNestedString(data["base"], "ref"),
		Draft:     jsonBool(data["draft"]),
		CreatedAt: jsonString(data["created_at"]),
		UpdatedAt: jsonString(data["updated_at"]),
	}
}

// toPullDetail converts a GitHub API PR response to a PullDetail.
func toPullDetail(data map[string]any) PullDetail {
	d := PullDetail{
		Number:    jsonInt(data["number"]),
		Title:     jsonString(data["title"]),
		State:     jsonString(data["state"]),
		User:      jsonNestedString(data["user"], "login"),
		Body:      jsonString(data["body"]),
		Head:      jsonNestedString(data["head"], "ref"),
		Base:      jsonNestedString(data["base"], "ref"),
		Draft:     jsonBool(data["draft"]),
		Labels:    jsonStringSliceFromLabels(data["labels"]),
		URL:       jsonString(data["html_url"]),
		Additions: jsonInt(data["additions"]),
		Deletions: jsonInt(data["deletions"]),
		Commits:   jsonInt(data["commits"]),
		CreatedAt: jsonString(data["created_at"]),
		UpdatedAt: jsonString(data["updated_at"]),
		MergedAt:  jsonString(data["merged_at"]),
	}
	if m, ok := data["mergeable"]; ok && m != nil {
		b := jsonBool(m)
		d.Mergeable = &b
	}
	return d
}

// toRunSummary converts a GitHub API workflow run to a RunSummary.
func toRunSummary(data map[string]any) RunSummary {
	return RunSummary{
		ID:         jsonInt64(data["id"]),
		Name:       jsonString(data["name"]),
		Status:     jsonString(data["status"]),
		Conclusion: jsonString(data["conclusion"]),
		Branch:     jsonString(data["head_branch"]),
		Event:      jsonString(data["event"]),
		CreatedAt:  jsonString(data["created_at"]),
	}
}

// toRunDetail converts a GitHub API workflow run to a RunDetail.
func toRunDetail(data map[string]any) RunDetail {
	return RunDetail{
		ID:           jsonInt64(data["id"]),
		Name:         jsonString(data["name"]),
		Status:       jsonString(data["status"]),
		Conclusion:   jsonString(data["conclusion"]),
		Branch:       jsonString(data["head_branch"]),
		Event:        jsonString(data["event"]),
		WorkflowID:   jsonInt64(data["workflow_id"]),
		RunNumber:    jsonInt(data["run_number"]),
		RunAttempt:   jsonInt(data["run_attempt"]),
		URL:          jsonString(data["html_url"]),
		CreatedAt:    jsonString(data["created_at"]),
		UpdatedAt:    jsonString(data["updated_at"]),
		RunStartedAt: jsonString(data["run_started_at"]),
	}
}

// toReleaseSummary converts a GitHub API release to a ReleaseSummary.
func toReleaseSummary(data map[string]any) ReleaseSummary {
	return ReleaseSummary{
		ID:         jsonInt64(data["id"]),
		TagName:    jsonString(data["tag_name"]),
		Name:       jsonString(data["name"]),
		Draft:      jsonBool(data["draft"]),
		Prerelease: jsonBool(data["prerelease"]),
		CreatedAt:  jsonString(data["created_at"]),
	}
}

// toReleaseDetail converts a GitHub API release to a ReleaseDetail.
func toReleaseDetail(data map[string]any) ReleaseDetail {
	return ReleaseDetail{
		ID:          jsonInt64(data["id"]),
		TagName:     jsonString(data["tag_name"]),
		Name:        jsonString(data["name"]),
		Body:        jsonString(data["body"]),
		Draft:       jsonBool(data["draft"]),
		Prerelease:  jsonBool(data["prerelease"]),
		Target:      jsonString(data["target_commitish"]),
		URL:         jsonString(data["html_url"]),
		CreatedAt:   jsonString(data["created_at"]),
		PublishedAt: jsonString(data["published_at"]),
	}
}

// --- JSON extraction helpers ---

func jsonString(v any) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func jsonBool(v any) bool {
	if v == nil {
		return false
	}
	b, _ := v.(bool)
	return b
}

func jsonInt(v any) int {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}

func jsonInt64(v any) int64 {
	if v == nil {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int64(n)
	case int64:
		return n
	case json.Number:
		i, _ := n.Int64()
		return i
	}
	return 0
}

func jsonNestedString(v any, key string) string {
	if v == nil {
		return ""
	}
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	s, _ := m[key].(string)
	return s
}

func jsonStringSliceFromLabels(v any) []string {
	if v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			if name, ok := m["name"].(string); ok {
				result = append(result, name)
			}
		}
	}
	return result
}

func jsonStringSliceFromUsers(v any) []string {
	if v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			if login, ok := m["login"].(string); ok {
				result = append(result, login)
			}
		}
	}
	return result
}

// --- Output helpers ---

// truncate shortens s to at most max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// confirmDestructive returns an error if the --confirm flag is absent or false.
func confirmDestructive(cmd *cobra.Command) error {
	confirmed, _ := cmd.Flags().GetBool("confirm")
	if !confirmed {
		return fmt.Errorf("this action is irreversible; re-run with --confirm to proceed")
	}
	return nil
}

// dryRunResult prints a standardised dry-run response and returns nil.
func dryRunResult(cmd *cobra.Command, description string, data any) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(data)
	}
	fmt.Printf("[DRY RUN] %s\n", description)
	return nil
}

// printRepoSummaries outputs repo summaries as JSON or a formatted text table.
func printRepoSummaries(cmd *cobra.Command, repos []RepoSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(repos)
	}
	if len(repos) == 0 {
		fmt.Println("No repositories found.")
		return nil
	}
	lines := make([]string, 0, len(repos)+1)
	lines = append(lines, fmt.Sprintf("%-40s  %-20s  %-7s  %s", "NAME", "OWNER", "PRIVATE", "UPDATED"))
	for _, r := range repos {
		lines = append(lines, fmt.Sprintf("%-40s  %-20s  %-7v  %s", truncate(r.FullName, 40), truncate(r.Owner, 20), r.Private, r.UpdatedAt))
	}
	cli.PrintText(lines)
	return nil
}

// printIssueSummaries outputs issue summaries as JSON or a formatted text table.
func printIssueSummaries(cmd *cobra.Command, issues []IssueSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(issues)
	}
	if len(issues) == 0 {
		fmt.Println("No issues found.")
		return nil
	}
	lines := make([]string, 0, len(issues)+1)
	lines = append(lines, fmt.Sprintf("%-6s  %-50s  %-8s  %-15s  %s", "NUM", "TITLE", "STATE", "USER", "UPDATED"))
	for _, i := range issues {
		lines = append(lines, fmt.Sprintf("%-6d  %-50s  %-8s  %-15s  %s", i.Number, truncate(i.Title, 50), i.State, truncate(i.User, 15), i.UpdatedAt))
	}
	cli.PrintText(lines)
	return nil
}

// printPullSummaries outputs PR summaries as JSON or a formatted text table.
func printPullSummaries(cmd *cobra.Command, pulls []PullSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(pulls)
	}
	if len(pulls) == 0 {
		fmt.Println("No pull requests found.")
		return nil
	}
	lines := make([]string, 0, len(pulls)+1)
	lines = append(lines, fmt.Sprintf("%-6s  %-45s  %-8s  %-15s  %-20s  %s", "NUM", "TITLE", "STATE", "USER", "HEAD", "BASE"))
	for _, p := range pulls {
		lines = append(lines, fmt.Sprintf("%-6d  %-45s  %-8s  %-15s  %-20s  %s", p.Number, truncate(p.Title, 45), p.State, truncate(p.User, 15), truncate(p.Head, 20), p.Base))
	}
	cli.PrintText(lines)
	return nil
}

// printRunSummaries outputs run summaries as JSON or a formatted text table.
func printRunSummaries(cmd *cobra.Command, runs []RunSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(runs)
	}
	if len(runs) == 0 {
		fmt.Println("No workflow runs found.")
		return nil
	}
	lines := make([]string, 0, len(runs)+1)
	lines = append(lines, fmt.Sprintf("%-12s  %-30s  %-12s  %-12s  %-20s  %s", "ID", "NAME", "STATUS", "CONCLUSION", "BRANCH", "CREATED"))
	for _, r := range runs {
		lines = append(lines, fmt.Sprintf("%-12d  %-30s  %-12s  %-12s  %-20s  %s", r.ID, truncate(r.Name, 30), r.Status, r.Conclusion, truncate(r.Branch, 20), r.CreatedAt))
	}
	cli.PrintText(lines)
	return nil
}

// printReleaseSummaries outputs release summaries as JSON or a formatted text table.
func printReleaseSummaries(cmd *cobra.Command, releases []ReleaseSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(releases)
	}
	if len(releases) == 0 {
		fmt.Println("No releases found.")
		return nil
	}
	lines := make([]string, 0, len(releases)+1)
	lines = append(lines, fmt.Sprintf("%-12s  %-20s  %-30s  %-7s  %s", "ID", "TAG", "NAME", "DRAFT", "CREATED"))
	for _, r := range releases {
		lines = append(lines, fmt.Sprintf("%-12d  %-20s  %-30s  %-7v  %s", r.ID, truncate(r.TagName, 20), truncate(r.Name, 30), r.Draft, r.CreatedAt))
	}
	cli.PrintText(lines)
	return nil
}

// printGistSummaries outputs gist summaries as JSON or a formatted text table.
func printGistSummaries(cmd *cobra.Command, gists []GistSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(gists)
	}
	if len(gists) == 0 {
		fmt.Println("No gists found.")
		return nil
	}
	lines := make([]string, 0, len(gists)+1)
	lines = append(lines, fmt.Sprintf("%-32s  %-40s  %-7s  %s", "ID", "DESCRIPTION", "PUBLIC", "UPDATED"))
	for _, g := range gists {
		lines = append(lines, fmt.Sprintf("%-32s  %-40s  %-7v  %s", g.ID, truncate(g.Description, 40), g.Public, g.UpdatedAt))
	}
	cli.PrintText(lines)
	return nil
}
