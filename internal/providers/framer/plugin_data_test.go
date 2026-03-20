package framer

import (
	"strings"
	"testing"
)

func TestPluginDataGet(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	pluginDataCmd := newPluginDataCmd(factory)
	root.AddCommand(pluginDataCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"plugin-data", "get", "--key", "mykey"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "stored-value") {
		t.Errorf("expected 'stored-value' in output, got: %s", output)
	}
}

func TestPluginDataSet(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	pluginDataCmd := newPluginDataCmd(factory)
	root.AddCommand(pluginDataCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"plugin-data", "set", "--key", "mykey", "--value", "myvalue"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "mykey") {
		t.Errorf("expected 'mykey' in output, got: %s", output)
	}
}

func TestPluginDataSetDryRun(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	pluginDataCmd := newPluginDataCmd(factory)
	root.AddCommand(pluginDataCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"plugin-data", "set", "--key", "mykey", "--value", "myvalue", "--dry-run"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "dry") {
		t.Errorf("expected dry-run output, got: %s", output)
	}
}

func TestPluginDataKeys(t *testing.T) {
	factory := mockBridge(defaultHandler())
	root := newTestRootCmd()
	pluginDataCmd := newPluginDataCmd(factory)
	root.AddCommand(pluginDataCmd)

	output := captureStdout(t, func() {
		root.SetArgs([]string{"plugin-data", "keys"})
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(output, "key1") {
		t.Errorf("expected 'key1' in output, got: %s", output)
	}
	if !strings.Contains(output, "key2") {
		t.Errorf("expected 'key2' in output, got: %s", output)
	}
	if !strings.Contains(output, "key3") {
		t.Errorf("expected 'key3' in output, got: %s", output)
	}
}
