package fly

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newRegionsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("regions",
		newRegionsListCmd(factory),
	)
}

func TestRegionsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newRegionsTestCmd(factory))
	output := runCmd(t, root, "regions", "list")

	mustContain(t, output, "iad")
	mustContain(t, output, "Ashburn, Virginia (US)")
	mustContain(t, output, "lhr")
	mustContain(t, output, "London, United Kingdom")
	mustContain(t, output, "nrt")
	mustContain(t, output, "Tokyo, Japan")
}

func TestRegionsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newRegionsTestCmd(factory))
	output := runCmd(t, root, "regions", "list", "--json")

	var results []Region
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 regions, got %d", len(results))
	}
	if results[0].Code != "iad" {
		t.Errorf("expected first region Code=iad, got %s", results[0].Code)
	}
	if results[0].Name != "Ashburn, Virginia (US)" {
		t.Errorf("expected first region Name=Ashburn, Virginia (US), got %s", results[0].Name)
	}
	if results[1].Code != "lhr" {
		t.Errorf("expected second region Code=lhr, got %s", results[1].Code)
	}
	if results[2].Code != "nrt" {
		t.Errorf("expected third region Code=nrt, got %s", results[2].Code)
	}
}

func TestRegionsList_HeaderPresent(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newRegionsTestCmd(factory))
	output := runCmd(t, root, "regions", "list")

	// Verify the column header is present
	mustContain(t, output, "CODE")
	mustContain(t, output, "NAME")
}
