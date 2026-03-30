package obituaries

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// ObituaryName holds the parsed name components from the Legacy.com API.
type ObituaryName struct {
	First string `json:"first"`
	Last  string `json:"last"`
	Full  string `json:"full"`
}

// ObituarySummary is a simplified obituary view returned by the search command.
type ObituarySummary struct {
	ID          int          `json:"id"`
	Name        ObituaryName `json:"name"`
	Age         int          `json:"age,omitempty"`
	City        string       `json:"city"`
	State       string       `json:"state"`
	PublishDate string       `json:"publish_date"`
	URL         string       `json:"url,omitempty"`
	Publication string       `json:"publication,omitempty"`
}

// NameEntry is a minimal name record used by the `names` command for ACRIS cross-referencing.
type NameEntry struct {
	First       string `json:"first"`
	Last        string `json:"last"`
	Full        string `json:"full"`
	PublishDate string `json:"publish_date"`
}

// SearchResponse maps to the Legacy.com search API response envelope.
type SearchResponse struct {
	Obituaries []rawObituary `json:"obituaries"`
	TotalCount int           `json:"totalCount"`
	Page       int           `json:"page"`
	PageSize   int           `json:"pageSize"`
}

// rawObituary matches the wire format of a single obituary from the Legacy.com API.
type rawObituary struct {
	ID          int          `json:"id"`
	Name        ObituaryName `json:"name"`
	City        string       `json:"city"`
	State       string       `json:"state"`
	PublishDate string       `json:"publishDate"`
	Age         int          `json:"age,omitempty"`
	URL         string       `json:"url,omitempty"`
	Publication string       `json:"publication,omitempty"`
}

// toSummary converts a rawObituary to an ObituarySummary.
func (r rawObituary) toSummary() ObituarySummary {
	return ObituarySummary{
		ID:          r.ID,
		Name:        r.Name,
		Age:         r.Age,
		City:        r.City,
		State:       r.State,
		PublishDate: r.PublishDate,
		URL:         r.URL,
		Publication: r.Publication,
	}
}

// toNameEntry converts a rawObituary to a NameEntry.
func (r rawObituary) toNameEntry() NameEntry {
	return NameEntry{
		First:       r.Name.First,
		Last:        r.Name.Last,
		Full:        r.Name.Full,
		PublishDate: r.PublishDate,
	}
}

// printObituaries outputs obituary summaries as JSON or a text table.
func printObituaries(cmd *cobra.Command, summaries []ObituarySummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(summaries)
	}

	if len(summaries) == 0 {
		fmt.Println("No obituaries found.")
		return nil
	}

	lines := make([]string, 0, len(summaries)+1)
	lines = append(lines, fmt.Sprintf("%-35s  %-4s  %-18s  %-12s  %s",
		"NAME", "AGE", "LOCATION", "DATE", "URL"))
	for _, s := range summaries {
		age := "-"
		if s.Age > 0 {
			age = fmt.Sprintf("%d", s.Age)
		}
		location := truncateName(s.City+", "+s.State, 18)
		name := truncateName(s.Name.Full, 35)
		lines = append(lines, fmt.Sprintf("%-35s  %-4s  %-18s  %-12s  %s",
			name, age, location, s.PublishDate, s.URL))
	}
	cli.PrintText(lines)
	return nil
}

// printNames outputs name entries as JSON or a text table.
func printNames(cmd *cobra.Command, entries []NameEntry) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(entries)
	}

	if len(entries) == 0 {
		fmt.Println("No obituaries found.")
		return nil
	}

	lines := make([]string, 0, len(entries)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-20s  %-35s  %s",
		"FIRST", "LAST", "FULL", "DATE"))
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("%-20s  %-20s  %-35s  %s",
			truncateName(e.First, 20),
			truncateName(e.Last, 20),
			truncateName(e.Full, 35),
			e.PublishDate))
	}
	cli.PrintText(lines)
	return nil
}

// truncateName shortens s to at most max runes, appending "..." if truncated.
func truncateName(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}
