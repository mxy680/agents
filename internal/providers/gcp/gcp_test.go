package gcp

import (
	"context"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderName(t *testing.T) {
	p := New()
	assert.Equal(t, "gcp", p.Name())
}

func TestRegisterCommands(t *testing.T) {
	p := New()
	parent := &cobra.Command{Use: "integrations"}
	p.RegisterCommands(parent)

	// Find the "gcp" subcommand.
	var gcpCmd *cobra.Command
	for _, sub := range parent.Commands() {
		if sub.Use == "gcp" {
			gcpCmd = sub
			break
		}
	}
	require.NotNil(t, gcpCmd, "gcp command should be registered")

	// Verify expected subcommand groups exist by name.
	subNames := map[string]bool{}
	for _, sub := range gcpCmd.Commands() {
		subNames[sub.Use] = true
	}
	assert.True(t, subNames["projects"], "projects subcommand should be registered")
	assert.True(t, subNames["services"], "services subcommand should be registered")
	assert.True(t, subNames["oauth"], "oauth subcommand should be registered")
	assert.True(t, subNames["brands"], "brands subcommand should be registered")
	assert.True(t, subNames["iam"], "iam subcommand should be registered")
}

func TestRegisterCommands_GCPAliases(t *testing.T) {
	p := New()
	parent := &cobra.Command{Use: "integrations"}
	p.RegisterCommands(parent)

	var gcpCmd *cobra.Command
	for _, sub := range parent.Commands() {
		if sub.Use == "gcp" {
			gcpCmd = sub
			break
		}
	}
	require.NotNil(t, gcpCmd)
	assert.Contains(t, gcpCmd.Aliases, "gcloud")
	assert.Contains(t, gcpCmd.Aliases, "gc")
}

func TestProviderClientFactory_Error(t *testing.T) {
	// When the factory returns an error, every command should propagate it.
	expectedErr := errors.New("auth failure")
	errFactory := func(ctx context.Context) (*Client, error) {
		return nil, expectedErr
	}

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(errFactory))

	err := runCmdErr(t, root, "projects", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth failure")
}

func TestResolveProject_NoProjectSet(t *testing.T) {
	// A client with empty projectID and no --project flag should return an error.
	noProjectFactory := func(ctx context.Context) (*Client, error) {
		return &Client{
			http:               nil,
			projectID:          "", // empty — forces error path
			resourceManagerURL: "",
			serviceUsageURL:    "",
			iamURL:             "",
			iapURL:             "",
		}, nil
	}

	root := newTestRootCmd()
	root.AddCommand(newServicesTestCmd(noProjectFactory))

	err := runCmdErr(t, root, "services", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no project specified")
}

func TestGCPError_Error(t *testing.T) {
	e := &GCPError{
		StatusCode: 403,
		Message:    "Permission denied",
	}
	assert.Contains(t, e.Error(), "403")
	assert.Contains(t, e.Error(), "Permission denied")
}

func TestClientDoJSON_APIError(t *testing.T) {
	// Use the full mock server; request a project that has no registered handler.
	// The default http.ServeMux returns 404 for unregistered paths, which should
	// propagate as a *GCPError from the client.
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))

	// /v3/projects/does-not-exist is not registered → 404 → GCPError.
	err := runCmdErr(t, root, "projects", "get", "--project", "does-not-exist")
	assert.Error(t, err)
}
