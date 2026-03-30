package instagram

import (
	"fmt"
	"net/url"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// edgeCount models Instagram's {"count": N} nested structure.
type edgeCount struct {
	Count int64 `json:"count"`
}

// webProfileInfoResponse is the response envelope for GET /api/v1/users/web_profile_info/.
type webProfileInfoResponse struct {
	Data struct {
		User struct {
			ID             string    `json:"id"`
			Username       string    `json:"username"`
			FullName       string    `json:"full_name"`
			IsPrivate      bool      `json:"is_private"`
			IsVerified     bool      `json:"is_verified"`
			Biography      string    `json:"biography"`
			ExternalURL    *string   `json:"external_url"`
			EdgeFollowedBy edgeCount `json:"edge_followed_by"`
			EdgeFollow     edgeCount `json:"edge_follow"`
			EdgeMedia      edgeCount `json:"edge_owner_to_timeline_media"`
			ProfilePicURL  string    `json:"profile_pic_url"`
			IsBusiness     bool      `json:"is_business_account"`
			IsProfessional bool      `json:"is_professional_account"`
			Category       string    `json:"category_name"`
		} `json:"user"`
	} `json:"data"`
	Status string `json:"status"`
}

// userInfoResponse is the response envelope for GET /api/v1/users/{id}/info/.
type userInfoResponse struct {
	User struct {
		PK              string `json:"pk"`
		Username        string `json:"username"`
		FullName        string `json:"full_name"`
		IsPrivate       bool   `json:"is_private"`
		IsVerified      bool   `json:"is_verified"`
		Biography       string `json:"biography"`
		ExternalURL     string `json:"external_url"`
		FollowerCount   int64  `json:"follower_count"`
		FollowingCount  int64  `json:"following_count"`
		MediaCount      int64  `json:"media_count"`
		TotalClipsCount int64  `json:"total_clips_count"`
		IsBusiness      bool   `json:"is_business"`
		AccountType     int    `json:"account_type"`
		ProfilePicURL   string `json:"profile_pic_url"`
		HasProfilePic   bool   `json:"has_profile_pic"`
		IsProfessional  bool   `json:"is_professional_account"`
	} `json:"user"`
	Status string `json:"status"`
}

func newProfileGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a user profile",
		Long:  "Get profile information for a user by username or user ID. Defaults to own profile when no flags are provided.",
		RunE:  makeRunProfileGet(factory),
	}
	cmd.Flags().String("username", "", "Instagram username to look up")
	cmd.Flags().String("user-id", "", "Instagram user ID to look up (uses /api/v1/users/{id}/info/)")
	return cmd
}

func makeRunProfileGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		username, _ := cmd.Flags().GetString("username")
		userID, _ := cmd.Flags().GetString("user-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		// --user-id takes precedence over --username; uses a different REST endpoint.
		if userID != "" {
			resp, err := client.MobileGet(ctx, "/api/v1/users/"+url.PathEscape(userID)+"/info/", nil)
			if err != nil {
				return fmt.Errorf("getting user info for %s: %w", userID, err)
			}

			var info userInfoResponse
			if err := client.DecodeJSON(resp, &info); err != nil {
				return fmt.Errorf("decoding user info: %w", err)
			}

			u := info.User
			detail := UserDetail{
				ID:              u.PK,
				Username:        u.Username,
				FullName:        u.FullName,
				ProfilePicURL:   u.ProfilePicURL,
				IsPrivate:       u.IsPrivate,
				IsVerified:      u.IsVerified,
				Biography:       u.Biography,
				ExternalURL:     u.ExternalURL,
				FollowerCount:   u.FollowerCount,
				FollowingCount:  u.FollowingCount,
				MediaCount:      u.MediaCount,
				TotalClipsCount: u.TotalClipsCount,
				IsBusiness:      u.IsBusiness,
				AccountType:     u.AccountType,
				HasProfilePic:   u.HasProfilePic,
				IsProfessional:  u.IsProfessional,
			}
			return printUserDetail(cmd, detail)
		}

		// Use the web_profile_info endpoint for username lookup (or own profile).
		if username == "" {
			return fmt.Errorf("provide --username or --user-id; own-profile lookup requires a username")
		}

		params := url.Values{}
		params.Set("username", username)

		resp, err := client.Get(ctx, "/api/v1/users/web_profile_info/", params)
		if err != nil {
			return fmt.Errorf("getting web profile info for %s: %w", username, err)
		}

		var info webProfileInfoResponse
		if err := client.DecodeJSON(resp, &info); err != nil {
			return fmt.Errorf("decoding web profile info: %w", err)
		}

		u := info.Data.User
		extURL := ""
		if u.ExternalURL != nil {
			extURL = *u.ExternalURL
		}
		detail := UserDetail{
			ID:             u.ID,
			Username:       u.Username,
			FullName:       u.FullName,
			ProfilePicURL:  u.ProfilePicURL,
			IsPrivate:      u.IsPrivate,
			IsVerified:     u.IsVerified,
			Biography:      u.Biography,
			ExternalURL:    extURL,
			FollowerCount:  u.EdgeFollowedBy.Count,
			FollowingCount: u.EdgeFollow.Count,
			MediaCount:     u.EdgeMedia.Count,
			IsBusiness:     u.IsBusiness,
			Category:       u.Category,
			IsProfessional: u.IsProfessional,
		}
		return printUserDetail(cmd, detail)
	}
}

// printUserDetail outputs a UserDetail as JSON or a formatted text block.
func printUserDetail(cmd *cobra.Command, detail UserDetail) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(detail)
	}

	lines := []string{
		fmt.Sprintf("ID:           %s", detail.ID),
		fmt.Sprintf("Username:     %s", detail.Username),
		fmt.Sprintf("Full Name:    %s", detail.FullName),
		fmt.Sprintf("Private:      %v", detail.IsPrivate),
		fmt.Sprintf("Verified:     %v", detail.IsVerified),
		fmt.Sprintf("Biography:    %s", detail.Biography),
		fmt.Sprintf("Followers:    %s", formatCount(detail.FollowerCount)),
		fmt.Sprintf("Following:    %s", formatCount(detail.FollowingCount)),
		fmt.Sprintf("Posts:        %s", formatCount(detail.MediaCount)),
		fmt.Sprintf("Reels:        %s", formatCount(detail.TotalClipsCount)),
		fmt.Sprintf("Business:     %v", detail.IsBusiness),
		fmt.Sprintf("Professional: %v", detail.IsProfessional),
	}
	if detail.ExternalURL != "" {
		lines = append(lines, fmt.Sprintf("Website:      %s", detail.ExternalURL))
	}
	if detail.Category != "" {
		lines = append(lines, fmt.Sprintf("Category:     %s", detail.Category))
	}
	cli.PrintText(lines)
	return nil
}
