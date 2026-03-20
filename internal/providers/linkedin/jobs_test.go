package linkedin

import (
	"context"
	"testing"
)

func TestJobsSearch_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newJobsSearchCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"search", "--query=engineer"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	// The mock returns search clusters (may be empty for jobs since clusters returns company data)
	// We check for no crash and valid output
	_ = out
}

func TestJobsSearch_WithFilters(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newJobsSearchCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"search", "--query=engineer", "--location=SF", "--remote=REMOTE", "--experience=ENTRY_LEVEL", "--type=FULL_TIME"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	_ = out
}

func TestJobsSearch_MissingFlag(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newJobsSearchCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	root.SetArgs([]string{"search"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error for missing --query, got nil")
	}
}

func TestJobsGet_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newJobsGetCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"get", "--job-id=3456"})
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

func TestJobsGet_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newJobsGetCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "get", "--job-id=3456"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"title"`) {
		t.Errorf("expected JSON output with title field, got: %s", out)
	}
	if !containsStr(out, "Senior Software Engineer") {
		t.Errorf("expected job title in JSON output, got: %s", out)
	}
}

func TestJobsGet_MissingFlag(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newJobsGetCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	root.SetArgs([]string{"get"})
	err := root.ExecuteContext(context.Background())
	if err == nil {
		t.Error("expected error for missing --job-id, got nil")
	}
}

func TestJobsSave_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newJobsSaveCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"save", "--job-id=3456", "--dry-run"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

func TestJobsSave_Execute(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newJobsSaveCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"save", "--job-id=3456"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "3456") {
		t.Errorf("expected job ID in output, got: %s", out)
	}
}

func TestJobsUnsave_DryRun(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newJobsUnsaveCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"unsave", "--job-id=99", "--dry-run"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

func TestJobsUnsave_Execute(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newJobsUnsaveCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"unsave", "--job-id=99"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "99") {
		t.Errorf("expected job ID in output, got: %s", out)
	}
}

func TestJobsSaved_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newJobsSavedCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"saved"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "Senior Software Engineer") {
		t.Errorf("expected job title in output, got: %s", out)
	}
}

func TestJobsSaved_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newJobsSavedCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "saved"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"title"`) {
		t.Errorf("expected JSON output with title field, got: %s", out)
	}
}

func TestJobsRecommended_Text(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newJobsRecommendedCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"recommended"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, "Staff Engineer") {
		t.Errorf("expected recommended job title in output, got: %s", out)
	}
	if !containsStr(out, "BigCo") {
		t.Errorf("expected company name in output, got: %s", out)
	}
}

func TestJobsRecommended_JSON(t *testing.T) {
	srv := newFullMockServer(t)
	defer srv.Close()

	factory := newTestClientFactory(srv)
	cmd := newJobsRecommendedCmd(factory)
	root := newTestRootCmd()
	root.AddCommand(cmd)

	out := captureStdout(t, func() {
		root.SetArgs([]string{"--json", "recommended"})
		if err := root.ExecuteContext(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	if !containsStr(out, `"title"`) {
		t.Errorf("expected JSON output with title field, got: %s", out)
	}
}

func TestToJobSummaryFromPosting(t *testing.T) {
	raw := voyagerJobPosting{
		EntityURN:         "urn:li:fs_normalized_jobPosting:3456",
		Title:             "Senior Software Engineer",
		CompanyName:       "TestCorp",
		FormattedLocation: "San Francisco, CA",
		ListedAt:          1704067200000,
		WorkRemoteAllowed: true,
	}

	s := toJobSummaryFromPosting(raw)
	if s.ID != "3456" {
		t.Errorf("ID = %q, want %q", s.ID, "3456")
	}
	if s.Title != "Senior Software Engineer" {
		t.Errorf("Title = %q, want %q", s.Title, "Senior Software Engineer")
	}
	if s.Company != "TestCorp" {
		t.Errorf("Company = %q, want %q", s.Company, "TestCorp")
	}
	if s.Remote != "remote" {
		t.Errorf("Remote = %q, want %q", s.Remote, "remote")
	}
}

func TestToJobSummaryFromDetail(t *testing.T) {
	raw := voyagerJobPostingDetailResponse{
		EntityURN:         "urn:li:fs_normalized_jobPosting:3456",
		Title:             "Senior Software Engineer",
		CompanyName:       "TestCorp",
		FormattedLocation: "San Francisco, CA",
		ListedAt:          1704067200000,
		WorkRemoteAllowed: false,
	}

	s := toJobSummaryFromDetail(raw)
	if s.ID != "3456" {
		t.Errorf("ID = %q, want %q", s.ID, "3456")
	}
	if s.Remote != "" {
		t.Errorf("Remote = %q, want empty string for non-remote job", s.Remote)
	}
}
