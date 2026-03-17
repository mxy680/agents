package github

import (
	"net/http"
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		input string
		max   int
		want  string
	}{
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"ab", 5, "ab"},
		{"", 5, ""},
		{"abcdef", 6, "abcdef"},
		{"abcdefg", 6, "abc..."},
	}
	for _, tt := range tests {
		got := truncate(tt.input, tt.max)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.max, got, tt.want)
		}
	}
}

func TestJsonString(t *testing.T) {
	if got := jsonString(nil); got != "" {
		t.Errorf("jsonString(nil) = %q, want empty", got)
	}
	if got := jsonString("hello"); got != "hello" {
		t.Errorf("jsonString(hello) = %q", got)
	}
	if got := jsonString(42); got != "" {
		t.Errorf("jsonString(42) = %q, want empty", got)
	}
}

func TestJsonBool(t *testing.T) {
	if got := jsonBool(nil); got != false {
		t.Error("jsonBool(nil) should be false")
	}
	if got := jsonBool(true); got != true {
		t.Error("jsonBool(true) should be true")
	}
}

func TestJsonInt(t *testing.T) {
	if got := jsonInt(nil); got != 0 {
		t.Errorf("jsonInt(nil) = %d", got)
	}
	if got := jsonInt(float64(42)); got != 42 {
		t.Errorf("jsonInt(42.0) = %d", got)
	}
	if got := jsonInt(10); got != 10 {
		t.Errorf("jsonInt(10) = %d", got)
	}
}

func TestJsonInt64(t *testing.T) {
	if got := jsonInt64(nil); got != 0 {
		t.Errorf("jsonInt64(nil) = %d", got)
	}
	if got := jsonInt64(float64(100)); got != 100 {
		t.Errorf("jsonInt64(100.0) = %d", got)
	}
}

func TestJsonNestedString(t *testing.T) {
	if got := jsonNestedString(nil, "key"); got != "" {
		t.Errorf("jsonNestedString(nil) = %q", got)
	}
	m := map[string]any{"login": "alice"}
	if got := jsonNestedString(m, "login"); got != "alice" {
		t.Errorf("jsonNestedString = %q", got)
	}
	if got := jsonNestedString(m, "missing"); got != "" {
		t.Errorf("jsonNestedString(missing) = %q", got)
	}
}

func TestJsonStringSliceFromLabels(t *testing.T) {
	labels := []any{
		map[string]any{"name": "bug"},
		map[string]any{"name": "enhancement"},
	}
	got := jsonStringSliceFromLabels(labels)
	if len(got) != 2 || got[0] != "bug" || got[1] != "enhancement" {
		t.Errorf("jsonStringSliceFromLabels = %v", got)
	}
	if got := jsonStringSliceFromLabels(nil); got != nil {
		t.Errorf("jsonStringSliceFromLabels(nil) = %v", got)
	}
}

func TestJsonStringSliceFromUsers(t *testing.T) {
	users := []any{
		map[string]any{"login": "alice"},
		map[string]any{"login": "bob"},
	}
	got := jsonStringSliceFromUsers(users)
	if len(got) != 2 || got[0] != "alice" || got[1] != "bob" {
		t.Errorf("jsonStringSliceFromUsers = %v", got)
	}
}

func TestParseLinkHeader(t *testing.T) {
	tests := []struct {
		input string
		want  map[string]string
	}{
		{"", map[string]string{}},
		{
			`<https://api.github.com/repos?page=2>; rel="next", <https://api.github.com/repos?page=5>; rel="last"`,
			map[string]string{
				"next": "https://api.github.com/repos?page=2",
				"last": "https://api.github.com/repos?page=5",
			},
		},
		{
			`<https://api.github.com/repos?page=1>; rel="prev"`,
			map[string]string{
				"prev": "https://api.github.com/repos?page=1",
			},
		},
	}
	for _, tt := range tests {
		got := parseLinkHeader(tt.input)
		for k, v := range tt.want {
			if got[k] != v {
				t.Errorf("parseLinkHeader(%q)[%q] = %q, want %q", tt.input, k, got[k], v)
			}
		}
	}
}

func TestHasNextPage(t *testing.T) {
	if hasNextPage(nil) {
		t.Error("hasNextPage(nil) should be false")
	}

	resp := &http.Response{Header: http.Header{}}
	if hasNextPage(resp) {
		t.Error("hasNextPage(no Link) should be false")
	}

	resp.Header.Set("Link", `<https://api.github.com/repos?page=2>; rel="next"`)
	if !hasNextPage(resp) {
		t.Error("hasNextPage(with next) should be true")
	}
}

func TestToRepoSummary(t *testing.T) {
	data := map[string]any{
		"id": float64(1), "name": "repo", "full_name": "alice/repo",
		"owner": map[string]any{"login": "alice"},
		"private": false, "description": "A repo",
		"html_url": "https://github.com/alice/repo", "updated_at": "2026-01-01T00:00:00Z",
	}
	s := toRepoSummary(data)
	if s.Name != "repo" || s.Owner != "alice" || s.FullName != "alice/repo" {
		t.Errorf("toRepoSummary = %+v", s)
	}
}

func TestToRepoDetail(t *testing.T) {
	data := map[string]any{
		"id": float64(1), "name": "repo", "full_name": "alice/repo",
		"owner": map[string]any{"login": "alice"},
		"private": false, "description": "A repo",
		"html_url": "https://github.com/alice/repo", "clone_url": "https://github.com/alice/repo.git",
		"default_branch": "main", "language": "Go",
		"stargazers_count": float64(42), "forks_count": float64(5), "open_issues_count": float64(3),
		"created_at": "2025-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
	}
	d := toRepoDetail(data)
	if d.Stars != 42 || d.DefaultBranch != "main" || d.Language != "Go" {
		t.Errorf("toRepoDetail = %+v", d)
	}
}

func TestToIssueSummary(t *testing.T) {
	data := map[string]any{
		"number": float64(1), "title": "Bug", "state": "open",
		"user":   map[string]any{"login": "alice"},
		"labels": []any{map[string]any{"name": "bug"}},
	}
	s := toIssueSummary(data)
	if s.Number != 1 || s.Title != "Bug" || s.User != "alice" || len(s.Labels) != 1 {
		t.Errorf("toIssueSummary = %+v", s)
	}
}

func TestToPullSummary(t *testing.T) {
	data := map[string]any{
		"number": float64(5), "title": "PR", "state": "open",
		"user": map[string]any{"login": "alice"},
		"head": map[string]any{"ref": "feature"},
		"base": map[string]any{"ref": "main"},
	}
	s := toPullSummary(data)
	if s.Number != 5 || s.Head != "feature" || s.Base != "main" {
		t.Errorf("toPullSummary = %+v", s)
	}
}

func TestToRunSummary(t *testing.T) {
	data := map[string]any{
		"id": float64(1001), "name": "CI", "status": "completed", "conclusion": "success",
		"head_branch": "main", "event": "push",
	}
	s := toRunSummary(data)
	if s.ID != 1001 || s.Status != "completed" || s.Branch != "main" {
		t.Errorf("toRunSummary = %+v", s)
	}
}

func TestToReleaseSummary(t *testing.T) {
	data := map[string]any{
		"id": float64(500), "tag_name": "v1.0.0", "name": "Release 1.0",
		"draft": false, "prerelease": false,
	}
	s := toReleaseSummary(data)
	if s.ID != 500 || s.TagName != "v1.0.0" {
		t.Errorf("toReleaseSummary = %+v", s)
	}
}
