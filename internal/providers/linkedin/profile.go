package linkedin

import (
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerProfileResponse is the response envelope for GET /voyager/api/identity/profiles/{publicId}.
type voyagerProfileResponse struct {
	Profile struct {
		EntityURN      string `json:"entityUrn"`
		FirstName      string `json:"firstName"`
		LastName       string `json:"lastName"`
		Headline       string `json:"headline"`
		Summary        string `json:"summary"`
		LocationName   string `json:"locationName"`
		IndustryName   string `json:"industryName"`
		ProfilePicture *struct {
			DisplayImageReference struct {
				VectorImage struct {
					RootURL   string `json:"rootUrl"`
					Artifacts []struct {
						FileIdentifyingURLPathSegment string `json:"fileIdentifyingUrlPathSegment"`
					} `json:"artifacts"`
				} `json:"vectorImage"`
			} `json:"displayImageReference"`
		} `json:"profilePicture"`
	} `json:"profile"`
	ConnectionCount int `json:"connectionCount"`
	FollowerCount   int `json:"followerCount"`
}

// voyagerMeResponse is the response envelope for GET /voyager/api/me.
type voyagerMeResponse struct {
	MiniProfile struct {
		EntityURN       string `json:"entityUrn"`
		FirstName       string `json:"firstName"`
		LastName        string `json:"lastName"`
		Occupation      string `json:"occupation"`
		PublicIdentifier string `json:"publicIdentifier"`
	} `json:"miniProfile"`
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

		path := "/voyager/api/identity/profiles/" + url.PathEscape(publicID)
		resp, err := client.Get(ctx, path, nil)
		if err != nil {
			return fmt.Errorf("getting profile %s: %w", publicID, err)
		}

		var raw voyagerProfileResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding profile: %w", err)
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
			PublicID:        publicID,
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

		var raw voyagerMeResponse
		if err := client.DecodeJSON(resp, &raw); err != nil {
			return fmt.Errorf("decoding own profile: %w", err)
		}

		mp := raw.MiniProfile
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
