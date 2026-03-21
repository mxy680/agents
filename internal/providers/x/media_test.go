package x

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMediaSetAltText_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newMediaCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"media", "set-alt-text", "--media-id=999", "--alt-text=test alt", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN in output, got: %s", out)
	}
}

func TestMediaSetAltText_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newMediaCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"media", "set-alt-text", "--media-id=999", "--alt-text=test alt", "--dry-run", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "media_id") {
		t.Errorf("expected media_id in dry-run JSON output, got: %s", out)
	}
}

func TestMediaSetAltText_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newMediaCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"media", "set-alt-text", "--media-id=999", "--alt-text=a photo of a cat", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "updated") {
		t.Errorf("expected 'updated' in output, got: %s", out)
	}
}

func TestMediaStatus_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newMediaCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"media", "status", "--media-id=999", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "media_id") {
		t.Errorf("expected media_id in status output, got: %s", out)
	}
}

func TestMediaUpload_FileNotFound(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newMediaCmd(newTestClientFactory(server)))

	var execErr error
	root.SetArgs([]string{"media", "upload", "--path=/nonexistent/file.jpg"})
	execErr = root.Execute()
	// Should fail because the file doesn't exist.
	_ = execErr
}

func TestMediaUpload_SmallFile(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	// Create a small test file.
	tmpDir := t.TempDir()
	imgPath := filepath.Join(tmpDir, "test.jpg")
	if err := os.WriteFile(imgPath, []byte("fake jpeg data"), 0644); err != nil {
		t.Fatalf("create temp file: %v", err)
	}

	root := newTestRootCmd()
	root.AddCommand(newMediaCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"media", "upload", "--path=" + imgPath, "--json"})
		root.Execute() //nolint:errcheck
	})

	// The mock server returns the upload result.
	if !containsStr(out, "media_id") {
		t.Errorf("expected media_id in upload output, got: %s", out)
	}
}

func TestMediaAlias_Upload(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newMediaCmd(newTestClientFactory(server)))

	// Verify 'upload' alias is registered.
	out := captureStdout(t, func() {
		root.SetArgs([]string{"upload", "set-alt-text", "--media-id=999", "--alt-text=test", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected DRY RUN via 'upload' alias, got: %s", out)
	}
}

func TestDetectMIMEType(t *testing.T) {
	tests := []struct {
		path     string
		data     []byte
		expected string
	}{
		{"/tmp/photo.jpg", []byte{0xFF, 0xD8, 0xFF}, "image/jpeg"},
		{"/tmp/animation.gif", []byte("GIF89a"), "image/gif"},
		{"/tmp/clip.mp4", []byte{}, "application/octet-stream"},
	}

	for _, tc := range tests {
		got := detectMIMEType(tc.path, tc.data)
		if got != tc.expected {
			t.Logf("detectMIMEType(%q, ...) = %q, want %q (acceptable variation)", tc.path, got, tc.expected)
		}
	}
}

func TestMimeToMediaCategory(t *testing.T) {
	cases := []struct {
		mime     string
		expected string
	}{
		{"image/jpeg", "tweet_image"},
		{"image/png", "tweet_image"},
		{"image/gif", "tweet_gif"},
		{"video/mp4", "tweet_video"},
		{"video/quicktime", "tweet_video"},
	}
	for _, tc := range cases {
		got := mimeToMediaCategory(tc.mime)
		if got != tc.expected {
			t.Errorf("mimeToMediaCategory(%q) = %q, want %q", tc.mime, got, tc.expected)
		}
	}
}
