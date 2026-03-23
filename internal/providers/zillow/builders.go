package zillow

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newBuildersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "builders",
		Short:   "Search and view home builders",
		Aliases: []string{"builder"},
	}

	cmd.AddCommand(newBuilderSearchCmd(factory))
	cmd.AddCommand(newBuilderGetCmd(factory))
	cmd.AddCommand(newBuilderCommunitiesCmd(factory))
	cmd.AddCommand(newBuilderReviewsCmd(factory))

	return cmd
}

func newBuilderSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for home builders by location",
		RunE:  makeRunBuilderSearch(factory),
	}
	cmd.Flags().String("location", "", "Location to search")
	cmd.Flags().Int("limit", 25, "Maximum results")
	cmd.MarkFlagRequired("location")
	return cmd
}

func makeRunBuilderSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		location, _ := cmd.Flags().GetString("location")
		limit, _ := cmd.Flags().GetInt("limit")

		reqURL := client.baseURL + "/graphql/?builderSearch=" + url.QueryEscape(location)
		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("search builders: %w", err)
		}

		summaries, err := parseBuilderSearchResults(body, limit)
		if err != nil {
			return fmt.Errorf("parse results: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(summaries)
		}

		if len(summaries) == 0 {
			fmt.Println("No builders found.")
			return nil
		}
		lines := []string{"Builders:"}
		lines = append(lines, fmt.Sprintf("  %-12s  %-30s  %-6s  %-5s",
			"ID", "NAME", "RATING", "REVS"))
		for _, b := range summaries {
			rating := "-"
			if b.Rating > 0 {
				rating = fmt.Sprintf("%.1f", b.Rating)
			}
			reviews := "-"
			if b.ReviewCount > 0 {
				reviews = fmt.Sprintf("%d", b.ReviewCount)
			}
			lines = append(lines, fmt.Sprintf("  %-12s  %-30s  %-6s  %-5s",
				b.BuilderID, truncate(b.Name, 30), rating, reviews))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newBuilderGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get builder details",
		RunE:  makeRunBuilderGet(factory),
	}
	cmd.Flags().String("builder-id", "", "Builder ID")
	cmd.MarkFlagRequired("builder-id")
	return cmd
}

func makeRunBuilderGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		builderID, _ := cmd.Flags().GetString("builder-id")

		reqURL := fmt.Sprintf("%s/graphql/?builderId=%s", client.baseURL, builderID)
		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("get builder: %w", err)
		}

		var resp map[string]any
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(resp)
		}

		data, _ := resp["data"].(map[string]any)
		if data == nil {
			fmt.Println("Builder not found.")
			return nil
		}
		builder, _ := data["builder"].(map[string]any)
		if builder == nil {
			fmt.Println("Builder not found.")
			return nil
		}

		lines := []string{
			fmt.Sprintf("Name:    %s", jsonStr(builder, "name")),
		}
		if rating, ok := builder["rating"].(float64); ok && rating > 0 {
			lines = append(lines, fmt.Sprintf("Rating:  %.1f", rating))
		}
		if rc, ok := builder["reviewCount"].(float64); ok && rc > 0 {
			lines = append(lines, fmt.Sprintf("Reviews: %d", int(rc)))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newBuilderCommunitiesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "communities",
		Short: "Get builder's communities",
		RunE:  makeRunBuilderCommunities(factory),
	}
	cmd.Flags().String("builder-id", "", "Builder ID")
	cmd.MarkFlagRequired("builder-id")
	return cmd
}

func makeRunBuilderCommunities(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		builderID, _ := cmd.Flags().GetString("builder-id")

		reqURL := fmt.Sprintf("%s/graphql/?builderId=%s&communities=true", client.baseURL, builderID)
		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("get communities: %w", err)
		}

		communities, err := parseBuilderCommunities(body)
		if err != nil {
			return fmt.Errorf("parse communities: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{
				"builderId":   builderID,
				"communities": communities,
			})
		}

		if len(communities) == 0 {
			fmt.Println("No communities found.")
			return nil
		}
		lines := []string{fmt.Sprintf("Communities for builder %s:", builderID)}
		for _, c := range communities {
			price := ""
			if c.PriceFrom > 0 {
				price = fmt.Sprintf("  %s - %s", formatPrice(c.PriceFrom), formatPrice(c.PriceTo))
			}
			lines = append(lines, fmt.Sprintf("  %s — %s%s", c.Name, c.Location, price))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newBuilderReviewsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reviews",
		Short: "Get builder reviews",
		RunE:  makeRunBuilderReviews(factory),
	}
	cmd.Flags().String("builder-id", "", "Builder ID")
	cmd.Flags().Int("limit", 10, "Maximum reviews")
	cmd.MarkFlagRequired("builder-id")
	return cmd
}

func makeRunBuilderReviews(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		builderID, _ := cmd.Flags().GetString("builder-id")

		reqURL := fmt.Sprintf("%s/graphql/?builderId=%s&reviews=true", client.baseURL, builderID)
		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("get reviews: %w", err)
		}

		var resp map[string]any
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(resp)
		}

		fmt.Println("Builder reviews loaded.")
		return nil
	}
}

// parseBuilderSearchResults extracts builder summaries from search results.
func parseBuilderSearchResults(body []byte, limit int) ([]BuilderSummary, error) {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	data, _ := resp["data"].(map[string]any)
	if data == nil {
		return nil, nil
	}
	builders, _ := data["builders"].([]any)

	var summaries []BuilderSummary
	for _, item := range builders {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		s := BuilderSummary{
			BuilderID: jsonStr(m, "builderId"),
			Name:      jsonStr(m, "name"),
			URL:       jsonStr(m, "url"),
		}
		if rating, ok := m["rating"].(float64); ok {
			s.Rating = rating
		}
		if rc, ok := m["reviewCount"].(float64); ok {
			s.ReviewCount = int(rc)
		}
		summaries = append(summaries, s)
		if limit > 0 && len(summaries) >= limit {
			break
		}
	}
	return summaries, nil
}

// parseBuilderCommunities extracts communities from the builder detail response.
func parseBuilderCommunities(body []byte) ([]BuilderCommunity, error) {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	data, _ := resp["data"].(map[string]any)
	if data == nil {
		return nil, nil
	}
	builder, _ := data["builder"].(map[string]any)
	if builder == nil {
		return nil, nil
	}
	communities, _ := builder["communities"].([]any)

	var result []BuilderCommunity
	for _, item := range communities {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		c := BuilderCommunity{
			Name:     jsonStr(m, "name"),
			Location: jsonStr(m, "location"),
			URL:      jsonStr(m, "url"),
		}
		if pf, ok := m["priceFrom"].(float64); ok {
			c.PriceFrom = int64(pf)
		}
		if pt, ok := m["priceTo"].(float64); ok {
			c.PriceTo = int64(pt)
		}
		result = append(result, c)
	}
	return result, nil
}
