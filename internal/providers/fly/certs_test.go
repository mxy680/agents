package fly

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newCertsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("certs",
		newCertsListCmd(factory),
		newCertsGetCmd(factory),
		newCertsAddCmd(factory),
		newCertsCheckCmd(factory),
		newCertsRemoveCmd(factory),
	)
}

func TestCertsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "list", "--app", "my-app")

	mustContain(t, output, "example.com")
	mustContain(t, output, "Ready")
	mustContain(t, output, "www.example.com")
	mustContain(t, output, "Pending")
}

func TestCertsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "list", "--app", "my-app", "--json")

	var results []CertSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 certs, got %d", len(results))
	}
	if results[0].Hostname != "example.com" {
		t.Errorf("expected first cert Hostname=example.com, got %s", results[0].Hostname)
	}
	if results[0].ClientStatus != "Ready" {
		t.Errorf("expected first cert ClientStatus=Ready, got %s", results[0].ClientStatus)
	}
	if !results[0].Issued {
		t.Error("expected first cert Issued=true")
	}
	if results[1].Hostname != "www.example.com" {
		t.Errorf("expected second cert Hostname=www.example.com, got %s", results[1].Hostname)
	}
	if results[1].Issued {
		t.Error("expected second cert Issued=false")
	}
}

func TestCertsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "get", "--app", "my-app", "--hostname", "example.com")

	mustContain(t, output, "example.com")
	mustContain(t, output, "Ready")
	mustContain(t, output, "true") // issued
	mustContain(t, output, "_acme-challenge.example.com")
}

func TestCertsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "get", "--app", "my-app", "--hostname", "example.com", "--json")

	var detail CertDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.Hostname != "example.com" {
		t.Errorf("expected Hostname=example.com, got %s", detail.Hostname)
	}
	if detail.ClientStatus != "Ready" {
		t.Errorf("expected ClientStatus=Ready, got %s", detail.ClientStatus)
	}
	if !detail.Issued {
		t.Error("expected Issued=true")
	}
	if !detail.AcmeDNSConfigured {
		t.Error("expected AcmeDNSConfigured=true")
	}
	if !detail.AcmeALPNConfigured {
		t.Error("expected AcmeALPNConfigured=true")
	}
	if detail.DNSValidationTarget != "_acme-challenge.example.com" {
		t.Errorf("expected DNSValidationTarget=_acme-challenge.example.com, got %s", detail.DNSValidationTarget)
	}
}

func TestCertsAdd_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "add", "--app", "my-app", "--hostname", "example.com")

	mustContain(t, output, "Added certificate")
	mustContain(t, output, "example.com")
}

func TestCertsAdd_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "add", "--app", "my-app", "--hostname", "new.example.com", "--json")

	var detail CertDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.Hostname != "new.example.com" {
		t.Errorf("expected Hostname=new.example.com, got %s", detail.Hostname)
	}
	if detail.ClientStatus != "Pending" {
		t.Errorf("expected ClientStatus=Pending, got %s", detail.ClientStatus)
	}
	if detail.Issued {
		t.Error("expected Issued=false for newly added cert")
	}
}

func TestCertsAdd_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "add", "--app", "my-app", "--hostname", "example.com", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "example.com")
}

func TestCertsAdd_DryRun_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "add", "--app", "my-app", "--hostname", "example.com", "--dry-run", "--json")

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON dry-run output, got: %s\nerror: %v", output, err)
	}
	if result["action"] != "add" {
		t.Errorf("expected action=add, got %v", result["action"])
	}
	if result["hostname"] != "example.com" {
		t.Errorf("expected hostname=example.com, got %v", result["hostname"])
	}
}

func TestCertsCheck_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "check", "--app", "my-app", "--hostname", "example.com")

	mustContain(t, output, "example.com")
	mustContain(t, output, "Ready")
}

func TestCertsCheck_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "check", "--app", "my-app", "--hostname", "example.com", "--json")

	var detail CertDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.Hostname != "example.com" {
		t.Errorf("expected Hostname=example.com, got %s", detail.Hostname)
	}
	if !detail.Issued {
		t.Error("expected Issued=true after check")
	}
}

func TestCertsRemove_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "remove", "--app", "my-app", "--hostname", "example.com", "--confirm")

	mustContain(t, output, "Removed certificate")
	mustContain(t, output, "example.com")
}

func TestCertsRemove_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	err := runCmdErr(t, root, "certs", "remove", "--app", "my-app", "--hostname", "example.com")

	if err == nil {
		t.Fatal("expected error when --confirm is absent")
	}
	mustContain(t, err.Error(), "irreversible")
}

func TestCertsRemove_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "remove", "--app", "my-app", "--hostname", "example.com", "--confirm", "--json")

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if result["status"] != "removed" {
		t.Errorf("expected status=removed, got %s", result["status"])
	}
	if result["hostname"] != "example.com" {
		t.Errorf("expected hostname=example.com, got %s", result["hostname"])
	}
}

func TestCertsRemove_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "remove", "--app", "my-app", "--hostname", "example.com", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "example.com")
}
