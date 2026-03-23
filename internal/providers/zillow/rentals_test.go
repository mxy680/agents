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
