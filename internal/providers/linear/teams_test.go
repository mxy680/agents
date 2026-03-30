package linear

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newTeamsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("teams",
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
	)
}

func TestTeamsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newTeamsTestCmd(factory))
	output := runCmd(t, root, "teams", "list")

	mustContain(t, output, "Engineering")
	mustContain(t, output, "ENG")
	mustContain(t, output, "Design")
	mustContain(t, output, "DES")
}

func TestTeamsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newTeamsTestCmd(factory))
	output := runCmd(t, root, "teams", "list", "--json")

	var results []TeamSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Len(t, results, 2)
	assert.Equal(t, "team-abc1", results[0].ID)
	assert.Equal(t, "Engineering", results[0].Name)
	assert.Equal(t, "ENG", results[0].Key)
	assert.Equal(t, "team-def2", results[1].ID)
	assert.Equal(t, "Design", results[1].Name)
	assert.Equal(t, "DES", results[1].Key)
}

func TestTeamsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newTeamsTestCmd(factory))
	output := runCmd(t, root, "teams", "get", "--id", "team-abc1")

	mustContain(t, output, "team-abc1")
	mustContain(t, output, "Engineering")
	mustContain(t, output, "ENG")
	mustContain(t, output, "Core engineering team")
	mustContain(t, output, "Members:")
}

func TestTeamsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newTeamsTestCmd(factory))
	output := runCmd(t, root, "teams", "get", "--id", "team-abc1", "--json")

	var detail TeamDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "team-abc1", detail.ID)
	assert.Equal(t, "Engineering", detail.Name)
	assert.Equal(t, "ENG", detail.Key)
	assert.Equal(t, "Core engineering team", detail.Description)
	assert.Len(t, detail.Members, 1)
	assert.Equal(t, "Alice", detail.Members[0].Name)
	assert.Equal(t, "alice@example.com", detail.Members[0].Email)
}

func TestTeamsMembers_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newTeamsTestCmd(factory))
	output := runCmd(t, root, "teams", "members", "--id", "team-abc1")

	mustContain(t, output, "Alice")
	mustContain(t, output, "alice@example.com")
	mustContain(t, output, "Bob")
	mustContain(t, output, "bob@example.com")
}

func TestTeamsMembers_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newTeamsTestCmd(factory))
	output := runCmd(t, root, "teams", "members", "--id", "team-abc1", "--json")

	var members []UserSummary
	if err := json.Unmarshal([]byte(output), &members); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Len(t, members, 2)
	assert.Equal(t, "usr-abc1", members[0].ID)
	assert.Equal(t, "Alice", members[0].Name)
	assert.Equal(t, "alice@example.com", members[0].Email)
	assert.Equal(t, "usr-def2", members[1].ID)
	assert.Equal(t, "Bob", members[1].Name)
}
