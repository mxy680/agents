package linear

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

// graphqlRequest is the parsed body of a Linear GraphQL POST.
type graphqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

// parseGQLRequest reads and parses the GraphQL request body.
func parseGQLRequest(r *http.Request) graphqlRequest {
	data, _ := io.ReadAll(r.Body)
	var req graphqlRequest
	json.Unmarshal(data, &req)
	return req
}

// respondJSON writes a JSON body to w.
func respondJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// gqlData wraps a value in a {"data": ...} GraphQL envelope.
func gqlData(v any) map[string]any {
	return map[string]any{"data": v}
}

// gqlError returns a GraphQL error envelope.
func gqlError(msg string) map[string]any {
	return map[string]any{
		"errors": []map[string]any{
			{"message": msg},
		},
	}
}

// newGraphQLHandler returns an http.Handler that dispatches GraphQL requests
// based on keywords in the query string.
func newGraphQLHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := parseGQLRequest(r)
		q := req.Query
		vars := req.Variables

		w.Header().Set("Content-Type", "application/json")

		switch {
		// --- Issues ---
		case strings.Contains(q, "issueDelete"):
			respondJSON(w, gqlData(map[string]any{
				"issueDelete": map[string]any{"success": true},
			}))

		case strings.Contains(q, "issueUpdate"):
			respondJSON(w, gqlData(map[string]any{
				"issueUpdate": map[string]any{
					"issue": map[string]any{
						"id":         "issue-abc1",
						"identifier": "ENG-1",
						"title":      "Updated issue title",
						"state":      map[string]any{"name": "In Progress"},
					},
				},
			}))

		case strings.Contains(q, "issueCreate"):
			input, _ := vars["input"].(map[string]any)
			title, _ := input["title"].(string)
			respondJSON(w, gqlData(map[string]any{
				"issueCreate": map[string]any{
					"issue": map[string]any{
						"id":         "issue-new1",
						"identifier": "ENG-42",
						"title":      title,
					},
				},
			}))

		case strings.Contains(q, "issue(") && strings.Contains(q, "comments"):
			respondJSON(w, gqlData(map[string]any{
				"issue": map[string]any{
					"comments": map[string]any{
						"nodes": []map[string]any{
							{
								"id":        "cmt-abc1",
								"body":      "First comment",
								"user":      map[string]any{"name": "Alice"},
								"createdAt": "2024-01-01T00:00:00Z",
							},
							{
								"id":        "cmt-def2",
								"body":      "Second comment",
								"user":      map[string]any{"name": "Bob"},
								"createdAt": "2024-01-02T00:00:00Z",
							},
						},
					},
				},
			}))

		case strings.Contains(q, "issue("):
			// issue get: query($id: String!) { issue(id: $id) { ... description ... } }
			respondJSON(w, gqlData(map[string]any{
				"issue": map[string]any{
					"id":          "issue-abc1",
					"identifier":  "ENG-1",
					"title":       "Fix login bug",
					"description": "Users can't log in with SSO",
					"state":       map[string]any{"name": "In Progress"},
					"priority":    2,
					"assignee":    map[string]any{"name": "Alice"},
					"team":        map[string]any{"name": "Engineering"},
					"labels":      map[string]any{"nodes": []map[string]any{{"name": "bug"}}},
					"createdAt":   "2024-01-01T00:00:00Z",
					"updatedAt":   "2024-01-02T00:00:00Z",
				},
			}))

		case strings.Contains(q, "issues("):
			respondJSON(w, gqlData(map[string]any{
				"issues": map[string]any{
					"nodes": []map[string]any{
						{
							"id":         "issue-abc1",
							"identifier": "ENG-1",
							"title":      "Fix login bug",
							"state":      map[string]any{"name": "In Progress"},
							"priority":   2,
							"assignee":   map[string]any{"name": "Alice"},
							"createdAt":  "2024-01-01T00:00:00Z",
						},
						{
							"id":         "issue-def2",
							"identifier": "ENG-2",
							"title":      "Add dark mode",
							"state":      map[string]any{"name": "Todo"},
							"priority":   3,
							"assignee":   nil,
							"createdAt":  "2024-01-02T00:00:00Z",
						},
					},
				},
			}))

		// --- Projects ---
		case strings.Contains(q, "projectDelete"):
			respondJSON(w, gqlData(map[string]any{
				"projectDelete": map[string]any{"success": true},
			}))

		case strings.Contains(q, "projectUpdate"):
			respondJSON(w, gqlData(map[string]any{
				"projectUpdate": map[string]any{
					"project": map[string]any{
						"id":   "proj-abc1",
						"name": "Updated Project",
					},
				},
			}))

		case strings.Contains(q, "projectCreate"):
			input, _ := vars["input"].(map[string]any)
			name, _ := input["name"].(string)
			respondJSON(w, gqlData(map[string]any{
				"projectCreate": map[string]any{
					"project": map[string]any{
						"id":   "proj-new1",
						"name": name,
					},
				},
			}))

		case strings.Contains(q, "project(id:"):
			respondJSON(w, gqlData(map[string]any{
				"project": map[string]any{
					"id":          "proj-abc1",
					"name":        "My Project",
					"description": "A great project",
					"state":       "started",
					"startDate":   "2024-01-01",
					"targetDate":  "2024-06-01",
					"progress":    0.4,
				},
			}))

		case strings.Contains(q, "projects("):
			respondJSON(w, gqlData(map[string]any{
				"projects": map[string]any{
					"nodes": []map[string]any{
						{
							"id":       "proj-abc1",
							"name":     "My Project",
							"state":    "started",
							"progress": 0.4,
							"teams":    map[string]any{"nodes": []map[string]any{{"name": "Engineering"}}},
						},
						{
							"id":       "proj-def2",
							"name":     "Another Project",
							"state":    "planned",
							"progress": 0.0,
							"teams":    map[string]any{"nodes": []map[string]any{}},
						},
					},
				},
			}))

		// --- Cycles ---
		case strings.Contains(q, "cycle(id:"):
			respondJSON(w, gqlData(map[string]any{
				"cycle": map[string]any{
					"id":       "cycle-abc1",
					"number":   3,
					"startsAt": "2024-01-01T00:00:00Z",
					"endsAt":   "2024-01-14T00:00:00Z",
					"issues": map[string]any{
						"nodes": []map[string]any{
							{"id": "issue-abc1", "title": "Fix login bug"},
						},
					},
				},
			}))

		case strings.Contains(q, "activeCycle"):
			respondJSON(w, gqlData(map[string]any{
				"team": map[string]any{
					"activeCycle": map[string]any{
						"id":       "cycle-abc1",
						"number":   3,
						"startsAt": "2024-01-01T00:00:00Z",
						"endsAt":   "2024-01-14T00:00:00Z",
					},
				},
			}))

		case strings.Contains(q, "cycles"):
			respondJSON(w, gqlData(map[string]any{
				"team": map[string]any{
					"cycles": map[string]any{
						"nodes": []map[string]any{
							{
								"id":       "cycle-abc1",
								"number":   1,
								"startsAt": "2024-01-01T00:00:00Z",
								"endsAt":   "2024-01-14T00:00:00Z",
							},
							{
								"id":       "cycle-def2",
								"number":   2,
								"startsAt": "2024-01-15T00:00:00Z",
								"endsAt":   "2024-01-28T00:00:00Z",
							},
						},
					},
				},
			}))

		// --- Teams ---
		case strings.Contains(q, "team(id:") && strings.Contains(q, "states"):
			respondJSON(w, gqlData(map[string]any{
				"team": map[string]any{
					"states": map[string]any{
						"nodes": []map[string]any{
							{
								"id":       "state-abc1",
								"name":     "Todo",
								"color":    "#e2e2e2",
								"type":     "unstarted",
								"position": 0.0,
							},
							{
								"id":       "state-def2",
								"name":     "In Progress",
								"color":    "#f2c94c",
								"type":     "started",
								"position": 1.0,
							},
						},
					},
				},
			}))

		case strings.Contains(q, "team(id:") && strings.Contains(q, "members") && !strings.Contains(q, "description"):
			respondJSON(w, gqlData(map[string]any{
				"team": map[string]any{
					"members": map[string]any{
						"nodes": []map[string]any{
							{"id": "usr-abc1", "name": "Alice", "email": "alice@example.com"},
							{"id": "usr-def2", "name": "Bob", "email": "bob@example.com"},
						},
					},
				},
			}))

		case strings.Contains(q, "team(id:"):
			respondJSON(w, gqlData(map[string]any{
				"team": map[string]any{
					"id":          "team-abc1",
					"name":        "Engineering",
					"key":         "ENG",
					"description": "Core engineering team",
					"members": map[string]any{
						"nodes": []map[string]any{
							{"id": "usr-abc1", "name": "Alice", "email": "alice@example.com"},
						},
					},
				},
			}))

		case strings.Contains(q, "teams"):
			respondJSON(w, gqlData(map[string]any{
				"teams": map[string]any{
					"nodes": []map[string]any{
						{"id": "team-abc1", "name": "Engineering", "key": "ENG"},
						{"id": "team-def2", "name": "Design", "key": "DES"},
					},
				},
			}))

		// --- Labels ---
		case strings.Contains(q, "issueLabelDelete"):
			respondJSON(w, gqlData(map[string]any{
				"issueLabelDelete": map[string]any{"success": true},
			}))

		case strings.Contains(q, "issueLabelCreate"):
			input, _ := vars["input"].(map[string]any)
			name, _ := input["name"].(string)
			color, _ := input["color"].(string)
			respondJSON(w, gqlData(map[string]any{
				"issueLabelCreate": map[string]any{
					"issueLabel": map[string]any{
						"id":    "lbl-new1",
						"name":  name,
						"color": color,
					},
				},
			}))

		case strings.Contains(q, "issueLabels"):
			respondJSON(w, gqlData(map[string]any{
				"issueLabels": map[string]any{
					"nodes": []map[string]any{
						{"id": "lbl-abc1", "name": "bug", "color": "#d73a4a"},
						{"id": "lbl-def2", "name": "enhancement", "color": "#a2eeef"},
					},
				},
			}))

		// --- Comments ---
		case strings.Contains(q, "commentDelete"):
			respondJSON(w, gqlData(map[string]any{
				"commentDelete": map[string]any{"success": true},
			}))

		case strings.Contains(q, "commentCreate"):
			input, _ := vars["input"].(map[string]any)
			body, _ := input["body"].(string)
			respondJSON(w, gqlData(map[string]any{
				"commentCreate": map[string]any{
					"comment": map[string]any{
						"id":   "cmt-new1",
						"body": body,
					},
				},
			}))

		// --- Users ---
		case strings.Contains(q, "viewer"):
			respondJSON(w, gqlData(map[string]any{
				"viewer": map[string]any{
					"id":          "usr-me1",
					"name":        "Mark Shteyn",
					"email":       "mark@example.com",
					"displayName": "markshteyn",
				},
			}))

		case strings.Contains(q, "user(id:"):
			respondJSON(w, gqlData(map[string]any{
				"user": map[string]any{
					"id":          "usr-abc1",
					"name":        "Alice",
					"email":       "alice@example.com",
					"displayName": "alice",
					"active":      true,
				},
			}))

		case strings.Contains(q, "users"):
			respondJSON(w, gqlData(map[string]any{
				"users": map[string]any{
					"nodes": []map[string]any{
						{
							"id":          "usr-abc1",
							"name":        "Alice",
							"email":       "alice@example.com",
							"displayName": "alice",
							"active":      true,
						},
						{
							"id":          "usr-def2",
							"name":        "Bob",
							"email":       "bob@example.com",
							"displayName": "bob",
							"active":      false,
						},
					},
				},
			}))

		// --- Webhooks ---
		case strings.Contains(q, "webhookDelete"):
			respondJSON(w, gqlData(map[string]any{
				"webhookDelete": map[string]any{"success": true},
			}))

		case strings.Contains(q, "webhookCreate"):
			input, _ := vars["input"].(map[string]any)
			url, _ := input["url"].(string)
			respondJSON(w, gqlData(map[string]any{
				"webhookCreate": map[string]any{
					"webhook": map[string]any{
						"id":  "wh-new1",
						"url": url,
					},
				},
			}))

		case strings.Contains(q, "webhooks"):
			respondJSON(w, gqlData(map[string]any{
				"webhooks": map[string]any{
					"nodes": []map[string]any{
						{
							"id":        "wh-abc1",
							"url":       "https://hooks.example.com/linear",
							"enabled":   true,
							"team":      map[string]any{"name": "Engineering"},
							"createdAt": "2024-01-01T00:00:00Z",
						},
						{
							"id":        "wh-def2",
							"url":       "https://hooks.example.com/linear2",
							"enabled":   false,
							"team":      nil,
							"createdAt": "2024-01-02T00:00:00Z",
						},
					},
				},
			}))

		default:
			http.Error(w, "unhandled GraphQL query", http.StatusBadRequest)
		}
	})
}

// newFullMockServer creates an httptest.Server with a single /graphql handler
// that dispatches all Linear GraphQL operations.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle("/graphql", newGraphQLHandler())
	return httptest.NewServer(mux)
}

// newTestClientFactory returns a ClientFactory that creates a *Client pointed at the test server.
func newTestClientFactory(server *httptest.Server) ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		return &Client{
			http:    server.Client(),
			baseURL: server.URL + "/graphql",
		}, nil
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

// buildTestCmd creates a subcommand group for a resource for use in tests.
func buildTestCmd(name string, cmds ...*cobra.Command) *cobra.Command {
	cmd := &cobra.Command{Use: name}
	for _, c := range cmds {
		cmd.AddCommand(c)
	}
	return cmd
}

// runCmd executes a cobra command tree with args and returns stdout.
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
