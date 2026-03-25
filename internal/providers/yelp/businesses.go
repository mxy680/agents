package yelp

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newBusinessesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "businesses",
		Short:   "Search and view Yelp businesses",
		Aliases: []string{"business", "biz"},
	}

	cmd.AddCommand(newBusinessSearchCmd(factory))
	cmd.AddCommand(newBusinessPhoneSearchCmd(factory))
	cmd.AddCommand(newBusinessMatchCmd(factory))
	cmd.AddCommand(newBusinessGetCmd(factory))

	return cmd
}

func newBusinessSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search businesses by term and location",
		RunE:  makeRunBusinessSearch(factory),
	}
	cmd.Flags().String("term", "", "Search term (e.g., 'food', 'restaurants')")
	cmd.Flags().String("location", "", "Location text (e.g., 'San Francisco, CA')")
	cmd.Flags().Float64("latitude", 0, "Latitude for geo-based search")
	cmd.Flags().Float64("longitude", 0, "Longitude for geo-based search")
	cmd.Flags().Int("radius", 0, "Search radius in meters (max 40000)")
	cmd.Flags().String("categories", "", "Comma-separated category aliases to filter by")
	cmd.Flags().String("price", "", "Comma-separated price tiers to filter by (1,2,3,4)")
	cmd.Flags().String("sort-by", "", "Sort: best_match, rating, review_count, distance")
	cmd.Flags().Bool("open-now", false, "Only return businesses open now")
	cmd.Flags().Int("limit", 20, "Maximum number of results (max 50)")
	cmd.Flags().Int("offset", 0, "Offset for pagination")
	return cmd
}

func makeRunBusinessSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		term, _ := cmd.Flags().GetString("term")
		location, _ := cmd.Flags().GetString("location")
		latitude, _ := cmd.Flags().GetFloat64("latitude")
		longitude, _ := cmd.Flags().GetFloat64("longitude")
		radius, _ := cmd.Flags().GetInt("radius")
		categories, _ := cmd.Flags().GetString("categories")
		price, _ := cmd.Flags().GetString("price")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		openNow, _ := cmd.Flags().GetBool("open-now")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		if location == "" && (latitude == 0 || longitude == 0) {
			return fmt.Errorf("either --location or both --latitude and --longitude are required")
		}

		params := url.Values{}
		if term != "" {
			params.Set("term", term)
		}
		if location != "" {
			params.Set("location", location)
		}
		if latitude != 0 {
			params.Set("latitude", strconv.FormatFloat(latitude, 'f', -1, 64))
		}
		if longitude != 0 {
			params.Set("longitude", strconv.FormatFloat(longitude, 'f', -1, 64))
		}
		if radius > 0 {
			params.Set("radius", strconv.Itoa(radius))
		}
		if categories != "" {
			params.Set("categories", categories)
		}
		if price != "" {
			params.Set("price", price)
		}
		if sortBy != "" {
			params.Set("sort_by", sortBy)
		}
		if openNow {
			params.Set("open_now", "true")
		}
		if limit > 0 {
			params.Set("limit", strconv.Itoa(limit))
		}
		if offset > 0 {
			params.Set("offset", strconv.Itoa(offset))
		}

		body, err := client.doYelp(ctx, "GET", "/businesses/search", params)
		if err != nil {
			return fmt.Errorf("business search: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(body, &raw); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var resp struct {
			Businesses []BusinessSummary `json:"businesses"`
			Total      int               `json:"total"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		return printBusinessSummaries(cmd, resp.Businesses)
	}
}

func newBusinessPhoneSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "phone-search",
		Short: "Search businesses by phone number",
		RunE:  makeRunBusinessPhoneSearch(factory),
	}
	cmd.Flags().String("phone", "", "Phone number including country code (e.g., +14159083801)")
	_ = cmd.MarkFlagRequired("phone")
	return cmd
}

func makeRunBusinessPhoneSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		phone, _ := cmd.Flags().GetString("phone")

		params := url.Values{}
		params.Set("phone", phone)

		body, err := client.doYelp(ctx, "GET", "/businesses/search/phone", params)
		if err != nil {
			return fmt.Errorf("phone search: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(body, &raw); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var resp struct {
			Businesses []BusinessSummary `json:"businesses"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		return printBusinessSummaries(cmd, resp.Businesses)
	}
}

func newBusinessMatchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "match",
		Short: "Find a business by name and address",
		RunE:  makeRunBusinessMatch(factory),
	}
	cmd.Flags().String("name", "", "Business name")
	cmd.Flags().String("city", "", "City")
	cmd.Flags().String("state", "", "State code (e.g., CA)")
	cmd.Flags().String("country", "", "Country code (e.g., US)")
	cmd.Flags().String("address1", "", "Street address")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("city")
	_ = cmd.MarkFlagRequired("state")
	_ = cmd.MarkFlagRequired("country")
	return cmd
}

func makeRunBusinessMatch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		name, _ := cmd.Flags().GetString("name")
		city, _ := cmd.Flags().GetString("city")
		state, _ := cmd.Flags().GetString("state")
		country, _ := cmd.Flags().GetString("country")
		address1, _ := cmd.Flags().GetString("address1")

		params := url.Values{}
		params.Set("name", name)
		params.Set("city", city)
		params.Set("state", state)
		params.Set("country", country)
		if address1 != "" {
			params.Set("address1", address1)
		}

		body, err := client.doYelp(ctx, "GET", "/businesses/matches", params)
		if err != nil {
			return fmt.Errorf("business match: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(body, &raw); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var resp struct {
			Businesses []BusinessSummary `json:"businesses"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		return printBusinessSummaries(cmd, resp.Businesses)
	}
}

func newBusinessGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific business by ID",
		RunE:  makeRunBusinessGet(factory),
	}
	cmd.Flags().String("id", "", "Yelp business ID or alias")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunBusinessGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		id, _ := cmd.Flags().GetString("id")

		body, err := client.doYelp(ctx, "GET", "/businesses/"+id, nil)
		if err != nil {
			return fmt.Errorf("get business: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw json.RawMessage
			if err := json.Unmarshal(body, &raw); err != nil {
				return fmt.Errorf("parse response: %w", err)
			}
			return cli.PrintJSON(raw)
		}

		var detail BusinessDetail
		if err := json.Unmarshal(body, &detail); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		lines := formatBusinessDetail(detail)
		cli.PrintText(lines)
		return nil
	}
}

// formatBusinessDetail formats a BusinessDetail for text output.
func formatBusinessDetail(b BusinessDetail) []string {
	lines := []string{
		fmt.Sprintf("Name:        %s", b.Name),
		fmt.Sprintf("ID:          %s", b.ID),
		fmt.Sprintf("Rating:      %s (%d reviews)", formatRating(b.Rating), b.ReviewCount),
		fmt.Sprintf("Price:       %s", orDash(b.Price)),
		fmt.Sprintf("Phone:       %s", orDash(b.Phone)),
		fmt.Sprintf("Address:     %s", formatAddress(b.Location)),
		fmt.Sprintf("Categories:  %s", orDash(formatCategories(b.Categories))),
		fmt.Sprintf("Claimed:     %v", b.IsClaimed),
		fmt.Sprintf("URL:         %s", b.URL),
	}

	if len(b.Photos) > 0 {
		lines = append(lines, fmt.Sprintf("Photos:      %d available", len(b.Photos)))
	}

	for _, h := range b.Hours {
		status := "closed"
		if h.IsOpenNow {
			status = "open now"
		}
		lines = append(lines, fmt.Sprintf("Hours (%s): %s", h.HoursType, status))
	}

	return lines
}

// orDash returns s or "-" if s is empty.
func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
