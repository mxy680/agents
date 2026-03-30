package vercel

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// --- Shared types ---

// ProjectSummary is the JSON-serializable summary of a Vercel project.
type ProjectSummary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Framework string `json:"framework,omitempty"`
	NodeJS    string `json:"nodeVersion,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}

// ProjectDetail is the JSON-serializable full project metadata.
type ProjectDetail struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Framework        string `json:"framework,omitempty"`
	NodeJS           string `json:"nodeVersion,omitempty"`
	RootDirectory    string `json:"rootDirectory,omitempty"`
	BuildCommand     string `json:"buildCommand,omitempty"`
	OutputDirectory  string `json:"outputDirectory,omitempty"`
	InstallCommand   string `json:"installCommand,omitempty"`
	DevCommand       string `json:"devCommand,omitempty"`
	UpdatedAt        int64  `json:"updatedAt,omitempty"`
	CreatedAt        int64  `json:"createdAt,omitempty"`
	AccountID        string `json:"accountId,omitempty"`
}

// DeploymentSummary is the JSON-serializable summary of a Vercel deployment.
type DeploymentSummary struct {
	ID        string `json:"id"`
	URL       string `json:"url,omitempty"`
	Name      string `json:"name,omitempty"`
	State     string `json:"state,omitempty"`
	Type      string `json:"type,omitempty"`
	Target    string `json:"target,omitempty"`
	Source    string `json:"source,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
}

// DeploymentDetail is the JSON-serializable full deployment metadata.
type DeploymentDetail struct {
	ID         string `json:"id"`
	URL        string `json:"url,omitempty"`
	Name       string `json:"name,omitempty"`
	State      string `json:"state,omitempty"`
	Type       string `json:"type,omitempty"`
	Target     string `json:"target,omitempty"`
	Source     string `json:"source,omitempty"`
	Creator    string `json:"creator,omitempty"`
	GitBranch  string `json:"gitBranch,omitempty"`
	GitCommit  string `json:"gitCommit,omitempty"`
	ReadyState string `json:"readyState,omitempty"`
	CreatedAt  int64  `json:"createdAt,omitempty"`
	BuildingAt int64  `json:"buildingAt,omitempty"`
	ReadyAt    int64  `json:"readyAt,omitempty"`
}

// DomainSummary is the JSON-serializable summary of a Vercel domain.
type DomainSummary struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	Verified   bool   `json:"verified"`
	CreatedAt  int64  `json:"createdAt,omitempty"`
	ExpiresAt  int64  `json:"expiresAt,omitempty"`
}

// DomainDetail is the JSON-serializable full domain metadata.
type DomainDetail struct {
	ID             string   `json:"id,omitempty"`
	Name           string   `json:"name"`
	Verified       bool     `json:"verified"`
	Nameservers    []string `json:"nameservers,omitempty"`
	IntendedNS     []string `json:"intendedNameservers,omitempty"`
	ServiceType    string   `json:"serviceType,omitempty"`
	CreatedAt      int64    `json:"createdAt,omitempty"`
	ExpiresAt      int64    `json:"expiresAt,omitempty"`
	TransferredAt  int64    `json:"transferredAt,omitempty"`
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

func jsonStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
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

// --- Conversion helpers ---

func toProjectSummary(data map[string]any) ProjectSummary {
	return ProjectSummary{
		ID:        jsonString(data["id"]),
		Name:      jsonString(data["name"]),
		Framework: jsonString(data["framework"]),
		NodeJS:    jsonString(data["nodeVersion"]),
		UpdatedAt: jsonInt64(data["updatedAt"]),
		CreatedAt: jsonInt64(data["createdAt"]),
	}
}

func toProjectDetail(data map[string]any) ProjectDetail {
	return ProjectDetail{
		ID:              jsonString(data["id"]),
		Name:            jsonString(data["name"]),
		Framework:       jsonString(data["framework"]),
		NodeJS:          jsonString(data["nodeVersion"]),
		RootDirectory:   jsonString(data["rootDirectory"]),
		BuildCommand:    jsonString(data["buildCommand"]),
		OutputDirectory: jsonString(data["outputDirectory"]),
		InstallCommand:  jsonString(data["installCommand"]),
		DevCommand:      jsonString(data["devCommand"]),
		UpdatedAt:       jsonInt64(data["updatedAt"]),
		CreatedAt:       jsonInt64(data["createdAt"]),
		AccountID:       jsonString(data["accountId"]),
	}
}

func toDeploymentSummary(data map[string]any) DeploymentSummary {
	return DeploymentSummary{
		ID:        jsonString(data["id"]),
		URL:       jsonString(data["url"]),
		Name:      jsonString(data["name"]),
		State:     jsonString(data["state"]),
		Type:      jsonString(data["type"]),
		Target:    jsonString(data["target"]),
		Source:    jsonString(data["source"]),
		CreatedAt: jsonInt64(data["createdAt"]),
	}
}

func toDeploymentDetail(data map[string]any) DeploymentDetail {
	creator := jsonNestedString(data["creator"], "username")
	if creator == "" {
		creator = jsonNestedString(data["creator"], "email")
	}

	gitBranch := jsonNestedString(data["meta"], "githubCommitRef")
	gitCommit := jsonNestedString(data["meta"], "githubCommitSha")

	return DeploymentDetail{
		ID:         jsonString(data["id"]),
		URL:        jsonString(data["url"]),
		Name:       jsonString(data["name"]),
		State:      jsonString(data["state"]),
		Type:       jsonString(data["type"]),
		Target:     jsonString(data["target"]),
		Source:     jsonString(data["source"]),
		Creator:    creator,
		GitBranch:  gitBranch,
		GitCommit:  gitCommit,
		ReadyState: jsonString(data["readyState"]),
		CreatedAt:  jsonInt64(data["createdAt"]),
		BuildingAt: jsonInt64(data["buildingAt"]),
		ReadyAt:    jsonInt64(data["ready"]),
	}
}

func toDomainSummary(data map[string]any) DomainSummary {
	return DomainSummary{
		ID:        jsonString(data["id"]),
		Name:      jsonString(data["name"]),
		Verified:  jsonBool(data["verified"]),
		CreatedAt: jsonInt64(data["createdAt"]),
		ExpiresAt: jsonInt64(data["expiresAt"]),
	}
}

func toDomainDetail(data map[string]any) DomainDetail {
	return DomainDetail{
		ID:            jsonString(data["id"]),
		Name:          jsonString(data["name"]),
		Verified:      jsonBool(data["verified"]),
		Nameservers:   jsonStringSlice(data["nameservers"]),
		IntendedNS:    jsonStringSlice(data["intendedNameservers"]),
		ServiceType:   jsonString(data["serviceType"]),
		CreatedAt:     jsonInt64(data["createdAt"]),
		ExpiresAt:     jsonInt64(data["expiresAt"]),
		TransferredAt: jsonInt64(data["transferredAt"]),
	}
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
	lines = append(lines, fmt.Sprintf("%-28s  %-30s  %s", "ID", "NAME", "FRAMEWORK"))
	for _, p := range projects {
		lines = append(lines, fmt.Sprintf("%-28s  %-30s  %s", truncate(p.ID, 28), truncate(p.Name, 30), p.Framework))
	}
	cli.PrintText(lines)
	return nil
}

// printDeploymentSummaries outputs deployment summaries as JSON or a formatted text table.
func printDeploymentSummaries(cmd *cobra.Command, deployments []DeploymentSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(deployments)
	}
	if len(deployments) == 0 {
		fmt.Println("No deployments found.")
		return nil
	}
	lines := make([]string, 0, len(deployments)+1)
	lines = append(lines, fmt.Sprintf("%-28s  %-35s  %-10s  %-12s  %s", "ID", "URL", "STATE", "TARGET", "NAME"))
	for _, d := range deployments {
		lines = append(lines, fmt.Sprintf("%-28s  %-35s  %-10s  %-12s  %s",
			truncate(d.ID, 28), truncate(d.URL, 35), d.State, d.Target, truncate(d.Name, 20)))
	}
	cli.PrintText(lines)
	return nil
}

// printDomainSummaries outputs domain summaries as JSON or a formatted text table.
func printDomainSummaries(cmd *cobra.Command, domains []DomainSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(domains)
	}
	if len(domains) == 0 {
		fmt.Println("No domains found.")
		return nil
	}
	lines := make([]string, 0, len(domains)+1)
	lines = append(lines, fmt.Sprintf("%-40s  %-10s", "DOMAIN", "VERIFIED"))
	for _, d := range domains {
		lines = append(lines, fmt.Sprintf("%-40s  %-10v", truncate(d.Name, 40), d.Verified))
	}
	cli.PrintText(lines)
	return nil
}
