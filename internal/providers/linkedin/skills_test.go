package linkedin

import (
	"context"
	"testing"
)

func TestSkillsList_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newSkillsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"skills", "list", "--username=marktest"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "Go") {
		t.Errorf("expected skill name 'Go' in output, got: %s", out)
	}
	if !containsStr(out, "Kubernetes") {
		t.Errorf("expected skill name 'Kubernetes' in output, got: %s", out)
	}
}

func TestSkillsList_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newSkillsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "skills", "list", "--username=marktest"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"name"`) {
		t.Errorf("expected JSON field 'name' in output, got: %s", out)
	}
	if !containsStr(out, `"endorsement_count"`) {
		t.Errorf("expected JSON field 'endorsement_count' in output, got: %s", out)
	}
	if !containsStr(out, "Go") {
		t.Errorf("expected skill name in JSON output, got: %s", out)
	}
}

func TestSkillsList_DefaultsToCurrentUser(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newSkillsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		// No --username; should fetch current user then list skills
		root.SetArgs([]string{"skills", "list"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "Go") {
		t.Errorf("expected skill name 'Go' in output for current user, got: %s", out)
	}
}

func TestSkillsList_WithAlias(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newSkillsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"skill", "list", "--username=marktest"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "Go") {
		t.Errorf("expected skill name in output via alias, got: %s", out)
	}
}

func TestSkillsEndorse_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newSkillsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"skills", "endorse", "--urn=urn:li:fs_skill:123", "--skill-id=123", "--dry-run"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestSkillsEndorse_Execute(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newSkillsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"skills", "endorse", "--urn=urn:li:fs_skill:123", "--skill-id=123"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "123") {
		t.Errorf("expected skill ID in output, got: %s", out)
	}
}

func TestSkillsEndorse_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newSkillsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "skills", "endorse", "--urn=urn:li:fs_skill:123", "--skill-id=123"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"endorsed"`) {
		t.Errorf("expected JSON field 'endorsed' in output, got: %s", out)
	}
}

func TestSkillsEndorse_MissingURN(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newSkillsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"skills", "endorse", "--skill-id=123"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error when --urn is missing")
	}
}

func TestSkillsEndorse_MissingSkillID(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newSkillsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"skills", "endorse", "--urn=urn:li:fs_skill:123"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error when --skill-id is missing")
	}
}

func TestSkillsEndorsements_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newSkillsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"skills", "endorsements", "--skill-id=123", "--username=marktest"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "Alice") {
		t.Errorf("expected endorser name 'Alice' in output, got: %s", out)
	}
}

func TestSkillsEndorsements_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newSkillsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "skills", "endorsements", "--skill-id=123", "--username=marktest"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"name"`) {
		t.Errorf("expected JSON field 'name' in output, got: %s", out)
	}
	if !containsStr(out, "Alice") {
		t.Errorf("expected endorser name in JSON output, got: %s", out)
	}
}

func TestSkillsEndorsements_DefaultsToCurrentUser(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newSkillsCmd(newTestClientFactory(srv)))

	out := captureStdout(t, func() {
		// No --username; should use current user
		root.SetArgs([]string{"skills", "endorsements", "--skill-id=123"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "Alice") {
		t.Errorf("expected endorser name for current user, got: %s", out)
	}
}

func TestSkillsEndorsements_MissingSkillID(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	root := newTestRootCmd()
	root.AddCommand(newSkillsCmd(newTestClientFactory(srv)))

	root.SetArgs([]string{"skills", "endorsements"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error when --skill-id is missing")
	}
}

func TestToSkillSummary(t *testing.T) {
	el := voyagerSkillElement{
		EntityURN:        "urn:li:fs_skill:123",
		Name:             "Go",
		EndorsementCount: 42,
	}
	s := toSkillSummary(el)
	if s.ID != "123" {
		t.Errorf("ID = %q, want %q", s.ID, "123")
	}
	if s.Name != "Go" {
		t.Errorf("Name = %q, want %q", s.Name, "Go")
	}
	if s.EndorsementCount != 42 {
		t.Errorf("EndorsementCount = %d, want 42", s.EndorsementCount)
	}
}
