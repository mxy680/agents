package supabase

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// --- API client helper ---

// doSupabase makes an authenticated request to the Supabase Management API.
// It returns the response body bytes. On non-2xx responses, it returns an error
// with the status code and response body.
func doSupabase(client *http.Client, method, path string, body io.Reader) ([]byte, error) {
	baseURL := auth.SupabaseBaseURL()
	url := baseURL + "/v1" + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	return data, nil
}

// --- Shared types ---

// ProjectSummary is the JSON-serializable summary of a Supabase project.
type ProjectSummary struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	OrganizationID string `json:"organizationId"`
	Region         string `json:"region"`
	Status         string `json:"status"`
	CreatedAt      string `json:"createdAt,omitempty"`
}

// ProjectDetail is the JSON-serializable full project metadata.
type ProjectDetail struct {
	ProjectSummary
	DatabaseHost    string `json:"databaseHost,omitempty"`
	DatabaseVersion string `json:"databaseVersion,omitempty"`
}

// BranchSummary is the JSON-serializable summary of a Supabase branch.
type BranchSummary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	GitBranch string `json:"gitBranch,omitempty"`
	IsDefault bool   `json:"isDefault"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt,omitempty"`
}

// OrgSummary is the JSON-serializable summary of a Supabase organization.
type OrgSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// APIKeySummary is the JSON-serializable summary of a Supabase API key.
type APIKeySummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	APIKey string `json:"apiKey"`
	Type   string `json:"type"`
}

// SecretSummary is the JSON-serializable summary of a Supabase secret.
type SecretSummary struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// AuthConfig holds the raw Supabase Auth configuration (complex nested config).
type AuthConfig struct {
	Config json.RawMessage `json:"config"`
}

// AdvisorResult is a single finding from the Supabase advisor.
type AdvisorResult struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// --- Output helpers ---

// truncate shortens s to at most max runes, appending "…" if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

// maskKey shows the first 4 and last 4 characters, masking the middle with "****".
// For strings shorter than 8 characters, masks entirely with "****".
func maskKey(s string) string {
	if len(s) < 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}

// confirmDestructive returns an error if the --confirm flag is absent or false.
func confirmDestructive(cmd *cobra.Command, message string) error {
	confirmed, _ := cmd.Flags().GetBool("confirm")
	if !confirmed {
		return fmt.Errorf("%s; re-run with --confirm to proceed", message)
	}
	return nil
}

// dryRunResult checks the --dry-run flag. If set, prints a dry-run message to
// stdout (or JSON) and returns true. Returns false if not a dry run.
func dryRunResult(cmd *cobra.Command, action string) bool {
	if !cli.IsDryRun(cmd) {
		return false
	}
	if cli.IsJSONOutput(cmd) {
		_ = cli.PrintJSON(map[string]string{"dryRun": action})
	} else {
		fmt.Printf("[DRY RUN] %s\n", action)
	}
	return true
}

// printProjectSummaries outputs project summaries as JSON or a formatted text table.
func printProjectSummaries(cmd *cobra.Command, projects []ProjectSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(projects)
	}
	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return nil
	}
	lines := make([]string, 0, len(projects)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-30s  %-20s  %-15s  %s", "ID", "NAME", "ORG", "REGION", "STATUS"))
	for _, p := range projects {
		lines = append(lines, fmt.Sprintf("%-20s  %-30s  %-20s  %-15s  %s",
			truncate(p.ID, 20), truncate(p.Name, 30), truncate(p.OrganizationID, 20), truncate(p.Region, 15), p.Status))
	}
	cli.PrintText(lines)
	return nil
}

// printOrgSummaries outputs organization summaries as JSON or a formatted text table.
func printOrgSummaries(cmd *cobra.Command, orgs []OrgSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(orgs)
	}
	if len(orgs) == 0 {
		fmt.Println("No organizations found.")
		return nil
	}
	lines := make([]string, 0, len(orgs)+1)
	lines = append(lines, fmt.Sprintf("%-30s  %s", "ID", "NAME"))
	for _, o := range orgs {
		lines = append(lines, fmt.Sprintf("%-30s  %s", truncate(o.ID, 30), o.Name))
	}
	cli.PrintText(lines)
	return nil
}
