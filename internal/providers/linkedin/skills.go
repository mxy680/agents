package linkedin

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerSkillsResponse is the response envelope for GET /voyager/api/identity/profiles/{id}/skills.
type voyagerSkillsResponse struct {
	Elements []voyagerSkillElement `json:"elements"`
	Paging   voyagerPaging         `json:"paging"`
}

type voyagerSkillElement struct {
	EntityURN        string `json:"entityUrn"`
	Name             string `json:"name"`
	EndorsementCount int    `json:"endorsementCount"`
}

// voyagerSkillEndorsementsResponse is the response for skill endorsements.
type voyagerSkillEndorsementsResponse struct {
	Elements []struct {
		EntityURN string `json:"entityUrn"`
		Endorser  struct {
			MiniProfile struct {
				FirstName        string `json:"firstName"`
				LastName         string `json:"lastName"`
				PublicIdentifier string `json:"publicIdentifier"`
			} `json:"miniProfile"`
		} `json:"endorser"`
	} `json:"elements"`
	Paging voyagerPaging `json:"paging"`
}

// toSkillSummary maps a voyagerSkillElement to SkillSummary.
func toSkillSummary(el voyagerSkillElement) SkillSummary {
	id := el.EntityURN
	if parts := strings.Split(el.EntityURN, ":"); len(parts) > 0 {
		id = parts[len(parts)-1]
	}
	return SkillSummary{
		ID:               id,
		Name:             el.Name,
		EndorsementCount: el.EndorsementCount,
	}
}

// newSkillsCmd builds the "skills" subcommand group.
func newSkillsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "skills",
		Short:   "Interact with LinkedIn profile skills",
		Aliases: []string{"skill"},
	}
	cmd.AddCommand(newSkillsListCmd(factory))
	cmd.AddCommand(newSkillsEndorseCmd(factory))
	cmd.AddCommand(newSkillsEndorsementsCmd(factory))
	return cmd
}

// newSkillsListCmd builds the "skills list" command.
func newSkillsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List skills on a LinkedIn profile",
		RunE:  makeRunSkillsList(factory),
	}
	cmd.Flags().String("username", "", "Profile public ID (defaults to current user)")
	return cmd
}

// newSkillsEndorseCmd builds the "skills endorse" command.
func newSkillsEndorseCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "endorse",
		Short: "Endorse a skill by URN",
		RunE:  makeRunSkillsEndorse(factory),
	}
	cmd.Flags().String("urn", "", "Skill entity URN (e.g. urn:li:fs_skill:123) (required)")
	cmd.Flags().String("skill-id", "", "Skill ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without endorsing")
	_ = cmd.MarkFlagRequired("urn")
	_ = cmd.MarkFlagRequired("skill-id")
	return cmd
}

// newSkillsEndorsementsCmd builds the "skills endorsements" command.
func newSkillsEndorsementsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "endorsements",
		Short: "List endorsers for a skill",
		RunE:  makeRunSkillsEndorsements(factory),
	}
	cmd.Flags().String("username", "", "Profile public ID (defaults to current user)")
	cmd.Flags().String("skill-id", "", "Skill ID (required)")
	cmd.Flags().Int("limit", 10, "Maximum number of endorsers to return")
	_ = cmd.MarkFlagRequired("skill-id")
	return cmd
}

func makeRunSkillsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		username, _ := cmd.Flags().GetString("username")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		// If no username provided, fetch the current user's public identifier.
		if username == "" {
			meResp, err := client.Get(ctx, "/voyager/api/me", nil)
			if err != nil {
				return fmt.Errorf("getting current user: %w", err)
			}
			var me struct {
				MiniProfile struct {
					PublicIdentifier string `json:"publicIdentifier"`
				} `json:"miniProfile"`
			}
			if err := client.DecodeJSON(meResp, &me); err != nil {
				return fmt.Errorf("decoding current user: %w", err)
			}
			username = me.MiniProfile.PublicIdentifier
		}

		path := "/voyager/api/identity/profiles/" + url.PathEscape(username) + "/skills"
		params := url.Values{"count": {"50"}}
		resp, err := client.Get(ctx, path, params)
		if err != nil {
			return fmt.Errorf("listing skills for %s: %w", username, err)
		}

		var raw voyagerSkillsResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding skills: %w", err)
		}

		summaries := make([]SkillSummary, 0, len(raw.Elements))
		for _, el := range raw.Elements {
			summaries = append(summaries, toSkillSummary(el))
		}
		return printSkillSummaries(cmd, summaries)
	}
}

func makeRunSkillsEndorse(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		skillURN, _ := cmd.Flags().GetString("urn")
		skillID, _ := cmd.Flags().GetString("skill-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would endorse skill %s (urn: %s)", skillID, skillURN), map[string]any{
				"action":    "endorse",
				"skill_id":  skillID,
				"skill_urn": skillURN,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/identity/normEntities/" + url.PathEscape(skillURN) + "/endorse"
		resp, err := client.PostJSON(ctx, path, map[string]any{})
		if err != nil {
			return fmt.Errorf("endorsing skill %s: %w", skillURN, err)
		}
		resp.Body.Close()

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]any{"endorsed": true, "skill_id": skillID, "skill_urn": skillURN})
		}
		fmt.Printf("Endorsed skill %s\n", skillID)
		return nil
	}
}

func makeRunSkillsEndorsements(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		username, _ := cmd.Flags().GetString("username")
		skillID, _ := cmd.Flags().GetString("skill-id")
		limit, _ := cmd.Flags().GetInt("limit")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		// If no username provided, fetch the current user's public identifier.
		if username == "" {
			meResp, err := client.Get(ctx, "/voyager/api/me", nil)
			if err != nil {
				return fmt.Errorf("getting current user: %w", err)
			}
			var me struct {
				MiniProfile struct {
					PublicIdentifier string `json:"publicIdentifier"`
				} `json:"miniProfile"`
			}
			if err := client.DecodeJSON(meResp, &me); err != nil {
				return fmt.Errorf("decoding current user: %w", err)
			}
			username = me.MiniProfile.PublicIdentifier
		}

		path := "/voyager/api/identity/profiles/" + url.PathEscape(username) + "/skillEndorsements/" + url.PathEscape(skillID)
		params := url.Values{"count": {fmt.Sprintf("%d", limit)}}
		resp, err := client.Get(ctx, path, params)
		if err != nil {
			return fmt.Errorf("listing endorsements for skill %s: %w", skillID, err)
		}

		var raw voyagerSkillEndorsementsResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding skill endorsements: %w", err)
		}

		type endorserEntry struct {
			URN      string `json:"urn"`
			PublicID string `json:"public_id"`
			Name     string `json:"name"`
		}
		endorsers := make([]endorserEntry, 0, len(raw.Elements))
		for _, el := range raw.Elements {
			endorsers = append(endorsers, endorserEntry{
				URN:      el.EntityURN,
				PublicID: el.Endorser.MiniProfile.PublicIdentifier,
				Name:     el.Endorser.MiniProfile.FirstName + " " + el.Endorser.MiniProfile.LastName,
			})
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(endorsers)
		}
		if len(endorsers) == 0 {
			fmt.Println("No endorsements found.")
			return nil
		}
		lines := make([]string, 0, len(endorsers)+1)
		lines = append(lines, fmt.Sprintf("%-25s  %-35s", "PUBLIC ID", "NAME"))
		for _, e := range endorsers {
			lines = append(lines, fmt.Sprintf("%-25s  %-35s",
				truncate(e.PublicID, 25),
				truncate(e.Name, 35),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}

// printSkillSummaries outputs skill summaries as JSON or text.
func printSkillSummaries(cmd *cobra.Command, skills []SkillSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(skills)
	}
	if len(skills) == 0 {
		fmt.Println("No skills found.")
		return nil
	}
	lines := make([]string, 0, len(skills)+1)
	lines = append(lines, fmt.Sprintf("%-15s  %-35s  %-12s", "ID", "NAME", "ENDORSEMENTS"))
	for _, s := range skills {
		lines = append(lines, fmt.Sprintf("%-15s  %-35s  %-12s",
			truncate(s.ID, 15),
			truncate(s.Name, 35),
			formatCount(s.EndorsementCount),
		))
	}
	cli.PrintText(lines)
	return nil
}
