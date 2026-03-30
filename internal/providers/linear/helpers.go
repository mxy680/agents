package linear

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// --- Issue types ---

// IssueSummary is the JSON-serializable summary of a Linear issue.
type IssueSummary struct {
	ID         string `json:"id"`
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
	State      string `json:"state,omitempty"`
	Priority   int    `json:"priority"`
	Assignee   string `json:"assignee,omitempty"`
	CreatedAt  string `json:"createdAt,omitempty"`
}

// IssueDetail is the JSON-serializable full issue metadata.
type IssueDetail struct {
	ID          string   `json:"id"`
	Identifier  string   `json:"identifier"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	State       string   `json:"state,omitempty"`
	Priority    int      `json:"priority"`
	Assignee    string   `json:"assignee,omitempty"`
	Team        string   `json:"team,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	CreatedAt   string   `json:"createdAt,omitempty"`
	UpdatedAt   string   `json:"updatedAt,omitempty"`
}

// --- Project types ---

// ProjectSummary is the JSON-serializable summary of a Linear project.
type ProjectSummary struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	State    string   `json:"state,omitempty"`
	Progress float64  `json:"progress,omitempty"`
	Teams    []string `json:"teams,omitempty"`
}

// ProjectDetail is the JSON-serializable full project metadata.
type ProjectDetail struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	State       string  `json:"state,omitempty"`
	StartDate   string  `json:"startDate,omitempty"`
	TargetDate  string  `json:"targetDate,omitempty"`
	Progress    float64 `json:"progress,omitempty"`
}

// --- Cycle types ---

// CycleSummary is the JSON-serializable summary of a Linear cycle.
type CycleSummary struct {
	ID       string `json:"id"`
	Number   int    `json:"number"`
	StartsAt string `json:"startsAt,omitempty"`
	EndsAt   string `json:"endsAt,omitempty"`
}

// CycleDetail is the JSON-serializable full cycle metadata including issues.
type CycleDetail struct {
	ID       string         `json:"id"`
	Number   int            `json:"number"`
	StartsAt string         `json:"startsAt,omitempty"`
	EndsAt   string         `json:"endsAt,omitempty"`
	Issues   []IssueSummary `json:"issues,omitempty"`
}

// --- Team types ---

// TeamSummary is the JSON-serializable summary of a Linear team.
type TeamSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key,omitempty"`
}

// TeamDetail is the JSON-serializable full team metadata.
type TeamDetail struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Key         string       `json:"key,omitempty"`
	Description string       `json:"description,omitempty"`
	Members     []UserSummary `json:"members,omitempty"`
}

// --- Comment types ---

// CommentSummary is the JSON-serializable summary of an issue comment.
type CommentSummary struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	User      string `json:"user,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

// --- Label types ---

// LabelSummary is the JSON-serializable summary of an issue label.
type LabelSummary struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// --- User types ---

// UserSummary is the JSON-serializable summary of a Linear user.
type UserSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	Active      bool   `json:"active"`
}

// --- Workflow state types ---

// WorkflowState is the JSON-serializable representation of a workflow state.
type WorkflowState struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Color    string  `json:"color,omitempty"`
	Type     string  `json:"type,omitempty"`
	Position float64 `json:"position,omitempty"`
}

// --- Webhook types ---

// WebhookSummary is the JSON-serializable summary of a Linear webhook.
type WebhookSummary struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Enabled   bool   `json:"enabled"`
	Team      string `json:"team,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
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

// priorityLabel converts a Linear priority integer to a human-readable label.
func priorityLabel(p int) string {
	switch p {
	case 0:
		return "No priority"
	case 1:
		return "Urgent"
	case 2:
		return "High"
	case 3:
		return "Medium"
	case 4:
		return "Low"
	default:
		return fmt.Sprintf("%d", p)
	}
}

// --- Print helpers ---

func printIssueSummaries(cmd *cobra.Command, issues []IssueSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(issues)
	}
	if len(issues) == 0 {
		fmt.Println("No issues found.")
		return nil
	}
	lines := make([]string, 0, len(issues)+1)
	lines = append(lines, fmt.Sprintf("%-12s  %-45s  %-14s  %-10s  %s", "IDENTIFIER", "TITLE", "STATE", "PRIORITY", "ASSIGNEE"))
	for _, i := range issues {
		lines = append(lines, fmt.Sprintf("%-12s  %-45s  %-14s  %-10s  %s",
			truncate(i.Identifier, 12), truncate(i.Title, 45), truncate(i.State, 14), priorityLabel(i.Priority), truncate(i.Assignee, 20)))
	}
	cli.PrintText(lines)
	return nil
}

func printProjectSummaries(cmd *cobra.Command, projects []ProjectSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(projects)
	}
	if len(projects) == 0 {
		fmt.Println("No projects found.")
		return nil
	}
	lines := make([]string, 0, len(projects)+1)
	lines = append(lines, fmt.Sprintf("%-28s  %-35s  %-12s  %s", "ID", "NAME", "STATE", "PROGRESS"))
	for _, p := range projects {
		lines = append(lines, fmt.Sprintf("%-28s  %-35s  %-12s  %.0f%%",
			truncate(p.ID, 28), truncate(p.Name, 35), truncate(p.State, 12), p.Progress*100))
	}
	cli.PrintText(lines)
	return nil
}

func printTeamSummaries(cmd *cobra.Command, teams []TeamSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(teams)
	}
	if len(teams) == 0 {
		fmt.Println("No teams found.")
		return nil
	}
	lines := make([]string, 0, len(teams)+1)
	lines = append(lines, fmt.Sprintf("%-28s  %-30s  %s", "ID", "NAME", "KEY"))
	for _, t := range teams {
		lines = append(lines, fmt.Sprintf("%-28s  %-30s  %s", truncate(t.ID, 28), truncate(t.Name, 30), t.Key))
	}
	cli.PrintText(lines)
	return nil
}

func printUserSummaries(cmd *cobra.Command, users []UserSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(users)
	}
	if len(users) == 0 {
		fmt.Println("No users found.")
		return nil
	}
	lines := make([]string, 0, len(users)+1)
	lines = append(lines, fmt.Sprintf("%-28s  %-25s  %-35s  %s", "ID", "NAME", "EMAIL", "ACTIVE"))
	for _, u := range users {
		lines = append(lines, fmt.Sprintf("%-28s  %-25s  %-35s  %v",
			truncate(u.ID, 28), truncate(u.Name, 25), truncate(u.Email, 35), u.Active))
	}
	cli.PrintText(lines)
	return nil
}
