package calendar

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFreebusyQueryJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestFreebusyCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"freebusy", "query",
			"--time-min=2026-03-16T00:00:00Z",
			"--time-max=2026-03-17T00:00:00Z",
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}

	var results []FreeBusyResult
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 calendar result, got %d", len(results))
	}
	if len(results[0].Busy) != 2 {
		t.Errorf("expected 2 busy slots, got %d", len(results[0].Busy))
	}
}

func TestFreebusyQueryText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestFreebusyCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"freebusy", "query",
			"--time-min=2026-03-16T00:00:00Z",
			"--time-max=2026-03-17T00:00:00Z",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected non-empty text output")
	}
}

func TestFreebusyQueryEmpty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/freeBusy", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"calendars": map[string]any{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestFreebusyCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"freebusy", "query",
			"--time-min=2026-03-16T00:00:00Z",
			"--time-max=2026-03-17T00:00:00Z",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if output == "" {
		t.Error("expected output for empty result")
	}
}

func TestFreebusyQueryMultipleCalendars(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestFreebusyCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"freebusy", "query",
			"--calendar-ids=primary,work@group.calendar.google.com",
			"--time-min=2026-03-16T00:00:00Z",
			"--time-max=2026-03-17T00:00:00Z",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
}
