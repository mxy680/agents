package cloudflare

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newFirewallTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("firewall",
		newFirewallListCmd(factory),
		newFirewallCreateCmd(factory),
		newFirewallDeleteCmd(factory),
	)
}

func TestFirewallList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newFirewallTestCmd(factory))
	output := runCmd(t, root, "firewall", "list", "--zone", testZoneID)

	mustContain(t, output, "rule_abc1")
	mustContain(t, output, "block")
	mustContain(t, output, "rule_def2")
	mustContain(t, output, "challenge")
}

func TestFirewallList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newFirewallTestCmd(factory))
	output := runCmd(t, root, "firewall", "list", "--zone", testZoneID, "--json")

	var rules []FirewallRuleSummary
	err := json.Unmarshal([]byte(output), &rules)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Len(t, rules, 2)
	assert.Equal(t, "rule_abc1", rules[0].ID)
	assert.Equal(t, "block", rules[0].Action)
	assert.Equal(t, "Block known bots", rules[0].Description)
	assert.False(t, rules[0].Paused)
	assert.Equal(t, "rule_def2", rules[1].ID)
	assert.Equal(t, "challenge", rules[1].Action)
	assert.True(t, rules[1].Paused)
}

func TestFirewallCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newFirewallTestCmd(factory))
	output := runCmd(t, root, "firewall", "create",
		"--zone", testZoneID,
		"--action", "block",
		"--expression", `(ip.geoip.country eq "RU")`,
		"--description", "block bad actors",
	)

	mustContain(t, output, "Created firewall rule")
	mustContain(t, output, "rule_new1")
	mustContain(t, output, "block")
}

func TestFirewallCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newFirewallTestCmd(factory))
	output := runCmd(t, root, "firewall", "create",
		"--zone", testZoneID,
		"--action", "block",
		"--expression", `(ip.geoip.country eq "RU")`,
		"--json",
	)

	var rule FirewallRuleSummary
	err := json.Unmarshal([]byte(output), &rule)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "rule_new1", rule.ID)
	assert.Equal(t, "block", rule.Action)
}

func TestFirewallCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newFirewallTestCmd(factory))
	output := runCmd(t, root, "firewall", "create",
		"--zone", testZoneID,
		"--action", "block",
		"--expression", `(ip.geoip.country eq "RU")`,
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "block")
}

func TestFirewallDelete_Confirm_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newFirewallTestCmd(factory))
	output := runCmd(t, root, "firewall", "delete",
		"--zone", testZoneID,
		"--rule", "rule_abc1",
		"--confirm",
	)

	mustContain(t, output, "Deleted firewall rule")
	mustContain(t, output, "rule_abc1")
}

func TestFirewallDelete_Confirm_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newFirewallTestCmd(factory))
	output := runCmd(t, root, "firewall", "delete",
		"--zone", testZoneID,
		"--rule", "rule_abc1",
		"--confirm",
		"--json",
	)

	var result map[string]string
	err := json.Unmarshal([]byte(output), &result)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, "deleted", result["status"])
	assert.Equal(t, "rule_abc1", result["rule_id"])
}

func TestFirewallDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newFirewallTestCmd(factory))
	output := runCmd(t, root, "firewall", "delete",
		"--zone", testZoneID,
		"--rule", "rule_abc1",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "rule_abc1")
}

func TestFirewallDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newFirewallTestCmd(factory))

	err := runCmdErr(t, root, "firewall", "delete",
		"--zone", testZoneID,
		"--rule", "rule_abc1",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}
