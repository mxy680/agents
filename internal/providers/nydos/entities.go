package nydos

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newEntitiesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "entities",
		Short:   "Query NY DOS entity formations and active corporations",
		Aliases: []string{"ent", "entity"},
	}

	cmd.AddCommand(newEntitiesRecentCmd(factory))
	cmd.AddCommand(newEntitiesSearchCmd(factory))
	cmd.AddCommand(newEntitiesMatchAddressCmd(factory))

	return cmd
}

// newEntitiesRecentCmd returns the `entities recent` subcommand.
func newEntitiesRecentCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recent",
		Short: "Get new entity formations in a date range",
		RunE:  makeRunEntitiesRecent(factory),
	}
	cmd.Flags().String("since", "", "Start date for filing_date filter (e.g. 2026-03-01)")
	cmd.Flags().String("type", "", "Entity type filter: llc, corp, lp")
	cmd.Flags().String("county", "", "Filter by county (e.g. NEW YORK, KINGS, BRONX, QUEENS, RICHMOND)")
	cmd.Flags().Int("limit", 1000, "Maximum number of results")
	if err := cmd.MarkFlagRequired("since"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunEntitiesRecent(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		since, _ := cmd.Flags().GetString("since")
		entityTypeFilter, _ := cmd.Flags().GetString("type")
		county, _ := cmd.Flags().GetString("county")
		limit, _ := cmd.Flags().GetInt("limit")

		whereClause := fmt.Sprintf("filing_date>='%s'", since)

		if entityTypeFilter != "" {
			fullType, err := lookupEntityType(entityTypeFilter)
			if err != nil {
				return err
			}
			whereClause += fmt.Sprintf(" AND entity_type='%s'", fullType)
		}

		if county != "" {
			whereClause += fmt.Sprintf(" AND jurisdiction='%s'", strings.ToUpper(county))
		}

		params := url.Values{
			"$where": []string{whereClause},
			"$limit": []string{fmt.Sprintf("%d", limit)},
			"$order": []string{"filing_date DESC"},
		}

		body, err := client.Query(ctx, client.dailyFilingsURL, params)
		if err != nil {
			return fmt.Errorf("fetch recent entities: %w", err)
		}

		var records []DailyFilingRecord
		if err := json.Unmarshal(body, &records); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		entities := make([]EntitySummary, len(records))
		for i, r := range records {
			entities[i] = dailyFilingToSummary(r)
		}

		return printEntities(cmd, entities)
	}
}

// newEntitiesSearchCmd returns the `entities search` subcommand.
func newEntitiesSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search active corporations by name",
		RunE:  makeRunEntitiesSearch(factory),
	}
	cmd.Flags().String("name", "", "Name search term (uses LIKE query)")
	cmd.Flags().String("type", "", "Entity type filter: llc, corp, lp")
	cmd.Flags().Int("limit", 50, "Maximum number of results")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunEntitiesSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		name, _ := cmd.Flags().GetString("name")
		entityTypeFilter, _ := cmd.Flags().GetString("type")
		limit, _ := cmd.Flags().GetInt("limit")

		whereClause := fmt.Sprintf("current_entity_name like '%%%s%%'", strings.ToUpper(name))

		if entityTypeFilter != "" {
			fullType, err := lookupEntityType(entityTypeFilter)
			if err != nil {
				return err
			}
			whereClause += fmt.Sprintf(" AND entity_type_code='%s'", fullType)
		}

		params := url.Values{
			"$where": []string{whereClause},
			"$limit": []string{fmt.Sprintf("%d", limit)},
		}

		body, err := client.Query(ctx, client.activeCorpsURL, params)
		if err != nil {
			return fmt.Errorf("search entities: %w", err)
		}

		var records []ActiveCorpRecord
		if err := json.Unmarshal(body, &records); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		entities := make([]EntitySummary, len(records))
		for i, r := range records {
			entities[i] = activeCorpToSummary(r)
		}

		return printEntities(cmd, entities)
	}
}

// newEntitiesMatchAddressCmd returns the `entities match-address` subcommand.
func newEntitiesMatchAddressCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "match-address",
		Short: "Find recently-formed entities referencing a property address",
		Long: `Search for entities whose corp_name contains tokens from the given address.
Useful for the "in contract" use case: find LLCs formed recently that may reference a specific address.
Address tokens (street number + street name) are searched with OR logic.`,
		RunE: makeRunEntitiesMatchAddress(factory),
	}
	cmd.Flags().String("address", "", `Address pattern to search for (e.g. "1776 SEMINOLE")`)
	cmd.Flags().String("since", "", "Filter by filing_date >= this date (default: 90 days ago)")
	cmd.Flags().Int("limit", 50, "Maximum number of results")
	if err := cmd.MarkFlagRequired("address"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunEntitiesMatchAddress(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		address, _ := cmd.Flags().GetString("address")
		since, _ := cmd.Flags().GetString("since")
		limit, _ := cmd.Flags().GetInt("limit")

		tokens := addressTokens(address)
		if len(tokens) == 0 {
			return fmt.Errorf("address %q produced no searchable tokens", address)
		}

		// Build OR clause: corp_name like '%TOKEN1%' OR corp_name like '%TOKEN2%'
		var clauses []string
		for _, tok := range tokens {
			clauses = append(clauses, fmt.Sprintf("corp_name like '%%%s%%'", tok))
		}
		whereClause := strings.Join(clauses, " OR ")

		if since != "" {
			whereClause = fmt.Sprintf("(%s) AND filing_date>='%s'", whereClause, since)
		}

		params := url.Values{
			"$where": []string{whereClause},
			"$limit": []string{fmt.Sprintf("%d", limit)},
			"$order": []string{"filing_date DESC"},
		}

		var entities []EntitySummary
		seen := map[string]bool{}

		// Strategy 1: Phrase search on active corps (most reliable)
		// Build a compact phrase from the address tokens: "540 WEST 29" -> LIKE '%540%WEST%29%'
		phrasePattern := "%" + strings.Join(tokens, "%") + "%"
		phraseParams := url.Values{
			"$where": []string{fmt.Sprintf("current_entity_name like '%s'", phrasePattern)},
			"$limit": []string{fmt.Sprintf("%d", limit)},
			"$order": []string{"initial_dos_filing_date DESC"},
		}

		acBody, acErr := client.Query(ctx, client.activeCorpsURL, phraseParams)
		if acErr == nil {
			var acRecords []ActiveCorpRecord
			if json.Unmarshal(acBody, &acRecords) == nil {
				for _, r := range acRecords {
					s := activeCorpToSummary(r)
					if !seen[s.Name] {
						seen[s.Name] = true
						entities = append(entities, s)
					}
				}
			}
		}

		// Strategy 2: Daily filings with OR tokens (catches recent filings with filer info)
		body, err := client.Query(ctx, client.dailyFilingsURL, params)
		if err == nil {
			var records []DailyFilingRecord
			if json.Unmarshal(body, &records) == nil {
				for _, r := range records {
					s := dailyFilingToSummary(r)
					// Only add if ALL tokens appear in the name (tighter matching)
					nameUpper := strings.ToUpper(s.Name)
					allMatch := true
					for _, tok := range tokens {
						if !strings.Contains(nameUpper, strings.ToUpper(tok)) {
							allMatch = false
							break
						}
					}
					if allMatch && !seen[s.Name] {
						seen[s.Name] = true
						entities = append(entities, s)
					}
				}
			}
		}

		if len(entities) == 0 && acErr != nil {
			return fmt.Errorf("match address entities: %w", acErr)
		}

		return printEntities(cmd, entities)
	}
}

// addressTokens splits an address string into uppercase tokens, filtering out
// common stop words that would produce too many false positives.
func addressTokens(address string) []string {
	stopWords := map[string]bool{
		"AVE": true, "AVENUE": true,
		"ST":     true,
		"STREET": true,
		"RD":     true,
		"ROAD":   true,
		"BLVD":   true,
		"DR":     true,
		"DRIVE":  true,
		"LN":     true,
		"LANE":   true,
		"CT":     true,
		"COURT":  true,
		"PL":     true,
		"PLACE":  true,
		"WAY":    true,
		"PKWY":   true,
		"HWY":    true,
		"NY":     true,
		"NYC":    true,
	}

	upper := strings.ToUpper(strings.TrimSpace(address))
	raw := strings.Fields(upper)

	var tokens []string
	for _, tok := range raw {
		// Strip trailing punctuation.
		tok = strings.TrimRight(tok, ".,;:")
		if tok == "" || stopWords[tok] {
			continue
		}
		// Strip ordinal suffixes from numbers: 29TH -> 29, 1ST -> 1, 2ND -> 2, 3RD -> 3
		for _, suffix := range []string{"TH", "ST", "ND", "RD"} {
			if strings.HasSuffix(tok, suffix) && len(tok) > len(suffix) {
				prefix := tok[:len(tok)-len(suffix)]
				if prefix != "" && prefix[0] >= '0' && prefix[0] <= '9' {
					tok = prefix
					break
				}
			}
		}
		tokens = append(tokens, tok)
	}
	return tokens
}
