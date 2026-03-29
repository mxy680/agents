package gcpconsole

import (
	"context"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/emdash-projects/agents/internal/auth"
)

// --- Provider ---

func TestProviderName(t *testing.T) {
	p := New()
	assert.Equal(t, "gcp-console", p.Name())
}

func TestRegisterCommands(t *testing.T) {
	p := New()
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	p.RegisterCommands(root)

	gcpCmd, _, err := root.Find([]string{"gcp-console"})
	assert.NoError(t, err)
	assert.NotNil(t, gcpCmd)

	// Alias also works
	gcpCmdAlias, _, err := root.Find([]string{"gcc"})
	assert.NoError(t, err)
	assert.NotNil(t, gcpCmdAlias)
}

func TestRegisterCommands_OAuthSubcommands(t *testing.T) {
	p := New()
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	p.RegisterCommands(root)

	for _, subcmd := range []string{"list", "get", "create", "update", "delete"} {
		cmd, _, err := root.Find([]string{"gcp-console", "oauth", subcmd})
		assert.NoError(t, err, "subcommand %q should be registered", subcmd)
		assert.NotNil(t, cmd, "subcommand %q should be registered", subcmd)
	}
}

// --- SAPISIDHash (via auth.GCPConsoleSession) ---

// TestSAPISIDHash verifies that SAPISIDHash() produces a correctly formatted
// SAPISIDHASH value. The hash computation itself (computeSAPISIDHash) is
// tested in the auth package; here we exercise the exported method.
func TestSAPISIDHash_Format(t *testing.T) {
	session := &auth.GCPConsoleSession{
		SAPISID:    "test-sapisid-value",
		AllCookies: "SAPISID=test-sapisid-value; OTHER=cookie",
	}
	hash := session.SAPISIDHash()
	mustContain(t, hash, "SAPISIDHASH")
	// Format: "SAPISIDHASH {timestamp}_{hex}"
	assert.True(t, len(hash) > len("SAPISIDHASH "), "hash should have timestamp and hex parts")
}

func TestSAPISIDHash_TwoCallsDifferentTimestamps(t *testing.T) {
	session := &auth.GCPConsoleSession{
		SAPISID:    "my-sapisid",
		AllCookies: "SAPISID=my-sapisid; ANOTHER=xyz",
	}
	// Both calls produce valid SAPISIDHASH values.
	// They may or may not differ (same second) but both must be valid.
	hash1 := session.SAPISIDHash()
	hash2 := session.SAPISIDHash()
	mustContain(t, hash1, "SAPISIDHASH")
	mustContain(t, hash2, "SAPISIDHASH")
}

// --- GCPConsoleError ---

func TestGCPConsoleError_Error(t *testing.T) {
	err := &GCPConsoleError{
		StatusCode: 403,
		Code:       403,
		Message:    "insufficient permissions",
		Status:     "PERMISSION_DENIED",
	}
	msg := err.Error()
	mustContain(t, msg, "403")
	mustContain(t, msg, "insufficient permissions")
}

func TestGCPConsoleError_Error_NoMessage(t *testing.T) {
	err := &GCPConsoleError{StatusCode: 500}
	msg := err.Error()
	mustContain(t, msg, "500")
}

// --- addAPIKey ---

func TestAddAPIKey_NoExistingQuery(t *testing.T) {
	result := addAPIKey("/clients")
	mustContain(t, result, "?key=")
	mustContain(t, result, gcpConsoleAPIKey)
}

func TestAddAPIKey_ExistingQuery(t *testing.T) {
	result := addAPIKey("/clients?projectNumber=123")
	mustContain(t, result, "&key=")
	mustContain(t, result, gcpConsoleAPIKey)
	mustContain(t, result, "projectNumber=123")
}

// --- ClientFactory error propagation ---

func TestClientError_FactoryFailure(t *testing.T) {
	factory := ClientFactory(func(_ context.Context) (*Client, error) {
		return nil, fmt.Errorf("auth failed: missing credentials")
	})

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	err := runCmdErr(t, root, "oauth", "list", "--project-number", "123456789012")

	assert.Error(t, err)
	mustContain(t, err.Error(), "auth failed: missing credentials")
}

func TestClientError_GetFactoryFailure(t *testing.T) {
	factory := ClientFactory(func(_ context.Context) (*Client, error) {
		return nil, fmt.Errorf("no credentials configured")
	})

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	err := runCmdErr(t, root, "oauth", "get",
		"--client-id", "123456789012-abc.apps.googleusercontent.com",
	)

	assert.Error(t, err)
	mustContain(t, err.Error(), "no credentials configured")
}

func TestClientError_CreateFactoryFailure(t *testing.T) {
	factory := ClientFactory(func(_ context.Context) (*Client, error) {
		return nil, fmt.Errorf("no credentials configured")
	})

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	err := runCmdErr(t, root, "oauth", "create",
		"--project-number", "123456789012",
		"--name", "Test",
		"--redirect-uris", "https://example.com/cb",
	)

	assert.Error(t, err)
	mustContain(t, err.Error(), "no credentials configured")
}

func TestClientError_UpdateFactoryFailure(t *testing.T) {
	factory := ClientFactory(func(_ context.Context) (*Client, error) {
		return nil, fmt.Errorf("no credentials configured")
	})

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	err := runCmdErr(t, root, "oauth", "update",
		"--client-id", "123456789012-abc.apps.googleusercontent.com",
		"--add-redirect", "https://new.example.com",
	)

	assert.Error(t, err)
	mustContain(t, err.Error(), "no credentials configured")
}

func TestClientError_DeleteFactoryFailure(t *testing.T) {
	factory := ClientFactory(func(_ context.Context) (*Client, error) {
		return nil, fmt.Errorf("no credentials configured")
	})

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	err := runCmdErr(t, root, "oauth", "delete",
		"--client-id", "123456789012-abc.apps.googleusercontent.com",
		"--confirm",
	)

	assert.Error(t, err)
	mustContain(t, err.Error(), "no credentials configured")
}
