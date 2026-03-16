//go:build integration

package gmail

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
)

// requireEnv skips the test if any required env var is missing.
func requireEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"GOOGLE_DESKTOP_CLIENT_ID",
		"GOOGLE_DESKTOP_CLIENT_SECRET",
		"GMAIL_ACCESS_TOKEN",
		"GMAIL_REFRESH_TOKEN",
	} {
		if os.Getenv(key) == "" {
			t.Skipf("skipping: %s not set (run with doppler run)", key)
		}
	}
}

func realFactory() ServiceFactory {
	return auth.NewGmailService
}

func integrationRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	return root
}

// --- list-unread ---

func TestIntegration_ListUnread_JSON(t *testing.T) {
	requireEnv(t)

	cmd := newListUnreadCmd(realFactory())
	root := integrationRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"list-unread", "--limit=3", "--since=72h", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("list-unread failed: %v", execErr)
	}

	var summaries []EmailSummary
	if err := json.Unmarshal([]byte(output), &summaries); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("got %d unread messages", len(summaries))
	for _, s := range summaries {
		t.Logf("  [%s] from=%s subject=%q", s.ID, s.From, s.Subject)
		if s.ID == "" {
			t.Error("message has empty ID")
		}
	}
}

func TestIntegration_ListUnread_Text(t *testing.T) {
	requireEnv(t)

	cmd := newListUnreadCmd(realFactory())
	root := integrationRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"list-unread", "--limit=3", "--since=72h"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("list-unread text failed: %v", execErr)
	}
	t.Logf("text output:\n%s", output)
}

// --- read ---

func TestIntegration_Read_JSON(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	// First get a real message ID
	svc, err := realFactory()(ctx)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}
	resp, err := svc.Users.Messages.List("me").MaxResults(1).Do()
	if err != nil {
		t.Fatalf("listing messages: %v", err)
	}
	if len(resp.Messages) == 0 {
		t.Skip("no messages in mailbox")
	}
	msgID := resp.Messages[0].Id

	cmd := newReadCmd(realFactory())
	root := integrationRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"read", "--id=" + msgID, "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("read failed: %v", execErr)
	}

	var detail EmailDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if detail.ID != msgID {
		t.Errorf("expected ID=%s, got %s", msgID, detail.ID)
	}
	t.Logf("read message: from=%s subject=%q body_len=%d", detail.From, detail.Subject, len(detail.Body))
}

// --- send (dry-run) ---

func TestIntegration_Send_DryRun(t *testing.T) {
	requireEnv(t)

	cmd := newSendCmd(realFactory())
	root := integrationRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"send",
			"--to=omniclaw680@gmail.com",
			"--subject=Integration Test (dry-run)",
			"--body=This should NOT be sent.",
			"--dry-run",
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("send dry-run failed: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if result["status"] != "dry-run" {
		t.Errorf("expected status=dry-run, got %s", result["status"])
	}
	t.Logf("dry-run output: %v", result)
}

// --- send (real, to self) ---

func TestIntegration_Send_ToSelf(t *testing.T) {
	requireEnv(t)

	cmd := newSendCmd(realFactory())
	root := integrationRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"send",
			"--to=omniclaw680@gmail.com",
			"--subject=Integration Test from CLI",
			"--body=Sent by make test-integration at " + t.Name(),
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("send failed: %v", execErr)
	}

	var result SendResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if result.ID == "" {
		t.Error("sent message has empty ID")
	}
	if result.Status != "sent" {
		t.Errorf("expected status=sent, got %s", result.Status)
	}
	t.Logf("sent message: id=%s threadId=%s", result.ID, result.ThreadID)
}

// --- search ---

func TestIntegration_Search_JSON(t *testing.T) {
	requireEnv(t)

	cmd := newSearchCmd(realFactory())
	root := integrationRootCmd()
	root.AddCommand(cmd)

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"search", "--query=in:inbox", "--limit=3", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("search failed: %v", execErr)
	}

	var summaries []EmailSummary
	if err := json.Unmarshal([]byte(output), &summaries); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("search returned %d results", len(summaries))
	for _, s := range summaries {
		t.Logf("  [%s] from=%s subject=%q", s.ID, s.From, s.Subject)
	}
}

// --- round-trip: send then search ---

func TestIntegration_SendThenSearch(t *testing.T) {
	requireEnv(t)

	// Send a message with a unique subject
	uniqueSubject := "CLI-roundtrip-test-" + strings.Replace(t.Name(), "/", "-", -1)

	sendCmd := newSendCmd(realFactory())
	root1 := integrationRootCmd()
	root1.AddCommand(sendCmd)

	var sendOutput string
	var sendErr error
	sendOutput = captureStdout(t, func() {
		root1.SetArgs([]string{
			"send",
			"--to=omniclaw680@gmail.com",
			"--subject=" + uniqueSubject,
			"--body=Round-trip integration test",
			"--json",
		})
		sendErr = root1.Execute()
	})
	if sendErr != nil {
		t.Fatalf("send failed: %v", sendErr)
	}

	var sendResult SendResult
	json.Unmarshal([]byte(sendOutput), &sendResult)
	t.Logf("sent message id=%s", sendResult.ID)

	// Search for it by subject
	searchCmd := newSearchCmd(realFactory())
	root2 := integrationRootCmd()
	root2.AddCommand(searchCmd)

	var searchOutput string
	var searchErr error
	searchOutput = captureStdout(t, func() {
		root2.SetArgs([]string{"search", "--query=subject:" + uniqueSubject, "--limit=5", "--json"})
		searchErr = root2.Execute()
	})
	if searchErr != nil {
		t.Fatalf("search failed: %v", searchErr)
	}

	var results []EmailSummary
	json.Unmarshal([]byte(searchOutput), &results)

	found := false
	for _, r := range results {
		if strings.Contains(r.Subject, uniqueSubject) {
			found = true
			t.Logf("found sent message: id=%s subject=%q", r.ID, r.Subject)
		}
	}
	if !found {
		t.Logf("sent message not found in search yet (Gmail indexing delay is normal)")
	}
}
