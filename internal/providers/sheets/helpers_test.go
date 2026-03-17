package sheets

import (
	"strings"
	"testing"
)

func TestParseValuesJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		rows    int
	}{
		{
			name:  "valid 2D array",
			input: `[["a","b"],["c","d"]]`,
			rows:  2,
		},
		{
			name:  "single cell",
			input: `[["hello"]]`,
			rows:  1,
		},
		{
			name:  "numbers and strings",
			input: `[["Name",42],["Age",30]]`,
			rows:  2,
		},
		{
			name:  "empty array",
			input: `[]`,
			rows:  0,
		},
		{
			name:    "invalid JSON",
			input:   `not json`,
			wantErr: true,
		},
		{
			name:    "not an array",
			input:   `{"key":"value"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseValuesJSON(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.rows {
				t.Errorf("expected %d rows, got %d", tt.rows, len(got))
			}
		})
	}
}

func TestParseBatchData(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		entries int
	}{
		{
			name:    "valid batch data",
			input:   `[{"range":"Sheet1!A1","values":[["a"]]},{"range":"Sheet2!B1","values":[["b"]]}]`,
			entries: 2,
		},
		{
			name:    "empty array",
			input:   `[]`,
			entries: 0,
		},
		{
			name:    "invalid JSON",
			input:   `not json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBatchData(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.entries {
				t.Errorf("expected %d entries, got %d", tt.entries, len(got))
			}
		})
	}
}

func TestValidateValueInput(t *testing.T) {
	if err := validateValueInput("RAW"); err != nil {
		t.Errorf("RAW should be valid: %v", err)
	}
	if err := validateValueInput("USER_ENTERED"); err != nil {
		t.Errorf("USER_ENTERED should be valid: %v", err)
	}
	if err := validateValueInput("INVALID"); err == nil {
		t.Error("INVALID should be rejected")
	}
}

func TestValidateMajorDimension(t *testing.T) {
	if err := validateMajorDimension("ROWS"); err != nil {
		t.Errorf("ROWS should be valid: %v", err)
	}
	if err := validateMajorDimension("COLUMNS"); err != nil {
		t.Errorf("COLUMNS should be valid: %v", err)
	}
	if err := validateMajorDimension("DIAGONAL"); err == nil {
		t.Error("DIAGONAL should be rejected")
	}
}

func TestSplitAndTrimRanges(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"A1:B2,C1:D2", []string{"A1:B2", "C1:D2"}},
		{"A1:B2, C1:D2", []string{"A1:B2", "C1:D2"}},
		{" A1:B2 , C1:D2 ", []string{"A1:B2", "C1:D2"}},
		{"A1:B2,,C1:D2", []string{"A1:B2", "C1:D2"}},
		{"A1:B2", []string{"A1:B2"}},
	}
	for _, tt := range tests {
		got := splitAndTrimRanges(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitAndTrimRanges(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitAndTrimRanges(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestFormatCellsTable(t *testing.T) {
	tests := []struct {
		name   string
		values [][]interface{}
		want   string
	}{
		{
			name:   "empty",
			values: nil,
			want:   "(empty)",
		},
		{
			name:   "single cell",
			values: [][]interface{}{{"hello"}},
			want:   "hello",
		},
		{
			name: "two rows aligned",
			values: [][]interface{}{
				{"Name", "Age"},
				{"Alice", 30},
			},
			want: "Name   Age\nAlice  30",
		},
		{
			name: "ragged rows",
			values: [][]interface{}{
				{"a", "b", "c"},
				{"d"},
			},
			want: "a  b  c\nd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := formatCellsTable(tt.values)
			got := strings.Join(lines, "\n")
			// Trim trailing whitespace from each line for comparison
			gotLines := strings.Split(got, "\n")
			for i, l := range gotLines {
				gotLines[i] = strings.TrimRight(l, " ")
			}
			got = strings.Join(gotLines, "\n")

			if got != tt.want {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}
