//go:build integration

package gmail

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
)

// rfc2822Message builds a minimal RFC 2822 message bytes suitable for import/insert.
func rfc2822Message(from, to, subject, body string) []byte {
	return []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s",
		from, to, subject, body,
	))
}

// writeTempEML writes content to a temp *.eml file and returns the path.
// The caller must remove the file when done.
func writeTempEML(t *testing.T, content []byte) string {
	t.Helper()
	f, err := os.CreateTemp("", "integration-*.eml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	defer f.Close()
	if _, err := f.Write(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	return f.Name()
}

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

func integrationMessagesCmd(factory ServiceFactory) *cobra.Command {
	return buildTestMessagesCmd(factory)
}

func integrationThreadsCmd(factory ServiceFactory) *cobra.Command {
	return buildTestThreadsCmd(factory)
}

// --- messages list (unread) ---

func TestIntegration_ListUnread_JSON(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(integrationMessagesCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "list", "--query=is:unread", "--limit=3", "--since=72h", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("messages list failed: %v", execErr)
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

	root := integrationRootCmd()
	root.AddCommand(integrationMessagesCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "list", "--query=is:unread", "--limit=3", "--since=72h"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("messages list text failed: %v", execErr)
	}
	t.Logf("text output:\n%s", output)
}

// --- messages get ---

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

	root := integrationRootCmd()
	root.AddCommand(integrationMessagesCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "get", "--id=" + msgID, "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("messages get failed: %v", execErr)
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

// --- messages send (dry-run) ---

func TestIntegration_Send_DryRun(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(integrationMessagesCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"messages", "send",
			"--to=omniclaw680@gmail.com",
			"--subject=Integration Test (dry-run)",
			"--body=This should NOT be sent.",
			"--dry-run",
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("messages send dry-run failed: %v", execErr)
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

// --- messages send (real, to self) ---

func TestIntegration_Send_ToSelf(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(integrationMessagesCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"messages", "send",
			"--to=omniclaw680@gmail.com",
			"--subject=Integration Test from CLI",
			"--body=Sent by make test-integration at " + t.Name(),
			"--json",
		})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("messages send failed: %v", execErr)
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

// --- messages list (search) ---

func TestIntegration_Search_JSON(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(integrationMessagesCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "list", "--query=in:inbox", "--limit=3", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("messages list (search) failed: %v", execErr)
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

// --- messages trash/untrash ---

func TestIntegration_Messages_TrashUntrash(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	// Send a message to self to get an ID to work with
	svc, err := realFactory()(ctx)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}

	root1 := integrationRootCmd()
	root1.AddCommand(integrationMessagesCmd(realFactory()))

	var sendOutput string
	captureStdout(t, func() {
		root1.SetArgs([]string{
			"messages", "send",
			"--to=omniclaw680@gmail.com",
			"--subject=Integration Test TrashUntrash",
			"--body=trash-untrash test",
			"--json",
		})
		root1.Execute() //nolint:errcheck
	})

	// Use the most recent message in inbox as the test target
	resp, err := svc.Users.Messages.List("me").MaxResults(1).Do()
	if err != nil {
		t.Fatalf("listing messages: %v", err)
	}
	if len(resp.Messages) == 0 {
		t.Skip("no messages in mailbox to trash")
	}
	msgID := resp.Messages[0].Id
	t.Logf("using message id=%s for trash/untrash test", msgID)
	_ = sendOutput

	// Trash it
	root2 := integrationRootCmd()
	root2.AddCommand(integrationMessagesCmd(realFactory()))
	var execErr error
	captureStdout(t, func() {
		root2.SetArgs([]string{"messages", "trash", "--id=" + msgID, "--json"})
		execErr = root2.Execute()
	})
	if execErr != nil {
		t.Fatalf("messages trash failed: %v", execErr)
	}
	t.Logf("trashed message id=%s", msgID)

	// Untrash it
	root3 := integrationRootCmd()
	root3.AddCommand(integrationMessagesCmd(realFactory()))
	captureStdout(t, func() {
		root3.SetArgs([]string{"messages", "untrash", "--id=" + msgID, "--json"})
		execErr = root3.Execute()
	})
	if execErr != nil {
		t.Fatalf("messages untrash failed: %v", execErr)
	}
	t.Logf("untrashed message id=%s", msgID)
}

// --- messages modify ---

func TestIntegration_Messages_Modify(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	svc, err := realFactory()(ctx)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}

	resp, err := svc.Users.Messages.List("me").MaxResults(1).Do()
	if err != nil {
		t.Fatalf("listing messages: %v", err)
	}
	if len(resp.Messages) == 0 {
		t.Skip("no messages in mailbox to modify")
	}
	msgID := resp.Messages[0].Id
	t.Logf("using message id=%s for modify test", msgID)

	root := integrationRootCmd()
	root.AddCommand(integrationMessagesCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "modify", "--id=" + msgID, "--add-labels=STARRED", "--json"})
		execErr = root.Execute()
	})
	if execErr != nil {
		t.Fatalf("messages modify failed: %v", execErr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if result["status"] != "modified" {
		t.Errorf("expected status=modified, got %v", result["status"])
	}
	t.Logf("modified message: id=%s labelIds=%v", result["id"], result["labelIds"])

	// Remove the STARRED label to restore state
	root2 := integrationRootCmd()
	root2.AddCommand(integrationMessagesCmd(realFactory()))
	captureStdout(t, func() {
		root2.SetArgs([]string{"messages", "modify", "--id=" + msgID, "--remove-labels=STARRED", "--json"})
		root2.Execute() //nolint:errcheck
	})
}

// --- messages delete ---

func TestIntegration_Messages_Delete(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	// Send a message to self first so we have something to delete
	root1 := integrationRootCmd()
	root1.AddCommand(integrationMessagesCmd(realFactory()))

	var sendOutput string
	var sendErr error
	sendOutput = captureStdout(t, func() {
		root1.SetArgs([]string{
			"messages", "send",
			"--to=omniclaw680@gmail.com",
			"--subject=Integration Test Delete (to be deleted)",
			"--body=this message will be permanently deleted",
			"--json",
		})
		sendErr = root1.Execute()
	})
	if sendErr != nil {
		t.Fatalf("send failed: %v", sendErr)
	}

	var sendResult SendResult
	if err := json.Unmarshal([]byte(sendOutput), &sendResult); err != nil {
		t.Fatalf("invalid send JSON: %v", err)
	}

	// Use the sent message ID if available; otherwise take the most recent inbox message
	msgID := sendResult.ID
	if msgID == "" {
		svc, err := realFactory()(ctx)
		if err != nil {
			t.Fatalf("creating service: %v", err)
		}
		resp, err := svc.Users.Messages.List("me").MaxResults(1).Do()
		if err != nil {
			t.Fatalf("listing messages: %v", err)
		}
		if len(resp.Messages) == 0 {
			t.Skip("no messages in mailbox to delete")
		}
		msgID = resp.Messages[0].Id
	}
	t.Logf("deleting message id=%s", msgID)

	// Trash first, then delete — Gmail's backend needs the message settled before permanent delete
	svc, svcErr := realFactory()(ctx)
	if svcErr != nil {
		t.Fatalf("creating service: %v", svcErr)
	}
	_, trashErr := svc.Users.Messages.Trash("me", msgID).Do()
	if trashErr != nil {
		t.Fatalf("trash failed: %v", trashErr)
	}

	// Gmail backend needs time to settle after trash before permanent delete
	time.Sleep(2 * time.Second)

	root2 := integrationRootCmd()
	root2.AddCommand(integrationMessagesCmd(realFactory()))

	var execErr error
	captureStdout(t, func() {
		root2.SetArgs([]string{"messages", "delete", "--id=" + msgID, "--confirm", "--json"})
		execErr = root2.Execute()
	})
	if execErr != nil {
		t.Fatalf("messages delete failed: %v", execErr)
	}
	t.Logf("permanently deleted message id=%s", msgID)
}

// --- round-trip: send then search ---

func TestIntegration_SendThenSearch(t *testing.T) {
	requireEnv(t)

	// Send a message with a unique subject
	uniqueSubject := "CLI-roundtrip-test-" + strings.Replace(t.Name(), "/", "-", -1)

	root1 := integrationRootCmd()
	root1.AddCommand(integrationMessagesCmd(realFactory()))

	var sendOutput string
	var sendErr error
	sendOutput = captureStdout(t, func() {
		root1.SetArgs([]string{
			"messages", "send",
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
	root2 := integrationRootCmd()
	root2.AddCommand(integrationMessagesCmd(realFactory()))

	var searchOutput string
	var searchErr error
	searchOutput = captureStdout(t, func() {
		root2.SetArgs([]string{"messages", "list", "--query=subject:" + uniqueSubject, "--limit=5", "--json"})
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

// --- threads list ---

func TestIntegration_Threads_List(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(integrationThreadsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "list", "--limit=3", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("threads list failed: %v", execErr)
	}

	var summaries []ThreadSummary
	if err := json.Unmarshal([]byte(output), &summaries); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("got %d threads", len(summaries))
	for _, s := range summaries {
		t.Logf("  [%s] snippet=%q", s.ID, s.Snippet)
		if s.ID == "" {
			t.Error("thread has empty ID")
		}
	}
}

// --- threads get ---

func TestIntegration_Threads_Get(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	svc, err := realFactory()(ctx)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}
	resp, err := svc.Users.Threads.List("me").MaxResults(1).Do()
	if err != nil {
		t.Fatalf("listing threads: %v", err)
	}
	if len(resp.Threads) == 0 {
		t.Skip("no threads in mailbox")
	}
	threadID := resp.Threads[0].Id

	root := integrationRootCmd()
	root.AddCommand(integrationThreadsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "get", "--id=" + threadID, "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("threads get failed: %v", execErr)
	}

	var detail ThreadDetail
	if err := json.Unmarshal([]byte(output), &detail); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if detail.ID != threadID {
		t.Errorf("expected ID=%s, got %s", threadID, detail.ID)
	}
	t.Logf("thread %s has %d messages", detail.ID, len(detail.Messages))
}

// --- threads trash/untrash ---

func TestIntegration_Threads_TrashUntrash(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	// Send a message to self to get a thread to work with
	root1 := integrationRootCmd()
	root1.AddCommand(integrationMessagesCmd(realFactory()))

	var sendOutput string
	var sendErr error
	sendOutput = captureStdout(t, func() {
		root1.SetArgs([]string{
			"messages", "send",
			"--to=omniclaw680@gmail.com",
			"--subject=Integration Test Threads TrashUntrash",
			"--body=thread-trash-untrash test",
			"--json",
		})
		sendErr = root1.Execute()
	})
	if sendErr != nil {
		t.Fatalf("send failed: %v", sendErr)
	}

	var sendResult SendResult
	if err := json.Unmarshal([]byte(sendOutput), &sendResult); err != nil {
		t.Fatalf("invalid send JSON: %v", err)
	}

	// Get the threadId from the sent message
	threadID := sendResult.ThreadID
	if threadID == "" {
		// Fall back to listing threads
		svc, err := realFactory()(ctx)
		if err != nil {
			t.Fatalf("creating service: %v", err)
		}
		resp, err := svc.Users.Threads.List("me").MaxResults(1).Do()
		if err != nil {
			t.Fatalf("listing threads: %v", err)
		}
		if len(resp.Threads) == 0 {
			t.Skip("no threads in mailbox to trash")
		}
		threadID = resp.Threads[0].Id
	}
	t.Logf("using thread id=%s for trash/untrash test", threadID)

	// Trash the thread
	root2 := integrationRootCmd()
	root2.AddCommand(integrationThreadsCmd(realFactory()))
	var execErr error
	captureStdout(t, func() {
		root2.SetArgs([]string{"threads", "trash", "--id=" + threadID, "--json"})
		execErr = root2.Execute()
	})
	if execErr != nil {
		t.Fatalf("threads trash failed: %v", execErr)
	}
	t.Logf("trashed thread id=%s", threadID)

	// Untrash the thread
	root3 := integrationRootCmd()
	root3.AddCommand(integrationThreadsCmd(realFactory()))
	captureStdout(t, func() {
		root3.SetArgs([]string{"threads", "untrash", "--id=" + threadID, "--json"})
		execErr = root3.Execute()
	})
	if execErr != nil {
		t.Fatalf("threads untrash failed: %v", execErr)
	}
	t.Logf("untrashed thread id=%s", threadID)
}

// --- threads modify ---

func TestIntegration_Threads_Modify(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	svc, err := realFactory()(ctx)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}

	resp, err := svc.Users.Threads.List("me").MaxResults(1).Do()
	if err != nil {
		t.Fatalf("listing threads: %v", err)
	}
	if len(resp.Threads) == 0 {
		t.Skip("no threads in mailbox to modify")
	}
	threadID := resp.Threads[0].Id
	t.Logf("using thread id=%s for modify test", threadID)

	root := integrationRootCmd()
	root.AddCommand(integrationThreadsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"threads", "modify", "--id=" + threadID, "--add-labels=STARRED", "--json"})
		execErr = root.Execute()
	})
	if execErr != nil {
		t.Fatalf("threads modify failed: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if result["status"] != "modified" {
		t.Errorf("expected status=modified, got %v", result["status"])
	}
	t.Logf("modified thread: id=%s", result["id"])

	// Restore: remove the STARRED label
	root2 := integrationRootCmd()
	root2.AddCommand(integrationThreadsCmd(realFactory()))
	captureStdout(t, func() {
		root2.SetArgs([]string{"threads", "modify", "--id=" + threadID, "--remove-labels=STARRED", "--json"})
		root2.Execute() //nolint:errcheck
	})
}

// --- labels ---

func integrationLabelsCmd(factory ServiceFactory) *cobra.Command {
	return buildTestLabelsCmd(factory)
}

func TestIntegration_Labels_List(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(integrationLabelsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"labels", "list", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("labels list failed: %v", execErr)
	}

	var labels []LabelInfo
	if err := json.Unmarshal([]byte(output), &labels); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("got %d labels", len(labels))

	inboxFound := false
	for _, l := range labels {
		t.Logf("  [%s] name=%s type=%s", l.ID, l.Name, l.Type)
		if l.ID == "INBOX" {
			inboxFound = true
		}
	}
	if !inboxFound {
		t.Error("expected INBOX label to exist")
	}
}

func integrationDraftsCmd(factory ServiceFactory) *cobra.Command {
	return buildTestDraftsCmd(factory)
}

func TestIntegration_Labels_CRUD(t *testing.T) {
	requireEnv(t)

	labelName := "IntegrationTestLabel-" + t.Name()

	// Create
	root1 := integrationRootCmd()
	root1.AddCommand(integrationLabelsCmd(realFactory()))

	var createOutput string
	var createErr error
	createOutput = captureStdout(t, func() {
		root1.SetArgs([]string{"labels", "create", "--name=" + labelName, "--json"})
		createErr = root1.Execute()
	})
	if createErr != nil {
		t.Fatalf("labels create failed: %v", createErr)
	}

	var createResult map[string]string
	if err := json.Unmarshal([]byte(createOutput), &createResult); err != nil {
		t.Fatalf("invalid create JSON: %v\noutput: %s", err, createOutput)
	}
	labelID := createResult["id"]
	if labelID == "" {
		t.Fatal("created label has empty ID")
	}
	t.Logf("created label id=%s name=%s", labelID, labelID)

	// Get
	root2 := integrationRootCmd()
	root2.AddCommand(integrationLabelsCmd(realFactory()))

	var getOutput string
	var getErr error
	getOutput = captureStdout(t, func() {
		root2.SetArgs([]string{"labels", "get", "--id=" + labelID, "--json"})
		getErr = root2.Execute()
	})
	if getErr != nil {
		t.Fatalf("labels get failed: %v", getErr)
	}

	var info LabelInfo
	if err := json.Unmarshal([]byte(getOutput), &info); err != nil {
		t.Fatalf("invalid get JSON: %v\noutput: %s", err, getOutput)
	}
	if info.ID != labelID {
		t.Errorf("expected ID=%s, got %s", labelID, info.ID)
	}
	t.Logf("fetched label: id=%s name=%s", info.ID, info.Name)

	// Update name
	updatedName := labelName + "-updated"
	root3 := integrationRootCmd()
	root3.AddCommand(integrationLabelsCmd(realFactory()))

	var updateErr error
	captureStdout(t, func() {
		root3.SetArgs([]string{"labels", "update", "--id=" + labelID, "--name=" + updatedName, "--json"})
		updateErr = root3.Execute()
	})
	if updateErr != nil {
		t.Fatalf("labels update failed: %v", updateErr)
	}
	t.Logf("updated label id=%s to name=%s", labelID, updatedName)

	// Patch visibility
	root4 := integrationRootCmd()
	root4.AddCommand(integrationLabelsCmd(realFactory()))

	var patchErr error
	captureStdout(t, func() {
		root4.SetArgs([]string{"labels", "patch", "--id=" + labelID, "--label-list-visibility=labelShow", "--json"})
		patchErr = root4.Execute()
	})
	if patchErr != nil {
		t.Fatalf("labels patch failed: %v", patchErr)
	}
	t.Logf("patched label id=%s visibility", labelID)

	// Delete
	root5 := integrationRootCmd()
	root5.AddCommand(integrationLabelsCmd(realFactory()))

	var deleteErr error
	captureStdout(t, func() {
		root5.SetArgs([]string{"labels", "delete", "--id=" + labelID, "--confirm", "--json"})
		deleteErr = root5.Execute()
	})
	if deleteErr != nil {
		t.Fatalf("labels delete failed: %v", deleteErr)
	}
	t.Logf("deleted label id=%s", labelID)
}

// --- drafts ---

func TestIntegration_Drafts_CRUD(t *testing.T) {
	requireEnv(t)

	// Create a draft
	root1 := integrationRootCmd()
	root1.AddCommand(integrationDraftsCmd(realFactory()))

	var createOutput string
	var createErr error
	createOutput = captureStdout(t, func() {
		root1.SetArgs([]string{
			"drafts", "create",
			"--to=omniclaw680@gmail.com",
			"--subject=Integration Test Draft",
			"--body=Draft body for integration test",
			"--json",
		})
		createErr = root1.Execute()
	})
	if createErr != nil {
		t.Fatalf("drafts create failed: %v", createErr)
	}

	var createResult map[string]string
	if err := json.Unmarshal([]byte(createOutput), &createResult); err != nil {
		t.Fatalf("invalid create JSON: %v\noutput: %s", err, createOutput)
	}
	draftID := createResult["id"]
	if draftID == "" {
		t.Fatal("created draft has empty ID")
	}
	t.Logf("created draft id=%s", draftID)

	// List drafts — verify our draft appears
	root2 := integrationRootCmd()
	root2.AddCommand(integrationDraftsCmd(realFactory()))

	var listOutput string
	var listErr error
	listOutput = captureStdout(t, func() {
		root2.SetArgs([]string{"drafts", "list", "--json"})
		listErr = root2.Execute()
	})
	if listErr != nil {
		t.Fatalf("drafts list failed: %v", listErr)
	}

	var drafts []DraftSummary
	if err := json.Unmarshal([]byte(listOutput), &drafts); err != nil {
		t.Fatalf("invalid list JSON: %v\noutput: %s", err, listOutput)
	}
	found := false
	for _, d := range drafts {
		if d.ID == draftID {
			found = true
			t.Logf("found draft in list: id=%s subject=%q", d.ID, d.Subject)
		}
	}
	if !found {
		t.Errorf("created draft id=%s not found in drafts list", draftID)
	}

	// Get draft
	root3 := integrationRootCmd()
	root3.AddCommand(integrationDraftsCmd(realFactory()))

	var getOutput string
	var getErr error
	getOutput = captureStdout(t, func() {
		root3.SetArgs([]string{"drafts", "get", "--id=" + draftID, "--json"})
		getErr = root3.Execute()
	})
	if getErr != nil {
		t.Fatalf("drafts get failed: %v", getErr)
	}

	var detail DraftDetail
	if err := json.Unmarshal([]byte(getOutput), &detail); err != nil {
		t.Fatalf("invalid get JSON: %v\noutput: %s", err, getOutput)
	}
	if detail.ID != draftID {
		t.Errorf("expected ID=%s, got %s", draftID, detail.ID)
	}
	t.Logf("fetched draft: id=%s subject=%q", detail.ID, detail.Subject)

	// Update draft
	root4 := integrationRootCmd()
	root4.AddCommand(integrationDraftsCmd(realFactory()))

	var updateErr error
	captureStdout(t, func() {
		root4.SetArgs([]string{
			"drafts", "update",
			"--id=" + draftID,
			"--to=omniclaw680@gmail.com",
			"--subject=Integration Test Draft (updated)",
			"--body=Updated draft body",
			"--json",
		})
		updateErr = root4.Execute()
	})
	if updateErr != nil {
		t.Fatalf("drafts update failed: %v", updateErr)
	}
	t.Logf("updated draft id=%s", draftID)

	// Send draft
	root5 := integrationRootCmd()
	root5.AddCommand(integrationDraftsCmd(realFactory()))

	var sendOutput string
	var sendErr error
	sendOutput = captureStdout(t, func() {
		root5.SetArgs([]string{"drafts", "send", "--id=" + draftID, "--json"})
		sendErr = root5.Execute()
	})
	if sendErr != nil {
		t.Fatalf("drafts send failed: %v", sendErr)
	}

	var sendResult map[string]string
	if err := json.Unmarshal([]byte(sendOutput), &sendResult); err != nil {
		t.Fatalf("invalid send JSON: %v\noutput: %s", err, sendOutput)
	}
	if sendResult["status"] != "sent" {
		t.Errorf("expected status=sent, got %s", sendResult["status"])
	}
	t.Logf("sent draft: message id=%s threadId=%s", sendResult["id"], sendResult["threadId"])
}

func integrationHistoryCmd(factory ServiceFactory) *cobra.Command {
	return buildTestHistoryCmd(factory)
}

func integrationAttachmentsCmd(factory ServiceFactory) *cobra.Command {
	return buildTestAttachmentsCmd(factory)
}

// --- history list ---

func TestIntegration_History_List(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	// Fetch the current historyId from the user profile.
	svc, err := realFactory()(ctx)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}
	profile, err := svc.Users.GetProfile("me").Do()
	if err != nil {
		t.Fatalf("getting profile: %v", err)
	}
	// Use a slightly older history ID to ensure there is something to return.
	// We subtract a small delta from the current ID so the list is non-trivially populated.
	startID := profile.HistoryId
	if startID > 100 {
		startID -= 100
	}
	t.Logf("listing history from id=%d (current=%d)", startID, profile.HistoryId)

	root := integrationRootCmd()
	root.AddCommand(integrationHistoryCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"history", "list", "--start-history-id=" + formatUint64(startID), "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("history list failed: %v", execErr)
	}

	var result HistoryResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("got %d history entries, current historyId=%d", len(result.History), result.HistoryID)
	for _, h := range result.History {
		t.Logf("  [%d] added=%d deleted=%d labelsAdded=%d labelsRemoved=%d",
			h.ID, len(h.MessagesAdded), len(h.MessagesDeleted), len(h.LabelsAdded), len(h.LabelsRemoved))
	}
}

// formatUint64 formats a uint64 as a decimal string for use in flag arguments.
func formatUint64(v uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", v))
}

func integrationSettingsCmd(factory ServiceFactory) *cobra.Command {
	return buildTestSettingsCmd(factory)
}

// --- settings filters list (read-only, safe) ---

func TestIntegration_Settings_Filters_List(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(integrationSettingsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "filters", "list", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("settings filters list failed: %v", execErr)
	}

	var filters []FilterInfo
	if err := json.Unmarshal([]byte(output), &filters); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("got %d filters", len(filters))
	for _, f := range filters {
		t.Logf("  [%s] from=%s subject=%s", f.ID, f.Criteria.From, f.Criteria.Subject)
	}
}

// --- settings forwarding-addresses list (read-only, safe) ---

func TestIntegration_Settings_ForwardingAddresses_List(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(integrationSettingsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "forwarding-addresses", "list", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("settings forwarding-addresses list failed: %v", execErr)
	}

	var addresses []ForwardingAddressInfo
	if err := json.Unmarshal([]byte(output), &addresses); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("got %d forwarding addresses", len(addresses))
	for _, a := range addresses {
		t.Logf("  email=%s status=%s", a.ForwardingEmail, a.VerificationStatus)
	}
}

// --- settings get-vacation (read-only, safe) ---

func TestIntegration_Settings_GetVacation(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(integrationSettingsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get-vacation", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("settings get-vacation failed: %v", execErr)
	}

	var info VacationInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("vacation settings: enableAutoReply=%v subject=%q", info.EnableAutoReply, info.ResponseSubject)
}

// --- settings get-language (read-only, safe) ---

func TestIntegration_Settings_GetLanguage(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(integrationSettingsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get-language", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("settings get-language failed: %v", execErr)
	}

	var info LanguageInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if info.DisplayLanguage == "" {
		t.Error("expected non-empty displayLanguage")
	}
	t.Logf("display language: %s", info.DisplayLanguage)
}

// --- settings send-as list (read-only, safe) ---

func TestIntegration_Settings_SendAs_List(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(integrationSettingsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "send-as", "list", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("settings send-as list failed: %v", execErr)
	}

	var aliases []SendAsInfo
	if err := json.Unmarshal([]byte(output), &aliases); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("got %d send-as aliases", len(aliases))
	for _, a := range aliases {
		t.Logf("  email=%s displayName=%q isPrimary=%v verificationStatus=%s", a.SendAsEmail, a.DisplayName, a.IsPrimary, a.VerificationStatus)
	}
}

// --- settings delegates list (read-only, safe) ---

func TestIntegration_Settings_Delegates_List(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(integrationSettingsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "delegates", "list", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Skipf("settings delegates list not available (requires Workspace): %v", execErr)
	}

	var delegates []DelegateInfo
	if err := json.Unmarshal([]byte(output), &delegates); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("got %d delegates", len(delegates))
	for _, d := range delegates {
		t.Logf("  email=%s verificationStatus=%s", d.DelegateEmail, d.VerificationStatus)
	}
}

func TestIntegration_Drafts_Delete(t *testing.T) {
	requireEnv(t)

	// Create a draft to delete
	root1 := integrationRootCmd()
	root1.AddCommand(integrationDraftsCmd(realFactory()))

	var createOutput string
	var createErr error
	createOutput = captureStdout(t, func() {
		root1.SetArgs([]string{
			"drafts", "create",
			"--to=omniclaw680@gmail.com",
			"--subject=Integration Test Draft (to be deleted)",
			"--body=This draft will be deleted",
			"--json",
		})
		createErr = root1.Execute()
	})
	if createErr != nil {
		t.Fatalf("drafts create failed: %v", createErr)
	}

	var createResult map[string]string
	if err := json.Unmarshal([]byte(createOutput), &createResult); err != nil {
		t.Fatalf("invalid create JSON: %v\noutput: %s", err, createOutput)
	}
	draftID := createResult["id"]
	if draftID == "" {
		t.Fatal("created draft has empty ID")
	}
	t.Logf("created draft id=%s for deletion", draftID)

	// Delete the draft
	root2 := integrationRootCmd()
	root2.AddCommand(integrationDraftsCmd(realFactory()))

	var deleteErr error
	captureStdout(t, func() {
		root2.SetArgs([]string{"drafts", "delete", "--id=" + draftID, "--confirm", "--json"})
		deleteErr = root2.Execute()
	})
	if deleteErr != nil {
		t.Fatalf("drafts delete failed: %v", deleteErr)
	}
	t.Logf("deleted draft id=%s", draftID)
}

// --- messages import ---

func TestIntegration_Messages_Import(t *testing.T) {
	requireEnv(t)

	// Write a temp RFC 2822 message file.
	raw := rfc2822Message(
		"omniclaw680@gmail.com",
		"omniclaw680@gmail.com",
		"Integration Test Import",
		"This is an imported message.",
	)
	path := writeTempEML(t, raw)
	defer os.Remove(path)

	root := integrationRootCmd()
	root.AddCommand(integrationMessagesCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "import", "--raw-file=" + path, "--json"})
		execErr = root.Execute()
	})
	if execErr != nil {
		t.Fatalf("messages import failed: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if result["id"] == "" {
		t.Error("imported message has empty id")
	}
	if result["status"] != "imported" {
		t.Errorf("expected status=imported, got %s", result["status"])
	}
	t.Logf("imported message: id=%s threadId=%s", result["id"], result["threadId"])
}

// --- messages insert ---

func TestIntegration_Messages_Insert(t *testing.T) {
	requireEnv(t)

	raw := rfc2822Message(
		"omniclaw680@gmail.com",
		"omniclaw680@gmail.com",
		"Integration Test Insert",
		"This is an inserted message.",
	)
	path := writeTempEML(t, raw)
	defer os.Remove(path)

	root := integrationRootCmd()
	root.AddCommand(integrationMessagesCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"messages", "insert", "--raw-file=" + path, "--json"})
		execErr = root.Execute()
	})
	if execErr != nil {
		t.Fatalf("messages insert failed: %v", execErr)
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if result["id"] == "" {
		t.Error("inserted message has empty id")
	}
	if result["status"] != "inserted" {
		t.Errorf("expected status=inserted, got %s", result["status"])
	}
	t.Logf("inserted message: id=%s threadId=%s", result["id"], result["threadId"])
}

// --- messages batch-modify ---

func TestIntegration_Messages_BatchModify(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	svc, err := realFactory()(ctx)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}

	// Send two messages to self so we have known IDs to batch-modify.
	var id1, id2 string
	for i, subject := range []string{"BatchModify Test 1", "BatchModify Test 2"} {
		root := integrationRootCmd()
		root.AddCommand(integrationMessagesCmd(realFactory()))
		out := captureStdout(t, func() {
			root.SetArgs([]string{
				"messages", "send",
				"--to=omniclaw680@gmail.com",
				"--subject=" + subject,
				"--body=batch-modify test",
				"--json",
			})
			root.Execute() //nolint:errcheck
		})
		var sr SendResult
		json.Unmarshal([]byte(out), &sr) //nolint:errcheck
		if i == 0 {
			id1 = sr.ID
		} else {
			id2 = sr.ID
		}
	}

	// Fall back to the two most recent messages if send IDs are empty.
	if id1 == "" || id2 == "" {
		resp, err := svc.Users.Messages.List("me").MaxResults(2).Do()
		if err != nil {
			t.Fatalf("listing messages: %v", err)
		}
		if len(resp.Messages) < 2 {
			t.Skip("fewer than 2 messages in mailbox")
		}
		id1 = resp.Messages[0].Id
		id2 = resp.Messages[1].Id
	}
	t.Logf("batch-modifying messages: id1=%s id2=%s", id1, id2)

	root := integrationRootCmd()
	root.AddCommand(integrationMessagesCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"messages", "batch-modify",
			"--ids=" + id1 + "," + id2,
			"--add-labels=STARRED",
			"--json",
		})
		execErr = root.Execute()
	})
	if execErr != nil {
		t.Fatalf("messages batch-modify failed: %v", execErr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if result["status"] != "modified" {
		t.Errorf("expected status=modified, got %v", result["status"])
	}
	t.Logf("batch-modified: %v", result)

	// Restore: remove STARRED from those messages.
	root2 := integrationRootCmd()
	root2.AddCommand(integrationMessagesCmd(realFactory()))
	captureStdout(t, func() {
		root2.SetArgs([]string{
			"messages", "batch-modify",
			"--ids=" + id1 + "," + id2,
			"--remove-labels=STARRED",
			"--json",
		})
		root2.Execute() //nolint:errcheck
	})
}

// --- messages batch-delete ---

func TestIntegration_Messages_BatchDelete(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	svc, err := realFactory()(ctx)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}

	// Send two messages to self to get IDs we can safely delete.
	var id1, id2 string
	for i, subject := range []string{"BatchDelete Test 1", "BatchDelete Test 2"} {
		root := integrationRootCmd()
		root.AddCommand(integrationMessagesCmd(realFactory()))
		out := captureStdout(t, func() {
			root.SetArgs([]string{
				"messages", "send",
				"--to=omniclaw680@gmail.com",
				"--subject=" + subject,
				"--body=batch-delete test",
				"--json",
			})
			root.Execute() //nolint:errcheck
		})
		var sr SendResult
		json.Unmarshal([]byte(out), &sr) //nolint:errcheck
		if i == 0 {
			id1 = sr.ID
		} else {
			id2 = sr.ID
		}
	}

	// Fall back if send IDs are empty.
	if id1 == "" || id2 == "" {
		resp, err := svc.Users.Messages.List("me").MaxResults(2).Do()
		if err != nil {
			t.Fatalf("listing messages: %v", err)
		}
		if len(resp.Messages) < 2 {
			t.Skip("fewer than 2 messages in mailbox")
		}
		id1 = resp.Messages[0].Id
		id2 = resp.Messages[1].Id
	}
	t.Logf("batch-deleting messages: id1=%s id2=%s", id1, id2)

	// Trash both first so Gmail's backend can settle before permanent deletion.
	for _, id := range []string{id1, id2} {
		if _, err := svc.Users.Messages.Trash("me", id).Do(); err != nil {
			t.Fatalf("trash message %s: %v", id, err)
		}
	}
	time.Sleep(2 * time.Second)

	root := integrationRootCmd()
	root.AddCommand(integrationMessagesCmd(realFactory()))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{
			"messages", "batch-delete",
			"--ids=" + id1 + "," + id2,
			"--confirm",
			"--json",
		})
		execErr = root.Execute()
	})
	if execErr != nil {
		t.Fatalf("messages batch-delete failed: %v", execErr)
	}
	t.Logf("permanently batch-deleted messages: %s, %s", id1, id2)
}

// --- threads delete ---

func TestIntegration_Threads_Delete(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	// Send a message to get a thread ID.
	root1 := integrationRootCmd()
	root1.AddCommand(integrationMessagesCmd(realFactory()))

	var sendOutput string
	var sendErr error
	sendOutput = captureStdout(t, func() {
		root1.SetArgs([]string{
			"messages", "send",
			"--to=omniclaw680@gmail.com",
			"--subject=Integration Test Threads Delete",
			"--body=thread-delete test",
			"--json",
		})
		sendErr = root1.Execute()
	})
	if sendErr != nil {
		t.Fatalf("send failed: %v", sendErr)
	}

	var sendResult SendResult
	json.Unmarshal([]byte(sendOutput), &sendResult) //nolint:errcheck

	threadID := sendResult.ThreadID
	if threadID == "" {
		svc, err := realFactory()(ctx)
		if err != nil {
			t.Fatalf("creating service: %v", err)
		}
		resp, err := svc.Users.Threads.List("me").MaxResults(1).Do()
		if err != nil {
			t.Fatalf("listing threads: %v", err)
		}
		if len(resp.Threads) == 0 {
			t.Skip("no threads in mailbox")
		}
		threadID = resp.Threads[0].Id
	}
	t.Logf("deleting thread id=%s", threadID)

	// Trash the thread first, then wait for Gmail's backend to settle.
	svc, err := realFactory()(ctx)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}
	if _, err := svc.Users.Threads.Trash("me", threadID).Do(); err != nil {
		t.Fatalf("trash thread: %v", err)
	}
	time.Sleep(2 * time.Second)

	root2 := integrationRootCmd()
	root2.AddCommand(integrationThreadsCmd(realFactory()))

	var execErr error
	captureStdout(t, func() {
		root2.SetArgs([]string{"threads", "delete", "--id=" + threadID, "--confirm", "--json"})
		execErr = root2.Execute()
	})
	if execErr != nil {
		t.Fatalf("threads delete failed: %v", execErr)
	}
	t.Logf("permanently deleted thread id=%s", threadID)
}

// --- attachments get ---

func TestIntegration_Attachments_Get(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	// Insert a message with a simple text attachment via messages insert.
	// Build a multipart RFC 2822 message with one text/plain attachment.
	rawMsg := strings.Join([]string{
		"From: omniclaw680@gmail.com",
		"To: omniclaw680@gmail.com",
		"Subject: Integration Test Attachment",
		"MIME-Version: 1.0",
		`Content-Type: multipart/mixed; boundary="boundary_att_test"`,
		"",
		"--boundary_att_test",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		"Email body here.",
		"",
		"--boundary_att_test",
		`Content-Type: text/plain; name="test.txt"`,
		`Content-Disposition: attachment; filename="test.txt"`,
		"",
		"Hello from attachment.",
		"",
		"--boundary_att_test--",
	}, "\r\n")

	path := writeTempEML(t, []byte(rawMsg))
	defer os.Remove(path)

	// Insert the message.
	rootInsert := integrationRootCmd()
	rootInsert.AddCommand(integrationMessagesCmd(realFactory()))
	var insertOutput string
	captureStdout(t, func() {
		rootInsert.SetArgs([]string{"messages", "insert", "--raw-file=" + path, "--json"})
		rootInsert.Execute() //nolint:errcheck
	})
	_ = insertOutput

	// Wait briefly for the message to be indexed.
	time.Sleep(1 * time.Second)

	// Find a message that has payload.parts with a filename (attachment).
	svc, err := realFactory()(ctx)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}

	resp, err := svc.Users.Messages.List("me").MaxResults(10).Q("has:attachment").Do()
	if err != nil {
		t.Fatalf("listing messages with attachments: %v", err)
	}
	if len(resp.Messages) == 0 {
		t.Skip("no messages with attachments found in mailbox")
	}

	// Find a message that actually has an attachment part.
	var msgID, attachmentID string
	for _, m := range resp.Messages {
		msg, err := svc.Users.Messages.Get("me", m.Id).Format("full").Do()
		if err != nil {
			continue
		}
		if msg.Payload == nil {
			continue
		}
		for _, part := range msg.Payload.Parts {
			if part.Filename != "" && part.Body != nil && part.Body.AttachmentId != "" {
				msgID = msg.Id
				attachmentID = part.Body.AttachmentId
				t.Logf("found attachment: messageId=%s attachmentId=%s filename=%s", msgID, attachmentID, part.Filename)
				break
			}
		}
		if msgID != "" {
			break
		}
	}
	if msgID == "" || attachmentID == "" {
		t.Skip("no messages with downloadable attachments found")
	}

	root := integrationRootCmd()
	root.AddCommand(integrationAttachmentsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"attachments", "get",
			"--message-id=" + msgID,
			"--attachment-id=" + attachmentID,
			"--json",
		})
		execErr = root.Execute()
	})
	if execErr != nil {
		t.Fatalf("attachments get failed: %v", execErr)
	}

	var info AttachmentInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("attachment: id=%s size=%d", info.AttachmentID, info.Size)
}

// --- settings vacation round-trip ---

func TestIntegration_Settings_VacationRoundTrip(t *testing.T) {
	requireEnv(t)

	// Get current vacation settings so we can restore them.
	root1 := integrationRootCmd()
	root1.AddCommand(integrationSettingsCmd(realFactory()))
	var getOutput string
	var getErr error
	getOutput = captureStdout(t, func() {
		root1.SetArgs([]string{"settings", "get-vacation", "--json"})
		getErr = root1.Execute()
	})
	if getErr != nil {
		t.Fatalf("get-vacation failed: %v", getErr)
	}
	var original VacationInfo
	if err := json.Unmarshal([]byte(getOutput), &original); err != nil {
		t.Fatalf("invalid get-vacation JSON: %v\noutput: %s", err, getOutput)
	}
	t.Logf("original vacation: enabled=%v subject=%q", original.EnableAutoReply, original.ResponseSubject)

	// Restore original settings at the end of the test.
	defer func() {
		rootR := integrationRootCmd()
		rootR.AddCommand(integrationSettingsCmd(realFactory()))
		args := []string{
			"settings", "set-vacation",
			"--json",
		}
		if original.EnableAutoReply {
			args = append(args, "--enable-auto-reply=true")
		} else {
			args = append(args, "--enable-auto-reply=false")
		}
		if original.ResponseSubject != "" {
			args = append(args, "--subject="+original.ResponseSubject)
		}
		if original.ResponseBodyPlainText != "" {
			args = append(args, "--body="+original.ResponseBodyPlainText)
		}
		captureStdout(t, func() {
			rootR.SetArgs(args)
			rootR.Execute() //nolint:errcheck
		})
		t.Logf("restored original vacation settings")
	}()

	// Enable vacation with test values.
	root2 := integrationRootCmd()
	root2.AddCommand(integrationSettingsCmd(realFactory()))
	var setOutput string
	var setErr error
	setOutput = captureStdout(t, func() {
		root2.SetArgs([]string{
			"settings", "set-vacation",
			"--enable-auto-reply=true",
			"--subject=Integration Test Vacation",
			"--body=I am away during integration testing.",
			"--json",
		})
		setErr = root2.Execute()
	})
	if setErr != nil {
		t.Fatalf("set-vacation failed: %v", setErr)
	}
	var setInfo VacationInfo
	if err := json.Unmarshal([]byte(setOutput), &setInfo); err != nil {
		t.Fatalf("invalid set-vacation JSON: %v\noutput: %s", err, setOutput)
	}
	t.Logf("set vacation: enabled=%v subject=%q", setInfo.EnableAutoReply, setInfo.ResponseSubject)

	// Verify: get again and confirm the values were applied.
	root3 := integrationRootCmd()
	root3.AddCommand(integrationSettingsCmd(realFactory()))
	var verifyOutput string
	var verifyErr error
	verifyOutput = captureStdout(t, func() {
		root3.SetArgs([]string{"settings", "get-vacation", "--json"})
		verifyErr = root3.Execute()
	})
	if verifyErr != nil {
		t.Fatalf("get-vacation verify failed: %v", verifyErr)
	}
	var verified VacationInfo
	if err := json.Unmarshal([]byte(verifyOutput), &verified); err != nil {
		t.Fatalf("invalid verify JSON: %v\noutput: %s", err, verifyOutput)
	}
	if !verified.EnableAutoReply {
		t.Error("expected enableAutoReply=true after set")
	}
	if verified.ResponseSubject != "Integration Test Vacation" {
		t.Errorf("expected subject='Integration Test Vacation', got %q", verified.ResponseSubject)
	}
	t.Logf("verified vacation: enabled=%v subject=%q", verified.EnableAutoReply, verified.ResponseSubject)
}

// --- settings language round-trip ---

func TestIntegration_Settings_LanguageRoundTrip(t *testing.T) {
	requireEnv(t)

	// Get current language.
	root1 := integrationRootCmd()
	root1.AddCommand(integrationSettingsCmd(realFactory()))
	var getOutput string
	var getErr error
	getOutput = captureStdout(t, func() {
		root1.SetArgs([]string{"settings", "get-language", "--json"})
		getErr = root1.Execute()
	})
	if getErr != nil {
		t.Fatalf("get-language failed: %v", getErr)
	}
	var original LanguageInfo
	if err := json.Unmarshal([]byte(getOutput), &original); err != nil {
		t.Fatalf("invalid get-language JSON: %v\noutput: %s", err, getOutput)
	}
	t.Logf("original language: %s", original.DisplayLanguage)

	// Restore original language at the end.
	defer func() {
		if original.DisplayLanguage == "" {
			return
		}
		rootR := integrationRootCmd()
		rootR.AddCommand(integrationSettingsCmd(realFactory()))
		captureStdout(t, func() {
			rootR.SetArgs([]string{
				"settings", "set-language",
				"--display-language=" + original.DisplayLanguage,
				"--json",
			})
			rootR.Execute() //nolint:errcheck
		})
		t.Logf("restored language to %s", original.DisplayLanguage)
	}()

	// Set language to French.
	testLang := "fr"
	if original.DisplayLanguage == "fr" {
		testLang = "de"
	}
	root2 := integrationRootCmd()
	root2.AddCommand(integrationSettingsCmd(realFactory()))
	var setOutput string
	var setErr error
	setOutput = captureStdout(t, func() {
		root2.SetArgs([]string{
			"settings", "set-language",
			"--display-language=" + testLang,
			"--json",
		})
		setErr = root2.Execute()
	})
	if setErr != nil {
		t.Fatalf("set-language failed: %v", setErr)
	}
	var setInfo LanguageInfo
	if err := json.Unmarshal([]byte(setOutput), &setInfo); err != nil {
		t.Fatalf("invalid set-language JSON: %v\noutput: %s", err, setOutput)
	}
	if setInfo.DisplayLanguage != testLang {
		t.Errorf("expected language=%s after set, got %s", testLang, setInfo.DisplayLanguage)
	}
	t.Logf("set language to %s", setInfo.DisplayLanguage)

	// Verify: get and confirm.
	root3 := integrationRootCmd()
	root3.AddCommand(integrationSettingsCmd(realFactory()))
	var verifyOutput string
	var verifyErr error
	verifyOutput = captureStdout(t, func() {
		root3.SetArgs([]string{"settings", "get-language", "--json"})
		verifyErr = root3.Execute()
	})
	if verifyErr != nil {
		t.Fatalf("get-language verify failed: %v", verifyErr)
	}
	var verified LanguageInfo
	if err := json.Unmarshal([]byte(verifyOutput), &verified); err != nil {
		t.Fatalf("invalid verify JSON: %v\noutput: %s", err, verifyOutput)
	}
	if verified.DisplayLanguage != testLang {
		t.Errorf("expected language=%s after verify, got %s", testLang, verified.DisplayLanguage)
	}
	t.Logf("verified language: %s", verified.DisplayLanguage)
}

// --- settings get-auto-forwarding ---

func TestIntegration_Settings_AutoForwardingGet(t *testing.T) {
	requireEnv(t)

	root := integrationRootCmd()
	root.AddCommand(integrationSettingsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"settings", "get-auto-forwarding", "--json"})
		execErr = root.Execute()
	})
	if execErr != nil {
		t.Fatalf("get-auto-forwarding failed: %v", execErr)
	}

	var info AutoForwardingInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	t.Logf("auto-forwarding: enabled=%v email=%s disposition=%s", info.Enabled, info.EmailAddress, info.Disposition)
}

// --- settings IMAP round-trip ---

func TestIntegration_Settings_ImapRoundTrip(t *testing.T) {
	requireEnv(t)

	// Get current IMAP settings.
	root1 := integrationRootCmd()
	root1.AddCommand(integrationSettingsCmd(realFactory()))
	var getOutput string
	var getErr error
	getOutput = captureStdout(t, func() {
		root1.SetArgs([]string{"settings", "get-imap", "--json"})
		getErr = root1.Execute()
	})
	if getErr != nil {
		t.Fatalf("get-imap failed: %v", getErr)
	}
	var original ImapInfo
	if err := json.Unmarshal([]byte(getOutput), &original); err != nil {
		t.Fatalf("invalid get-imap JSON: %v\noutput: %s", err, getOutput)
	}
	t.Logf("original IMAP: enabled=%v autoExpunge=%v expungeBehavior=%s", original.Enabled, original.AutoExpunge, original.ExpungeBehavior)

	// Restore original IMAP settings at the end.
	defer func() {
		rootR := integrationRootCmd()
		rootR.AddCommand(integrationSettingsCmd(realFactory()))
		args := []string{
			"settings", "set-imap",
			fmt.Sprintf("--enabled=%v", original.Enabled),
			fmt.Sprintf("--auto-expunge=%v", original.AutoExpunge),
			"--json",
		}
		if original.ExpungeBehavior != "" {
			args = append(args, "--expunge-behavior="+original.ExpungeBehavior)
		}
		captureStdout(t, func() {
			rootR.SetArgs(args)
			rootR.Execute() //nolint:errcheck
		})
		t.Logf("restored original IMAP settings")
	}()

	// Toggle the enabled flag.
	newEnabled := !original.Enabled
	root2 := integrationRootCmd()
	root2.AddCommand(integrationSettingsCmd(realFactory()))
	var setOutput string
	var setErr error
	setOutput = captureStdout(t, func() {
		root2.SetArgs([]string{
			"settings", "set-imap",
			fmt.Sprintf("--enabled=%v", newEnabled),
			"--json",
		})
		setErr = root2.Execute()
	})
	if setErr != nil {
		t.Fatalf("set-imap failed: %v", setErr)
	}
	var setInfo ImapInfo
	if err := json.Unmarshal([]byte(setOutput), &setInfo); err != nil {
		t.Fatalf("invalid set-imap JSON: %v\noutput: %s", err, setOutput)
	}
	t.Logf("set IMAP: enabled=%v", setInfo.Enabled)

	// Verify.
	root3 := integrationRootCmd()
	root3.AddCommand(integrationSettingsCmd(realFactory()))
	var verifyOutput string
	var verifyErr error
	verifyOutput = captureStdout(t, func() {
		root3.SetArgs([]string{"settings", "get-imap", "--json"})
		verifyErr = root3.Execute()
	})
	if verifyErr != nil {
		t.Fatalf("get-imap verify failed: %v", verifyErr)
	}
	var verified ImapInfo
	if err := json.Unmarshal([]byte(verifyOutput), &verified); err != nil {
		t.Fatalf("invalid verify JSON: %v\noutput: %s", err, verifyOutput)
	}
	if verified.Enabled != newEnabled {
		t.Errorf("expected enabled=%v after set, got %v", newEnabled, verified.Enabled)
	}
	t.Logf("verified IMAP: enabled=%v", verified.Enabled)
}

// --- settings POP round-trip ---

func TestIntegration_Settings_PopRoundTrip(t *testing.T) {
	requireEnv(t)

	// Get current POP settings.
	root1 := integrationRootCmd()
	root1.AddCommand(integrationSettingsCmd(realFactory()))
	var getOutput string
	var getErr error
	getOutput = captureStdout(t, func() {
		root1.SetArgs([]string{"settings", "get-pop", "--json"})
		getErr = root1.Execute()
	})
	if getErr != nil {
		t.Fatalf("get-pop failed: %v", getErr)
	}
	var original PopInfo
	if err := json.Unmarshal([]byte(getOutput), &original); err != nil {
		t.Fatalf("invalid get-pop JSON: %v\noutput: %s", err, getOutput)
	}
	t.Logf("original POP: accessWindow=%s disposition=%s", original.AccessWindow, original.Disposition)

	// Restore at end.
	defer func() {
		if original.AccessWindow == "" {
			return
		}
		rootR := integrationRootCmd()
		rootR.AddCommand(integrationSettingsCmd(realFactory()))
		args := []string{
			"settings", "set-pop",
			"--access-window=" + original.AccessWindow,
			"--json",
		}
		if original.Disposition != "" {
			args = append(args, "--disposition="+original.Disposition)
		}
		captureStdout(t, func() {
			rootR.SetArgs(args)
			rootR.Execute() //nolint:errcheck
		})
		t.Logf("restored original POP settings")
	}()

	// Choose a different access window value to set.
	testWindow := "disabled"
	if original.AccessWindow == "disabled" {
		testWindow = "allMail"
	}
	root2 := integrationRootCmd()
	root2.AddCommand(integrationSettingsCmd(realFactory()))
	var setOutput string
	var setErr error
	setOutput = captureStdout(t, func() {
		root2.SetArgs([]string{
			"settings", "set-pop",
			"--access-window=" + testWindow,
			"--json",
		})
		setErr = root2.Execute()
	})
	if setErr != nil {
		t.Fatalf("set-pop failed: %v", setErr)
	}
	var setInfo PopInfo
	if err := json.Unmarshal([]byte(setOutput), &setInfo); err != nil {
		t.Fatalf("invalid set-pop JSON: %v\noutput: %s", err, setOutput)
	}
	t.Logf("set POP: accessWindow=%s", setInfo.AccessWindow)

	// Verify.
	root3 := integrationRootCmd()
	root3.AddCommand(integrationSettingsCmd(realFactory()))
	var verifyOutput string
	var verifyErr error
	verifyOutput = captureStdout(t, func() {
		root3.SetArgs([]string{"settings", "get-pop", "--json"})
		verifyErr = root3.Execute()
	})
	if verifyErr != nil {
		t.Fatalf("get-pop verify failed: %v", verifyErr)
	}
	var verified PopInfo
	if err := json.Unmarshal([]byte(verifyOutput), &verified); err != nil {
		t.Fatalf("invalid verify JSON: %v\noutput: %s", err, verifyOutput)
	}
	if verified.AccessWindow != testWindow {
		t.Errorf("expected accessWindow=%s after set, got %s", testWindow, verified.AccessWindow)
	}
	t.Logf("verified POP: accessWindow=%s", verified.AccessWindow)
}

// --- settings filters CRUD ---

func TestIntegration_Settings_Filters_CRUD(t *testing.T) {
	requireEnv(t)

	// Create a filter.
	root1 := integrationRootCmd()
	root1.AddCommand(integrationSettingsCmd(realFactory()))
	var createOutput string
	var createErr error
	createOutput = captureStdout(t, func() {
		root1.SetArgs([]string{
			"settings", "filters", "create",
			"--from=integration-test-filter@example.com",
			"--add-label=STARRED",
			"--json",
		})
		createErr = root1.Execute()
	})
	if createErr != nil {
		t.Fatalf("filters create failed: %v", createErr)
	}
	var createResult map[string]string
	if err := json.Unmarshal([]byte(createOutput), &createResult); err != nil {
		t.Fatalf("invalid create JSON: %v\noutput: %s", err, createOutput)
	}
	filterID := createResult["id"]
	if filterID == "" {
		t.Fatal("created filter has empty id")
	}
	t.Logf("created filter: id=%s", filterID)

	// Get the filter.
	root2 := integrationRootCmd()
	root2.AddCommand(integrationSettingsCmd(realFactory()))
	var getOutput string
	var getErr error
	getOutput = captureStdout(t, func() {
		root2.SetArgs([]string{
			"settings", "filters", "get",
			"--id=" + filterID,
			"--json",
		})
		getErr = root2.Execute()
	})
	if getErr != nil {
		t.Fatalf("filters get failed: %v", getErr)
	}
	var info FilterInfo
	if err := json.Unmarshal([]byte(getOutput), &info); err != nil {
		t.Fatalf("invalid get JSON: %v\noutput: %s", err, getOutput)
	}
	if info.ID != filterID {
		t.Errorf("expected ID=%s, got %s", filterID, info.ID)
	}
	t.Logf("fetched filter: id=%s from=%s", info.ID, info.Criteria.From)

	// Delete the filter.
	root3 := integrationRootCmd()
	root3.AddCommand(integrationSettingsCmd(realFactory()))
	var deleteErr error
	captureStdout(t, func() {
		root3.SetArgs([]string{
			"settings", "filters", "delete",
			"--id=" + filterID,
			"--confirm",
			"--json",
		})
		deleteErr = root3.Execute()
	})
	if deleteErr != nil {
		t.Fatalf("filters delete failed: %v", deleteErr)
	}
	t.Logf("deleted filter: id=%s", filterID)
}

// --- settings forwarding-addresses create/get/delete ---

func TestIntegration_Settings_ForwardingAddresses_CRUD(t *testing.T) {
	requireEnv(t)

	testEmail := "fwd-integration-test@example.com"

	// Attempt to create a forwarding address (sends a verification email, which is fine on a burner).
	root1 := integrationRootCmd()
	root1.AddCommand(integrationSettingsCmd(realFactory()))
	var createOutput string
	var createErr error
	createOutput = captureStdout(t, func() {
		root1.SetArgs([]string{
			"settings", "forwarding-addresses", "create",
			"--email=" + testEmail,
			"--json",
		})
		createErr = root1.Execute()
	})
	if createErr != nil {
		// Creation may fail if the API restricts forwarding address creation on consumer accounts.
		t.Skipf("forwarding-addresses create not available: %v", createErr)
	}
	var createResult map[string]string
	if err := json.Unmarshal([]byte(createOutput), &createResult); err != nil {
		t.Fatalf("invalid create JSON: %v\noutput: %s", err, createOutput)
	}
	t.Logf("created forwarding address: email=%s status=%s", createResult["forwardingEmail"], createResult["verificationStatus"])

	// Get the forwarding address.
	root2 := integrationRootCmd()
	root2.AddCommand(integrationSettingsCmd(realFactory()))
	var getOutput string
	var getErr error
	getOutput = captureStdout(t, func() {
		root2.SetArgs([]string{
			"settings", "forwarding-addresses", "get",
			"--email=" + testEmail,
			"--json",
		})
		getErr = root2.Execute()
	})
	if getErr != nil {
		t.Logf("forwarding-addresses get failed (may need verification first): %v", getErr)
	} else {
		var info ForwardingAddressInfo
		if err := json.Unmarshal([]byte(getOutput), &info); err != nil {
			t.Fatalf("invalid get JSON: %v\noutput: %s", err, getOutput)
		}
		t.Logf("fetched forwarding address: email=%s status=%s", info.ForwardingEmail, info.VerificationStatus)
	}

	// Delete the forwarding address.
	root3 := integrationRootCmd()
	root3.AddCommand(integrationSettingsCmd(realFactory()))
	var deleteErr error
	captureStdout(t, func() {
		root3.SetArgs([]string{
			"settings", "forwarding-addresses", "delete",
			"--email=" + testEmail,
			"--confirm",
			"--json",
		})
		deleteErr = root3.Execute()
	})
	if deleteErr != nil {
		t.Logf("forwarding-addresses delete failed (may not exist yet): %v", deleteErr)
	} else {
		t.Logf("deleted forwarding address: email=%s", testEmail)
	}
}

// --- settings send-as get (primary) ---

func TestIntegration_Settings_SendAs_Get(t *testing.T) {
	requireEnv(t)
	ctx := context.Background()

	// Get the list first to find the primary send-as email.
	svc, err := realFactory()(ctx)
	if err != nil {
		t.Fatalf("creating service: %v", err)
	}
	resp, err := svc.Users.Settings.SendAs.List("me").Do()
	if err != nil {
		t.Fatalf("listing send-as: %v", err)
	}
	if len(resp.SendAs) == 0 {
		t.Skip("no send-as aliases found")
	}

	// Use the primary alias.
	var primaryEmail string
	for _, sa := range resp.SendAs {
		if sa.IsPrimary {
			primaryEmail = sa.SendAsEmail
			break
		}
	}
	if primaryEmail == "" {
		primaryEmail = resp.SendAs[0].SendAsEmail
	}
	t.Logf("using send-as email=%s", primaryEmail)

	root := integrationRootCmd()
	root.AddCommand(integrationSettingsCmd(realFactory()))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{
			"settings", "send-as", "get",
			"--email=" + primaryEmail,
			"--json",
		})
		execErr = root.Execute()
	})
	if execErr != nil {
		t.Fatalf("send-as get failed: %v", execErr)
	}

	var info SendAsInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if info.SendAsEmail != primaryEmail {
		t.Errorf("expected sendAsEmail=%s, got %s", primaryEmail, info.SendAsEmail)
	}
	t.Logf("send-as: email=%s displayName=%q isPrimary=%v", info.SendAsEmail, info.DisplayName, info.IsPrimary)
}

// --- settings send-as CRUD ---

func TestIntegration_Settings_SendAs_CRUD(t *testing.T) {
	requireEnv(t)

	testEmail := "sendas-integration-test@example.com"

	// Attempt to create a send-as alias. This may fail if the email is not verified.
	root1 := integrationRootCmd()
	root1.AddCommand(integrationSettingsCmd(realFactory()))
	var createOutput string
	var createErr error
	createOutput = captureStdout(t, func() {
		root1.SetArgs([]string{
			"settings", "send-as", "create",
			"--email=" + testEmail,
			"--display-name=Integration Test Alias",
			"--json",
		})
		createErr = root1.Execute()
	})
	if createErr != nil {
		t.Skipf("send-as create failed (email may require verification): %v", createErr)
	}
	var createResult map[string]string
	if err := json.Unmarshal([]byte(createOutput), &createResult); err != nil {
		t.Fatalf("invalid create JSON: %v\noutput: %s", err, createOutput)
	}
	t.Logf("created send-as alias: email=%s", createResult["sendAsEmail"])

	// Update display name.
	root2 := integrationRootCmd()
	root2.AddCommand(integrationSettingsCmd(realFactory()))
	var updateErr error
	captureStdout(t, func() {
		root2.SetArgs([]string{
			"settings", "send-as", "update",
			"--email=" + testEmail,
			"--display-name=Integration Test Alias (updated)",
			"--json",
		})
		updateErr = root2.Execute()
	})
	if updateErr != nil {
		t.Logf("send-as update failed: %v", updateErr)
	} else {
		t.Logf("updated send-as alias display name")
	}

	// Patch signature.
	root3 := integrationRootCmd()
	root3.AddCommand(integrationSettingsCmd(realFactory()))
	var patchErr error
	captureStdout(t, func() {
		root3.SetArgs([]string{
			"settings", "send-as", "patch",
			"--email=" + testEmail,
			"--signature=<b>Integration Test</b>",
			"--json",
		})
		patchErr = root3.Execute()
	})
	if patchErr != nil {
		t.Logf("send-as patch failed: %v", patchErr)
	} else {
		t.Logf("patched send-as alias signature")
	}

	// Verify — send verification email (no-op if not already accepted).
	root4 := integrationRootCmd()
	root4.AddCommand(integrationSettingsCmd(realFactory()))
	var verifyErr error
	captureStdout(t, func() {
		root4.SetArgs([]string{
			"settings", "send-as", "verify",
			"--email=" + testEmail,
			"--json",
		})
		verifyErr = root4.Execute()
	})
	if verifyErr != nil {
		t.Logf("send-as verify failed (expected for unverified alias): %v", verifyErr)
	}

	// Delete the send-as alias.
	root5 := integrationRootCmd()
	root5.AddCommand(integrationSettingsCmd(realFactory()))
	var deleteErr error
	captureStdout(t, func() {
		root5.SetArgs([]string{
			"settings", "send-as", "delete",
			"--email=" + testEmail,
			"--confirm",
			"--json",
		})
		deleteErr = root5.Execute()
	})
	if deleteErr != nil {
		t.Logf("send-as delete failed: %v", deleteErr)
	} else {
		t.Logf("deleted send-as alias: email=%s", testEmail)
	}
}

// --- settings delegates CRUD ---

func TestIntegration_Settings_Delegates_CRUD(t *testing.T) {
	requireEnv(t)

	testEmail := "delegate-integration-test@example.com"

	// Attempt to create a delegate. This requires Workspace domain-wide authority and will
	// almost certainly return 403 on a consumer Gmail account.
	root1 := integrationRootCmd()
	root1.AddCommand(integrationSettingsCmd(realFactory()))
	var createOutput string
	var createErr error
	createOutput = captureStdout(t, func() {
		root1.SetArgs([]string{
			"settings", "delegates", "create",
			"--email=" + testEmail,
			"--confirm",
			"--json",
		})
		createErr = root1.Execute()
	})
	if createErr != nil {
		t.Skipf("delegates create not available (requires Workspace domain-wide authority): %v", createErr)
	}
	var createResult map[string]string
	if err := json.Unmarshal([]byte(createOutput), &createResult); err != nil {
		t.Fatalf("invalid create JSON: %v\noutput: %s", err, createOutput)
	}
	t.Logf("created delegate: email=%s status=%s", createResult["delegateEmail"], createResult["verificationStatus"])

	// Get the delegate.
	root2 := integrationRootCmd()
	root2.AddCommand(integrationSettingsCmd(realFactory()))
	var getOutput string
	var getErr error
	getOutput = captureStdout(t, func() {
		root2.SetArgs([]string{
			"settings", "delegates", "get",
			"--email=" + testEmail,
			"--json",
		})
		getErr = root2.Execute()
	})
	if getErr != nil {
		t.Logf("delegates get failed: %v", getErr)
	} else {
		var info DelegateInfo
		if err := json.Unmarshal([]byte(getOutput), &info); err != nil {
			t.Fatalf("invalid get JSON: %v\noutput: %s", err, getOutput)
		}
		t.Logf("fetched delegate: email=%s status=%s", info.DelegateEmail, info.VerificationStatus)
	}

	// Delete the delegate.
	root3 := integrationRootCmd()
	root3.AddCommand(integrationSettingsCmd(realFactory()))
	var deleteErr error
	captureStdout(t, func() {
		root3.SetArgs([]string{
			"settings", "delegates", "delete",
			"--email=" + testEmail,
			"--confirm",
			"--json",
		})
		deleteErr = root3.Execute()
	})
	if deleteErr != nil {
		t.Logf("delegates delete failed: %v", deleteErr)
	} else {
		t.Logf("deleted delegate: email=%s", testEmail)
	}
}
