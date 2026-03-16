package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintJSON(t *testing.T) {
	data := map[string]string{"key": "value"}
	output := captureStdout(func() {
		if err := PrintJSON(data); err != nil {
			t.Fatal(err)
		}
	})

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("expected key=value, got key=%s", result["key"])
	}
}

func TestPrintJSON_Array(t *testing.T) {
	data := []int{1, 2, 3}
	output := captureStdout(func() {
		if err := PrintJSON(data); err != nil {
			t.Fatal(err)
		}
	})

	var result []int
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 items, got %d", len(result))
	}
}

func TestPrintText(t *testing.T) {
	output := captureStdout(func() {
		PrintText([]string{"line1", "line2", "line3"})
	})
	expected := "line1\nline2\nline3\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func newTestCmd(jsonFlag, dryRunFlag bool) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().Bool("json", jsonFlag, "")
	cmd.Flags().Bool("dry-run", dryRunFlag, "")
	return cmd
}

func TestPrintResult_JSON(t *testing.T) {
	cmd := newTestCmd(true, false)
	data := map[string]string{"status": "ok"}

	output := captureStdout(func() {
		if err := PrintResult(cmd, data, []string{"Status: ok"}); err != nil {
			t.Fatal(err)
		}
	})

	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", output)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status=ok, got %s", result["status"])
	}
}

func TestPrintResult_Text(t *testing.T) {
	cmd := newTestCmd(false, false)
	data := map[string]string{"status": "ok"}

	output := captureStdout(func() {
		if err := PrintResult(cmd, data, []string{"Status: ok"}); err != nil {
			t.Fatal(err)
		}
	})

	if output != "Status: ok\n" {
		t.Errorf("expected text output, got: %q", output)
	}
}

func TestIsJSONOutput(t *testing.T) {
	cmd := newTestCmd(true, false)
	if !IsJSONOutput(cmd) {
		t.Error("expected IsJSONOutput to return true")
	}

	cmd2 := newTestCmd(false, false)
	if IsJSONOutput(cmd2) {
		t.Error("expected IsJSONOutput to return false")
	}
}

func TestIsDryRun(t *testing.T) {
	cmd := newTestCmd(false, true)
	if !IsDryRun(cmd) {
		t.Error("expected IsDryRun to return true")
	}

	cmd2 := newTestCmd(false, false)
	if IsDryRun(cmd2) {
		t.Error("expected IsDryRun to return false")
	}
}
