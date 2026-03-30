package vercel

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
)

func TestProvider_Name(t *testing.T) {
	p := New()
	if p.Name() != "vercel" {
		t.Errorf("expected Name()=vercel, got %s", p.Name())
	}
}

func TestProvider_RegisterCommands(t *testing.T) {
	p := New()
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	// Should not panic
	p.RegisterCommands(root)

	// Verify the vercel subcommand was registered
	vercelCmd, _, err := root.Find([]string{"vercel"})
	if err != nil || vercelCmd == nil {
		t.Error("expected vercel subcommand to be registered")
	}
}

func TestAddTeamID_NoTeam(t *testing.T) {
	c := &Client{teamID: ""}
	result := c.addTeamID("/v10/projects")
	if result != "/v10/projects" {
		t.Errorf("expected path unchanged, got %s", result)
	}
}

func TestAddTeamID_WithTeam(t *testing.T) {
	c := &Client{teamID: "team_abc1"}
	result := c.addTeamID("/v10/projects")
	if result != "/v10/projects?teamId=team_abc1" {
		t.Errorf("expected teamId appended, got %s", result)
	}
}

func TestAddTeamID_WithExistingQuery(t *testing.T) {
	c := &Client{teamID: "team_abc1"}
	result := c.addTeamID("/v10/projects?limit=20")
	if result != "/v10/projects?limit=20&teamId=team_abc1" {
		t.Errorf("expected & separator, got %s", result)
	}
}

func TestVercelError_Error(t *testing.T) {
	err := &VercelError{StatusCode: 404, Message: "not found"}
	msg := err.Error()
	mustContain(t, msg, "404")
	mustContain(t, msg, "not found")
}

func TestClient_Do_APIError_WithErrorBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v10/projects", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"code":"forbidden","message":"token is not authorized"}}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("projects", newProjectsListCmd(factory)))
	err := runCmdErr(t, root, "projects", "list")
	if err == nil {
		t.Fatal("expected error from API")
	}
	mustContain(t, err.Error(), "token is not authorized")
}

func TestClient_Do_APIError_PlainBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v10/projects", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("projects", newProjectsListCmd(factory)))
	err := runCmdErr(t, root, "projects", "list")
	if err == nil {
		t.Fatal("expected error from API")
	}
	mustContain(t, err.Error(), "500")
}

func TestClient_DoJSON_EmptyResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v9/projects/empty-proj", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := &Client{
		http:    server.Client(),
		baseURL: server.URL,
	}
	err := c.doJSON(context.Background(), http.MethodGet, "/v9/projects/empty-proj", nil, nil)
	// nil result should not error
	if err != nil {
		t.Errorf("expected nil error for nil result, got %v", err)
	}
}
