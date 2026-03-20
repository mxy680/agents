package linkedin

import (
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerConnectionsResponse is the response envelope for GET /voyager/api/relationships/dash/connections.
type voyagerConnectionsResponse struct {
	Elements []struct {
		ConnectedMember         string `json:"connectedMember"`
		ConnectedMemberResolved struct {
			EntityURN       string `json:"entityUrn"`
			FirstName       string `json:"firstName"`
			LastName        string `json:"lastName"`
			Occupation      string `json:"occupation"`
			PublicIdentifier string `json:"publicIdentifier"`
		} `json:"connectedMemberResolved"`
		CreatedAt int64 `json:"createdAt"`
	} `json:"elements"`
	Paging struct {
		Start int `json:"start"`
		Count int `json:"count"`
		Total int `json:"total"`
	} `json:"paging"`
}

// newConnectionsCmd builds the "connections" subcommand group.
func newConnectionsCmd(factory ClientFactory) *cobra.Command {
	connectionsCmd := &cobra.Command{
		Use:     "connections",
		Short:   "Manage LinkedIn connections",
		Aliases: []string{"conn"},
	}
	connectionsCmd.AddCommand(newConnectionsListCmd(factory))
	connectionsCmd.AddCommand(newConnectionsGetCmd(factory))
	connectionsCmd.AddCommand(newConnectionsRemoveCmd(factory))
	return connectionsCmd
}

// newConnectionsListCmd builds the "connections list" command.
func newConnectionsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your LinkedIn connections",
		Long:  "List your first-degree LinkedIn connections, sorted by recently added by default.",
		RunE:  makeRunConnectionsList(factory),
	}
	cmd.Flags().Int("limit", 10, "Maximum number of connections to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	cmd.Flags().String("sort", "RECENTLY_ADDED", "Sort order: RECENTLY_ADDED, LAST_NAME, or FIRST_NAME")
	return cmd
}

// newConnectionsGetCmd builds the "connections get" command.
func newConnectionsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a connection's profile by URN",
		Long:  "Retrieve a connection's LinkedIn profile using their member URN.",
		RunE:  makeRunConnectionsGet(factory),
	}
	cmd.Flags().String("urn", "", "Connection member URN (e.g. urn:li:fs_miniProfile:...)")
	return cmd
}

// newConnectionsRemoveCmd builds the "connections remove" command.
func newConnectionsRemoveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove a LinkedIn connection",
		Long:  "Remove a first-degree LinkedIn connection by their member URN.",
		RunE:  makeRunConnectionsRemove(factory),
	}
	cmd.Flags().String("urn", "", "Connection member URN (e.g. urn:li:fs_miniProfile:...)")
	cmd.Flags().Bool("confirm", false, "Confirm removal (required for destructive action)")
	cmd.Flags().Bool("dry-run", false, "Preview the action without making changes")
	return cmd
}

func makeRunConnectionsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")
		sort, _ := cmd.Flags().GetString("sort")

		start := 0
		if cursor != "" {
			if _, err := fmt.Sscanf(cursor, "%d", &start); err != nil {
				return fmt.Errorf("invalid cursor %q: must be a numeric start offset", cursor)
			}
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{
			"q":        {"search"},
			"start":    {fmt.Sprintf("%d", start)},
			"count":    {fmt.Sprintf("%d", limit)},
			"sortType": {sort},
		}
		resp, err := client.Get(ctx, "/voyager/api/relationships/dash/connections", params)
		if err != nil {
			return fmt.Errorf("listing connections: %w", err)
		}

		var raw voyagerConnectionsResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding connections: %w", err)
		}

		summaries := make([]ConnectionSummary, 0, len(raw.Elements))
		for _, el := range raw.Elements {
			summaries = append(summaries, ConnectionSummary{
				URN:       el.ConnectedMember,
				PublicID:  el.ConnectedMemberResolved.PublicIdentifier,
				FirstName: el.ConnectedMemberResolved.FirstName,
				LastName:  el.ConnectedMemberResolved.LastName,
				Headline:  el.ConnectedMemberResolved.Occupation,
				CreatedAt: el.CreatedAt,
			})
		}

		return printConnectionSummaries(cmd, summaries)
	}
}

func makeRunConnectionsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		urn, _ := cmd.Flags().GetString("urn")
		if urn == "" {
			return fmt.Errorf("--urn is required")
		}

		// Extract the public ID from the URN, or use the URN as a profile lookup.
		// For profile get we reuse the profile endpoint with the public identifier.
		// Since we have the URN, we call the profile endpoint by URN using the identity/profiles path.
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/identity/profiles/" + url.PathEscape(urn)
		resp, err := client.Get(ctx, path, nil)
		if err != nil {
			return fmt.Errorf("getting connection profile %s: %w", urn, err)
		}

		var raw voyagerProfileResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding connection profile: %w", err)
		}

		picURL := ""
		if raw.Profile.ProfilePicture != nil {
			vi := raw.Profile.ProfilePicture.DisplayImageReference.VectorImage
			if len(vi.Artifacts) > 0 {
				picURL = vi.RootURL + vi.Artifacts[0].FileIdentifyingURLPathSegment
			}
		}

		detail := ProfileDetail{
			URN:             raw.Profile.EntityURN,
			PublicID:        urn,
			FirstName:       raw.Profile.FirstName,
			LastName:        raw.Profile.LastName,
			Headline:        raw.Profile.Headline,
			Summary:         raw.Profile.Summary,
			Location:        raw.Profile.LocationName,
			Industry:        raw.Profile.IndustryName,
			ProfilePicURL:   picURL,
			ConnectionCount: raw.ConnectionCount,
			FollowerCount:   raw.FollowerCount,
		}
		return printProfileDetail(cmd, detail)
	}
}

func makeRunConnectionsRemove(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		urn, _ := cmd.Flags().GetString("urn")
		if urn == "" {
			return fmt.Errorf("--urn is required")
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("Remove connection: %s", urn), map[string]string{"urn": urn})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"action":    "delete",
			"memberUrn": urn,
		}
		resp, err := client.PostJSON(ctx, "/voyager/api/relationships/dash/connections", body)
		if err != nil {
			return fmt.Errorf("removing connection %s: %w", urn, err)
		}

		if err := client.DecodeJSON(resp, &struct{}{}); err != nil {
			return fmt.Errorf("removing connection: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "removed", "urn": urn})
		}
		fmt.Printf("Connection removed: %s\n", urn)
		return nil
	}
}

// printConnectionSummaries outputs connection summaries as JSON or text.
func printConnectionSummaries(cmd *cobra.Command, connections []ConnectionSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(connections)
	}
	if len(connections) == 0 {
		fmt.Println("No connections found.")
		return nil
	}
	lines := make([]string, 0, len(connections)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-20s  %-40s  %-16s", "FIRST NAME", "LAST NAME", "HEADLINE", "CONNECTED"))
	for _, c := range connections {
		lines = append(lines, fmt.Sprintf("%-20s  %-20s  %-40s  %-16s",
			truncate(c.FirstName, 20),
			truncate(c.LastName, 20),
			truncate(c.Headline, 40),
			formatTimestamp(c.CreatedAt),
		))
	}
	cli.PrintText(lines)
	return nil
}
