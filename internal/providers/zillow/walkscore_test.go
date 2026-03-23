package zillow

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestWalkScoreGet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "walkscore", "get", "--zpid=12345678"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "12345678") {
			t.Errorf("expected ZPID in output, got: %s", out)
		}
		if !strings.Contains(out, "Walk Score") {
			t.Errorf("expected Walk Score label in output, got: %s", out)
		}
		if !strings.Contains(out, "82") {
			t.Errorf("expected walk score 82 in output, got: %s", out)
		}
		if !strings.Contains(out, "Very Walkable") {
			t.Errorf("expected description in output, got: %s", out)
		}
		if !strings.Contains(out, "Transit Score") {
			t.Errorf("expected Transit Score label in output, got: %s", out)
		}
		if !strings.Contains(out, "65") {
			t.Errorf("expected transit score 65 in output, got: %s", out)
		}
		if !strings.Contains(out, "Bike Score") {
			t.Errorf("expected Bike Score label in output, got: %s", out)
		}
		if !strings.Contains(out, "70") {
			t.Errorf("expected bike score 70 in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "walkscore", "get", "--zpid=12345678", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var result WalkScoreResult
		if err := json.Unmarshal([]byte(out), &result); err != nil {
			t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
		}
		if result.ZPID != "12345678" {
			t.Errorf("expected ZPID 12345678, got %s", result.ZPID)
		}
		if result.WalkScore != 82 {
			t.Errorf("expected WalkScore 82, got %d", result.WalkScore)
		}
		if result.TransitScore != 65 {
			t.Errorf("expected TransitScore 65, got %d", result.TransitScore)
		}
		if result.BikeScore != 70 {
			t.Errorf("expected BikeScore 70, got %d", result.BikeScore)
		}
		if result.WalkDesc != "Very Walkable" {
			t.Errorf("expected WalkDesc 'Very Walkable', got %s", result.WalkDesc)
		}
	})
}

func TestParseWalkScore(t *testing.T) {
	t.Run("full_response", func(t *testing.T) {
		body := []byte(`{
			"data": {
				"property": {
					"walkScore": {
						"walkscore": 82,
						"description": "Very Walkable",
						"ws_link": "/walk-score/test"
					},
					"transitScore": {
						"transit_score": 65,
						"description": "Excellent Transit"
					},
					"bikeScore": {
						"bike_score": 70,
						"description": "Very Bikeable"
					}
				}
			}
		}`)

		result, err := parseWalkScore(body, "12345678")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ZPID != "12345678" {
			t.Errorf("expected ZPID 12345678, got %s", result.ZPID)
		}
		if result.WalkScore != 82 {
			t.Errorf("expected WalkScore 82, got %d", result.WalkScore)
		}
		if result.WalkDesc != "Very Walkable" {
			t.Errorf("expected WalkDesc 'Very Walkable', got %s", result.WalkDesc)
		}
		if result.TransitScore != 65 {
			t.Errorf("expected TransitScore 65, got %d", result.TransitScore)
		}
		if result.TransitDesc != "Excellent Transit" {
			t.Errorf("expected TransitDesc 'Excellent Transit', got %s", result.TransitDesc)
		}
		if result.BikeScore != 70 {
			t.Errorf("expected BikeScore 70, got %d", result.BikeScore)
		}
		if result.BikeDesc != "Very Bikeable" {
			t.Errorf("expected BikeDesc 'Very Bikeable', got %s", result.BikeDesc)
		}
	})

	t.Run("missing_data", func(t *testing.T) {
		body := []byte(`{}`)
		result, err := parseWalkScore(body, "99999")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ZPID != "99999" {
			t.Errorf("expected ZPID 99999, got %s", result.ZPID)
		}
		if result.WalkScore != 0 {
			t.Errorf("expected WalkScore 0 for missing data, got %d", result.WalkScore)
		}
	})

	t.Run("missing_property", func(t *testing.T) {
		body := []byte(`{"data": {}}`)
		result, err := parseWalkScore(body, "99999")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.WalkScore != 0 {
			t.Errorf("expected WalkScore 0 for missing property, got %d", result.WalkScore)
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		_, err := parseWalkScore([]byte(`not-json`), "12345678")
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}
