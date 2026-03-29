package gcp

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newOAuthTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("oauth",
		newOAuthListCmd(factory),
		newOAuthCreateCmd(factory),
		newOAuthUpdateCmd(factory),
		newOAuthDeleteCmd(factory),
		newOAuthCreateCredentialsCmd(factory),
		newOAuthListCredentialsCmd(factory),
	)
}

func TestOAuthList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "list", "--project", "my-project-1")

	// The text output truncates long names; check for display names which are shorter.
	mustContain(t, output, "My App Client")
	mustContain(t, output, "Another Client")
}

func TestOAuthList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "list", "--project", "my-project-1", "--json")

	var results []OAuthClientSummary
	assert.NoError(t, json.Unmarshal([]byte(output), &results))
	assert.Len(t, results, 2)
	assert.Contains(t, results[0].Name, "my-app-client")
	assert.Equal(t, "My App Client", results[0].DisplayName)
}

func TestOAuthList_UsesClientProject(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "list")

	mustContain(t, output, "My App Client")
}

func TestOAuthCreate_DryRun_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "create",
		"--project", "my-project-1",
		"--client-id", "my-new-client",
		"--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "my-new-client")
}

func TestOAuthCreate_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "create",
		"--project", "my-project-1",
		"--client-id", "my-new-client",
		"--dry-run", "--json")

	var result map[string]any
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "create", result["action"])
	assert.Equal(t, "my-new-client", result["oauthClientId"])
}

func TestOAuthCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "create",
		"--project", "my-project-1",
		"--client-id", "my-app-client",
		"--display-name", "My App Client")

	mustContain(t, output, "Created OAuth client")
}

func TestOAuthCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "create",
		"--project", "my-project-1",
		"--client-id", "my-app-client",
		"--json")

	var result OAuthClientSummary
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Contains(t, result.Name, "my-app-client")
}

func TestOAuthUpdate_DryRun_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "update",
		"--project", "my-project-1",
		"--client-id", "my-app-client",
		"--redirect-uris", "https://app.example.com/callback",
		"--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "my-app-client")
}

func TestOAuthUpdate_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "update",
		"--project", "my-project-1",
		"--client-id", "my-app-client",
		"--redirect-uris", "https://app.example.com/callback",
		"--dry-run", "--json")

	var result map[string]any
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "update", result["action"])
	assert.Equal(t, "my-app-client", result["clientId"])
}

func TestOAuthUpdate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "update",
		"--project", "my-project-1",
		"--client-id", "my-app-client",
		"--redirect-uris", "https://updated.example.com/callback")

	mustContain(t, output, "Updated OAuth client")
	mustContain(t, output, "my-app-client")
}

func TestOAuthUpdate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "update",
		"--project", "my-project-1",
		"--client-id", "my-app-client",
		"--redirect-uris", "https://updated.example.com/callback",
		"--json")

	var result OAuthClientSummary
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Contains(t, result.Name, "my-app-client")
}

func TestOAuthDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "delete",
		"--project", "my-project-1",
		"--client-id", "my-app-client",
		"--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "my-app-client")
}

func TestOAuthDelete_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "delete",
		"--project", "my-project-1",
		"--client-id", "my-app-client",
		"--dry-run", "--json")

	var result map[string]any
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "delete", result["action"])
	assert.Equal(t, "my-app-client", result["clientId"])
}

func TestOAuthDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "delete",
		"--project", "my-project-1",
		"--client-id", "my-app-client",
		"--confirm")

	mustContain(t, output, "Deleted OAuth client")
	mustContain(t, output, "my-app-client")
}

func TestOAuthDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	err := runCmdErr(t, root, "oauth", "delete",
		"--project", "my-project-1",
		"--client-id", "my-app-client")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}

func TestOAuthDelete_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "delete",
		"--project", "my-project-1",
		"--client-id", "my-app-client",
		"--confirm", "--json")

	var result map[string]string
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "deleted", result["status"])
	assert.Equal(t, "my-app-client", result["clientId"])
}

func TestOAuthCreateCredentials_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "create-credentials",
		"--project", "my-project-1",
		"--client-id", "my-app-client",
		"--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "my-app-client")
}

func TestOAuthCreateCredentials_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "create-credentials",
		"--project", "my-project-1",
		"--client-id", "my-app-client",
		"--dry-run", "--json")

	var result map[string]any
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Equal(t, "create-credentials", result["action"])
}

func TestOAuthCreateCredentials_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "create-credentials",
		"--project", "my-project-1",
		"--client-id", "my-app-client")

	mustContain(t, output, "Client Secret")
	mustContain(t, output, "super-secret-value")
}

func TestOAuthCreateCredentials_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "create-credentials",
		"--project", "my-project-1",
		"--client-id", "my-app-client",
		"--json")

	var result OAuthCredential
	assert.NoError(t, json.Unmarshal([]byte(output), &result))
	assert.Contains(t, result.Name, "cred-1")
	assert.Equal(t, "super-secret-value", result.ClientSecret)
}

func TestOAuthListCredentials_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "list-credentials",
		"--project", "my-project-1",
		"--client-id", "my-app-client")

	// Names are truncated; check for the column header and disabled values.
	mustContain(t, output, "DISABLED")
	mustContain(t, output, "false")
	mustContain(t, output, "true")
}

func TestOAuthListCredentials_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "list-credentials",
		"--project", "my-project-1",
		"--client-id", "my-app-client",
		"--json")

	var results []OAuthCredential
	assert.NoError(t, json.Unmarshal([]byte(output), &results))
	assert.Len(t, results, 2)
	assert.Contains(t, results[0].Name, "cred-1")
	assert.False(t, results[0].Disabled)
	assert.True(t, results[1].Disabled)
}
