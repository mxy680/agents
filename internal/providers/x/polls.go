package x

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// PollResult holds information about a poll creation or vote.
type PollResult struct {
	CardURI string `json:"card_uri,omitempty"`
	Status  string `json:"status"`
}

// newPollsCmd builds the "polls" subcommand group.
func newPollsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "polls",
		Short:   "Create polls and vote on them",
		Aliases: []string{"poll"},
	}
	cmd.AddCommand(newPollsCreateCmd(factory))
	cmd.AddCommand(newPollsVoteCmd(factory))
	return cmd
}

func newPollsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a poll card",
		RunE:  makeRunPollsCreate(factory),
	}
	cmd.Flags().StringSlice("options", nil, "Poll options (2-4, required)")
	_ = cmd.MarkFlagRequired("options")
	cmd.Flags().Int("duration", 1440, "Poll duration in minutes (required)")
	_ = cmd.MarkFlagRequired("duration")
	return cmd
}

func newPollsVoteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote",
		Short: "Vote on a poll",
		RunE:  makeRunPollsVote(factory),
	}
	cmd.Flags().String("tweet-id", "", "Tweet ID containing the poll (required)")
	_ = cmd.MarkFlagRequired("tweet-id")
	cmd.Flags().Int("choice", 0, "Choice number (1-indexed, required)")
	_ = cmd.MarkFlagRequired("choice")
	cmd.Flags().Bool("dry-run", false, "Print what would be voted without voting")
	return cmd
}

// --- RunE implementations ---

func makeRunPollsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		options, _ := cmd.Flags().GetStringSlice("options")
		duration, _ := cmd.Flags().GetInt("duration")

		// Filter empty options.
		cleaned := make([]string, 0, len(options))
		for _, o := range options {
			o = strings.TrimSpace(o)
			if o != "" {
				cleaned = append(cleaned, o)
			}
		}

		if len(cleaned) < 2 || len(cleaned) > 4 {
			return fmt.Errorf("polls require 2-4 options, got %d", len(cleaned))
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"card_data": buildPollCardData(cleaned, duration),
		}

		var result struct {
			CardURI string `json:"card_uri"`
		}
		if err := client.postJSONToFullURL(ctx, client.capsURL+"/v2/cards/create.json", body, &result); err != nil {
			return fmt.Errorf("creating poll: %w", err)
		}

		pollResult := PollResult{
			CardURI: result.CardURI,
			Status:  "created",
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(pollResult)
		}
		fmt.Printf("Poll created. Card URI: %s\n", result.CardURI)
		return nil
	}
}

func makeRunPollsVote(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		tweetID, _ := cmd.Flags().GetString("tweet-id")
		choice, _ := cmd.Flags().GetInt("choice")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if choice < 1 || choice > 4 {
			return fmt.Errorf("choice must be between 1 and 4, got %d", choice)
		}

		selectedChoice := fmt.Sprintf("choice%d", choice)

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("vote %s on poll in tweet %s", selectedChoice, tweetID), map[string]string{
				"tweet_id":        tweetID,
				"selected_choice": selectedChoice,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"twitter:string:card_uri":         fmt.Sprintf("card://tweet:%s", tweetID),
			"twitter:string:selected_choice":  selectedChoice,
		}

		var raw json.RawMessage
		if err := client.postJSONToFullURL(ctx, client.capsURL+"/v2/capi/passthrough/1", body, &raw); err != nil {
			return fmt.Errorf("voting on poll in tweet %s: %w", tweetID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(raw)
		}
		fmt.Printf("Voted %s on poll in tweet %s\n", selectedChoice, tweetID)
		return nil
	}
}

// postJSONToFullURL sends a POST JSON request to any full URL and decodes the response.
func (c *Client) postJSONToFullURL(ctx context.Context, fullURL string, body any, target any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	c.applyHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request to %s: %w", fullURL, err)
	}
	c.captureResponseHeaders(resp)

	return c.DecodeJSON(resp, target)
}

// buildPollCardData assembles the card data payload for poll creation.
func buildPollCardData(options []string, durationMinutes int) map[string]any {
	data := map[string]any{
		"card_name":        fmt.Sprintf("poll%dchoice_text_only", len(options)),
		"duration_minutes": durationMinutes,
		"counts_are_final": false,
		"choices_count":    len(options),
	}
	for i, o := range options {
		data[fmt.Sprintf("choice%d_label", i+1)] = o
	}
	return data
}
