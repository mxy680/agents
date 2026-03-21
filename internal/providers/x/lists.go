package x

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// GraphQL query hashes for list operations.
const (
	hashListByRestId                = "9hbYpeVBMq8-yB8slayGWQ"
	hashListsManagementPageTimeline = "47170qwZCt5aFo9cBwFoNA"
	hashListLatestTweetsTimeline    = "HjsWc-nwwHKYwHenbHm-tw"
	hashListMembers                 = "BQp2IEYkgxuSxqbTAr1e1g"
	hashListSubscribers             = "74wGEkaBxrdoXakWTWMxRQ"
	hashCreateList                  = "EYg7JZU3A1eJ-wr2eygPHQ"
	hashUpdateList                  = "dIEI1sbSAuZlxhE0ggrezA"
	hashDeleteList                  = "UnN9Th1BDbeLjpgjGzn3MQ"
	hashListAddMember               = "lLNsL7mW6gSEQG6rXP7TNw"
	hashListRemoveMember            = "cvDFkG5WjcXV0Qw5nfe1qQ"
	hashEditListBanner              = "t_DsROHldculsB0B9BUAWw"
	hashDeleteListBanner            = "Y90WuxdWugtMRJhkXTdvzg"
)

// newListsCmd builds the "lists" subcommand group.
func newListsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "lists",
		Short:   "Manage X lists",
		Aliases: []string{"list"},
	}
	cmd.AddCommand(newListsGetCmd(factory))
	cmd.AddCommand(newListsOwnedCmd(factory))
	cmd.AddCommand(newListsSearchCmd(factory))
	cmd.AddCommand(newListsTweetsCmd(factory))
	cmd.AddCommand(newListsMembersCmd(factory))
	cmd.AddCommand(newListsSubscribersCmd(factory))
	cmd.AddCommand(newListsCreateCmd(factory))
	cmd.AddCommand(newListsUpdateCmd(factory))
	cmd.AddCommand(newListsDeleteCmd(factory))
	cmd.AddCommand(newListsAddMemberCmd(factory))
	cmd.AddCommand(newListsRemoveMemberCmd(factory))
	cmd.AddCommand(newListsSetBannerCmd(factory))
	cmd.AddCommand(newListsRemoveBannerCmd(factory))
	return cmd
}

// newListsGetCmd builds the "lists get" command.
func newListsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a list by ID",
		RunE:  makeRunListsGet(factory),
	}
	cmd.Flags().String("list-id", "", "List ID (required)")
	_ = cmd.MarkFlagRequired("list-id")
	return cmd
}

// newListsOwnedCmd builds the "lists owned" command.
func newListsOwnedCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "owned",
		Short: "List lists owned by the authenticated user",
		RunE:  makeRunListsOwned(factory),
	}
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newListsSearchCmd builds the "lists search" command.
func newListsSearchCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search for lists",
		RunE:  makeRunListsSearch(factory),
	}
	cmd.Flags().String("query", "", "Search query (required)")
	_ = cmd.MarkFlagRequired("query")
	cmd.Flags().Int("limit", 20, "Maximum number of results")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newListsTweetsCmd builds the "lists tweets" command.
func newListsTweetsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tweets",
		Short: "Get tweets from a list",
		RunE:  makeRunListsTweets(factory),
	}
	cmd.Flags().String("list-id", "", "List ID (required)")
	_ = cmd.MarkFlagRequired("list-id")
	cmd.Flags().Int("limit", 20, "Maximum number of tweets")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newListsMembersCmd builds the "lists members" command.
func newListsMembersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "members",
		Short: "Get members of a list",
		RunE:  makeRunListsMembers(factory),
	}
	cmd.Flags().String("list-id", "", "List ID (required)")
	_ = cmd.MarkFlagRequired("list-id")
	cmd.Flags().Int("limit", 20, "Maximum number of members")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newListsSubscribersCmd builds the "lists subscribers" command.
func newListsSubscribersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscribers",
		Short: "Get subscribers of a list",
		RunE:  makeRunListsSubscribers(factory),
	}
	cmd.Flags().String("list-id", "", "List ID (required)")
	_ = cmd.MarkFlagRequired("list-id")
	cmd.Flags().Int("limit", 20, "Maximum number of subscribers")
	cmd.Flags().String("cursor", "", "Pagination cursor")
	return cmd
}

// newListsCreateCmd builds the "lists create" command.
func newListsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new list",
		RunE:  makeRunListsCreate(factory),
	}
	cmd.Flags().String("name", "", "List name (required)")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().String("description", "", "List description")
	cmd.Flags().Bool("private", false, "Make the list private")
	cmd.Flags().Bool("dry-run", false, "Print what would be created without creating")
	return cmd
}

// newListsUpdateCmd builds the "lists update" command.
func newListsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a list",
		RunE:  makeRunListsUpdate(factory),
	}
	cmd.Flags().String("list-id", "", "List ID (required)")
	_ = cmd.MarkFlagRequired("list-id")
	cmd.Flags().String("name", "", "New list name")
	cmd.Flags().String("description", "", "New list description")
	cmd.Flags().Bool("private", false, "Make the list private")
	cmd.Flags().Bool("dry-run", false, "Print what would be updated without updating")
	return cmd
}

// newListsDeleteCmd builds the "lists delete" command.
func newListsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a list",
		RunE:  makeRunListsDelete(factory),
	}
	cmd.Flags().String("list-id", "", "List ID (required)")
	_ = cmd.MarkFlagRequired("list-id")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Print what would be deleted without deleting")
	return cmd
}

// newListsAddMemberCmd builds the "lists add-member" command.
func newListsAddMemberCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-member",
		Short: "Add a member to a list",
		RunE:  makeRunListsAddMember(factory),
	}
	cmd.Flags().String("list-id", "", "List ID (required)")
	_ = cmd.MarkFlagRequired("list-id")
	cmd.Flags().String("user-id", "", "User ID to add (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without doing it")
	return cmd
}

// newListsRemoveMemberCmd builds the "lists remove-member" command.
func newListsRemoveMemberCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-member",
		Short: "Remove a member from a list",
		RunE:  makeRunListsRemoveMember(factory),
	}
	cmd.Flags().String("list-id", "", "List ID (required)")
	_ = cmd.MarkFlagRequired("list-id")
	cmd.Flags().String("user-id", "", "User ID to remove (required)")
	_ = cmd.MarkFlagRequired("user-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without doing it")
	return cmd
}

// newListsSetBannerCmd builds the "lists set-banner" command.
func newListsSetBannerCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-banner",
		Short: "Set a banner image for a list",
		RunE:  makeRunListsSetBanner(factory),
	}
	cmd.Flags().String("list-id", "", "List ID (required)")
	_ = cmd.MarkFlagRequired("list-id")
	cmd.Flags().String("path", "", "Path to image file (required)")
	_ = cmd.MarkFlagRequired("path")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without doing it")
	return cmd
}

// newListsRemoveBannerCmd builds the "lists remove-banner" command.
func newListsRemoveBannerCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-banner",
		Short: "Remove the banner image from a list",
		RunE:  makeRunListsRemoveBanner(factory),
	}
	cmd.Flags().String("list-id", "", "List ID (required)")
	_ = cmd.MarkFlagRequired("list-id")
	cmd.Flags().Bool("dry-run", false, "Print what would be done without doing it")
	return cmd
}

// --- RunE implementations ---

func makeRunListsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		listID, _ := cmd.Flags().GetString("list-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"listId": listID,
		}

		data, err := client.GraphQL(ctx, hashListByRestId, "ListByRestId", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("getting list %s: %w", listID, err)
		}

		var payload struct {
			List json.RawMessage `json:"list"`
		}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("parse list response: %w", err)
		}

		list, err := parseListResult(payload.List)
		if err != nil {
			return fmt.Errorf("parse list: %w", err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(list)
		}

		lines := []string{
			fmt.Sprintf("ID:          %s", list.ID),
			fmt.Sprintf("Name:        %s", list.Name),
			fmt.Sprintf("Description: %s", list.Description),
			fmt.Sprintf("Members:     %d", list.MemberCount),
			fmt.Sprintf("Subscribers: %d", list.SubscriberCount),
			fmt.Sprintf("Private:     %v", list.Private),
			fmt.Sprintf("Owner:       %s (@%s)", list.OwnerName, list.OwnerUsername),
		}
		cli.PrintText(lines)
		return nil
	}
}

func makeRunListsOwned(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"count": limit,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashListsManagementPageTimeline, "ListsManagementPageTimeline", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching owned lists: %w", err)
		}

		lists, nextCursor, err := extractListTimeline(data)
		if err != nil {
			return fmt.Errorf("extract owned lists: %w", err)
		}

		return printListSummaries(cmd, lists, nextCursor)
	}
}

func makeRunListsSearch(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"rawQuery":    query,
			"count":       limit,
			"product":     "Lists",
			"querySource": "typed_query",
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashSearchTimeline, "SearchTimeline", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("searching lists: %w", err)
		}

		lists, nextCursor, err := extractListTimeline(data)
		if err != nil {
			// Fall back to empty if the search response uses tweet timeline format.
			lists = []ListSummary{}
			nextCursor = ""
		}

		return printListSummaries(cmd, lists, nextCursor)
	}
}

func makeRunListsTweets(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		listID, _ := cmd.Flags().GetString("list-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"listId": listID,
			"count":  limit,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashListLatestTweetsTimeline, "ListLatestTweetsTimeline", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching list tweets for %s: %w", listID, err)
		}

		return printTimelineResult(cmd, data)
	}
}

func makeRunListsMembers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		listID, _ := cmd.Flags().GetString("list-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"listId": listID,
			"count":  limit,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashListMembers, "ListMembers", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching list members for %s: %w", listID, err)
		}

		return printUserListResult(cmd, data)
	}
}

func makeRunListsSubscribers(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		listID, _ := cmd.Flags().GetString("list-id")
		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"listId": listID,
			"count":  limit,
		}
		if cursor != "" {
			vars["cursor"] = cursor
		}

		data, err := client.GraphQL(ctx, hashListSubscribers, "ListSubscribers", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("fetching list subscribers for %s: %w", listID, err)
		}

		return printUserListResult(cmd, data)
	}
}

func makeRunListsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		private, _ := cmd.Flags().GetBool("private")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		mode := "Public"
		if private {
			mode = "Private"
		}

		vars := map[string]any{
			"name":        name,
			"description": description,
			"mode":        mode,
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("create list %q", name), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := client.GraphQLPost(ctx, hashCreateList, "CreateList", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("creating list %q: %w", name, err)
		}

		list, err := parseListMutationResult(data)
		if err != nil {
			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]string{"status": "created", "name": name})
			}
			fmt.Printf("List created: %s\n", name)
			return nil
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(list)
		}
		fmt.Printf("List created: %s (ID: %s)\n", list.Name, list.ID)
		return nil
	}
}

func makeRunListsUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		listID, _ := cmd.Flags().GetString("list-id")
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		private, _ := cmd.Flags().GetBool("private")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"listId": listID,
		}
		if name != "" {
			vars["name"] = name
		}
		if description != "" {
			vars["description"] = description
		}
		if private {
			vars["mode"] = "Private"
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("update list %s", listID), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, hashUpdateList, "UpdateList", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("updating list %s: %w", listID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "updated", "list_id": listID})
		}
		fmt.Printf("List updated: %s\n", listID)
		return nil
	}
}

func makeRunListsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		listID, _ := cmd.Flags().GetString("list-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("delete list %s", listID), map[string]string{"list_id": listID})
		}

		if err := confirmDestructive(cmd, "this action is irreversible"); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		vars := map[string]any{
			"listId": listID,
		}

		_, err = client.GraphQLPost(ctx, hashDeleteList, "DeleteList", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("deleting list %s: %w", listID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "list_id": listID})
		}
		fmt.Printf("List deleted: %s\n", listID)
		return nil
	}
}

func makeRunListsAddMember(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		listID, _ := cmd.Flags().GetString("list-id")
		userID, _ := cmd.Flags().GetString("user-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"listId": listID,
			"userId": userID,
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("add user %s to list %s", userID, listID), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, hashListAddMember, "ListAddMember", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("adding user %s to list %s: %w", userID, listID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "added", "list_id": listID, "user_id": userID})
		}
		fmt.Printf("User %s added to list %s\n", userID, listID)
		return nil
	}
}

func makeRunListsRemoveMember(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		listID, _ := cmd.Flags().GetString("list-id")
		userID, _ := cmd.Flags().GetString("user-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"listId": listID,
			"userId": userID,
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("remove user %s from list %s", userID, listID), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, hashListRemoveMember, "ListRemoveMember", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("removing user %s from list %s: %w", userID, listID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "removed", "list_id": listID, "user_id": userID})
		}
		fmt.Printf("User %s removed from list %s\n", userID, listID)
		return nil
	}
}

func makeRunListsSetBanner(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		listID, _ := cmd.Flags().GetString("list-id")
		path, _ := cmd.Flags().GetString("path")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("set banner for list %s from %s", listID, path), map[string]string{"list_id": listID, "path": path})
		}

		// Read image file.
		imageData, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading image file %s: %w", path, err)
		}

		vars := map[string]any{
			"listId": listID,
			"banner": fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(imageData), base64.StdEncoding.EncodeToString(imageData)),
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, hashEditListBanner, "EditListBanner", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("setting banner for list %s: %w", listID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "banner_set", "list_id": listID})
		}
		fmt.Printf("Banner set for list %s\n", listID)
		return nil
	}
}

func makeRunListsRemoveBanner(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		listID, _ := cmd.Flags().GetString("list-id")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		vars := map[string]any{
			"listId": listID,
		}

		if dryRun {
			return dryRunResult(cmd, fmt.Sprintf("remove banner from list %s", listID), vars)
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		_, err = client.GraphQLPost(ctx, hashDeleteListBanner, "DeleteListBanner", vars, DefaultFeatures)
		if err != nil {
			return fmt.Errorf("removing banner from list %s: %w", listID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "banner_removed", "list_id": listID})
		}
		fmt.Printf("Banner removed from list %s\n", listID)
		return nil
	}
}

// extractListTimeline extracts list summaries and a cursor from a GraphQL timeline response.
// List timeline responses contain list entries rather than tweet entries.
func extractListTimeline(data json.RawMessage) ([]ListSummary, string, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, "", fmt.Errorf("parse list timeline data: %w", err)
	}

	instructionsRaw, err := findInstructions(top)
	if err != nil {
		return nil, "", err
	}

	var instructions []struct {
		Type    string            `json:"type"`
		Entries []json.RawMessage `json:"entries"`
	}
	if err := json.Unmarshal(instructionsRaw, &instructions); err != nil {
		return nil, "", fmt.Errorf("parse list timeline instructions: %w", err)
	}

	var lists []ListSummary
	cursor := ""

	for _, instr := range instructions {
		if instr.Type != "TimelineAddEntries" {
			continue
		}
		for _, entryRaw := range instr.Entries {
			var entry struct {
				Content struct {
					EntryType  string `json:"entryType"`
					Value      string `json:"value"`
					CursorType string `json:"cursorType"`
					ItemContent struct {
						ItemType    string          `json:"itemType"`
						ListResults json.RawMessage `json:"list_results"`
					} `json:"itemContent"`
				} `json:"content"`
			}
			if err := json.Unmarshal(entryRaw, &entry); err != nil {
				continue
			}

			switch entry.Content.EntryType {
			case "TimelineTimelineCursor":
				if entry.Content.CursorType == "Bottom" {
					cursor = entry.Content.Value
				}
			case "TimelineTimelineItem":
				if entry.Content.ItemContent.ItemType == "TimelineList" &&
					entry.Content.ItemContent.ListResults != nil {
					var listWrapper struct {
						Result json.RawMessage `json:"result"`
					}
					if err := json.Unmarshal(entry.Content.ItemContent.ListResults, &listWrapper); err != nil {
						continue
					}
					list, err := parseListResult(listWrapper.Result)
					if err != nil {
						continue
					}
					lists = append(lists, *list)
				}
			}
		}
	}

	return lists, cursor, nil
}

// parseListMutationResult extracts a ListSummary from a create/update mutation response.
func parseListMutationResult(data json.RawMessage) (*ListSummary, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		return nil, fmt.Errorf("parse list mutation data: %w", err)
	}

	for _, v := range top {
		var wrapper struct {
			List json.RawMessage `json:"list"`
		}
		if err := json.Unmarshal(v, &wrapper); err == nil && wrapper.List != nil {
			return parseListResult(wrapper.List)
		}
		// Try direct list object.
		list, err := parseListResult(v)
		if err == nil && list.ID != "" {
			return list, nil
		}
	}

	return nil, fmt.Errorf("list not found in mutation response")
}

// printListSummaries outputs list summaries as JSON or text.
func printListSummaries(cmd *cobra.Command, lists []ListSummary, nextCursor string) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(map[string]any{
			"lists":       lists,
			"next_cursor": nextCursor,
		})
	}

	if len(lists) == 0 {
		fmt.Println("No lists found.")
		return nil
	}

	if nextCursor != "" {
		fmt.Printf("Next cursor: %s\n", nextCursor)
	}

	lines := make([]string, 0, len(lists)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-30s  %-10s  %-10s  %-8s", "ID", "NAME", "MEMBERS", "SUBS", "PRIVATE"))
	for _, l := range lists {
		lines = append(lines, fmt.Sprintf("%-20s  %-30s  %-10d  %-10d  %-8v",
			truncate(l.ID, 20),
			truncate(l.Name, 30),
			l.MemberCount,
			l.SubscriberCount,
			l.Private,
		))
	}
	cli.PrintText(lines)
	return nil
}

