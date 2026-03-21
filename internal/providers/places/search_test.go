package places

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSearchTextJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	searchCmd := newSearchTextCmd(factory)
	root.AddCommand(searchCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"text", "--query=coffee", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	places, ok := result["places"].([]any)
	if !ok || len(places) != 2 {
		t.Fatalf("expected 2 places, got %v", result["places"])
	}
	if result["nextPageToken"] != "token123" {
		t.Errorf("expected nextPageToken=token123, got %v", result["nextPageToken"])
	}
}

func TestSearchTextText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	searchCmd := newSearchTextCmd(factory)
	root.AddCommand(searchCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"text", "--query=coffee"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "Coffee Corner") {
		t.Errorf("expected output to contain 'Coffee Corner', got: %s", out)
	}
	if !strings.Contains(out, "Bean & Leaf") {
		t.Errorf("expected output to contain 'Bean & Leaf', got: %s", out)
	}
	if !strings.Contains(out, "Next page") {
		t.Errorf("expected output to contain pagination info, got: %s", out)
	}
}

func TestSearchTextWithOptions(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	searchCmd := newSearchTextCmd(factory)
	root.AddCommand(searchCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"text", "--query=coffee",
			"--type=cafe",
			"--location-bias=41.4993,-81.6944,5000",
			"--min-rating=4.0",
			"--open-now",
			"--price-levels=1,2",
			"--rank=DISTANCE",
			"--region=us",
			"--lang=en",
			"--limit=10",
			"--fields=basic",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestSearchTextWithLocationRestrict(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	searchCmd := newSearchTextCmd(factory)
	root.AddCommand(searchCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"text", "--query=coffee",
			"--location-restrict=41.0,-82.0,42.0,-81.0",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestSearchTextMissingQuery(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	searchCmd := newSearchTextCmd(factory)
	root.AddCommand(searchCmd)

	root.SetArgs([]string{"text"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing --query")
	}
}

func TestSearchNearbyJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	nearbyCmd := newSearchNearbyCmd(factory)
	root.AddCommand(nearbyCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"nearby", "--lat=41.4993", "--lng=-81.6944", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var places []any
	if err := json.Unmarshal([]byte(out), &places); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(places) != 1 {
		t.Fatalf("expected 1 place, got %d", len(places))
	}
}

func TestSearchNearbyText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	nearbyCmd := newSearchNearbyCmd(factory)
	root.AddCommand(nearbyCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"nearby", "--lat=41.4993", "--lng=-81.6944"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "Nearby Diner") {
		t.Errorf("expected 'Nearby Diner' in output, got: %s", out)
	}
}

func TestSearchNearbyWithOptions(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	nearbyCmd := newSearchNearbyCmd(factory)
	root.AddCommand(nearbyCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"nearby",
			"--lat=41.4993", "--lng=-81.6944",
			"--radius=10000",
			"--types=restaurant,cafe",
			"--exclude-types=gas_station",
			"--primary-types=restaurant",
			"--rank=DISTANCE",
			"--region=us",
			"--lang=en",
			"--limit=5",
			"--fields=preferred",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var places []any
	if err := json.Unmarshal([]byte(out), &places); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}
