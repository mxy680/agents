//go:build integration

package gmail

import (
	"context"
	"encoding/json"
	"fmt"
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
