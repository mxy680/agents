package cloudflare

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newIPsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("ips",
		newIPsListCmd(factory),
	)
}

func TestIPsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIPsTestCmd(factory))
	output := runCmd(t, root, "ips", "list")

	mustContain(t, output, "IPv4 CIDRs")
	mustContain(t, output, "103.21.244.0/22")
	mustContain(t, output, "103.22.200.0/22")
	mustContain(t, output, "IPv6 CIDRs")
	mustContain(t, output, "2400:cb00::/32")
	mustContain(t, output, "2606:4700::/32")
}

func TestIPsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIPsTestCmd(factory))
	output := runCmd(t, root, "ips", "list", "--json")

	var ranges IPRanges
	err := json.Unmarshal([]byte(output), &ranges)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Len(t, ranges.IPv4CIDRs, 2)
	assert.Contains(t, ranges.IPv4CIDRs, "103.21.244.0/22")
	assert.Contains(t, ranges.IPv4CIDRs, "103.22.200.0/22")
	assert.Len(t, ranges.IPv6CIDRs, 2)
	assert.Contains(t, ranges.IPv6CIDRs, "2400:cb00::/32")
	assert.Contains(t, ranges.IPv6CIDRs, "2606:4700::/32")
}

func newIPsWithEthereumMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/ips", func(w http.ResponseWriter, r *http.Request) {
		result := map[string]any{
			"ipv4_cidrs":   []any{"103.21.244.0/22"},
			"ipv6_cidrs":   []any{"2400:cb00::/32"},
			"ethereum_ips": []any{"198.41.128.0/20"},
		}
		writeJSON(w, cfEnvelope(result))
	})
	return httptest.NewServer(mux)
}

func TestIPsList_WithEthereumIPs_Text(t *testing.T) {
	server := newIPsWithEthereumMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIPsTestCmd(factory))
	output := runCmd(t, root, "ips", "list")

	mustContain(t, output, "Ethereum IPs")
	mustContain(t, output, "198.41.128.0/20")
}

func TestIPsList_WithEthereumIPs_JSON(t *testing.T) {
	server := newIPsWithEthereumMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newIPsTestCmd(factory))
	output := runCmd(t, root, "ips", "list", "--json")

	var ranges IPRanges
	err := json.Unmarshal([]byte(output), &ranges)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Len(t, ranges.EthereumIPs, 1)
	assert.Contains(t, ranges.EthereumIPs, "198.41.128.0/20")
}
