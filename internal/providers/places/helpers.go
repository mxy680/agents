package places

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// Entry represents a scraped Google Maps place with all available fields.
// Matches the JSON output of gosom/google-maps-scraper.
type Entry struct {
	Title            string              `json:"title"`
	Category         string              `json:"category"`
	Categories       []string            `json:"categories,omitempty"`
	Address          string              `json:"address"`
	Phone            string              `json:"phone,omitempty"`
	Website          string              `json:"web_site,omitempty"`
	Link             string              `json:"link,omitempty"`
	CID              string              `json:"cid,omitempty"`
	PlaceID          string              `json:"place_id,omitempty"`
	Latitude         float64             `json:"latitude,omitempty"`
	Longitude        float64             `json:"longtitude,omitempty"` // note: typo matches scraper output
	Rating           float64             `json:"review_rating,omitempty"`
	ReviewCount      int                 `json:"review_count,omitempty"`
	ReviewsPerRating map[int]int         `json:"reviews_per_rating,omitempty"`
	PriceRange       string              `json:"price_range,omitempty"`
	Status           string              `json:"status,omitempty"`
	Description      string              `json:"description,omitempty"`
	OpenHours        map[string][]string `json:"open_hours,omitempty"`
	PopularTimes     map[string]any      `json:"popular_times,omitempty"`
	PlusCode         string              `json:"plus_code,omitempty"`
	ReviewsLink      string              `json:"reviews_link,omitempty"`
	Thumbnail        string              `json:"thumbnail,omitempty"`
	Timezone         string              `json:"timezone,omitempty"`
	DataID           string              `json:"data_id,omitempty"`
	Images           []Image             `json:"images,omitempty"`
	Reservations     []LinkSource        `json:"reservations,omitempty"`
	OrderOnline      []LinkSource        `json:"order_online,omitempty"`
	Menu             *LinkSource         `json:"menu,omitempty"`
	Owner            *Owner              `json:"owner,omitempty"`
	CompleteAddress  *CompleteAddress     `json:"complete_address,omitempty"`
	About            []AboutSection      `json:"about,omitempty"`
	UserReviews      []Review            `json:"user_reviews,omitempty"`
	Emails           []string            `json:"emails,omitempty"`
	InputID          string              `json:"input_id,omitempty"`
}

// Image represents a place image.
type Image struct {
	Title string `json:"title,omitempty"`
	Image string `json:"image,omitempty"`
}

// LinkSource represents a link with its source attribution.
type LinkSource struct {
	Link   string `json:"link,omitempty"`
	Source string `json:"source,omitempty"`
}

// Owner represents the business owner.
type Owner struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
	Link string `json:"link,omitempty"`
}

// CompleteAddress is a structured address.
type CompleteAddress struct {
	Borough    string `json:"borough,omitempty"`
	Street     string `json:"street,omitempty"`
	City       string `json:"city,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	State      string `json:"state,omitempty"`
	Country    string `json:"country,omitempty"`
}

// AboutSection represents an "about" info section.
type AboutSection struct {
	ID      string       `json:"id,omitempty"`
	Name    string       `json:"name,omitempty"`
	Options []AboutOption `json:"options,omitempty"`
}

// AboutOption is a single about option (e.g., "Outdoor seating: Yes").
type AboutOption struct {
	Name    string `json:"name,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
}

// Review is a user review.
type Review struct {
	Name           string   `json:"name,omitempty"`
	ProfilePicture string   `json:"profile_picture,omitempty"`
	Rating         int      `json:"rating,omitempty"`
	Description    string   `json:"description,omitempty"`
	Images         []string `json:"images,omitempty"`
	When           string   `json:"when,omitempty"`
}

// PlaceSummary is a simplified view for text table output.
type PlaceSummary struct {
	Title       string  `json:"title"`
	Address     string  `json:"address"`
	Phone       string  `json:"phone,omitempty"`
	Rating      float64 `json:"rating,omitempty"`
	ReviewCount int     `json:"reviewCount,omitempty"`
	Website     string  `json:"website,omitempty"`
	PriceRange  string  `json:"priceRange,omitempty"`
	Status      string  `json:"status,omitempty"`
}

// toPlaceSummary converts an Entry to a PlaceSummary.
func toPlaceSummary(e Entry) PlaceSummary {
	return PlaceSummary{
		Title:       e.Title,
		Address:     e.Address,
		Phone:       e.Phone,
		Rating:      e.Rating,
		ReviewCount: e.ReviewCount,
		Website:     e.Website,
		PriceRange:  e.PriceRange,
		Status:      e.Status,
	}
}

// truncate shortens s to at most max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// parseLatLng parses a "lat,lng" string into latitude and longitude.
func parseLatLng(s string) (float64, float64, error) {
	var lat, lng float64
	parts := strings.SplitN(s, ",", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid lat,lng format: %q (expected lat,lng)", s)
	}
	if _, err := fmt.Sscanf(parts[0], "%f", &lat); err != nil {
		return 0, 0, fmt.Errorf("invalid latitude: %w", err)
	}
	if _, err := fmt.Sscanf(parts[1], "%f", &lng); err != nil {
		return 0, 0, fmt.Errorf("invalid longitude: %w", err)
	}
	return lat, lng, nil
}

// printPlaceSummaries outputs place summaries as JSON or a formatted text table.
func printPlaceSummaries(cmd *cobra.Command, summaries []PlaceSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(summaries)
	}

	if len(summaries) == 0 {
		fmt.Println("No places found.")
		return nil
	}

	lines := make([]string, 0, len(summaries)+1)
	lines = append(lines, fmt.Sprintf("%-30s  %-35s  %-15s  %-6s  %-5s  %-5s",
		"NAME", "ADDRESS", "PHONE", "RATING", "REVS", "PRICE"))
	for _, s := range summaries {
		name := truncate(s.Title, 30)
		addr := truncate(s.Address, 35)
		phone := truncate(s.Phone, 15)
		rating := "-"
		if s.Rating > 0 {
			rating = fmt.Sprintf("%.1f", s.Rating)
		}
		reviews := "-"
		if s.ReviewCount > 0 {
			reviews = fmt.Sprintf("%d", s.ReviewCount)
		}
		lines = append(lines, fmt.Sprintf("%-30s  %-35s  %-15s  %-6s  %-5s  %-5s",
			name, addr, phone, rating, reviews, s.PriceRange))
	}
	cli.PrintText(lines)
	return nil
}

// printEntryDetail outputs a single entry in detailed text format.
func printEntryDetail(entry Entry) {
	lines := []string{
		fmt.Sprintf("Name:       %s", entry.Title),
		fmt.Sprintf("Category:   %s", entry.Category),
		fmt.Sprintf("Address:    %s", entry.Address),
	}
	if entry.Phone != "" {
		lines = append(lines, fmt.Sprintf("Phone:      %s", entry.Phone))
	}
	if entry.Website != "" {
		lines = append(lines, fmt.Sprintf("Website:    %s", entry.Website))
	}
	if entry.Rating > 0 {
		lines = append(lines, fmt.Sprintf("Rating:     %.1f (%d reviews)", entry.Rating, entry.ReviewCount))
	}
	if entry.PriceRange != "" {
		lines = append(lines, fmt.Sprintf("Price:      %s", entry.PriceRange))
	}
	if entry.Status != "" {
		lines = append(lines, fmt.Sprintf("Status:     %s", entry.Status))
	}
	if entry.Link != "" {
		lines = append(lines, fmt.Sprintf("Maps:       %s", entry.Link))
	}
	if entry.Description != "" {
		lines = append(lines, fmt.Sprintf("Description: %s", entry.Description))
	}
	if entry.Latitude != 0 || entry.Longitude != 0 {
		lines = append(lines, fmt.Sprintf("Location:   %.6f, %.6f", entry.Latitude, entry.Longitude))
	}
	if len(entry.Emails) > 0 {
		lines = append(lines, fmt.Sprintf("Emails:     %s", strings.Join(entry.Emails, ", ")))
	}

	// Hours
	if len(entry.OpenHours) > 0 {
		lines = append(lines, "\nHours:")
		dayOrder := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
		for _, day := range dayOrder {
			if hours, ok := entry.OpenHours[day]; ok {
				lines = append(lines, fmt.Sprintf("  %-10s %s", day+":", strings.Join(hours, ", ")))
			}
		}
	}

	// Reviews
	if len(entry.UserReviews) > 0 {
		lines = append(lines, fmt.Sprintf("\nReviews (%d):", len(entry.UserReviews)))
		for _, r := range entry.UserReviews {
			header := fmt.Sprintf("  %d★ by %s", r.Rating, r.Name)
			if r.When != "" {
				header += " (" + r.When + ")"
			}
			lines = append(lines, header)
			if r.Description != "" {
				lines = append(lines, fmt.Sprintf("    %s", truncate(r.Description, 120)))
			}
		}
	}

	// Images
	if len(entry.Images) > 0 {
		lines = append(lines, fmt.Sprintf("\nImages (%d):", len(entry.Images)))
		for _, img := range entry.Images {
			if img.Title != "" {
				lines = append(lines, fmt.Sprintf("  %s: %s", img.Title, truncate(img.Image, 80)))
			} else {
				lines = append(lines, fmt.Sprintf("  %s", truncate(img.Image, 80)))
			}
		}
	}

	fmt.Println(strings.Join(lines, "\n"))
}
