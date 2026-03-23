package zillow

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSchoolsNearby(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	t.Run("text", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "schools", "nearby", "--zpid=12345678"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if !strings.Contains(out, "12345678") {
			t.Errorf("expected ZPID in output, got: %s", out)
		}
		if !strings.Contains(out, "Lincoln Elementary") {
			t.Errorf("expected school name in output, got: %s", out)
		}
		if !strings.Contains(out, "8/10") {
			t.Errorf("expected school rating in output, got: %s", out)
		}
		if !strings.Contains(out, "Elementary") {
			t.Errorf("expected school level in output, got: %s", out)
		}
		if !strings.Contains(out, "Public") {
			t.Errorf("expected school type in output, got: %s", out)
		}
		if !strings.Contains(out, "0.3") {
			t.Errorf("expected school distance in output, got: %s", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "schools", "nearby", "--zpid=12345678", "--json"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		var results []SchoolSummary
		if err := json.Unmarshal([]byte(out), &results); err != nil {
			t.Fatalf("expected valid JSON array, got: %s, error: %v", out, err)
		}
		if len(results) == 0 {
			t.Errorf("expected at least one school")
		}
		if results[0].Name != "Lincoln Elementary" {
			t.Errorf("expected school name 'Lincoln Elementary', got: %s", results[0].Name)
		}
		if results[0].Rating != 8 {
			t.Errorf("expected rating 8, got: %d", results[0].Rating)
		}
		if results[0].Level != "Elementary" {
			t.Errorf("expected level 'Elementary', got: %s", results[0].Level)
		}
		if results[0].Type != "Public" {
			t.Errorf("expected type 'Public', got: %s", results[0].Type)
		}
		if results[0].Distance != 0.3 {
			t.Errorf("expected distance 0.3, got: %f", results[0].Distance)
		}
	})

	t.Run("with_limit", func(t *testing.T) {
		root := newTestRootCmd()
		p := &Provider{ClientFactory: newTestClientFactory(server)}
		p.RegisterCommands(root)
		out := captureStdout(t, func() {
			root.SetArgs([]string{"zillow", "schools", "nearby", "--zpid=12345678", "--limit=1"})
			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		// With limit=1 and 1 school in mock, should still show Lincoln Elementary
		if !strings.Contains(out, "Lincoln Elementary") {
			t.Errorf("expected school name in output, got: %s", out)
		}
	})
}

func TestFetchNearbySchools(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	client := newTestClient(server)

	schools, err := fetchNearbySchools(t.Context(), client, "12345678")
	if err != nil {
		t.Fatalf("fetchNearbySchools: %v", err)
	}
	if len(schools) == 0 {
		t.Fatal("expected at least one school")
	}
	if schools[0].Name != "Lincoln Elementary" {
		t.Errorf("expected 'Lincoln Elementary', got %s", schools[0].Name)
	}
	if schools[0].Rating != 8 {
		t.Errorf("expected rating 8, got %d", schools[0].Rating)
	}
}
