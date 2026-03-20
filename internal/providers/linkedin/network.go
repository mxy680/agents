package linkedin

import (
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerFollowersResponse is the response envelope for followers/following list calls.
type voyagerFollowersResponse struct {
	Elements []struct {
		EntityURN  string `json:"entityUrn"`
		FirstName  string `json:"firstName"`
		LastName   string `json:"lastName"`
		Occupation string `json:"occupation"`
		PublicIdentifier string `json:"publicIdentifier"`
		// Some shapes wrap the mini profile
		MiniProfile struct {
			EntityURN        string `json:"entityUrn"`
			FirstName        string `json:"firstName"`
			LastName         string `json:"lastName"`
			Occupation       string `json:"occupation"`
			PublicIdentifier string `json:"publicIdentifier"`
		} `json:"miniProfile"`
	} `json:"elements"`
	Paging struct {
		Start int `json:"start"`
		Count int `json:"count"`
		Total int `json:"total"`
	} `json:"paging"`
}

// voyagerSuggestionsResponse is the response envelope for connection suggestions.
type voyagerSuggestionsResponse struct {
	Elements []struct {
		MemberDistance struct{} `json:"memberDistance"`
		// The suggested member's mini profile
		SubText struct{ Text string `json:"text"` } `json:"subText"`
		Title   struct{ Text string `json:"text"` } `json:"title"`
		EntityURN string `json:"entityUrn"`
		// Some shapes have a connectedMemberResolved or similar
		EntityResult struct {
			EntityURN        string `json:"entityUrn"`
			FirstName        string `json:"firstName"`
			LastName         string `json:"lastName"`
			Occupation       string `json:"occupation"`
			PublicIdentifier string `json:"publicIdentifier"`
		} `json:"entityResult"`
	} `json:"elements"`
	Paging struct {
		Start int `json:"start"`
		Count int `json:"count"`
		Total int `json:"total"`
	} `json:"paging"`
}

// newNetworkCmd builds the "network" subcommand group.
func newNetworkCmd(factory ClientFactory) *cobra.Command {
	networkCmd := &cobra.Command{
		Use:   "network",
		Short: "Manage your LinkedIn network (followers, following, suggestions)",
	}
	networkCmd.AddCommand(newNetworkFollowersCmd(factory))
	networkCmd.AddCommand(newNetworkFollowingCmd(factory))
	networkCmd.AddCommand(newNetworkFollowCmd(factory))
	networkCmd.AddCommand(newNetworkUnfollowCmd(factory))
	networkCmd.AddCommand(newNetworkSuggestionsCmd(factory))
	return networkCmd
}

// newNetworkFollowersCmd builds the "network followers" command.
func newNetworkFollowersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "followers",
		Short: "List your LinkedIn followers",
		Long:  "List members who follow you on LinkedIn.",
		RunE:  makeRunNetworkFollowers(factory),
	}
	cmd.Flags().Int("limit", 10, "Maximum number of followers to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	return cmd
}

// newNetworkFollowingCmd builds the "network following" command.
func newNetworkFollowingCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "following",
		Short: "List members and entities you follow on LinkedIn",
		RunE:  makeRunNetworkFollowing(factory),
	}
	cmd.Flags().Int("limit", 10, "Maximum number of results to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	return cmd
}

// newNetworkFollowCmd builds the "network follow" command.
func newNetworkFollowCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "follow",
		Short: "Follow a member or entity by URN",
		RunE:  makeRunNetworkFollow(factory),
	}
	cmd.Flags().String("urn", "", "URN of the member or entity to follow (required)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without making changes")
	return cmd
}

// newNetworkUnfollowCmd builds the "network unfollow" command.
func newNetworkUnfollowCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfollow",
		Short: "Unfollow a member or entity by URN",
		RunE:  makeRunNetworkUnfollow(factory),
	}
	cmd.Flags().String("urn", "", "URN of the member or entity to unfollow (required)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without making changes")
	return cmd
}

// newNetworkSuggestionsCmd builds the "network suggestions" command.
func newNetworkSuggestionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suggestions",
		Short: "List people you may know on LinkedIn",
		RunE:  makeRunNetworkSuggestions(factory),
	}
	cmd.Flags().Int("limit", 10, "Maximum number of suggestions to return")
	return cmd
}

func makeRunNetworkFollowers(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

func makeRunNetworkFollowing(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

func makeRunNetworkFollow(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		urn, _ := cmd.Flags().GetString("urn")
		if urn == "" {
			return fmt.Errorf("--urn is required")
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("Follow: %s", urn), map[string]string{"urn": urn})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{"urn": urn}
		resp, err := client.PostJSON(ctx, "/voyager/api/feed/follows", body)
		if err != nil {
			return fmt.Errorf("following %s: %w", urn, err)
		}

		if err := client.DecodeJSON(resp, &struct{}{}); err != nil {
			return fmt.Errorf("following %s: %w", urn, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "following", "urn": urn})
		}
		fmt.Printf("Now following: %s\n", urn)
		return nil
	}
}

func makeRunNetworkUnfollow(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		urn, _ := cmd.Flags().GetString("urn")
		if urn == "" {
			return fmt.Errorf("--urn is required")
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("Unfollow: %s", urn), map[string]string{"urn": urn})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/feed/follows/" + url.PathEscape(urn)
		resp, err := client.Delete(ctx, path)
		if err != nil {
			return fmt.Errorf("unfollowing %s: %w", urn, err)
		}
		resp.Body.Close()

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "unfollowed", "urn": urn})
		}
		fmt.Printf("Unfollowed: %s\n", urn)
		return nil
	}
}

func makeRunNetworkSuggestions(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

// followersResponseToProfiles maps a voyagerFollowersResponse to []ProfileSummary.
// It handles both flat miniProfile shapes and direct fields.
func followersResponseToProfiles(raw voyagerFollowersResponse) []ProfileSummary {
	summaries := make([]ProfileSummary, 0, len(raw.Elements))
	for _, el := range raw.Elements {
		// Prefer nested miniProfile if present
		if el.MiniProfile.EntityURN != "" {
			summaries = append(summaries, ProfileSummary{
				URN:       el.MiniProfile.EntityURN,
				PublicID:  el.MiniProfile.PublicIdentifier,
				FirstName: el.MiniProfile.FirstName,
				LastName:  el.MiniProfile.LastName,
				Headline:  el.MiniProfile.Occupation,
			})
			continue
		}
		summaries = append(summaries, ProfileSummary{
			URN:       el.EntityURN,
			PublicID:  el.PublicIdentifier,
			FirstName: el.FirstName,
			LastName:  el.LastName,
			Headline:  el.Occupation,
		})
	}
	return summaries
}
