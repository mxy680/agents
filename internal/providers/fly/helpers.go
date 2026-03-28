package fly

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// --- App types ---

// AppSummary is the JSON-serializable summary of a Fly.io app.
type AppSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	OrgID  string `json:"org_slug,omitempty"`
}

// AppDetail is the JSON-serializable full app metadata.
type AppDetail struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	OrgID    string `json:"org_slug,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	AppURL   string `json:"app_url,omitempty"`
}

// --- Machine types ---

// MachineSummary is the JSON-serializable summary of a Fly.io machine.
type MachineSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	State  string `json:"state"`
	Region string `json:"region"`
	Image  string `json:"image,omitempty"`
}

// MachineDetail is the JSON-serializable full machine metadata.
type MachineDetail struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	State      string         `json:"state"`
	Region     string         `json:"region"`
	InstanceID string         `json:"instance_id,omitempty"`
	PrivateIP  string         `json:"private_ip,omitempty"`
	CreatedAt  string         `json:"created_at,omitempty"`
	UpdatedAt  string         `json:"updated_at,omitempty"`
	Config     *MachineConfig `json:"config,omitempty"`
}

// MachineConfig holds the configuration for creating or updating a machine.
type MachineConfig struct {
	Image string `json:"image"`
}

// --- Volume types ---

// VolumeSummary is the JSON-serializable summary of a Fly.io volume.
type VolumeSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	State  string `json:"state"`
	Region string `json:"region"`
	SizeGB int    `json:"size_gb"`
}

// VolumeDetail is the JSON-serializable full volume metadata.
type VolumeDetail struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	State              string `json:"state"`
	Region             string `json:"region"`
	SizeGB             int    `json:"size_gb"`
	Encrypted          bool   `json:"encrypted"`
	AttachedMachineID  string `json:"attached_machine_id,omitempty"`
	AttachedAllocID    string `json:"attached_alloc_id,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
}

// VolumeSnapshot is a point-in-time snapshot of a volume.
type VolumeSnapshot struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	SizeBytes int64  `json:"size"`
	CreatedAt string `json:"created_at,omitempty"`
	Digest    string `json:"digest,omitempty"`
}

// --- Certificate types ---

// CertSummary is the JSON-serializable summary of a Fly.io certificate.
type CertSummary struct {
	Hostname       string `json:"hostname"`
	ClientStatus   string `json:"client_status"`
	Issued         bool   `json:"issued"`
}

// CertDetail is the JSON-serializable full certificate metadata.
type CertDetail struct {
	Hostname             string `json:"hostname"`
	ClientStatus         string `json:"client_status"`
	Issued               bool   `json:"issued"`
	DNSValidationTarget  string `json:"dns_validation_target,omitempty"`
	DNSValidationHostname string `json:"dns_validation_hostname,omitempty"`
	DNSValidationInstructions string `json:"dns_validation_instructions,omitempty"`
	AcmeDNSConfigured    bool   `json:"acme_dns_configured"`
	AcmeALPNConfigured   bool   `json:"acme_alpn_configured"`
	CreatedAt            string `json:"created_at,omitempty"`
	ExpiresAt            string `json:"expires_at,omitempty"`
}

// --- Secret types ---

// SecretSummary is the JSON-serializable summary of a Fly.io secret.
type SecretSummary struct {
	Name      string `json:"name"`
	Digest    string `json:"digest,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

// --- Region types ---

// Region represents a Fly.io deployment region.
type Region struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// --- Helpers ---

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

// truncate shortens s to at most max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// --- Print helpers ---

func printAppSummaries(cmd *cobra.Command, apps []AppSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(apps)
	}
	if len(apps) == 0 {
		fmt.Println("No apps found.")
		return nil
	}
	lines := make([]string, 0, len(apps)+1)
	lines = append(lines, fmt.Sprintf("%-30s  %-12s  %s", "NAME", "STATUS", "ORG"))
	for _, a := range apps {
		lines = append(lines, fmt.Sprintf("%-30s  %-12s  %s", truncate(a.Name, 30), a.Status, a.OrgID))
	}
	cli.PrintText(lines)
	return nil
}

func printMachineSummaries(cmd *cobra.Command, machines []MachineSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(machines)
	}
	if len(machines) == 0 {
		fmt.Println("No machines found.")
		return nil
	}
	lines := make([]string, 0, len(machines)+1)
	lines = append(lines, fmt.Sprintf("%-28s  %-30s  %-10s  %-10s  %s", "ID", "NAME", "STATE", "REGION", "IMAGE"))
	for _, m := range machines {
		lines = append(lines, fmt.Sprintf("%-28s  %-30s  %-10s  %-10s  %s",
			truncate(m.ID, 28), truncate(m.Name, 30), m.State, m.Region, truncate(m.Image, 40)))
	}
	cli.PrintText(lines)
	return nil
}

func printVolumeSummaries(cmd *cobra.Command, volumes []VolumeSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(volumes)
	}
	if len(volumes) == 0 {
		fmt.Println("No volumes found.")
		return nil
	}
	lines := make([]string, 0, len(volumes)+1)
	lines = append(lines, fmt.Sprintf("%-28s  %-20s  %-10s  %-10s  %s", "ID", "NAME", "STATE", "REGION", "SIZE GB"))
	for _, v := range volumes {
		lines = append(lines, fmt.Sprintf("%-28s  %-20s  %-10s  %-10s  %d",
			truncate(v.ID, 28), truncate(v.Name, 20), v.State, v.Region, v.SizeGB))
	}
	cli.PrintText(lines)
	return nil
}

func printCertSummaries(cmd *cobra.Command, certs []CertSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(certs)
	}
	if len(certs) == 0 {
		fmt.Println("No certificates found.")
		return nil
	}
	lines := make([]string, 0, len(certs)+1)
	lines = append(lines, fmt.Sprintf("%-40s  %-15s  %s", "HOSTNAME", "CLIENT STATUS", "ISSUED"))
	for _, c := range certs {
		lines = append(lines, fmt.Sprintf("%-40s  %-15s  %v",
			truncate(c.Hostname, 40), c.ClientStatus, c.Issued))
	}
	cli.PrintText(lines)
	return nil
}

func printSecretSummaries(cmd *cobra.Command, secrets []SecretSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(secrets)
	}
	if len(secrets) == 0 {
		fmt.Println("No secrets found.")
		return nil
	}
	lines := make([]string, 0, len(secrets)+1)
	lines = append(lines, fmt.Sprintf("%-40s  %s", "NAME", "DIGEST"))
	for _, s := range secrets {
		lines = append(lines, fmt.Sprintf("%-40s  %s", truncate(s.Name, 40), truncate(s.Digest, 20)))
	}
	cli.PrintText(lines)
	return nil
}

func printRegions(cmd *cobra.Command, regions []Region) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(regions)
	}
	if len(regions) == 0 {
		fmt.Println("No regions found.")
		return nil
	}
	lines := make([]string, 0, len(regions)+1)
	lines = append(lines, fmt.Sprintf("%-8s  %s", "CODE", "NAME"))
	for _, r := range regions {
		lines = append(lines, fmt.Sprintf("%-8s  %s", r.Code, r.Name))
	}
	cli.PrintText(lines)
	return nil
}
