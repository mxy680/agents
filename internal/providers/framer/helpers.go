package framer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/emdash-projects/agents/internal/cli"
)

// Generic types for JSON results from bridge

// ProjectInfo holds basic project metadata.
type ProjectInfo struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	VersionedID string `json:"versioned_id,omitempty"`
}

// User holds basic user information.
type User struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Deployment holds information about a single deployment.
type Deployment struct {
	ID        string `json:"id"`
	URL       string `json:"url,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

// PublishResult holds the outcome of a publish operation.
type PublishResult struct {
	DeploymentID string `json:"deploymentId,omitempty"`
	URL          string `json:"url,omitempty"`
}

// PublishInfo holds information about the project's last published state.
type PublishInfo struct {
	URL           string `json:"url,omitempty"`
	LastPublished string `json:"lastPublished,omitempty"`
}

// ChangedPaths lists which project paths have been added, removed, or modified.
type ChangedPaths struct {
	Added    []string `json:"added"`
	Removed  []string `json:"removed"`
	Modified []string `json:"modified"`
}

// CollectionSummary holds the ID and name of a CMS collection.
type CollectionSummary struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// Field describes a single field in a CMS collection.
type Field struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

// CollectionItem represents a single item in a CMS collection.
type CollectionItem struct {
	ID        string         `json:"id"`
	Slug      string         `json:"slug,omitempty"`
	FieldData map[string]any `json:"fieldData,omitempty"`
}

// NodeSummary holds the essential attributes of a canvas node.
type NodeSummary struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Type    string `json:"__class,omitempty"`
	Locked  bool   `json:"locked,omitempty"`
	Visible bool   `json:"visible,omitempty"`
}

// Rect describes the bounding rectangle of a node on the canvas.
type Rect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// ColorStyle holds a color style definition.
type ColorStyle struct {
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Light string `json:"light,omitempty"`
	Dark  string `json:"dark,omitempty"`
}

// TextStyle holds a text style definition.
type TextStyle struct {
	ID       string  `json:"id"`
	Name     string  `json:"name,omitempty"`
	Font     string  `json:"font,omitempty"`
	FontSize float64 `json:"fontSize,omitempty"`
}

// Font holds font family and variant information.
type Font struct {
	Family string `json:"family"`
	Style  string `json:"style,omitempty"`
	Weight int    `json:"weight,omitempty"`
}

// Locale holds locale metadata for localization.
type Locale struct {
	ID   string `json:"id"`
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// Redirect holds a URL redirect rule.
type Redirect struct {
	ID   string `json:"id"`
	From string `json:"from"`
	To   string `json:"to"`
	Type int    `json:"type,omitempty"`
}

// CodeFile holds metadata about a code file in the project.
type CodeFile struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// CustomCode holds the project-level custom code snippets.
type CustomCode struct {
	HeadStart string `json:"headStart,omitempty"`
	HeadEnd   string `json:"headEnd,omitempty"`
	BodyStart string `json:"bodyStart,omitempty"`
	BodyEnd   string `json:"bodyEnd,omitempty"`
}

// ScreenshotResult holds the output of a screenshot or SVG export operation.
type ScreenshotResult struct {
	Image []byte `json:"image,omitempty"`
	URL   string `json:"url,omitempty"`
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// confirmDestructive checks if --confirm is set; if not, prints a warning and returns false.
func confirmDestructive(cmd *cobra.Command, action string) bool {
	confirm, _ := cmd.Flags().GetBool("confirm")
	if !confirm {
		fmt.Fprintf(os.Stderr, "This will %s. Use --confirm to proceed.\n", action)
		return false
	}
	return true
}

// dryRunResult returns a dry-run response if --dry-run is set.
func dryRunResult(cmd *cobra.Command, action string, data any) (bool, error) {
	if cli.IsDryRun(cmd) {
		result := map[string]any{
			"dry_run": true,
			"action":  action,
		}
		if data != nil {
			result["data"] = data
		}
		return true, cli.PrintResult(cmd, result, []string{fmt.Sprintf("[dry-run] Would %s", action)})
	}
	return false, nil
}

// parseJSONFlag parses a JSON string from a flag value.
func parseJSONFlag(value string) (json.RawMessage, error) {
	if value == "" {
		return nil, fmt.Errorf("JSON value is required")
	}
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(value), &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return raw, nil
}

// parseJSONFlagOrFile parses JSON from either a flag value or a file path.
func parseJSONFlagOrFile(value, filePath string) (json.RawMessage, error) {
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", filePath, err)
		}
		return parseJSONFlag(string(data))
	}
	return parseJSONFlag(value)
}

// parseStringList splits a comma-separated string into a slice.
func parseStringList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// readFileOrFlag returns text from either --flag or --flag-file.
func readFileOrFlag(value, filePath string) (string, error) {
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("read file %s: %w", filePath, err)
		}
		return string(data), nil
	}
	return value, nil
}
