package linkedin

import (
	"testing"
)

func TestFeedList_Deprecated(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFeedCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"feed", "list"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for deprecated feed list endpoint")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message, got: %s", err.Error())
	}
}

func TestFeedHashtag_Deprecated(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFeedCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"feed", "hashtag", "--tag", "golang"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for deprecated feed hashtag endpoint")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message, got: %s", err.Error())
	}
}

func TestFeedHashtag_MissingTag(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newFeedCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"feed", "hashtag"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --tag is missing")
	}
}
