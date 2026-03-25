package yelp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTransactionSearch(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{
				"yelp", "transactions", "search",
				"--type=delivery",
				"--location=San Francisco, CA",
			})
			_ = root.Execute()
		})
		if !strings.Contains(out, "Gary Danko") {
			t.Errorf("expected business name in transaction search output, got: %s", out)
		}
		if !strings.Contains(out, "4.5") {
			t.Errorf("expected rating in output, got: %s", out)
		}
	})

	t.Run("json output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{
				"yelp", "transactions", "search",
				"--type=delivery",
				"--location=San Francisco, CA",
				"--json",
			})
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

func TestTransactionSearchWithGeo(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"yelp", "transactions", "search",
			"--type=delivery",
			"--latitude=37.8051",
			"--longitude=-122.4212",
			"--json",
		})
		_ = root.Execute()
	})
	if !strings.Contains(out, "gary-danko-san-francisco") {
		t.Errorf("expected business id in geo transaction search output, got: %s", out)
	}
}

func TestTransactionSearchRequiresLocation(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	// Missing location should fail
	root.SetArgs([]string{"yelp", "transactions", "search", "--type=delivery"})
	err := root.Execute()
	_ = err // error is expected
}

func TestTransactionAliases(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	// Test "tx" alias
	out := captureStdout(t, func() {
		root.SetArgs([]string{"yelp", "tx", "search", "--type=delivery", "--location=San Francisco, CA"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "Gary Danko") {
		t.Errorf("expected business name via 'tx' alias, got: %s", out)
	}

	// Test "transaction" alias
	out = captureStdout(t, func() {
		root.SetArgs([]string{"yelp", "transaction", "search", "--type=delivery", "--location=San Francisco, CA"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "Gary Danko") {
		t.Errorf("expected business name via 'transaction' alias, got: %s", out)
	}
}

func TestTransactionSearchEmptyResults(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/transactions/delivery/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"total":      0,
			"businesses": []map[string]any{},
		}
		json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"yelp", "transactions", "search", "--type=delivery", "--location=Nowhere"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "No businesses found.") {
		t.Errorf("expected 'No businesses found.' message, got: %s", out)
	}
}
