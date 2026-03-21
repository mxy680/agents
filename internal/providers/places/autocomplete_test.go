package places

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAutocompleteJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	acCmd := newAutocompleteCmd(factory)
	root.AddCommand(acCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"autocomplete", "--input=coffee", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var suggestions []AutocompleteSuggestion
	if err := json.Unmarshal([]byte(out), &suggestions); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(suggestions))
	}
	if suggestions[0].Type != "place" {
		t.Errorf("first suggestion type = %q, want place", suggestions[0].Type)
	}
	if suggestions[0].PlaceID != "ChIJ1" {
		t.Errorf("first suggestion placeID = %q, want ChIJ1", suggestions[0].PlaceID)
	}
	if suggestions[0].Distance != 1500 {
		t.Errorf("first suggestion distance = %d, want 1500", suggestions[0].Distance)
	}
	if suggestions[1].Type != "query" {
		t.Errorf("second suggestion type = %q, want query", suggestions[1].Type)
	}
	if suggestions[1].Text != "coffee shops near me" {
		t.Errorf("second suggestion text = %q", suggestions[1].Text)
	}
}

func TestAutocompleteText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	acCmd := newAutocompleteCmd(factory)
	root.AddCommand(acCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"autocomplete", "--input=coffee"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "Coffee Corner") {
		t.Errorf("expected 'Coffee Corner' in output, got: %s", out)
	}
	if !strings.Contains(out, "coffee shops near me") {
		t.Errorf("expected 'coffee shops near me' in output, got: %s", out)
	}
	if !strings.Contains(out, "place") {
		t.Errorf("expected 'place' type in output, got: %s", out)
	}
	if !strings.Contains(out, "query") {
		t.Errorf("expected 'query' type in output, got: %s", out)
	}
}

func TestAutocompleteWithOptions(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	acCmd := newAutocompleteCmd(factory)
	root.AddCommand(acCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"autocomplete", "--input=coffee",
			"--types=cafe,restaurant",
			"--regions=us",
			"--location-bias=41.4993,-81.6944,5000",
			"--origin=41.5,-81.7",
			"--lang=en",
			"--region=us",
			"--include-queries",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var suggestions []AutocompleteSuggestion
	if err := json.Unmarshal([]byte(out), &suggestions); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestAutocompleteMissingInput(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	acCmd := newAutocompleteCmd(factory)
	root.AddCommand(acCmd)

	root.SetArgs([]string{"autocomplete"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing --input")
	}
}
