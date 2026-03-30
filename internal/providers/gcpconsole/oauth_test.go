package gcpconsole

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// newOAuthTestCmd builds the "oauth" subcommand tree backed by the given factory.
func newOAuthTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("oauth",
		newOAuthListCmd(factory),
		newOAuthGetCmd(factory),
		newOAuthCreateCmd(factory),
		newOAuthUpdateCmd(factory),
		newOAuthDeleteCmd(factory),
	)
}

// --- List ---

func TestOAuthList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "list", "--project-number", "123456789012", "--json")

	var clients []OAuthClient
	if err := json.Unmarshal([]byte(output), &clients); err != nil {
		t.Fatalf("expected JSON array, got: %s\nerror: %v", output, err)
	}
	assert.Len(t, clients, 2)
	assert.Equal(t, "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com", clients[0].ClientID)
	assert.Equal(t, "My Web App", clients[0].DisplayName)
	assert.Equal(t, "WEB", clients[0].Type)
}

func TestOAuthList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "list", "--project-number", "123456789012")

	mustContain(t, output, "My Web App")
	mustContain(t, output, "Mobile Client")
	mustContain(t, output, "WEB")
}

func TestOAuthList_Empty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/clients", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"clients": []any{}})
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "list", "--project-number", "123456789012")

	mustContain(t, output, "No OAuth clients found.")
}

// --- Get ---

func TestOAuthGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	clientID := "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com"
	output := runCmd(t, root, "oauth", "get",
		"--client-id", clientID,
		"--json",
	)

	var client OAuthClient
	if err := json.Unmarshal([]byte(output), &client); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, clientID, client.ClientID)
	assert.Equal(t, "My Web App", client.DisplayName)
	assert.Equal(t, "WEB", client.Type)
	assert.Len(t, client.RedirectURIs, 2)
	assert.Len(t, client.ClientSecrets, 1)
	assert.Equal(t, "ENABLED", client.ClientSecrets[0].State)
}

func TestOAuthGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	clientID := "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com"
	output := runCmd(t, root, "oauth", "get", "--client-id", clientID)

	mustContain(t, output, "Client ID:")
	mustContain(t, output, clientID)
	mustContain(t, output, "My Web App")
	mustContain(t, output, "Redirect URIs:")
	mustContain(t, output, "https://example.com/callback")
	mustContain(t, output, "Client Secrets:")
}

// --- Create ---

func TestOAuthCreate_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "create",
		"--project-number", "123456789012",
		"--name", "New App",
		"--redirect-uris", "https://newapp.example.com/callback",
		"--json",
	)

	var client OAuthClient
	if err := json.Unmarshal([]byte(output), &client); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "New App", client.DisplayName)
	assert.NotEmpty(t, client.ClientID)
	assert.NotEmpty(t, client.ClientSecrets)
}

func TestOAuthCreate_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "create",
		"--project-number", "123456789012",
		"--name", "New App",
		"--redirect-uris", "https://newapp.example.com/callback",
	)

	mustContain(t, output, "Created OAuth client:")
	mustContain(t, output, "Client Secret:")
}

func TestOAuthCreate_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "create",
		"--project-number", "123456789012",
		"--name", "New App",
		"--redirect-uris", "https://newapp.example.com/callback",
		"--dry-run",
	)

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output for dry-run, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "create", result["action"])
	assert.Equal(t, "123456789012", result["projectNumber"])
	assert.Equal(t, "New App", result["displayName"])
}

func TestOAuthCreate_MultipleRedirectURIs(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "create",
		"--project-number", "123456789012",
		"--name", "Multi URI App",
		"--redirect-uris", "https://app.example.com/callback, https://app.example.com/auth",
		"--dry-run",
	)

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	uris, _ := result["redirectUris"].([]any)
	assert.Len(t, uris, 2)
}

// --- Update ---

func TestOAuthUpdate_AddRedirect(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	clientID := "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com"
	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "update",
		"--client-id", clientID,
		"--add-redirect", "https://example.com/new-callback",
		"--json",
	)

	var client OAuthClient
	if err := json.Unmarshal([]byte(output), &client); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, clientID, client.ClientID)

	// The PUT handler echoes back the sent redirectUris
	found := false
	for _, uri := range client.RedirectURIs {
		if uri == "https://example.com/new-callback" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected new redirect URI to be present in response")
}

func TestOAuthUpdate_RemoveRedirect(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	clientID := "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com"
	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "update",
		"--client-id", clientID,
		"--remove-redirect", "https://example.com/auth",
		"--json",
	)

	var client OAuthClient
	if err := json.Unmarshal([]byte(output), &client); err != nil {
		t.Fatalf("expected JSON object, got: %s\nerror: %v", output, err)
	}

	// The mock PUT echoes the sent redirectUris — verify the removed one is absent
	for _, uri := range client.RedirectURIs {
		if uri == "https://example.com/auth" {
			t.Errorf("expected 'https://example.com/auth' to be removed, but it is still present")
		}
	}
}

func TestOAuthUpdate_DryRun_AddRedirect(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	clientID := "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com"
	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "update",
		"--client-id", clientID,
		"--add-redirect", "https://example.com/dry",
		"--dry-run",
	)

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output for dry-run, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "update", result["action"])
	assert.Equal(t, clientID, result["clientId"])
}

func TestOAuthUpdate_NoFlags_Error(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	clientID := "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com"
	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	err := runCmdErr(t, root, "oauth", "update", "--client-id", clientID)

	assert.Error(t, err)
	mustContain(t, err.Error(), "--add-redirect or --remove-redirect")
}

// --- Delete ---

func TestOAuthDelete_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	clientID := "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com"
	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "delete",
		"--client-id", clientID,
		"--confirm",
	)

	mustContain(t, output, "Deleted OAuth client:")
	mustContain(t, output, clientID)
}

func TestOAuthDelete_Confirm_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	clientID := "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com"
	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "delete",
		"--client-id", clientID,
		"--confirm",
		"--json",
	)

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "deleted", result["status"])
	assert.Equal(t, clientID, result["clientId"])
}

func TestOAuthDelete_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	clientID := "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com"
	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	err := runCmdErr(t, root, "oauth", "delete", "--client-id", clientID)

	assert.Error(t, err)
	mustContain(t, err.Error(), "irreversible")
}

func TestOAuthDelete_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	clientID := "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com"
	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	output := runCmd(t, root, "oauth", "delete",
		"--client-id", clientID,
		"--dry-run",
	)

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output for dry-run, got: %s\nerror: %v", output, err)
	}
	assert.Equal(t, "delete", result["action"])
	assert.Equal(t, clientID, result["clientId"])
}

// --- Error handling ---

func TestOAuthList_APIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/clients", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"error":{"code":401,"message":"Request had invalid authentication credentials","status":"UNAUTHENTICATED"}}`)
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	err := runCmdErr(t, root, "oauth", "list", "--project-number", "123456789012")

	assert.Error(t, err)
	mustContain(t, err.Error(), "invalid authentication credentials")
}

func TestOAuthGet_APIError_PlainBody(t *testing.T) {
	clientID := "123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com"
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/clients/"+clientID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "client not found")
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newOAuthTestCmd(factory))
	err := runCmdErr(t, root, "oauth", "get", "--client-id", clientID)

	assert.Error(t, err)
	mustContain(t, err.Error(), "404")
}

// --- splitAndTrim unit tests ---

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , b , c ", []string{"a", "b", "c"}},
		{"single", []string{"single"}},
		{"", []string{}},
		{"  ,  ,  ", []string{}},
	}

	for _, tt := range tests {
		result := splitAndTrim(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("splitAndTrim(%q): got %v, want %v", tt.input, result, tt.expected)
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("splitAndTrim(%q)[%d]: got %q, want %q", tt.input, i, result[i], tt.expected[i])
			}
		}
	}
}

// --- applyRedirectChanges unit tests ---

func TestApplyRedirectChanges_Add(t *testing.T) {
	existing := []string{"https://a.com", "https://b.com"}
	result := applyRedirectChanges(existing, "https://c.com", "")
	assert.Contains(t, result, "https://a.com")
	assert.Contains(t, result, "https://b.com")
	assert.Contains(t, result, "https://c.com")
	assert.Len(t, result, 3)
}

func TestApplyRedirectChanges_Remove(t *testing.T) {
	existing := []string{"https://a.com", "https://b.com", "https://c.com"}
	result := applyRedirectChanges(existing, "", "https://b.com")
	assert.Contains(t, result, "https://a.com")
	assert.NotContains(t, result, "https://b.com")
	assert.Contains(t, result, "https://c.com")
	assert.Len(t, result, 2)
}

func TestApplyRedirectChanges_AddDuplicate(t *testing.T) {
	existing := []string{"https://a.com", "https://b.com"}
	result := applyRedirectChanges(existing, "https://a.com", "")
	// Should not duplicate
	assert.Len(t, result, 2)
}

func TestApplyRedirectChanges_RemoveNonExistent(t *testing.T) {
	existing := []string{"https://a.com", "https://b.com"}
	result := applyRedirectChanges(existing, "", "https://z.com")
	assert.Len(t, result, 2)
}

func TestApplyRedirectChanges_AddAndRemove(t *testing.T) {
	existing := []string{"https://a.com", "https://b.com"}
	result := applyRedirectChanges(existing, "https://c.com", "https://a.com")
	assert.NotContains(t, result, "https://a.com")
	assert.Contains(t, result, "https://b.com")
	assert.Contains(t, result, "https://c.com")
	assert.Len(t, result, 2)
}
