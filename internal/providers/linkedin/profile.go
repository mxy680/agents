package linkedin

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// miniProfileEntity represents a MiniProfile entity found in the included array
// of a normalized /voyager/api/me response.
type miniProfileEntity struct {
	EntityURN        string `json:"entityUrn"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	Occupation       string `json:"occupation"`
	PublicIdentifier string `json:"publicIdentifier"`
}

// dashProfileEntity represents a Profile entity found in the included array
// of a normalized /voyager/api/identity/dash/profiles response.
type dashProfileEntity struct {
	EntityURN        string `json:"entityUrn"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	Headline         string `json:"headline"`
	Summary          string `json:"summary"`
	GeoLocationName  string `json:"geoLocationName"`
	PublicIdentifier string `json:"publicIdentifier"`
}

// newProfileCmd builds the "profile" subcommand group.
func newProfileCmd(factory ClientFactory) *cobra.Command {
	profileCmd := &cobra.Command{
		Use:     "profile",
		Short:   "View LinkedIn profiles",
		Aliases: []string{"prof"},
	}
	profileCmd.AddCommand(newProfileGetCmd(factory))
	profileCmd.AddCommand(newProfileMeCmd(factory))
	return profileCmd
}

// newProfileGetCmd builds the "profile get" command.
func newProfileGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a LinkedIn profile by public ID",
		Long:  "Retrieve a LinkedIn profile by its public identifier (the slug in the profile URL).",
		RunE:  makeRunProfileGet(factory),
	}
	cmd.Flags().String("public-id", "", "LinkedIn public profile ID (URL slug)")
	return cmd
}

// newProfileMeCmd builds the "profile me" command.
func newProfileMeCmd(factory ClientFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "me",
		Short: "Get your own LinkedIn profile",
		Long:  "Retrieve the authenticated user's own LinkedIn profile via /voyager/api/me.",
		RunE:  makeRunProfileMe(factory),
	}
}

func makeRunProfileGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		publicID, _ := cmd.Flags().GetString("public-id")
		if publicID == "" {
			return fmt.Errorf("--public-id is required")
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		params := url.Values{
			"q":              {"memberIdentity"},
			"memberIdentity": {publicID},
		}
		resp, err := client.Get(ctx, "/voyager/api/identity/dash/profiles", params)
		if err != nil {
			return fmt.Errorf("getting profile %s: %w", publicID, err)
		}

		var normalized NormalizedResponse
		if err := client.DecodeJSON(resp, &normalized); err != nil {
			return fmt.Errorf("decoding profile: %w", err)
		}

		raw := FindIncluded(normalized.Included, "Profile")
		if raw == nil {
			return fmt.Errorf("profile not found for %s", publicID)
		}

		var entity dashProfileEntity
		if err := json.Unmarshal(raw, &entity); err != nil {
			return fmt.Errorf("parsing profile entity: %w", err)
		}

		detail := ProfileDetail{
			URN:       entity.EntityURN,
			PublicID:  entity.PublicIdentifier,
			FirstName: entity.FirstName,
			LastName:  entity.LastName,
			Headline:  entity.Headline,
			Summary:   entity.Summary,
			Location:  entity.GeoLocationName,
		}
		if detail.PublicID == "" {
			detail.PublicID = publicID
		}
		return printProfileDetail(cmd, detail)
	}
}

func makeRunProfileMe(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := client.Get(ctx, "/voyager/api/me", nil)
		if err != nil {
			return fmt.Errorf("getting own profile: %w", err)
		}

		var normalized NormalizedResponse
		if err := client.DecodeJSON(resp, &normalized); err != nil {
			return fmt.Errorf("decoding own profile: %w", err)
		}

		raw := FindIncluded(normalized.Included, "MiniProfile")
		if raw == nil {
			return fmt.Errorf("miniProfile not found in /me response")
		}

		var mp miniProfileEntity
		if err := json.Unmarshal(raw, &mp); err != nil {
			return fmt.Errorf("parsing miniProfile entity: %w", err)
		}

		detail := ProfileDetail{
			URN:      mp.EntityURN,
			PublicID: mp.PublicIdentifier,
			FirstName: mp.FirstName,
			LastName:  mp.LastName,
			Headline:  mp.Occupation,
		}
		return printProfileDetail(cmd, detail)
	}
}

// printProfileDetail outputs a ProfileDetail as JSON or a formatted text block.
func printProfileDetail(cmd *cobra.Command, detail ProfileDetail) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(detail)
	}

	lines := []string{
		fmt.Sprintf("URN:         %s", detail.URN),
		fmt.Sprintf("Public ID:   %s", detail.PublicID),
		fmt.Sprintf("Name:        %s %s", detail.FirstName, detail.LastName),
		fmt.Sprintf("Headline:    %s", detail.Headline),
		fmt.Sprintf("Location:    %s", detail.Location),
		fmt.Sprintf("Industry:    %s", detail.Industry),
		fmt.Sprintf("Connections: %s", formatCount(detail.ConnectionCount)),
		fmt.Sprintf("Followers:   %s", formatCount(detail.FollowerCount)),
	}
	if detail.Summary != "" {
		lines = append(lines, fmt.Sprintf("Summary:     %s", truncate(detail.Summary, 200)))
	}
	if detail.ProfilePicURL != "" {
		lines = append(lines, fmt.Sprintf("Picture:     %s", detail.ProfilePicURL))
	}
	cli.PrintText(lines)
	return nil
}
