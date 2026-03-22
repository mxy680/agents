package x

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// TrendSummary is a condensed representation of an X trending topic.
type TrendSummary struct {
	Name       string `json:"name"`
	URL        string `json:"url,omitempty"`
	TweetCount int    `json:"tweet_count,omitempty"`
}

// TrendLocation is a location that supports trending topics.
type TrendLocation struct {
	Name    string `json:"name"`
	WOEID   int    `json:"woeid"`
	Country string `json:"country,omitempty"`
}

// newTrendsCmd builds the "trends" subcommand group.
func newTrendsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "trends",
		Short:   "View trending topics",
		Aliases: []string{"trend"},
	}
	cmd.AddCommand(newTrendsListCmd(factory))
	cmd.AddCommand(newTrendsLocationsCmd(factory))
	cmd.AddCommand(newTrendsByPlaceCmd(factory))
	return cmd
}

func newTrendsListCmd(factory ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List current trending topics",
		RunE:  makeRunTrendsList(factory),
	}
}

func newTrendsLocationsCmd(factory ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "locations",
		Short: "List locations that support trending topics",
		RunE:  makeRunTrendsLocations(factory),
	}
}

func newTrendsByPlaceCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "by-place",
		Short: "Get trending topics for a specific location",
		RunE:  makeRunTrendsByPlace(factory),
	}
	cmd.Flags().Int("woeid", 0, "Where On Earth ID (WOEID) for the location (required)")
	_ = cmd.MarkFlagRequired("woeid")
	return cmd
}

// --- RunE implementations ---

func makeRunTrendsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Get(ctx, "/i/api/2/guide.json", nil)
		if err != nil {
			return fmt.Errorf("fetching trends: %w", err)
		}

		var raw json.RawMessage
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decode trends response: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(raw)
		}

		trends := extractTrendsFromGuide(raw)
		return printTrendSummaries(cmd, trends)
	}
}

func makeRunTrendsLocations(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Get(ctx, "/i/api/1.1/trends/available.json", nil)
		if err != nil {
			return fmt.Errorf("fetching trend locations: %w", err)
		}

		var locations []json.RawMessage
		if err := client.DecodeJSON(resp, &locations); err != nil {
			return fmt.Errorf("decode locations response: %w", err)
		}

		parsed := make([]TrendLocation, 0, len(locations))
		for _, raw := range locations {
			var loc struct {
				Name    string `json:"name"`
				WoeID   int    `json:"woeid"`
				Country string `json:"country"`
			}
			if err := json.Unmarshal(raw, &loc); err != nil {
				continue
			}
			parsed = append(parsed, TrendLocation{
				Name:    loc.Name,
				WOEID:   loc.WoeID,
				Country: loc.Country,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(parsed)
		}

		if len(parsed) == 0 {
			fmt.Println("No trend locations found.")
			return nil
		}

		lines := make([]string, 0, len(parsed)+1)
		lines = append(lines, fmt.Sprintf("%-10s  %-30s  %-20s", "WOEID", "NAME", "COUNTRY"))
		for _, loc := range parsed {
			lines = append(lines, fmt.Sprintf("%-10d  %-30s  %-20s",
				loc.WOEID,
				truncate(loc.Name, 30),
				truncate(loc.Country, 20),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

func makeRunTrendsByPlace(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		woeid, _ := cmd.Flags().GetInt("woeid")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{}
		params.Set("id", fmt.Sprintf("%d", woeid))

		resp, err := client.Get(ctx, "/i/api/1.1/trends/place.json", params)
		if err != nil {
			return fmt.Errorf("fetching trends for woeid %d: %w", woeid, err)
		}

		var places []struct {
			Trends []struct {
				Name        string `json:"name"`
				URL         string `json:"url"`
				TweetVolume *int   `json:"tweet_volume"`
			} `json:"trends"`
		}
		if err := client.DecodeJSON(resp, &places); err != nil {
			return fmt.Errorf("decode trends by place response: %w", err)
		}

		var trends []TrendSummary
		if len(places) > 0 {
			for _, t := range places[0].Trends {
				count := 0
				if t.TweetVolume != nil {
					count = *t.TweetVolume
				}
				trends = append(trends, TrendSummary{
					Name:       t.Name,
					URL:        t.URL,
					TweetCount: count,
				})
			}
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(trends)
		}

		return printTrendSummaries(cmd, trends)
	}
}

// extractTrendsFromGuide parses trend entries from the /i/api/2/guide.json response.
func extractTrendsFromGuide(data json.RawMessage) []TrendSummary {
	var guide struct {
		Timeline struct {
			Instructions []struct {
				Type    string `json:"type"`
				Entries []struct {
					Content struct {
						TimelineModule *struct {
							Items []struct {
								Item struct {
									Content struct {
										Trend *struct {
											Name       string `json:"name"`
											TrendURL   string `json:"trendUrl"`
											DomainContext struct {
												EntityCount int `json:"entityCount"`
											} `json:"domainContext"`
										} `json:"trend"`
									} `json:"content"`
								} `json:"item"`
							} `json:"items"`
						} `json:"timelineModule"`
					} `json:"content"`
				} `json:"entries"`
			} `json:"instructions"`
		} `json:"timeline"`
	}

	if err := json.Unmarshal(data, &guide); err != nil {
		return nil
	}

	var trends []TrendSummary
	for _, instr := range guide.Timeline.Instructions {
		for _, entry := range instr.Entries {
			if entry.Content.TimelineModule == nil {
				continue
			}
			for _, item := range entry.Content.TimelineModule.Items {
				if t := item.Item.Content.Trend; t != nil {
					trends = append(trends, TrendSummary{
						Name:       t.Name,
						URL:        t.TrendURL,
						TweetCount: t.DomainContext.EntityCount,
					})
				}
			}
		}
	}

	return trends
}

// printTrendSummaries outputs trend summaries as JSON or text.
func printTrendSummaries(cmd *cobra.Command, trends []TrendSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(trends)
	}
	if len(trends) == 0 {
		fmt.Println("No trends found.")
		return nil
	}
	lines := make([]string, 0, len(trends)+1)
	lines = append(lines, fmt.Sprintf("%-30s  %-10s", "TREND", "TWEETS"))
	for _, t := range trends {
		lines = append(lines, fmt.Sprintf("%-30s  %-10d",
			truncate(t.Name, 30),
			t.TweetCount,
		))
	}
	cli.PrintText(lines)
	return nil
}
