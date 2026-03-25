package citibike

import (
	"encoding/json"
	"strings"
	"testing"
)

// Times Square coordinates used throughout tests.
const (
	timesSquareLat = 40.7580
	timesSquareLng = -73.9855
)

// --------------------------------------------------------------------------
// filterStationsByRadius unit tests
// --------------------------------------------------------------------------

func TestFilterStationsByRadius_WithinRadius(t *testing.T) {
	// station-001 is AT the centre (distance ≈ 0), station-002 is ~800 m away,
	// station-003 is ~2 km away. With radius=1000 we expect stations 1 and 2.
	got := filterStationsByRadius(testStations, timesSquareLat, timesSquareLng, 1000)
	if len(got) < 1 {
		t.Fatalf("expected at least 1 station within 1000 m, got %d", len(got))
	}
	// First result must be the closest station.
	if got[0].Name != "W 42 St & 8 Ave" {
		t.Errorf("expected closest station to be W 42 St & 8 Ave, got %q", got[0].Name)
	}
}

func TestFilterStationsByRadius_NoneInRadius(t *testing.T) {
	// Search from a point that is far from all test stations (midtown NJ).
	// All test stations are in Manhattan so none should be within 100 m.
	got := filterStationsByRadius(testStations, 40.7282, -74.0776, 100)
	if len(got) != 0 {
		t.Errorf("expected 0 stations within 100 m of NJ point, got %d", len(got))
	}
}

func TestFilterStationsByRadius_SortedByDistance(t *testing.T) {
	got := filterStationsByRadius(testStations, timesSquareLat, timesSquareLng, 5000)
	for i := 1; i < len(got); i++ {
		if got[i].DistanceM < got[i-1].DistanceM {
			t.Errorf("results not sorted by distance: got[%d].DistanceM=%.1f < got[%d].DistanceM=%.1f",
				i, got[i].DistanceM, i-1, got[i-1].DistanceM)
		}
	}
}

func TestFilterStationsByRadius_AllReturned(t *testing.T) {
	// Very large radius should return all test stations.
	got := filterStationsByRadius(testStations, timesSquareLat, timesSquareLng, 100_000)
	if len(got) != len(testStations) {
		t.Errorf("expected %d stations, got %d", len(testStations), len(got))
	}
}

// --------------------------------------------------------------------------
// computeDensity unit tests
// --------------------------------------------------------------------------

func TestComputeDensity_AllStations(t *testing.T) {
	d := computeDensity(testStations, timesSquareLat, timesSquareLng, 100_000)
	if d.Count != len(testStations) {
		t.Errorf("expected count %d, got %d", len(testStations), d.Count)
	}
	expectedTotal := 35 + 27 + 42 // 104
	if d.TotalCapacity != expectedTotal {
		t.Errorf("expected total capacity %d, got %d", expectedTotal, d.TotalCapacity)
	}
	expectedAvg := float64(expectedTotal) / float64(len(testStations))
	if d.AvgCapacity != expectedAvg {
		t.Errorf("expected avg capacity %.4f, got %.4f", expectedAvg, d.AvgCapacity)
	}
	if d.RadiusM != 100_000 {
		t.Errorf("expected radius_m 100000, got %.0f", d.RadiusM)
	}
}

func TestComputeDensity_NoStations(t *testing.T) {
	// Search from NJ — all test stations are in Manhattan, none within 100 m.
	d := computeDensity(testStations, 40.7282, -74.0776, 100)
	if d.Count != 0 {
		t.Errorf("expected count 0, got %d", d.Count)
	}
	if d.AvgCapacity != 0 {
		t.Errorf("expected avg_capacity 0 when count is 0, got %.2f", d.AvgCapacity)
	}
	if d.TotalCapacity != 0 {
		t.Errorf("expected total_capacity 0, got %d", d.TotalCapacity)
	}
}

// --------------------------------------------------------------------------
// haversineMeters unit tests
// --------------------------------------------------------------------------

func TestHaversineMeters_SamePoint(t *testing.T) {
	d := haversineMeters(40.758, -73.985, 40.758, -73.985)
	if d != 0 {
		t.Errorf("expected 0 m for same point, got %.2f", d)
	}
}

func TestHaversineMeters_KnownDistance(t *testing.T) {
	// Approx distance between Times Square and Empire State Building ~650 m.
	d := haversineMeters(40.7580, -73.9855, 40.7484, -73.9857)
	if d < 500 || d > 1500 {
		t.Errorf("expected ~650 m between Times Sq and ESB, got %.0f m", d)
	}
}

// --------------------------------------------------------------------------
// `stations search` command integration tests (via mock server)
// --------------------------------------------------------------------------

func TestStationsSearch_TextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"citibike", "stations", "search",
			"--lat", "40.758", "--lng", "-73.9855", "--radius", "5000",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "W 42 St") {
		t.Errorf("expected station name in output, got: %s", out)
	}
}

func TestStationsSearch_JSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"citibike", "stations", "search",
			"--lat", "40.758", "--lng", "-73.9855", "--radius", "5000",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var results []StationSummary
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
	}
	if len(results) == 0 {
		t.Errorf("expected at least one station in JSON output")
	}
	if results[0].Name == "" {
		t.Errorf("expected non-empty station name")
	}
	if results[0].DistanceM < 0 {
		t.Errorf("expected non-negative distance_m")
	}
}

func TestStationsSearch_NearbyAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		// Use the "nearby" alias.
		root.SetArgs([]string{
			"citibike", "stations", "nearby",
			"--lat", "40.758", "--lng", "-73.9855",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// Default radius is 500 m; station-001 is at the exact centre so it
	// should always appear.
	if !strings.Contains(out, "W 42 St") {
		t.Errorf("expected station name via nearby alias, got: %s", out)
	}
}

func TestStationsSearch_EmptyResult(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		// Search from NJ with tiny radius — no test stations should match.
		root.SetArgs([]string{
			"citibike", "stations", "search",
			"--lat", "40.7282", "--lng", "-74.0776", "--radius", "100",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "No stations") {
		t.Errorf("expected empty-results message, got: %s", out)
	}
}

// --------------------------------------------------------------------------
// `stations density` command integration tests (via mock server)
// --------------------------------------------------------------------------

func TestStationsDensity_TextOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"citibike", "stations", "density",
			"--lat", "40.758", "--lng", "-73.9855", "--radius", "100000",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if !strings.Contains(out, "Station count:") {
		t.Errorf("expected 'Station count:' in output, got: %s", out)
	}
	if !strings.Contains(out, "Total capacity:") {
		t.Errorf("expected 'Total capacity:' in output, got: %s", out)
	}
}

func TestStationsDensity_JSONOutput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"citibike", "stations", "density",
			"--lat", "40.758", "--lng", "-73.9855", "--radius", "100000",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var d DensitySummary
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		t.Fatalf("expected valid JSON object, got: %s, error: %v", out, err)
	}
	if d.Count != len(testStations) {
		t.Errorf("expected count %d, got %d", len(testStations), d.Count)
	}
	if d.TotalCapacity != 104 { // 35+27+42
		t.Errorf("expected total_capacity 104, got %d", d.TotalCapacity)
	}
	if d.RadiusM != 100000 {
		t.Errorf("expected radius_m 100000, got %.0f", d.RadiusM)
	}
}

func TestStationsDensity_DefaultRadius(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{ClientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		// Omit --radius to exercise the 1000 m default.
		root.SetArgs([]string{
			"citibike", "stations", "density",
			"--lat", "40.758", "--lng", "-73.9855", "--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	var d DensitySummary
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		t.Fatalf("expected valid JSON, got: %s, error: %v", out, err)
	}
	if d.RadiusM != 1000 {
		t.Errorf("expected default radius_m 1000, got %.0f", d.RadiusM)
	}
}

// --------------------------------------------------------------------------
// truncateName unit tests
// --------------------------------------------------------------------------

func TestTruncateName(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"Short", 10, "Short"},
		{"Exactly ten", 11, "Exactly ten"},
		{"A long station name here", 10, "A long ..."},
	}
	for _, tt := range tests {
		got := truncateName(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncateName(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}
