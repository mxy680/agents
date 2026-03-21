package places

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	getCmd := newGetCmd(factory)
	root.AddCommand(getCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"get", "--place-id=ChIJ1", "--json"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var detail PlaceDetail
	if err := json.Unmarshal([]byte(out), &detail); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if detail.ID != "ChIJ1" {
		t.Errorf("ID = %q, want ChIJ1", detail.ID)
	}
	if detail.Name != "Coffee Corner" {
		t.Errorf("Name = %q, want Coffee Corner", detail.Name)
	}
	if detail.Rating != 4.5 {
		t.Errorf("Rating = %f, want 4.5", detail.Rating)
	}
	if detail.PrimaryType != "cafe" {
		t.Errorf("PrimaryType = %q, want cafe", detail.PrimaryType)
	}
	if !detail.Delivery {
		t.Error("expected Delivery=true")
	}
	if len(detail.Reviews) != 1 {
		t.Errorf("expected 1 review, got %d", len(detail.Reviews))
	}
	if len(detail.Photos) != 1 {
		t.Errorf("expected 1 photo, got %d", len(detail.Photos))
	}
	if len(detail.AddressComponents) != 2 {
		t.Errorf("expected 2 address components, got %d", len(detail.AddressComponents))
	}
}

func TestGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	getCmd := newGetCmd(factory)
	root.AddCommand(getCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"get", "--place-id=ChIJ1"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	expected := []string{
		"Coffee Corner",
		"123 Main St, Cleveland",
		"cafe",
		"4.5",
		"$$",
		"+1 216-555-0100",
		"coffeecorner.example.com",
		"Monday:",
		"Alice",
		"Best coffee in town!",
		"Delivery",
		"Dine-in",
		"Takeout",
	}
	for _, e := range expected {
		if !strings.Contains(out, e) {
			t.Errorf("expected output to contain %q", e)
		}
	}
}

func TestGetWithOptions(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	getCmd := newGetCmd(factory)
	root.AddCommand(getCmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"get", "--place-id=ChIJ1",
			"--lang=en",
			"--region=us",
			"--fields=preferred",
			"--json",
		})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	var detail PlaceDetail
	if err := json.Unmarshal([]byte(out), &detail); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestGetMissingPlaceID(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	getCmd := newGetCmd(factory)
	root.AddCommand(getCmd)

	root.SetArgs([]string{"get"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing --place-id")
	}
}
