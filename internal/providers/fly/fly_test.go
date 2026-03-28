package fly

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
)

func TestProvider_Name(t *testing.T) {
	p := New()
	if p.Name() != "fly" {
		t.Errorf("expected Name()=fly, got %s", p.Name())
	}
}

func TestProvider_RegisterCommands(t *testing.T) {
	p := New()
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	p.RegisterCommands(root)

	flyCmd, _, err := root.Find([]string{"fly"})
	if err != nil || flyCmd == nil {
		t.Error("expected fly subcommand to be registered")
	}
}

func TestProvider_RegisterCommands_Subcommands(t *testing.T) {
	p := New()
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	p.RegisterCommands(root)

	subcommands := [][]string{
		{"fly", "apps"},
		{"fly", "machines"},
		{"fly", "volumes"},
		{"fly", "certs"},
		{"fly", "secrets"},
		{"fly", "regions"},
	}
	for _, path := range subcommands {
		cmd, _, err := root.Find(path)
		if err != nil || cmd == nil {
			t.Errorf("expected subcommand %v to be registered", path)
		}
	}
}

func TestProvider_Alias(t *testing.T) {
	p := New()
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	p.RegisterCommands(root)

	// The fly command has alias "f"
	flyCmd, _, err := root.Find([]string{"f"})
	if err != nil || flyCmd == nil {
		t.Error("expected 'f' alias to resolve to fly subcommand")
	}
}

func TestFlyError_Error(t *testing.T) {
	err := &FlyError{StatusCode: 404, Message: "app not found"}
	msg := err.Error()
	mustContain(t, msg, "404")
	mustContain(t, msg, "app not found")
}

func TestFlyError_Error_500(t *testing.T) {
	err := &FlyError{StatusCode: 500, Message: "internal server error"}
	msg := err.Error()
	mustContain(t, msg, "500")
	mustContain(t, msg, "internal server error")
}

func TestClient_Do_APIError_WithJSONBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/apps", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"unauthorized"}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("apps", newAppsListCmd(factory)))
	err := runCmdErr(t, root, "apps", "list")
	if err == nil {
		t.Fatal("expected error from API")
	}
	mustContain(t, err.Error(), "unauthorized")
}

func TestClient_Do_APIError_PlainBody(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/apps", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("apps", newAppsListCmd(factory)))
	err := runCmdErr(t, root, "apps", "list")
	if err == nil {
		t.Fatal("expected error from API")
	}
	mustContain(t, err.Error(), "500")
}

func TestClient_DoJSON_EmptyResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/apps/empty-app", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := &Client{
		http:    server.Client(),
		baseURL: server.URL,
	}
	err := c.doJSON(context.Background(), http.MethodGet, "/v1/apps/empty-app", nil, nil)
	if err != nil {
		t.Errorf("expected nil error for empty response with nil result, got %v", err)
	}
}

func TestClient_GraphQL_ErrorResponse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"errors":[{"message":"app not found"}]}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("secrets", newSecretsListCmd(factory)))
	err := runCmdErr(t, root, "secrets", "list", "--app", "nonexistent-app")
	if err == nil {
		t.Fatal("expected graphql error")
	}
	mustContain(t, err.Error(), "app not found")
}

func TestClient_GraphQL_HTTPError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized"))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	factory := newTestClientFactory(server)
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("secrets", newSecretsListCmd(factory)))
	err := runCmdErr(t, root, "secrets", "list", "--app", "my-app")
	if err == nil {
		t.Fatal("expected HTTP error from graphql endpoint")
	}
	mustContain(t, err.Error(), "401")
}

func TestClient_GraphQL_MalformedEnvelope(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not valid json`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	c := &Client{
		http:       server.Client(),
		baseURL:    server.URL,
		graphqlURL: server.URL + "/graphql",
	}
	var result any
	err := c.graphQL(context.Background(), `query { platform { regions { code name } } }`, nil, &result)
	if err == nil {
		t.Fatal("expected error for malformed graphql envelope")
	}
	mustContain(t, err.Error(), "decoding graphql envelope")
}

func TestClientFactory_Error(t *testing.T) {
	factory := ClientFactory(func(ctx context.Context) (*Client, error) {
		return nil, &FlyError{StatusCode: 401, Message: "no token"}
	})
	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("apps", newAppsListCmd(factory)))
	err := runCmdErr(t, root, "apps", "list")
	if err == nil {
		t.Fatal("expected factory error to propagate")
	}
	mustContain(t, err.Error(), "no token")
}
