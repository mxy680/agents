package framer

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
)

// bridgeCommand is a JSON-RPC command sent to the Node.js bridge.
type bridgeCommand struct {
	Method string         `json:"method"`
	Params map[string]any `json:"params,omitempty"`
}

// bridgeResponse is a JSON-RPC response from the Node.js bridge.
type bridgeResponse struct {
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// BridgeClient manages a Node.js subprocess that communicates with the Framer API.
type BridgeClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	reader *bufio.Reader
	mu     sync.Mutex
}

// BridgeClientFactory is the function signature for creating a BridgeClient.
type BridgeClientFactory func(ctx context.Context) (*BridgeClient, error)

// DefaultBridgeClientFactory returns a factory that spawns the real Node.js bridge.
func DefaultBridgeClientFactory() BridgeClientFactory {
	return func(ctx context.Context) (*BridgeClient, error) {
		return NewBridgeClient(ctx)
	}
}

// bridgeDir returns the directory containing the Node.js bridge code.
// It looks relative to the current executable first, then falls back
// to the source-tree location for development.
func bridgeDir() string {
	// Try relative to executable (for installed binary)
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Join(filepath.Dir(exe), "..", "lib", "framer-bridge")
		if _, err := os.Stat(filepath.Join(dir, "bridge.js")); err == nil {
			return dir
		}
	}

	// Fall back to source tree location (development)
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "bridge")
}

// NewBridgeClient starts the Node.js bridge subprocess.
func NewBridgeClient(ctx context.Context) (*BridgeClient, error) {
	dir := bridgeDir()
	bridgeScript := filepath.Join(dir, "bridge.js")

	if _, err := os.Stat(bridgeScript); err != nil {
		return nil, fmt.Errorf("framer bridge not found at %s: %w", bridgeScript, err)
	}

	// Check node_modules exist
	nodeModules := filepath.Join(dir, "node_modules")
	if _, err := os.Stat(nodeModules); err != nil {
		return nil, fmt.Errorf("framer bridge dependencies not installed (run 'npm install' in %s): %w", dir, err)
	}

	cmd := exec.CommandContext(ctx, "node", bridgeScript)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr // Pass through bridge errors

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start framer bridge: %w", err)
	}

	return &BridgeClient{
		cmd:    cmd,
		stdin:  stdin,
		reader: bufio.NewReader(stdout),
	}, nil
}

// Call sends a command to the bridge and returns the result.
func (b *BridgeClient) Call(method string, params map[string]any) (json.RawMessage, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	cmd := bridgeCommand{Method: method, Params: params}
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("marshal command: %w", err)
	}

	data = append(data, '\n')
	if _, err := b.stdin.Write(data); err != nil {
		return nil, fmt.Errorf("write to bridge: %w", err)
	}

	line, err := b.reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("read from bridge: %w", err)
	}

	var resp bridgeResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("framer API: %s", resp.Error)
	}

	return resp.Result, nil
}

// Close disconnects from Framer and stops the bridge subprocess.
func (b *BridgeClient) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Send disconnect command
	cmd := bridgeCommand{Method: "disconnect"}
	data, _ := json.Marshal(cmd)
	data = append(data, '\n')
	b.stdin.Write(data) //nolint:errcheck

	b.stdin.Close()

	// cmd is nil when using a mock client (no subprocess to wait on).
	if b.cmd == nil {
		return nil
	}
	return b.cmd.Wait()
}
