package drive

import (
	"testing"

	api "google.golang.org/api/drive/v3"
)

func TestToFileSummary(t *testing.T) {
	file := &api.File{
		Id:           "f1",
		Name:         "test.txt",
		MimeType:     "text/plain",
		Size:         1024,
		ModifiedTime: "2026-03-16T10:00:00Z",
		Parents:      []string{"root"},
		Trashed:      false,
	}
	s := toFileSummary(file)
	if s.ID != "f1" {
		t.Errorf("expected ID=f1, got %s", s.ID)
	}
	if s.Name != "test.txt" {
		t.Errorf("expected Name=test.txt, got %s", s.Name)
	}
	if s.Size != 1024 {
		t.Errorf("expected Size=1024, got %d", s.Size)
	}
}

func TestToFileDetail(t *testing.T) {
	file := &api.File{
		Id:           "f1",
		Name:         "test.txt",
		MimeType:     "text/plain",
		Size:         2048,
		ModifiedTime: "2026-03-16T10:00:00Z",
		CreatedTime:  "2026-03-01T10:00:00Z",
		Description:  "A test file",
		WebViewLink:  "https://example.com/view",
		Shared:       true,
		Owners: []*api.User{
			{EmailAddress: "alice@example.com", DisplayName: "Alice"},
		},
	}
	d := toFileDetail(file)
	if d.Description != "A test file" {
		t.Errorf("expected Description='A test file', got %s", d.Description)
	}
	if d.WebViewLink != "https://example.com/view" {
		t.Errorf("expected WebViewLink, got %s", d.WebViewLink)
	}
	if len(d.Owners) != 1 || d.Owners[0].Email != "alice@example.com" {
		t.Errorf("expected owner alice@example.com, got %v", d.Owners)
	}
}

func TestToPermissionInfo(t *testing.T) {
	perm := &api.Permission{
		Id:           "p1",
		Role:         "writer",
		Type:         "user",
		EmailAddress: "bob@example.com",
		DisplayName:  "Bob",
	}
	info := toPermissionInfo(perm)
	if info.Role != "writer" {
		t.Errorf("expected Role=writer, got %s", info.Role)
	}
	if info.EmailAddress != "bob@example.com" {
		t.Errorf("expected Email=bob@example.com, got %s", info.EmailAddress)
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "-"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{2684354560, "2.5 GB"},
	}
	for _, tt := range tests {
		got := formatSize(tt.bytes)
		if got != tt.want {
			t.Errorf("formatSize(%d) = %s, want %s", tt.bytes, got, tt.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hi", 2, "hi"},
		{"abcdef", 5, "ab..."},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestConfirmDestructive(t *testing.T) {
	// Without --confirm flag
	cmd := newTestRootCmd()
	cmd.Flags().Bool("confirm", false, "")
	if err := confirmDestructive(cmd); err == nil {
		t.Error("expected error without --confirm")
	}

	// With --confirm flag set
	cmd2 := newTestRootCmd()
	cmd2.Flags().Bool("confirm", false, "")
	cmd2.Flags().Set("confirm", "true")
	if err := confirmDestructive(cmd2); err != nil {
		t.Errorf("unexpected error with --confirm: %v", err)
	}
}
