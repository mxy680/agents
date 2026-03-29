package gcp

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// --- Project types ---

// ProjectSummary is the JSON-serializable summary of a GCP project.
type ProjectSummary struct {
	ProjectID   string `json:"projectId"`
	DisplayName string `json:"displayName,omitempty"`
	State       string `json:"state,omitempty"`
	CreateTime  string `json:"createTime,omitempty"`
}

// ProjectDetail is the JSON-serializable full metadata of a GCP project.
type ProjectDetail struct {
	Name        string `json:"name"`
	ProjectID   string `json:"projectId"`
	DisplayName string `json:"displayName,omitempty"`
	State       string `json:"state,omitempty"`
	Parent      string `json:"parent,omitempty"`
	CreateTime  string `json:"createTime,omitempty"`
	UpdateTime  string `json:"updateTime,omitempty"`
	Etag        string `json:"etag,omitempty"`
}

// --- Service types ---

// ServiceSummary is the JSON-serializable summary of a GCP API service.
type ServiceSummary struct {
	Name  string `json:"name"`
	Title string `json:"title,omitempty"`
	State string `json:"state,omitempty"`
}

// --- OAuth client types ---

// OAuthClientSummary is the JSON-serializable summary of an IAM OAuth client.
type OAuthClientSummary struct {
	Name        string   `json:"name"`
	ClientID    string   `json:"clientId,omitempty"`
	DisplayName string   `json:"displayName,omitempty"`
	RedirectURIs []string `json:"allowedRedirectUris,omitempty"`
	Disabled    bool     `json:"disabled,omitempty"`
}

// OAuthCredential is the JSON-serializable representation of an OAuth client credential.
// This contains the client_id and client_secret needed for OAuth flows.
type OAuthCredential struct {
	Name         string `json:"name"`
	ClientID     string `json:"clientId,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
	Disabled     bool   `json:"disabled,omitempty"`
}

// --- Brand (consent screen) types ---

// BrandSummary is the JSON-serializable summary of an IAP brand.
type BrandSummary struct {
	Name             string `json:"name"`
	ApplicationTitle string `json:"applicationTitle,omitempty"`
	SupportEmail     string `json:"supportEmail,omitempty"`
	OrgInternalOnly  bool   `json:"orgInternalOnly,omitempty"`
}

// --- IAM service account types ---

// ServiceAccountSummary is the JSON-serializable summary of a GCP service account.
type ServiceAccountSummary struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
}

// ServiceAccountKey is the JSON-serializable representation of a service account key.
type ServiceAccountKey struct {
	Name           string `json:"name"`
	PrivateKeyData string `json:"privateKeyData,omitempty"` // base64-encoded JSON key
	KeyType        string `json:"keyType,omitempty"`
	ValidAfterTime string `json:"validAfterTime,omitempty"`
}

// --- Conversion helpers ---

func toProjectSummary(data map[string]any) ProjectSummary {
	return ProjectSummary{
		ProjectID:   jsonString(data["projectId"]),
		DisplayName: jsonString(data["displayName"]),
		State:       jsonString(data["state"]),
		CreateTime:  jsonString(data["createTime"]),
	}
}

func toProjectDetail(data map[string]any) ProjectDetail {
	return ProjectDetail{
		Name:        jsonString(data["name"]),
		ProjectID:   jsonString(data["projectId"]),
		DisplayName: jsonString(data["displayName"]),
		State:       jsonString(data["state"]),
		Parent:      jsonString(data["parent"]),
		CreateTime:  jsonString(data["createTime"]),
		UpdateTime:  jsonString(data["updateTime"]),
		Etag:        jsonString(data["etag"]),
	}
}

func toServiceSummary(data map[string]any) ServiceSummary {
	config, _ := data["config"].(map[string]any)
	title := ""
	if config != nil {
		title = jsonString(config["title"])
	}
	// Service name is like "projects/123/services/iap.googleapis.com"
	// Extract just the service name from the end.
	name := jsonString(data["name"])
	return ServiceSummary{
		Name:  name,
		Title: title,
		State: jsonString(data["state"]),
	}
}

func toOAuthClientSummary(data map[string]any) OAuthClientSummary {
	return OAuthClientSummary{
		Name:         jsonString(data["name"]),
		ClientID:     jsonString(data["clientId"]),
		DisplayName:  jsonString(data["displayName"]),
		RedirectURIs: jsonStringSlice(data["allowedRedirectUris"]),
		Disabled:     jsonBool(data["disabled"]),
	}
}

func toOAuthCredential(data map[string]any) OAuthCredential {
	return OAuthCredential{
		Name:         jsonString(data["name"]),
		ClientID:     jsonString(data["clientId"]),
		ClientSecret: jsonString(data["clientSecret"]),
		Disabled:     jsonBool(data["disabled"]),
	}
}

func toBrandSummary(data map[string]any) BrandSummary {
	return BrandSummary{
		Name:             jsonString(data["name"]),
		ApplicationTitle: jsonString(data["applicationTitle"]),
		SupportEmail:     jsonString(data["supportEmail"]),
		OrgInternalOnly:  jsonBool(data["orgInternalOnly"]),
	}
}

func toServiceAccountSummary(data map[string]any) ServiceAccountSummary {
	return ServiceAccountSummary{
		Name:        jsonString(data["name"]),
		Email:       jsonString(data["email"]),
		DisplayName: jsonString(data["displayName"]),
		Disabled:    jsonBool(data["disabled"]),
	}
}

func toServiceAccountKey(data map[string]any) ServiceAccountKey {
	return ServiceAccountKey{
		Name:           jsonString(data["name"]),
		PrivateKeyData: jsonString(data["privateKeyData"]),
		KeyType:        jsonString(data["keyType"]),
		ValidAfterTime: jsonString(data["validAfterTime"]),
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
	lines = append(lines, fmt.Sprintf("%-30s  %-30s  %s", "PROJECT ID", "DISPLAY NAME", "STATE"))
	for _, p := range projects {
		lines = append(lines, fmt.Sprintf("%-30s  %-30s  %s",
			truncate(p.ProjectID, 30), truncate(p.DisplayName, 30), p.State))
	}
	cli.PrintText(lines)
	return nil
}

// printServiceSummaries outputs service summaries as JSON or a formatted text table.
func printServiceSummaries(cmd *cobra.Command, services []ServiceSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(services)
	}
	if len(services) == 0 {
		fmt.Println("No services found.")
		return nil
	}
	lines := make([]string, 0, len(services)+1)
	lines = append(lines, fmt.Sprintf("%-50s  %-30s  %s", "NAME", "TITLE", "STATE"))
	for _, s := range services {
		lines = append(lines, fmt.Sprintf("%-50s  %-30s  %s",
			truncate(s.Name, 50), truncate(s.Title, 30), s.State))
	}
	cli.PrintText(lines)
	return nil
}

// printOAuthClientSummaries outputs OAuth client summaries as JSON or a formatted text table.
func printOAuthClientSummaries(cmd *cobra.Command, clients []OAuthClientSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(clients)
	}
	if len(clients) == 0 {
		fmt.Println("No OAuth clients found.")
		return nil
	}
	lines := make([]string, 0, len(clients)+1)
	lines = append(lines, fmt.Sprintf("%-40s  %-30s  %-10s", "NAME", "DISPLAY NAME", "DISABLED"))
	for _, c := range clients {
		lines = append(lines, fmt.Sprintf("%-40s  %-30s  %-10v",
			truncate(c.Name, 40), truncate(c.DisplayName, 30), c.Disabled))
	}
	cli.PrintText(lines)
	return nil
}

// printOAuthCredentials outputs OAuth credentials as JSON or a formatted text table.
func printOAuthCredentials(cmd *cobra.Command, creds []OAuthCredential) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(creds)
	}
	if len(creds) == 0 {
		fmt.Println("No credentials found.")
		return nil
	}
	lines := make([]string, 0, len(creds)+1)
	lines = append(lines, fmt.Sprintf("%-50s  %-10s", "NAME", "DISABLED"))
	for _, c := range creds {
		lines = append(lines, fmt.Sprintf("%-50s  %-10v", truncate(c.Name, 50), c.Disabled))
	}
	cli.PrintText(lines)
	return nil
}

// printServiceAccountSummaries outputs service account summaries as JSON or a formatted text table.
func printServiceAccountSummaries(cmd *cobra.Command, accounts []ServiceAccountSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(accounts)
	}
	if len(accounts) == 0 {
		fmt.Println("No service accounts found.")
		return nil
	}
	lines := make([]string, 0, len(accounts)+1)
	lines = append(lines, fmt.Sprintf("%-50s  %-30s", "EMAIL", "DISPLAY NAME"))
	for _, a := range accounts {
		lines = append(lines, fmt.Sprintf("%-50s  %-30s",
			truncate(a.Email, 50), truncate(a.DisplayName, 30)))
	}
	cli.PrintText(lines)
	return nil
}
