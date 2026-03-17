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
	withRelationshipsMock(mux)
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
