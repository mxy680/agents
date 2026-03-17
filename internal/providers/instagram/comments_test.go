package instagram

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCommentsListText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "list", "--media-id=111222333")

	mustContain(t, out, "ID")
	mustContain(t, out, "comment_111")
	mustContain(t, out, "Great post!")
	mustContain(t, out, "commenter1")
}

func TestCommentsListJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "list", "--media-id=111222333", "--json")

	dec := json.NewDecoder(strings.NewReader(out))
	var comments []CommentSummary
	if err := dec.Decode(&comments); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(comments) == 0 {
		t.Fatal("expected at least one comment")
	}
	if comments[0].PK != "comment_111" {
		t.Errorf("expected pk=comment_111, got %s", comments[0].PK)
	}
	if comments[0].Text != "Great post!" {
		t.Errorf("expected text='Great post!', got %s", comments[0].Text)
	}
}

func TestCommentsListMissingFlag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	err := runCmdErr(t, root, "comments", "list")
	if err == nil {
		t.Error("expected error when --media-id not provided")
	}
}

func TestCommentsRepliesText(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "replies", "--media-id=111222333", "--comment-id=comment_111")

	mustContain(t, out, "reply_222")
	mustContain(t, out, "Agreed!")
	mustContain(t, out, "replier1")
}

func TestCommentsRepliesJSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "replies", "--media-id=111222333", "--comment-id=comment_111", "--json")

	var comments []CommentSummary
	if err := json.Unmarshal([]byte(out), &comments); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(comments) == 0 {
		t.Fatal("expected at least one reply")
	}
	if comments[0].PK != "reply_222" {
		t.Errorf("expected pk=reply_222, got %s", comments[0].PK)
	}
}

func TestCommentsCreateDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "create", "--media-id=111222333", "--text=Nice!", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
	mustContain(t, out, "111222333")
}

func TestCommentsCreate(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "create", "--media-id=111222333", "--text=Nice!")

	mustContain(t, out, "Comment posted on media 111222333")
}

func TestCommentsDeleteDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "delete", "--media-id=111222333", "--comment-id=comment_111", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
}

func TestCommentsDeleteNoConfirm(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	err := runCmdErr(t, root, "comments", "delete", "--media-id=111222333", "--comment-id=comment_111")
	if err == nil {
		t.Error("expected error when --confirm not provided")
	}
}

func TestCommentsDeleteConfirmed(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "delete", "--media-id=111222333", "--comment-id=comment_111", "--confirm")

	mustContain(t, out, "Deleted comment comment_111 on media 111222333")
}

func TestCommentsLikeDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "like", "--comment-id=comment_111", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
}

func TestCommentsLike(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "like", "--comment-id=comment_111")

	mustContain(t, out, "Liked comment comment_111")
}

func TestCommentsUnlikeDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "unlike", "--comment-id=comment_111", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
}

func TestCommentsUnlike(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "unlike", "--comment-id=comment_111")

	mustContain(t, out, "Unliked comment comment_111")
}

func TestCommentsDisableDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "disable", "--media-id=111222333", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
}

func TestCommentsDisable(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "disable", "--media-id=111222333")

	mustContain(t, out, "Disabled comments on media 111222333")
}

func TestCommentsEnableDryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "enable", "--media-id=111222333", "--dry-run")

	mustContain(t, out, "[DRY RUN]")
}

func TestCommentsEnable(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "enable", "--media-id=111222333")

	mustContain(t, out, "Enabled comments on media 111222333")
}

func TestCommentsAlias(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	// "comment" is an alias for "comments"
	out := runCmd(t, root, "comment", "list", "--media-id=111222333")
	mustContain(t, out, "comment_111")
}

// TestCommentsListGraphQLFallback verifies that when the mobile API returns
// status:fail, the command falls back to the GraphQL endpoint and returns comments.
func TestCommentsListGraphQLFallback(t *testing.T) {
	// Build a mock server where the mobile comments endpoint returns status:fail,
	// forcing the GraphQL fallback path.
	mux := http.NewServeMux()
	withProfileMock(mux)
	withStoriesMock(mux)
	withReelsMock(mux)
	withLikesMock(mux)
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

	// Register a media mock that returns status:fail for comments but works for info.
	mux.HandleFunc("/api/v1/media/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/media/")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}
		action := strings.TrimSuffix(parts[1], "/")
		switch {
		case action == "comments":
			// Return status:fail to trigger GraphQL fallback.
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"comments": []any{},
				"status":   "fail",
			})
		case action == "info":
			// Return media info with shortcode (needed for GraphQL fallback).
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{"id": parts[0], "code": "TEST_SHORTCODE"},
				},
				"status": "ok",
			})
		default:
			http.NotFound(w, r)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "list", "--media-id=111222333")

	// Should contain the media detail GraphQL comment.
	mustContain(t, out, "comment_detail_111")
	mustContain(t, out, "Comment from media detail!")
	mustContain(t, out, "detail_comme")
}

// TestCommentsListGraphQLFallbackJSON verifies JSON output from the GraphQL fallback.
func TestCommentsListGraphQLFallbackJSON(t *testing.T) {
	mux := http.NewServeMux()
	withProfileMock(mux)
	withStoriesMock(mux)
	withReelsMock(mux)
	withLikesMock(mux)
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

	mux.HandleFunc("/api/v1/media/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/media/")
		parts := strings.SplitN(path, "/", 2)
		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}
		action := strings.TrimSuffix(parts[1], "/")
		switch {
		case action == "comments":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"comments": []any{}, "status": "fail"})
		case action == "info":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{
				"items":  []map[string]any{{"id": parts[0], "code": "TEST_SHORTCODE"}},
				"status": "ok",
			})
		default:
			http.NotFound(w, r)
		}
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	factory := newTestClientFactory(server)

	root := newTestRootCmd()
	root.AddCommand(buildTestCommentsCmd(factory))
	out := runCmd(t, root, "comments", "list", "--media-id=111222333", "--json")

	var comments []CommentSummary
	if err := json.Unmarshal([]byte(out), &comments); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, out)
	}
	if len(comments) == 0 {
		t.Fatal("expected at least one comment from GraphQL fallback")
	}
	if comments[0].PK != "comment_detail_111" {
		t.Errorf("expected pk=comment_detail_111, got %s", comments[0].PK)
	}
}
