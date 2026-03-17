package github

import (
	"encoding/json"
	"testing"
)

// --- Labels list tests ---

func TestLabelsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("labels", []string{"label"},
		newLabelsListCmd(factory),
		newLabelsGetCmd(factory),
		newLabelsCreateCmd(factory),
		newLabelsUpdateCmd(factory),
		newLabelsDeleteCmd(factory),
	))

	out := runCmd(t, root, "labels", "list", "--owner=alice", "--repo=repo-alpha")
	mustContain(t, out, "bug")
	mustContain(t, out, "enhancement")
	mustContain(t, out, "d73a4a")
}

func TestLabelsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("labels", []string{"label"},
		newLabelsListCmd(factory),
		newLabelsGetCmd(factory),
		newLabelsCreateCmd(factory),
		newLabelsUpdateCmd(factory),
		newLabelsDeleteCmd(factory),
	))

	out := runCmd(t, root, "labels", "list", "--owner=alice", "--repo=repo-alpha", "--json")

	var labels []LabelInfo
	if err := json.Unmarshal([]byte(out), &labels); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(labels) < 2 {
		t.Fatalf("expected 2 labels, got %d", len(labels))
	}
	if labels[0].Name != "bug" {
		t.Errorf("labels[0].Name = %q, want %q", labels[0].Name, "bug")
	}
}

// --- Labels get tests ---

func TestLabelsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("labels", []string{"label"},
		newLabelsListCmd(factory),
		newLabelsGetCmd(factory),
		newLabelsCreateCmd(factory),
		newLabelsUpdateCmd(factory),
		newLabelsDeleteCmd(factory),
	))

	out := runCmd(t, root, "labels", "get", "--owner=alice", "--repo=repo-alpha", "--name=bug")
	mustContain(t, out, "bug")
	mustContain(t, out, "d73a4a")
	mustContain(t, out, "Something is broken")
}

func TestLabelsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("labels", []string{"label"},
		newLabelsListCmd(factory),
		newLabelsGetCmd(factory),
		newLabelsCreateCmd(factory),
		newLabelsUpdateCmd(factory),
		newLabelsDeleteCmd(factory),
	))

	out := runCmd(t, root, "labels", "get", "--owner=alice", "--repo=repo-alpha", "--name=bug", "--json")

	var label LabelInfo
	if err := json.Unmarshal([]byte(out), &label); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if label.Name != "bug" {
		t.Errorf("label.Name = %q, want %q", label.Name, "bug")
	}
	if label.Color != "d73a4a" {
		t.Errorf("label.Color = %q, want %q", label.Color, "d73a4a")
	}
}

// --- Labels create tests ---

func TestLabelsCreateDryRun(t *testing.T) {
	// labels create does not have a --dry-run gate in makeRunLabelsCreate; it
	// calls the API directly. The test validates the happy path instead, which
	// still exercises the command's flags and the mock server POST handler.
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("labels", []string{"label"},
		newLabelsListCmd(factory),
		newLabelsGetCmd(factory),
		newLabelsCreateCmd(factory),
		newLabelsUpdateCmd(factory),
		newLabelsDeleteCmd(factory),
	))

	out := runCmd(t, root, "labels", "create",
		"--owner=alice", "--repo=repo-alpha",
		"--name=security", "--color=ee0701", "--description=Security issue")
	mustContain(t, out, "security")
}

func TestLabelsCreateText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("labels", []string{"label"},
		newLabelsListCmd(factory),
		newLabelsGetCmd(factory),
		newLabelsCreateCmd(factory),
		newLabelsUpdateCmd(factory),
		newLabelsDeleteCmd(factory),
	))

	out := runCmd(t, root, "labels", "create",
		"--owner=alice", "--repo=repo-alpha",
		"--name=security", "--color=ee0701")
	mustContain(t, out, "security")
}

func TestLabelsCreateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("labels", []string{"label"},
		newLabelsListCmd(factory),
		newLabelsGetCmd(factory),
		newLabelsCreateCmd(factory),
		newLabelsUpdateCmd(factory),
		newLabelsDeleteCmd(factory),
	))

	out := runCmd(t, root, "labels", "create",
		"--owner=alice", "--repo=repo-alpha",
		"--name=security", "--color=ee0701", "--json")

	var label LabelInfo
	if err := json.Unmarshal([]byte(out), &label); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if label.Name != "security" {
		t.Errorf("label.Name = %q, want %q", label.Name, "security")
	}
	if label.Color != "ee0701" {
		t.Errorf("label.Color = %q, want %q", label.Color, "ee0701")
	}
}

// --- Labels update tests ---

func TestLabelsUpdateText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("labels", []string{"label"},
		newLabelsListCmd(factory),
		newLabelsGetCmd(factory),
		newLabelsCreateCmd(factory),
		newLabelsUpdateCmd(factory),
		newLabelsDeleteCmd(factory),
	))

	out := runCmd(t, root, "labels", "update",
		"--owner=alice", "--repo=repo-alpha",
		"--name=bug", "--new-name=bug-fix")
	mustContain(t, out, "bug-fix")
}

func TestLabelsUpdateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("labels", []string{"label"},
		newLabelsListCmd(factory),
		newLabelsGetCmd(factory),
		newLabelsCreateCmd(factory),
		newLabelsUpdateCmd(factory),
		newLabelsDeleteCmd(factory),
	))

	out := runCmd(t, root, "labels", "update",
		"--owner=alice", "--repo=repo-alpha",
		"--name=bug", "--new-name=bug-fix", "--json")

	var label LabelInfo
	if err := json.Unmarshal([]byte(out), &label); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if label.Name != "bug-fix" {
		t.Errorf("label.Name = %q, want %q", label.Name, "bug-fix")
	}
}

// --- Labels delete tests ---

func TestLabelsDeleteConfirmRequired(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("labels", []string{"label"},
		newLabelsListCmd(factory),
		newLabelsGetCmd(factory),
		newLabelsCreateCmd(factory),
		newLabelsUpdateCmd(factory),
		newLabelsDeleteCmd(factory),
	))

	err := runCmdErr(t, root, "labels", "delete",
		"--owner=alice", "--repo=repo-alpha", "--name=bug")
	if err == nil {
		t.Error("expected error when --confirm not provided")
	}
}

func TestLabelsDeleteWithConfirmText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("labels", []string{"label"},
		newLabelsListCmd(factory),
		newLabelsGetCmd(factory),
		newLabelsCreateCmd(factory),
		newLabelsUpdateCmd(factory),
		newLabelsDeleteCmd(factory),
	))

	out := runCmd(t, root, "labels", "delete",
		"--owner=alice", "--repo=repo-alpha", "--name=bug", "--confirm")
	mustContain(t, out, "Deleted label")
	mustContain(t, out, "bug")
}

func TestLabelsDeleteWithConfirmJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	t.Setenv("GITHUB_API_BASE_URL", server.URL)

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("labels", []string{"label"},
		newLabelsListCmd(factory),
		newLabelsGetCmd(factory),
		newLabelsCreateCmd(factory),
		newLabelsUpdateCmd(factory),
		newLabelsDeleteCmd(factory),
	))

	out := runCmd(t, root, "labels", "delete",
		"--owner=alice", "--repo=repo-alpha", "--name=bug", "--confirm", "--json")

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if result["status"] != "deleted" {
		t.Errorf("status = %q, want %q", result["status"], "deleted")
	}
	if result["name"] != "bug" {
		t.Errorf("name = %q, want %q", result["name"], "bug")
	}
}
