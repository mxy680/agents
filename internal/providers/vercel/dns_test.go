package vercel

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newDNSTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("dns",
		newDNSListCmd(factory),
		newDNSAddCmd(factory),
		newDNSRemoveCmd(factory),
	)
}

func TestDnsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "list", "--domain", "example.com", "--json")

	var records []DNSRecord
	if err := json.Unmarshal([]byte(output), &records); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 DNS records, got %d", len(records))
	}
	if records[0].ID != "rec_abc1" {
		t.Errorf("expected first record ID=rec_abc1, got %s", records[0].ID)
	}
	if records[0].Name != "www" {
		t.Errorf("expected first record Name=www, got %s", records[0].Name)
	}
	if records[0].Type != "CNAME" {
		t.Errorf("expected first record Type=CNAME, got %s", records[0].Type)
	}
	if records[0].Value != "cname.vercel-dns.com" {
		t.Errorf("expected first record Value=cname.vercel-dns.com, got %s", records[0].Value)
	}
	if records[1].Type != "A" {
		t.Errorf("expected second record Type=A, got %s", records[1].Type)
	}
}

func TestDnsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "list", "--domain", "example.com")

	mustContain(t, output, "rec_abc1")
	mustContain(t, output, "CNAME")
	mustContain(t, output, "cname.vercel-dns.com")
}

func TestDnsAdd_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "add",
		"--domain", "example.com",
		"--type", "TXT",
		"--name", "verify",
		"--value", "google-site-verification=abc123",
		"--json",
	)

	var rec DNSRecord
	if err := json.Unmarshal([]byte(output), &rec); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if rec.ID != "rec_new1" {
		t.Errorf("expected ID=rec_new1, got %s", rec.ID)
	}
	if rec.Name != "verify" {
		t.Errorf("expected Name=verify, got %s", rec.Name)
	}
	if rec.Type != "TXT" {
		t.Errorf("expected Type=TXT, got %s", rec.Type)
	}
}

func TestDnsAdd_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	output := runCmd(t, root, "dns", "add",
		"--domain", "example.com",
		"--type", "TXT",
		"--name", "verify",
		"--value", "somevalue",
		"--dry-run",
	)

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "example.com")
}

func TestDnsRemove_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))

	var execErr error
	output := captureStdout(t, func() {
		root.SetArgs([]string{"dns", "remove", "--domain", "example.com", "--id", "rec_abc1", "--confirm"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	mustContain(t, output, "Deleted")
	mustContain(t, output, "rec_abc1")
}

func TestDnsRemove_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDNSTestCmd(factory))
	err := runCmdErr(t, root, "dns", "remove", "--domain", "example.com", "--id", "rec_abc1")

	if err == nil {
		t.Fatal("expected error when --confirm is absent")
	}
	mustContain(t, err.Error(), "irreversible")
}
