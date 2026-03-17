package instagram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
)

// withProfileMock registers all profile-related mock handlers on mux.
func withProfileMock(mux *http.ServeMux) {
	// GET /api/v1/users/web_profile_info/?username=X
	mux.HandleFunc("/api/v1/users/web_profile_info/", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, `{"status":"fail","message":"missing username"}`, http.StatusBadRequest)
			return
		}
		resp := map[string]any{
			"data": map[string]any{
				"user": map[string]any{
					"id":                      "42544748138",
					"username":                username,
					"full_name":               "Test User",
					"is_private":              false,
					"is_verified":             false,
					"biography":               "Test bio",
					"external_url":            "https://example.com",
					"profile_pic_url":         "https://example.com/pic.jpg",
					"is_business_account":     false,
					"is_professional_account": false,
					"category_name":           "",
					"edge_followed_by":        map[string]any{"count": 100},
					"edge_follow":             map[string]any{"count": 50},
					"edge_owner_to_timeline_media": map[string]any{"count": 10},
				},
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /api/v1/users/{id}/info/
	mux.HandleFunc("/api/v1/users/", func(w http.ResponseWriter, r *http.Request) {
		// path: /api/v1/users/{id}/info/
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) < 2 || parts[1] != "info/" {
			http.NotFound(w, r)
			return
		}
		userID := parts[0]
		resp := map[string]any{
			"user": map[string]any{
				"pk":                     userID,
				"username":               "testuser",
				"full_name":              "Test User",
				"is_private":             false,
				"is_verified":            false,
				"biography":              "Bio from user info",
				"external_url":           "",
				"follower_count":         int64(200),
				"following_count":        int64(80),
				"media_count":            int64(15),
				"total_clips_count":      int64(3),
				"is_business":            false,
				"account_type":           1,
				"profile_pic_url":        "https://example.com/pic2.jpg",
				"has_profile_pic":        true,
				"is_professional_account": false,
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /api/v1/accounts/edit/web_form_data/
	mux.HandleFunc("/api/v1/accounts/edit/web_form_data/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"form_data": map[string]any{
				"first_name":        "Test",
				"last_name":         "User",
				"email":             "testuser@example.com",
				"username":          "testuser",
				"phone_number":      "+15551234567",
				"gender":            1,
				"biography":         "My bio",
				"external_url":      "https://example.com",
				"is_email_confirmed": true,
				"is_phone_confirmed": false,
				"business_account":  false,
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withMediaMock registers all media-related mock handlers on mux.
func withMediaMock(mux *http.ServeMux) {
	// GET /api/v1/feed/user/{user_id}/
	mux.HandleFunc("/api/v1/feed/user/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/feed/user/")
		// Route story requests: /api/v1/feed/user/{id}/story/
		if strings.HasSuffix(path, "/story/") {
			withStoriesListHandler(w, r, path)
			return
		}
		// Route media feed: /api/v1/feed/user/{id}/
		resp := map[string]any{
			"items": []map[string]any{
				{
					"id":            "111222333",
					"code":          "abc123",
					"media_type":    1,
					"caption":       map[string]any{"text": "Test caption"},
					"taken_at":      int64(1700000000),
					"like_count":    int64(42),
					"comment_count": int64(7),
					"image_versions2": map[string]any{
						"candidates": []map[string]any{
							{"url": "https://example.com/img.jpg"},
						},
					},
				},
			},
			"next_max_id":    "cursor_abc",
			"more_available": true,
			"status":         "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /api/v1/media/{id}/info/  — shared by media get, story get, reel get
	mux.HandleFunc("/api/v1/media/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/media/")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}
		mediaID := parts[0]
		action := strings.TrimSuffix(parts[1], "/")

		switch {
		case action == "info":
			resp := map[string]any{
				"items": []map[string]any{
					{
						"id":            mediaID,
						"code":          "abc123",
						"media_type":    1,
						"caption":       map[string]any{"text": "Single post"},
						"taken_at":      int64(1700000000),
						"like_count":    int64(100),
						"comment_count": int64(10),
						"play_count":    int64(500),
						"image_versions2": map[string]any{
							"candidates": []map[string]any{
								{"url": "https://example.com/img.jpg"},
							},
						},
					},
				},
				"status": "ok",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		case action == "delete" || strings.HasPrefix(action, "delete?"):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"did_delete": true, "status": "ok"})

		case action == "only_me":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})

		case action == "undo_only_me":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})

		case action == "likers":
			resp := map[string]any{
				"users": []map[string]any{
					{
						"pk":              "555666777",
						"username":        "liker_user",
						"full_name":       "Liker User",
						"profile_pic_url": "https://example.com/liker.jpg",
						"is_private":      false,
						"is_verified":     false,
					},
				},
				"status": "ok",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		case action == "save":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})

		case action == "unsave":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})

		case action == "list_reel_media_viewer":
			resp := map[string]any{
				"users": []map[string]any{
					{
						"pk":              "888999000",
						"username":        "viewer_user",
						"full_name":       "Viewer User",
						"profile_pic_url": "https://example.com/viewer.jpg",
						"is_private":      false,
						"is_verified":     true,
					},
				},
				"user_count": int64(1),
				"next_max_id": "",
				"status":      "ok",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		default:
			http.NotFound(w, r)
		}
	})
}

// withStoriesListHandler handles GET /api/v1/feed/user/{id}/story/ requests.
func withStoriesListHandler(w http.ResponseWriter, _ *http.Request, path string) {
	userID := strings.TrimSuffix(path, "/story/")
	_ = userID
	resp := map[string]any{
		"reel": map[string]any{
			"items": []map[string]any{
				{
					"id":          "story_111",
					"media_type":  1,
					"taken_at":    int64(1700000000),
					"expiring_at": int64(1700086400),
					"image_versions2": map[string]any{
						"candidates": []map[string]any{
							{"url": "https://example.com/story.jpg"},
						},
					},
				},
			},
			"status": "ok",
		},
		"status": "ok",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// withStoriesMock registers stories-only mock handlers on mux (reels_tray).
func withStoriesMock(mux *http.ServeMux) {
	// GET /api/v1/feed/reels_tray/
	mux.HandleFunc("/api/v1/feed/reels_tray/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"tray": []map[string]any{
				{
					"id": "tray_entry_1",
					"user": map[string]any{
						"pk":              "777888999",
						"username":        "followed_user",
						"full_name":       "Followed User",
						"profile_pic_url": "https://example.com/followed.jpg",
						"is_private":      false,
						"is_verified":     false,
					},
					"items": []map[string]any{
						{"id": "story_aaa"},
						{"id": "story_bbb"},
					},
					"seen": int64(0),
				},
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withReelsMock registers reels-only mock handlers on mux.
func withReelsMock(mux *http.ServeMux) {
	// GET /api/v1/clips/user/
	mux.HandleFunc("/api/v1/clips/user/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"items": []map[string]any{
				{
					"media": map[string]any{
						"id":            "reel_111",
						"code":          "reel_code_1",
						"media_type":    2,
						"caption":       map[string]any{"text": "My reel caption"},
						"taken_at":      int64(1700000000),
						"like_count":    int64(200),
						"comment_count": int64(15),
						"play_count":    int64(1000),
						"image_versions2": map[string]any{
							"candidates": []map[string]any{
								{"url": "https://example.com/reel_thumb.jpg"},
							},
						},
					},
				},
			},
			"paging_info": map[string]any{
				"max_id":         "reel_cursor_1",
				"more_available": true,
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// POST /api/v1/clips/reels_tab_feed_items/
	mux.HandleFunc("/api/v1/clips/reels_tab_feed_items/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"items": []map[string]any{
				{
					"media": map[string]any{
						"id":            "feed_reel_222",
						"code":          "feed_reel_code_1",
						"media_type":    2,
						"caption":       map[string]any{"text": "Feed reel caption"},
						"taken_at":      int64(1700010000),
						"like_count":    int64(500),
						"comment_count": int64(30),
						"play_count":    int64(5000),
						"image_versions2": map[string]any{
							"candidates": []map[string]any{
								{"url": "https://example.com/feed_reel.jpg"},
							},
						},
					},
				},
			},
			"paging_info": map[string]any{
				"max_id":         "",
				"more_available": false,
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// newFullMockServer creates an httptest server with all Instagram mock handlers.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withProfileMock(mux)
	withMediaMock(mux)
	withStoriesMock(mux)
	withReelsMock(mux)
	return httptest.NewServer(mux)
}

// newTestClientFactory returns a ClientFactory that creates an Instagram Client
// backed by the given httptest server, bypassing real auth entirely.
func newTestClientFactory(server *httptest.Server) ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		session := &auth.InstagramSession{
			SessionID: "test-session-id",
			CSRFToken: "test-csrf-token",
			DSUserID:  "42544748138",
			UserAgent: "test-agent/1.0",
		}
		return newClientWithBase(session, server.Client(), server.URL), nil
	}
}

// captureStdout runs f with os.Stdout redirected to a pipe and returns the output.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	out, _ := io.ReadAll(r)
	return string(out)
}

// newTestRootCmd creates a root command with global flags for testing.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().Bool("dry-run", false, "")
	return root
}

// buildTestProfileCmd creates a `profile` subcommand tree for use in tests.
func buildTestProfileCmd(factory ClientFactory) *cobra.Command {
	profileCmd := &cobra.Command{Use: "profile", Aliases: []string{"prof"}}
	profileCmd.AddCommand(newProfileGetCmd(factory))
	profileCmd.AddCommand(newProfileEditFormCmd(factory))
	return profileCmd
}

// buildTestMediaCmd creates a `media` subcommand tree for use in tests.
func buildTestMediaCmd(factory ClientFactory) *cobra.Command {
	return newMediaCmd(factory)
}

// buildTestStoriesCmd creates a `stories` subcommand tree for use in tests.
func buildTestStoriesCmd(factory ClientFactory) *cobra.Command {
	return newStoriesCmd(factory)
}

// buildTestReelsCmd creates a `reels` subcommand tree for use in tests.
func buildTestReelsCmd(factory ClientFactory) *cobra.Command {
	return newReelsCmd(factory)
}

// runCmd is a test helper that executes a cobra command tree with args and returns stdout.
func runCmd(t *testing.T, root *cobra.Command, args ...string) string {
	t.Helper()
	return captureStdout(t, func() {
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			t.Fatalf("command failed: %v", err)
		}
	})
}

// runCmdErr executes a cobra command tree and returns any error (does not fatal).
func runCmdErr(t *testing.T, root *cobra.Command, args ...string) error {
	t.Helper()
	root.SetArgs(args)
	// Silence usage on error so test output is clean
	root.SilenceUsage = true
	root.SilenceErrors = true
	return root.Execute()
}

// mustContain asserts that output contains substr.
func mustContain(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("expected output to contain %q, got:\n%s", substr, output)
	}
}
