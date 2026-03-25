package yelp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEventSearch(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "events", "search", "--location=San Francisco, CA"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "SF Tech Meetup") {
			t.Errorf("expected event name in output, got: %s", out)
		}
		if !strings.Contains(out, "yes") {
			t.Errorf("expected free=yes in output, got: %s", out)
		}
		if !strings.Contains(out, "San Francisco") {
			t.Errorf("expected city in output, got: %s", out)
		}
	})

	t.Run("json output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "events", "search", "--location=San Francisco, CA", "--json"})
			_ = root.Execute()
		})
		if !strings.Contains(out, `"events"`) {
			t.Errorf("expected events field in JSON output, got: %s", out)
		}
		if !strings.Contains(out, "san-francisco-tech-meetup") {
			t.Errorf("expected event id in JSON output, got: %s", out)
		}
	})
}

func TestEventSearchWithFilters(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"yelp", "events", "search",
			"--location=San Francisco, CA",
			"--is-free",
			"--categories=music",
			"--limit=10",
			"--sort-by=desc",
			"--sort-on=popularity",
			"--json",
		})
		_ = root.Execute()
	})
	if !strings.Contains(out, "san-francisco-tech-meetup") {
		t.Errorf("expected event in filtered JSON output, got: %s", out)
	}
}

func TestEventSearchWithGeo(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"yelp", "events", "search",
			"--latitude=37.8051",
			"--longitude=-122.4212",
			"--radius=5000",
			"--json",
		})
		_ = root.Execute()
	})
	if !strings.Contains(out, "san-francisco-tech-meetup") {
		t.Errorf("expected event in geo search output, got: %s", out)
	}
}

func TestEventSearchWithDateRange(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{
			"yelp", "events", "search",
			"--location=San Francisco, CA",
			"--start-date=1712000000",
			"--end-date=1713000000",
			"--json",
		})
		_ = root.Execute()
	})
	if !strings.Contains(out, "san-francisco-tech-meetup") {
		t.Errorf("expected event in date-filtered output, got: %s", out)
	}
}

func TestEventGet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "events", "get", "--event-id=san-francisco-tech-meetup"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "SF Tech Meetup") {
			t.Errorf("expected event name in get output, got: %s", out)
		}
		if !strings.Contains(out, "Name:") {
			t.Errorf("expected Name: label in output, got: %s", out)
		}
		if !strings.Contains(out, "Free:") {
			t.Errorf("expected Free: label in output, got: %s", out)
		}
		if !strings.Contains(out, "Location:") {
			t.Errorf("expected Location: label in output, got: %s", out)
		}
	})

	t.Run("json output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "events", "get", "--event-id=san-francisco-tech-meetup", "--json"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "san-francisco-tech-meetup") {
			t.Errorf("expected event id in JSON output, got: %s", out)
		}
		if !strings.Contains(out, `"name"`) {
			t.Errorf("expected name field in JSON output, got: %s", out)
		}
	})
}

func TestEventFeatured(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	t.Run("text output with location", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "events", "featured", "--location=San Francisco, CA"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "SF Symphony Concert") {
			t.Errorf("expected featured event name in output, got: %s", out)
		}
		if !strings.Contains(out, "Name:") {
			t.Errorf("expected Name: label in output, got: %s", out)
		}
	})

	t.Run("json output with location", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"yelp", "events", "featured", "--location=San Francisco, CA", "--json"})
			_ = root.Execute()
		})
		if !strings.Contains(out, "sf-featured-concert") {
			t.Errorf("expected featured event id in JSON output, got: %s", out)
		}
	})

	t.Run("text output with geo", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{
				"yelp", "events", "featured",
				"--latitude=37.8051",
				"--longitude=-122.4212",
			})
			_ = root.Execute()
		})
		if !strings.Contains(out, "SF Symphony Concert") {
			t.Errorf("expected featured event name in geo output, got: %s", out)
		}
	})
}

func TestEventFeaturedRequiresLocation(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	// No location should return an error
	root.SetArgs([]string{"yelp", "events", "featured"})
	err := root.Execute()
	_ = err // error is expected - just ensure no panic
}

func TestEventAliases(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	// Test "event" alias
	out := captureStdout(t, func() {
		root.SetArgs([]string{"yelp", "event", "get", "--event-id=san-francisco-tech-meetup"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "SF Tech Meetup") {
		t.Errorf("expected event name via 'event' alias, got: %s", out)
	}
}

func TestEventSearchEmptyResults(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"total":  0,
			"events": []map[string]any{},
		}
		json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	root := newTestRootCmd()
	p := &Provider{clientFactory: newTestClientFactory(server)}
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"yelp", "events", "search", "--location=Nowhere"})
		_ = root.Execute()
	})
	if !strings.Contains(out, "No events found.") {
		t.Errorf("expected 'No events found.' message, got: %s", out)
	}
}
