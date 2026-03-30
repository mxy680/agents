package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderName(t *testing.T) {
	p := New()
	assert.Equal(t, "cloudflare", p.Name())
}

func TestRegisterCommands(t *testing.T) {
	p := New()
	root := newTestRootCmd()
	p.RegisterCommands(root)

	// Verify the cloudflare command is registered
	cfCmd, _, err := root.Find([]string{"cloudflare"})
	assert.NoError(t, err)
	assert.NotNil(t, cfCmd)
	assert.Equal(t, "cloudflare", cfCmd.Name())
}

func TestRegisterCommands_Alias(t *testing.T) {
	p := New()
	root := newTestRootCmd()
	p.RegisterCommands(root)

	// Verify the "cf" alias resolves
	cfCmd, _, err := root.Find([]string{"cf"})
	assert.NoError(t, err)
	assert.NotNil(t, cfCmd)
}

func TestClientFactory_Error(t *testing.T) {
	// A factory that always returns an error
	errFactory := ClientFactory(func(ctx context.Context) (*Client, error) {
		return nil, fmt.Errorf("auth failed")
	})

	root := newTestRootCmd()
	zonesCmd := buildTestCmd("zones", newZonesListCmd(errFactory))
	root.AddCommand(zonesCmd)

	err := runCmdErr(t, root, "zones", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth failed")
}

func TestClient_APIError(t *testing.T) {
	// Server that returns success=false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, `{"success":false,"result":null,"errors":[{"code":10000,"message":"Authentication error"}],"messages":[]}`)
	}))
	defer server.Close()

	factory := ClientFactory(func(ctx context.Context) (*Client, error) {
		return &Client{
			http:    server.Client(),
			baseURL: server.URL,
		}, nil
	})

	root := newTestRootCmd()
	zonesCmd := buildTestCmd("zones", newZonesListCmd(factory))
	root.AddCommand(zonesCmd)

	err := runCmdErr(t, root, "zones", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Authentication error")
}

func TestClient_AccountPath_Missing(t *testing.T) {
	// Factory with no account ID
	factory := ClientFactory(func(ctx context.Context) (*Client, error) {
		return &Client{
			http:      http.DefaultClient,
			baseURL:   "http://localhost",
			accountID: "",
		}, nil
	})

	root := newTestRootCmd()
	workersCmd := buildTestCmd("workers", newWorkersListCmd(factory))
	root.AddCommand(workersCmd)

	err := runCmdErr(t, root, "workers", "list")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CLOUDFLARE_ACCOUNT_ID")
}

func TestCloudflareError_WithErrors(t *testing.T) {
	e := &CloudflareError{
		StatusCode: 403,
		Errors: []CloudflareAPIError{
			{Code: 9109, Message: "Invalid access token"},
		},
	}
	assert.Contains(t, e.Error(), "403")
	assert.Contains(t, e.Error(), "Invalid access token")
}

func TestCloudflareError_NoErrors(t *testing.T) {
	e := &CloudflareError{StatusCode: 500}
	assert.Contains(t, e.Error(), "500")
}

func TestClient_Do_NonJSONError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "internal server error text")
	}))
	defer server.Close()

	client := &Client{
		http:    server.Client(),
		baseURL: server.URL,
	}

	_, err := client.do(context.Background(), http.MethodGet, "/some-path", nil)
	assert.Error(t, err)
	var cfErr *CloudflareError
	assert.ErrorAs(t, err, &cfErr)
	assert.Equal(t, 500, cfErr.StatusCode)
}

func TestClient_Do_NonJSONSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "plain text response")
	}))
	defer server.Close()

	client := &Client{
		http:    server.Client(),
		baseURL: server.URL,
	}

	raw, err := client.do(context.Background(), http.MethodGet, "/some-path", nil)
	assert.NoError(t, err)
	assert.Equal(t, "plain text response", string(raw))
}

func TestDryRunResult_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("zones",
		newZonesPurgeCacheCmd(factory),
	))

	output := runCmd(t, root, "zones", "purge-cache", "--zone", testZoneID, "--dry-run")
	mustContain(t, output, "DRY RUN")
	mustContain(t, output, testZoneID)
}

func TestConfirmDestructive_Required(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCmd("dns",
		newDNSDeleteCmd(factory),
	))

	err := runCmdErr(t, root, "dns", "delete",
		"--zone", testZoneID,
		"--record", "rec_abc1",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}
