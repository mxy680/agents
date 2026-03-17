package drive

import (
	"encoding/json"
	"testing"
)

func TestPermissionsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestPermissionsCmd(factory))
	out := runCmd(t, root, "permissions", "list", "--file-id=file1")

	mustContain(t, out, "ID")
	mustContain(t, out, "ROLE")
	mustContain(t, out, "perm1")
	mustContain(t, out, "owner")
	mustContain(t, out, "alice@example.com")
}

func TestPermissionsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestPermissionsCmd(factory))
	out := runCmd(t, root, "permissions", "list", "--file-id=file1", "--json")

	var perms []PermissionInfo
	if err := json.Unmarshal([]byte(out), &perms); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(perms) != 2 {
		t.Errorf("expected 2 permissions, got %d", len(perms))
	}
	if perms[0].Role != "owner" {
		t.Errorf("expected first permission role=owner, got %s", perms[0].Role)
	}
}

func TestPermissionsGetText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestPermissionsCmd(factory))
	out := runCmd(t, root, "permissions", "get", "--file-id=file1", "--permission-id=perm1")

	mustContain(t, out, "ID:      perm1")
	mustContain(t, out, "Role:    owner")
	mustContain(t, out, "Email:   alice@example.com")
}

func TestPermissionsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestPermissionsCmd(factory))
	out := runCmd(t, root, "permissions", "get", "--file-id=file1", "--permission-id=perm1", "--json")

	var info PermissionInfo
	if err := json.Unmarshal([]byte(out), &info); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if info.ID != "perm1" {
		t.Errorf("expected ID=perm1, got %s", info.ID)
	}
}

func TestPermissionsCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestPermissionsCmd(factory))
	out := runCmd(t, root, "permissions", "create",
		"--file-id=file1", "--role=reader", "--type=user", "--email=bob@example.com", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "Would share file file1")
}

func TestPermissionsCreateDryRunJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestPermissionsCmd(factory))
	out := runCmd(t, root, "permissions", "create",
		"--file-id=file1", "--role=reader", "--type=user", "--email=bob@example.com", "--dry-run", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["action"] != "create_permission" {
		t.Errorf("expected action=create_permission, got %v", result["action"])
	}
}

func TestPermissionsCreateText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestPermissionsCmd(factory))
	out := runCmd(t, root, "permissions", "create",
		"--file-id=file1", "--role=reader", "--type=user", "--email=bob@example.com")

	mustContain(t, out, "Created permission:")
	mustContain(t, out, "role=reader")
}

func TestPermissionsCreateJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestPermissionsCmd(factory))
	out := runCmd(t, root, "permissions", "create",
		"--file-id=file1", "--role=reader", "--type=user", "--email=bob@example.com", "--json")

	var info PermissionInfo
	if err := json.Unmarshal([]byte(out), &info); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if info.Role != "reader" {
		t.Errorf("expected role=reader, got %s", info.Role)
	}
}

func TestPermissionsDeleteRequiresConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestPermissionsCmd(factory))
	err := runCmdErr(t, root, "permissions", "delete",
		"--file-id=file1", "--permission-id=perm1")

	if err == nil {
		t.Fatal("expected error without --confirm")
	}
	mustContain(t, err.Error(), "irreversible")
}

func TestPermissionsDeleteWithConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestPermissionsCmd(factory))
	out := runCmd(t, root, "permissions", "delete",
		"--file-id=file1", "--permission-id=perm1", "--confirm")

	mustContain(t, out, "Deleted permission: perm1")
}

func TestPermissionsDeleteJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestPermissionsCmd(factory))
	out := runCmd(t, root, "permissions", "delete",
		"--file-id=file1", "--permission-id=perm1", "--confirm", "--json")

	var result map[string]string
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result["status"] != "deleted" {
		t.Errorf("expected status=deleted, got %s", result["status"])
	}
}

func TestPermissionsDeleteDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestPermissionsCmd(factory))
	out := runCmd(t, root, "permissions", "delete",
		"--file-id=file1", "--permission-id=perm1", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "Would delete permission perm1")
}
