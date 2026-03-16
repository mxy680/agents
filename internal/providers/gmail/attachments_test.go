package gmail

import (
	"encoding/json"
	"os"
	"testing"
)

func TestAttachmentsGetJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestAttachmentsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"attachments", "get", "--message-id=msg1", "--attachment-id=att1", "--json"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("attachments get --json failed: %v", execErr)
	}

	var info AttachmentInfo
	if err := json.Unmarshal([]byte(output), &info); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if info.AttachmentID != "att1" {
		t.Errorf("expected attachmentId=att1, got %s", info.AttachmentID)
	}
	if info.Size != 1234 {
		t.Errorf("expected size=1234, got %d", info.Size)
	}
}

func TestAttachmentsGetToStdout(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestAttachmentsCmd(factory))

	var output string
	var execErr error
	output = captureStdout(t, func() {
		root.SetArgs([]string{"attachments", "get", "--message-id=msg1", "--attachment-id=att1"})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("attachments get (stdout) failed: %v", execErr)
	}

	// The mock returns base64 of "Hello World"
	if output != "Hello World" {
		t.Errorf("expected stdout to contain raw bytes %q, got %q", "Hello World", output)
	}
}

func TestAttachmentsGetToFile(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	factory := newTestServiceFactory(server)

	tmp, err := os.CreateTemp(t.TempDir(), "attachment-*.bin")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	tmp.Close()

	root := newTestRootCmd()
	root.AddCommand(buildTestAttachmentsCmd(factory))

	var execErr error
	captureStdout(t, func() {
		root.SetArgs([]string{"attachments", "get", "--message-id=msg1", "--attachment-id=att1", "--output=" + tmp.Name()})
		execErr = root.Execute()
	})

	if execErr != nil {
		t.Fatalf("attachments get --output failed: %v", execErr)
	}

	contents, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatalf("reading output file: %v", err)
	}
	if string(contents) != "Hello World" {
		t.Errorf("expected file contents %q, got %q", "Hello World", string(contents))
	}
}
