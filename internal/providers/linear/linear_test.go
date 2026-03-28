package linear

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// --- Provider tests ---

func TestProviderName(t *testing.T) {
	p := New()
	assert.Equal(t, "linear", p.Name())
}

func TestRegisterCommands(t *testing.T) {
	p := New()
	parent := &cobra.Command{Use: "integrations"}
	p.RegisterCommands(parent)

	// Verify "linear" command was registered.
	var linearCmd *cobra.Command
	for _, cmd := range parent.Commands() {
		if cmd.Use == "linear" {
			linearCmd = cmd
			break
		}
	}
	assert.NotNil(t, linearCmd, "expected linear subcommand to be registered")

	// Verify all expected resource subcommands are present.
	expectedGroups := []string{"issues", "projects", "cycles", "teams", "comments", "labels", "users", "workflows", "webhooks"}
	registeredGroups := make(map[string]bool)
	for _, cmd := range linearCmd.Commands() {
		registeredGroups[cmd.Use] = true
	}
	for _, group := range expectedGroups {
		assert.True(t, registeredGroups[group], "expected %q subcommand to be registered", group)
	}
}

func TestRegisterCommands_Aliases(t *testing.T) {
	p := New()
	parent := &cobra.Command{Use: "integrations"}
	p.RegisterCommands(parent)

	var linearCmd *cobra.Command
	for _, cmd := range parent.Commands() {
		if cmd.Use == "linear" {
			linearCmd = cmd
			break
		}
	}
	assert.NotNil(t, linearCmd)
	assert.Contains(t, linearCmd.Aliases, "ln")
}

// --- GraphQL error handling ---

func TestGraphQLError_ReturnedByAPI(t *testing.T) {
	// Server that always returns a GraphQL error.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(gqlError("access denied"))
	}))
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))

	err := runCmdErr(t, root, "issues", "list", "--team", "team-abc1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}

func TestGraphQLError_MultipleErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]any{
				{"message": "field not found"},
				{"message": "unauthorized"},
			},
		})
	}))
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newTeamsTestCmd(factory))

	err := runCmdErr(t, root, "teams", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field not found")
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestGraphQLError_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("this is not json"))
	}))
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(newProjectsTestCmd(factory))

	err := runCmdErr(t, root, "projects", "list")
	assert.Error(t, err)
}

func TestGraphQLError_ClientFactoryFails(t *testing.T) {
	factory := func(ctx context.Context) (*Client, error) {
		return nil, assert.AnError
	}

	root := newTestRootCmd()
	root.AddCommand(newIssuesTestCmd(factory))

	err := runCmdErr(t, root, "issues", "list", "--team", "team-abc1")
	assert.Error(t, err)
}

// --- Helpers unit tests ---

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"hi", 2, "hi"},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
		{"", 5, ""},
	}
	for _, tc := range tests {
		got := truncate(tc.input, tc.max)
		assert.Equal(t, tc.expected, got, "truncate(%q, %d)", tc.input, tc.max)
	}
}

func TestPriorityLabel(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "No priority"},
		{1, "Urgent"},
		{2, "High"},
		{3, "Medium"},
		{4, "Low"},
		{5, "5"},
		{99, "99"},
	}
	for _, tc := range tests {
		got := priorityLabel(tc.input)
		assert.Equal(t, tc.expected, got, "priorityLabel(%d)", tc.input)
	}
}

func TestConfirmDestructive_WithConfirm(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("confirm", false, "")
	cmd.Flags().Set("confirm", "true")

	err := confirmDestructive(cmd)
	assert.NoError(t, err)
}

func TestConfirmDestructive_WithoutConfirm(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("confirm", false, "")

	err := confirmDestructive(cmd)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}

func TestLinearError_ErrorMethod(t *testing.T) {
	e := &LinearError{Message: "something went wrong"}
	assert.Equal(t, "Linear API error: something went wrong", e.Error())
}
