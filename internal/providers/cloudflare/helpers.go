package cloudflare

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// --- Zone types ---

// ZoneSummary is the JSON-serializable summary of a Cloudflare zone.
type ZoneSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Plan   string `json:"plan,omitempty"`
}

// ZoneDetail is the JSON-serializable full zone metadata.
type ZoneDetail struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Status              string `json:"status"`
	Plan                string `json:"plan,omitempty"`
	NameServers         []string `json:"name_servers,omitempty"`
	OriginalNameServers []string `json:"original_name_servers,omitempty"`
	Type                string `json:"type,omitempty"`
	Paused              bool   `json:"paused"`
}

// --- DNS types ---

// DNSRecordSummary is the JSON-serializable summary of a DNS record.
type DNSRecordSummary struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl,omitempty"`
}

// --- Worker types ---

// WorkerSummary is the JSON-serializable summary of a Worker script.
type WorkerSummary struct {
	ID         string `json:"id"`
	ETAG       string `json:"etag,omitempty"`
	CreatedOn  string `json:"created_on,omitempty"`
	ModifiedOn string `json:"modified_on,omitempty"`
}

// --- Pages types ---

// PagesSummary is the JSON-serializable summary of a Pages project.
type PagesSummary struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	SubDomain         string `json:"subdomain,omitempty"`
	ProductionBranch  string `json:"production_branch,omitempty"`
	CreatedOn         string `json:"created_on,omitempty"`
}

// PagesDeploymentSummary is the JSON-serializable summary of a Pages deployment.
type PagesDeploymentSummary struct {
	ID          string `json:"id"`
	URL         string `json:"url,omitempty"`
	Environment string `json:"environment,omitempty"`
	Stage       string `json:"stage,omitempty"`
	CreatedOn   string `json:"created_on,omitempty"`
}

// --- R2 types ---

// R2BucketSummary is the JSON-serializable summary of an R2 bucket.
type R2BucketSummary struct {
	Name         string `json:"name"`
	CreationDate string `json:"creation_date,omitempty"`
}

// --- KV types ---

// KVNamespaceSummary is the JSON-serializable summary of a KV namespace.
type KVNamespaceSummary struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	SupportsTTL    bool   `json:"supports_url_encoding,omitempty"`
}

// KVKeySummary is the JSON-serializable summary of a KV key.
type KVKeySummary struct {
	Name       string `json:"name"`
	Expiration int64  `json:"expiration,omitempty"`
}

// --- Firewall types ---

// FirewallRuleSummary is the JSON-serializable summary of a firewall rule.
type FirewallRuleSummary struct {
	ID          string `json:"id"`
	Action      string `json:"action"`
	Description string `json:"description,omitempty"`
	Paused      bool   `json:"paused"`
}

// --- Certificate types ---

// CertPackSummary is the JSON-serializable summary of an SSL certificate pack.
type CertPackSummary struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Hosts              []string `json:"hosts,omitempty"`
	PrimaryCertificate int    `json:"primary_certificate,omitempty"`
	Status             string `json:"status,omitempty"`
}

// --- Account types ---

// AccountSummary is the JSON-serializable summary of a Cloudflare account.
type AccountSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
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

func toZoneSummary(data map[string]any) ZoneSummary {
	return ZoneSummary{
		ID:     jsonString(data["id"]),
		Name:   jsonString(data["name"]),
		Status: jsonString(data["status"]),
		Plan:   jsonNestedString(data["plan"], "name"),
	}
}

func toZoneDetail(data map[string]any) ZoneDetail {
	return ZoneDetail{
		ID:                  jsonString(data["id"]),
		Name:                jsonString(data["name"]),
		Status:              jsonString(data["status"]),
		Plan:                jsonNestedString(data["plan"], "name"),
		NameServers:         jsonStringSlice(data["name_servers"]),
		OriginalNameServers: jsonStringSlice(data["original_name_servers"]),
		Type:                jsonString(data["type"]),
		Paused:              jsonBool(data["paused"]),
	}
}

func toDNSRecordSummary(data map[string]any) DNSRecordSummary {
	return DNSRecordSummary{
		ID:      jsonString(data["id"]),
		Type:    jsonString(data["type"]),
		Name:    jsonString(data["name"]),
		Content: jsonString(data["content"]),
		Proxied: jsonBool(data["proxied"]),
		TTL:     jsonInt(data["ttl"]),
	}
}

func toWorkerSummary(data map[string]any) WorkerSummary {
	return WorkerSummary{
		ID:         jsonString(data["id"]),
		ETAG:       jsonString(data["etag"]),
		CreatedOn:  jsonString(data["created_on"]),
		ModifiedOn: jsonString(data["modified_on"]),
	}
}

func toPagesSummary(data map[string]any) PagesSummary {
	return PagesSummary{
		ID:               jsonString(data["id"]),
		Name:             jsonString(data["name"]),
		SubDomain:        jsonString(data["subdomain"]),
		ProductionBranch: jsonString(data["production_branch"]),
		CreatedOn:        jsonString(data["created_on"]),
	}
}

func toPagesDeploymentSummary(data map[string]any) PagesDeploymentSummary {
	stage := jsonNestedString(data["latest_stage"], "name")
	if stage == "" {
		stage = jsonString(data["environment"])
	}
	return PagesDeploymentSummary{
		ID:          jsonString(data["id"]),
		URL:         jsonString(data["url"]),
		Environment: jsonString(data["environment"]),
		Stage:       stage,
		CreatedOn:   jsonString(data["created_on"]),
	}
}

func toR2BucketSummary(data map[string]any) R2BucketSummary {
	return R2BucketSummary{
		Name:         jsonString(data["name"]),
		CreationDate: jsonString(data["creation_date"]),
	}
}

func toKVNamespaceSummary(data map[string]any) KVNamespaceSummary {
	return KVNamespaceSummary{
		ID:    jsonString(data["id"]),
		Title: jsonString(data["title"]),
	}
}

func toKVKeySummary(data map[string]any) KVKeySummary {
	return KVKeySummary{
		Name:       jsonString(data["name"]),
		Expiration: jsonInt64(data["expiration"]),
	}
}

func toFirewallRuleSummary(data map[string]any) FirewallRuleSummary {
	return FirewallRuleSummary{
		ID:          jsonString(data["id"]),
		Action:      jsonString(data["action"]),
		Description: jsonString(data["description"]),
		Paused:      jsonBool(data["paused"]),
	}
}

func toCertPackSummary(data map[string]any) CertPackSummary {
	return CertPackSummary{
		ID:     jsonString(data["id"]),
		Type:   jsonString(data["type"]),
		Hosts:  jsonStringSlice(data["hosts"]),
		Status: jsonString(data["status"]),
	}
}

func toAccountSummary(data map[string]any) AccountSummary {
	return AccountSummary{
		ID:   jsonString(data["id"]),
		Name: jsonString(data["name"]),
		Type: jsonString(data["type"]),
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

// --- Print helpers ---

func printZoneSummaries(cmd *cobra.Command, zones []ZoneSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(zones)
	}
	if len(zones) == 0 {
		fmt.Println("No zones found.")
		return nil
	}
	lines := make([]string, 0, len(zones)+1)
	lines = append(lines, fmt.Sprintf("%-32s  %-40s  %-10s  %s", "ID", "NAME", "STATUS", "PLAN"))
	for _, z := range zones {
		lines = append(lines, fmt.Sprintf("%-32s  %-40s  %-10s  %s",
			truncate(z.ID, 32), truncate(z.Name, 40), z.Status, z.Plan))
	}
	cli.PrintText(lines)
	return nil
}

func printDNSRecords(cmd *cobra.Command, records []DNSRecordSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(records)
	}
	if len(records) == 0 {
		fmt.Println("No DNS records found.")
		return nil
	}
	lines := make([]string, 0, len(records)+1)
	lines = append(lines, fmt.Sprintf("%-32s  %-6s  %-40s  %-40s  %s", "ID", "TYPE", "NAME", "CONTENT", "PROXIED"))
	for _, r := range records {
		lines = append(lines, fmt.Sprintf("%-32s  %-6s  %-40s  %-40s  %v",
			truncate(r.ID, 32), r.Type, truncate(r.Name, 40), truncate(r.Content, 40), r.Proxied))
	}
	cli.PrintText(lines)
	return nil
}

func printWorkers(cmd *cobra.Command, workers []WorkerSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(workers)
	}
	if len(workers) == 0 {
		fmt.Println("No workers found.")
		return nil
	}
	lines := make([]string, 0, len(workers)+1)
	lines = append(lines, fmt.Sprintf("%-40s  %s", "ID", "MODIFIED"))
	for _, w := range workers {
		lines = append(lines, fmt.Sprintf("%-40s  %s", truncate(w.ID, 40), w.ModifiedOn))
	}
	cli.PrintText(lines)
	return nil
}

func printPagesProjects(cmd *cobra.Command, projects []PagesSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(projects)
	}
	if len(projects) == 0 {
		fmt.Println("No Pages projects found.")
		return nil
	}
	lines := make([]string, 0, len(projects)+1)
	lines = append(lines, fmt.Sprintf("%-36s  %-30s  %-30s  %s", "ID", "NAME", "SUBDOMAIN", "BRANCH"))
	for _, p := range projects {
		lines = append(lines, fmt.Sprintf("%-36s  %-30s  %-30s  %s",
			truncate(p.ID, 36), truncate(p.Name, 30), truncate(p.SubDomain, 30), p.ProductionBranch))
	}
	cli.PrintText(lines)
	return nil
}

func printPagesDeployments(cmd *cobra.Command, deployments []PagesDeploymentSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(deployments)
	}
	if len(deployments) == 0 {
		fmt.Println("No deployments found.")
		return nil
	}
	lines := make([]string, 0, len(deployments)+1)
	lines = append(lines, fmt.Sprintf("%-36s  %-12s  %-30s  %s", "ID", "ENVIRONMENT", "URL", "CREATED"))
	for _, d := range deployments {
		lines = append(lines, fmt.Sprintf("%-36s  %-12s  %-30s  %s",
			truncate(d.ID, 36), d.Environment, truncate(d.URL, 30), d.CreatedOn))
	}
	cli.PrintText(lines)
	return nil
}

func printR2Buckets(cmd *cobra.Command, buckets []R2BucketSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(buckets)
	}
	if len(buckets) == 0 {
		fmt.Println("No R2 buckets found.")
		return nil
	}
	lines := make([]string, 0, len(buckets)+1)
	lines = append(lines, fmt.Sprintf("%-50s  %s", "NAME", "CREATED"))
	for _, b := range buckets {
		lines = append(lines, fmt.Sprintf("%-50s  %s", truncate(b.Name, 50), b.CreationDate))
	}
	cli.PrintText(lines)
	return nil
}

func printKVNamespaces(cmd *cobra.Command, namespaces []KVNamespaceSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(namespaces)
	}
	if len(namespaces) == 0 {
		fmt.Println("No KV namespaces found.")
		return nil
	}
	lines := make([]string, 0, len(namespaces)+1)
	lines = append(lines, fmt.Sprintf("%-36s  %s", "ID", "TITLE"))
	for _, n := range namespaces {
		lines = append(lines, fmt.Sprintf("%-36s  %s", truncate(n.ID, 36), n.Title))
	}
	cli.PrintText(lines)
	return nil
}

func printKVKeys(cmd *cobra.Command, keys []KVKeySummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(keys)
	}
	if len(keys) == 0 {
		fmt.Println("No keys found.")
		return nil
	}
	lines := make([]string, 0, len(keys)+1)
	lines = append(lines, fmt.Sprintf("%-60s  %s", "KEY", "EXPIRATION"))
	for _, k := range keys {
		exp := ""
		if k.Expiration > 0 {
			exp = fmt.Sprintf("%d", k.Expiration)
		}
		lines = append(lines, fmt.Sprintf("%-60s  %s", truncate(k.Name, 60), exp))
	}
	cli.PrintText(lines)
	return nil
}

func printFirewallRules(cmd *cobra.Command, rules []FirewallRuleSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(rules)
	}
	if len(rules) == 0 {
		fmt.Println("No firewall rules found.")
		return nil
	}
	lines := make([]string, 0, len(rules)+1)
	lines = append(lines, fmt.Sprintf("%-32s  %-10s  %-7s  %s", "ID", "ACTION", "PAUSED", "DESCRIPTION"))
	for _, r := range rules {
		lines = append(lines, fmt.Sprintf("%-32s  %-10s  %-7v  %s",
			truncate(r.ID, 32), r.Action, r.Paused, truncate(r.Description, 40)))
	}
	cli.PrintText(lines)
	return nil
}

func printCertPacks(cmd *cobra.Command, packs []CertPackSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(packs)
	}
	if len(packs) == 0 {
		fmt.Println("No certificate packs found.")
		return nil
	}
	lines := make([]string, 0, len(packs)+1)
	lines = append(lines, fmt.Sprintf("%-36s  %-10s  %s", "ID", "TYPE", "STATUS"))
	for _, p := range packs {
		lines = append(lines, fmt.Sprintf("%-36s  %-10s  %s", truncate(p.ID, 36), p.Type, p.Status))
	}
	cli.PrintText(lines)
	return nil
}

func printAccounts(cmd *cobra.Command, accounts []AccountSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(accounts)
	}
	if len(accounts) == 0 {
		fmt.Println("No accounts found.")
		return nil
	}
	lines := make([]string, 0, len(accounts)+1)
	lines = append(lines, fmt.Sprintf("%-32s  %-40s  %s", "ID", "NAME", "TYPE"))
	for _, a := range accounts {
		lines = append(lines, fmt.Sprintf("%-32s  %-40s  %s", truncate(a.ID, 32), truncate(a.Name, 40), a.Type))
	}
	cli.PrintText(lines)
	return nil
}
