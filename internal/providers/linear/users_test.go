package linear

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newUsersTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("users",
		newUsersListCmd(factory),
		newUsersGetCmd(factory),
		newUsersMeCmd(factory),
	)
}

func TestUsersList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newUsersTestCmd(factory))
	output := runCmd(t, root, "users", "list")

	mustContain(t, output, "usr-abc1")
	mustContain(t, output, "Alice")
	mustContain(t, output, "alice@example.com")
	mustContain(t, output, "usr-def2")
	mustContain(t, output, "Bob")
	mustContain(t, output, "bob@example.com")
}

func TestUsersList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newUsersTestCmd(factory))
	output := runCmd(t, root, "users", "list", "--json")

	var results []UserSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Len(t, results, 2)
	assert.Equal(t, "usr-abc1", results[0].ID)
	assert.Equal(t, "Alice", results[0].Name)
	assert.Equal(t, "alice@example.com", results[0].Email)
	assert.Equal(t, "alice", results[0].DisplayName)
	assert.True(t, results[0].Active)
	assert.Equal(t, "usr-def2", results[1].ID)
	assert.Equal(t, "Bob", results[1].Name)
	assert.False(t, results[1].Active)
}

func TestUsersGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newUsersTestCmd(factory))
	output := runCmd(t, root, "users", "get", "--id", "usr-abc1")

	mustContain(t, output, "usr-abc1")
	mustContain(t, output, "Alice")
	mustContain(t, output, "alice@example.com")
	mustContain(t, output, "alice")
	mustContain(t, output, "true")
}

func TestUsersGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newUsersTestCmd(factory))
	output := runCmd(t, root, "users", "get", "--id", "usr-abc1", "--json")

	var user UserSummary
	if err := json.Unmarshal([]byte(output), &user); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "usr-abc1", user.ID)
	assert.Equal(t, "Alice", user.Name)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.Equal(t, "alice", user.DisplayName)
	assert.True(t, user.Active)
}

func TestUsersMe_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newUsersTestCmd(factory))
	output := runCmd(t, root, "users", "me")

	mustContain(t, output, "usr-me1")
	mustContain(t, output, "Mark Shteyn")
	mustContain(t, output, "mark@example.com")
	mustContain(t, output, "markshteyn")
}

func TestUsersMe_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newUsersTestCmd(factory))
	output := runCmd(t, root, "users", "me", "--json")

	var user UserSummary
	if err := json.Unmarshal([]byte(output), &user); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "usr-me1", user.ID)
	assert.Equal(t, "Mark Shteyn", user.Name)
	assert.Equal(t, "mark@example.com", user.Email)
	assert.Equal(t, "markshteyn", user.DisplayName)
	assert.True(t, user.Active)
}
