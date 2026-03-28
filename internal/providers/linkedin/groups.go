package linkedin

import (
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// voyagerGroupsResponse is the response envelope for GET /voyager/api/groups/groups.
type voyagerGroupsResponse struct {
	Elements []voyagerGroupElement `json:"elements"`
	Paging   voyagerPaging         `json:"paging"`
}

type voyagerGroupElement struct {
	EntityURN   string `json:"entityUrn"`
	Name        string `json:"name"`
	MemberCount int    `json:"memberCount"`
	Description string `json:"description"`
}

// voyagerGroupMembersResponse is the response for GET /voyager/api/groups/groups/{id}/members.
type voyagerGroupMembersResponse struct {
	Elements []struct {
		EntityURN string `json:"entityUrn"`
		MiniProfile struct {
			FirstName        string `json:"firstName"`
			LastName         string `json:"lastName"`
			PublicIdentifier string `json:"publicIdentifier"`
			Occupation       string `json:"occupation"`
		} `json:"miniProfile"`
	} `json:"elements"`
	Paging voyagerPaging `json:"paging"`
}

// voyagerGroupPostsResponse is the response for GET /voyager/api/groups/groups/{id}/posts.
type voyagerGroupPostsResponse struct {
	Elements []struct {
		UpdateMetadata struct {
			URN string `json:"urn"`
		} `json:"updateMetadata"`
		Actor struct {
			Name struct {
				Text string `json:"text"`
			} `json:"name"`
			URN string `json:"urn"`
		} `json:"actor"`
		Commentary struct {
			Text struct {
				Text string `json:"text"`
			} `json:"text"`
		} `json:"commentary"`
		SocialDetail struct {
			TotalSocialActivityCounts struct {
				NumLikes    int `json:"numLikes"`
				NumComments int `json:"numComments"`
				NumShares   int `json:"numShares"`
			} `json:"totalSocialActivityCounts"`
		} `json:"socialDetail"`
		CreatedAt int64 `json:"createdAt"`
	} `json:"elements"`
	Paging voyagerPaging `json:"paging"`
}

// toGroupSummary maps a voyagerGroupElement to GroupSummary.
func toGroupSummary(el voyagerGroupElement) GroupSummary {
	id := el.EntityURN
	if parts := strings.Split(el.EntityURN, ":"); len(parts) > 0 {
		id = parts[len(parts)-1]
	}
	return GroupSummary{
		ID:          id,
		Name:        el.Name,
		MemberCount: el.MemberCount,
		Description: el.Description,
	}
}

// newGroupsCmd builds the "groups" subcommand group.
func newGroupsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "groups",
		Short:   "Interact with LinkedIn groups",
		Aliases: []string{"group"},
	}
	cmd.AddCommand(newGroupsListCmd(factory))
	cmd.AddCommand(newGroupsGetCmd(factory))
	cmd.AddCommand(newGroupsMembersCmd(factory))
	cmd.AddCommand(newGroupsPostsCmd(factory))
	cmd.AddCommand(newGroupsJoinCmd(factory))
	cmd.AddCommand(newGroupsLeaveCmd(factory))
	return cmd
}

// newGroupsListCmd builds the "groups list" command.
func newGroupsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List groups you are a member of",
		RunE:  makeRunGroupsList(factory),
	}
	cmd.Flags().Int("limit", 10, "Maximum number of groups to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	return cmd
}

// newGroupsGetCmd builds the "groups get" command.
func newGroupsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a group by ID",
		RunE:  makeRunGroupsGet(factory),
	}
	cmd.Flags().String("group-id", "", "Group ID (required)")
	_ = cmd.MarkFlagRequired("group-id")
	return cmd
}

// newGroupsMembersCmd builds the "groups members" command.
func newGroupsMembersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members",
		Short: "List members of a group",
		RunE:  makeRunGroupsMembers(factory),
	}
	cmd.Flags().String("group-id", "", "Group ID (required)")
	cmd.Flags().Int("limit", 10, "Maximum number of members to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	_ = cmd.MarkFlagRequired("group-id")
	return cmd
}

// newGroupsPostsCmd builds the "groups posts" command.
func newGroupsPostsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "posts",
		Short: "List posts in a group",
		RunE:  makeRunGroupsPosts(factory),
	}
	cmd.Flags().String("group-id", "", "Group ID (required)")
	cmd.Flags().Int("limit", 10, "Maximum number of posts to return")
	cmd.Flags().String("cursor", "", "Pagination cursor (start offset)")
	_ = cmd.MarkFlagRequired("group-id")
	return cmd
}

func makeRunGroupsList(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

func makeRunGroupsGet(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

func makeRunGroupsMembers(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

func makeRunGroupsPosts(_ ClientFactory) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		return errEndpointDeprecated
	}
}

// newGroupsJoinCmd builds the "groups join" command.
func newGroupsJoinCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join",
		Short: "Join a LinkedIn group",
		RunE:  makeRunGroupsJoin(factory),
	}
	cmd.Flags().String("group-id", "", "Group ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview action without executing it")
	_ = cmd.MarkFlagRequired("group-id")
	return cmd
}

func makeRunGroupsJoin(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		groupID, _ := cmd.Flags().GetString("group-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("join group %s", groupID),
				map[string]string{"joined": "true", "group_id": groupID})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/groups/groups/" + groupID + "/members"
		_, err = client.PostJSON(ctx, path, map[string]any{"groupUrn": "urn:li:fs_group:" + groupID})
		if err != nil {
			return fmt.Errorf("joining group %s: %w", groupID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"joined": "true", "group_id": groupID})
		}
		fmt.Printf("Joined group %s\n", groupID)
		return nil
	}
}

// newGroupsLeaveCmd builds the "groups leave" command.
func newGroupsLeaveCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leave",
		Short: "Leave a LinkedIn group",
		RunE:  makeRunGroupsLeave(factory),
	}
	cmd.Flags().String("group-id", "", "Group ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm the leave action")
	cmd.Flags().Bool("dry-run", false, "Preview action without executing it")
	_ = cmd.MarkFlagRequired("group-id")
	return cmd
}

func makeRunGroupsLeave(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		groupID, _ := cmd.Flags().GetString("group-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("leave group %s", groupID),
				map[string]string{"group_id": groupID})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := "/voyager/api/groups/groups/" + groupID + "/members/me"
		_, err = client.Delete(ctx, path)
		if err != nil {
			return fmt.Errorf("leaving group %s: %w", groupID, err)
		}

		fmt.Printf("Left group %s\n", groupID)
		return nil
	}
}

// printGroupSummaries outputs group summaries as JSON or text.
func printGroupSummaries(cmd *cobra.Command, groups []GroupSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(groups)
	}
	if len(groups) == 0 {
		fmt.Println("No groups found.")
		return nil
	}
	lines := make([]string, 0, len(groups)+1)
	lines = append(lines, fmt.Sprintf("%-15s  %-35s  %-10s  %-50s", "ID", "NAME", "MEMBERS", "DESCRIPTION"))
	for _, g := range groups {
		lines = append(lines, fmt.Sprintf("%-15s  %-35s  %-10s  %-50s",
			truncate(g.ID, 15),
			truncate(g.Name, 35),
			formatCount(g.MemberCount),
			truncate(g.Description, 50),
		))
	}
	cli.PrintText(lines)
	return nil
}

// printGroupDetail outputs a single group as JSON or formatted text block.
func printGroupDetail(cmd *cobra.Command, g GroupSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(g)
	}
	lines := []string{
		fmt.Sprintf("ID:          %s", g.ID),
		fmt.Sprintf("Name:        %s", g.Name),
		fmt.Sprintf("Members:     %s", formatCount(g.MemberCount)),
		fmt.Sprintf("Description: %s", g.Description),
	}
	cli.PrintText(lines)
	return nil
}
