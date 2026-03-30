package cloudflare

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newAccountsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("accounts",
		newAccountsListCmd(factory),
		newAccountsGetCmd(factory),
	)
}

func TestAccountsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAccountsTestCmd(factory))
	output := runCmd(t, root, "accounts", "list")

	mustContain(t, output, testAccountID)
	mustContain(t, output, "My Cloudflare Account")
	mustContain(t, output, "acct_other456")
	mustContain(t, output, "Another Account")
}

func TestAccountsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAccountsTestCmd(factory))
	output := runCmd(t, root, "accounts", "list", "--json")

	var accounts []AccountSummary
	err := json.Unmarshal([]byte(output), &accounts)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Len(t, accounts, 2)
	assert.Equal(t, testAccountID, accounts[0].ID)
	assert.Equal(t, "My Cloudflare Account", accounts[0].Name)
	assert.Equal(t, "standard", accounts[0].Type)
	assert.Equal(t, "acct_other456", accounts[1].ID)
	assert.Equal(t, "enterprise", accounts[1].Type)
}

func TestAccountsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAccountsTestCmd(factory))
	output := runCmd(t, root, "accounts", "get", "--account", testAccountID)

	mustContain(t, output, testAccountID)
	mustContain(t, output, "My Cloudflare Account")
	mustContain(t, output, "standard")
}

func TestAccountsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newAccountsTestCmd(factory))
	output := runCmd(t, root, "accounts", "get", "--account", testAccountID, "--json")

	var account AccountSummary
	err := json.Unmarshal([]byte(output), &account)
	assert.NoError(t, err, "expected valid JSON, got: %s", output)
	assert.Equal(t, testAccountID, account.ID)
	assert.Equal(t, "My Cloudflare Account", account.Name)
	assert.Equal(t, "standard", account.Type)
}
