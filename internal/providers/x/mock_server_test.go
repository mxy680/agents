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

// newFullMockServer creates a test HTTP server with all mock handlers registered.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withPostsMock(mux)
	withUsersMock(mux)
	return httptest.NewServer(mux)
}
