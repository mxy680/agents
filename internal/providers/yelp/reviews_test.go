package yelp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReviewsList(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "reviews", "list", "--id=gary-danko-san-francisco"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "John D.") {
			t.Errorf("expected reviewer name in output, got: %s", out)
		}
		if !strings.Contains(out, "5.0") {
			t.Errorf("expected rating in output, got: %s", out)
		}
		if !strings.Contains(out, "2024-03-15") {
			t.Errorf("expected date in output, got: %s", out)
		}
		if !strings.Contains(out, "Absolutely amazing") {
			t.Errorf("expected review text in output, got: %s", out)
		}
	})

	t.Run("json output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "reviews", "list", "--id=gary-danko-san-francisco", "--json"})
			_ = root.Execute()
		})
		if !strings.Contains(out, `"reviews"`) {
			t.Errorf("expected reviews field in JSON output, got: %s", out)
		}
		if !strings.Contains(out, "xAG4O7l-t1ubbwVAlPnDKg") {
			t.Errorf("expected review id in JSON output, got: %s", out)
		}
	})
}

func TestReviewsListWithOptions(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"yelp", "reviews", "list",
			"--id=gary-danko-san-francisco",
			"--locale=en_US",
			"--sort-by=newest",
			"--limit=10",
			"--offset=0",
			"--json",
		})
		_ = root.Execute()
	})
	if !strings.Contains(out, "xAG4O7l-t1ubbwVAlPnDKg") {
		t.Errorf("expected review id in JSON output with options, got: %s", out)
	}
}

func TestReviewsAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"yelp", "review", "list", "--id=gary-danko-san-francisco"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "John D.") {
		t.Errorf("expected reviewer name via 'review' alias, got: %s", out)
	}
}

func TestReviewsEmptyList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/biz/empty-biz/review_feed", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"total":   0,
			"reviews": []map[string]any{},
		}
		json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"yelp", "reviews", "list", "--id=empty-biz"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "No reviews found.") {
		t.Errorf("expected 'No reviews found.' message, got: %s", out)
	}
}
