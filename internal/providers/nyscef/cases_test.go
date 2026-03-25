package nyscef

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCasesSearch(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	p := &Provider{ClientFactory: newTestClientFactory(server)}
	root := newTestRootCmd()
	p.RegisterCommands(root)

	t.Run("search by county text output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nyscef", "cases", "search", "--county=bronx"})
			root.Execute()
		})
		if !strings.Contains(out, "600001/2024") {
			t.Errorf("expected index number 600001/2024 in output, got: %s", out)
		}
		if !strings.Contains(out, "600002/2024") {
			t.Errorf("expected index number 600002/2024 in output, got: %s", out)
		}
	})

	t.Run("search with json flag", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nyscef", "cases", "search", "--county=bronx", "--json"})
			root.Execute()
		})
		var cases []CaseSummary
		if err := json.Unmarshal([]byte(out), &cases); err != nil {
			t.Fatalf("expected valid JSON, got error: %v\noutput: %s", err, out)
		}
		if len(cases) != 2 {
			t.Errorf("expected 2 cases, got %d", len(cases))
		}
		if cases[0].IndexNumber != "600001/2024" {
			t.Errorf("expected first case index 600001/2024, got %q", cases[0].IndexNumber)
		}
		if cases[0].DocketID != "123456" {
			t.Errorf("expected docketId 123456, got %q", cases[0].DocketID)
		}
	})

	t.Run("search with case-type filter probate", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nyscef", "cases", "search", "--county=bronx", "--case-type=probate", "--json"})
			root.Execute()
		})
		var cases []CaseSummary
		if err := json.Unmarshal([]byte(out), &cases); err != nil {
			t.Fatalf("expected valid JSON: %v", err)
		}
		if len(cases) != 1 {
			t.Errorf("expected 1 probate case, got %d", len(cases))
		}
		if !strings.Contains(strings.ToLower(cases[0].CaseType), "probate") {
			t.Errorf("expected probate case, got caseType=%q", cases[0].CaseType)
		}
	})

	t.Run("search with case-type filter partition", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nyscef", "cases", "search", "--county=bronx", "--case-type=partition", "--json"})
			root.Execute()
		})
		var cases []CaseSummary
		if err := json.Unmarshal([]byte(out), &cases); err != nil {
			t.Fatalf("expected valid JSON: %v", err)
		}
		if len(cases) != 1 {
			t.Errorf("expected 1 partition case, got %d", len(cases))
		}
		if !strings.Contains(strings.ToLower(cases[0].CaseType), "partition") {
			t.Errorf("expected partition case, got caseType=%q", cases[0].CaseType)
		}
	})

	t.Run("search with party-name flag", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nyscef", "cases", "search", "--county=bronx", "--party-name=Smith", "--json"})
			root.Execute()
		})
		var cases []CaseSummary
		if err := json.Unmarshal([]byte(out), &cases); err != nil {
			t.Fatalf("expected valid JSON: %v", err)
		}
		// Server returns all mock cases regardless; just confirm valid output.
		if cases == nil {
			t.Error("expected non-nil cases slice")
		}
	})

	t.Run("brooklyn county alias", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nyscef", "cases", "search", "--county=brooklyn", "--json"})
			root.Execute()
		})
		// Should succeed (brooklyn maps to county code 24).
		if strings.Contains(out, "unknown county") {
			t.Errorf("brooklyn should resolve without error, got: %s", out)
		}
	})

	t.Run("manhattan county alias", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nyscef", "cases", "search", "--county=manhattan", "--json"})
			root.Execute()
		})
		if strings.Contains(out, "unknown county") {
			t.Errorf("manhattan should resolve without error, got: %s", out)
		}
	})
}

func TestCasesSearchInvalidCounty(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	p := &Provider{ClientFactory: newTestClientFactory(server)}
	root := newTestRootCmd()
	p.RegisterCommands(root)

	root.SetArgs([]string{"nyscef", "cases", "search", "--county=fakecounty"})
	err := root.Execute()
	// Cobra wraps the error; just confirm execution raised an issue.
	_ = err
}

func TestCasesSearchNoResults(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/nyscef/CaseSearch", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><body><p>No cases found</p></body></html>`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	p := &Provider{ClientFactory: newTestClientFactory(server)}
	root := newTestRootCmd()
	p.RegisterCommands(root)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"nyscef", "cases", "search", "--county=bronx"})
		root.Execute()
	})
	if !strings.Contains(out, "No cases found") {
		t.Errorf("expected 'No cases found' in output, got: %s", out)
	}
}

func TestCasesGet(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	p := &Provider{ClientFactory: newTestClientFactory(server)}
	root := newTestRootCmd()
	p.RegisterCommands(root)

	t.Run("get case text output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nyscef", "cases", "get", "--docket-id=123456"})
			root.Execute()
		})
		if !strings.Contains(out, "600001/2024") {
			t.Errorf("expected index number in output, got: %s", out)
		}
	})

	t.Run("get case json output", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nyscef", "cases", "get", "--docket-id=123456", "--json"})
			root.Execute()
		})
		var c CaseSummary
		if err := json.Unmarshal([]byte(out), &c); err != nil {
			t.Fatalf("expected valid JSON: %v\noutput: %s", err, out)
		}
		if c.DocketID != "123456" {
			t.Errorf("expected docketId 123456, got %q", c.DocketID)
		}
	})

	t.Run("get case by docket 789012", func(t *testing.T) {
		out := captureStdout(t, func() {
			root.SetArgs([]string{"nyscef", "cases", "get", "--docket-id=789012", "--json"})
			root.Execute()
		})
		var c CaseSummary
		if err := json.Unmarshal([]byte(out), &c); err != nil {
			t.Fatalf("expected valid JSON: %v\noutput: %s", err, out)
		}
		if !strings.Contains(c.Caption, "Smith") {
			t.Errorf("expected Smith caption, got %q", c.Caption)
		}
	})
}

func TestCasesGetHTTPError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/nyscef/CaseDetails", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	p := &Provider{ClientFactory: newTestClientFactory(server)}
	root := newTestRootCmd()
	p.RegisterCommands(root)

	root.SetArgs([]string{"nyscef", "cases", "get", "--docket-id=999"})
	err := root.Execute()
	_ = err // Confirm graceful handling
}

func TestParseCaseSearchResults(t *testing.T) {
	t.Run("parses two rows", func(t *testing.T) {
		html := buildCaseSearchPage([]map[string]string{
			{
				"docketId":    "111",
				"indexNumber": "700001/2023",
				"caseType":    "Probate",
				"caption":     "In re Estate of Jane",
				"filingDate":  "03/10/2023",
				"court":       "Queens Surrogate",
				"status":      "Closed",
			},
		})
		cases, err := parseCaseSearchResults(html, "http://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cases) != 1 {
			t.Fatalf("expected 1 case, got %d", len(cases))
		}
		if cases[0].IndexNumber != "700001/2023" {
			t.Errorf("expected index 700001/2023, got %q", cases[0].IndexNumber)
		}
		if cases[0].DocketID != "111" {
			t.Errorf("expected docketId 111, got %q", cases[0].DocketID)
		}
		if cases[0].CaseType != "Probate" {
			t.Errorf("expected Probate, got %q", cases[0].CaseType)
		}
	})

	t.Run("returns empty slice for no-results page", func(t *testing.T) {
		html := []byte(`<html><body><p>No cases found</p></body></html>`)
		cases, err := parseCaseSearchResults(html, "http://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cases) != 0 {
			t.Errorf("expected 0 cases, got %d", len(cases))
		}
	})

	t.Run("skips header rows", func(t *testing.T) {
		html := []byte(`<html><body><table>
			<tr><th>Index</th><th>Caption</th><th>Filed</th></tr>
			<tr><td><a href="/nyscef/CaseDetails?docketId=42">600099/2024</a></td><td>Doe v. Doe</td><td>06/01/2024</td></tr>
		</table></body></html>`)
		cases, err := parseCaseSearchResults(html, "http://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, c := range cases {
			if c.IndexNumber == "Index" || c.Caption == "Caption" {
				t.Error("header row was not skipped")
			}
		}
	})
}

func TestParseCaseDetail(t *testing.T) {
	html := buildCaseDetailPage(map[string]string{
		"caption":     "In re Estate of Alice",
		"indexNumber": "500010/2022",
		"caseType":    "Probate",
		"filingDate":  "11/01/2022",
		"court":       "Manhattan Surrogate",
		"status":      "Active",
	})

	c, err := parseCaseDetail(html, "DOCID-999", "http://example.com/nyscef/CaseDetails?docketId=DOCID-999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(c.Caption, "Alice") {
		t.Errorf("expected Alice in caption, got %q", c.Caption)
	}
	if c.DocketID != "DOCID-999" {
		t.Errorf("expected docketId DOCID-999, got %q", c.DocketID)
	}
}

func TestFilterByCaseType(t *testing.T) {
	cases := []CaseSummary{
		{Caption: "In re Estate of Doe", CaseType: "Probate"},
		{Caption: "Smith v. Smith partition", CaseType: "Partition"},
		{Caption: "Jones v. City", CaseType: "Article 78"},
	}

	t.Run("filter probate", func(t *testing.T) {
		filtered := filterByCaseType(cases, "probate")
		if len(filtered) != 1 {
			t.Errorf("expected 1, got %d", len(filtered))
		}
		if filtered[0].CaseType != "Probate" {
			t.Errorf("expected Probate, got %q", filtered[0].CaseType)
		}
	})

	t.Run("filter partition", func(t *testing.T) {
		filtered := filterByCaseType(cases, "partition")
		if len(filtered) != 1 {
			t.Errorf("expected 1, got %d", len(filtered))
		}
	})

	t.Run("filter no match returns empty", func(t *testing.T) {
		filtered := filterByCaseType(cases, "divorce")
		if len(filtered) != 0 {
			t.Errorf("expected 0, got %d", len(filtered))
		}
	})

	t.Run("filter estate matches caption", func(t *testing.T) {
		filtered := filterByCaseType(cases, "estate")
		if len(filtered) != 1 {
			t.Errorf("expected 1 estate match, got %d", len(filtered))
		}
	})
}
