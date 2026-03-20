package linkedin

import (
	"context"
	"testing"
)

func TestCompaniesGet_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newCompaniesGetCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"get", "--company-id=testcorp"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "TestCorp") {
		t.Errorf("expected company name in output, got: %s", out)
	}
	if !containsStr(out, "Computer Software") {
		t.Errorf("expected industry in output, got: %s", out)
	}
}

func TestCompaniesGet_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newCompaniesGetCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "get", "--company-id=testcorp"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"name"`) {
		t.Errorf("expected JSON output with name field, got: %s", out)
	}
	if !containsStr(out, "TestCorp") {
		t.Errorf("expected company name in JSON output, got: %s", out)
	}
}

func TestCompaniesGet_MissingFlag(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newCompaniesGetCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	root.SetArgs([]string{"get"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error for missing --company-id, got nil")
	}
}

func TestCompaniesSearch_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newCompaniesSearchCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"search", "--query=TestCorp"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "TestCorp") {
		t.Errorf("expected company name in output, got: %s", out)
	}
}

func TestCompaniesSearch_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newCompaniesSearchCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "search", "--query=TestCorp"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"urn"`) {
		t.Errorf("expected JSON output with urn field, got: %s", out)
	}
}

func TestCompaniesEmployees_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newCompaniesEmployeesCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"employees", "--company-id=1234"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "TestCorp") {
		t.Errorf("expected result in output, got: %s", out)
	}
}

func TestCompaniesFollow_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newCompaniesFollowCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follow", "--company-id=1234", "--dry-run"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

func TestCompaniesFollow_Execute(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newCompaniesFollowCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"follow", "--company-id=1234"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "1234") {
		t.Errorf("expected company ID in output, got: %s", out)
	}
}

func TestCompaniesUnfollow_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newCompaniesUnfollowCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"unfollow", "--company-id=1234", "--dry-run"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

func TestCompaniesUnfollow_Execute(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newCompaniesUnfollowCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"unfollow", "--company-id=1234"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "1234") {
		t.Errorf("expected company ID in output, got: %s", out)
	}
}

func TestCompaniesJobs_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newCompaniesJobsCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"jobs", "--company-id=testcorp"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "Senior Software Engineer") {
		t.Errorf("expected job title in output, got: %s", out)
	}
	if !containsStr(out, "TestCorp") {
		t.Errorf("expected company name in output, got: %s", out)
	}
}

func TestCompaniesJobs_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newCompaniesJobsCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "jobs", "--company-id=testcorp"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"title"`) {
		t.Errorf("expected JSON output with title field, got: %s", out)
	}
}

func TestToCompanySummary(t *testing.T) {
	raw := voyagerCompanyResponse{
		EntityURN:     "urn:li:fs_normalized_company:1234",
		Name:          "TestCorp",
		IndustryName:  "Computer Software",
		StaffCount:    500,
		FollowerCount: 10000,
	}

	s := toCompanySummary(raw)
	if s.ID != "1234" {
		t.Errorf("ID = %q, want %q", s.ID, "1234")
	}
	if s.Name != "TestCorp" {
		t.Errorf("Name = %q, want %q", s.Name, "TestCorp")
	}
	if s.EmployeeCount != 500 {
		t.Errorf("EmployeeCount = %d, want 500", s.EmployeeCount)
	}
	if s.FollowerCount != 10000 {
		t.Errorf("FollowerCount = %d, want 10000", s.FollowerCount)
	}
}
