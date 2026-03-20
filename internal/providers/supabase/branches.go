package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

func newBranchesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "branches",
		Aliases: []string{"branch"},
		Short:   "Manage preview branches",
	}
	cmd.AddCommand(
		newBranchesListCmd(factory),
		newBranchesGetCmd(factory),
		newBranchesCreateCmd(factory),
		newBranchesUpdateCmd(factory),
		newBranchesDeleteCmd(factory),
		newBranchesPushCmd(factory),
		newBranchesMergeCmd(factory),
		newBranchesResetCmd(factory),
		newBranchesDiffCmd(factory),
		newBranchesDisableCmd(factory),
	)
	return cmd
}

// toBranchSummary converts a raw API response map to a BranchSummary.
func toBranchSummary(data map[string]any) BranchSummary {
	id, _ := data["id"].(string)
	name, _ := data["name"].(string)
	gitBranch, _ := data["git_branch"].(string)
	isDefault, _ := data["is_default"].(bool)
	status, _ := data["status"].(string)
	createdAt, _ := data["created_at"].(string)
	return BranchSummary{
		ID:        id,
		Name:      name,
		GitBranch: gitBranch,
		IsDefault: isDefault,
		Status:    status,
		CreatedAt: createdAt,
	}
}

// printBranchSummaries outputs branch summaries as JSON or a formatted text table.
func printBranchSummaries(cmd *cobra.Command, branches []BranchSummary) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(branches)
	}
	if len(branches) == 0 {
		fmt.Println("No branches found.")
		return nil
	}
	lines := make([]string, 0, len(branches)+1)
	lines = append(lines, fmt.Sprintf("%-20s  %-25s  %-25s  %s", "ID", "NAME", "GIT_BRANCH", "STATUS"))
	for _, b := range branches {
		lines = append(lines, fmt.Sprintf("%-20s  %-25s  %-25s  %s",
			truncate(b.ID, 20), truncate(b.Name, 25), truncate(b.GitBranch, 25), b.Status))
	}
	cli.PrintText(lines)
	return nil
}

// --- list ---

func newBranchesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List preview branches for a project",
		RunE:  makeRunBranchesList(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunBranchesList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/projects/%s/branches", ref), nil)
		if err != nil {
			return fmt.Errorf("listing branches: %w", err)
		}

		var raw []map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		branches := make([]BranchSummary, 0, len(raw))
		for _, r := range raw {
			branches = append(branches, toBranchSummary(r))
		}
		return printBranchSummaries(cmd, branches)
	}
}

// --- get ---

func newBranchesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details of a preview branch",
		RunE:  makeRunBranchesGet(factory),
	}
	cmd.Flags().String("branch-id", "", "Branch ID (required)")
	_ = cmd.MarkFlagRequired("branch-id")
	return cmd
}

func makeRunBranchesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		branchID, _ := cmd.Flags().GetString("branch-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/branches/%s", branchID), nil)
		if err != nil {
			return fmt.Errorf("getting branch %s: %w", branchID, err)
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		b := toBranchSummary(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(b)
		}

		lines := []string{
			fmt.Sprintf("ID:         %s", b.ID),
			fmt.Sprintf("Name:       %s", b.Name),
			fmt.Sprintf("Git Branch: %s", b.GitBranch),
			fmt.Sprintf("Default:    %v", b.IsDefault),
			fmt.Sprintf("Status:     %s", b.Status),
			fmt.Sprintf("Created:    %s", b.CreatedAt),
		}
		cli.PrintText(lines)
		return nil
	}
}

// --- create ---

func newBranchesCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a preview branch",
		RunE:  makeRunBranchesCreate(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().String("git-branch", "", "Git branch name (required)")
	cmd.Flags().String("region", "", "Region for the branch")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("git-branch")
	return cmd
}

func makeRunBranchesCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")
		gitBranch, _ := cmd.Flags().GetString("git-branch")
		region, _ := cmd.Flags().GetString("region")

		if dryRunResult(cmd, fmt.Sprintf("Would create preview branch from git branch %q in project %s", gitBranch, ref)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		bodyMap := map[string]any{
			"git_branch": gitBranch,
		}
		if region != "" {
			bodyMap["region"] = region
		}

		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}

		data, err := doSupabase(client, http.MethodPost, fmt.Sprintf("/projects/%s/branches", ref), bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("creating branch: %w", err)
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		b := toBranchSummary(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(b)
		}
		fmt.Printf("Created branch: %s (ID: %s)\n", b.Name, b.ID)
		return nil
	}
}

// --- update ---

func newBranchesUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a preview branch",
		RunE:  makeRunBranchesUpdate(factory),
	}
	cmd.Flags().String("branch-id", "", "Branch ID (required)")
	cmd.Flags().String("git-branch", "", "New git branch name")
	cmd.Flags().Bool("reset-on-push", false, "Reset branch on push")
	_ = cmd.MarkFlagRequired("branch-id")
	return cmd
}

func makeRunBranchesUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		branchID, _ := cmd.Flags().GetString("branch-id")
		gitBranch, _ := cmd.Flags().GetString("git-branch")
		resetOnPush, _ := cmd.Flags().GetBool("reset-on-push")

		if dryRunResult(cmd, fmt.Sprintf("Would update branch %s", branchID)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		bodyMap := map[string]any{
			"reset_on_push": resetOnPush,
		}
		if gitBranch != "" {
			bodyMap["git_branch"] = gitBranch
		}

		bodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			return fmt.Errorf("encoding request: %w", err)
		}

		data, err := doSupabase(client, http.MethodPatch, fmt.Sprintf("/branches/%s", branchID), bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("updating branch %s: %w", branchID, err)
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		b := toBranchSummary(raw)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(b)
		}
		fmt.Printf("Updated branch: %s (ID: %s)\n", b.Name, b.ID)
		return nil
	}
}

// --- delete ---

func newBranchesDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a preview branch (irreversible)",
		RunE:  makeRunBranchesDelete(factory),
	}
	cmd.Flags().String("branch-id", "", "Branch ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("branch-id")
	return cmd
}

func makeRunBranchesDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		branchID, _ := cmd.Flags().GetString("branch-id")

		if dryRunResult(cmd, fmt.Sprintf("Would permanently delete branch %s", branchID)) {
			return nil
		}

		if err := confirmDestructive(cmd, fmt.Sprintf("deleting branch %s is irreversible", branchID)); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if _, err := doSupabase(client, http.MethodDelete, fmt.Sprintf("/branches/%s", branchID), nil); err != nil {
			return fmt.Errorf("deleting branch %s: %w", branchID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "branchId": branchID})
		}
		fmt.Printf("Deleted branch: %s\n", branchID)
		return nil
	}
}

// --- push ---

func newBranchesPushCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push migrations to a preview branch",
		RunE:  makeRunBranchesPush(factory),
	}
	cmd.Flags().String("branch-id", "", "Branch ID (required)")
	_ = cmd.MarkFlagRequired("branch-id")
	return cmd
}

func makeRunBranchesPush(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		branchID, _ := cmd.Flags().GetString("branch-id")

		if dryRunResult(cmd, fmt.Sprintf("Would push migrations to branch %s", branchID)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodPost, fmt.Sprintf("/branches/%s/push", branchID), nil)
		if err != nil {
			return fmt.Errorf("pushing to branch %s: %w", branchID, err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw map[string]any
			if err := json.Unmarshal(data, &raw); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}
			return cli.PrintJSON(raw)
		}
		fmt.Printf("Push initiated for branch: %s\n", branchID)
		return nil
	}
}

// --- merge ---

func newBranchesMergeCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge a preview branch into the main branch",
		RunE:  makeRunBranchesMerge(factory),
	}
	cmd.Flags().String("branch-id", "", "Branch ID (required)")
	_ = cmd.MarkFlagRequired("branch-id")
	return cmd
}

func makeRunBranchesMerge(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		branchID, _ := cmd.Flags().GetString("branch-id")

		if dryRunResult(cmd, fmt.Sprintf("Would merge branch %s into main", branchID)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodPost, fmt.Sprintf("/branches/%s/merge", branchID), nil)
		if err != nil {
			return fmt.Errorf("merging branch %s: %w", branchID, err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw map[string]any
			if err := json.Unmarshal(data, &raw); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}
			return cli.PrintJSON(raw)
		}
		fmt.Printf("Merge initiated for branch: %s\n", branchID)
		return nil
	}
}

// --- reset ---

func newBranchesResetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Reset a preview branch to match the main branch",
		RunE:  makeRunBranchesReset(factory),
	}
	cmd.Flags().String("branch-id", "", "Branch ID (required)")
	_ = cmd.MarkFlagRequired("branch-id")
	return cmd
}

func makeRunBranchesReset(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		branchID, _ := cmd.Flags().GetString("branch-id")

		if dryRunResult(cmd, fmt.Sprintf("Would reset branch %s to match main", branchID)) {
			return nil
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodPost, fmt.Sprintf("/branches/%s/reset", branchID), nil)
		if err != nil {
			return fmt.Errorf("resetting branch %s: %w", branchID, err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw map[string]any
			if err := json.Unmarshal(data, &raw); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}
			return cli.PrintJSON(raw)
		}
		fmt.Printf("Reset initiated for branch: %s\n", branchID)
		return nil
	}
}

// --- diff ---

func newBranchesDiffCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Show schema diff for a preview branch",
		RunE:  makeRunBranchesDiff(factory),
	}
	cmd.Flags().String("branch-id", "", "Branch ID (required)")
	_ = cmd.MarkFlagRequired("branch-id")
	return cmd
}

func makeRunBranchesDiff(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		branchID, _ := cmd.Flags().GetString("branch-id")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		data, err := doSupabase(client, http.MethodGet, fmt.Sprintf("/branches/%s/diff", branchID), nil)
		if err != nil {
			return fmt.Errorf("getting diff for branch %s: %w", branchID, err)
		}

		if cli.IsJSONOutput(cmd) {
			var raw any
			if err := json.Unmarshal(data, &raw); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}
			return cli.PrintJSON(raw)
		}
		fmt.Print(string(data))
		return nil
	}
}

// --- disable ---

func newBranchesDisableCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable preview branching for a project",
		RunE:  makeRunBranchesDisable(factory),
	}
	cmd.Flags().String("ref", "", "Project ref (required)")
	cmd.Flags().Bool("confirm", false, "Confirm disabling branching")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunBranchesDisable(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ref, _ := cmd.Flags().GetString("ref")

		if dryRunResult(cmd, fmt.Sprintf("Would disable branching for project %s", ref)) {
			return nil
		}

		if err := confirmDestructive(cmd, fmt.Sprintf("disabling branching for project %s is irreversible", ref)); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if _, err := doSupabase(client, http.MethodDelete, fmt.Sprintf("/projects/%s/branches", ref), nil); err != nil {
			return fmt.Errorf("disabling branching for project %s: %w", ref, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "disabled", "ref": ref})
		}
		fmt.Printf("Branching disabled for project: %s\n", ref)
		return nil
	}
}
