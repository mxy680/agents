package places

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	"google.golang.org/api/googleapi"
)

func newGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get detailed information about a place",
		Long:  "Retrieve full details for a place by its place ID.",
		RunE:  makeRunGet(factory),
	}
	cmd.Flags().String("place-id", "", "Place ID (required)")
	cmd.Flags().String("lang", "", "Language code (e.g. en)")
	cmd.Flags().String("region", "", "CLDR region code (e.g. us)")
	cmd.Flags().String("fields", "advanced", "Field tier: basic, advanced, preferred, or all")
	cmd.Flags().String("session-token", "", "Session token for billing (UUID from autocomplete)")
	_ = cmd.MarkFlagRequired("place-id")
	return cmd
}

func makeRunGet(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create service: %w", err)
		}

		placeID, _ := cmd.Flags().GetString("place-id")
		name := "places/" + placeID

		tier, _ := cmd.Flags().GetString("fields")
		call := svc.Places.Get(name).Fields(googleapi.Field(detailFieldMask(tier)))

		if lang, _ := cmd.Flags().GetString("lang"); lang != "" {
			call = call.LanguageCode(lang)
		}
		if region, _ := cmd.Flags().GetString("region"); region != "" {
			call = call.RegionCode(region)
		}
		if st, _ := cmd.Flags().GetString("session-token"); st != "" {
			call = call.SessionToken(st)
		}

		place, err := call.Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("get place: %w", err)
		}

		detail := toPlaceDetail(place)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		// Text output
		lines := []string{
			fmt.Sprintf("Name:       %s", detail.Name),
			fmt.Sprintf("ID:         %s", detail.ID),
			fmt.Sprintf("Address:    %s", detail.Address),
		}
		if detail.PrimaryType != "" {
			lines = append(lines, fmt.Sprintf("Type:       %s", detail.PrimaryType))
		}
		if detail.Rating > 0 {
			lines = append(lines, fmt.Sprintf("Rating:     %.1f (%d reviews)", detail.Rating, detail.UserRatingCount))
		}
		if detail.PriceLevel != "" {
			lines = append(lines, fmt.Sprintf("Price:      %s", detail.PriceLevel))
		}
		lines = append(lines, fmt.Sprintf("Status:     %s", detail.BusinessStatus))
		if detail.OpenNow != nil {
			if *detail.OpenNow {
				lines = append(lines, "Open Now:   Yes")
			} else {
				lines = append(lines, "Open Now:   No")
			}
		}
		if detail.PhoneNumber != "" {
			lines = append(lines, fmt.Sprintf("Phone:      %s", detail.PhoneNumber))
		}
		if detail.WebsiteURI != "" {
			lines = append(lines, fmt.Sprintf("Website:    %s", detail.WebsiteURI))
		}
		if detail.GoogleMapsURI != "" {
			lines = append(lines, fmt.Sprintf("Maps:       %s", detail.GoogleMapsURI))
		}
		if detail.EditorialSummary != "" {
			lines = append(lines, fmt.Sprintf("Summary:    %s", detail.EditorialSummary))
		}
		if detail.Location != nil {
			lines = append(lines, fmt.Sprintf("Location:   %.6f, %.6f", detail.Location.Latitude, detail.Location.Longitude))
		}

		// Services
		services := []string{}
		if detail.Delivery {
			services = append(services, "Delivery")
		}
		if detail.DineIn {
			services = append(services, "Dine-in")
		}
		if detail.Takeout {
			services = append(services, "Takeout")
		}
		if detail.CurbsidePickup {
			services = append(services, "Curbside")
		}
		if detail.Reservable {
			services = append(services, "Reservable")
		}
		if len(services) > 0 {
			lines = append(lines, fmt.Sprintf("Services:   %s", strings.Join(services, ", ")))
		}

		// Hours
		if len(detail.WeekdayHours) > 0 {
			lines = append(lines, "\nHours:")
			for _, h := range detail.WeekdayHours {
				lines = append(lines, fmt.Sprintf("  %s", h))
			}
		}

		// Reviews
		if len(detail.Reviews) > 0 {
			lines = append(lines, fmt.Sprintf("\nReviews (%d):", len(detail.Reviews)))
			for _, r := range detail.Reviews {
				header := fmt.Sprintf("  %.0f★ by %s", r.Rating, r.Author)
				if r.RelativeTime != "" {
					header += " (" + r.RelativeTime + ")"
				}
				lines = append(lines, header)
				if r.Text != "" {
					lines = append(lines, fmt.Sprintf("    %s", truncate(r.Text, 120)))
				}
			}
		}

		// Photos
		if len(detail.Photos) > 0 {
			lines = append(lines, fmt.Sprintf("\nPhotos (%d):", len(detail.Photos)))
			for _, ph := range detail.Photos {
				lines = append(lines, fmt.Sprintf("  %s (%dx%d)", ph.Name, ph.WidthPx, ph.HeightPx))
			}
		}

		cli.PrintText(lines)
		return nil
	}
}
