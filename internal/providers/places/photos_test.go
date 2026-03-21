package places

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPhotosListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	listCmd := newPhotosListCmd(factory)
	root.AddCommand(listCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"list", "--place-id=ChIJ1", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var photos []PhotoReference
	if err := json.Unmarshal([]byte(out), &photos); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if len(photos) != 1 {
		t.Fatalf("expected 1 photo, got %d", len(photos))
	}
	if photos[0].Name != "places/ChIJ1/photos/abc123" {
		t.Errorf("photo name = %q", photos[0].Name)
	}
	if photos[0].WidthPx != 4032 {
		t.Errorf("photo width = %d, want 4032", photos[0].WidthPx)
	}
}

func TestPhotosListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	listCmd := newPhotosListCmd(factory)
	root.AddCommand(listCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"list", "--place-id=ChIJ1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "abc123") {
		t.Errorf("expected photo name in output, got: %s", out)
	}
	if !strings.Contains(out, "4032x3024") {
		t.Errorf("expected dimensions in output, got: %s", out)
	}
}

func TestPhotosGetURLOnly(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	getCmd := newPhotosGetCmd(factory)
	root.AddCommand(getCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"get",
			"--photo-name=places/ChIJ1/photos/abc123",
			"--url-only",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var result PhotoMedia
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if result.PhotoURI != "https://lh3.googleusercontent.com/places/photo123" {
		t.Errorf("photoUri = %q", result.PhotoURI)
	}
}

func TestPhotosGetURLOnlyText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	getCmd := newPhotosGetCmd(factory)
	root.AddCommand(getCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"get",
			"--photo-name=places/ChIJ1/photos/abc123",
			"--url-only",
		})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "googleusercontent.com") {
		t.Errorf("expected photo URL in output, got: %s", out)
	}
}

func TestPhotosGetMissingPhotoName(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	getCmd := newPhotosGetCmd(factory)
	root.AddCommand(getCmd)

	root.SetArgs([]string{"get"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing --photo-name")
	}
}

func TestPhotosListMissingPlaceID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	listCmd := newPhotosListCmd(factory)
	root.AddCommand(listCmd)

	root.SetArgs([]string{"list"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing --place-id")
	}
}
