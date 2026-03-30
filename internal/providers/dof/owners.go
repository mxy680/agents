package dof

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
)

func newOwnersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "owners",
		Short:   "Query NYC DOF property owner records",
		Aliases: []string{"owner", "own"},
	}

	cmd.AddCommand(newOwnersLookupCmd(factory))
	cmd.AddCommand(newOwnersSearchCmd(factory))
	cmd.AddCommand(newOwnersByEntityCmd(factory))

	return cmd
}

// newOwnersLookupCmd returns the `owners lookup` subcommand.
func newOwnersLookupCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lookup",
		Short: "Look up property owner by BBL",
		RunE:  makeRunOwnersLookup(factory),
	}
	cmd.Flags().String("bbl", "", "10-digit Borough-Block-Lot identifier (e.g. 2029640028)")
	if err := cmd.MarkFlagRequired("bbl"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunOwnersLookup(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		bbl, _ := cmd.Flags().GetString("bbl")

		params := url.Values{
			"$where": []string{fmt.Sprintf("bble='%s'", bbl)},
			"$limit": []string{"1"},
		}

		body, err := client.Get(ctx, params)
		if err != nil {
			return fmt.Errorf("fetch owner: %w", err)
		}

		var raw []rawOwnerRecord
		if err := json.Unmarshal(body, &raw); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		if len(raw) == 0 {
			return fmt.Errorf("no property found for BBL %q", bbl)
		}

		return printOwnerRecord(cmd, toOwnerRecord(raw[0]))
	}
}

// newOwnersSearchCmd returns the `owners search` subcommand.
func newOwnersSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for properties by owner name",
		RunE:  makeRunOwnersSearch(factory),
	}
	cmd.Flags().String("name", "", "Owner name search term (LIKE match, e.g. 'SMITH' or 'LLC')")
	cmd.Flags().String("borough", "", "Filter by borough: manhattan, bronx, brooklyn, queens, staten-island")
	cmd.Flags().Int("limit", 50, "Maximum number of results to return")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunOwnersSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		name, _ := cmd.Flags().GetString("name")
		borough, _ := cmd.Flags().GetString("borough")
		limit, _ := cmd.Flags().GetInt("limit")

		where := fmt.Sprintf("upper(owner) LIKE upper('%%%s%%')", name)
		if borough != "" {
			code, err := lookupBoroughCode(borough)
			if err != nil {
				return err
			}
			where += fmt.Sprintf(" AND boro='%s'", code)
		}

		params := url.Values{
			"$where": []string{where},
			"$limit": []string{fmt.Sprintf("%d", limit)},
		}

		body, err := client.Get(ctx, params)
		if err != nil {
			return fmt.Errorf("fetch owners: %w", err)
		}

		var raw []rawOwnerRecord
		if err := json.Unmarshal(body, &raw); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		return printOwnerRecords(cmd, toOwnerRecords(raw))
	}
}

// newOwnersByEntityCmd returns the `owners by-entity` subcommand.
func newOwnersByEntityCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "by-entity",
		Short: "Find all properties owned by entities matching a pattern",
		RunE:  makeRunOwnersByEntity(factory),
	}
	cmd.Flags().String("pattern", "", "Entity name pattern to match (e.g. 'LLC' or 'REALTY')")
	cmd.Flags().String("borough", "", "Filter by borough: manhattan, bronx, brooklyn, queens, staten-island")
	cmd.Flags().Int("limit", 100, "Maximum number of results to return")
	if err := cmd.MarkFlagRequired("pattern"); err != nil {
		panic(err)
	}
	return cmd
}

func makeRunOwnersByEntity(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		pattern, _ := cmd.Flags().GetString("pattern")
		borough, _ := cmd.Flags().GetString("borough")
		limit, _ := cmd.Flags().GetInt("limit")

		where := fmt.Sprintf("upper(owner) LIKE upper('%%%s%%')", pattern)
		if borough != "" {
			code, err := lookupBoroughCode(borough)
			if err != nil {
				return err
			}
			where += fmt.Sprintf(" AND boro='%s'", code)
		}

		params := url.Values{
			"$where":  []string{where},
			"$limit":  []string{fmt.Sprintf("%d", limit)},
			"$order":  []string{"owner ASC"},
		}

		body, err := client.Get(ctx, params)
		if err != nil {
			return fmt.Errorf("fetch entity properties: %w", err)
		}

		var raw []rawOwnerRecord
		if err := json.Unmarshal(body, &raw); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		return printOwnerRecords(cmd, toOwnerRecords(raw))
	}
}
