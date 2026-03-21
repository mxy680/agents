package imessage

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newTestClient creates a Client pointing at the given test server.
func newTestClient(server *httptest.Server) *Client {
	return newClientWithBase(server.Client(), server.URL, "test-password")
}

// newTestClientFactory returns a ClientFactory pointing at the given test server.
func newTestClientFactory(server *httptest.Server) ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		return newTestClient(server), nil
	}
}

// newTestRootCmd creates a root command with --json and --dry-run flags.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "Output as JSON")
	root.PersistentFlags().Bool("dry-run", false, "Preview actions")
	return root
}

// captureStdout captures stdout during f() and returns the output.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	buf := make([]byte, 64*1024)
	n, _ := r.Read(buf)
	r.Close()
	return string(buf[:n])
}

// bbResponse wraps data in the standard BlueBubbles envelope.
func bbResponse(data any) string {
	switch v := data.(type) {
	case string:
		return `{"status":200,"message":"Success","data":` + v + `}`
	default:
		return fmt.Sprintf(`{"status":200,"message":"Success","data":%v}`, v)
	}
}

// bbJSONResponse wraps pre-marshaled JSON in the BlueBubbles envelope.
func bbJSONResponse(jsonData string) string {
	return `{"status":200,"message":"Success","data":` + jsonData + `}`
}

// bbErrorResponse creates an error response.
func bbErrorResponse(status int, msg string) string {
	return fmt.Sprintf(`{"status":%d,"message":"%s","error":{"message":"%s"}}`, status, msg, msg)
}

// containsStr is a test helper that checks if s contains sub.
func containsStr(s, sub string) bool {
	return strings.Contains(s, sub)
}

// Reusable mock data constants.
const (
	mockChatGUID1 = "iMessage;-;chat123456"
	mockChatGUID2 = "iMessage;+;chat789012"
	mockMsgGUID1  = "msg-guid-001"
	mockMsgGUID2  = "msg-guid-002"
	mockPhone1    = "+1234567890"
	mockPhone2    = "+0987654321"
	mockAttGUID1  = "att-guid-001"
)

func mockChatJSON(guid, displayName string) string {
	return fmt.Sprintf(`{
		"guid": %q,
		"displayName": %q,
		"chatIdentifier": %q,
		"isArchived": false,
		"participants": [
			{"handle": {"address": %q}}
		],
		"lastMessage": {"text": "Hello from mock"}
	}`, guid, displayName, guid, mockPhone1)
}

func mockMessageJSON(guid string, isFromMe bool) string {
	return fmt.Sprintf(`{
		"guid": %q,
		"text": "Test message text",
		"isFromMe": %v,
		"dateCreated": 1700000000000,
		"handle": {"address": %q},
		"chats": [{"guid": %q}],
		"attachments": []
	}`, guid, isFromMe, mockPhone1, mockChatGUID1)
}

func mockAttachmentJSON(guid string) string {
	return fmt.Sprintf(`{
		"guid": %q,
		"transferName": "photo.jpg",
		"mimeType": "image/jpeg",
		"totalBytes": 204800,
		"isOutgoing": false,
		"createdDate": 1700000000000
	}`, guid)
}

func mockHandleJSON(address string) string {
	return fmt.Sprintf(`{
		"address": %q,
		"service": "iMessage",
		"country": "US",
		"uncanonicalizedId": %q
	}`, address, address)
}

func mockContactJSON(id, firstName, lastName string) string {
	return fmt.Sprintf(`{
		"id": %q,
		"firstName": %q,
		"lastName": %q,
		"displayName": %q,
		"phoneNumbers": [{"address": %q}],
		"emails": [{"address": "test@example.com"}]
	}`, id, firstName, lastName, firstName+" "+lastName, mockPhone1)
}

func mockScheduledJSON(id int) string {
	return fmt.Sprintf(`{
		"id": %d,
		"chatGuid": %q,
		"message": "Scheduled test message",
		"scheduledFor": 1800000000000,
		"status": "pending"
	}`, id, mockChatGUID1)
}

func mockWebhookJSON(id int, url string) string {
	return fmt.Sprintf(`{
		"id": %d,
		"url": %q,
		"events": ["new-message", "updated-message"]
	}`, id, url)
}

// withChatMocks registers chat-related mock handlers on mux.
func withChatMocks(mux *http.ServeMux) {
	// POST /api/v1/chat/query
	mux.HandleFunc("/api/v1/chat/query", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		chats := `[` + mockChatJSON(mockChatGUID1, "Group Chat") + `,` + mockChatJSON(mockChatGUID2, "") + `]`
		w.Write([]byte(bbJSONResponse(chats)))
	})

	// GET /api/v1/chat/count
	mux.HandleFunc("/api/v1/chat/count", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"total": 42}`)))
	})

	// POST /api/v1/chat/new
	mux.HandleFunc("/api/v1/chat/new", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(mockChatJSON(mockChatGUID1, "New Chat"))))
	})

	// Wildcard for /api/v1/chat/{guid}[/sub-resource]
	mux.HandleFunc("/api/v1/chat/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/chat/")

		switch {
		case strings.HasSuffix(path, "/message"):
			msgs := `[` + mockMessageJSON(mockMsgGUID1, false) + `,` + mockMessageJSON(mockMsgGUID2, true) + `]`
			w.Write([]byte(bbJSONResponse(msgs)))
		case strings.HasSuffix(path, "/read"):
			w.Write([]byte(bbJSONResponse(`{"success": true}`)))
		case strings.HasSuffix(path, "/unread"):
			w.Write([]byte(bbJSONResponse(`{"success": true}`)))
		case strings.HasSuffix(path, "/leave"):
			w.Write([]byte(bbJSONResponse(`{"success": true}`)))
		case strings.HasSuffix(path, "/typing") && r.Method == http.MethodPost:
			w.Write([]byte(bbJSONResponse(`{"success": true}`)))
		case strings.HasSuffix(path, "/typing") && r.Method == http.MethodDelete:
			w.Write([]byte(bbJSONResponse(`{"success": true}`)))
		case strings.HasSuffix(path, "/icon") && r.Method == http.MethodPost:
			w.Write([]byte(bbJSONResponse(`{"updated": true}`)))
		case strings.HasSuffix(path, "/icon") && r.Method == http.MethodDelete:
			w.Write([]byte(bbJSONResponse(`{"success": true}`)))
		case strings.HasSuffix(path, "/participant"):
			w.Write([]byte(bbJSONResponse(mockChatJSON(mockChatGUID1, "Group Chat"))))
		case strings.HasSuffix(path, "/participant/remove"):
			w.Write([]byte(bbJSONResponse(mockChatJSON(mockChatGUID1, "Group Chat"))))
		case r.Method == http.MethodDelete:
			w.Write([]byte(bbJSONResponse(`{"success": true}`)))
		case r.Method == http.MethodPut:
			guid := strings.Split(path, "/")[0]
			w.Write([]byte(bbJSONResponse(mockChatJSON(guid, "Updated Chat"))))
		default:
			// GET /api/v1/chat/{guid}
			guid := strings.Split(path, "/")[0]
			w.Write([]byte(bbJSONResponse(mockChatJSON(guid, "Test Chat"))))
		}
	})
}

// withMessageMocks registers message-related mock handlers on mux.
func withMessageMocks(mux *http.ServeMux) {
	// POST /api/v1/message/query
	mux.HandleFunc("/api/v1/message/query", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		msgs := `[` + mockMessageJSON(mockMsgGUID1, false) + `,` + mockMessageJSON(mockMsgGUID2, true) + `]`
		w.Write([]byte(bbJSONResponse(msgs)))
	})

	// POST /api/v1/message/text
	mux.HandleFunc("/api/v1/message/text", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(mockMessageJSON(mockMsgGUID1, true))))
	})

	// POST /api/v1/message/multipart
	mux.HandleFunc("/api/v1/message/multipart", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(mockMessageJSON(mockMsgGUID1, true))))
	})

	// POST /api/v1/message/react
	mux.HandleFunc("/api/v1/message/react", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"success": true}`)))
	})

	// GET /api/v1/message/count/updated
	mux.HandleFunc("/api/v1/message/count/updated", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"total": 5}`)))
	})

	// GET /api/v1/message/count/me
	mux.HandleFunc("/api/v1/message/count/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"total": 50}`)))
	})

	// GET /api/v1/message/count
	mux.HandleFunc("/api/v1/message/count", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"total": 100}`)))
	})

	// GET/POST /api/v1/message/schedule[/{id}]
	mux.HandleFunc("/api/v1/message/schedule", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.Write([]byte(bbJSONResponse(mockScheduledJSON(1))))
		default:
			items := `[` + mockScheduledJSON(1) + `,` + mockScheduledJSON(2) + `]`
			w.Write([]byte(bbJSONResponse(items)))
		}
	})

	// Wildcard for /api/v1/message/{guid}[/sub-resource] and /api/v1/message/schedule/{id}
	mux.HandleFunc("/api/v1/message/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/message/")

		switch {
		case strings.HasPrefix(path, "schedule/"):
			switch r.Method {
			case http.MethodPut:
				w.Write([]byte(bbJSONResponse(mockScheduledJSON(1))))
			case http.MethodDelete:
				w.Write([]byte(bbJSONResponse(`{"success": true}`)))
			default:
				w.Write([]byte(bbJSONResponse(mockScheduledJSON(1))))
			}
		case strings.HasSuffix(path, "/edit"):
			w.Write([]byte(bbJSONResponse(mockMessageJSON(mockMsgGUID1, true))))
		case strings.HasSuffix(path, "/unsend"):
			w.Write([]byte(bbJSONResponse(`{"success": true}`)))
		case strings.HasSuffix(path, "/embedded-media"):
			media := `[` + mockAttachmentJSON(mockAttGUID1) + `]`
			w.Write([]byte(bbJSONResponse(media)))
		case strings.HasSuffix(path, "/notify"):
			w.Write([]byte(bbJSONResponse(`{"success": true}`)))
		case r.Method == http.MethodDelete:
			w.Write([]byte(bbJSONResponse(`{"success": true}`)))
		default:
			w.Write([]byte(bbJSONResponse(mockMessageJSON(mockMsgGUID1, false))))
		}
	})
}

// withHandleMocks registers handle-related mock handlers on mux.
func withHandleMocks(mux *http.ServeMux) {
	// POST /api/v1/handle/query
	mux.HandleFunc("/api/v1/handle/query", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handles := `[` + mockHandleJSON(mockPhone1) + `,` + mockHandleJSON(mockPhone2) + `]`
		w.Write([]byte(bbJSONResponse(handles)))
	})

	// GET /api/v1/handle/count
	mux.HandleFunc("/api/v1/handle/count", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"total": 20}`)))
	})

	// GET /api/v1/handle/availability/{service}
	mux.HandleFunc("/api/v1/handle/availability/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"available": true, "address": "+1234567890"}`)))
	})

	// Wildcard for /api/v1/handle/{guid}[/focus]
	mux.HandleFunc("/api/v1/handle/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/handle/")

		if strings.HasSuffix(path, "/focus") {
			w.Write([]byte(bbJSONResponse(`{"focusStatus": "focused"}`)))
			return
		}
		w.Write([]byte(bbJSONResponse(mockHandleJSON(mockPhone1))))
	})
}

// withAttachmentMocks registers attachment-related mock handlers on mux.
func withAttachmentMocks(mux *http.ServeMux) {
	// GET /api/v1/attachment/count
	mux.HandleFunc("/api/v1/attachment/count", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"total": 15}`)))
	})

	// POST /api/v1/attachment/upload
	mux.HandleFunc("/api/v1/attachment/upload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(mockAttachmentJSON(mockAttGUID1))))
	})

	// Wildcard for /api/v1/attachment/{guid}[/download|/live|/blurhash]
	mux.HandleFunc("/api/v1/attachment/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/attachment/")

		if strings.HasSuffix(path, "/download") || strings.HasSuffix(path, "/download/force") || strings.HasSuffix(path, "/live") {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write([]byte("mock binary data"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(path, "/blurhash") {
			w.Write([]byte(bbJSONResponse(`{"blurhash": "L6PZfSi_.AyE_3t7t7R**0o#DgR4"}`)))
			return
		}

		w.Write([]byte(bbJSONResponse(mockAttachmentJSON(mockAttGUID1))))
	})
}

// withContactMocks registers contact-related mock handlers on mux.
func withContactMocks(mux *http.ServeMux) {
	// GET /api/v1/contact
	// POST /api/v1/contact (create)
	mux.HandleFunc("/api/v1/contact", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.Write([]byte(bbJSONResponse(mockContactJSON("contact-001", "Jane", "Doe"))))
		default:
			contacts := `[` + mockContactJSON("contact-001", "John", "Smith") + `,` + mockContactJSON("contact-002", "Jane", "Doe") + `]`
			w.Write([]byte(bbJSONResponse(contacts)))
		}
	})

	// POST /api/v1/contact/query
	mux.HandleFunc("/api/v1/contact/query", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		contacts := `[` + mockContactJSON("contact-001", "John", "Smith") + `]`
		w.Write([]byte(bbJSONResponse(contacts)))
	})
}

// withServerMocks registers server-related mock handlers on mux.
func withServerMocks(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/server/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{
			"server_version": "1.9.0",
			"os_version": "macOS 14.0",
			"detected_icloud": "MacBook Pro",
			"private_api": true,
			"proxy_service": "Dynamic DNS"
		}`)))
	})

	mux.HandleFunc("/api/v1/server/logs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`"[INFO] Server started successfully"`)))
	})

	mux.HandleFunc("/api/v1/server/restart/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"success": true}`)))
	})

	mux.HandleFunc("/api/v1/server/update/check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"available": false, "version": "1.9.0"}`)))
	})

	mux.HandleFunc("/api/v1/server/update/install", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"success": true}`)))
	})

	mux.HandleFunc("/api/v1/server/alert", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`[]`)))
	})

	mux.HandleFunc("/api/v1/server/alert/read", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"success": true}`)))
	})

	mux.HandleFunc("/api/v1/server/statistics/totals", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"messages": 1000, "chats": 42}`)))
	})

	mux.HandleFunc("/api/v1/server/statistics/media", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"images": 50, "videos": 10}`)))
	})

	mux.HandleFunc("/api/v1/server/statistics/media/chat", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`[{"guid": "chat123", "count": 15}]`)))
	})
}

// withWebhookMocks registers webhook-related mock handlers on mux.
func withWebhookMocks(mux *http.ServeMux) {
	// GET /api/v1/webhook and POST /api/v1/webhook
	mux.HandleFunc("/api/v1/webhook", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.Write([]byte(bbJSONResponse(mockWebhookJSON(1, "https://example.com/hook"))))
		default:
			hooks := `[` + mockWebhookJSON(1, "https://example.com/hook") + `]`
			w.Write([]byte(bbJSONResponse(hooks)))
		}
	})

	// DELETE /api/v1/webhook/{id}
	mux.HandleFunc("/api/v1/webhook/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"success": true}`)))
	})
}

// withFaceTimeMocks registers FaceTime-related mock handlers on mux.
func withFaceTimeMocks(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/facetime/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"success": true}`)))
	})
}

// withFindMyMocks registers FindMy-related mock handlers on mux.
func withFindMyMocks(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/findmy/devices", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		devices := `[{"id": "device-001", "name": "My iPhone", "batteryLevel": 0.75}]`
		w.Write([]byte(bbJSONResponse(devices)))
	})

	mux.HandleFunc("/api/v1/findmy/friends", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		friends := `[{"id": "friend-001", "handle": "+1234567890", "firstName": "Alice"}]`
		w.Write([]byte(bbJSONResponse(friends)))
	})
}

// withICloudMocks registers iCloud-related mock handlers on mux.
func withICloudMocks(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/icloud/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"success": true}`)))
	})
}

// withMacMocks registers Mac-related mock handlers on mux.
func withMacMocks(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/mac/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(bbJSONResponse(`{"success": true}`)))
	})
}

// newFullMockServer creates an httptest.Server with all BlueBubbles API endpoints registered.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	withChatMocks(mux)
	withMessageMocks(mux)
	withHandleMocks(mux)
	withAttachmentMocks(mux)
	withContactMocks(mux)
	withServerMocks(mux)
	withWebhookMocks(mux)
	withFaceTimeMocks(mux)
	withFindMyMocks(mux)
	withICloudMocks(mux)
	withMacMocks(mux)

	return httptest.NewServer(mux)
}
