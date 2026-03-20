package linkedin

import (
	"context"
	"testing"
)

func TestGroupsList_Deprecated(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"groups", "list"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error for deprecated groups list endpoint")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message, got: %s", err.Error())
	}
}

func TestGroupsGet_Deprecated(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"groups", "get", "--group-id=12345"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error for deprecated groups get endpoint")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message, got: %s", err.Error())
	}
}

func TestGroupsGet_MissingFlag(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"groups", "get"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error when --group-id is missing")
	}
}

func TestGroupsMembers_Deprecated(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"groups", "members", "--group-id=12345"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error for deprecated groups members endpoint")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message, got: %s", err.Error())
	}
}

func TestGroupsMembers_MissingFlag(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"groups", "members"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error when --group-id is missing")
	}
}

func TestGroupsPosts_Deprecated(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"groups", "posts", "--group-id=12345"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error for deprecated groups posts endpoint")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message, got: %s", err.Error())
	}
}

func TestGroupsPosts_MissingFlag(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"groups", "posts"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error when --group-id is missing")
	}
}

func TestGroupsJoin_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"groups", "join", "--group-id=12345", "--dry-run"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestGroupsJoin_Execute(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"groups", "join", "--group-id=12345"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "12345") {
		t.Errorf("expected group ID in output, got: %s", out)
	}
}

func TestGroupsJoin_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "groups", "join", "--group-id=12345"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"joined"`) {
		t.Errorf("expected JSON field 'joined' in output, got: %s", out)
	}
}

func TestGroupsLeave_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"groups", "leave", "--group-id=12345", "--dry-run"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestGroupsLeave_RequiresConfirm(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"groups", "leave", "--group-id=12345"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error when --confirm is not provided")
	}
	if !containsStr(err.Error(), "irreversible") {
		t.Errorf("expected 'irreversible' in error message, got: %s", err.Error())
	}
}

func TestGroupsLeave_WithConfirm(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"groups", "leave", "--group-id=12345", "--confirm"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "Left") {
		t.Errorf("expected 'Left' in output, got: %s", out)
	}
}

func TestGroupsLeave_MissingFlag(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newGroupsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"groups", "leave"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error when --group-id is missing")
	}
}

func TestToGroupSummary(t *testing.T) {
	el := voyagerGroupElement{
		EntityURN:   "urn:li:fs_group:12345",
		Name:        "Go Developers",
		MemberCount: 50000,
		Description: "A community for Go developers",
	}
	s := toGroupSummary(el)
	if s.ID != "12345" {
		t.Errorf("ID = %q, want %q", s.ID, "12345")
	}
	if s.Name != "Go Developers" {
		t.Errorf("Name = %q, want %q", s.Name, "Go Developers")
	}
	if s.MemberCount != 50000 {
		t.Errorf("MemberCount = %d, want 50000", s.MemberCount)
	}
}
