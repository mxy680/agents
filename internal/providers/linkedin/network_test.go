package linkedin

import (
	"testing"
)

func TestNetworkFollowers_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"network", "followers"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Follower") {
		t.Errorf("expected 'Follower' in output, got: %s", out)
	}
	if !containsStr(out, "One") {
		t.Errorf("expected 'One' in output, got: %s", out)
	}
}

func TestNetworkFollowers_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"network", "followers", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"public_id"`) {
		t.Errorf("expected JSON field 'public_id' in output, got: %s", out)
	}
	if !containsStr(out, "follower-one") {
		t.Errorf("expected 'follower-one' in JSON output, got: %s", out)
	}
}

func TestNetworkFollowers_InvalidCursor(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"network", "followers", "--cursor", "bad"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for invalid cursor")
	}
}

func TestNetworkFollowing_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"network", "following"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Following") {
		t.Errorf("expected 'Following' in output, got: %s", out)
	}
}

func TestNetworkFollowing_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"network", "following", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "following-one") {
		t.Errorf("expected 'following-one' in JSON output, got: %s", out)
	}
}

func TestNetworkFollowing_InvalidCursor(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"network", "following", "--cursor", "bad"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for invalid cursor")
	}
}

func TestNetworkFollow_MissingURN(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"network", "follow"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --urn is missing")
	}
	if !containsStr(err.Error(), "--urn") {
		t.Errorf("expected '--urn' in error message, got: %s", err.Error())
	}
}

func TestNetworkFollow_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"network", "follow", "--urn", "urn:li:fs_normalized_company:1234", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestNetworkFollow_Success(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"network", "follow", "--urn", "urn:li:fs_normalized_company:1234"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "following") {
		t.Errorf("expected 'following' in output, got: %s", out)
	}
}

func TestNetworkFollow_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"network", "follow", "--urn", "urn:li:fs_normalized_company:1234", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON 'status' field in output, got: %s", out)
	}
	if !containsStr(out, "following") {
		t.Errorf("expected 'following' in JSON output, got: %s", out)
	}
}

func TestNetworkUnfollow_MissingURN(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"network", "unfollow"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error when --urn is missing")
	}
}

func TestNetworkUnfollow_DryRun(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"network", "unfollow", "--urn", "urn:li:fs_normalized_company:1234", "--dry-run"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "DRY RUN") {
		t.Errorf("expected '[DRY RUN]' in output, got: %s", out)
	}
}

func TestNetworkUnfollow_Success(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"network", "unfollow", "--urn", "urn:li:fs_normalized_company:1234"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Unfollowed") {
		t.Errorf("expected 'Unfollowed' in output, got: %s", out)
	}
}

func TestNetworkUnfollow_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"network", "unfollow", "--urn", "urn:li:fs_normalized_company:1234", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, `"status"`) {
		t.Errorf("expected JSON 'status' field in output, got: %s", out)
	}
	if !containsStr(out, "unfollowed") {
		t.Errorf("expected 'unfollowed' in JSON output, got: %s", out)
	}
}

func TestNetworkSuggestions_Text(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"network", "suggestions"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "Suggested") {
		t.Errorf("expected 'Suggested' in output, got: %s", out)
	}
}

func TestNetworkSuggestions_JSON(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	out := captureStdout(t, func() {
		root.SetArgs([]string{"network", "suggestions", "--json"})
		root.Execute() //nolint:errcheck
	})

	if !containsStr(out, "suggested-person") {
		t.Errorf("expected 'suggested-person' in JSON output, got: %s", out)
	}
}

func TestFollowersResponseToProfiles_MiniProfile(t *testing.T) {
	raw := voyagerFollowersResponse{
		Elements: []struct {
			EntityURN  string `json:"entityUrn"`
			FirstName  string `json:"firstName"`
			LastName   string `json:"lastName"`
			Occupation string `json:"occupation"`
			PublicIdentifier string `json:"publicIdentifier"`
			MiniProfile struct {
				EntityURN        string `json:"entityUrn"`
				FirstName        string `json:"firstName"`
				LastName         string `json:"lastName"`
				Occupation       string `json:"occupation"`
				PublicIdentifier string `json:"publicIdentifier"`
			} `json:"miniProfile"`
		}{
			{
				MiniProfile: struct {
					EntityURN        string `json:"entityUrn"`
					FirstName        string `json:"firstName"`
					LastName         string `json:"lastName"`
					Occupation       string `json:"occupation"`
					PublicIdentifier string `json:"publicIdentifier"`
				}{
					EntityURN:        "urn:li:fs_miniProfile:Mini1",
					FirstName:        "Mini",
					LastName:         "Profile",
					Occupation:       "Dev",
					PublicIdentifier: "mini-profile",
				},
			},
		},
	}
	summaries := followersResponseToProfiles(raw)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].URN != "urn:li:fs_miniProfile:Mini1" {
		t.Errorf("expected URN from miniProfile, got %s", summaries[0].URN)
	}
	if summaries[0].PublicID != "mini-profile" {
		t.Errorf("expected publicID 'mini-profile', got %s", summaries[0].PublicID)
	}
}
