package zillow

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// PropertySummary is a simplified property view for list output.
type PropertySummary struct {
	ZPID        string  `json:"zpid"`
	Address     string  `json:"address"`
	Price       int64   `json:"price,omitempty"`
	Beds        int     `json:"beds,omitempty"`
	Baths       float64 `json:"baths,omitempty"`
	Sqft        int     `json:"sqft,omitempty"`
	HomeType    string  `json:"homeType,omitempty"`
	Status      string  `json:"status,omitempty"`
	ZillowURL   string  `json:"zillowUrl,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	DaysOnMarket int    `json:"daysOnMarket,omitempty"`
}

// PropertyDetail is a full property details view.
type PropertyDetail struct {
	ZPID             string          `json:"zpid"`
	Address          string          `json:"address"`
	StreetAddress    string          `json:"streetAddress,omitempty"`
	City             string          `json:"city,omitempty"`
	State            string          `json:"state,omitempty"`
	Zipcode          string          `json:"zipcode,omitempty"`
	Price            int64           `json:"price,omitempty"`
	Beds             int             `json:"beds,omitempty"`
	Baths            float64         `json:"baths,omitempty"`
	Sqft             int             `json:"sqft,omitempty"`
	LotSize          int             `json:"lotSize,omitempty"`
	YearBuilt        int             `json:"yearBuilt,omitempty"`
	HomeType         string          `json:"homeType,omitempty"`
	Status           string          `json:"status,omitempty"`
	Description      string          `json:"description,omitempty"`
	Zestimate        int64           `json:"zestimate,omitempty"`
	RentZestimate    int64           `json:"rentZestimate,omitempty"`
	Latitude         float64         `json:"latitude,omitempty"`
	Longitude        float64         `json:"longitude,omitempty"`
	ZillowURL        string          `json:"zillowUrl,omitempty"`
	Photos           []string        `json:"photos,omitempty"`
	PriceHistory     []PriceEvent    `json:"priceHistory,omitempty"`
	TaxHistory       []TaxRecord     `json:"taxHistory,omitempty"`
	Schools          []SchoolSummary `json:"schools,omitempty"`
	DaysOnMarket     int             `json:"daysOnMarket,omitempty"`
	MonthlyHOA       int             `json:"monthlyHoa,omitempty"`
	ParkingSpaces    int             `json:"parkingSpaces,omitempty"`
	HeatingType      string          `json:"heatingType,omitempty"`
	CoolingType      string          `json:"coolingType,omitempty"`
	ListingAgent     string          `json:"listingAgent,omitempty"`
	ListingBrokerage string          `json:"listingBrokerage,omitempty"`
}

// PriceEvent represents a single price history entry.
type PriceEvent struct {
	Date    string `json:"date"`
	Event   string `json:"event"`
	Price   int64  `json:"price,omitempty"`
	Source  string `json:"source,omitempty"`
}

// TaxRecord represents a single tax history entry.
type TaxRecord struct {
	Year        int   `json:"year"`
	TaxPaid     int64 `json:"taxPaid,omitempty"`
	TaxAssessed int64 `json:"taxAssessed,omitempty"`
}

// ZestimateSummary holds Zestimate data for a property.
type ZestimateSummary struct {
	ZPID          string `json:"zpid"`
	Address       string `json:"address,omitempty"`
	Zestimate     int64  `json:"zestimate,omitempty"`
	RentZestimate int64  `json:"rentZestimate,omitempty"`
	ValueLow      int64  `json:"valueLow,omitempty"`
	ValueHigh     int64  `json:"valueHigh,omitempty"`
	ValueChange   int64  `json:"valueChange,omitempty"`
}

// ZestimateChartPoint is a single data point in a Zestimate chart.
type ZestimateChartPoint struct {
	Date  string `json:"date"`
	Value int64  `json:"value"`
}

// AgentSummary is a simplified agent view for list output.
type AgentSummary struct {
	AgentID      string  `json:"agentId"`
	Name         string  `json:"name"`
	Phone        string  `json:"phone,omitempty"`
	Rating       float64 `json:"rating,omitempty"`
	ReviewCount  int     `json:"reviewCount,omitempty"`
	RecentSales  int     `json:"recentSales,omitempty"`
	ProfileURL   string  `json:"profileUrl,omitempty"`
	Photo        string  `json:"photo,omitempty"`
	Specialties  string  `json:"specialties,omitempty"`
}

// AgentReview is a single agent review.
type AgentReview struct {
	Rating      float64 `json:"rating"`
	Date        string  `json:"date,omitempty"`
	Description string  `json:"description,omitempty"`
	Reviewer    string  `json:"reviewer,omitempty"`
}

// MortgageRate holds mortgage rate data.
type MortgageRate struct {
	Program     string  `json:"program"`
	Rate        float64 `json:"rate"`
	APR         float64 `json:"apr,omitempty"`
	LoanType    string  `json:"loanType,omitempty"`
	State       string  `json:"state,omitempty"`
	Date        string  `json:"date,omitempty"`
}

// MortgageCalculation holds mortgage payment calculation results.
type MortgageCalculation struct {
	MonthlyPayment  float64 `json:"monthlyPayment"`
	Principal       float64 `json:"principal"`
	Interest        float64 `json:"interest"`
	Tax             float64 `json:"tax,omitempty"`
	Insurance       float64 `json:"insurance,omitempty"`
	TotalCost       float64 `json:"totalCost"`
}

// LenderReview is a single lender review from Zillow's mortgage API.
type LenderReview struct {
	Rating                    float64 `json:"rating"`
	Title                     string  `json:"title,omitempty"`
	Content                   string  `json:"content,omitempty"`
	LoanType                  string  `json:"loanType,omitempty"`
	LoanProgram               string  `json:"loanProgram,omitempty"`
	ClosingCostsSatisfaction  float64 `json:"closingCostsSatisfaction,omitempty"`
	InterestRateSatisfaction  float64 `json:"interestRateSatisfaction,omitempty"`
	VerifiedReviewer          bool    `json:"verifiedReviewer,omitempty"`
}

// LenderInfo holds lender overview data returned by the reviews endpoint.
type LenderInfo struct {
	ProfileURL   string         `json:"profileUrl,omitempty"`
	ReviewURL    string         `json:"reviewUrl,omitempty"`
	TotalReviews int            `json:"totalReviews,omitempty"`
	Rating       float64        `json:"rating,omitempty"`
	Reviews      []LenderReview `json:"reviews,omitempty"`
}

// AutocompleteResult is a single autocomplete suggestion.
type AutocompleteResult struct {
	Display   string  `json:"display"`
	ZPID      string  `json:"zpid,omitempty"`
	Type      string  `json:"type,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	URL       string  `json:"url,omitempty"`
}

// WalkScoreResult holds walk/transit/bike scores.
type WalkScoreResult struct {
	ZPID         string `json:"zpid"`
	WalkScore    int    `json:"walkScore,omitempty"`
	TransitScore int    `json:"transitScore,omitempty"`
	BikeScore    int    `json:"bikeScore,omitempty"`
	WalkDesc     string `json:"walkDescription,omitempty"`
	TransitDesc  string `json:"transitDescription,omitempty"`
	BikeDesc     string `json:"bikeDescription,omitempty"`
}

// SchoolSummary is a simplified school view.
type SchoolSummary struct {
	Name     string  `json:"name"`
	Rating   int     `json:"rating,omitempty"`
	Level    string  `json:"level,omitempty"`
	Type     string  `json:"type,omitempty"`
	Grades   string  `json:"grades,omitempty"`
	Distance float64 `json:"distance,omitempty"`
	Link     string  `json:"link,omitempty"`
}

// NeighborhoodSummary is a simplified neighborhood view.
type NeighborhoodSummary struct {
	RegionID        string `json:"regionId"`
	Name            string `json:"name"`
	Type            string `json:"type,omitempty"`
	MedianHomeValue int64  `json:"medianHomeValue,omitempty"`
	MedianRent      int64  `json:"medianRent,omitempty"`
	ZillowURL       string `json:"zillowUrl,omitempty"`
}

// BuilderSummary is a simplified builder view.
type BuilderSummary struct {
	BuilderID   string  `json:"builderId"`
	Name        string  `json:"name"`
	Rating      float64 `json:"rating,omitempty"`
	ReviewCount int     `json:"reviewCount,omitempty"`
	URL         string  `json:"url,omitempty"`
}

// BuilderCommunity is a new construction community.
type BuilderCommunity struct {
	Name      string `json:"name"`
	Location  string `json:"location,omitempty"`
	PriceFrom int64  `json:"priceFrom,omitempty"`
	PriceTo   int64  `json:"priceTo,omitempty"`
	URL       string `json:"url,omitempty"`
}

// truncate shortens s to at most max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// formatPrice formats a price as "$1,234,567".
func formatPrice(price int64) string {
	if price == 0 {
		return "-"
	}
	s := fmt.Sprintf("%d", price)
	// Insert commas
	n := len(s)
	if n <= 3 {
		return "$" + s
	}
	var parts []string
	for n > 3 {
		parts = append([]string{s[n-3 : n]}, parts...)
		n -= 3
	}
	parts = append([]string{s[:n]}, parts...)
	return "$" + strings.Join(parts, ",")
}

// printPropertySummaries outputs property summaries as JSON or a formatted text table.
func printPropertySummaries(cmd *cobra.Command, summaries []PropertySummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(summaries)
	}

	if len(summaries) == 0 {
		fmt.Println("No properties found.")
		return nil
	}

	lines := make([]string, 0, len(summaries)+1)
	lines = append(lines, fmt.Sprintf("%-12s  %-40s  %-12s  %-4s  %-5s  %-8s  %-12s",
		"ZPID", "ADDRESS", "PRICE", "BEDS", "BATHS", "SQFT", "STATUS"))
	for _, s := range summaries {
		addr := truncate(s.Address, 40)
		price := formatPrice(s.Price)
		beds := "-"
		if s.Beds > 0 {
			beds = fmt.Sprintf("%d", s.Beds)
		}
		baths := "-"
		if s.Baths > 0 {
			baths = fmt.Sprintf("%.1f", s.Baths)
		}
		sqft := "-"
		if s.Sqft > 0 {
			sqft = fmt.Sprintf("%d", s.Sqft)
		}
		lines = append(lines, fmt.Sprintf("%-12s  %-40s  %-12s  %-4s  %-5s  %-8s  %-12s",
			s.ZPID, addr, price, beds, baths, sqft, s.Status))
	}
	cli.PrintText(lines)
	return nil
}

// printPropertyDetail outputs a single property in detailed text format.
func printPropertyDetail(detail PropertyDetail) {
	lines := []string{
		fmt.Sprintf("ZPID:        %s", detail.ZPID),
		fmt.Sprintf("Address:     %s", detail.Address),
		fmt.Sprintf("Price:       %s", formatPrice(detail.Price)),
	}
	if detail.Beds > 0 {
		lines = append(lines, fmt.Sprintf("Beds:        %d", detail.Beds))
	}
	if detail.Baths > 0 {
		lines = append(lines, fmt.Sprintf("Baths:       %.1f", detail.Baths))
	}
	if detail.Sqft > 0 {
		lines = append(lines, fmt.Sprintf("Sqft:        %d", detail.Sqft))
	}
	if detail.LotSize > 0 {
		lines = append(lines, fmt.Sprintf("Lot Size:    %d sqft", detail.LotSize))
	}
	if detail.YearBuilt > 0 {
		lines = append(lines, fmt.Sprintf("Year Built:  %d", detail.YearBuilt))
	}
	if detail.HomeType != "" {
		lines = append(lines, fmt.Sprintf("Type:        %s", detail.HomeType))
	}
	if detail.Status != "" {
		lines = append(lines, fmt.Sprintf("Status:      %s", detail.Status))
	}
	if detail.Zestimate > 0 {
		lines = append(lines, fmt.Sprintf("Zestimate:   %s", formatPrice(detail.Zestimate)))
	}
	if detail.RentZestimate > 0 {
		lines = append(lines, fmt.Sprintf("Rent Zest:   %s/mo", formatPrice(detail.RentZestimate)))
	}
	if detail.MonthlyHOA > 0 {
		lines = append(lines, fmt.Sprintf("HOA:         %s/mo", formatPrice(int64(detail.MonthlyHOA))))
	}
	if detail.DaysOnMarket > 0 {
		lines = append(lines, fmt.Sprintf("Days Listed: %d", detail.DaysOnMarket))
	}
	if detail.ListingAgent != "" {
		lines = append(lines, fmt.Sprintf("Agent:       %s", detail.ListingAgent))
	}
	if detail.ListingBrokerage != "" {
		lines = append(lines, fmt.Sprintf("Brokerage:   %s", detail.ListingBrokerage))
	}
	if detail.Description != "" {
		lines = append(lines, fmt.Sprintf("\nDescription:\n  %s", truncate(detail.Description, 500)))
	}
	if detail.ZillowURL != "" {
		lines = append(lines, fmt.Sprintf("\nURL:         %s", detail.ZillowURL))
	}
	if detail.Latitude != 0 || detail.Longitude != 0 {
		lines = append(lines, fmt.Sprintf("Location:    %.6f, %.6f", detail.Latitude, detail.Longitude))
	}

	if len(detail.Photos) > 0 {
		lines = append(lines, fmt.Sprintf("\nPhotos (%d):", len(detail.Photos)))
		limit := 5
		if len(detail.Photos) < limit {
			limit = len(detail.Photos)
		}
		for _, p := range detail.Photos[:limit] {
			lines = append(lines, fmt.Sprintf("  %s", p))
		}
		if len(detail.Photos) > 5 {
			lines = append(lines, fmt.Sprintf("  ... and %d more", len(detail.Photos)-5))
		}
	}

	fmt.Println(strings.Join(lines, "\n"))
}

// printAgentSummaries outputs agent summaries as JSON or a formatted text table.
func printAgentSummaries(cmd *cobra.Command, summaries []AgentSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(summaries)
	}

	if len(summaries) == 0 {
		fmt.Println("No agents found.")
		return nil
	}

	lines := make([]string, 0, len(summaries)+1)
	lines = append(lines, fmt.Sprintf("%-12s  %-30s  %-15s  %-6s  %-5s  %-6s",
		"ID", "NAME", "PHONE", "RATING", "REVS", "SALES"))
	for _, a := range summaries {
		name := truncate(a.Name, 30)
		rating := "-"
		if a.Rating > 0 {
			rating = fmt.Sprintf("%.1f", a.Rating)
		}
		reviews := "-"
		if a.ReviewCount > 0 {
			reviews = fmt.Sprintf("%d", a.ReviewCount)
		}
		sales := "-"
		if a.RecentSales > 0 {
			sales = fmt.Sprintf("%d", a.RecentSales)
		}
		lines = append(lines, fmt.Sprintf("%-12s  %-30s  %-15s  %-6s  %-5s  %-6s",
			a.AgentID, name, a.Phone, rating, reviews, sales))
	}
	cli.PrintText(lines)
	return nil
}
