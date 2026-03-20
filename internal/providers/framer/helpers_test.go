package framer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"short string unchanged", "hello", 10, "hello"},
		{"exact length unchanged", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"empty string", "", 5, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := truncate(tc.input, tc.maxLen)
			if got != tc.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tc.input, tc.maxLen, got, tc.want)
			}
		})
	}
}

func TestParseStringList(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty string returns nil", "", nil},
		{"single item", "foo", []string{"foo"}},
		{"multiple items", "foo,bar,baz", []string{"foo", "bar", "baz"}},
		{"items with spaces", " foo , bar , baz ", []string{"foo", "bar", "baz"}},
		{"trailing comma ignored", "foo,bar,", []string{"foo", "bar"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseStringList(tc.input)
			if len(got) != len(tc.want) {
				t.Errorf("parseStringList(%q) = %v, want %v", tc.input, got, tc.want)
				return
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("parseStringList(%q)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestParseJSONFlag(t *testing.T) {
	t.Run("valid JSON", func(t *testing.T) {
		result, err := parseJSONFlag(`{"key":"value"}`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(result) != `{"key":"value"}` {
			t.Errorf("unexpected result: %s", result)
		}
	})

	t.Run("empty string returns error", func(t *testing.T) {
		_, err := parseJSONFlag("")
		if err == nil {
			t.Error("expected error for empty string")
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		_, err := parseJSONFlag("not-json")
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("valid JSON array", func(t *testing.T) {
		result, err := parseJSONFlag(`[1,2,3]`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(result) != `[1,2,3]` {
			t.Errorf("unexpected result: %s", result)
		}
	})
}

func TestParseJSONFlagOrFile(t *testing.T) {
	t.Run("from value", func(t *testing.T) {
		result, err := parseJSONFlagOrFile(`{"key":"val"}`, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(result) != `{"key":"val"}` {
			t.Errorf("unexpected result: %s", result)
		}
	})

	t.Run("from file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "data.json")
		if err := os.WriteFile(path, []byte(`{"file":"data"}`), 0644); err != nil {
			t.Fatal(err)
		}

		result, err := parseJSONFlagOrFile("", path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(result) != `{"file":"data"}` {
			t.Errorf("unexpected result: %s", result)
		}
	})
}

func TestReadFileOrFlag(t *testing.T) {
	t.Run("returns value when no file path", func(t *testing.T) {
		got, err := readFileOrFlag("hello", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "hello" {
			t.Errorf("expected 'hello', got: %s", got)
		}
	})

	t.Run("reads from file when file path provided", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(path, []byte("file content"), 0644); err != nil {
			t.Fatal(err)
		}

		got, err := readFileOrFlag("", path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "file content" {
			t.Errorf("expected 'file content', got: %s", got)
		}
	})

	t.Run("missing file returns error", func(t *testing.T) {
		_, err := readFileOrFlag("", "/nonexistent/path/file.txt")
		if err == nil {
			t.Error("expected error for missing file")
		}
	})
}
