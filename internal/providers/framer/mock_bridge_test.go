package framer

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

// mockBridgeHandler is a function that handles a bridge command and returns a response.
type mockBridgeHandler func(method string, params map[string]any) (any, error)

// mockBridge creates a BridgeClientFactory that uses an in-process handler
// instead of spawning a Node.js subprocess.
func mockBridge(handler mockBridgeHandler) BridgeClientFactory {
	return func(ctx context.Context) (*BridgeClient, error) {
		// Create pipes for stdin/stdout
		stdinReader, stdinWriter := io.Pipe()
		stdoutReader, stdoutWriter := io.Pipe()

		// Start a goroutine that reads commands and writes responses
		go func() {
			defer stdoutWriter.Close()
			scanner := bufio.NewScanner(stdinReader)
			for scanner.Scan() {
				line := scanner.Text()
				var cmd bridgeCommand
				if err := json.Unmarshal([]byte(line), &cmd); err != nil {
					resp, _ := json.Marshal(bridgeResponse{Error: "invalid JSON"})
					stdoutWriter.Write(append(resp, '\n')) //nolint:errcheck
					continue
				}

				if cmd.Method == "disconnect" {
					resp, _ := json.Marshal(bridgeResponse{Result: json.RawMessage(`{"success":true}`)})
					stdoutWriter.Write(append(resp, '\n')) //nolint:errcheck
					continue
				}

				result, err := handler(cmd.Method, cmd.Params)
				if err != nil {
					resp, _ := json.Marshal(bridgeResponse{Error: err.Error()})
					stdoutWriter.Write(append(resp, '\n')) //nolint:errcheck
					continue
				}

				resultJSON, _ := json.Marshal(result)
				resp, _ := json.Marshal(bridgeResponse{Result: resultJSON})
				stdoutWriter.Write(append(resp, '\n')) //nolint:errcheck
			}
		}()

		return &BridgeClient{
			stdin:  stdinWriter,
			reader: bufio.NewReader(stdoutReader),
		}, nil
	}
}

// captureStdout captures stdout output during execution of f.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r) //nolint:errcheck
	return buf.String()
}

// newTestRootCmd creates a fresh root command for testing (avoids flag conflicts).
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "Output in JSON format")
	root.PersistentFlags().Bool("dry-run", false, "Preview actions without executing them")
	return root
}

// defaultHandler returns a mock handler that responds to all common methods.
func defaultHandler() mockBridgeHandler {
	return func(method string, params map[string]any) (any, error) {
		switch method {
		// Project
		case "getProjectInfo":
			return map[string]any{"name": "Test Project", "id": "proj-1", "versioned_id": "v-proj-1"}, nil
		case "getCurrentUser":
			return map[string]any{"id": "user-1", "name": "Test User"}, nil
		case "getChangedPaths":
			return map[string]any{"added": []string{"/new-page"}, "removed": []string{}, "modified": []string{"/home"}}, nil
		case "getChangeContributors":
			return []string{"alice@example.com", "bob@example.com"}, nil

		// Publishing
		case "publish":
			return map[string]any{"deploymentId": "deploy-1", "url": "https://preview.framer.app"}, nil
		case "deploy":
			return map[string]any{"id": "deploy-1", "url": "https://example.com", "createdAt": "2026-03-16T10:00:00Z"}, nil
		case "getDeployments":
			return []map[string]any{
				{"id": "deploy-1", "url": "https://example.com", "createdAt": "2026-03-16T10:00:00Z"},
				{"id": "deploy-2", "url": "https://staging.example.com", "createdAt": "2026-03-15T10:00:00Z"},
			}, nil
		case "getPublishInfo":
			return map[string]any{"url": "https://example.com", "lastPublished": "2026-03-16T10:00:00Z"}, nil

		// Collections
		case "getCollections":
			return []map[string]any{
				{"id": "col-1", "name": "Blog Posts"},
				{"id": "col-2", "name": "Products"},
			}, nil
		case "getCollection":
			return map[string]any{"id": params["id"], "name": "Blog Posts"}, nil
		case "createCollection":
			return map[string]any{"id": "col-new", "name": params["name"]}, nil
		case "getCollectionFields":
			return []map[string]any{
				{"id": "field-1", "name": "Title", "type": "string"},
				{"id": "field-2", "name": "Body", "type": "richtext"},
			}, nil
		case "addCollectionFields":
			return []map[string]any{
				{"id": "field-3", "name": "Author", "type": "string"},
			}, nil
		case "removeCollectionFields", "setCollectionFieldOrder":
			return map[string]any{"success": true}, nil
		case "getCollectionItems":
			return []map[string]any{
				{"id": "item-1", "slug": "hello-world", "fieldData": map[string]any{"Title": "Hello World"}},
				{"id": "item-2", "slug": "second-post", "fieldData": map[string]any{"Title": "Second Post"}},
			}, nil
		case "addCollectionItems":
			return []map[string]any{
				{"id": "item-new", "slug": "new-item"},
			}, nil
		case "removeCollectionItems", "setCollectionItemOrder":
			return map[string]any{"success": true}, nil

		// Managed Collections
		case "getManagedCollections":
			return []map[string]any{
				{"id": "mcol-1", "name": "Managed Posts"},
			}, nil
		case "createManagedCollection":
			return map[string]any{"id": "mcol-new", "name": params["name"]}, nil
		case "getManagedCollectionFields":
			return []map[string]any{
				{"id": "mfield-1", "name": "Title", "type": "string"},
			}, nil
		case "setManagedCollectionFields":
			return []map[string]any{
				{"id": "mfield-1", "name": "Title", "type": "string"},
			}, nil
		case "addManagedCollectionItems":
			return []map[string]any{
				{"id": "item-new", "slug": "new-item"},
			}, nil
		case "removeManagedCollectionItems", "setManagedCollectionItemOrder":
			return map[string]any{"success": true}, nil
		case "getManagedCollectionItemIds":
			return []string{"item-1", "item-2"}, nil

		// Nodes
		case "getNode":
			return map[string]any{"id": params["nodeId"], "name": "Frame 1", "__class": "FrameNode"}, nil
		case "getChildren":
			return []map[string]any{
				{"id": "child-1", "name": "Text", "__class": "TextNode"},
				{"id": "child-2", "name": "Image", "__class": "FrameNode"},
			}, nil
		case "getParent":
			return map[string]any{"id": "parent-1", "name": "Page", "__class": "WebPageNode"}, nil
		case "getNodesWithType":
			nodeType := ""
			if t, ok := params["type"]; ok {
				nodeType, _ = t.(string)
			}
			return []map[string]any{
				{"id": "node-1", "name": "Frame A", "__class": nodeType},
				{"id": "node-2", "name": "Frame B", "__class": nodeType},
			}, nil
		case "createFrameNode":
			return map[string]any{"id": "frame-new", "name": "New Frame", "__class": "FrameNode"}, nil
		case "createTextNode":
			return map[string]any{"id": "text-new", "name": "New Text", "__class": "TextNode"}, nil
		case "createComponentNode":
			return map[string]any{"id": "comp-new", "name": params["name"], "__class": "ComponentNode"}, nil
		case "createWebPage":
			return map[string]any{"id": "page-new", "name": "New Page", "__class": "WebPageNode"}, nil
		case "createDesignPage":
			return map[string]any{"id": "design-new", "name": params["name"], "__class": "DesignPageNode"}, nil
		case "cloneNode":
			return map[string]any{"id": "clone-1", "name": "Frame 1 Copy", "__class": "FrameNode"}, nil
		case "removeNodes":
			return map[string]any{"success": true}, nil
		case "setAttributes":
			return map[string]any{"id": params["nodeId"], "name": "Updated", "__class": "FrameNode"}, nil
		case "setParent":
			return map[string]any{"id": params["nodeId"], "name": "Moved Node", "__class": "FrameNode"}, nil
		case "getRect":
			return map[string]any{"x": 0.0, "y": 0.0, "width": 100.0, "height": 200.0}, nil
		case "getCanvasRoot":
			return map[string]any{"id": "root", "name": "Canvas Root", "__class": "CanvasRootNode"}, nil

		// Agent
		case "getAgentSystemPrompt":
			return "You are a Framer design agent. Use DSL commands to modify the project.", nil
		case "getAgentContext":
			return map[string]any{"fonts": []string{"Inter"}, "components": []string{"Button", "Card"}}, nil
		case "readProjectForAgent":
			return []map[string]any{{"query": "pages", "result": "2 pages found"}}, nil
		case "applyAgentChanges":
			return map[string]any{"applied": true, "nodesCreated": 3}, nil

		// Styles
		case "getColorStyles":
			return []map[string]any{
				{"id": "cs-1", "name": "Primary", "light": "#0066FF", "dark": "#3399FF"},
				{"id": "cs-2", "name": "Background", "light": "#FFFFFF", "dark": "#1A1A1A"},
			}, nil
		case "getColorStyle":
			return map[string]any{"id": params["id"], "name": "Primary", "light": "#0066FF", "dark": "#3399FF"}, nil
		case "createColorStyle":
			return map[string]any{"id": "cs-new", "name": "New Color"}, nil
		case "getTextStyles":
			return []map[string]any{
				{"id": "ts-1", "name": "Heading 1", "font": "Inter", "fontSize": 32.0},
				{"id": "ts-2", "name": "Body", "font": "Inter", "fontSize": 16.0},
			}, nil
		case "getTextStyle":
			return map[string]any{"id": params["id"], "name": "Heading 1", "font": "Inter", "fontSize": 32.0}, nil
		case "createTextStyle":
			return map[string]any{"id": "ts-new", "name": "New Text Style"}, nil

		// Fonts
		case "getFonts":
			return []map[string]any{
				{"family": "Inter", "style": "normal", "weight": 400},
				{"family": "Inter", "style": "normal", "weight": 700},
			}, nil
		case "getFont":
			if family, ok := params["family"].(string); ok && family == "NotFound" {
				return nil, nil
			}
			return map[string]any{"family": params["family"], "style": "normal", "weight": 400}, nil

		// Locales
		case "getLocales":
			return []map[string]any{
				{"id": "loc-1", "code": "en-US", "name": "English", "slug": "en"},
				{"id": "loc-2", "code": "fr-FR", "name": "French", "slug": "fr"},
			}, nil
		case "getDefaultLocale":
			return map[string]any{"id": "loc-1", "code": "en-US", "name": "English", "slug": "en"}, nil
		case "createLocale":
			return map[string]any{"id": "loc-new", "code": "de-DE", "name": "German", "slug": "de"}, nil
		case "getLocaleLanguages":
			return []map[string]any{
				{"code": "en", "name": "English"},
				{"code": "fr", "name": "French"},
			}, nil
		case "getLocaleRegions":
			return []map[string]any{
				{"code": "US", "name": "United States", "isCommon": true},
				{"code": "GB", "name": "United Kingdom", "isCommon": true},
			}, nil
		case "getLocalizationGroups":
			return []map[string]any{{"id": "grp-1", "name": "Default"}}, nil
		case "setLocalizationData":
			return map[string]any{"success": true}, nil

		// Redirects
		case "getRedirects":
			return []map[string]any{
				{"id": "redir-1", "from": "/old", "to": "/new", "type": 301},
				{"id": "redir-2", "from": "/legacy", "to": "/modern", "type": 302},
			}, nil
		case "addRedirects":
			return []map[string]any{
				{"id": "redir-new", "from": "/test", "to": "/dest", "type": 301},
			}, nil
		case "removeRedirects", "setRedirectOrder":
			return map[string]any{"success": true}, nil

		// Code
		case "getCodeFiles":
			return []map[string]any{
				{"id": "code-1", "name": "analytics.tsx"},
				{"id": "code-2", "name": "utils.ts"},
			}, nil
		case "getCodeFile":
			return map[string]any{"id": params["id"], "name": "analytics.tsx"}, nil
		case "createCodeFile":
			return map[string]any{"id": "code-new", "name": params["name"]}, nil
		case "typecheckCode":
			return []map[string]any{}, nil
		case "getCustomCode":
			return map[string]any{"headEnd": "<script>console.log('hi')</script>"}, nil
		case "setCustomCode":
			return map[string]any{"headEnd": "<script>alert(1)</script>"}, nil

		// Images & Files
		case "uploadImage":
			return map[string]any{"url": "https://framer.com/images/uploaded.png", "id": "img-1"}, nil
		case "uploadImages":
			return []map[string]any{
				{"url": "https://framer.com/images/1.png", "id": "img-1"},
				{"url": "https://framer.com/images/2.png", "id": "img-2"},
			}, nil
		case "uploadFile":
			return map[string]any{"url": "https://framer.com/files/uploaded.pdf", "id": "file-1"}, nil
		case "uploadFiles":
			return []map[string]any{
				{"url": "https://framer.com/files/1.pdf", "id": "file-1"},
			}, nil

		// SVG
		case "addSVG":
			return map[string]any{"success": true}, nil
		case "getVectorSets":
			return []map[string]any{{"id": "vs-1", "name": "Icons"}}, nil

		// Screenshots
		case "screenshot":
			return map[string]any{"image": "iVBORw0KGgo=", "url": "https://framer.com/screenshots/1.png"}, nil
		case "exportSVG":
			return "<svg xmlns=\"http://www.w3.org/2000/svg\"><rect width=\"100\" height=\"100\"/></svg>", nil

		// Plugin Data
		case "getPluginData":
			return "stored-value", nil
		case "setPluginData":
			return map[string]any{"success": true}, nil
		case "getPluginDataKeys":
			return []string{"key1", "key2", "key3"}, nil

		default:
			return nil, fmt.Errorf("unknown method: %s", method)
		}
	}
}
