package vercel

import (
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

func newDomainsTestCmd(factory ClientFactory) *cobra.Command {
	return buildTestCmd("domains",
		newDomainsListCmd(factory),
		newDomainsGetCmd(factory),
		newDomainsAddCmd(factory),
		newDomainsVerifyCmd(factory),
		newDomainsRemoveCmd(factory),
	)
}

func TestDomainsList_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDomainsTestCmd(factory))
	output := runCmd(t, root, "domains", "list")

	mustContain(t, output, "example.com")
	mustContain(t, output, "example.org")
}

func TestDomainsList_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDomainsTestCmd(factory))
	output := runCmd(t, root, "domains", "list", "--json")

	var results []DomainSummary
	if err := json.Unmarshal([]byte(output), &results); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(results))
	}
	if results[0].Name != "example.com" {
		t.Errorf("expected first domain Name=example.com, got %s", results[0].Name)
	}
	if !results[0].Verified {
		t.Error("expected first domain to be verified")
	}
	if results[1].Name != "example.org" {
		t.Errorf("expected second domain Name=example.org, got %s", results[1].Name)
	}
	if results[1].Verified {
		t.Error("expected second domain to NOT be verified")
	}
}

func TestDomainsGet_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDomainsTestCmd(factory))
	output := runCmd(t, root, "domains", "get", "--domain", "example.com", "--json")

	var detail DomainDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if detail.Name != "example.com" {
		t.Errorf("expected Name=example.com, got %s", detail.Name)
	}
	if detail.ID != "dom_abc1" {
		t.Errorf("expected ID=dom_abc1, got %s", detail.ID)
	}
	if !detail.Verified {
		t.Error("expected domain to be verified")
	}
	if detail.ServiceType != "external" {
		t.Errorf("expected ServiceType=external, got %s", detail.ServiceType)
	}
	if len(detail.Nameservers) != 2 {
		t.Errorf("expected 2 nameservers, got %d", len(detail.Nameservers))
	}
}

func TestDomainsGet_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDomainsTestCmd(factory))
	output := runCmd(t, root, "domains", "get", "--domain", "example.com")

	mustContain(t, output, "example.com")
	mustContain(t, output, "external")
	mustContain(t, output, "ns1.vercel.com")
}

func TestDomainsAdd_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDomainsTestCmd(factory))
	output := runCmd(t, root, "domains", "add", "--domain", "newdomain.com", "--json")

	var data map[string]any
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if name, _ := data["name"].(string); name != "newdomain.com" {
		t.Errorf("expected name=newdomain.com, got %v", data["name"])
	}
}

func TestDomainsAdd_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDomainsTestCmd(factory))
	output := runCmd(t, root, "domains", "add", "--domain", "newdomain.com", "--dry-run")

	mustContain(t, output, "DRY RUN")
	mustContain(t, output, "newdomain.com")
}

func TestDomainsVerify_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDomainsTestCmd(factory))
	output := runCmd(t, root, "domains", "verify", "--domain", "example.com", "--json")

	var data map[string]any
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		t.Fatalf("expected JSON output, got: %s\nerror: %v", output, err)
	}
	if verified, _ := data["verified"].(bool); !verified {
		t.Error("expected verified=true")
	}
}

func TestDomainsRemove_Confirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDomainsTestCmd(factory))
	output := runCmd(t, root, "domains", "remove", "--domain", "example.com", "--confirm")

	mustContain(t, output, "Removed")
	mustContain(t, output, "example.com")
}

func TestDomainsVerify_Text_Verified(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDomainsTestCmd(factory))
	output := runCmd(t, root, "domains", "verify", "--domain", "example.com")

	mustContain(t, output, "verified")
}

func TestDomainsAdd_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDomainsTestCmd(factory))
	output := runCmd(t, root, "domains", "add", "--domain", "newdomain.com")

	mustContain(t, output, "Added domain")
	mustContain(t, output, "newdomain.com")
}

func TestDomainsRemove_NoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(newDomainsTestCmd(factory))
	err := runCmdErr(t, root, "domains", "remove", "--domain", "example.com")

	if err == nil {
		t.Fatal("expected error when --confirm is absent")
	}
	mustContain(t, err.Error(), "irreversible")
}
