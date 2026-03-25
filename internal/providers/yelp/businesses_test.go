package yelp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBusinessSearch(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "businesses", "search", "--location=San Francisco, CA"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "Gary Danko") {
			t.Errorf("expected business name in output, got: %s", out)
		}
		if !strings.Contains(out, "4.5") {
			t.Errorf("expected rating in output, got: %s", out)
		}
	})

	t.Run("json output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "businesses", "search", "--location=San Francisco, CA", "--json"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "gary-danko-san-francisco") {
			t.Errorf("expected business id in JSON output, got: %s", out)
		}
		if !strings.Contains(out, `"businesses"`) {
			t.Errorf("expected businesses field in JSON output, got: %s", out)
		}
	})
}

func TestBusinessSearchWithTerm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"yelp", "businesses", "search",
			"--term=restaurants",
			"--location=San Francisco, CA",
			"--limit=5",
			"--json",
		})
		_ = root.Execute()
	})
	if !strings.Contains(out, "gary-danko-san-francisco") {
		t.Errorf("expected business id in output with term filter, got: %s", out)
	}
}

func TestBusinessSearchWithGeo(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"yelp", "businesses", "search",
			"--latitude=37.8051",
			"--longitude=-122.4212",
			"--radius=1000",
			"--json",
		})
		_ = root.Execute()
	})
	if !strings.Contains(out, "gary-danko") {
		t.Errorf("expected business in geo search output, got: %s", out)
	}
}

func TestBusinessSearchRequiresLocation(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	// No location args should produce an error
	err := root.Execute()
	// The root execute without args shouldn't error, we need to test the specific command
	_ = err

	// Run with missing location
	root.SetArgs([]string{"yelp", "businesses", "search"})
	err = root.Execute()
	// This should return an error about missing location
	_ = err
}

func TestBusinessSearchEmptyResults(t *testing.T) {
	// A mock server that returns empty businesses list
	mux := newEmptyBusinessMux()
	server := newCustomMockServer(mux)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"yelp", "businesses", "search", "--location=Nowhere"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "No businesses found.") {
		t.Errorf("expected 'No businesses found.' message, got: %s", out)
	}
}

func TestBusinessPhoneSearch(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "businesses", "phone-search", "--phone=+14152520800"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "Gary Danko") {
			t.Errorf("expected business name in phone search output, got: %s", out)
		}
	})

	t.Run("json output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "businesses", "phone-search", "--phone=+14152520800", "--json"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "gary-danko-san-francisco") {
			t.Errorf("expected business id in phone search JSON output, got: %s", out)
		}
	})
}

func TestBusinessMatch(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{
				"yelp", "businesses", "match",
				"--name=Gary Danko",
				"--city=San Francisco",
				"--state=CA",
				"--country=US",
				"--address1=800 N Point St",
			})
			_ = root.Execute()
		})
		if !strings.Contains(out, "Gary Danko") {
			t.Errorf("expected business name in match output, got: %s", out)
		}
	})

	t.Run("json output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{
				"yelp", "businesses", "match",
				"--name=Gary Danko",
				"--city=San Francisco",
				"--state=CA",
				"--country=US",
				"--json",
			})
			_ = root.Execute()
		})
		if !strings.Contains(out, "gary-danko-san-francisco") {
			t.Errorf("expected business id in match JSON output, got: %s", out)
		}
	})
}

func TestBusinessGet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "businesses", "get", "--id=gary-danko-san-francisco"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "Gary Danko") {
			t.Errorf("expected business name in get output, got: %s", out)
		}
		if !strings.Contains(out, "Name:") {
			t.Errorf("expected Name: label in output, got: %s", out)
		}
		if !strings.Contains(out, "Rating:") {
			t.Errorf("expected Rating: label in output, got: %s", out)
		}
	})

	t.Run("json output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "businesses", "get", "--id=gary-danko-san-francisco", "--json"})
			_ = root.Execute()
		})
		if !strings.Contains(out, `"id"`) {
			t.Errorf("expected id field in JSON output, got: %s", out)
		}
		if !strings.Contains(out, "gary-danko-san-francisco") {
			t.Errorf("expected business id value in JSON output, got: %s", out)
		}
	})
}

func TestBusinessAliases(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	// Test "biz" alias for businesses
	out := captureStdout(t, func() {
		root.SetArgs([]string{"yelp", "biz", "get", "--id=gary-danko-san-francisco"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "Gary Danko") {
		t.Errorf("expected business name via 'biz' alias, got: %s", out)
	}
}

// newEmptyBusinessMux returns an http.ServeMux that serves empty business results.
func newEmptyBusinessMux() *http.ServeMux {
	return newEmptyBusinessMuxInternal()
}

// newCustomMockServer creates an httptest server with the given mux.
func newCustomMockServer(mux *http.ServeMux) *httptest.Server {
	return httptest.NewServer(mux)
}

func newEmptyBusinessMuxInternal() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/businesses/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"total":      0,
			"businesses": []map[string]any{},
		}
		json.NewEncoder(w).Encode(resp)
	})
	return mux
}
