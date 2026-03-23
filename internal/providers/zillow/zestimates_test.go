package zillow

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestZestimateGet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "zestimates", "get", "--zpid=12345678"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "12345678") {
			t.Errorf("expected ZPID in output, got: %s", out)
		}
		if !strings.Contains(out, "Zestimate") {
			t.Errorf("expected Zestimate label in output, got: %s", out)
		}
		if !strings.Contains(out, "460,000") {
			t.Errorf("expected zestimate value 460,000 in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "zestimates", "get", "--zpid=12345678", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result ZestimateSummary
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if result.ZPID != "12345678" {
			t.Errorf("expected ZPID 12345678, got: %s", result.ZPID)
		}
		if result.Zestimate != 460000 {
			t.Errorf("expected Zestimate 460000, got: %d", result.Zestimate)
		}
		if result.RentZestimate != 2200 {
			t.Errorf("expected RentZestimate 2200, got: %d", result.RentZestimate)
		}
	})
}

func TestZestimateRent(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "zestimates", "rent", "--zpid=12345678"})
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
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "zestimates", "rent", "--zpid=12345678", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result map[string]any
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if _, ok := result["rentZestimate"]; !ok {
			t.Errorf("expected rentZestimate key in JSON, got: %v", result)
		}
		if result["zpid"] != "12345678" {
			t.Errorf("expected zpid 12345678, got: %v", result["zpid"])
		}
	})
}

func TestZestimateChart(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	// The mock server returns no homeValueChartData, so the chart will be empty.
	// We verify the command runs without error and handles empty data gracefully.
	t.Run("text_empty", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "zestimates", "chart", "--zpid=12345678"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		// Either prints chart data or "No chart data found."
		if out == "" {
			t.Errorf("expected some output, got empty string")
		}
	})

	t.Run("json_empty", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "zestimates", "chart", "--zpid=12345678", "--json"})
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
	})

	t.Run("with_duration", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "zestimates", "chart", "--zpid=12345678", "--duration=5y"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		_ = out // just verify it runs
	})
}

func TestFetchZestimate(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	client := newTestClient(server)

	result, err := fetchZestimate(t.Context(), client, "12345678")
	if err != nil {
		t.Fatalf("fetchZestimate: %v", err)
	}
	if result.ZPID != "12345678" {
		t.Errorf("expected ZPID 12345678, got %s", result.ZPID)
	}
	if result.Zestimate != 460000 {
		t.Errorf("expected Zestimate 460000, got %d", result.Zestimate)
	}
	if result.RentZestimate != 2200 {
		t.Errorf("expected RentZestimate 2200, got %d", result.RentZestimate)
	}
}
