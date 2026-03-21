package imessage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --- newClientWithBase / buildURL ---

func TestBuildURL(t *testing.T) {
	c := newClientWithBase(&http.Client{}, "http://localhost:1234", "mypassword")

	t.Run("password is appended as query param", func(t *testing.T) {
		got := c.buildURL("chat/query", nil)
		if !strings.Contains(got, "password=mypassword") {
			t.Errorf("buildURL output %q does not contain password param", got)
		}
	})

	t.Run("path is appended to api/v1", func(t *testing.T) {
		got := c.buildURL("chat/count", nil)
		if !strings.Contains(got, "/api/v1/chat/count") {
			t.Errorf("buildURL output %q does not contain expected path", got)
		}
	})

	t.Run("leading slash is stripped from path", func(t *testing.T) {
		got := c.buildURL("/chat/count", nil)
		// Should not produce double slash in the path segment.
		if strings.Contains(got, "/api/v1//") {
			t.Errorf("buildURL output %q has double slash", got)
		}
		if !strings.Contains(got, "/api/v1/chat/count") {
			t.Errorf("buildURL output %q does not contain expected path", got)
		}
	})

	t.Run("base trailing slash is stripped", func(t *testing.T) {
		c2 := newClientWithBase(&http.Client{}, "http://localhost:1234/", "pw")
		got := c2.buildURL("server/info", nil)
		if strings.Contains(got, "//api") {
			t.Errorf("buildURL output %q has double slash before api", got)
		}
	})
}

// --- Client.Get ---

func TestClientGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":200,"message":"OK","data":{}}`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		body, err := client.Get(context.Background(), "test/path", nil)
		if err != nil {
			t.Fatalf("Get returned error: %v", err)
		}
		if len(body) == 0 {
			t.Error("Get returned empty body")
		}
	})

	t.Run("non-2xx status returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`not found`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		_, err := client.Get(context.Background(), "missing", nil)
		if err == nil {
			t.Error("expected error for 404, got nil")
		}
		if !strings.Contains(err.Error(), "404") {
			t.Errorf("error should mention HTTP status, got: %v", err)
		}
	})
}

// --- Client.Post ---

func TestClientPost(t *testing.T) {
	t.Run("success with JSON body", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if ct := r.Header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":200,"message":"OK","data":{"created":true}}`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		body, err := client.Post(context.Background(), "test/path", map[string]any{"key": "value"})
		if err != nil {
			t.Fatalf("Post returned error: %v", err)
		}
		if len(body) == 0 {
			t.Error("Post returned empty body")
		}
	})

	t.Run("nil body sends no content-type", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ct := r.Header.Get("Content-Type"); ct != "" {
				t.Errorf("nil body should not set Content-Type, got %q", ct)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":200,"message":"OK","data":null}`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		_, err := client.Post(context.Background(), "test/path", nil)
		if err != nil {
			t.Fatalf("Post with nil body returned error: %v", err)
		}
	})
}

// --- Client.Put ---

func TestClientPut(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("expected PUT, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":200,"message":"OK","data":{"updated":true}}`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		body, err := client.Put(context.Background(), "test/update", map[string]any{"name": "new-name"})
		if err != nil {
			t.Fatalf("Put returned error: %v", err)
		}
		if len(body) == 0 {
			t.Error("Put returned empty body")
		}
	})

	t.Run("non-2xx returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`server error`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		_, err := client.Put(context.Background(), "test/fail", nil)
		if err == nil {
			t.Error("expected error for 500, got nil")
		}
	})
}

// --- Client.Delete ---

func TestClientDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodDelete {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":200,"message":"OK","data":{"deleted":true}}`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		body, err := client.Delete(context.Background(), "test/resource/123")
		if err != nil {
			t.Fatalf("Delete returned error: %v", err)
		}
		if len(body) == 0 {
			t.Error("Delete returned empty body")
		}
	})

	t.Run("non-2xx returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`forbidden`))
		}))
		defer srv.Close()

		client := newTestClient(srv)
		_, err := client.Delete(context.Background(), "test/protected")
		if err == nil {
			t.Error("expected error for 403, got nil")
		}
		if !strings.Contains(err.Error(), "403") {
			t.Errorf("error should mention status 403, got: %v", err)
		}
	})
}

// --- ParseResponse ---

func TestParseResponse(t *testing.T) {
	t.Run("success extracts data field", func(t *testing.T) {
		input := []byte(`{"status":200,"message":"Success","data":{"key":"value"}}`)
		data, err := ParseResponse(input)
		if err != nil {
			t.Fatalf("ParseResponse returned error: %v", err)
		}
		if string(data) != `{"key":"value"}` {
			t.Errorf("ParseResponse data = %s, want {\"key\":\"value\"}", string(data))
		}
	})

	t.Run("non-200 status returns error", func(t *testing.T) {
		input := []byte(`{"status":404,"message":"Not Found","error":{"message":"chat not found"}}`)
		_, err := ParseResponse(input)
		if err == nil {
			t.Error("expected error for status 404, got nil")
		}
		if !strings.Contains(err.Error(), "API error") {
			t.Errorf("error should mention 'API error', got: %v", err)
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		_, err := ParseResponse([]byte(`not json`))
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})

	t.Run("null data is valid", func(t *testing.T) {
		input := []byte(`{"status":200,"message":"Success","data":null}`)
		data, err := ParseResponse(input)
		if err != nil {
			t.Fatalf("ParseResponse with null data returned error: %v", err)
		}
		if string(data) != "null" {
			t.Errorf("ParseResponse null data = %s, want null", string(data))
		}
	})
}

// --- ParseResponseRaw ---

func TestParseResponseRaw(t *testing.T) {
	t.Run("returns raw body unchanged", func(t *testing.T) {
		input := []byte(`{"some":"raw","data":123}`)
		data, err := ParseResponseRaw(input)
		if err != nil {
			t.Fatalf("ParseResponseRaw returned error: %v", err)
		}
		if string(data) != string(input) {
			t.Errorf("ParseResponseRaw = %s, want %s", string(data), string(input))
		}
	})
}
