package vercel

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newTeamsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("teams",
		newTeamsListCmd(factory),
		newTeamsGetCmd(factory),
		newTeamsMembersCmd(factory),
	)
}

func TestTeamsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newTeamsTestCmd(factory))
	output := runCmd(t, root, "teams", "list", "--json")

	var teams []TeamSummary
	if err := json.Unmarshal([]byte(output), &teams); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(teams) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(teams))
	}
	if teams[0].ID != "team_abc1" {
		t.Errorf("expected first team ID=team_abc1, got %s", teams[0].ID)
	}
	if teams[0].Slug != "acme-corp" {
		t.Errorf("expected first team Slug=acme-corp, got %s", teams[0].Slug)
	}
	if teams[0].Name != "Acme Corp" {
		t.Errorf("expected first team Name=Acme Corp, got %s", teams[0].Name)
	}
}

func TestTeamsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newTeamsTestCmd(factory))
	output := runCmd(t, root, "teams", "list")

	mustContain(t, output, "team_abc1")
	mustContain(t, output, "acme-corp")
	mustContain(t, output, "Acme Corp")
}

func TestTeamsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newTeamsTestCmd(factory))
	output := runCmd(t, root, "teams", "get", "--id", "team_abc1", "--json")

	var team TeamSummary
	if err := json.Unmarshal([]byte(output), &team); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if team.ID != "team_abc1" {
		t.Errorf("expected ID=team_abc1, got %s", team.ID)
	}
	if team.Slug != "acme-corp" {
		t.Errorf("expected Slug=acme-corp, got %s", team.Slug)
	}
	if team.Name != "Acme Corp" {
		t.Errorf("expected Name=Acme Corp, got %s", team.Name)
	}
}

func TestTeamsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newTeamsTestCmd(factory))
	output := runCmd(t, root, "teams", "get", "--id", "team_abc1")

	mustContain(t, output, "team_abc1")
	mustContain(t, output, "acme-corp")
	mustContain(t, output, "Acme Corp")
}

func TestTeamsMembers_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newTeamsTestCmd(factory))
	output := runCmd(t, root, "teams", "members", "--id", "team_abc1", "--json")

	var members []TeamMember
	if err := json.Unmarshal([]byte(output), &members); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
	if members[0].UID != "usr_alice1" {
		t.Errorf("expected first member UID=usr_alice1, got %s", members[0].UID)
	}
	if members[0].Username != "alice" {
		t.Errorf("expected first member Username=alice, got %s", members[0].Username)
	}
	if members[0].Email != "alice@example.com" {
		t.Errorf("expected first member Email=alice@example.com, got %s", members[0].Email)
	}
	if members[0].Role != "OWNER" {
		t.Errorf("expected first member Role=OWNER, got %s", members[0].Role)
	}
	if members[1].Role != "MEMBER" {
		t.Errorf("expected second member Role=MEMBER, got %s", members[1].Role)
	}
}

func TestTeamsMembers_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newTeamsTestCmd(factory))
	output := runCmd(t, root, "teams", "members", "--id", "team_abc1")

	mustContain(t, output, "alice")
	mustContain(t, output, "OWNER")
	mustContain(t, output, "bob")
}
