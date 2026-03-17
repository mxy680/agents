package instagram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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

		// comments list: GET /api/v1/media/{id}/comments/
		case action == "comments":
			resp := map[string]any{
				"comments": []map[string]any{
					{
						"pk":                 "comment_111",
						"text":               "Great post!",
						"created_at":         int64(1700000100),
						"comment_like_count": int64(5),
						"user": map[string]any{
							"pk":       "user_abc",
							"username": "commenter1",
						},
					},
				},
				"next_max_id": "cmt_cursor_1",
				"status":      "ok",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		// comments replies: GET /api/v1/media/{id}/comments/{comment_id}/inline_child_comments/
		case strings.HasPrefix(action, "comments/") && strings.HasSuffix(action, "/inline_child_comments"):
			resp := map[string]any{
				"comments": []map[string]any{
					{
						"pk":                 "reply_222",
						"text":               "Agreed!",
						"created_at":         int64(1700000200),
						"comment_like_count": int64(1),
						"user": map[string]any{
							"pk":       "user_def",
							"username": "replier1",
						},
					},
				},
				"status": "ok",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		// comment create: POST /api/v1/media/{id}/comment/
		case action == "comment":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})

		// comment delete: POST /api/v1/media/{id}/comment/{comment_id}/delete/
		case strings.HasPrefix(action, "comment/") && strings.HasSuffix(action, "/delete"):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})

		// comment like/unlike: POST /api/v1/media/{comment_id}/comment_like/ or comment_unlike/
		case action == "comment_like" || action == "comment_unlike":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})

		// disable/enable comments on a post
		case action == "disable_comments" || action == "enable_comments":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})

		// like/unlike a post
		case action == "like" || action == "unlike":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})

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

// withDirectMock registers all direct-message-related mock handlers on mux.
func withDirectMock(mux *http.ServeMux) {
	// GET /api/v1/direct_v2/inbox/
	mux.HandleFunc("/api/v1/direct_v2/inbox/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"inbox": map[string]any{
				"threads": []map[string]any{
					{
						"thread_id":        "thread_111",
						"thread_title":     "Test Thread",
						"last_activity_at": int64(1700000000000000),
						"is_group":         false,
						"users": []map[string]any{
							{"pk": "user_999", "username": "dm_user", "full_name": "DM User"},
						},
						"items": []map[string]any{
							{
								"item_id":   "item_aaa",
								"item_type": "text",
								"text":      "Hello there",
								"timestamp": int64(1700000000000000),
								"user_id":   "user_999",
							},
						},
					},
				},
				"oldest_cursor": "inbox_cursor_1",
				"has_older":     true,
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /api/v1/direct_v2/pending_inbox/
	mux.HandleFunc("/api/v1/direct_v2/pending_inbox/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"inbox": map[string]any{
				"threads": []map[string]any{
					{
						"thread_id":    "pending_thread_222",
						"thread_title": "Pending Thread",
						"is_group":     false,
					},
				},
				"oldest_cursor": "",
				"has_older":     false,
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /api/v1/direct_v2/threads/{thread_id}/
	// POST /api/v1/direct_v2/threads/{thread_id}/approve/
	// POST /api/v1/direct_v2/threads/{thread_id}/hide/
	// POST /api/v1/direct_v2/threads/{thread_id}/items/{item_id}/delete/
	// POST /api/v1/direct_v2/threads/{thread_id}/items/{item_id}/seen/
	mux.HandleFunc("/api/v1/direct_v2/threads/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/direct_v2/threads/")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) < 2 {
			// GET thread detail: /api/v1/direct_v2/threads/{thread_id}/
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"thread": map[string]any{
					"thread_id":        parts[0],
					"thread_title":     "Test Thread",
					"last_activity_at": int64(1700000000000000),
					"is_group":         false,
					"items": []map[string]any{
						{
							"item_id":   "item_aaa",
							"item_type": "text",
							"text":      "Hello there",
							"timestamp": int64(1700000000000000),
							"user_id":   "user_999",
						},
					},
					"oldest_cursor": "",
					"has_older":     false,
				},
				"status": "ok",
			})
			return
		}
		threadID := parts[0]
		action := strings.TrimSuffix(parts[1], "/")
		_ = threadID

		switch {
		case action == "":
			// thread detail response
			resp := map[string]any{
				"thread": map[string]any{
					"thread_id":        threadID,
					"thread_title":     "Test Thread",
					"last_activity_at": int64(1700000000000000),
					"is_group":         false,
					"items": []map[string]any{
						{
							"item_id":   "item_aaa",
							"item_type": "text",
							"text":      "Hello there",
							"timestamp": int64(1700000000000000),
							"user_id":   "user_999",
						},
					},
					"oldest_cursor": "",
					"has_older":     false,
				},
				"status": "ok",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		case action == "approve" || action == "hide":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
		case strings.HasPrefix(action, "items/") && strings.HasSuffix(action, "/delete"):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
		case strings.HasPrefix(action, "items/") && strings.HasSuffix(action, "/seen"):
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
		case action == "text":
			// broadcast/text: threadID="broadcast", action="text"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
		default:
			http.NotFound(w, r)
		}
	})

	// POST /api/v1/direct_v2/create_group_thread/
	mux.HandleFunc("/api/v1/direct_v2/create_group_thread/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})
}

// withLikesMock registers likes feed mock handler on mux.
// Note: like/unlike on /api/v1/media/{id}/like|unlike/ are handled by withMediaMock.
func withLikesMock(mux *http.ServeMux) {
	// GET /api/v1/feed/liked/
	mux.HandleFunc("/api/v1/feed/liked/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"items": []map[string]any{
				{
					"id":            "liked_post_111",
					"code":          "liked_code_1",
					"media_type":    1,
					"caption":       map[string]any{"text": "Liked post caption"},
					"taken_at":      int64(1700005000),
					"like_count":    int64(300),
					"comment_count": int64(20),
					"image_versions2": map[string]any{
						"candidates": []map[string]any{
							{"url": "https://example.com/liked.jpg"},
						},
					},
				},
			},
			"next_max_id":    "liked_cursor_1",
			"more_available": true,
			"status":         "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withRelationshipsMock registers relationships-related mock handlers on mux.
func withRelationshipsMock(mux *http.ServeMux) {
	// GET /api/v1/friendships/{user_id}/followers/
	// GET /api/v1/friendships/{user_id}/following/
	// GET /api/v1/friendships/show/{user_id}/
	// POST /api/v1/friendships/create/{user_id}/
	// POST /api/v1/friendships/destroy/{user_id}/
	// POST /api/v1/friendships/remove_follower/{user_id}/
	// POST /api/v1/friendships/block/{user_id}/
	// POST /api/v1/friendships/unblock/{user_id}/
	// POST /api/v1/friendships/mute_posts_or_story_from_follow/
	// POST /api/v1/friendships/unmute_posts_or_story_from_follow/
	mux.HandleFunc("/api/v1/friendships/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/friendships/")
		path = strings.TrimSuffix(path, "/")
		parts := strings.SplitN(path, "/", 2)

		action := parts[0]

		switch {
		// Fixed-path POST actions (no user_id in path segment 0)
		case action == "mute_posts_or_story_from_follow" || action == "unmute_posts_or_story_from_follow":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"friendship_status": map[string]any{"following": true},
				"status":            "ok",
			})
		// Actions where path is /action/user_id/  e.g. /show/123/, /create/123/
		case action == "show" || action == "create" || action == "destroy" ||
			action == "remove_follower" || action == "block" || action == "unblock":
			if len(parts) < 2 {
				http.NotFound(w, r)
				return
			}
			switch action {
			case "show":
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"friendship_status": map[string]any{
						"following":        true,
						"followed_by":      false,
						"blocking":         false,
						"muting":           false,
						"is_private":       false,
						"incoming_request": false,
						"outgoing_request": false,
						"is_restricted":    false,
					},
					"status": "ok",
				})
			default:
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{
					"friendship_status": map[string]any{"following": true},
					"status":            "ok",
				})
			}
		// Actions where path is /user_id/followers/ or /user_id/following/
		default:
			// action is actually the user_id here; parts[1] is "followers" or "following"
			if len(parts) < 2 {
				http.NotFound(w, r)
				return
			}
			subAction := parts[1]
			switch subAction {
			case "followers", "following":
				resp := map[string]any{
					"users": []map[string]any{
						{
							"pk":              "rel_user_111",
							"username":        "rel_user",
							"full_name":       "Rel User",
							"profile_pic_url": "https://example.com/rel.jpg",
							"is_private":      false,
							"is_verified":     false,
						},
					},
					"next_max_id": "",
					"big_list":    false,
					"status":      "ok",
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			default:
				http.NotFound(w, r)
			}
		}
	})

	// GET /api/v1/users/blocked_list/
	// Note: /api/v1/users/ is already registered by withProfileMock for user info.
	// We need a more specific handler — register on the exact path.
	mux.HandleFunc("/api/v1/users/blocked_list/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"blocked_list": []map[string]any{
				{
					"pk":              "blocked_user_111",
					"username":        "blocked_user",
					"full_name":       "Blocked User",
					"profile_pic_url": "https://example.com/blocked.jpg",
					"is_private":      true,
					"is_verified":     false,
				},
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// POST /api/v1/restrict_action/restrict/
	// POST /api/v1/restrict_action/unrestrict/
	mux.HandleFunc("/api/v1/restrict_action/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"friendship_status": map[string]any{"is_restricted": true},
			"status":            "ok",
		})
	})
}

// withSearchMock registers search-related mock handlers on mux.
func withSearchMock(mux *http.ServeMux) {
	// GET /api/v1/users/search/
	mux.HandleFunc("/api/v1/users/search/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"users": []map[string]any{
				{
					"pk":              "search_user_111",
					"username":        "search_result",
					"full_name":       "Search Result User",
					"profile_pic_url": "https://example.com/search.jpg",
					"is_private":      false,
					"is_verified":     false,
				},
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /api/v1/tags/search/
	mux.HandleFunc("/api/v1/tags/search/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"results": []map[string]any{
				{
					"id":              "tag_111",
					"name":            "golang",
					"media_count":     int64(50000),
					"following_count": int64(1200),
					"following":       false,
				},
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /api/v1/location_search/
	mux.HandleFunc("/api/v1/location_search/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"venues": []map[string]any{
				{
					"pk":          int64(999111),
					"name":        "Test Location",
					"address":     "123 Main St",
					"city":        "Test City",
					"lat":         37.7749,
					"lng":         -122.4194,
					"media_count": int64(500),
				},
			},
			"status": "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	// GET /api/v1/fbsearch/topsearch_flat/
	// POST /api/v1/fbsearch/clear_search_history/
	mux.HandleFunc("/api/v1/fbsearch/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/fbsearch/")
		path = strings.TrimSuffix(path, "/")
		switch path {
		case "topsearch_flat":
			// Real API uses the key "list", not "ranked_list".
			resp := map[string]any{
				"list": []map[string]any{
					{
						"position": 1,
						"type":     "user",
						"user": map[string]any{
							"pk":       12190648480,
							"username": "top_result",
						},
					},
				},
				"status": "ok",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		case "clear_search_history":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
		default:
			http.NotFound(w, r)
		}
	})

	// GET /api/v1/discover/explore/
	mux.HandleFunc("/api/v1/discover/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"items":          []map[string]any{{"id": "explore_item_1"}},
			"next_max_id":    "explore_cursor",
			"more_available": false,
			"status":         "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withCollectionsMock registers collections-related mock handlers on mux.
func withCollectionsMock(mux *http.ServeMux) {
	// GET /api/v1/collections/list/
	// POST /api/v1/collections/create/
	// POST /api/v1/collections/{id}/edit/
	// POST /api/v1/collections/{id}/delete/
	mux.HandleFunc("/api/v1/collections/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/collections/")
		path = strings.TrimSuffix(path, "/")

		switch path {
		case "list":
			resp := map[string]any{
				"items": []map[string]any{
					{
						"collection_id":   "col_111",
						"collection_name": "My Saves",
						"collection_type": "MEDIA",
						"media_count":     int64(10),
					},
				},
				"next_max_id":    "",
				"more_available": false,
				"status":         "ok",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		case "create":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"collection": map[string]any{
					"collection_id":   "col_new",
					"collection_name": "New Collection",
					"collection_type": "MEDIA",
					"media_count":     int64(0),
				},
				"status": "ok",
			})
		default:
			// /api/v1/collections/{id}/edit/ or /api/v1/collections/{id}/delete/
			parts := strings.SplitN(path, "/", 2)
			if len(parts) == 2 && (parts[1] == "edit" || parts[1] == "delete") {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
			} else {
				http.NotFound(w, r)
			}
		}
	})

	// GET /api/v1/feed/collection/{id}/
	// GET /api/v1/feed/saved/posts/
	// (both under /api/v1/feed/ — handled via the feed handler)
	mux.HandleFunc("/api/v1/feed/collection/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"items": []map[string]any{
				{
					"id":            "saved_media_111",
					"code":          "saved_code_1",
					"media_type":    1,
					"caption":       map[string]any{"text": "Saved post"},
					"taken_at":      int64(1700000000),
					"like_count":    int64(50),
					"comment_count": int64(3),
					"image_versions2": map[string]any{
						"candidates": []map[string]any{
							{"url": "https://example.com/saved.jpg"},
						},
					},
				},
			},
			"next_max_id":    "",
			"more_available": false,
			"status":         "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/api/v1/feed/saved/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"items": []map[string]any{
				{
					"media": map[string]any{
						"id":            "saved_post_222",
						"code":          "saved_code_2",
						"media_type":    1,
						"caption":       map[string]any{"text": "All saves post"},
						"taken_at":      int64(1700001000),
						"like_count":    int64(75),
						"comment_count": int64(5),
						"image_versions2": map[string]any{
							"candidates": []map[string]any{
								{"url": "https://example.com/allsaved.jpg"},
							},
						},
					},
				},
			},
			"next_max_id":    "",
			"more_available": false,
			"status":         "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

// withTagsMock registers hashtag-related mock handlers on mux.
func withTagsMock(mux *http.ServeMux) {
	// GET /api/v1/tags/{name}/info/
	// GET /api/v1/tags/{name}/sections/
	// POST /api/v1/tags/follow/{name}/
	// POST /api/v1/tags/unfollow/{name}/
	// GET /api/v1/tags/{name}/related/
	mux.HandleFunc("/api/v1/tags/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/tags/")
		path = strings.TrimSuffix(path, "/")
		parts := strings.SplitN(path, "/", 2)

		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}

		first := parts[0]
		second := parts[1]

		// follow/{tag_name} and unfollow/{tag_name}
		if first == "follow" || first == "unfollow" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"result": "following", "status": "ok"})
			return
		}

		// search/ handled by the more-specific /api/v1/tags/search/ handler.
		// Here: {tag_name}/{action}
		tagName := first
		action := second
		switch action {
		case "info":
			// Real API returns tag fields at the top level (flat), not wrapped in a "tag" key.
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"id":              "tag_111",
				"name":            tagName,
				"media_count":     int64(50000),
				"following_count": int64(1200),
				"following":       false,
				"status":          "ok",
			})
		case "sections":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"sections": []map[string]any{
					{
						"feed_type": "media",
						"layout_content": map[string]any{
							"medias": []map[string]any{
								{
									"media": map[string]any{
										"id":            "tag_media_111",
										"code":          "tag_code_1",
										"media_type":    1,
										"caption":       map[string]any{"text": "Tag post"},
										"taken_at":      int64(1700000000),
										"like_count":    int64(100),
										"comment_count": int64(8),
										"image_versions2": map[string]any{
											"candidates": []map[string]any{
												{"url": "https://example.com/tag.jpg"},
											},
										},
									},
								},
							},
						},
					},
				},
				"next_max_id":    "",
				"more_available": false,
				"status":         "ok",
			})
		case "related":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"related": []map[string]any{
					{
						"type": "hashtag",
						"tag": map[string]any{
							"id":              "related_tag_222",
							"name":            "related" + tagName,
							"media_count":     int64(10000),
							"following_count": int64(500),
							"following":       false,
						},
					},
				},
				"status": "ok",
			})
		default:
			http.NotFound(w, r)
		}
	})

	// GET /api/v1/users/self/following_tag_list/
	mux.HandleFunc("/api/v1/users/self/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/users/self/")
		path = strings.TrimSuffix(path, "/")
		if path == "following_tag_list" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"tags": []map[string]any{
					{
						"id":              "followed_tag_111",
						"name":            "photography",
						"media_count":     int64(200000),
						"following_count": int64(50000),
						"following":       true,
					},
				},
				"status": "ok",
			})
			return
		}
		http.NotFound(w, r)
	})
}

// withLocationsMock registers location-related mock handlers on mux.
func withLocationsMock(mux *http.ServeMux) {
	// GET /api/v1/locations/{id}/info/
	// GET /api/v1/locations/{id}/sections/
	// GET /api/v1/locations/{id}/story/
	mux.HandleFunc("/api/v1/locations/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/locations/")
		path = strings.TrimSuffix(path, "/")
		parts := strings.SplitN(path, "/", 2)

		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}

		locationID := parts[0]
		action := parts[1]

		switch action {
		case "info":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"location": map[string]any{
					"pk":          int64(999111),
					"name":        "Test Location " + locationID,
					"address":     "123 Main St",
					"city":        "Test City",
					"lat":         37.7749,
					"lng":         -122.4194,
					"media_count": int64(500),
				},
				"status": "ok",
			})
		case "sections":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"sections": []map[string]any{
					{
						"feed_type": "media",
						"layout_content": map[string]any{
							"medias": []map[string]any{
								{
									"media": map[string]any{
										"id":            "loc_media_111",
										"code":          "loc_code_1",
										"media_type":    1,
										"caption":       map[string]any{"text": "Location post"},
										"taken_at":      int64(1700000000),
										"like_count":    int64(200),
										"comment_count": int64(15),
										"image_versions2": map[string]any{
											"candidates": []map[string]any{
												{"url": "https://example.com/loc.jpg"},
											},
										},
									},
								},
							},
						},
					},
				},
				"next_max_id":    "",
				"more_available": false,
				"status":         "ok",
			})
		case "story":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"story":  map[string]any{"id": "loc_story_111"},
				"status": "ok",
			})
		default:
			http.NotFound(w, r)
		}
	})
}

// withActivityMock registers activity/notifications mock handlers on mux.
func withActivityMock(mux *http.ServeMux) {
	// GET /api/v1/news/inbox/
	// POST /api/v1/news/inbox_seen/
	mux.HandleFunc("/api/v1/news/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/news/")
		path = strings.TrimSuffix(path, "/")

		switch path {
		case "inbox":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"new_stories": []map[string]any{
					{
						"pk":   "notif_111",
						"type": 1,
						"args": map[string]any{
							"timestamp":    int64(1700000000),
							"text":         "liked your photo",
							"profile_id":   "notif_user_111",
							"profile_name": "notif_user",
						},
					},
				},
				"old_stories": []map[string]any{},
				"status":      "ok",
			})
		case "inbox_seen":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
		default:
			http.NotFound(w, r)
		}
	})
}

// withLiveMock registers live broadcast mock handlers on mux.
func withLiveMock(mux *http.ServeMux) {
	// GET /api/v1/live/reels_tray_broadcasts/
	// GET /api/v1/live/{id}/info/
	// GET /api/v1/live/{id}/get_comment/
	// POST /api/v1/live/{id}/heartbeat_and_get_viewer_count/
	// POST /api/v1/live/{id}/like/
	// POST /api/v1/live/{id}/comment/
	mux.HandleFunc("/api/v1/live/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/live/")
		path = strings.TrimSuffix(path, "/")

		if path == "reels_tray_broadcasts" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"broadcasts": []map[string]any{
					{
						"id":               "broadcast_111",
						"broadcast_status": "active",
						"cover_frame_url":  "https://example.com/live.jpg",
						"viewer_count":     int64(500),
						"published_time":   int64(1700000000),
					},
				},
				"status": "ok",
			})
			return
		}

		parts := strings.SplitN(path, "/", 2)
		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}
		broadcastID := parts[0]
		action := parts[1]
		_ = broadcastID

		switch action {
		case "info":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"broadcast": map[string]any{
					"id":               broadcastID,
					"broadcast_status": "active",
					"cover_frame_url":  "https://example.com/live.jpg",
					"viewer_count":     int64(500),
					"published_time":   int64(1700000000),
				},
				"status": "ok",
			})
		case "get_comment":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"comments": []map[string]any{
					{
						"pk":         "live_comment_111",
						"text":       "Hello!",
						"created_at": 1700000001.0,
						"user": map[string]any{
							"pk":       "live_commenter_111",
							"username": "live_commenter",
						},
					},
				},
				"status": "ok",
			})
		case "heartbeat_and_get_viewer_count":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"viewer_count": int64(505), "status": "ok"})
		case "like":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"like_ts": int64(1700000002), "status": "ok"})
		case "comment":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"comment": map[string]any{
					"pk":         "live_new_comment",
					"text":       "New comment",
					"created_at": 1700000003.0,
					"user": map[string]any{
						"pk":       "42544748138",
						"username": "testuser",
					},
				},
				"status": "ok",
			})
		default:
			http.NotFound(w, r)
		}
	})
}

// withHighlightsMock registers story highlights mock handlers on mux.
func withHighlightsMock(mux *http.ServeMux) {
	// GET /api/v1/highlights/{user_id}/highlights_tray/
	// POST /api/v1/highlights/create_reel/
	// POST /api/v1/highlights/{id}/edit_reel/
	// POST /api/v1/highlights/{id}/delete_reel/
	mux.HandleFunc("/api/v1/highlights/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/highlights/")
		path = strings.TrimSuffix(path, "/")
		parts := strings.SplitN(path, "/", 2)

		if len(parts) == 1 && parts[0] == "create_reel" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"reel": map[string]any{
					"id":          "highlight_new",
					"title":       "New Highlight",
					"media_count": 2,
					"created_at":  int64(1700000000),
				},
				"status": "ok",
			})
			return
		}

		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}

		id := parts[0]
		action := parts[1]

		switch action {
		case "highlights_tray":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"tray": []map[string]any{
					{
						"id":          "highlight:hl_111",
						"title":       "Travel",
						"media_count": 5,
						"created_at":  int64(1700000000),
					},
				},
				"status": "ok",
			})
		case "edit_reel":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"reel": map[string]any{
					"id":          id,
					"title":       "Edited",
					"media_count": 3,
					"created_at":  int64(1700000000),
				},
				"status": "ok",
			})
		case "delete_reel":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
		default:
			http.NotFound(w, r)
		}
	})

	// GET /api/v1/feed/reels_media/
	mux.HandleFunc("/api/v1/feed/reels_media/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"reels":  map[string]any{"highlight:hl_111": map[string]any{"id": "highlight:hl_111", "items": []map[string]any{}}},
			"status": "ok",
		})
	})
}

// withCloseFriendsMock registers close friends mock handlers on mux.
// Note: /api/v1/friendships/ is already registered; besties is handled by extending it.
// We use a more specific path registration.
func withCloseFriendsMock(mux *http.ServeMux) {
	// GET /api/v1/friendships/besties/
	// POST /api/v1/friendships/set_besties/
	// These are already under /api/v1/friendships/ — we extend withRelationshipsMock
	// to handle these paths. Since we cannot re-register, we handle them in
	// a separate specific handler registered before the generic one.
	// Registration order matters: more-specific paths are matched first by http.ServeMux.
	mux.HandleFunc("/api/v1/friendships/besties/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"users": []map[string]any{
				{
					"pk":              "bestie_user_111",
					"username":        "bestie_user",
					"full_name":       "Bestie User",
					"profile_pic_url": "https://example.com/bestie.jpg",
					"is_private":      false,
					"is_verified":     false,
				},
			},
			"next_max_id": "",
			"big_list":    false,
			"status":      "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/api/v1/friendships/set_besties/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})
}

// withSettingsMock registers account settings mock handlers on mux.
func withSettingsMock(mux *http.ServeMux) {
	// GET /api/v1/accounts/current_user/
	// POST /api/v1/accounts/set_private/
	// POST /api/v1/accounts/set_public/
	mux.HandleFunc("/api/v1/accounts/current_user/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"user": map[string]any{
				"pk":              "42544748138",
				"username":        "testuser",
				"full_name":       "Test User",
				"is_private":      false,
				"is_verified":     false,
				"email":           "test@example.com",
				"phone_number":    "+15551234567",
				"biography":       "My bio",
				"external_url":    "https://example.com",
				"profile_pic_url": "https://example.com/pic.jpg",
			},
			"status": "ok",
		})
	})

	mux.HandleFunc("/api/v1/accounts/set_private/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	mux.HandleFunc("/api/v1/accounts/set_public/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	})

	// GET /api/v1/session/login_activity/
	mux.HandleFunc("/api/v1/session/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"login_activity": []map[string]any{
				{
					"device_name":  "iPhone 14",
					"location":     "San Francisco, CA",
					"login_time":   int64(1700000000),
				},
			},
			"status": "ok",
		})
	})
}

// withGraphQLMock registers handlers for the Instagram web GraphQL endpoints.
// It handles both /graphql/query (web frontend) and /api/graphql (Polaris DM).
func withGraphQLMock(mux *http.ServeMux) {
	// POST /graphql/query — form-encoded GraphQL used by web frontend.
	// Dispatches by doc_id to the appropriate mock response.
	mux.HandleFunc("/graphql/query", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		vals, _ := url.ParseQuery(string(body))
		docID := vals.Get("doc_id")

		w.Header().Set("Content-Type", "application/json")
		switch docID {
		case "26914912424764761": // PolarisPostChildCommentsQuery
			resp := map[string]any{
				"data": map[string]any{
					"xdt_api__v1__media__media_id__comments__parent_comment_id__child_comments__connection": map[string]any{
						"edges": []map[string]any{
							{
								"node": map[string]any{
									"pk":                 "comment_gql_111",
									"text":               "GraphQL comment!",
									"created_at_utc":     int64(1700000100),
									"comment_like_count": int64(3),
									"has_liked_comment":  false,
									"user": map[string]any{
										"pk":       "user_gql_abc",
										"username": "gql_commenter",
									},
								},
							},
						},
					},
				},
				"status": "ok",
			}
			json.NewEncoder(w).Encode(resp)

		case "8845758582119845": // PolarisPostActionLoadPostQueryQuery (media detail with comments)
			resp := map[string]any{
				"data": map[string]any{
					"xdt_shortcode_media": map[string]any{
						"edge_media_to_parent_comment": map[string]any{
							"count": 100,
							"edges": []map[string]any{
								{
									"node": map[string]any{
										"id":         "comment_detail_111",
										"text":       "Comment from media detail!",
										"created_at": int64(1700000100),
										"owner": map[string]any{
											"id":       "user_detail_abc",
											"username": "detail_commenter",
										},
										"edge_liked_by": map[string]any{"count": int64(5)},
									},
								},
							},
						},
					},
				},
				"status": "ok",
			}
			json.NewEncoder(w).Encode(resp)

		case "26136666099278270": // PolarisClipsTabDesktopPaginationQuery (reels feed)
			resp := map[string]any{
				"data": map[string]any{
					"xdt_api__v1__clips__home__connection": map[string]any{
						"edges": []map[string]any{
							{
								"node": map[string]any{
									"media": map[string]any{
										"pk":            "feed_reel_222",
										"code":          "feed_reel_gql_code",
										"taken_at":      int64(1700010000),
										"like_count":    int64(500),
										"comment_count": int64(30),
										"play_count":    int64(5000),
										"caption":       map[string]any{"text": "Feed reel caption"},
										"image_versions2": map[string]any{
											"candidates": []map[string]any{
												{"url": "https://example.com/feed_reel.jpg"},
											},
										},
									},
								},
							},
						},
						"page_info": map[string]any{
							"has_next_page": false,
							"end_cursor":    "",
						},
					},
				},
				"status": "ok",
			}
			json.NewEncoder(w).Encode(resp)

		default:
			http.Error(w, `{"errors":[{"message":"unknown doc_id"}]}`, http.StatusBadRequest)
		}
	})

	// POST /api/graphql — used by PolarisDirectInboxQuery for DM inbox fallback.
	mux.HandleFunc("/api/graphql", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, _ := io.ReadAll(r.Body)
		vals, _ := url.ParseQuery(string(body))
		docID := vals.Get("doc_id")

		w.Header().Set("Content-Type", "application/json")
		switch docID {
		case "34623053607292942": // PolarisDirectInboxQuery
			resp := map[string]any{
				"data": map[string]any{
					"get_slide_mailbox_for_iris_subscription": map[string]any{
						"threads_by_folder": map[string]any{
							"edges": []map[string]any{
								{
									"node": map[string]any{
										"thread_id":    "thread_gql_111",
										"thread_title": "GraphQL Thread",
										"is_group":     false,
										"updated_at":   int64(1700000000000000),
									},
								},
							},
						},
					},
				},
				"status": "ok",
			}
			json.NewEncoder(w).Encode(resp)

		default:
			http.Error(w, `{"errors":[{"message":"unknown doc_id"}]}`, http.StatusBadRequest)
		}
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
	withDirectMock(mux)
	withLikesMock(mux)
	// Register close friends before relationships so specific paths match first.
	withCloseFriendsMock(mux)
	withRelationshipsMock(mux)
	withSearchMock(mux)
	withCollectionsMock(mux)
	withTagsMock(mux)
	withLocationsMock(mux)
	withActivityMock(mux)
	withLiveMock(mux)
	withHighlightsMock(mux)
	withSettingsMock(mux)
	withGraphQLMock(mux)
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

// buildTestDirectCmd creates a `direct` subcommand tree for use in tests.
func buildTestDirectCmd(factory ClientFactory) *cobra.Command {
	return newDirectCmd(factory)
}

// buildTestCommentsCmd creates a `comments` subcommand tree for use in tests.
func buildTestCommentsCmd(factory ClientFactory) *cobra.Command {
	return newCommentsCmd(factory)
}

// buildTestLikesCmd creates a `likes` subcommand tree for use in tests.
func buildTestLikesCmd(factory ClientFactory) *cobra.Command {
	return newLikesCmd(factory)
}

// buildTestRelationshipsCmd creates a `relationships` subcommand tree for use in tests.
func buildTestRelationshipsCmd(factory ClientFactory) *cobra.Command {
	return newRelationshipsCmd(factory)
}

// buildTestSearchCmd creates a `search` subcommand tree for use in tests.
func buildTestSearchCmd(factory ClientFactory) *cobra.Command {
	return newSearchCmd(factory)
}

// buildTestCollectionsCmd creates a `collections` subcommand tree for use in tests.
func buildTestCollectionsCmd(factory ClientFactory) *cobra.Command {
	return newCollectionsCmd(factory)
}

// buildTestTagsCmd creates a `tags` subcommand tree for use in tests.
func buildTestTagsCmd(factory ClientFactory) *cobra.Command {
	return newTagsCmd(factory)
}

// buildTestLocationsCmd creates a `locations` subcommand tree for use in tests.
func buildTestLocationsCmd(factory ClientFactory) *cobra.Command {
	return newLocationsCmd(factory)
}

// buildTestActivityCmd creates an `activity` subcommand tree for use in tests.
func buildTestActivityCmd(factory ClientFactory) *cobra.Command {
	return newActivityCmd(factory)
}

// buildTestLiveCmd creates a `live` subcommand tree for use in tests.
func buildTestLiveCmd(factory ClientFactory) *cobra.Command {
	return newLiveCmd(factory)
}

// buildTestHighlightsCmd creates a `highlights` subcommand tree for use in tests.
func buildTestHighlightsCmd(factory ClientFactory) *cobra.Command {
	return newHighlightsCmd(factory)
}

// buildTestCloseFriendsCmd creates a `closefriends` subcommand tree for use in tests.
func buildTestCloseFriendsCmd(factory ClientFactory) *cobra.Command {
	return newCloseFriendsCmd(factory)
}

// buildTestSettingsCmd creates a `settings` subcommand tree for use in tests.
func buildTestSettingsCmd(factory ClientFactory) *cobra.Command {
	return newSettingsCmd(factory)
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
