package linkedin

import (
	"testing"
)

func TestNetworkFollowers_Deprecated(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"network", "followers"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for deprecated network followers endpoint")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message, got: %s", err.Error())
	}
}

func TestNetworkFollowing_Deprecated(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"network", "following"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for deprecated network following endpoint")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message, got: %s", err.Error())
	}
}

func TestNetworkSuggestions_Deprecated(t *testing.T) {
	server := newFullMockServer(t)
	defer server.Close()

	root := newTestRootCmd()
	root.AddCommand(newNetworkCmd(newTestClientFactory(server)))

	root.SetArgs([]string{"network", "suggestions"})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for deprecated network suggestions endpoint")
	}
	if !containsStr(err.Error(), "deprecated") {
		t.Errorf("expected 'deprecated' in error message, got: %s", err.Error())
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
