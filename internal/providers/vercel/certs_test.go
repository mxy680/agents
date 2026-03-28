package vercel

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newCertsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("certs",
		newCertsListCmd(factory),
		newCertsGetCmd(factory),
	)
}

func TestCertsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "list", "--json")

	var certs []CertSummary
	if err := json.Unmarshal([]byte(output), &certs); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(certs) != 2 {
		t.Fatalf("expected 2 certs, got %d", len(certs))
	}
	if certs[0].ID != "cert_abc1" {
		t.Errorf("expected first cert ID=cert_abc1, got %s", certs[0].ID)
	}
	if !certs[0].AutoRenew {
		t.Error("expected first cert AutoRenew=true")
	}
	if len(certs[0].CNs) != 2 {
		t.Errorf("expected 2 CNs on first cert, got %d", len(certs[0].CNs))
	}
	if certs[1].ID != "cert_def2" {
		t.Errorf("expected second cert ID=cert_def2, got %s", certs[1].ID)
	}
	if certs[1].AutoRenew {
		t.Error("expected second cert AutoRenew=false")
	}
}

func TestCertsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "list")

	mustContain(t, output, "cert_abc1")
	mustContain(t, output, "example.com")
}

func TestCertsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "get", "--id", "cert_abc1", "--json")

	var cert CertSummary
	if err := json.Unmarshal([]byte(output), &cert); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if cert.ID != "cert_abc1" {
		t.Errorf("expected ID=cert_abc1, got %s", cert.ID)
	}
	if !cert.AutoRenew {
		t.Error("expected AutoRenew=true")
	}
	if len(cert.CNs) != 2 {
		t.Errorf("expected 2 CNs, got %d", len(cert.CNs))
	}
}

func TestCertsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newCertsTestCmd(factory))
	output := runCmd(t, root, "certs", "get", "--id", "cert_abc1")

	mustContain(t, output, "cert_abc1")
	mustContain(t, output, "example.com")
	mustContain(t, output, "www.example.com")
}
