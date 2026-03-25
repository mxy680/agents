package yelp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAutocompleteQuery(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "autocomplete", "query", "--text=pizza"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "Terms:") {
			t.Errorf("expected Terms: label in output, got: %s", out)
		}
		if !strings.Contains(out, "pizza") {
			t.Errorf("expected 'pizza' in output, got: %s", out)
		}
		if !strings.Contains(out, "Businesses:") {
			t.Errorf("expected Businesses: section in output, got: %s", out)
		}
		if !strings.Contains(out, "Gary Danko") {
			t.Errorf("expected business name in output, got: %s", out)
		}
		if !strings.Contains(out, "Categories:") {
			t.Errorf("expected Categories: section in output, got: %s", out)
		}
		if !strings.Contains(out, "Pizza") {
			t.Errorf("expected category in output, got: %s", out)
		}
	})

	t.Run("json output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "autocomplete", "query", "--text=pizza", "--json"})
			_ = root.Execute()
		})
		if !strings.Contains(out, `"terms"`) {
			t.Errorf("expected terms field in JSON output, got: %s", out)
		}
		if !strings.Contains(out, `"businesses"`) {
			t.Errorf("expected businesses field in JSON output, got: %s", out)
		}
		if !strings.Contains(out, `"categories"`) {
			t.Errorf("expected categories field in JSON output, got: %s", out)
		}
	})
}

func TestAutocompleteQueryWithGeo(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"yelp", "autocomplete", "query",
			"--text=pizza",
			"--latitude=37.8051",
			"--longitude=-122.4212",
			"--locale=en_US",
			"--json",
		})
		_ = root.Execute()
	})
	if !strings.Contains(out, "pizza") {
		t.Errorf("expected 'pizza' in geo-biased autocomplete output, got: %s", out)
	}
}

func TestAutocompleteAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	// Test "ac" alias
	out := captureStdout(t, func() {
		root.SetArgs([]string{"yelp", "ac", "query", "--text=pizza"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "pizza") {
		t.Errorf("expected 'pizza' via 'ac' alias, got: %s", out)
	}
}

func TestAutocompleteQueryEmpty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/autocomplete", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"terms":      []map[string]any{},
			"businesses": []map[string]any{},
			"categories": []map[string]any{},
		}
		json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"yelp", "autocomplete", "query", "--text=zzzzz"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "No suggestions found.") {
		t.Errorf("expected 'No suggestions found.' message, got: %s", out)
	}
}
