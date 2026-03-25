package yelp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCategoryList(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "categories", "list"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "restaurants") {
			t.Errorf("expected 'restaurants' alias in output, got: %s", out)
		}
		if !strings.Contains(out, "Restaurants") {
			t.Errorf("expected 'Restaurants' title in output, got: %s", out)
		}
		if !strings.Contains(out, "pizza") {
			t.Errorf("expected 'pizza' alias in output, got: %s", out)
		}
	})

	t.Run("json output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "categories", "list", "--json"})
			_ = root.Execute()
		})
		if !strings.Contains(out, `"categories"`) {
			t.Errorf("expected categories field in JSON output, got: %s", out)
		}
		if !strings.Contains(out, "restaurants") {
			t.Errorf("expected 'restaurants' in JSON output, got: %s", out)
		}
	})
}

func TestCategoryListWithLocale(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"yelp", "categories", "list", "--locale=en_US", "--json"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "restaurants") {
		t.Errorf("expected 'restaurants' in locale-filtered output, got: %s", out)
	}
}

func TestCategoryGet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "categories", "get", "--alias=pizza"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "Pizza") {
			t.Errorf("expected 'Pizza' title in output, got: %s", out)
		}
		if !strings.Contains(out, "Alias:") {
			t.Errorf("expected Alias: label in output, got: %s", out)
		}
		if !strings.Contains(out, "Title:") {
			t.Errorf("expected Title: label in output, got: %s", out)
		}
		if !strings.Contains(out, "Parents:") {
			t.Errorf("expected Parents: label in output, got: %s", out)
		}
		if !strings.Contains(out, "restaurants") {
			t.Errorf("expected parent alias 'restaurants' in output, got: %s", out)
		}
	})

	t.Run("json output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "categories", "get", "--alias=pizza", "--json"})
			_ = root.Execute()
		})
		if !strings.Contains(out, `"category"`) {
			t.Errorf("expected category field in JSON output, got: %s", out)
		}
		if !strings.Contains(out, "pizza") {
			t.Errorf("expected 'pizza' in JSON output, got: %s", out)
		}
	})
}

func TestCategoryAliases(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	// Test "cat" alias
	out := captureStdout(t, func() {
		root.SetArgs([]string{"yelp", "cat", "get", "--alias=pizza"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "Pizza") {
		t.Errorf("expected 'Pizza' title via 'cat' alias, got: %s", out)
	}

	// Test "category" alias
	out = captureStdout(t, func() {
		root.SetArgs([]string{"yelp", "category", "list"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "restaurants") {
		t.Errorf("expected 'restaurants' via 'category' alias, got: %s", out)
	}
}

func TestCategoryListEmpty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/categories", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
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
		root.SetArgs([]string{"yelp", "categories", "list"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "No categories found.") {
		t.Errorf("expected 'No categories found.' message, got: %s", out)
	}
}
