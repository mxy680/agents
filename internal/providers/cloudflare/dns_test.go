package cloudflare

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newDNSTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("dns",
		newDNSListCmd(factory),
		newDNSGetCmd(factory),
		newDNSCreateCmd(factory),
		newDNSUpdateCmd(factory),
		newDNSDeleteCmd(factory),
	)
}

func TestDNSList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "list", "--zone", testZoneID)

	mustContain(t, output, "rec_abc1")
	mustContain(t, output, "example.com")
	mustContain(t, output, "1.2.3.4")
	mustContain(t, output, "CNAME")
}

func TestDNSList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "list", "--zone", testZoneID, "--json")

	var records []DNSRecordSummary
	err := json.Unmarshal([]byte(output), &records)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Len(t, records, 2)
	assert.Equal(t, "rec_abc1", records[0].ID)
	assert.Equal(t, "A", records[0].Type)
	assert.Equal(t, "example.com", records[0].Name)
	assert.Equal(t, "1.2.3.4", records[0].Content)
	assert.True(t, records[0].Proxied)
	assert.Equal(t, "rec_def2", records[1].ID)
	assert.Equal(t, "CNAME", records[1].Type)
}

func TestDNSGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "get", "--zone", testZoneID, "--record", "rec_abc1")

	mustContain(t, output, "rec_abc1")
	mustContain(t, output, "A")
	mustContain(t, output, "example.com")
	mustContain(t, output, "1.2.3.4")
}

func TestDNSGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "get", "--zone", testZoneID, "--record", "rec_abc1", "--json")

	var record DNSRecordSummary
	err := json.Unmarshal([]byte(output), &record)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "rec_abc1", record.ID)
	assert.Equal(t, "A", record.Type)
	assert.Equal(t, "example.com", record.Name)
	assert.Equal(t, "1.2.3.4", record.Content)
	assert.True(t, record.Proxied)
}

func TestDNSCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "create",
		"--zone", testZoneID,
		"--type", "A",
		"--name", "api.example.com",
		"--content", "9.9.9.9",
	)

	mustContain(t, output, "Created DNS record")
}

func TestDNSCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "create",
		"--zone", testZoneID,
		"--type", "A",
		"--name", "api.example.com",
		"--content", "9.9.9.9",
		"--json",
	)

	var record DNSRecordSummary
	err := json.Unmarshal([]byte(output), &record)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "rec_new1", record.ID)
	assert.Equal(t, "A", record.Type)
}

func TestDNSCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "create",
		"--zone", testZoneID,
		"--type", "A",
		"--name", "api.example.com",
		"--content", "9.9.9.9",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "api.example.com")
}

func TestDNSUpdate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "update",
		"--zone", testZoneID,
		"--record", "rec_abc1",
		"--content", "5.6.7.8",
	)

	mustContain(t, output, "Updated DNS record")
}

func TestDNSUpdate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "update",
		"--zone", testZoneID,
		"--record", "rec_abc1",
		"--content", "5.6.7.8",
		"--json",
	)

	var record DNSRecordSummary
	err := json.Unmarshal([]byte(output), &record)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "rec_abc1", record.ID)
}

func TestDNSUpdate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "update",
		"--zone", testZoneID,
		"--record", "rec_abc1",
		"--content", "5.6.7.8",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "rec_abc1")
}

func TestDNSDelete_Confirm_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "delete",
		"--zone", testZoneID,
		"--record", "rec_abc1",
		"--confirm",
	)

	mustContain(t, output, "Deleted DNS record")
	mustContain(t, output, "rec_abc1")
}

func TestDNSDelete_Confirm_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "delete",
		"--zone", testZoneID,
		"--record", "rec_abc1",
		"--confirm",
		"--json",
	)

	var result map[string]string
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "deleted", result["status"])
	assert.Equal(t, "rec_abc1", result["record_id"])
}

func TestDNSDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "delete",
		"--zone", testZoneID,
		"--record", "rec_abc1",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "rec_abc1")
}

func TestDNSDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))

	err := runCmdErr(t, root, "dns", "delete",
		"--zone", testZoneID,
		"--record", "rec_abc1",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}
