package zillow

import (
	"encoding/json"
	"fmt"
	"math"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newMortgageCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mortgage",
		Short:   "View mortgage rates and calculators",
		Aliases: []string{"mort"},
	}

	cmd.AddCommand(newMortgageRatesCmd(factory))
	cmd.AddCommand(newMortgageRatesHistoryCmd(factory))
	cmd.AddCommand(newMortgageCalculateCmd(factory))
	cmd.AddCommand(newMortgageLenderReviewsCmd(factory))

	return cmd
}

func newMortgageRatesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rates",
		Short: "Get current mortgage rates",
		RunE:  makeRunMortgageRates(factory),
	}
	cmd.Flags().String("state", "", "State abbreviation (e.g., CO)")
	cmd.Flags().String("program", "", "Loan program: Fixed30Year, Fixed15Year, Fixed20Year, ARM5, ARM7")
	cmd.Flags().String("loan-type", "", "Loan type: Conventional, FHA, VA, USDA, Jumbo")
	cmd.Flags().String("credit-score", "", "Credit score bucket: Low, High, VeryHigh")
	return cmd
}

func makeRunMortgageRates(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		state, _ := cmd.Flags().GetString("state")
		program, _ := cmd.Flags().GetString("program")
		loanType, _ := cmd.Flags().GetString("loan-type")
		creditScore, _ := cmd.Flags().GetString("credit-score")

		reqURL := client.mortURL + "/api/getCurrentRates"
		params := "?"
		if state != "" {
			params += "stateAbbreviation=" + state + "&"
		}
		if program != "" {
			params += "program=" + program + "&"
		}
		if loanType != "" {
			params += "loanType=" + loanType + "&"
		}
		if creditScore != "" {
			params += "creditScoreBucket=" + creditScore + "&"
		}
		if params != "?" {
			reqURL += params[:len(params)-1]
		}

		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("get rates: %w", err)
		}

		rates, err := parseCurrentRates(body)
		if err != nil {
			return fmt.Errorf("parse rates: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(rates)
		}

		if len(rates) == 0 {
			fmt.Println("No rates available.")
			return nil
		}
		lines := []string{"Current Mortgage Rates:"}
		lines = append(lines, fmt.Sprintf("  %-15s  %-8s  %-8s", "PROGRAM", "RATE", "APR"))
		for _, r := range rates {
			lines = append(lines, fmt.Sprintf("  %-15s  %-8s  %-8s",
				r.Program, fmt.Sprintf("%.3f%%", r.Rate), fmt.Sprintf("%.3f%%", r.APR)))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newMortgageRatesHistoryCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rates-history",
		Short: "Get historical mortgage rates",
		RunE:  makeRunMortgageRatesHistory(factory),
	}
	cmd.Flags().String("state", "", "State abbreviation")
	cmd.Flags().String("program", "Fixed30Year", "Loan program")
	cmd.Flags().Int("duration-days", 30, "History duration in days (1-4000)")
	cmd.Flags().String("aggregation", "Daily", "Aggregation: Daily, Weekly, Monthly")
	return cmd
}

func makeRunMortgageRatesHistory(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		state, _ := cmd.Flags().GetString("state")
		program, _ := cmd.Flags().GetString("program")
		durationDays, _ := cmd.Flags().GetInt("duration-days")
		aggregation, _ := cmd.Flags().GetString("aggregation")

		reqURL := fmt.Sprintf("%s/api/getRates?durationDays=%d&aggregation=%s",
			client.mortURL, durationDays, aggregation)
		if state != "" {
			reqURL += "&stateAbbreviation=" + state
		}
		if program != "" {
			reqURL += "&program=" + program
		}

		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("get rates history: %w", err)
		}

		rates, err := parseRatesHistory(body)
		if err != nil {
			return fmt.Errorf("parse rates history: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(rates)
		}

		if len(rates) == 0 {
			fmt.Println("No rate history available.")
			return nil
		}
		lines := []string{fmt.Sprintf("Rate History (%d days):", durationDays)}
		lines = append(lines, fmt.Sprintf("  %-12s  %-15s  %-8s  %-8s", "DATE", "PROGRAM", "RATE", "APR"))
		for _, r := range rates {
			lines = append(lines, fmt.Sprintf("  %-12s  %-15s  %-8s  %-8s",
				r.Date, r.Program, fmt.Sprintf("%.3f%%", r.Rate), fmt.Sprintf("%.3f%%", r.APR)))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newMortgageCalculateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calculate",
		Short: "Calculate monthly mortgage payment",
		RunE:  makeRunMortgageCalculate(factory),
	}
	cmd.Flags().Int64("price", 0, "Home price")
	cmd.Flags().Float64("down-payment", 20, "Down payment percentage")
	cmd.Flags().Float64("rate", 0, "Interest rate (if 0, uses current average)")
	cmd.Flags().Int("term", 30, "Loan term in years")
	cmd.MarkFlagRequired("price")
	return cmd
}

func makeRunMortgageCalculate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		price, _ := cmd.Flags().GetInt64("price")
		downPct, _ := cmd.Flags().GetFloat64("down-payment")
		rate, _ := cmd.Flags().GetFloat64("rate")
		term, _ := cmd.Flags().GetInt("term")

		downPayment := float64(price) * (downPct / 100.0)
		principal := float64(price) - downPayment
		monthlyRate := rate / 100.0 / 12.0
		numPayments := float64(term * 12)

		var monthlyPayment float64
		if monthlyRate > 0 {
			monthlyPayment = principal * (monthlyRate * math.Pow(1+monthlyRate, numPayments)) / (math.Pow(1+monthlyRate, numPayments) - 1)
		} else {
			monthlyPayment = principal / numPayments
		}

		totalCost := monthlyPayment * numPayments
		totalInterest := totalCost - principal

		calc := MortgageCalculation{
			MonthlyPayment: math.Round(monthlyPayment*100) / 100,
			Principal:      math.Round(principal*100) / 100,
			Interest:       math.Round(totalInterest*100) / 100,
			TotalCost:      math.Round(totalCost*100) / 100,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(calc)
		}

		lines := []string{
			"Mortgage Calculation:",
			fmt.Sprintf("  Home Price:      %s", formatPrice(price)),
			fmt.Sprintf("  Down Payment:    %s (%.0f%%)", formatPrice(int64(downPayment)), downPct),
			fmt.Sprintf("  Loan Amount:     %s", formatPrice(int64(principal))),
			fmt.Sprintf("  Interest Rate:   %.3f%%", rate),
			fmt.Sprintf("  Term:            %d years", term),
			"",
			fmt.Sprintf("  Monthly Payment: $%.2f", calc.MonthlyPayment),
			fmt.Sprintf("  Total Interest:  %s", formatPrice(int64(calc.Interest))),
			fmt.Sprintf("  Total Cost:      %s", formatPrice(int64(calc.TotalCost))),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newMortgageLenderReviewsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lender-reviews",
		Short: "Get lender reviews by NMLS ID",
		RunE:  makeRunMortgageLenderReviews(factory),
	}
	cmd.Flags().String("nmls-id", "", "NMLS ID of the lender/loan officer")
	cmd.Flags().String("company", "", "Company name (required for institutional lenders)")
	cmd.Flags().Int("limit", 3, "Maximum reviews (1-10)")
	cmd.MarkFlagRequired("nmls-id")
	return cmd
}

func makeRunMortgageLenderReviews(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return fmt.Errorf("create client: %w", err)
		}

		nmlsID, _ := cmd.Flags().GetString("nmls-id")
		company, _ := cmd.Flags().GetString("company")
		limit, _ := cmd.Flags().GetInt("limit")

		reqURL := fmt.Sprintf("%s/api/zillowLenderReviews?nmlsId=%s&reviewLimit=%d",
			client.mortURL, nmlsID, limit)
		if company != "" {
			reqURL += "&companyName=" + company
		}

		body, err := client.Get(ctx, reqURL)
		if err != nil {
			return fmt.Errorf("get lender reviews: %w", err)
		}

		info, err := parseLenderReviews(body)
		if err != nil {
			return fmt.Errorf("parse lender reviews: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("Lender NMLS %s:", nmlsID),
			fmt.Sprintf("  Rating:  %.1f (%d reviews)", info.Rating, info.TotalReviews),
		}
		if info.ProfileURL != "" {
			lines = append(lines, fmt.Sprintf("  Profile: %s", info.ProfileURL))
		}
		if len(info.Reviews) > 0 {
			lines = append(lines, "")
			for _, r := range info.Reviews {
				lines = append(lines, fmt.Sprintf("  %.0f★ — %s", r.Rating, r.Title))
				if r.Content != "" {
					lines = append(lines, fmt.Sprintf("    %s", truncate(r.Content, 120)))
				}
			}
		}
		cli.PrintText(lines)
		return nil
	}
}

// parseCurrentRates parses the getCurrentRates API response.
func parseCurrentRates(body []byte) ([]MortgageRate, error) {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	rates, _ := resp["rates"].(map[string]any)
	if rates == nil {
		return nil, nil
	}

	var result []MortgageRate
	for program, data := range rates {
		dm, ok := data.(map[string]any)
		if !ok {
			continue
		}
		r := MortgageRate{Program: program}
		if rate, ok := dm["rate"].(float64); ok {
			r.Rate = rate
		}
		if apr, ok := dm["apr"].(float64); ok {
			r.APR = apr
		}
		result = append(result, r)
	}
	return result, nil
}

// parseRatesHistory parses the getRates API response.
func parseRatesHistory(body []byte) ([]MortgageRate, error) {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	rates, _ := resp["rates"].(map[string]any)
	if rates == nil {
		return nil, nil
	}

	var result []MortgageRate
	for program, data := range rates {
		dm, ok := data.(map[string]any)
		if !ok {
			continue
		}
		samples, _ := dm["samples"].([]any)
		for _, s := range samples {
			sm, ok := s.(map[string]any)
			if !ok {
				continue
			}
			r := MortgageRate{
				Program: program,
				Date:    jsonStr(sm, "date"),
			}
			if rate, ok := sm["rate"].(float64); ok {
				r.Rate = rate
			}
			if apr, ok := sm["apr"].(float64); ok {
				r.APR = apr
			}
			result = append(result, r)
		}
	}
	return result, nil
}

// parseLenderReviews parses the lender reviews API response.
func parseLenderReviews(body []byte) (LenderInfo, error) {
	var resp map[string]any
	if err := json.Unmarshal(body, &resp); err != nil {
		return LenderInfo{}, err
	}

	info := LenderInfo{
		ProfileURL: jsonStr(resp, "profileURL"),
		ReviewURL:  jsonStr(resp, "reviewURL"),
	}
	if total, ok := resp["totalReviews"].(float64); ok {
		info.TotalReviews = int(total)
	}
	if rating, ok := resp["rating"].(float64); ok {
		info.Rating = rating
	}

	reviews, _ := resp["reviews"].([]any)
	for _, item := range reviews {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		r := LenderReview{
			Title:       jsonStr(m, "title"),
			Content:     jsonStr(m, "content"),
			LoanType:    jsonStr(m, "loanType"),
			LoanProgram: jsonStr(m, "loanProgram"),
		}
		if rating, ok := m["rating"].(float64); ok {
			r.Rating = rating
		}
		if cc, ok := m["closingCostsSatisfaction"].(float64); ok {
			r.ClosingCostsSatisfaction = cc
		}
		if ir, ok := m["interestRateSatisfaction"].(float64); ok {
			r.InterestRateSatisfaction = ir
		}
		if vr, ok := m["verifiedReviewer"].(bool); ok {
			r.VerifiedReviewer = vr
		}
		info.Reviews = append(info.Reviews, r)
	}
	return info, nil
}
