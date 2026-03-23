package zillow

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newAgentsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agents",
		Short:   "Search and view real estate agents",
		Aliases: []string{"agent"},
	}

	cmd.AddCommand(newAgentSearchCmd(factory))
	cmd.AddCommand(newAgentGetCmd(factory))
	cmd.AddCommand(newAgentReviewsCmd(factory))
	cmd.AddCommand(newAgentListingsCmd(factory))

	return cmd
}

func newAgentSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for agents by location",
		RunE:  makeRunAgentSearch(factory),
	}
	cmd.Flags().String("location", "", "Location to search (e.g., 'Denver, CO')")
	cmd.Flags().String("name", "", "Agent name filter")
	cmd.Flags().String("specialty", "", "Specialty: buying, selling")
	cmd.Flags().Float64("rating", 0, "Minimum rating")
	cmd.Flags().Int("limit", 25, "Maximum results")
	cmd.MarkFlagRequired("location")
	return cmd
}

func makeRunAgentSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		location, _ := cmd.Flags().GetString("location")
		name, _ := cmd.Flags().GetString("name")
		limit, _ := cmd.Flags().GetInt("limit")

		// Zillow agent search uses a different endpoint
		qs := url.Values{}
		qs.Set("searchQueryState", fmt.Sprintf(`{"usersSearchTerm":"%s"}`, location))
		qs.Set("wants", `{"cat3":["agentResults"]}`)
		if name != "" {
			qs.Set("agentName", name)
		}
		reqURL := client.baseURL + "/search/GetSearchPageState.htm?" + qs.Encode()

		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("agent search: %w", err)
		}

		summaries, err := parseAgentSearchResults(body, limit)
		if err != nil {
			return fmt.Errorf("parse agent results: %w", err)
		}

		return printAgentSummaries(cmd, summaries)
	}
}

func newAgentGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get agent details by ID",
		RunE:  makeRunAgentGet(factory),
	}
	cmd.Flags().String("agent-id", "", "Agent ID")
	cmd.MarkFlagRequired("agent-id")
	return cmd
}

func makeRunAgentGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		agentID, _ := cmd.Flags().GetString("agent-id")

		reqURL := fmt.Sprintf("%s/graphql/?agentId=%s", client.baseURL, agentID)
		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("get agent: %w", err)
		}

		var resp map[string]any
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(resp)
		}

		// Parse and display agent details
		data, _ := resp["data"].(map[string]any)
		if data == nil {
			fmt.Println("Agent not found.")
			return nil
		}
		agent, _ := data["agent"].(map[string]any)
		if agent == nil {
			fmt.Println("Agent not found.")
			return nil
		}

		lines := []string{
			fmt.Sprintf("Name:     %s", jsonStr(agent, "name")),
			fmt.Sprintf("Phone:    %s", jsonStr(agent, "phone")),
		}
		if rating, ok := agent["rating"].(float64); ok && rating > 0 {
			lines = append(lines, fmt.Sprintf("Rating:   %.1f", rating))
		}
		if reviews, ok := agent["reviewCount"].(float64); ok && reviews > 0 {
			lines = append(lines, fmt.Sprintf("Reviews:  %d", int(reviews)))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newAgentReviewsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reviews",
		Short: "Get agent reviews",
		RunE:  makeRunAgentReviews(factory),
	}
	cmd.Flags().String("agent-id", "", "Agent ID")
	cmd.Flags().Int("limit", 10, "Maximum reviews")
	cmd.MarkFlagRequired("agent-id")
	return cmd
}

func makeRunAgentReviews(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		agentID, _ := cmd.Flags().GetString("agent-id")
		limit, _ := cmd.Flags().GetInt("limit")

		reqURL := fmt.Sprintf("%s/graphql/?agentId=%s&reviews=true", client.baseURL, agentID)
		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("get reviews: %w", err)
		}

		var resp map[string]any
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		reviews := parseAgentReviews(resp, limit)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{
				"agentId": agentID,
				"reviews": reviews,
			})
		}

		if len(reviews) == 0 {
			fmt.Println("No reviews found.")
			return nil
		}
		lines := []string{fmt.Sprintf("Reviews for agent %s:", agentID)}
		for _, r := range reviews {
			header := fmt.Sprintf("  %.0f★", r.Rating)
			if r.Reviewer != "" {
				header += " by " + r.Reviewer
			}
			if r.Date != "" {
				header += " (" + r.Date + ")"
			}
			lines = append(lines, header)
			if r.Description != "" {
				lines = append(lines, fmt.Sprintf("    %s", truncate(r.Description, 120)))
			}
		}
		cli.PrintText(lines)
		return nil
	}
}

func newAgentListingsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "listings",
		Short: "Get agent's listings",
		RunE:  makeRunAgentListings(factory),
	}
	cmd.Flags().String("agent-id", "", "Agent ID")
	cmd.Flags().String("status", "for_sale", "Listing status: for_sale, for_rent, sold")
	cmd.Flags().Int("limit", 25, "Maximum results")
	cmd.MarkFlagRequired("agent-id")
	return cmd
}

func makeRunAgentListings(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		agentID, _ := cmd.Flags().GetString("agent-id")
		status, _ := cmd.Flags().GetString("status")
		limit, _ := cmd.Flags().GetInt("limit")

		reqURL := fmt.Sprintf("%s/graphql/?agentId=%s&listings=%s", client.baseURL, agentID, status)
		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("get listings: %w", err)
		}

		var resp map[string]any
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		summaries := parseAgentListings(resp, limit)
		return printPropertySummaries(cmd, summaries)
	}
}

// parseAgentSearchResults extracts agent summaries from search results.
func parseAgentSearchResults(body []byte, limit int) ([]AgentSummary, error) {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	cat3, _ := resp["cat3"].(map[string]any)
	if cat3 == nil {
		return nil, nil
	}
	results, _ := cat3["agentResults"].([]any)

	var summaries []AgentSummary
	for _, item := range results {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		s := AgentSummary{
			AgentID:    jsonStr(m, "agentId"),
			Name:       jsonStr(m, "name"),
			Phone:      jsonStr(m, "phone"),
			ProfileURL: jsonStr(m, "profileUrl"),
			Photo:      jsonStr(m, "photo"),
		}
		if rating, ok := m["rating"].(float64); ok {
			s.Rating = rating
		}
		if rc, ok := m["reviewCount"].(float64); ok {
			s.ReviewCount = int(rc)
		}
		if sales, ok := m["recentSales"].(float64); ok {
			s.RecentSales = int(sales)
		}
		summaries = append(summaries, s)
		if limit > 0 && len(summaries) >= limit {
			break
		}
	}
	return summaries, nil
}

// parseAgentReviews extracts reviews from the agent detail response.
func parseAgentReviews(resp map[string]any, limit int) []AgentReview {
	data, _ := resp["data"].(map[string]any)
	if data == nil {
		return nil
	}
	agent, _ := data["agent"].(map[string]any)
	if agent == nil {
		return nil
	}
	reviews, _ := agent["reviews"].([]any)

	var result []AgentReview
	for _, item := range reviews {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		r := AgentReview{
			Date:        jsonStr(m, "date"),
			Description: jsonStr(m, "description"),
			Reviewer:    jsonStr(m, "reviewer"),
		}
		if rating, ok := m["rating"].(float64); ok {
			r.Rating = rating
		}
		result = append(result, r)
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

// parseAgentListings extracts property summaries from agent listings response.
func parseAgentListings(resp map[string]any, limit int) []PropertySummary {
	data, _ := resp["data"].(map[string]any)
	if data == nil {
		return nil
	}
	agent, _ := data["agent"].(map[string]any)
	if agent == nil {
		return nil
	}
	listings, _ := agent["listings"].([]any)

	var summaries []PropertySummary
	for _, item := range listings {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		s := PropertySummary{
			ZPID:    jsonStr(m, "zpid"),
			Address: jsonStr(m, "address"),
			Status:  jsonStr(m, "status"),
		}
		if price, ok := m["price"].(float64); ok {
			s.Price = int64(price)
		}
		if beds, ok := m["beds"].(float64); ok {
			s.Beds = int(beds)
		}
		if baths, ok := m["baths"].(float64); ok {
			s.Baths = baths
		}
		if sqft, ok := m["sqft"].(float64); ok {
			s.Sqft = int(sqft)
		}
		summaries = append(summaries, s)
		if limit > 0 && len(summaries) >= limit {
			break
		}
	}
	return summaries
}
