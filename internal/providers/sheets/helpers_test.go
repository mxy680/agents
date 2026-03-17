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
