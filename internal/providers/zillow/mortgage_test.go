package zillow

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
)

func TestMortgageRates(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "mortgage", "rates"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Fixed30Year") {
			t.Errorf("expected Fixed30Year in output, got: %s", out)
		}
		if !strings.Contains(out, "6.875") {
			t.Errorf("expected rate value 6.875 in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "mortgage", "rates", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []MortgageRate
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
		}
		if len(results) == 0 {
			t.Errorf("expected at least one rate")
		}
		found30 := false
		for _, r := range results {
			if r.Program == "Fixed30Year" {
				found30 = true
				if r.Rate != 6.875 {
					t.Errorf("expected rate 6.875, got %f", r.Rate)
				}
			}
		}
		if !found30 {
			t.Errorf("expected Fixed30Year program in results")
		}
	})
}

func TestMortgageRatesHistory(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "mortgage", "rates-history"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Rate History") {
			t.Errorf("expected Rate History in output, got: %s", out)
		}
		if !strings.Contains(out, "Fixed30Year") {
			t.Errorf("expected Fixed30Year in output, got: %s", out)
		}
		if !strings.Contains(out, "2024-12-15") {
			t.Errorf("expected date 2024-12-15 in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "mortgage", "rates-history", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []MortgageRate
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
		}
		if len(results) == 0 {
			t.Errorf("expected at least one rate history entry")
		}
		if results[0].Date == "" {
			t.Errorf("expected date in history entry")
		}
	})
}

func TestMortgageCalculate(t *testing.T) {
	// Pure calculation — does not need a mock server.
	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: DefaultClientFactory()}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{
				"zillow", "mortgage", "calculate",
				"--price=500000",
				"--down-payment=20",
				"--rate=7.0",
				"--term=30",
			})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Mortgage Calculation") {
			t.Errorf("expected Mortgage Calculation in output, got: %s", out)
		}
		if !strings.Contains(out, "500,000") {
			t.Errorf("expected home price 500,000 in output, got: %s", out)
		}
		// Down payment = 20% of 500000 = 100000 → loan amount = 400000
		if !strings.Contains(out, "400,000") {
			t.Errorf("expected loan amount 400,000 in output, got: %s", out)
		}
		// Monthly payment for 400k at 7% for 30 years ≈ $2661
		if !strings.Contains(out, "2661") && !strings.Contains(out, "2,661") {
			t.Errorf("expected monthly payment ~2661 in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: DefaultClientFactory()}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{
				"zillow", "mortgage", "calculate",
				"--price=500000",
				"--down-payment=20",
				"--rate=7.0",
				"--term=30",
				"--json",
			})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result MortgageCalculation
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		// principal = 400000
		if result.Principal != 400000 {
			t.Errorf("expected principal 400000, got %f", result.Principal)
		}
		// monthly payment for P=400000, r=7%/12, n=360 ≈ 2661.21
		expectedPayment := 2661.21
		if math.Abs(result.MonthlyPayment-expectedPayment) > 1.0 {
			t.Errorf("expected monthly payment ~%.2f, got %.2f", expectedPayment, result.MonthlyPayment)
		}
	})
}

func TestMortgageLenderReviews(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "mortgage", "lender-reviews", "--nmls-id=12345"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "12345") {
			t.Errorf("expected NMLS ID in output, got: %s", out)
		}
		if !strings.Contains(out, "4.8") {
			t.Errorf("expected rating 4.8 in output, got: %s", out)
		}
		if !strings.Contains(out, "Great experience") {
			t.Errorf("expected review title in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "mortgage", "lender-reviews", "--nmls-id=12345", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result LenderInfo
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if result.Rating != 4.8 {
			t.Errorf("expected rating 4.8, got %f", result.Rating)
		}
		if result.TotalReviews != 42 {
			t.Errorf("expected totalReviews 42, got %d", result.TotalReviews)
		}
		if len(result.Reviews) == 0 {
			t.Errorf("expected at least one review")
		}
		if result.Reviews[0].Title != "Great experience" {
			t.Errorf("expected review title 'Great experience', got %s", result.Reviews[0].Title)
		}
	})
}

func TestParseCurrentRates(t *testing.T) {
	t.Run("valid_rates", func(t *testing.T) {
		body := []byte(`{
			"status": "OK",
			"rates": {
				"Fixed30Year": {"rate": 6.875, "apr": 7.012},
				"Fixed15Year": {"rate": 6.125, "apr": 6.287}
			}
		}`)

		rates, err := parseCurrentRates(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rates) != 2 {
			t.Fatalf("expected 2 rates, got %d", len(rates))
		}

		rateMap := make(map[string]MortgageRate)
		for _, r := range rates {
			rateMap[r.Program] = r
		}

		r30, ok := rateMap["Fixed30Year"]
		if !ok {
			t.Fatal("expected Fixed30Year in rates")
		}
		if r30.Rate != 6.875 {
			t.Errorf("expected rate 6.875, got %f", r30.Rate)
		}
		if r30.APR != 7.012 {
			t.Errorf("expected APR 7.012, got %f", r30.APR)
		}
	})

	t.Run("empty_rates", func(t *testing.T) {
		body := []byte(`{"status": "OK"}`)
		rates, err := parseCurrentRates(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rates != nil {
			t.Errorf("expected nil rates for empty response, got %v", rates)
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		_, err := parseCurrentRates([]byte(`not-json`))
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

func TestParseLenderReviews(t *testing.T) {
	t.Run("valid_response", func(t *testing.T) {
		body := []byte(`{
			"profileURL": "https://www.zillow.com/lender-profile/12345/",
			"reviewURL": "https://www.zillow.com/lender-reviews/12345/",
			"totalReviews": 42,
			"rating": 4.8,
			"reviews": [
				{
					"rating": 5.0,
					"title": "Great experience",
					"content": "Very helpful and responsive.",
					"loanType": "Conventional",
					"loanProgram": "Fixed30Year",
					"closingCostsSatisfaction": 5.0,
					"interestRateSatisfaction": 4.5,
					"verifiedReviewer": true
				}
			]
		}`)

		info, err := parseLenderReviews(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.ProfileURL != "https://www.zillow.com/lender-profile/12345/" {
			t.Errorf("unexpected ProfileURL: %s", info.ProfileURL)
		}
		if info.TotalReviews != 42 {
			t.Errorf("expected totalReviews 42, got %d", info.TotalReviews)
		}
		if info.Rating != 4.8 {
			t.Errorf("expected rating 4.8, got %f", info.Rating)
		}
		if len(info.Reviews) != 1 {
			t.Fatalf("expected 1 review, got %d", len(info.Reviews))
		}
		r := info.Reviews[0]
		if r.Title != "Great experience" {
			t.Errorf("expected title 'Great experience', got %s", r.Title)
		}
		if r.Rating != 5.0 {
			t.Errorf("expected rating 5.0, got %f", r.Rating)
		}
		if !r.VerifiedReviewer {
			t.Errorf("expected verifiedReviewer true")
		}
		if r.ClosingCostsSatisfaction != 5.0 {
			t.Errorf("expected closingCostsSatisfaction 5.0, got %f", r.ClosingCostsSatisfaction)
		}
		if r.InterestRateSatisfaction != 4.5 {
			t.Errorf("expected interestRateSatisfaction 4.5, got %f", r.InterestRateSatisfaction)
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		_, err := parseLenderReviews([]byte(`not-json`))
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("empty_response", func(t *testing.T) {
		body := []byte(`{}`)
		info, err := parseLenderReviews(body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.TotalReviews != 0 {
			t.Errorf("expected 0 reviews, got %d", info.TotalReviews)
		}
	})
}
