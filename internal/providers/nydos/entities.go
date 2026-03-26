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

		body, err := client.Query(ctx, client.dailyFilingsURL, params)
		if err != nil {
			return fmt.Errorf("match address entities: %w", err)
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
		tokens = append(tokens, tok)
	}
	return tokens
}
