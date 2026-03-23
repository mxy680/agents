package zillow

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRentalSearch(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "rentals", "search", "--location=Denver, CO"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "123 Main St") {
			t.Errorf("expected address in output, got: %s", out)
		}
		if !strings.Contains(out, "12345678") {
			t.Errorf("expected ZPID in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "rentals", "search", "--location=Denver, CO", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []PropertySummary
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
		}
		if len(results) == 0 {
			t.Errorf("expected at least one rental listing")
		}
	})

	t.Run("with_filters", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{
				"zillow", "rentals", "search",
				"--location=Denver, CO",
				"--min-price=1000",
				"--max-price=3000",
				"--min-beds=1",
				"--max-beds=3",
			})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		_ = out // just verify it runs without error
	})
}

func TestRentalGet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "rentals", "get", "--zpid=12345678"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "123 Main St") {
			t.Errorf("expected address in output, got: %s", out)
		}
		if !strings.Contains(out, "450,000") {
			t.Errorf("expected price in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "rentals", "get", "--zpid=12345678", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result PropertyDetail
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if result.ZPID != "12345678" {
			t.Errorf("expected ZPID 12345678, got %s", result.ZPID)
		}
		if result.Price != 450000 {
			t.Errorf("expected price 450000, got %d", result.Price)
		}
	})
}

func TestRentalEstimate(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "rentals", "estimate", "--zpid=12345678"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "Rent Zestimate") {
			t.Errorf("expected Rent Zestimate label in output, got: %s", out)
		}
		if !strings.Contains(out, "2,200") {
			t.Errorf("expected rent value 2,200 in output, got: %s", out)
		}
		if !strings.Contains(out, "12345678") {
			t.Errorf("expected ZPID in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "rentals", "estimate", "--zpid=12345678", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result map[string]any
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if result["zpid"] != "12345678" {
			t.Errorf("expected zpid in JSON, got: %v", result)
		}
		rentEst, ok := result["rentZestimate"].(float64)
		if !ok {
			t.Fatalf("expected rentZestimate as number in JSON, got: %T", result["rentZestimate"])
		}
		if int64(rentEst) != 2200 {
			t.Errorf("expected rentZestimate 2200, got %d", int64(rentEst))
		}
	})
}
