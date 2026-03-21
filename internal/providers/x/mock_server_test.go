package x

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
)

// newTestSession returns a minimal XSession for use in tests.
func newTestSession() *auth.XSession {
	return &auth.XSession{
		AuthToken: "test-auth-token",
		CSRFToken: "test-csrf-token",
		UserAgent: "TestAgent/1.0",
	}
}

// newTestClient creates a Client pointing at the given test server.
func newTestClient(server *httptest.Server) *Client {
	return newClientWithBase(newTestSession(), server.Client(), server.URL)
}

// newTestClientFactory returns a ClientFactory pointing at the given test server.
func newTestClientFactory(server *httptest.Server) ClientFactory {
	return func(ctx context.Context) (*Client, error) {
		return newTestClient(server), nil
	}
}

// newTestRootCmd creates a root command with --json flag wired up.
func newTestRootCmd() *cobra.Command {
	root := &cobra.Command{Use: "integrations"}
	root.PersistentFlags().Bool("json", false, "Output as JSON")
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

// containsStr is a test helper that checks if s contains sub.
func containsStr(s, sub string) bool {
	return strings.Contains(s, sub)
}

// mockTweetResult returns a mock GraphQL tweet_results JSON object.
func mockTweetResult(id, text, authorID, authorName, authorUsername string) json.RawMessage {
	raw, _ := json.Marshal(map[string]any{
		"result": map[string]any{
			"__typename": "Tweet",
			"rest_id":    id,
			"legacy": map[string]any{
				"full_text":       text,
				"favorite_count":  42,
				"retweet_count":   7,
				"reply_count":     3,
				"quote_count":     1,
				"created_at":      "Mon Jan 01 12:00:00 +0000 2024",
				"retweeted_status_id_str": "",
			},
			"core": map[string]any{
				"user_results": map[string]any{
					"result": map[string]any{
						"rest_id": authorID,
						"legacy": map[string]any{
							"screen_name":    authorUsername,
							"name":           authorName,
							"followers_count": 1000,
							"friends_count":  500,
							"statuses_count": 2000,
						},
					},
				},
			},
			"views": map[string]any{
				"count": "1500",
			},
		},
	})
	return raw
}

// mockUserResult returns a mock GraphQL user result JSON object.
func mockUserResult(id, username, name string) json.RawMessage {
	raw, _ := json.Marshal(map[string]any{
		"__typename": "User",
		"rest_id":    id,
		"legacy": map[string]any{
			"screen_name":              username,
			"name":                     name,
			"description":              "Test user bio",
			"location":                 "Test City",
			"verified":                 false,
			"followers_count":          1000,
			"friends_count":            500,
			"statuses_count":           2000,
			"profile_image_url_https":  "https://pbs.twimg.com/profile_images/test/photo.jpg",
			"created_at":               "Mon Jan 01 00:00:00 +0000 2020",
		},
		"is_blue_verified": true,
	})
	return raw
}

// mockTimelineResponse builds a GraphQL timeline response with one tweet entry and a cursor.
func mockTimelineResponse(tweetRaw json.RawMessage, cursor string) []byte {
	resp := map[string]any{
		"data": map[string]any{
			"home_timeline_by_raw_query": map[string]any{
				"timeline": map[string]any{
					"instructions": []any{
						map[string]any{
							"type": "TimelineAddEntries",
							"entries": []any{
								map[string]any{
									"entryId":   "tweet-123456789",
									"sortIndex": "9999",
									"content": map[string]any{
										"entryType": "TimelineTimelineItem",
										"itemContent": map[string]any{
											"itemType":     "TimelineTweet",
											"tweet_results": tweetRaw,
										},
									},
								},
								map[string]any{
									"entryId":   "cursor-bottom-123",
									"sortIndex": "0",
									"content": map[string]any{
										"entryType":  "TimelineTimelineCursor",
										"value":      cursor,
										"cursorType": "Bottom",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	raw, _ := json.Marshal(resp)
	return raw
}

// mockUserListResponse builds a GraphQL response with one user entry.
func mockUserListResponse(userResultRaw json.RawMessage) []byte {
	resp := map[string]any{
		"data": map[string]any{
			"retweeters_timeline": map[string]any{
				"timeline": map[string]any{
					"instructions": []any{
						map[string]any{
							"type": "TimelineAddEntries",
							"entries": []any{
								map[string]any{
									"entryId":   "user-111",
									"sortIndex": "9999",
									"content": map[string]any{
										"entryType": "TimelineTimelineItem",
										"itemContent": map[string]any{
											"itemType": "TimelineUser",
											"user_results": map[string]any{
												"result": json.RawMessage(userResultRaw),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	raw, _ := json.Marshal(resp)
	return raw
}

// graphqlResponse wraps data in a standard GraphQL envelope.
func graphqlResponse(data any) []byte {
	raw, _ := json.Marshal(map[string]any{"data": data})
	return raw
}

// withPostsMock registers mock handlers for tweet/post GraphQL operations.
func withPostsMock(mux *http.ServeMux) {
	tweetRaw := mockTweetResult("123456789", "Hello X world!", "999", "Test User", "testuser")

	// TweetResultByRestId — posts get
	mux.HandleFunc("/i/api/graphql/"+hashTweetResultByRestId+"/TweetResultByRestId", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := map[string]any{
			"tweetResult": tweetRaw,
		}
		w.Write(graphqlResponse(data))
	})

	// TweetResultsByRestIds — posts lookup
	mux.HandleFunc("/i/api/graphql/"+hashTweetResultsByRestIds+"/TweetResultsByRestIds", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := map[string]any{
			"tweetResults": []any{tweetRaw},
		}
		w.Write(graphqlResponse(data))
	})

	// CreateTweet — posts create
	mux.HandleFunc("/i/api/graphql/"+hashCreateTweet+"/CreateTweet", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := map[string]any{
			"create_tweet": map[string]any{
				"tweet_results": tweetRaw,
			},
		}
		w.Write(graphqlResponse(data))
	})

	// DeleteTweet — posts delete
	mux.HandleFunc("/i/api/graphql/"+hashDeleteTweet+"/DeleteTweet", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := map[string]any{
			"delete_tweet": map[string]any{},
		}
		w.Write(graphqlResponse(data))
	})

	// SearchTimeline — posts search (and users search with product=People)
	mux.HandleFunc("/i/api/graphql/"+hashSearchTimeline+"/SearchTimeline", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return timeline format for tweet search; user search will parse via user list result.
		w.Write(mockTimelineResponse(tweetRaw, "next-cursor-abc"))
	})

	// HomeTimeline — posts timeline
	mux.HandleFunc("/i/api/graphql/"+hashHomeTimeline+"/HomeTimeline", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, "timeline-cursor-xyz"))
	})

	// HomeLatestTimeline — posts latest-timeline
	mux.HandleFunc("/i/api/graphql/"+hashHomeLatestTimeline+"/HomeLatestTimeline", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, "latest-cursor-xyz"))
	})

	// UserTweets — posts user-tweets
	mux.HandleFunc("/i/api/graphql/"+hashUserTweets+"/UserTweets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, "user-tweets-cursor"))
	})

	// UserTweetsAndReplies — posts user-replies
	mux.HandleFunc("/i/api/graphql/"+hashUserTweetsAndReplies+"/UserTweetsAndReplies", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, "user-replies-cursor"))
	})

	// Retweeters — posts retweeters
	userRaw := mockUserResult("111", "retweeter1", "Retweeter One")
	mux.HandleFunc("/i/api/graphql/"+hashRetweeters+"/Retweeters", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockUserListResponse(userRaw))
	})

	// Favoriters — posts favoriters
	mux.HandleFunc("/i/api/graphql/"+hashFavoriters+"/Favoriters", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockUserListResponse(userRaw))
	})

	// SimilarPosts — posts similar
	mux.HandleFunc("/i/api/graphql/"+hashSimilarPosts+"/SimilarPosts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, ""))
	})
}

// withUsersMock registers mock handlers for user GraphQL operations.
func withUsersMock(mux *http.ServeMux) {
	userRaw := mockUserResult("999", "testuser", "Test User")

	// UserByScreenName — users get
	mux.HandleFunc("/i/api/graphql/"+hashUserByScreenName+"/UserByScreenName", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := map[string]any{
			"user": map[string]any{
				"result": userRaw,
			},
		}
		w.Write(graphqlResponse(data))
	})

	// UserByRestId — users get-by-id
	mux.HandleFunc("/i/api/graphql/"+hashUserByRestId+"/UserByRestId", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		data := map[string]any{
			"user": map[string]any{
				"result": userRaw,
			},
		}
		w.Write(graphqlResponse(data))
	})

	// UserHighlightsTweets — users highlights
	tweetRaw := mockTweetResult("123456789", "Hello X world!", "999", "Test User", "testuser")
	mux.HandleFunc("/i/api/graphql/"+hashUserHighlightsTweets+"/UserHighlightsTweets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, "highlights-cursor"))
	})

	// UserMedia — users media
	mux.HandleFunc("/i/api/graphql/"+hashUserMedia+"/UserMedia", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, "media-cursor"))
	})

	// UserCreatorSubscriptions — users subscriptions
	mux.HandleFunc("/i/api/graphql/"+hashUserCreatorSubscriptions+"/UserCreatorSubscriptions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockUserListResponse(userRaw))
	})
}

// mockFriendshipResponse returns a minimal v1.1 user object used by friendship endpoints.
func mockFriendshipResponse(userID string) []byte {
	raw, _ := json.Marshal(map[string]any{
		"id_str":          userID,
		"screen_name":     "testuser",
		"name":            "Test User",
		"followers_count": 1000,
		"friends_count":   500,
	})
	return raw
}

// withFollowsMock registers mock handlers for follower/following GraphQL and friendship v1.1 endpoints.
func withFollowsMock(mux *http.ServeMux) {
	userRaw := mockUserResult("111", "follower1", "Follower One")

	// Followers — follows followers
	mux.HandleFunc("/i/api/graphql/"+hashFollowers+"/Followers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockUserListResponse(userRaw))
	})

	// Following — follows following
	mux.HandleFunc("/i/api/graphql/"+hashFollowing+"/Following", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockUserListResponse(userRaw))
	})

	// BlueVerifiedFollowers — follows verified-followers
	mux.HandleFunc("/i/api/graphql/"+hashVerifiedFollowers+"/BlueVerifiedFollowers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockUserListResponse(userRaw))
	})

	// FollowersYouKnow — follows followers-you-know
	mux.HandleFunc("/i/api/graphql/"+hashFollowersYouKnow+"/FollowersYouKnow", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockUserListResponse(userRaw))
	})

	// friendships/create — follows follow
	mux.HandleFunc("/i/api/1.1/friendships/create.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockFriendshipResponse("111"))
	})

	// friendships/destroy — follows unfollow
	mux.HandleFunc("/i/api/1.1/friendships/destroy.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockFriendshipResponse("111"))
	})
}

// withBlocksMock registers mock handlers for block v1.1 endpoints.
func withBlocksMock(mux *http.ServeMux) {
	// blocks/create — blocks block
	mux.HandleFunc("/i/api/1.1/blocks/create.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockFriendshipResponse("222"))
	})

	// blocks/destroy — blocks unblock
	mux.HandleFunc("/i/api/1.1/blocks/destroy.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockFriendshipResponse("222"))
	})
}

// withMutesMock registers mock handlers for mute v1.1 endpoints.
func withMutesMock(mux *http.ServeMux) {
	// mutes/users/create — mutes mute
	mux.HandleFunc("/i/api/1.1/mutes/users/create.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockFriendshipResponse("333"))
	})

	// mutes/users/destroy — mutes unmute
	mux.HandleFunc("/i/api/1.1/mutes/users/destroy.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockFriendshipResponse("333"))
	})
}

// withLikesMock registers mock handlers for likes GraphQL operations.
func withLikesMock(mux *http.ServeMux) {
	tweetRaw := mockTweetResult("123456789", "Hello X world!", "999", "Test User", "testuser")

	// Likes — likes list
	mux.HandleFunc("/i/api/graphql/"+hashLikes+"/Likes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, "likes-cursor-abc"))
	})

	// FavoriteTweet — likes like
	mux.HandleFunc("/i/api/graphql/"+hashFavoriteTweet+"/FavoriteTweet", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"favorite_tweet": map[string]any{
				"tweet_results": tweetRaw,
			},
		}))
	})

	// UnfavoriteTweet — likes unlike
	mux.HandleFunc("/i/api/graphql/"+hashUnfavoriteTweet+"/UnfavoriteTweet", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"unfavorite_tweet": "Done",
		}))
	})
}

// withRetweetsMock registers mock handlers for retweet GraphQL operations.
func withRetweetsMock(mux *http.ServeMux) {
	// CreateRetweet — retweets retweet
	mux.HandleFunc("/i/api/graphql/"+hashCreateRetweet+"/CreateRetweet", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"create_retweet": map[string]any{},
		}))
	})

	// DeleteRetweet — retweets undo
	mux.HandleFunc("/i/api/graphql/"+hashDeleteRetweet+"/DeleteRetweet", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"delete_retweet": map[string]any{},
		}))
	})
}

// withBookmarksMock registers mock handlers for bookmark GraphQL operations.
func withBookmarksMock(mux *http.ServeMux) {
	tweetRaw := mockTweetResult("123456789", "Hello X world!", "999", "Test User", "testuser")

	// Bookmarks — bookmarks list
	mux.HandleFunc("/i/api/graphql/"+hashBookmarks+"/Bookmarks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, "bookmarks-cursor-abc"))
	})

	// CreateBookmark — bookmarks add (no folder)
	mux.HandleFunc("/i/api/graphql/"+hashCreateBookmark+"/CreateBookmark", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"tweet_bookmark_put": "Done",
		}))
	})

	// BookmarkToFolder — bookmarks add (with folder)
	mux.HandleFunc("/i/api/graphql/"+hashBookmarkToFolder+"/BookmarkToFolder", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"tweet_bookmark_put": "Done",
		}))
	})

	// DeleteBookmark — bookmarks remove
	mux.HandleFunc("/i/api/graphql/"+hashDeleteBookmark+"/DeleteBookmark", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"tweet_bookmark_delete": "Done",
		}))
	})

	// BookmarksAllDelete — bookmarks clear
	mux.HandleFunc("/i/api/graphql/"+hashBookmarksAllDelete+"/BookmarksAllDelete", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"bookmarks_all_delete": "Done",
		}))
	})

	// BookmarkFoldersSlice — bookmarks folders
	mux.HandleFunc("/i/api/graphql/"+hashBookmarkFoldersSlice+"/BookmarkFoldersSlice", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"bookmark_collections_slice": []any{
				map[string]any{"id": "folder-001", "name": "My Reading List"},
				map[string]any{"id": "folder-002", "name": "Inspiration"},
			},
		}))
	})

	// BookmarkFolderTimeline — bookmarks folder-tweets
	mux.HandleFunc("/i/api/graphql/"+hashBookmarkFolderTimeline+"/BookmarkFolderTimeline", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, "folder-tweets-cursor"))
	})

	// CreateBookmarkFolder — bookmarks create-folder
	mux.HandleFunc("/i/api/graphql/"+hashCreateBookmarkFolder+"/CreateBookmarkFolder", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"create_bookmark_collection": map[string]any{
				"id":   "folder-new-001",
				"name": "Test Folder",
			},
		}))
	})

	// EditBookmarkFolder — bookmarks edit-folder
	mux.HandleFunc("/i/api/graphql/"+hashEditBookmarkFolder+"/EditBookmarkFolder", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"edit_bookmark_collection": map[string]any{},
		}))
	})

	// DeleteBookmarkFolder — bookmarks delete-folder
	mux.HandleFunc("/i/api/graphql/"+hashDeleteBookmarkFolder+"/DeleteBookmarkFolder", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"delete_bookmark_collection": map[string]any{},
		}))
	})
}

// newFullMockServer creates a test HTTP server with all mock handlers registered.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withPostsMock(mux)
	withUsersMock(mux)
	withFollowsMock(mux)
	withBlocksMock(mux)
	withMutesMock(mux)
	withLikesMock(mux)
	withRetweetsMock(mux)
	withBookmarksMock(mux)
	return httptest.NewServer(mux)
}
