package linkedin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/emdash-projects/agents/internal/auth"
	"github.com/spf13/cobra"
)

// newTestSession returns a minimal LinkedInSession for use in tests.
func newTestSession() *auth.LinkedInSession {
	return &auth.LinkedInSession{
		LiAt:       "test-li-at",
		JSESSIONID: "ajax:test-jsessionid",
		UserAgent:  "TestAgent/1.0",
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

// withProfileMock registers profile-related mock handlers on mux.
func withProfileMock(mux *http.ServeMux) {
	// GET /voyager/api/identity/profiles/{publicId}[/skills|/skillEndorsements/...]
	mux.HandleFunc("/voyager/api/identity/profiles/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/voyager/api/identity/profiles/")

		// Route sub-paths for skills endpoints.
		if strings.Contains(path, "/skillEndorsements/") {
			// GET /voyager/api/identity/profiles/{publicId}/skillEndorsements/{skillId}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"elements": [
					{
						"entityUrn": "urn:li:endorsement:555",
						"endorser": {
							"miniProfile": {
								"firstName": "Alice",
								"lastName": "Smith",
								"publicIdentifier": "alice-smith"
							}
						}
					}
				],
				"paging": {"start": 0, "count": 10, "total": 1}
			}`))
			return
		}
		if strings.HasSuffix(path, "/skills") {
			// GET /voyager/api/identity/profiles/{publicId}/skills
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"elements": [
					{
						"entityUrn": "urn:li:fs_skill:123",
						"name": "Go",
						"endorsementCount": 42
					},
					{
						"entityUrn": "urn:li:fs_skill:456",
						"name": "Kubernetes",
						"endorsementCount": 18
					}
				],
				"paging": {"start": 0, "count": 50, "total": 2}
			}`))
			return
		}

		publicID := strings.TrimSuffix(path, "/")
		if publicID == "" {
			http.Error(w, `{"message":"missing publicId"}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"profile": {
				"entityUrn": "urn:li:fs_profile:ACoAABtest123",
				"firstName": "Test",
				"lastName": "User",
				"headline": "Software Engineer at TestCorp",
				"summary": "A test user profile.",
				"locationName": "San Francisco, CA",
				"industryName": "Computer Software",
				"profilePicture": {
					"displayImageReference": {
						"vectorImage": {
							"rootUrl": "https://example.com/pic/",
							"artifacts": [{"fileIdentifyingUrlPathSegment": "200x200.jpg"}]
						}
					}
				}
			},
			"connectionCount": 500,
			"followerCount": 1200
		}`))
	})

	// GET /voyager/api/me (current user profile)
	mux.HandleFunc("/voyager/api/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"miniProfile": {
				"entityUrn": "urn:li:fs_miniProfile:ACoAABtest123",
				"firstName": "Mark",
				"lastName": "Test",
				"occupation": "Software Engineer",
				"publicIdentifier": "marktest"
			}
		}`))
	})
}

// withConnectionsMock registers connections-related mock handlers on mux.
func withConnectionsMock(mux *http.ServeMux) {
	// GET /voyager/api/relationships/dash/connections
	mux.HandleFunc("/voyager/api/relationships/dash/connections", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			// Remove connection action
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"elements": [
				{
					"connectedMember": "urn:li:fs_miniProfile:ACoAAConn1",
					"connectedMemberResolved": {
						"entityUrn": "urn:li:fs_miniProfile:ACoAAConn1",
						"firstName": "Alice",
						"lastName": "Smith",
						"occupation": "Product Manager at Acme",
						"publicIdentifier": "alice-smith"
					},
					"createdAt": 1704067200000
				},
				{
					"connectedMember": "urn:li:fs_miniProfile:ACoAAConn2",
					"connectedMemberResolved": {
						"entityUrn": "urn:li:fs_miniProfile:ACoAAConn2",
						"firstName": "Bob",
						"lastName": "Jones",
						"occupation": "Engineer at Widgets Inc",
						"publicIdentifier": "bob-jones"
					},
					"createdAt": 1703980800000
				}
			],
			"paging": {"start": 0, "count": 10, "total": 2}
		}`))
	})
}

// withInvitationsMock registers invitation-related mock handlers on mux.
func withInvitationsMock(mux *http.ServeMux) {
	// GET /voyager/api/relationships/invitationViews (received)
	mux.HandleFunc("/voyager/api/relationships/invitationViews", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"elements": [
				{
					"invitation": {
						"invitationId": "123456",
						"sharedSecret": "secret1",
						"sentTime": 1704067200000,
						"message": "Hi, let's connect",
						"inviterResolved": {
							"entityUrn": "urn:li:fs_miniProfile:ACoAAInviter1",
							"firstName": "Carol",
							"lastName": "White"
						}
					}
				}
			],
			"paging": {"start": 0, "count": 10, "total": 1}
		}`))
	})

	// GET /voyager/api/relationships/sentInvitationViewsV2 (sent)
	// DELETE /voyager/api/relationships/sentInvitationViewsV2/{id} (withdraw)
	mux.HandleFunc("/voyager/api/relationships/sentInvitationViewsV2/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
			return
		}
		http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/voyager/api/relationships/sentInvitationViewsV2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"elements": [
				{
					"invitation": {
						"invitationId": "789012",
						"sharedSecret": "secret2",
						"sentTime": 1703980800000,
						"message": "Looking to connect",
						"inviterResolved": {
							"entityUrn": "urn:li:fs_miniProfile:ACoAAMe",
							"firstName": "Mark",
							"lastName": "Test"
						}
					}
				}
			],
			"paging": {"start": 0, "count": 10, "total": 1}
		}`))
	})

	// POST /voyager/api/relationships/invitation (send)
	mux.HandleFunc("/voyager/api/relationships/invitation", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{}`))
	})

	// PUT /voyager/api/relationships/invitations/{id} (accept/reject)
	mux.HandleFunc("/voyager/api/relationships/invitations/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})
}

// withPostsMock registers posts-related mock handlers on mux.
func withPostsMock(mux *http.ServeMux) {
	// GET /voyager/api/identity/profileUpdatesV2 (list posts)
	mux.HandleFunc("/voyager/api/identity/profileUpdatesV2", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"elements": [
				{
					"updateMetadata": {"urn": "urn:li:activity:1001"},
					"actor": {
						"name": {"text": "Test User"},
						"urn": "urn:li:fs_miniProfile:ACoAABtest123"
					},
					"commentary": {"text": {"text": "Hello LinkedIn world!"}},
					"socialDetail": {
						"totalSocialActivityCounts": {
							"numLikes": 42,
							"numComments": 5,
							"numShares": 2
						}
					},
					"createdAt": 1704067200000
				}
			],
			"paging": {"start": 0, "count": 10, "total": 1}
		}`))
	})

	// GET/DELETE /voyager/api/feed/updates/{urn}
	mux.HandleFunc("/voyager/api/feed/updates/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"entityUrn": "urn:li:activity:1001",
			"actor": {
				"name": {"text": "Test User"},
				"urn": "urn:li:fs_miniProfile:ACoAABtest123"
			},
			"commentary": {"text": {"text": "Hello LinkedIn world!"}},
			"socialDetail": {
				"totalSocialActivityCounts": {
					"numLikes": 42,
					"numComments": 5,
					"numShares": 2
				}
			},
			"createdAt": 1704067200000
		}`))
	})

	// POST /voyager/api/contentCreation/normalizedContent (create post)
	mux.HandleFunc("/voyager/api/contentCreation/normalizedContent", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"value": {"entityUrn": "urn:li:activity:9999"}}`))
	})

	// GET /voyager/api/socialActions/{postUrn}/reactions
	// POST /voyager/api/socialActions/{postUrn}/reactions
	mux.HandleFunc("/voyager/api/socialActions/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		switch {
		case strings.HasSuffix(path, "/reactions") && r.Method == "GET":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"elements": [
					{
						"reactionType": "LIKE",
						"actor": {
							"name": {"text": "Alice Smith"},
							"urn": "urn:li:fs_miniProfile:ACoAAConn1"
						}
					}
				],
				"paging": {"start": 0, "count": 10, "total": 1}
			}`))

		case strings.HasSuffix(path, "/reactions") && r.Method == "POST":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{}`))

		case strings.HasSuffix(path, "/comments") && r.Method == "GET":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"elements": [
					{
						"urn": "urn:li:comment:(activity:1001,5001)",
						"commenter": {
							"com.linkedin.voyager.feed.MemberActor": {
								"miniProfile": {
									"firstName": "Jane",
									"lastName": "Doe",
									"entityUrn": "urn:li:fs_miniProfile:ACoAAJane"
								}
							}
						},
						"comment": {"values": [{"value": "Great post!"}]},
						"socialDetail": {
							"totalSocialActivityCounts": {"numLikes": 3}
						},
						"createdAt": 1704070800000
					}
				],
				"paging": {"start": 0, "count": 10, "total": 1}
			}`))

		case strings.HasSuffix(path, "/comments") && r.Method == "POST":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"value": {"urn": "urn:li:comment:(activity:1001,6001)"}}`))

		case strings.HasSuffix(path, "/likes") && r.Method == "POST":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{}`))

		case strings.HasSuffix(path, "/likes") && r.Method == "DELETE":
			w.WriteHeader(http.StatusNoContent)

		case strings.HasSuffix(path, "/impressions") && r.Method == "GET":
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"impressionCount": 500, "uniqueImpressionsCount": 300}`))

		default:
			// DELETE on a comment URN itself
			if r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
		}
	})
}

// withCommentsMock registers comments-related mock handlers on mux.
// Comments share the /voyager/api/socialActions/ prefix with posts reactions.
// All handlers are registered via withPostsMock; this function is a no-op
// kept for symmetry so callers can call withCommentsMock explicitly.
func withCommentsMock(_ *http.ServeMux) {
	// Intentionally empty: comment endpoints share the /voyager/api/socialActions/
	// handler registered in withPostsMock.
}

// withFeedMock registers feed-related mock handlers on mux.
func withFeedMock(mux *http.ServeMux) {
	// GET /voyager/api/feed/dash/feedUpdates
	mux.HandleFunc("/voyager/api/feed/dash/feedUpdates", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"elements": [
				{
					"updateMetadata": {"urn": "urn:li:activity:2001"},
					"actor": {
						"name": {"text": "Feed Author"},
						"urn": "urn:li:fs_miniProfile:ACoAAFeed1"
					},
					"commentary": {"text": {"text": "Interesting feed post"}},
					"socialDetail": {
						"totalSocialActivityCounts": {
							"numLikes": 15,
							"numComments": 3,
							"numShares": 1
						}
					},
					"createdAt": 1704153600000
				}
			],
			"paging": {"start": 0, "count": 10, "total": 1}
		}`))
	})
}

// withMessagesMock registers messaging-related mock handlers on mux.
func withMessagesMock(mux *http.ServeMux) {
	// GET/POST /voyager/api/messaging/conversations — list and create
	mux.HandleFunc("/voyager/api/messaging/conversations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"elements": [
					{
						"entityUrn": "urn:li:fs_conversation:conv1",
						"conversationId": "conv1",
						"lastActivityAt": 1704067200000,
						"unreadCount": 2,
						"participants": [
							{"com.linkedin.voyager.messaging.MessagingMember": {"miniProfile": {"entityUrn":"urn:li:fs_miniProfile:ABC","firstName":"Jane","lastName":"Doe"}}}
						]
					}
				],
				"paging": {"start": 0, "count": 20, "total": 1}
			}`))
		case http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"conversationId":"new-conv-1"}`))
		default:
			http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	// /voyager/api/messaging/conversations/{id}[/events]
	mux.HandleFunc("/voyager/api/messaging/conversations/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/voyager/api/messaging/conversations/")

		// /conversations/{id}/events
		if strings.HasSuffix(path, "/events") {
			switch r.Method {
			case http.MethodGet:
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{
					"elements": [
						{
							"entityUrn": "urn:li:fs_event:msg1",
							"from": {"com.linkedin.voyager.messaging.MessagingMember": {"miniProfile": {"entityUrn":"urn:li:fs_miniProfile:ABC","firstName":"Jane","lastName":"Doe"}}},
							"eventContent": {"com.linkedin.voyager.messaging.event.MessageEvent": {"body": "Hello!"}},
							"createdAt": 1704067200000
						}
					],
					"paging": {"start": 0, "count": 20, "total": 1}
				}`))
			case http.MethodPost:
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(`{"entityUrn":"urn:li:fs_event:msg2"}`))
			default:
				http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
			}
			return
		}

		// /conversations/{id} — delete or patch
		switch r.Method {
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPatch:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{}`))
		default:
			http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})
}

// withCompaniesMock registers company-related mock handlers on mux.
func withCompaniesMock(mux *http.ServeMux) {
	// GET /voyager/api/organization/companies
	mux.HandleFunc("/voyager/api/organization/companies", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"entityUrn": "urn:li:fs_normalized_company:1234",
			"name": "TestCorp",
			"industryName": "Computer Software",
			"staffCount": 500,
			"followerCount": 10000,
			"description": "A test company"
		}`))
	})

	// POST /voyager/api/feed/follows — follow company
	mux.HandleFunc("/voyager/api/feed/follows", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})

	// DELETE /voyager/api/feed/follows/... — unfollow company
	mux.HandleFunc("/voyager/api/feed/follows/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// GET /voyager/api/jobs/jobPostings (company jobs)
	mux.HandleFunc("/voyager/api/jobs/jobPostings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"elements": [
				{
					"entityUrn": "urn:li:fs_normalized_jobPosting:3456",
					"title": "Senior Software Engineer",
					"companyName": "TestCorp",
					"formattedLocation": "San Francisco, CA",
					"listedAt": 1704067200000,
					"workRemoteAllowed": true
				}
			],
			"paging": {"start": 0, "count": 10, "total": 1}
		}`))
	})

	// GET /voyager/api/search/dash/clusters — company/employee search
	mux.HandleFunc("/voyager/api/search/dash/clusters", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"elements": [
				{
					"elements": [
						{
							"item": {
								"com.linkedin.voyager.search.SearchEntityResult": {
									"entityUrn": "urn:li:fs_normalized_company:1234",
									"title": {"text": "TestCorp"},
									"primarySubtitle": {"text": "Computer Software"}
								}
							}
						}
					]
				}
			]
		}`))
	})
}

// withJobsMock registers job-related mock handlers on mux.
func withJobsMock(mux *http.ServeMux) {
	// GET /voyager/api/jobs/jobPostings/{id}
	mux.HandleFunc("/voyager/api/jobs/jobPostings/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"entityUrn": "urn:li:fs_normalized_jobPosting:3456",
			"title": "Senior Software Engineer",
			"companyName": "TestCorp",
			"formattedLocation": "San Francisco, CA",
			"listedAt": 1704067200000,
			"workRemoteAllowed": true,
			"description": {"text": "An exciting role."}
		}`))
	})

	// GET/POST /voyager/api/jobs/savedJobs
	mux.HandleFunc("/voyager/api/jobs/savedJobs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"elements": [
					{
						"entityUrn": "urn:li:fs_savedJob:99",
						"jobPosting": {
							"entityUrn": "urn:li:fs_normalized_jobPosting:3456",
							"title": "Senior Software Engineer",
							"companyName": "TestCorp",
							"formattedLocation": "San Francisco, CA",
							"listedAt": 1704067200000,
							"workRemoteAllowed": true
						}
					}
				],
				"paging": {"start": 0, "count": 20, "total": 1}
			}`))
		case http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{}`))
		default:
			http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
		}
	})

	// DELETE /voyager/api/jobs/savedJobs/{id}
	mux.HandleFunc("/voyager/api/jobs/savedJobs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// GET /voyager/api/jobs/jobRecommendations
	mux.HandleFunc("/voyager/api/jobs/jobRecommendations", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"elements": [
				{
					"entityUrn": "urn:li:fs_normalized_jobPosting:7890",
					"title": "Staff Engineer",
					"companyName": "BigCo",
					"formattedLocation": "New York, NY",
					"listedAt": 1704067200000,
					"workRemoteAllowed": false
				}
			],
			"paging": {"start": 0, "count": 20, "total": 1}
		}`))
	})
}

// withAnalyticsMock registers analytics-related mock handlers on mux.
// Post impressions are handled by /voyager/api/socialActions/ in withPostsMock.
func withAnalyticsMock(mux *http.ServeMux) {
	// GET /voyager/api/identity/wvmpCards (profile views)
	mux.HandleFunc("/voyager/api/identity/wvmpCards", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"viewsCount": 42, "timePeriod": "7 days"}`))
	})

	// GET /voyager/api/identity/searchAppearances
	mux.HandleFunc("/voyager/api/identity/searchAppearances", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"searchAppearanceCount": 15, "timePeriod": "7 days"}`))
	})
}

// withSettingsMock registers settings-related mock handlers on mux.
func withSettingsMock(mux *http.ServeMux) {
	// GET/POST /voyager/api/identity/profileSettings
	mux.HandleFunc("/voyager/api/identity/profileSettings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"profileVisibility": "PUBLIC",
			"messagingPreference": "CONNECTIONS",
			"activeStatus": true
		}`))
	})

	// GET /voyager/api/identity/privacySettings
	mux.HandleFunc("/voyager/api/identity/privacySettings", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"profileVisibility": "PUBLIC",
			"connectionsVisibility": "ALL",
			"lastNameVisibility": "ALL",
			"profilePhotoVisibility": "PUBLIC"
		}`))
	})
}

// withSearchMock registers search-related mock handlers on mux.
// NOTE: /voyager/api/search/dash/clusters is already registered by withCompaniesMock.
// This function is a no-op kept for symmetry so callers can call withSearchMock explicitly.
func withSearchMock(_ *http.ServeMux) {
	// Intentionally empty: the /voyager/api/search/dash/clusters handler is
	// registered by withCompaniesMock which is called earlier in newFullMockServer.
}

// withNetworkMock registers network (followers/following/suggestions) mock handlers on mux.
func withNetworkMock(mux *http.ServeMux) {
	// GET /voyager/api/relationships/dash/followers
	mux.HandleFunc("/voyager/api/relationships/dash/followers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"elements": [
				{
					"entityUrn": "urn:li:fs_miniProfile:ACoAAFollower1",
					"firstName": "Follower",
					"lastName": "One",
					"occupation": "Engineer at Acme",
					"publicIdentifier": "follower-one"
				}
			],
			"paging": {"start": 0, "count": 10, "total": 1}
		}`))
	})

	// GET /voyager/api/relationships/dash/following
	mux.HandleFunc("/voyager/api/relationships/dash/following", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"elements": [
				{
					"entityUrn": "urn:li:fs_miniProfile:ACoAAFollowing1",
					"firstName": "Following",
					"lastName": "One",
					"occupation": "Designer at Corp",
					"publicIdentifier": "following-one"
				}
			],
			"paging": {"start": 0, "count": 10, "total": 1}
		}`))
	})

	// GET /voyager/api/relationships/dash/connectionsYouMayKnow
	mux.HandleFunc("/voyager/api/relationships/dash/connectionsYouMayKnow", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"elements": [
				{
					"entityUrn": "urn:li:fs_miniProfile:ACoAASuggested1",
					"firstName": "Suggested",
					"lastName": "Person",
					"occupation": "Manager at BigCo",
					"publicIdentifier": "suggested-person"
				}
			],
			"paging": {"start": 0, "count": 10, "total": 1}
		}`))
	})
}

// withNotificationsMock registers notification-related mock handlers on mux.
func withNotificationsMock(mux *http.ServeMux) {
	// GET /voyager/api/identity/notifications
	mux.HandleFunc("/voyager/api/identity/notifications", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"elements": [
				{
					"entityUrn": "urn:li:notification:111",
					"headline": {"text": "Alice Smith viewed your profile"},
					"publishedAt": 1704067200000,
					"read": false
				},
				{
					"entityUrn": "urn:li:notification:222",
					"headline": {"text": "Bob Jones liked your post"},
					"publishedAt": 1703980800000,
					"read": true
				}
			],
			"paging": {"start": 0, "count": 20, "total": 2}
		}`))
	})

	// POST /voyager/api/identity/notifications/markAllAsRead
	mux.HandleFunc("/voyager/api/identity/notifications/markAllAsRead", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, `{"message":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})
}

// newFullMockServer creates a test server with all LinkedIn mock handlers.
func newFullMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	withProfileMock(mux)
	withConnectionsMock(mux)
	withInvitationsMock(mux)
	withPostsMock(mux)
	withCommentsMock(mux)
	withFeedMock(mux)
	withMessagesMock(mux)
	withCompaniesMock(mux)
	withJobsMock(mux)
	withAnalyticsMock(mux)
	withSettingsMock(mux)
	withSearchMock(mux)
	withNetworkMock(mux)
	withNotificationsMock(mux)
	return httptest.NewServer(mux)
}
