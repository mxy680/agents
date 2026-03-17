package drive

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/drive/v3"
)

func newFilesListCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List files in Drive",
		RunE:  makeRunFilesList(factory),
	}
	cmd.Flags().String("query", "", "Drive search query (e.g. name contains 'report')")
	cmd.Flags().Int("limit", 20, "Maximum number of files to return")
	cmd.Flags().String("page-token", "", "Token for next page of results")
	cmd.Flags().String("order-by", "", "Sort order (e.g. modifiedTime desc, name)")
	cmd.Flags().String("corpora", "", "Corpora to search: user, drive, allDrives")
	cmd.Flags().String("drive-id", "", "Shared drive ID (used with --corpora=drive)")
	cmd.Flags().Bool("include-trashed", false, "Include trashed files in results")
	return cmd
}

func makeRunFilesList(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		query, _ := cmd.Flags().GetString("query")
		limit, _ := cmd.Flags().GetInt("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")
		orderBy, _ := cmd.Flags().GetString("order-by")
		corpora, _ := cmd.Flags().GetString("corpora")
		driveID, _ := cmd.Flags().GetString("drive-id")
		includeTrashed, _ := cmd.Flags().GetBool("include-trashed")

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		// Build query: auto-exclude trashed unless requested
		var queryParts []string
		if !includeTrashed {
			queryParts = append(queryParts, "trashed = false")
		}
		if query != "" {
			queryParts = append(queryParts, query)
		}
		fullQuery := strings.Join(queryParts, " and ")

		req := svc.Files.List().
			PageSize(int64(limit)).
			Fields("files(id,name,mimeType,size,modifiedTime,parents,trashed),nextPageToken")
		if fullQuery != "" {
			req = req.Q(fullQuery)
		}
		if pageToken != "" {
			req = req.PageToken(pageToken)
		}
		if orderBy != "" {
			req = req.OrderBy(orderBy)
		}
		if corpora != "" {
			req = req.Corpora(corpora)
		}
		if driveID != "" {
			req = req.DriveId(driveID).IncludeItemsFromAllDrives(true).SupportsAllDrives(true)
		}

		resp, err := req.Do()
		if err != nil {
			return fmt.Errorf("listing files: %w", err)
		}

		summaries := make([]FileSummary, 0, len(resp.Files))
		for _, f := range resp.Files {
			summaries = append(summaries, toFileSummary(f))
		}
		return printFileSummaries(cmd, summaries)
	}
}

func newFilesGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get file metadata by ID",
		RunE:  makeRunFilesGet(factory),
	}
	cmd.Flags().String("file-id", "", "File ID (required)")
	_ = cmd.MarkFlagRequired("file-id")
	return cmd
}

func makeRunFilesGet(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		fileID, _ := cmd.Flags().GetString("file-id")

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		file, err := svc.Files.Get(fileID).
			Fields("id,name,mimeType,size,modifiedTime,createdTime,parents,trashed,description,webViewLink,webContentLink,shared,owners(emailAddress,displayName)").
			SupportsAllDrives(true).
			Do()
		if err != nil {
			return fmt.Errorf("getting file %s: %w", fileID, err)
		}

		detail := toFileDetail(file)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}

		lines := []string{
			fmt.Sprintf("ID:           %s", detail.ID),
			fmt.Sprintf("Name:         %s", detail.Name),
			fmt.Sprintf("Type:         %s", detail.MimeType),
			fmt.Sprintf("Size:         %s", formatSize(detail.Size)),
			fmt.Sprintf("Modified:     %s", detail.ModifiedTime),
			fmt.Sprintf("Created:      %s", detail.CreatedTime),
			fmt.Sprintf("Description:  %s", detail.Description),
			fmt.Sprintf("Shared:       %v", detail.Shared),
		}
		if detail.WebViewLink != "" {
			lines = append(lines, fmt.Sprintf("View Link:    %s", detail.WebViewLink))
		}
		for _, o := range detail.Owners {
			lines = append(lines, fmt.Sprintf("Owner:        %s (%s)", o.Email, o.DisplayName))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newFilesDownloadCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download file content",
		RunE:  makeRunFilesDownload(factory),
	}
	cmd.Flags().String("file-id", "", "File ID (required)")
	cmd.Flags().String("output", "", "Output file path (default: stdout)")
	cmd.Flags().String("export-mime", "", "Export MIME type for Google Workspace files (e.g. application/pdf)")
	_ = cmd.MarkFlagRequired("file-id")
	return cmd
}

// maxDownloadBytes is the maximum file size allowed for downloads (500 MB).
const maxDownloadBytes = 500 * 1024 * 1024

func makeRunFilesDownload(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		fileID, _ := cmd.Flags().GetString("file-id")
		output, _ := cmd.Flags().GetString("output")
		exportMime, _ := cmd.Flags().GetString("export-mime")

		// Validate output path to prevent path traversal
		if output != "" {
			absOut, err := filepath.Abs(output)
			if err != nil {
				return fmt.Errorf("invalid output path: %w", err)
			}
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			if !strings.HasPrefix(absOut, cwd+string(os.PathSeparator)) && absOut != cwd {
				return fmt.Errorf("output path %q must be within the working directory", output)
			}
			output = absOut
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		var body io.ReadCloser
		if exportMime != "" {
			httpResp, err := svc.Files.Export(fileID, exportMime).Download()
			if err != nil {
				return fmt.Errorf("exporting file %s: %w", fileID, err)
			}
			body = httpResp.Body
		} else {
			httpResp, err := svc.Files.Get(fileID).SupportsAllDrives(true).Download()
			if err != nil {
				return fmt.Errorf("downloading file %s: %w", fileID, err)
			}
			body = httpResp.Body
		}
		defer body.Close()

		var dst io.Writer
		if output != "" {
			f, err := os.Create(output)
			if err != nil {
				return fmt.Errorf("creating output file: %w", err)
			}
			defer f.Close()
			dst = f
		} else {
			dst = os.Stdout
		}

		limited := io.LimitReader(body, maxDownloadBytes+1)
		n, err := io.Copy(dst, limited)
		if err != nil {
			return fmt.Errorf("writing file content: %w", err)
		}
		if n > maxDownloadBytes {
			return fmt.Errorf("file exceeds maximum download size (%s)", formatSize(maxDownloadBytes))
		}

		// Print download info to stderr so it doesn't mix with file content on stdout
		if output != "" {
			fmt.Fprintf(os.Stderr, "Downloaded %s to %s\n", formatSize(n), output)
		}
		return nil
	}
}

func newFilesUploadCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload a file to Drive",
		RunE:  makeRunFilesUpload(factory),
	}
	cmd.Flags().String("path", "", "Local file path to upload (required)")
	cmd.Flags().String("name", "", "Name for the file in Drive (default: local filename)")
	cmd.Flags().String("parent", "", "Parent folder ID")
	cmd.Flags().String("mime-type", "", "MIME type override")
	cmd.Flags().String("description", "", "File description")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func makeRunFilesUpload(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		path, _ := cmd.Flags().GetString("path")
		name, _ := cmd.Flags().GetString("name")
		parent, _ := cmd.Flags().GetString("parent")
		mimeType, _ := cmd.Flags().GetString("mime-type")
		description, _ := cmd.Flags().GetString("description")

		// Default name to filename from path
		if name == "" {
			parts := strings.Split(path, "/")
			name = parts[len(parts)-1]
		}

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would upload %q as %q", path, name), map[string]any{
				"action":   "upload",
				"path":     path,
				"name":     name,
				"parent":   parent,
				"mimeType": mimeType,
			})
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("opening file %s: %w", path, err)
		}
		defer f.Close()

		fileMeta := &api.File{
			Name:        name,
			Description: description,
		}
		if parent != "" {
			fileMeta.Parents = []string{parent}
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		createCall := svc.Files.Create(fileMeta).Media(f).SupportsAllDrives(true)

		created, err := createCall.Do()
		if err != nil {
			return fmt.Errorf("uploading file: %w", err)
		}

		detail := toFileDetail(created)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Uploaded: %s (%s)\n", detail.Name, detail.ID)
		return nil
	}
}

func newFilesCopyCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy a file",
		RunE:  makeRunFilesCopy(factory),
	}
	cmd.Flags().String("file-id", "", "File ID to copy (required)")
	cmd.Flags().String("name", "", "Name for the copy")
	cmd.Flags().String("parent", "", "Parent folder ID for the copy")
	_ = cmd.MarkFlagRequired("file-id")
	return cmd
}

func makeRunFilesCopy(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		fileID, _ := cmd.Flags().GetString("file-id")
		name, _ := cmd.Flags().GetString("name")
		parent, _ := cmd.Flags().GetString("parent")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would copy file %s", fileID), map[string]any{
				"action": "copy",
				"fileId": fileID,
				"name":   name,
				"parent": parent,
			})
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		copyMeta := &api.File{}
		if name != "" {
			copyMeta.Name = name
		}
		if parent != "" {
			copyMeta.Parents = []string{parent}
		}

		copied, err := svc.Files.Copy(fileID, copyMeta).SupportsAllDrives(true).Do()
		if err != nil {
			return fmt.Errorf("copying file %s: %w", fileID, err)
		}

		detail := toFileDetail(copied)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Copied: %s → %s (%s)\n", fileID, detail.Name, detail.ID)
		return nil
	}
}

func newFilesMoveCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move a file to a different folder",
		RunE:  makeRunFilesMove(factory),
	}
	cmd.Flags().String("file-id", "", "File ID to move (required)")
	cmd.Flags().String("parent", "", "Destination folder ID (required)")
	_ = cmd.MarkFlagRequired("file-id")
	_ = cmd.MarkFlagRequired("parent")
	return cmd
}

func makeRunFilesMove(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		fileID, _ := cmd.Flags().GetString("file-id")
		newParent, _ := cmd.Flags().GetString("parent")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would move file %s to folder %s", fileID, newParent), map[string]any{
				"action":    "move",
				"fileId":    fileID,
				"newParent": newParent,
			})
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		// Get current parents to remove them
		file, err := svc.Files.Get(fileID).Fields("parents").SupportsAllDrives(true).Do()
		if err != nil {
			return fmt.Errorf("getting file %s parents: %w", fileID, err)
		}

		oldParents := strings.Join(file.Parents, ",")
		moved, err := svc.Files.Update(fileID, nil).
			AddParents(newParent).
			RemoveParents(oldParents).
			SupportsAllDrives(true).
			Do()
		if err != nil {
			return fmt.Errorf("moving file %s: %w", fileID, err)
		}

		detail := toFileDetail(moved)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(detail)
		}
		fmt.Printf("Moved: %s → folder %s\n", detail.Name, newParent)
		return nil
	}
}

func newFilesTrashCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trash",
		Short: "Move a file to trash",
		RunE:  makeRunFilesTrash(factory),
	}
	cmd.Flags().String("file-id", "", "File ID to trash (required)")
	_ = cmd.MarkFlagRequired("file-id")
	return cmd
}

func makeRunFilesTrash(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		fileID, _ := cmd.Flags().GetString("file-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would trash file %s", fileID), map[string]any{
				"action": "trash",
				"fileId": fileID,
			})
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		trashed, err := svc.Files.Update(fileID, &api.File{Trashed: true}).
			SupportsAllDrives(true).
			Do()
		if err != nil {
			return fmt.Errorf("trashing file %s: %w", fileID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "trashed", "fileId": trashed.Id, "name": trashed.Name})
		}
		fmt.Printf("Trashed: %s (%s)\n", trashed.Name, trashed.Id)
		return nil
	}
}

func newFilesUntrashCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "untrash",
		Short: "Restore a file from trash",
		RunE:  makeRunFilesUntrash(factory),
	}
	cmd.Flags().String("file-id", "", "File ID to restore (required)")
	_ = cmd.MarkFlagRequired("file-id")
	return cmd
}

func makeRunFilesUntrash(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		fileID, _ := cmd.Flags().GetString("file-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would untrash file %s", fileID), map[string]any{
				"action": "untrash",
				"fileId": fileID,
			})
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		// ForceSendFields is required because false is the zero value for bool
		restored, err := svc.Files.Update(fileID, &api.File{
			Trashed:         false,
			ForceSendFields: []string{"Trashed"},
		}).SupportsAllDrives(true).Do()
		if err != nil {
			return fmt.Errorf("untrashing file %s: %w", fileID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "restored", "fileId": restored.Id, "name": restored.Name})
		}
		fmt.Printf("Restored: %s (%s)\n", restored.Name, restored.Id)
		return nil
	}
}

func newFilesDeleteCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Permanently delete a file (bypasses trash)",
		RunE:  makeRunFilesDelete(factory),
	}
	cmd.Flags().String("file-id", "", "File ID to delete (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("file-id")
	return cmd
}

func makeRunFilesDelete(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		fileID, _ := cmd.Flags().GetString("file-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would permanently delete file %s", fileID), map[string]any{
				"action": "delete",
				"fileId": fileID,
			})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		if err := svc.Files.Delete(fileID).SupportsAllDrives(true).Do(); err != nil {
			return fmt.Errorf("deleting file %s: %w", fileID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "fileId": fileID})
		}
		fmt.Printf("Deleted: %s\n", fileID)
		return nil
	}
}
