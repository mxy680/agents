package cloudflare

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newCertsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("certs",
		newCertsListCmd(factory),
		newCertsGetCmd(factory),
	)
}

func TestCertsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "list", "--zone", testZoneID)

	mustContain(t, output, "cert_abc1")
	mustContain(t, output, "advanced")
	mustContain(t, output, "active")
	mustContain(t, output, "cert_def2")
	mustContain(t, output, "universal")
}

func TestCertsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "list", "--zone", testZoneID, "--json")

	var packs []CertPackSummary
	err := json.Unmarshal([]byte(output), &packs)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Len(t, packs, 2)
	assert.Equal(t, "cert_abc1", packs[0].ID)
	assert.Equal(t, "advanced", packs[0].Type)
	assert.Equal(t, "active", packs[0].Status)
	assert.Contains(t, packs[0].Hosts, "example.com")
	assert.Contains(t, packs[0].Hosts, "*.example.com")
	assert.Equal(t, "cert_def2", packs[1].ID)
	assert.Equal(t, "universal", packs[1].Type)
}

func TestCertsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "get", "--zone", testZoneID, "--cert", "cert_abc1")

	mustContain(t, output, "cert_abc1")
	mustContain(t, output, "advanced")
	mustContain(t, output, "active")
	mustContain(t, output, "example.com")
}

func TestCertsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "get", "--zone", testZoneID, "--cert", "cert_abc1", "--json")

	var pack CertPackSummary
	err := json.Unmarshal([]byte(output), &pack)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "cert_abc1", pack.ID)
	assert.Equal(t, "advanced", pack.Type)
	assert.Equal(t, "active", pack.Status)
	assert.Contains(t, pack.Hosts, "example.com")
	assert.Contains(t, pack.Hosts, "*.example.com")
}
