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

// withDMMock registers mock handlers for DM v1.1 and GraphQL mutation endpoints.
func withDMMock(mux *http.ServeMux) {
	// dm/inbox_initial_state.json — dm inbox
	mux.HandleFunc("/i/api/1.1/dm/inbox_initial_state.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		raw, _ := json.Marshal(map[string]any{
			"inbox_initial_state": map[string]any{
				"conversations": map[string]any{
					"conv-abc-123": map[string]any{
						"conversation_id": "conv-abc-123",
						"type":            "ONE_TO_ONE",
						"participants": []any{
							map[string]any{"user_id": "111"},
							map[string]any{"user_id": "222"},
						},
					},
				},
				"entries": []any{
					map[string]any{
						"message": map[string]any{
							"id":              "msg-001",
							"conversation_id": "conv-abc-123",
							"message_data": map[string]any{
								"text":      "Hello there!",
								"sender_id": "111",
								"time":      "1710000000000",
							},
						},
					},
				},
			},
		})
		w.Write(raw)
	})

	// dm/conversation/{id}.json — dm conversation
	mux.HandleFunc("/i/api/1.1/dm/conversation/", func(w http.ResponseWriter, r *http.Request) {
		// Handle both conversation fetch and rename-group (update_name).
		if strings.HasSuffix(r.URL.Path, "/update_name.json") {
			w.Header().Set("Content-Type", "application/json")
			raw, _ := json.Marshal(map[string]any{"conversation_id": "conv-abc-123", "name": "New Group Name"})
			w.Write(raw)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		raw, _ := json.Marshal(map[string]any{
			"conversation_timeline": map[string]any{
				"entries": []any{
					map[string]any{
						"message": map[string]any{
							"id":              "msg-001",
							"conversation_id": "conv-abc-123",
							"message_data": map[string]any{
								"text":      "Hello there!",
								"sender_id": "111",
								"time":      "1710000000000",
							},
						},
					},
				},
				"min_entry_id": "msg-000",
			},
		})
		w.Write(raw)
	})

	// dm/new2.json — dm send and dm send-group
	mux.HandleFunc("/i/api/1.1/dm/new2.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		raw, _ := json.Marshal(map[string]any{
			"event": map[string]any{
				"id":   "msg-new-001",
				"type": "message_create",
			},
		})
		w.Write(raw)
	})

	// DMMessageDeleteMutation — dm delete
	mux.HandleFunc("/i/api/graphql/"+hashDMMessageDelete+"/DMMessageDeleteMutation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"dm_message_delete": map[string]any{}}))
	})

	// useDMReactionMutationAddMutation — dm react
	mux.HandleFunc("/i/api/graphql/"+hashDMReactionAdd+"/useDMReactionMutationAddMutation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"dm_reaction_add": map[string]any{}}))
	})

	// useDMReactionMutationRemoveMutation — dm unreact
	mux.HandleFunc("/i/api/graphql/"+hashDMReactionRemove+"/useDMReactionMutationRemoveMutation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"dm_reaction_remove": map[string]any{}}))
	})

	// AddParticipantsMutation — dm add-members
	mux.HandleFunc("/i/api/graphql/"+hashDMAddParticipants+"/AddParticipantsMutation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"add_participants": map[string]any{}}))
	})
}

// mockListResult returns a mock list GraphQL result object.
func mockListResult(id, name string) json.RawMessage {
	raw, _ := json.Marshal(map[string]any{
		"id_str":           id,
		"name":             name,
		"description":      "A test list",
		"member_count":     42,
		"subscriber_count": 10,
		"mode":             "Public",
		"created_at":       1710000000,
		"user_results": map[string]any{
			"result": map[string]any{
				"legacy": map[string]any{
					"screen_name": "listowner",
					"name":        "List Owner",
				},
			},
		},
	})
	return raw
}

// mockListTimelineResponse builds a GraphQL timeline response containing one list entry and a cursor.
func mockListTimelineResponse(listRaw json.RawMessage, cursor string) []byte {
	entries := []any{
		map[string]any{
			"entryId":   "list-" + string(listRaw[:10]),
			"sortIndex": "9999",
			"content": map[string]any{
				"entryType": "TimelineTimelineItem",
				"itemContent": map[string]any{
					"itemType": "TimelineList",
					"list_results": map[string]any{
						"result": listRaw,
					},
				},
			},
		},
	}
	if cursor != "" {
		entries = append(entries, map[string]any{
			"entryId":   "cursor-bottom",
			"sortIndex": "0",
			"content": map[string]any{
				"entryType":  "TimelineTimelineCursor",
				"value":      cursor,
				"cursorType": "Bottom",
			},
		})
	}

	resp := map[string]any{
		"data": map[string]any{
			"lists_management_page_timeline": map[string]any{
				"timeline": map[string]any{
					"instructions": []any{
						map[string]any{
							"type":    "TimelineAddEntries",
							"entries": entries,
						},
					},
				},
			},
		},
	}
	raw, _ := json.Marshal(resp)
	return raw
}

// withListsMock registers mock handlers for list GraphQL queries and mutations.
func withListsMock(mux *http.ServeMux) {
	listRaw := mockListResult("list-001", "My Test List")
	userRaw := mockUserResult("111", "listmember", "List Member")
	tweetRaw := mockTweetResult("123456789", "List tweet!", "999", "Test User", "testuser")

	// ListByRestId — lists get
	mux.HandleFunc("/i/api/graphql/"+hashListByRestId+"/ListByRestId", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"list": listRaw}))
	})

	// ListsManagementPageTimeline — lists owned
	mux.HandleFunc("/i/api/graphql/"+hashListsManagementPageTimeline+"/ListsManagementPageTimeline", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockListTimelineResponse(listRaw, "lists-cursor-abc"))
	})

	// SearchTimeline for lists — lists search (reuses the existing hashSearchTimeline handler
	// already registered by withPostsMock; the mock returns a timeline response that may not
	// contain list entries, so extractListTimeline will return empty — that is acceptable for tests).

	// ListLatestTweetsTimeline — lists tweets
	mux.HandleFunc("/i/api/graphql/"+hashListLatestTweetsTimeline+"/ListLatestTweetsTimeline", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, "list-tweets-cursor"))
	})

	// ListMembers — lists members
	mux.HandleFunc("/i/api/graphql/"+hashListMembers+"/ListMembers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockUserListResponse(userRaw))
	})

	// ListSubscribers — lists subscribers
	mux.HandleFunc("/i/api/graphql/"+hashListSubscribers+"/ListSubscribers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockUserListResponse(userRaw))
	})

	// CreateList — lists create
	mux.HandleFunc("/i/api/graphql/"+hashCreateList+"/CreateList", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"create_list": map[string]any{
				"list": listRaw,
			},
		}))
	})

	// UpdateList — lists update
	mux.HandleFunc("/i/api/graphql/"+hashUpdateList+"/UpdateList", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"update_list": map[string]any{}}))
	})

	// DeleteList — lists delete
	mux.HandleFunc("/i/api/graphql/"+hashDeleteList+"/DeleteList", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"delete_list": map[string]any{}}))
	})

	// ListAddMember — lists add-member
	mux.HandleFunc("/i/api/graphql/"+hashListAddMember+"/ListAddMember", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"list_add_member": map[string]any{}}))
	})

	// ListRemoveMember — lists remove-member
	mux.HandleFunc("/i/api/graphql/"+hashListRemoveMember+"/ListRemoveMember", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"list_remove_member": map[string]any{}}))
	})

	// EditListBanner — lists set-banner
	mux.HandleFunc("/i/api/graphql/"+hashEditListBanner+"/EditListBanner", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"edit_list_banner": map[string]any{}}))
	})

	// DeleteListBanner — lists remove-banner
	mux.HandleFunc("/i/api/graphql/"+hashDeleteListBanner+"/DeleteListBanner", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"delete_list_banner": map[string]any{}}))
	})
}

// withCommunitiesMock registers mock handlers for community GraphQL operations.
func withCommunitiesMock(mux *http.ServeMux) {
	tweetRaw := mockTweetResult("123456789", "Community tweet!", "999", "Test User", "testuser")
	userRaw := mockUserResult("111", "member1", "Member One")

	communityData := map[string]any{
		"id_str":       "111",
		"name":         "Test Community",
		"description":  "A test community",
		"member_count": 42,
		"created_at":   "2024-01-01",
	}

	// CommunityQuery — communities get
	mux.HandleFunc("/i/api/graphql/"+hashCommunityQuery+"/CommunityQuery", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"community_by_id": communityData}))
	})

	// CommunitiesSearchQuery — communities search
	mux.HandleFunc("/i/api/graphql/"+hashCommunitiesSearchQuery+"/CommunitiesSearchQuery", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"communities_search_by_query": []any{communityData}}))
	})

	// CommunityTweetsTimeline — communities tweets
	mux.HandleFunc("/i/api/graphql/"+hashCommunityTweetsTimeline+"/CommunityTweetsTimeline", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, "community-tweets-cursor"))
	})

	// CommunityMediaTimeline — communities media
	mux.HandleFunc("/i/api/graphql/"+hashCommunityMediaTimeline+"/CommunityMediaTimeline", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, "community-media-cursor"))
	})

	// membersSliceTimeline_Query — communities members
	mux.HandleFunc("/i/api/graphql/"+hashMembersSliceTimeline+"/membersSliceTimeline_Query", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockUserListResponse(userRaw))
	})

	// moderatorsSliceTimeline_Query — communities moderators
	mux.HandleFunc("/i/api/graphql/"+hashModeratorsSliceTimeline+"/moderatorsSliceTimeline_Query", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockUserListResponse(userRaw))
	})

	// CommunitiesMainPageTimeline — communities timeline
	mux.HandleFunc("/i/api/graphql/"+hashCommunitiesMainPageTimeline+"/CommunitiesMainPageTimeline", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, "communities-main-cursor"))
	})

	// JoinCommunity — communities join
	mux.HandleFunc("/i/api/graphql/"+hashJoinCommunity+"/JoinCommunity", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"join_community": map[string]any{}}))
	})

	// LeaveCommunity — communities leave
	mux.HandleFunc("/i/api/graphql/"+hashLeaveCommunity+"/LeaveCommunity", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"leave_community": map[string]any{}}))
	})

	// RequestToJoinCommunity — communities request-join
	mux.HandleFunc("/i/api/graphql/"+hashRequestToJoinCommunity+"/RequestToJoinCommunity", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{"request_to_join_community": map[string]any{}}))
	})

	// CommunityTweetSearchModuleQuery — communities search-tweets
	mux.HandleFunc("/i/api/graphql/"+hashCommunityTweetSearch+"/CommunityTweetSearchModuleQuery", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(mockTimelineResponse(tweetRaw, "community-search-cursor"))
	})
}

// withNotificationsMock registers mock handlers for notification v2 endpoints.
func withNotificationsMock(mux *http.ServeMux) {
	notifResponse := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		raw, _ := json.Marshal(map[string]any{
			"globalObjects": map[string]any{
				"notifications": map[string]any{
					"notif-001": map[string]any{
						"type":        "like",
						"timestampMs": "1700000000000",
						"message": map[string]any{
							"text": "Someone liked your tweet",
						},
					},
				},
				"tweets": map[string]any{},
				"users":  map[string]any{},
			},
			"timeline": map[string]any{
				"instructions": []any{},
			},
		})
		w.Write(raw)
	}

	mux.HandleFunc("/i/api/2/notifications/all.json", notifResponse)
	mux.HandleFunc("/i/api/2/notifications/mentions.json", notifResponse)
	mux.HandleFunc("/i/api/2/notifications/verified.json", notifResponse)
}

// withMediaMock registers mock handlers for media upload and metadata endpoints.
func withMediaMock(mux *http.ServeMux) {
	// Upload INIT, APPEND, FINALIZE, STATUS — /i/media/upload.json (handled by the test client's base URL)
	mux.HandleFunc("/i/media/upload.json", func(w http.ResponseWriter, r *http.Request) {
		// APPEND is multipart; try FormValue first (works for url-encoded) then fallback.
		cmd := r.FormValue("command")
		switch cmd {
		case "INIT":
			w.Header().Set("Content-Type", "application/json")
			raw, _ := json.Marshal(map[string]any{
				"media_id":        999,
				"media_id_string": "999",
			})
			w.Write(raw)
		case "APPEND":
			// APPEND returns 204 No Content.
			w.WriteHeader(http.StatusNoContent)
		case "FINALIZE":
			w.Header().Set("Content-Type", "application/json")
			raw, _ := json.Marshal(map[string]any{
				"media_id":        999,
				"media_id_string": "999",
				"processing_info": map[string]any{"state": "succeeded"},
			})
			w.Write(raw)
		case "STATUS":
			w.Header().Set("Content-Type", "application/json")
			raw, _ := json.Marshal(map[string]any{
				"media_id":        999,
				"media_id_string": "999",
				"processing_info": map[string]any{"state": "succeeded"},
			})
			w.Write(raw)
		default:
			// Unknown command or multipart without readable command field — treat as APPEND (204).
			w.WriteHeader(http.StatusNoContent)
		}
	})

	// Alt text — api.x.com/1.1/media/metadata/create.json
	mux.HandleFunc("/1.1/media/metadata/create.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	})
}

// withScheduledMock registers mock handlers for scheduled tweet GraphQL operations.
func withScheduledMock(mux *http.ServeMux) {
	scheduledItem := map[string]any{
		"rest_id":          "sched-123",
		"scheduled_status": "scheduled",
		"execute_at":       1800000000,
		"tweet_create_request": map[string]any{
			"status": "My scheduled tweet",
		},
	}

	// FetchScheduledTweets — scheduled list
	mux.HandleFunc("/i/api/graphql/"+hashFetchScheduledTweets+"/FetchScheduledTweets", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"scheduled_tweet_list": []any{scheduledItem},
		}))
	})

	// CreateScheduledTweet — scheduled create
	mux.HandleFunc("/i/api/graphql/"+hashCreateScheduledTweet+"/CreateScheduledTweet", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"create_scheduled_tweet": scheduledItem,
		}))
	})

	// DeleteScheduledTweet — scheduled delete
	mux.HandleFunc("/i/api/graphql/"+hashDeleteScheduledTweet+"/DeleteScheduledTweet", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(graphqlResponse(map[string]any{
			"delete_scheduled_tweet": map[string]any{},
		}))
	})
}

// withTrendsMock registers mock handlers for trends v2 and v1.1 endpoints.
func withTrendsMock(mux *http.ServeMux) {
	// /i/api/2/guide.json — trends list
	mux.HandleFunc("/i/api/2/guide.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		raw, _ := json.Marshal(map[string]any{
			"timeline": map[string]any{
				"instructions": []any{
					map[string]any{
						"type": "TimelineAddEntries",
						"entries": []any{
							map[string]any{
								"entryId": "trends-module",
								"content": map[string]any{
									"timelineModule": map[string]any{
										"items": []any{
											map[string]any{
												"item": map[string]any{
													"content": map[string]any{
														"trend": map[string]any{
															"name":     "#GoLang",
															"trendUrl": "https://x.com/search?q=%23GoLang",
															"domainContext": map[string]any{
																"entityCount": 5000,
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
					},
				},
			},
		})
		w.Write(raw)
	})

	// /i/api/1.1/trends/available.json — trends locations
	mux.HandleFunc("/i/api/1.1/trends/available.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		raw, _ := json.Marshal([]any{
			map[string]any{
				"name":    "Worldwide",
				"woeid":   1,
				"country": "",
			},
			map[string]any{
				"name":    "United States",
				"woeid":   23424977,
				"country": "United States",
			},
		})
		w.Write(raw)
	})

	// /i/api/1.1/trends/place.json — trends by-place
	mux.HandleFunc("/i/api/1.1/trends/place.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		tweetVolume := 12345
		raw, _ := json.Marshal([]any{
			map[string]any{
				"trends": []any{
					map[string]any{
						"name":         "#Testing",
						"url":          "https://x.com/search?q=%23Testing",
						"tweet_volume": tweetVolume,
					},
				},
			},
		})
		w.Write(raw)
	})
}

// withPollsMock registers mock handlers for polls (caps.x.com routed through base URL).
func withPollsMock(mux *http.ServeMux) {
	// Polls use caps.x.com full URLs, handled directly by the client.
	// In tests, the base URL is the test server, so register under the path portion.
	mux.HandleFunc("/v2/cards/create.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		raw, _ := json.Marshal(map[string]any{
			"card_uri": "card://1234567890",
		})
		w.Write(raw)
	})

	mux.HandleFunc("/v2/capi/passthrough/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		raw, _ := json.Marshal(map[string]any{
			"card": map[string]any{
				"name": "poll2choice_text_only",
			},
		})
		w.Write(raw)
	})
}

// withGeoMock registers mock handlers for geo v1.1 endpoints.
func withGeoMock(mux *http.ServeMux) {
	placeRaw := map[string]any{
		"id":           "place-001",
		"name":         "New York",
		"full_name":    "New York, NY",
		"place_type":   "city",
		"country":      "United States",
		"country_code": "US",
	}

	// /i/api/1.1/geo/reverse_geocode.json
	mux.HandleFunc("/i/api/1.1/geo/reverse_geocode.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		raw, _ := json.Marshal(map[string]any{
			"result": map[string]any{
				"places": []any{placeRaw},
			},
		})
		w.Write(raw)
	})

	// /i/api/1.1/geo/search.json
	mux.HandleFunc("/i/api/1.1/geo/search.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		raw, _ := json.Marshal(map[string]any{
			"result": map[string]any{
				"places": []any{placeRaw},
			},
		})
		w.Write(raw)
	})

	// /i/api/1.1/geo/id/{place_id}.json — catch-all for geo get
	mux.HandleFunc("/i/api/1.1/geo/id/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		raw, _ := json.Marshal(placeRaw)
		w.Write(raw)
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
	withDMMock(mux)
	withListsMock(mux)
	withCommunitiesMock(mux)
	withNotificationsMock(mux)
	withMediaMock(mux)
	withScheduledMock(mux)
	withTrendsMock(mux)
	withPollsMock(mux)
	withGeoMock(mux)
	return httptest.NewServer(mux)
}
