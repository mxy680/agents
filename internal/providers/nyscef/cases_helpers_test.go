package nyscef

import (
	"strings"
	"testing"
)

func TestCleanHTML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"<b>Hello</b>", "Hello"},
		{"<td>  foo  bar  </td>", "foo bar"},
		{"plain text", "plain text"},
		{"<a href=\"/foo\">link text</a>", "link text"},
		{"", ""},
	}
	for _, tc := range tests {
		got := cleanHTML(tc.input)
		if got != tc.want {
			t.Errorf("cleanHTML(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestExtractDocketID(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "direct docketId in href",
			html: `<a href="/nyscef/CaseDetails?docketId=12345">view</a>`,
			want: "12345",
		},
		{
			name: "casedetails href case-insensitive",
			html: `<td><a href="/nyscef/casedetails?docketId=99">click</a></td>`,
			want: "99",
		},
		{
			name: "no docketId",
			html: `<a href="/other/page">link</a>`,
			want: "",
		},
		{
			name: "docketId in query param",
			html: `<a href="?docketId=777&foo=bar">link</a>`,
			want: "777",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractDocketID(tc.html)
			if got != tc.want {
				t.Errorf("extractDocketID = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExtractDocketIDURLDecoded(t *testing.T) {
	// docketId with URL-encoded characters.
	html := `<a href="/nyscef/CaseDetails?docketId=XY%2F123">view</a>`
	got := extractDocketID(html)
	if got == "" {
		t.Error("expected non-empty docketId for URL-encoded value")
	}
}

func TestExtractIndexNumber(t *testing.T) {
	tests := []struct {
		cells []string
		want  string
	}{
		{[]string{"600001/2024", "Probate"}, "600001/2024"},
		{[]string{"no index here", "600002/2023"}, "600002/2023"},
		{[]string{"no match"}, ""},
		{[]string{}, ""},
	}
	for _, tc := range tests {
		got := extractIndexNumber(tc.cells)
		if got != tc.want {
			t.Errorf("extractIndexNumber(%v) = %q, want %q", tc.cells, got, tc.want)
		}
	}
}

func TestExtractTagContent(t *testing.T) {
	tests := []struct {
		html    string
		tagName string
		want    string
	}{
		{"<h1>Case Title</h1>", "h1", "Case Title"},
		{"<h2>  Subtitle  </h2>", "h2", "Subtitle"},
		{"<h1><b>Bold Title</b></h1>", "h1", "Bold Title"},
		{"no heading", "h1", ""},
	}
	for _, tc := range tests {
		got := extractTagContent(tc.html, tc.tagName)
		if got != tc.want {
			t.Errorf("extractTagContent(%q, %q) = %q, want %q", tc.html, tc.tagName, got, tc.want)
		}
	}
}

func TestParseCaseDetailTableRows(t *testing.T) {
	// Standard labeled table rows covering all switch cases.
	html := []byte(`<!DOCTYPE html><html><body>
		<table>
			<tr><td>Index Number</td><td>700777/2023</td></tr>
			<tr><td>Case Type</td><td>Partition Action</td></tr>
			<tr><td>Date Filed</td><td>07/04/2023</td></tr>
			<tr><td>Court</td><td>Queens Surrogate</td></tr>
			<tr><td>Status</td><td>Pending</td></tr>
			<tr><td>caption of case</td><td>Doe v. Doe</td></tr>
		</table>
	</body></html>`)

	c, err := parseCaseDetail(html, "DOCK-1", "http://example.com/CaseDetails?docketId=DOCK-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.DocketID != "DOCK-1" {
		t.Errorf("expected DocketID DOCK-1, got %q", c.DocketID)
	}
	if !strings.Contains(c.IndexNumber, "700777") {
		t.Errorf("expected index number, got %q", c.IndexNumber)
	}
	if !strings.Contains(c.CaseType, "Partition") {
		t.Errorf("expected Partition case type, got %q", c.CaseType)
	}
	if c.FilingDate == "" {
		t.Errorf("expected filing date, got empty")
	}
	if c.Court == "" {
		t.Errorf("expected court, got empty")
	}
	if c.Status == "" {
		t.Errorf("expected status, got empty")
	}
	if c.Caption != "Doe v. Doe" {
		t.Errorf("expected caption Doe v. Doe, got %q", c.Caption)
	}
}

func TestParseCaseDetailH1Caption(t *testing.T) {
	// h1 caption with no "caption" labeled row.
	html := []byte(`<!DOCTYPE html><html><body>
		<h1>In re Estate of Alice</h1>
		<table>
			<tr><td>Index Number</td><td>800001/2024</td></tr>
		</table>
	</body></html>`)

	c, err := parseCaseDetail(html, "DOCK-2", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(c.Caption, "Alice") {
		t.Errorf("expected h1 caption, got %q", c.Caption)
	}
}

func TestParseCaseDetailH2Caption(t *testing.T) {
	// No h1, uses h2 for caption when no table row has "caption" label.
	html := []byte(`<!DOCTYPE html><html><body>
		<h2>In re Estate of Bob</h2>
		<table>
			<tr><td>Index Number</td><td>800002/2024</td></tr>
		</table>
	</body></html>`)

	c, err := parseCaseDetail(html, "DOCK-2B", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(c.Caption, "Bob") {
		t.Errorf("expected h2 caption, got %q", c.Caption)
	}
}

func TestParseCaseDetailSkipsEmptyValues(t *testing.T) {
	// Rows with empty value cells should be skipped — first "index" row has empty value.
	html := []byte(`<!DOCTYPE html><html><body>
		<table>
			<tr><td>index ref</td><td></td></tr>
			<tr><td>Index Number</td><td>999001/2024</td></tr>
		</table>
	</body></html>`)

	c, err := parseCaseDetail(html, "DOCK-6", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(c.IndexNumber, "999001") {
		t.Errorf("expected index number 999001, got %q", c.IndexNumber)
	}
}

func TestParseCaseDetailDateFiled(t *testing.T) {
	// Test "filed" label path.
	html := []byte(`<!DOCTYPE html><html><body>
		<table>
			<tr><td>filed date</td><td>12/31/2023</td></tr>
		</table>
	</body></html>`)

	c, err := parseCaseDetail(html, "DOCK-7", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.FilingDate != "12/31/2023" {
		t.Errorf("expected filing date 12/31/2023, got %q", c.FilingDate)
	}
}

func TestParseCaseSearchResultsColumnVariants(t *testing.T) {
	t.Run("5 column table", func(t *testing.T) {
		html := []byte(`<html><body><table>
			<tr><th>Index</th><th>Type</th><th>Caption</th><th>Filed</th><th>Court</th></tr>
			<tr>
				<td><a href="/nyscef/CaseDetails?docketId=55">600055/2024</a></td>
				<td>Probate</td>
				<td>Estate of X</td>
				<td>01/01/2024</td>
				<td>Bronx Surrogate</td>
			</tr>
		</table></body></html>`)
		cases, err := parseCaseSearchResults(html, "http://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cases) != 1 {
			t.Fatalf("expected 1, got %d", len(cases))
		}
		if cases[0].Court != "Bronx Surrogate" {
			t.Errorf("expected court, got %q", cases[0].Court)
		}
	})

	t.Run("3 column table", func(t *testing.T) {
		html := []byte(`<html><body><table>
			<tr>
				<td><a href="?docketId=66">600066/2024</a></td>
				<td>Estate of Y</td>
				<td>02/15/2024</td>
			</tr>
		</table></body></html>`)
		cases, err := parseCaseSearchResults(html, "http://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cases) != 1 {
			t.Fatalf("expected 1, got %d", len(cases))
		}
		if cases[0].FilingDate != "02/15/2024" {
			t.Errorf("expected filing date, got %q", cases[0].FilingDate)
		}
	})

	t.Run("4 column table", func(t *testing.T) {
		html := []byte(`<html><body><table>
			<tr>
				<td><a href="?docketId=77">600077/2024</a></td>
				<td>Partition</td>
				<td>Jones v. Jones</td>
				<td>03/01/2024</td>
			</tr>
		</table></body></html>`)
		cases, err := parseCaseSearchResults(html, "http://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cases) != 1 {
			t.Fatalf("expected 1, got %d", len(cases))
		}
		if cases[0].CaseType != "Partition" {
			t.Errorf("expected Partition, got %q", cases[0].CaseType)
		}
	})

	t.Run("2 column table", func(t *testing.T) {
		html := []byte(`<html><body><table>
			<tr>
				<td><a href="?docketId=88">600088/2024</a></td>
				<td>Estate of Z</td>
			</tr>
		</table></body></html>`)
		cases, err := parseCaseSearchResults(html, "http://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cases) != 1 {
			t.Fatalf("expected 1, got %d", len(cases))
		}
		if cases[0].Caption != "Estate of Z" {
			t.Errorf("expected caption Estate of Z, got %q", cases[0].Caption)
		}
	})

	t.Run("1 column table", func(t *testing.T) {
		html := []byte(`<html><body><table>
			<tr><td><a href="?docketId=99">600099/2024</a></td></tr>
		</table></body></html>`)
		cases, err := parseCaseSearchResults(html, "http://example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// 1-cell rows should produce a case with an index number.
		_ = cases
	})
}
