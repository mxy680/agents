package x

import (
	"testing"
)

func TestGeoReverse_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newGeoCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"geo", "reverse", "--lat=40.7128", "--lng=-74.0060", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestGeoReverse_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newGeoCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"geo", "reverse", "--lat=40.7128", "--lng=-74.0060"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestGeoSearch_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newGeoCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"geo", "search", "--query=New York", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestGeoSearch_WithLatLng(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newGeoCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"geo", "search", "--query=pizza", "--lat=40.7128", "--lng=-74.0060", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestGeoGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newGeoCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"geo", "get", "--place-id=place123", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestGeoGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newGeoCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"geo", "get", "--place-id=place123"})
		root.Execute() //nolint:errcheck
	})

	// Either shows place details or the ID.
	if out == "" {
		t.Errorf("expected some output, got empty string")
	}
}

func TestGeoAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newGeoCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"location", "search", "--query=NYC", "--json"})
		root.Execute() //nolint:errcheck
	})

	if out == "" {
		t.Errorf("expected some output via 'location' alias, got empty string")
	}
}

func TestParseSinglePlace(t *testing.T) {
	raw := []byte(`{
		"id": "place-001",
		"name": "New York",
		"full_name": "New York, NY",
		"place_type": "city",
		"country": "United States",
		"country_code": "US"
	}`)

	place := parseSinglePlace(raw)

	if place.ID != "place-001" {
		t.Errorf("expected ID place-001, got: %s", place.ID)
	}
	if place.Name != "New York" {
		t.Errorf("expected Name New York, got: %s", place.Name)
	}
	if place.CountryCode != "US" {
		t.Errorf("expected CountryCode US, got: %s", place.CountryCode)
	}
}

func TestParseSinglePlace_Empty(t *testing.T) {
	raw := []byte(`{}`)
	place := parseSinglePlace(raw)
	if place.ID != "" {
		t.Errorf("expected empty ID for empty raw, got: %s", place.ID)
	}
}

func TestPrintGeoPlaces_Empty(t *testing.T) {
	root := newTestRootCmd()
	out := captureStdout(t, func() {
		err := printGeoPlaces(root, []GeoPlace{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !containsStr(out, "No places") {
		t.Errorf("expected 'No places' in output, got: %s", out)
	}
}
