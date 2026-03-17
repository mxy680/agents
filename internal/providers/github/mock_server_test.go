package github

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// --- Mock server setup ---

// withReposMock registers repo-related mock handlers on mux.
func withReposMock(mux *http.ServeMux) {
	// GET/POST /user/repos (list or create authenticated user's repos)
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			name, _ := body["name"].(string)
			resp := map[string]any{
				"id": 1, "name": name, "full_name": "alice/" + name,
				"owner":            map[string]any{"login": "alice"},
				"private":          body["private"],
				"description":      body["description"],
				"html_url":         "https://github.com/alice/" + name,
				"clone_url":        "https://github.com/alice/" + name + ".git",
				"default_branch":   "main",
				"language":         "",
				"stargazers_count": 0, "forks_count": 0, "open_issues_count": 0,
				"created_at": "2026-03-17T00:00:00Z", "updated_at": "2026-03-17T00:00:00Z",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := []map[string]any{
			{
				"id": 1, "name": "repo-alpha", "full_name": "alice/repo-alpha",
				"owner":            map[string]any{"login": "alice"},
				"private":          false,
				"description":      "First repo",
				"html_url":         "https://github.com/alice/repo-alpha",
				"clone_url":        "https://github.com/alice/repo-alpha.git",
				"default_branch":   "main",
				"language":         "Go",
				"stargazers_count": 42, "forks_count": 5, "open_issues_count": 3,
				"created_at": "2025-01-01T00:00:00Z", "updated_at": "2026-03-15T10:00:00Z",
			},
			{
				"id": 2, "name": "repo-beta", "full_name": "alice/repo-beta",
				"owner":            map[string]any{"login": "alice"},
				"private":          true,
				"description":      "Second repo",
				"html_url":         "https://github.com/alice/repo-beta",
				"clone_url":        "https://github.com/alice/repo-beta.git",
				"default_branch":   "main",
				"language":         "TypeScript",
				"stargazers_count": 10, "forks_count": 1, "open_issues_count": 0,
				"created_at": "2025-06-01T00:00:00Z", "updated_at": "2026-03-14T10:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET/POST /repos/alice/repo-alpha, DELETE
	mux.HandleFunc("/repos/alice/repo-alpha", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		resp := map[string]any{
			"id": 1, "name": "repo-alpha", "full_name": "alice/repo-alpha",
			"owner":            map[string]any{"login": "alice"},
			"private":          false,
			"description":      "First repo",
			"html_url":         "https://github.com/alice/repo-alpha",
			"clone_url":        "https://github.com/alice/repo-alpha.git",
			"default_branch":   "main",
			"language":         "Go",
			"stargazers_count": 42, "forks_count": 5, "open_issues_count": 3,
			"created_at": "2025-01-01T00:00:00Z", "updated_at": "2026-03-15T10:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// POST /repos/alice/repo-alpha/forks
	mux.HandleFunc("/repos/alice/repo-alpha/forks", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id": 99, "name": "repo-alpha", "full_name": "bob/repo-alpha",
			"owner":          map[string]any{"login": "bob"},
			"private":        false,
			"html_url":       "https://github.com/bob/repo-alpha",
			"default_branch": "main",
			"created_at":     "2026-03-16T00:00:00Z", "updated_at": "2026-03-16T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(resp)
	})
}

// withIssuesMock registers issue-related mock handlers on mux.
func withIssuesMock(mux *http.ServeMux) {
	// GET /repos/alice/repo-alpha/issues
	mux.HandleFunc("/repos/alice/repo-alpha/issues", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"number":     3,
				"title":      body["title"],
				"state":      "open",
				"user":       map[string]any{"login": "alice"},
				"body":       body["body"],
				"html_url":   "https://github.com/alice/repo-alpha/issues/3",
				"created_at": "2026-03-16T00:00:00Z",
				"updated_at": "2026-03-16T00:00:00Z",
				"comments":   0,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := []map[string]any{
			{
				"number": 1, "title": "Bug report", "state": "open",
				"user":       map[string]any{"login": "alice"},
				"labels":     []map[string]any{{"name": "bug"}},
				"created_at": "2026-03-10T00:00:00Z", "updated_at": "2026-03-15T00:00:00Z",
			},
			{
				"number": 2, "title": "Feature request", "state": "open",
				"user":       map[string]any{"login": "bob"},
				"labels":     []map[string]any{{"name": "enhancement"}},
				"created_at": "2026-03-11T00:00:00Z", "updated_at": "2026-03-14T00:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET/PATCH /repos/alice/repo-alpha/issues/1
	mux.HandleFunc("/repos/alice/repo-alpha/issues/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			state := "open"
			if s, ok := body["state"].(string); ok {
				state = s
			}
			resp := map[string]any{
				"number": 1, "title": "Bug report", "state": state,
				"user":       map[string]any{"login": "alice"},
				"body":       "This is a bug",
				"html_url":   "https://github.com/alice/repo-alpha/issues/1",
				"labels":     []map[string]any{{"name": "bug"}},
				"created_at": "2026-03-10T00:00:00Z", "updated_at": "2026-03-16T00:00:00Z",
				"comments":   2,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"number": 1, "title": "Bug report", "state": "open",
			"user":       map[string]any{"login": "alice"},
			"body":       "This is a bug",
			"html_url":   "https://github.com/alice/repo-alpha/issues/1",
			"labels":     []map[string]any{{"name": "bug"}},
			"created_at": "2026-03-10T00:00:00Z", "updated_at": "2026-03-15T00:00:00Z",
			"comments":   2,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// POST /repos/alice/repo-alpha/issues/1/comments
	mux.HandleFunc("/repos/alice/repo-alpha/issues/1/comments", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &body)
		resp := map[string]any{
			"id":         101,
			"user":       map[string]any{"login": "alice"},
			"body":       body["body"],
			"created_at": "2026-03-16T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})
}

// withPullsMock registers PR-related mock handlers on mux.
func withPullsMock(mux *http.ServeMux) {
	// GET/POST /repos/alice/repo-alpha/pulls
	mux.HandleFunc("/repos/alice/repo-alpha/pulls", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"number": 10, "title": body["title"], "state": "open",
				"user":       map[string]any{"login": "alice"},
				"head":       map[string]any{"ref": body["head"]},
				"base":       map[string]any{"ref": body["base"]},
				"draft":      body["draft"],
				"html_url":   "https://github.com/alice/repo-alpha/pull/10",
				"additions":  0, "deletions": 0, "commits": 1,
				"created_at": "2026-03-16T00:00:00Z", "updated_at": "2026-03-16T00:00:00Z",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := []map[string]any{
			{
				"number": 5, "title": "Add feature X", "state": "open",
				"user":       map[string]any{"login": "alice"},
				"head":       map[string]any{"ref": "feature-x"},
				"base":       map[string]any{"ref": "main"},
				"draft":      false,
				"created_at": "2026-03-12T00:00:00Z", "updated_at": "2026-03-15T00:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET/PATCH /repos/alice/repo-alpha/pulls/5
	mux.HandleFunc("/repos/alice/repo-alpha/pulls/5", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"number": 5, "title": "Add feature X", "state": "open",
			"user":       map[string]any{"login": "alice"},
			"body":       "Adds feature X",
			"head":       map[string]any{"ref": "feature-x"},
			"base":       map[string]any{"ref": "main"},
			"draft":      false, "mergeable": true,
			"html_url":   "https://github.com/alice/repo-alpha/pull/5",
			"additions":  50, "deletions": 10, "commits": 3,
			"created_at": "2026-03-12T00:00:00Z", "updated_at": "2026-03-15T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// PUT /repos/alice/repo-alpha/pulls/5/merge
	mux.HandleFunc("/repos/alice/repo-alpha/pulls/5/merge", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"sha":     "abc123def456",
			"merged":  true,
			"message": "Pull Request successfully merged",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// POST /repos/alice/repo-alpha/pulls/5/reviews
	mux.HandleFunc("/repos/alice/repo-alpha/pulls/5/reviews", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		data, _ := io.ReadAll(r.Body)
		json.Unmarshal(data, &body)
		resp := map[string]any{
			"id":    201,
			"state": body["event"],
			"body":  body["body"],
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withRunsMock registers workflow run mock handlers on mux.
func withRunsMock(mux *http.ServeMux) {
	// GET /repos/alice/repo-alpha/actions/runs
	mux.HandleFunc("/repos/alice/repo-alpha/actions/runs", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"total_count": 1,
			"workflow_runs": []map[string]any{
				{
					"id": 1001, "name": "CI", "status": "completed", "conclusion": "success",
					"head_branch": "main", "event": "push",
					"workflow_id": 100, "run_number": 42, "run_attempt": 1,
					"html_url":       "https://github.com/alice/repo-alpha/actions/runs/1001",
					"created_at":     "2026-03-15T10:00:00Z", "updated_at": "2026-03-15T10:05:00Z",
					"run_started_at": "2026-03-15T10:00:00Z",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /repos/alice/repo-alpha/actions/runs/1001
	mux.HandleFunc("/repos/alice/repo-alpha/actions/runs/1001", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id": 1001, "name": "CI", "status": "completed", "conclusion": "success",
			"head_branch": "main", "event": "push",
			"workflow_id": 100, "run_number": 42, "run_attempt": 1,
			"html_url":       "https://github.com/alice/repo-alpha/actions/runs/1001",
			"created_at":     "2026-03-15T10:00:00Z", "updated_at": "2026-03-15T10:05:00Z",
			"run_started_at": "2026-03-15T10:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// POST /repos/alice/repo-alpha/actions/runs/1001/rerun
	mux.HandleFunc("/repos/alice/repo-alpha/actions/runs/1001/rerun", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	// GET /repos/alice/repo-alpha/actions/workflows
	mux.HandleFunc("/repos/alice/repo-alpha/actions/workflows", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"total_count": 1,
			"workflows": []map[string]any{
				{
					"id":    100,
					"name":  "CI",
					"path":  ".github/workflows/ci.yml",
					"state": "active",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withReleasesMock registers release-related mock handlers on mux.
func withReleasesMock(mux *http.ServeMux) {
	// GET/POST /repos/alice/repo-alpha/releases
	mux.HandleFunc("/repos/alice/repo-alpha/releases", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"id": 501, "tag_name": body["tag_name"], "name": body["name"],
				"body": body["body"], "draft": body["draft"], "prerelease": body["prerelease"],
				"target_commitish": body["target_commitish"],
				"html_url":         "https://github.com/alice/repo-alpha/releases/tag/" + jsonString(body["tag_name"]),
				"created_at":       "2026-03-16T00:00:00Z", "published_at": "2026-03-16T00:00:00Z",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := []map[string]any{
			{
				"id": 500, "tag_name": "v1.0.0", "name": "Release 1.0",
				"draft": false, "prerelease": false,
				"created_at": "2026-03-01T00:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /repos/alice/repo-alpha/releases/latest
	mux.HandleFunc("/repos/alice/repo-alpha/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id": 500, "tag_name": "v1.0.0", "name": "Release 1.0",
			"body": "First stable release", "draft": false, "prerelease": false,
			"html_url":     "https://github.com/alice/repo-alpha/releases/tag/v1.0.0",
			"created_at":   "2026-03-01T00:00:00Z",
			"published_at": "2026-03-01T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /repos/alice/repo-alpha/releases/tags/v1.0.0
	mux.HandleFunc("/repos/alice/repo-alpha/releases/tags/v1.0.0", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id": 500, "tag_name": "v1.0.0", "name": "Release 1.0",
			"body": "First stable release", "draft": false, "prerelease": false,
			"html_url":     "https://github.com/alice/repo-alpha/releases/tag/v1.0.0",
			"created_at":   "2026-03-01T00:00:00Z",
			"published_at": "2026-03-01T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET/DELETE /repos/alice/repo-alpha/releases/500
	mux.HandleFunc("/repos/alice/repo-alpha/releases/500", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		resp := map[string]any{
			"id": 500, "tag_name": "v1.0.0", "name": "Release 1.0",
			"body": "First stable release", "draft": false, "prerelease": false,
			"html_url":     "https://github.com/alice/repo-alpha/releases/tag/v1.0.0",
			"created_at":   "2026-03-01T00:00:00Z",
			"published_at": "2026-03-01T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withGistsMock registers gist-related mock handlers on mux.
func withGistsMock(mux *http.ServeMux) {
	// GET/POST /gists
	mux.HandleFunc("/gists", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"id":          "gist-new1",
				"description": body["description"],
				"public":      body["public"],
				"files": map[string]any{
					"hello.txt": map[string]any{
						"filename": "hello.txt", "language": "Text", "size": 5,
						"content": "hello",
					},
				},
				"html_url":   "https://gist.github.com/gist-new1",
				"created_at": "2026-03-16T00:00:00Z", "updated_at": "2026-03-16T00:00:00Z",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := []map[string]any{
			{
				"id": "gist1", "description": "My snippet", "public": true,
				"files": map[string]any{
					"snippet.go": map[string]any{"filename": "snippet.go", "size": 100},
				},
				"created_at": "2026-03-10T00:00:00Z", "updated_at": "2026-03-15T00:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET/PATCH/DELETE /gists/gist1
	mux.HandleFunc("/gists/gist1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		resp := map[string]any{
			"id": "gist1", "description": "My snippet", "public": true,
			"files": map[string]any{
				"snippet.go": map[string]any{
					"filename": "snippet.go", "language": "Go", "size": 100,
					"content": "package main\n",
				},
			},
			"html_url":   "https://gist.github.com/gist1",
			"created_at": "2026-03-10T00:00:00Z", "updated_at": "2026-03-15T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withSearchMock registers search-related mock handlers on mux.
func withSearchMock(mux *http.ServeMux) {
	mux.HandleFunc("/search/repositories", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"total_count": 1,
			"items": []map[string]any{
				{
					"id": 1, "name": "repo-alpha", "full_name": "alice/repo-alpha",
					"owner":       map[string]any{"login": "alice"},
					"private":     false,
					"description": "First repo",
					"html_url":    "https://github.com/alice/repo-alpha",
					"updated_at":  "2026-03-15T10:00:00Z",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/search/code", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"total_count": 1,
			"items": []map[string]any{
				{
					"path": "main.go", "sha": "abc123",
					"repository": map[string]any{"full_name": "alice/repo-alpha"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/search/issues", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"total_count": 1,
			"items": []map[string]any{
				{
					"number": 1, "title": "Bug report", "state": "open",
					"user":       map[string]any{"login": "alice"},
					"created_at": "2026-03-10T00:00:00Z", "updated_at": "2026-03-15T00:00:00Z",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/search/commits", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"total_count": 1,
			"items": []map[string]any{
				{
					"sha": "abc1234567890",
					"commit": map[string]any{
						"message": "feat: add feature",
						"author":  map[string]any{"name": "Alice", "date": "2026-03-15T00:00:00Z"},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/search/users", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"total_count": 1,
			"items": []map[string]any{
				{"login": "alice", "id": 1, "type": "User"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withGitMock registers git data mock handlers on mux.
func withGitMock(mux *http.ServeMux) {
	// refs
	mux.HandleFunc("/repos/alice/repo-alpha/git/refs/heads", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{
			{
				"ref": "refs/heads/main", "node_id": "node1",
				"object": map[string]any{"type": "commit", "sha": "abc123"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/repos/alice/repo-alpha/git/ref/heads/main", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"ref": "refs/heads/main", "node_id": "node1",
			"object": map[string]any{"type": "commit", "sha": "abc123"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/repos/alice/repo-alpha/git/refs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"ref":    body["ref"],
				"object": map[string]any{"type": "commit", "sha": body["sha"]},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		// list all refs
		resp := []map[string]any{
			{
				"ref": "refs/heads/main", "node_id": "node1",
				"object": map[string]any{"type": "commit", "sha": "abc123"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// commits
	mux.HandleFunc("/repos/alice/repo-alpha/git/commits/abc123", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"sha": "abc123", "message": "Initial commit",
			"author": map[string]any{"name": "Alice", "email": "alice@example.com", "date": "2025-01-01T00:00:00Z"},
			"tree":   map[string]any{"sha": "tree123"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/repos/alice/repo-alpha/git/commits", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"sha": "newcommit123", "message": "New commit",
			"author": map[string]any{"name": "Alice", "email": "alice@example.com", "date": "2026-03-16T00:00:00Z"},
			"tree":   map[string]any{"sha": "tree456"},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	// trees
	mux.HandleFunc("/repos/alice/repo-alpha/git/trees/tree123", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"sha": "tree123",
			"tree": []map[string]any{
				{"path": "main.go", "mode": "100644", "type": "blob", "size": 100, "sha": "blob1"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/repos/alice/repo-alpha/git/trees", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"sha": "newtree123",
			"tree": []map[string]any{
				{"path": "main.go", "mode": "100644", "type": "blob", "sha": "blob1"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	// blobs
	mux.HandleFunc("/repos/alice/repo-alpha/git/blobs/blob1", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"sha": "blob1", "size": 100, "content": "cGFja2FnZSBtYWlu", "encoding": "base64",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/repos/alice/repo-alpha/git/blobs", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{"sha": "newblob1", "size": 5}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})

	// tags
	mux.HandleFunc("/repos/alice/repo-alpha/git/tags/tag123", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"tag": "v1.0.0", "sha": "tag123", "message": "Release v1.0.0",
			"tagger": map[string]any{"name": "Alice", "email": "alice@example.com", "date": "2026-03-01T00:00:00Z"},
			"object": map[string]any{"type": "commit", "sha": "abc123"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/repos/alice/repo-alpha/git/tags", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"tag": "v2.0.0", "sha": "newtag123", "message": "Release v2.0.0",
			"tagger": map[string]any{"name": "Alice", "email": "alice@example.com", "date": "2026-03-16T00:00:00Z"},
			"object": map[string]any{"type": "commit", "sha": "abc123"},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	})
}

// withOrgsMock registers org-related mock handlers on mux.
func withOrgsMock(mux *http.ServeMux) {
	mux.HandleFunc("/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{
			{"login": "acme-corp", "id": 100, "description": "Acme Corporation"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/orgs/acme-corp", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"login": "acme-corp", "id": 100, "name": "Acme Corporation",
			"description": "Building the future", "email": "info@acme.com",
			"public_repos": 25, "created_at": "2020-01-01T00:00:00Z",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/orgs/acme-corp/members", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{
			{"login": "alice", "id": 1},
			{"login": "bob", "id": 2},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/orgs/acme-corp/repos", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{
			{
				"id": 1, "name": "repo-alpha", "full_name": "acme-corp/repo-alpha",
				"owner": map[string]any{"login": "acme-corp"}, "private": false,
				"html_url": "https://github.com/acme-corp/repo-alpha", "updated_at": "2026-03-15T00:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withTeamsMock registers team-related mock handlers on mux.
func withTeamsMock(mux *http.ServeMux) {
	mux.HandleFunc("/orgs/acme-corp/teams", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{
			{"id": 10, "name": "Backend", "slug": "backend", "description": "Backend team", "permission": "push"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/orgs/acme-corp/teams/backend", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"id": 10, "name": "Backend", "slug": "backend", "description": "Backend team", "permission": "push",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/orgs/acme-corp/teams/backend/members", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{
			{"login": "alice", "id": 1},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/orgs/acme-corp/teams/backend/repos", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{
			{
				"id": 1, "name": "repo-alpha", "full_name": "acme-corp/repo-alpha",
				"owner": map[string]any{"login": "acme-corp"}, "private": false,
				"html_url": "https://github.com/acme-corp/repo-alpha", "updated_at": "2026-03-15T00:00:00Z",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/orgs/acme-corp/teams/backend/repos/alice/repo-alpha", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// PUT
		w.WriteHeader(http.StatusNoContent)
	})
}

// withLabelsMock registers label-related mock handlers on mux.
func withLabelsMock(mux *http.ServeMux) {
	mux.HandleFunc("/repos/alice/repo-alpha/labels", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body map[string]any
			data, _ := io.ReadAll(r.Body)
			json.Unmarshal(data, &body)
			resp := map[string]any{
				"id": 301, "name": body["name"], "color": body["color"], "description": body["description"],
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := []map[string]any{
			{"id": 300, "name": "bug", "color": "d73a4a", "description": "Something is broken"},
			{"id": 301, "name": "enhancement", "color": "a2eeef", "description": "New feature"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/repos/alice/repo-alpha/labels/bug", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodPatch {
			resp := map[string]any{"id": 300, "name": "bug-fix", "color": "d73a4a", "description": "Bug fix"}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{"id": 300, "name": "bug", "color": "d73a4a", "description": "Something is broken"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withBranchesMock registers branch-related mock handlers on mux.
func withBranchesMock(mux *http.ServeMux) {
	mux.HandleFunc("/repos/alice/repo-alpha/branches", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{
			{"name": "main", "commit": map[string]any{"sha": "abc123"}, "protected": true},
			{"name": "dev", "commit": map[string]any{"sha": "def456"}, "protected": false},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/repos/alice/repo-alpha/branches/main", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"name": "main", "commit": map[string]any{"sha": "abc123"}, "protected": true,
			"_links": map[string]any{"html": "https://github.com/alice/repo-alpha/tree/main"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/repos/alice/repo-alpha/branches/main/protection", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodPut {
			resp := map[string]any{
				"url":            "https://api.github.com/repos/alice/repo-alpha/branches/main/protection",
				"enforce_admins": map[string]any{"enabled": true},
				"required_pull_request_reviews": map[string]any{
					"required_approving_review_count": 2,
					"require_code_owner_reviews":      true,
				},
				"required_status_checks": map[string]any{
					"contexts": []any{"ci/build", "ci/test"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := map[string]any{
			"url":            "https://api.github.com/repos/alice/repo-alpha/branches/main/protection",
			"enforce_admins": map[string]any{"enabled": true},
			"required_pull_request_reviews": map[string]any{
				"required_approving_review_count": 1,
				"require_code_owner_reviews":      false,
			},
			"required_status_checks": map[string]any{
				"contexts": []any{"ci/build"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// newFullMockServer creates an httptest server with all GitHub mock handlers.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withReposMock(mux)
	withIssuesMock(mux)
	withPullsMock(mux)
	withRunsMock(mux)
	withReleasesMock(mux)
	withGistsMock(mux)
	withSearchMock(mux)
	withGitMock(mux)
	withOrgsMock(mux)
	withTeamsMock(mux)
	withLabelsMock(mux)
	withBranchesMock(mux)
	return httptest.NewServer(mux)
}

// newTestClientFactory returns a ClientFactory that creates an *http.Client
// pointed at the given httptest server.
func newTestClientFactory(server *httptest.Server) ClientFactory {
	return func(ctx context.Context) (*http.Client, error) {
		return server.Client(), nil
	}
}

// captureStdout runs f with os.Stdout redirected to a pipe and returns the output.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	return string(out)
}

// newTestRootCmd creates a root command with global flags for testing.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	return root
}

// buildTestCmd creates a subcommand tree for a resource group, for use in tests.
func buildTestCmd(name string, aliases []string, cmds ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{Use: name, Aliases: aliases}
	for _, c := range cmds {
		cmd.AddCommand(c)
	}
	return cmd
}

// runCmd is a test helper that executes a cobra command tree with args and returns stdout.
func runCmd(t *testing.T, root *cobra.Command, args ...string) string {
	t.Helper()
	return captureStdout(t, func() {
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}
	})
}

// runCmdErr executes a cobra command tree and returns any error (does not fatal).
func runCmdErr(t *testing.T, root *cobra.Command, args ...string) error {
	t.Helper()
	root.SetArgs(args)
	root.SilenceUsage = true
	root.SilenceErrors = true
	return root.Execute()
}

// mustContain asserts that output contains substr.
func mustContain(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("expected output to contain %q, got:\n%s", substr, output)
	}
}
