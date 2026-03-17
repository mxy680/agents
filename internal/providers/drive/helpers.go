package drive

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/drive/v3"
)

// FileSummary is the JSON-serializable summary of a Drive file.
type FileSummary struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	MimeType     string   `json:"mimeType"`
	Size         int64    `json:"size"`
	ModifiedTime string   `json:"modifiedTime,omitempty"`
	Parents      []string `json:"parents,omitempty"`
	Trashed      bool     `json:"trashed,omitempty"`
}

// FileDetail is the JSON-serializable full file metadata.
type FileDetail struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	MimeType       string      `json:"mimeType"`
	Size           int64       `json:"size"`
	ModifiedTime   string      `json:"modifiedTime,omitempty"`
	CreatedTime    string      `json:"createdTime,omitempty"`
	Parents        []string    `json:"parents,omitempty"`
	Trashed        bool        `json:"trashed,omitempty"`
	Description    string      `json:"description,omitempty"`
	WebViewLink    string      `json:"webViewLink,omitempty"`
	WebContentLink string      `json:"webContentLink,omitempty"`
	Owners         []OwnerInfo `json:"owners,omitempty"`
	Shared         bool        `json:"shared,omitempty"`
}

// OwnerInfo is a JSON-serializable file owner entry.
type OwnerInfo struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName,omitempty"`
}

// PermissionInfo is the JSON-serializable representation of a Drive permission.
type PermissionInfo struct {
	ID           string `json:"id"`
	Role         string `json:"role"`
	Type         string `json:"type"`
	EmailAddress string `json:"emailAddress,omitempty"`
	Domain       string `json:"domain,omitempty"`
	DisplayName  string `json:"displayName,omitempty"`
}

// toFileSummary converts a Drive API file to a FileSummary.
func toFileSummary(file *api.File) FileSummary {
	return FileSummary{
		ID:           file.Id,
		Name:         file.Name,
		MimeType:     file.MimeType,
		Size:         file.Size,
		ModifiedTime: file.ModifiedTime,
		Parents:      file.Parents,
		Trashed:      file.Trashed,
	}
}

// toFileDetail converts a Drive API file to a FileDetail.
func toFileDetail(file *api.File) FileDetail {
	d := FileDetail{
		ID:             file.Id,
		Name:           file.Name,
		MimeType:       file.MimeType,
		Size:           file.Size,
		ModifiedTime:   file.ModifiedTime,
		CreatedTime:    file.CreatedTime,
		Parents:        file.Parents,
		Trashed:        file.Trashed,
		Description:    file.Description,
		WebViewLink:    file.WebViewLink,
		WebContentLink: file.WebContentLink,
		Shared:         file.Shared,
	}
	for _, o := range file.Owners {
		d.Owners = append(d.Owners, OwnerInfo{
			Email:       o.EmailAddress,
			DisplayName: o.DisplayName,
		})
	}
	return d
}

// toPermissionInfo converts a Drive API permission to a PermissionInfo.
func toPermissionInfo(perm *api.Permission) PermissionInfo {
	return PermissionInfo{
		ID:           perm.Id,
		Role:         perm.Role,
		Type:         perm.Type,
		EmailAddress: perm.EmailAddress,
		Domain:       perm.Domain,
		DisplayName:  perm.DisplayName,
	}
}

// printFileSummaries outputs file summaries as JSON or a formatted text table.
func printFileSummaries(cmd *cobra.Command, files []FileSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(files)
	}

	if len(files) == 0 {
		fmt.Println("No files found.")
		return nil
	}

	lines := make([]string, 0, len(files)+1)
	lines = append(lines, fmt.Sprintf("%-40s  %-30s  %-10s  %s", "NAME", "TYPE", "SIZE", "MODIFIED"))
	for _, f := range files {
		name := truncate(f.Name, 40)
		lines = append(lines, fmt.Sprintf("%-40s  %-30s  %-10s  %s", name, truncate(f.MimeType, 30), formatSize(f.Size), f.ModifiedTime))
	}
	cli.PrintText(lines)
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

// formatSize converts bytes to a human-readable string.
func formatSize(bytes int64) string {
	if bytes == 0 {
		return "-"
	}
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
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
