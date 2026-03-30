package cloudflare

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newZonesTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("zones",
		newZonesListCmd(factory),
		newZonesGetCmd(factory),
		newZonesPurgeCacheCmd(factory),
	)
}

func TestZonesList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newZonesTestCmd(factory))
	output := runCmd(t, root, "zones", "list")

	mustContain(t, output, testZoneID)
	mustContain(t, output, "example.com")
	mustContain(t, output, "example.org")
	mustContain(t, output, "active")
	mustContain(t, output, "pending")
}

func TestZonesList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newZonesTestCmd(factory))
	output := runCmd(t, root, "zones", "list", "--json")

	var results []ZoneSummary
	err := json.Unmarshal([]byte(output), &results)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Len(t, results, 2)
	assert.Equal(t, testZoneID, results[0].ID)
	assert.Equal(t, "example.com", results[0].Name)
	assert.Equal(t, "active", results[0].Status)
	assert.Equal(t, "Free Website", results[0].Plan)
	assert.Equal(t, "zone_def456", results[1].ID)
	assert.Equal(t, "Pro", results[1].Plan)
}

func TestZonesGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newZonesTestCmd(factory))
	output := runCmd(t, root, "zones", "get", "--zone", testZoneID)

	mustContain(t, output, testZoneID)
	mustContain(t, output, "example.com")
	mustContain(t, output, "active")
	mustContain(t, output, "Free Website")
	mustContain(t, output, "ns1.cloudflare.com")
}

func TestZonesGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newZonesTestCmd(factory))
	output := runCmd(t, root, "zones", "get", "--zone", testZoneID, "--json")

	var detail ZoneDetail
	err := json.Unmarshal([]byte(output), &detail)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, testZoneID, detail.ID)
	assert.Equal(t, "example.com", detail.Name)
	assert.Equal(t, "active", detail.Status)
	assert.Equal(t, "full", detail.Type)
	assert.False(t, detail.Paused)
	assert.Contains(t, detail.NameServers, "ns1.cloudflare.com")
}

func TestZonesPurgeCache_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newZonesTestCmd(factory))
	output := runCmd(t, root, "zones", "purge-cache", "--zone", testZoneID, "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, testZoneID)
}

func TestZonesPurgeCache_Confirm_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newZonesTestCmd(factory))
	output := runCmd(t, root, "zones", "purge-cache", "--zone", testZoneID, "--confirm")

	mustContain(t, output, "purged")
	mustContain(t, output, testZoneID)
}

func TestZonesPurgeCache_Confirm_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newZonesTestCmd(factory))
	output := runCmd(t, root, "zones", "purge-cache", "--zone", testZoneID, "--confirm", "--json")

	var result map[string]string
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "purged", result["status"])
	assert.Equal(t, testZoneID, result["zone_id"])
}

func TestZonesPurgeCache_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newZonesTestCmd(factory))

	err := runCmdErr(t, root, "zones", "purge-cache", "--zone", testZoneID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}

func TestZonesPurgeCache_WithFiles(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newZonesTestCmd(factory))
	output := runCmd(t, root, "zones", "purge-cache",
		"--zone", testZoneID,
		"--files", "https://example.com/style.css",
		"--confirm",
	)

	mustContain(t, output, "purged")
}
