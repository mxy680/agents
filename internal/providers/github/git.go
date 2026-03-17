package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// --- Conversion helpers for Git types ---

func toRefInfo(data map[string]any) RefInfo {
	r := RefInfo{
		Ref:    jsonString(data["ref"]),
		NodeID: jsonString(data["node_id"]),
		URL:    jsonString(data["url"]),
	}
	if obj, ok := data["object"].(map[string]any); ok {
		r.Object.Type = jsonString(obj["type"])
		r.Object.SHA = jsonString(obj["sha"])
	}
	return r
}

func toGitCommitInfo(data map[string]any) GitCommitInfo {
	c := GitCommitInfo{
		SHA:     jsonString(data["sha"]),
		Message: jsonString(data["message"]),
		URL:     jsonString(data["url"]),
	}
	if author, ok := data["author"].(map[string]any); ok {
		c.Author.Name = jsonString(author["name"])
		c.Author.Email = jsonString(author["email"])
		c.Author.Date = jsonString(author["date"])
	}
	if tree, ok := data["tree"].(map[string]any); ok {
		c.Tree.SHA = jsonString(tree["sha"])
	}
	if parents, ok := data["parents"].([]any); ok {
		for _, p := range parents {
			if pm, ok := p.(map[string]any); ok {
				parent := struct {
					SHA string `json:"sha"`
				}{SHA: jsonString(pm["sha"])}
				c.Parents = append(c.Parents, parent)
			}
		}
	}
	return c
}

func toGitTreeInfo(data map[string]any) GitTreeInfo {
	t := GitTreeInfo{
		SHA:       jsonString(data["sha"]),
		URL:       jsonString(data["url"]),
		Truncated: jsonBool(data["truncated"]),
	}
	if entries, ok := data["tree"].([]any); ok {
		for _, e := range entries {
			if em, ok := e.(map[string]any); ok {
				t.Tree = append(t.Tree, TreeEntryInfo{
					Path: jsonString(em["path"]),
					Mode: jsonString(em["mode"]),
					Type: jsonString(em["type"]),
					Size: jsonInt(em["size"]),
					SHA:  jsonString(em["sha"]),
				})
			}
		}
	}
	return t
}

func toGitBlobInfo(data map[string]any) GitBlobInfo {
	return GitBlobInfo{
		SHA:      jsonString(data["sha"]),
		Size:     jsonInt(data["size"]),
		URL:      jsonString(data["url"]),
		Content:  jsonString(data["content"]),
		Encoding: jsonString(data["encoding"]),
	}
}

func toGitTagInfo(data map[string]any) GitTagInfo {
	t := GitTagInfo{
		Tag:     jsonString(data["tag"]),
		SHA:     jsonString(data["sha"]),
		Message: jsonString(data["message"]),
		URL:     jsonString(data["url"]),
	}
	if tagger, ok := data["tagger"].(map[string]any); ok {
		t.Tagger.Name = jsonString(tagger["name"])
		t.Tagger.Email = jsonString(tagger["email"])
		t.Tagger.Date = jsonString(tagger["date"])
	}
	if obj, ok := data["object"].(map[string]any); ok {
		t.Object.Type = jsonString(obj["type"])
		t.Object.SHA = jsonString(obj["sha"])
	}
	return t
}

// --- Refs commands ---

func newRefsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Git references in a repository",
		RunE:  makeRunRefsList(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("namespace", "", "Reference namespace to filter (e.g. heads, tags)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunRefsList(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		namespace, _ := cmd.Flags().GetString("namespace")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/git/refs", owner, repo)
		if namespace != "" {
			path = fmt.Sprintf("%s/%s", path, namespace)
		}

		var data []map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("listing refs for %s/%s: %w", owner, repo, err)
		}

		refs := make([]RefInfo, 0, len(data))
		for _, d := range data {
			refs = append(refs, toRefInfo(d))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(refs)
		}

		if len(refs) == 0 {
			fmt.Println("No refs found.")
			return nil
		}

		lines := make([]string, 0, len(refs)+1)
		lines = append(lines, fmt.Sprintf("%-50s  %-10s  %s", "REF", "TYPE", "SHA"))
		for _, r := range refs {
			lines = append(lines, fmt.Sprintf("%-50s  %-10s  %s", truncate(r.Ref, 50), r.Object.Type, r.Object.SHA))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newRefsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a single Git reference",
		RunE:  makeRunRefsGet(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("ref", "", "Reference name without refs/ prefix (e.g. heads/main) (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunRefsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		ref, _ := cmd.Flags().GetString("ref")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodGet, fmt.Sprintf("/repos/%s/%s/git/ref/%s", owner, repo, ref), nil, &data); err != nil {
			return fmt.Errorf("getting ref %s in %s/%s: %w", ref, owner, repo, err)
		}

		info := toRefInfo(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("Ref:     %s", info.Ref),
			fmt.Sprintf("NodeID:  %s", info.NodeID),
			fmt.Sprintf("URL:     %s", info.URL),
			fmt.Sprintf("Type:    %s", info.Object.Type),
			fmt.Sprintf("SHA:     %s", info.Object.SHA),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newRefsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Git reference",
		RunE:  makeRunRefsCreate(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("ref", "", "Full reference name (e.g. refs/heads/new-branch) (required)")
	cmd.Flags().String("sha", "", "SHA1 to point the reference to (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("sha")
	return cmd
}

func makeRunRefsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		ref, _ := cmd.Flags().GetString("ref")
		sha, _ := cmd.Flags().GetString("sha")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create ref %s pointing to %s in %s/%s", ref, sha, owner, repo), map[string]any{
				"action": "create_ref",
				"owner":  owner,
				"repo":   repo,
				"ref":    ref,
				"sha":    sha,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{"ref": ref, "sha": sha}
		var data map[string]any
		if _, err := doGitHub(client, http.MethodPost, fmt.Sprintf("/repos/%s/%s/git/refs", owner, repo), body, &data); err != nil {
			return fmt.Errorf("creating ref %s in %s/%s: %w", ref, owner, repo, err)
		}

		info := toRefInfo(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}
		fmt.Printf("Created ref: %s → %s\n", info.Ref, info.Object.SHA)
		return nil
	}
}

func newRefsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a Git reference",
		RunE:  makeRunRefsUpdate(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("ref", "", "Reference name without refs/ prefix (e.g. heads/main) (required)")
	cmd.Flags().String("sha", "", "New SHA1 to point the reference to (required)")
	cmd.Flags().Bool("force", false, "Force update (non-fast-forward)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("ref")
	_ = cmd.MarkFlagRequired("sha")
	return cmd
}

func makeRunRefsUpdate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		ref, _ := cmd.Flags().GetString("ref")
		sha, _ := cmd.Flags().GetString("sha")
		force, _ := cmd.Flags().GetBool("force")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update ref %s to %s in %s/%s (force=%v)", ref, sha, owner, repo, force), map[string]any{
				"action": "update_ref",
				"owner":  owner,
				"repo":   repo,
				"ref":    ref,
				"sha":    sha,
				"force":  force,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{"sha": sha, "force": force}
		var data map[string]any
		if _, err := doGitHub(client, http.MethodPatch, fmt.Sprintf("/repos/%s/%s/git/refs/%s", owner, repo, ref), body, &data); err != nil {
			return fmt.Errorf("updating ref %s in %s/%s: %w", ref, owner, repo, err)
		}

		info := toRefInfo(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}
		fmt.Printf("Updated ref: %s → %s\n", info.Ref, info.Object.SHA)
		return nil
	}
}

func newRefsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Git reference",
		RunE:  makeRunRefsDelete(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("ref", "", "Reference name without refs/ prefix (e.g. heads/old-branch) (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("ref")
	return cmd
}

func makeRunRefsDelete(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		ref, _ := cmd.Flags().GetString("ref")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete ref %s from %s/%s", ref, owner, repo), map[string]any{
				"action": "delete_ref",
				"owner":  owner,
				"repo":   repo,
				"ref":    ref,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		if _, err := doGitHub(client, http.MethodDelete, fmt.Sprintf("/repos/%s/%s/git/refs/%s", owner, repo, ref), nil, nil); err != nil {
			return fmt.Errorf("deleting ref %s from %s/%s: %w", ref, owner, repo, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "ref": ref})
		}
		fmt.Printf("Deleted ref: %s\n", ref)
		return nil
	}
}

// --- Commits commands ---

func newGitCommitsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a Git commit object by SHA",
		RunE:  makeRunGitCommitsGet(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("sha", "", "Commit SHA (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("sha")
	return cmd
}

func makeRunGitCommitsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		sha, _ := cmd.Flags().GetString("sha")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodGet, fmt.Sprintf("/repos/%s/%s/git/commits/%s", owner, repo, sha), nil, &data); err != nil {
			return fmt.Errorf("getting commit %s in %s/%s: %w", sha, owner, repo, err)
		}

		info := toGitCommitInfo(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		parentSHAs := make([]string, 0, len(info.Parents))
		for _, p := range info.Parents {
			parentSHAs = append(parentSHAs, p.SHA)
		}

		lines := []string{
			fmt.Sprintf("SHA:      %s", info.SHA),
			fmt.Sprintf("Message:  %s", info.Message),
			fmt.Sprintf("Tree:     %s", info.Tree.SHA),
			fmt.Sprintf("Author:   %s <%s> at %s", info.Author.Name, info.Author.Email, info.Author.Date),
			fmt.Sprintf("Parents:  %s", strings.Join(parentSHAs, ", ")),
			fmt.Sprintf("URL:      %s", info.URL),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newGitCommitsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Git commit object",
		RunE:  makeRunGitCommitsCreate(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("message", "", "Commit message (required)")
	cmd.Flags().String("tree", "", "SHA of the tree object (required)")
	cmd.Flags().String("parents", "", "Comma-separated parent commit SHAs")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("message")
	_ = cmd.MarkFlagRequired("tree")
	return cmd
}

func makeRunGitCommitsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		message, _ := cmd.Flags().GetString("message")
		tree, _ := cmd.Flags().GetString("tree")
		parentsRaw, _ := cmd.Flags().GetString("parents")

		var parents []string
		if parentsRaw != "" {
			for _, p := range strings.Split(parentsRaw, ",") {
				if p = strings.TrimSpace(p); p != "" {
					parents = append(parents, p)
				}
			}
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create commit %q on tree %s in %s/%s", message, tree, owner, repo), map[string]any{
				"action":  "create_commit",
				"owner":   owner,
				"repo":    repo,
				"message": message,
				"tree":    tree,
				"parents": parents,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"message": message,
			"tree":    tree,
			"parents": parents,
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodPost, fmt.Sprintf("/repos/%s/%s/git/commits", owner, repo), body, &data); err != nil {
			return fmt.Errorf("creating commit in %s/%s: %w", owner, repo, err)
		}

		info := toGitCommitInfo(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}
		fmt.Printf("Created commit: %s\n", info.SHA)
		return nil
	}
}

// --- Trees commands ---

func newGitTreesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a Git tree by SHA",
		RunE:  makeRunGitTreesGet(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("sha", "", "Tree SHA (required)")
	cmd.Flags().Bool("recursive", false, "Recursively retrieve the tree")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("sha")
	return cmd
}

func makeRunGitTreesGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		sha, _ := cmd.Flags().GetString("sha")
		recursive, _ := cmd.Flags().GetBool("recursive")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/git/trees/%s", owner, repo, sha)
		if recursive {
			path += "?recursive=1"
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodGet, path, nil, &data); err != nil {
			return fmt.Errorf("getting tree %s in %s/%s: %w", sha, owner, repo, err)
		}

		info := toGitTreeInfo(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("SHA:       %s", info.SHA),
			fmt.Sprintf("URL:       %s", info.URL),
			fmt.Sprintf("Truncated: %v", info.Truncated),
			fmt.Sprintf("Entries:   %d", len(info.Tree)),
		}
		if len(info.Tree) > 0 {
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("%-10s  %-6s  %-50s  %s", "MODE", "TYPE", "PATH", "SHA"))
			for _, e := range info.Tree {
				lines = append(lines, fmt.Sprintf("%-10s  %-6s  %-50s  %s", e.Mode, e.Type, truncate(e.Path, 50), e.SHA))
			}
		}
		cli.PrintText(lines)
		return nil
	}
}

func newGitTreesCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Git tree",
		RunE:  makeRunGitTreesCreate(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("tree", "", "JSON array of tree entries (e.g. [{\"path\":\"file.txt\",\"mode\":\"100644\",\"type\":\"blob\",\"sha\":\"...\"}])")
	cmd.Flags().String("tree-file", "", "Path to JSON file containing tree entries")
	cmd.Flags().String("base-tree", "", "SHA of the base tree to build upon (optional)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	return cmd
}

func makeRunGitTreesCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		treeJSON, _ := cmd.Flags().GetString("tree")
		treeFile, _ := cmd.Flags().GetString("tree-file")
		baseTree, _ := cmd.Flags().GetString("base-tree")

		if treeJSON == "" && treeFile == "" {
			return fmt.Errorf("one of --tree or --tree-file is required")
		}

		var rawTree string
		if treeFile != "" {
			data, err := os.ReadFile(treeFile)
			if err != nil {
				return fmt.Errorf("reading tree file %s: %w", treeFile, err)
			}
			rawTree = string(data)
		} else {
			rawTree = treeJSON
		}

		var treeEntries []any
		if err := json.Unmarshal([]byte(rawTree), &treeEntries); err != nil {
			return fmt.Errorf("parsing tree JSON: %w", err)
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create tree with %d entries in %s/%s", len(treeEntries), owner, repo), map[string]any{
				"action":   "create_tree",
				"owner":    owner,
				"repo":     repo,
				"entries":  len(treeEntries),
				"baseTree": baseTree,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{"tree": treeEntries}
		if baseTree != "" {
			body["base_tree"] = baseTree
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodPost, fmt.Sprintf("/repos/%s/%s/git/trees", owner, repo), body, &data); err != nil {
			return fmt.Errorf("creating tree in %s/%s: %w", owner, repo, err)
		}

		info := toGitTreeInfo(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}
		fmt.Printf("Created tree: %s (%d entries)\n", info.SHA, len(info.Tree))
		return nil
	}
}

// --- Blobs commands ---

func newGitBlobsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a Git blob by SHA",
		RunE:  makeRunGitBlobsGet(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("sha", "", "Blob SHA (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("sha")
	return cmd
}

func makeRunGitBlobsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		sha, _ := cmd.Flags().GetString("sha")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodGet, fmt.Sprintf("/repos/%s/%s/git/blobs/%s", owner, repo, sha), nil, &data); err != nil {
			return fmt.Errorf("getting blob %s in %s/%s: %w", sha, owner, repo, err)
		}

		info := toGitBlobInfo(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("SHA:      %s", info.SHA),
			fmt.Sprintf("Size:     %d", info.Size),
			fmt.Sprintf("Encoding: %s", info.Encoding),
			fmt.Sprintf("URL:      %s", info.URL),
		}
		if info.Content != "" {
			lines = append(lines, fmt.Sprintf("Content:  %s", truncate(info.Content, 80)))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newGitBlobsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Git blob",
		RunE:  makeRunGitBlobsCreate(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("content", "", "Blob content (required)")
	cmd.Flags().String("encoding", "utf-8", "Content encoding: utf-8 or base64")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("content")
	return cmd
}

func makeRunGitBlobsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		content, _ := cmd.Flags().GetString("content")
		encoding, _ := cmd.Flags().GetString("encoding")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create blob (%d bytes, encoding=%s) in %s/%s", len(content), encoding, owner, repo), map[string]any{
				"action":   "create_blob",
				"owner":    owner,
				"repo":     repo,
				"size":     len(content),
				"encoding": encoding,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{"content": content, "encoding": encoding}
		var data map[string]any
		if _, err := doGitHub(client, http.MethodPost, fmt.Sprintf("/repos/%s/%s/git/blobs", owner, repo), body, &data); err != nil {
			return fmt.Errorf("creating blob in %s/%s: %w", owner, repo, err)
		}

		info := toGitBlobInfo(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}
		fmt.Printf("Created blob: %s\n", info.SHA)
		return nil
	}
}

// --- Tags commands ---

func newGitTagsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a Git tag object by SHA",
		RunE:  makeRunGitTagsGet(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("sha", "", "Tag object SHA (required)")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("sha")
	return cmd
}

func makeRunGitTagsGet(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		sha, _ := cmd.Flags().GetString("sha")

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodGet, fmt.Sprintf("/repos/%s/%s/git/tags/%s", owner, repo, sha), nil, &data); err != nil {
			return fmt.Errorf("getting tag %s in %s/%s: %w", sha, owner, repo, err)
		}

		info := toGitTagInfo(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("Tag:     %s", info.Tag),
			fmt.Sprintf("SHA:     %s", info.SHA),
			fmt.Sprintf("Message: %s", info.Message),
			fmt.Sprintf("Tagger:  %s <%s> at %s", info.Tagger.Name, info.Tagger.Email, info.Tagger.Date),
			fmt.Sprintf("Object:  %s (%s)", info.Object.SHA, info.Object.Type),
			fmt.Sprintf("URL:     %s", info.URL),
		}
		cli.PrintText(lines)
		return nil
	}
}

func newGitTagsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Git tag object",
		RunE:  makeRunGitTagsCreate(factory),
	}
	cmd.Flags().String("owner", "", "Repository owner (required)")
	cmd.Flags().String("repo", "", "Repository name (required)")
	cmd.Flags().String("tag", "", "Tag name (required)")
	cmd.Flags().String("message", "", "Tag message (required)")
	cmd.Flags().String("object", "", "SHA of the object being tagged (required)")
	cmd.Flags().String("type", "commit", "Type of the object being tagged: commit, tree, or blob")
	_ = cmd.MarkFlagRequired("owner")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("tag")
	_ = cmd.MarkFlagRequired("message")
	_ = cmd.MarkFlagRequired("object")
	return cmd
}

func makeRunGitTagsCreate(factory ClientFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		owner, _ := cmd.Flags().GetString("owner")
		repo, _ := cmd.Flags().GetString("repo")
		tag, _ := cmd.Flags().GetString("tag")
		message, _ := cmd.Flags().GetString("message")
		object, _ := cmd.Flags().GetString("object")
		objectType, _ := cmd.Flags().GetString("type")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create tag %q pointing to %s (%s) in %s/%s", tag, object, objectType, owner, repo), map[string]any{
				"action":  "create_tag",
				"owner":   owner,
				"repo":    repo,
				"tag":     tag,
				"message": message,
				"object":  object,
				"type":    objectType,
			})
		}

		ctx := cmd.Context()
		client, err := factory(ctx)
		if err != nil {
			return err
		}

		body := map[string]any{
			"tag":     tag,
			"message": message,
			"object":  object,
			"type":    objectType,
		}

		var data map[string]any
		if _, err := doGitHub(client, http.MethodPost, fmt.Sprintf("/repos/%s/%s/git/tags", owner, repo), body, &data); err != nil {
			return fmt.Errorf("creating tag %s in %s/%s: %w", tag, owner, repo, err)
		}

		info := toGitTagInfo(data)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}
		fmt.Printf("Created tag: %s → %s (%s)\n", info.Tag, info.Object.SHA, info.Object.Type)
		return nil
	}
}
