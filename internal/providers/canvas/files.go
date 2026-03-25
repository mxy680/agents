package canvas

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newFilesCmd returns the parent "files" command with all subcommands attached.
func newFilesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "files",
		Short:   "Manage Canvas files and folders",
		Aliases: []string{"file", "f"},
	}

	cmd.AddCommand(newFilesListCmd(factory))
	cmd.AddCommand(newFilesGetCmd(factory))
	cmd.AddCommand(newFilesDownloadCmd(factory))
	cmd.AddCommand(newFilesFoldersCmd(factory))
	cmd.AddCommand(newFilesFolderContentsCmd(factory))

	return cmd
}

func newFilesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List files for a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}

			search, _ := cmd.Flags().GetString("search")
			contentTypes, _ := cmd.Flags().GetString("content-types")
			sort, _ := cmd.Flags().GetString("sort")
			limit, _ := cmd.Flags().GetInt("limit")

			params := url.Values{}
			if search != "" {
				params.Set("search_term", search)
			}
			if contentTypes != "" {
				for _, ct := range strings.Split(contentTypes, ",") {
					params.Add("content_types[]", strings.TrimSpace(ct))
				}
			}
			if sort != "" {
				params.Set("sort", sort)
			}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/files", params)
			if err != nil {
				return err
			}

			var files []FileSummary
			if err := json.Unmarshal(data, &files); err != nil {
				return fmt.Errorf("parse files: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(files)
			}

			if len(files) == 0 {
				fmt.Println("No files found.")
				return nil
			}
			for _, f := range files {
				locked := " "
				if f.Locked {
					locked = "L"
				}
				fmt.Printf("%-8d  [%s]  %-10s  %s\n", f.ID, locked, formatSize(f.Size), truncate(f.DisplayName, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("search", "", "Search files by name")
	cmd.Flags().String("content-types", "", "Comma-separated MIME types to filter by")
	cmd.Flags().String("sort", "", "Sort by: name|size|created_at|updated_at")
	cmd.Flags().Int("limit", 0, "Maximum number of files to return")
	return cmd
}

func newFilesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific file",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			fileID, _ := cmd.Flags().GetString("file-id")
			if fileID == "" {
				return fmt.Errorf("--file-id is required")
			}

			data, err := client.Get(ctx, "/files/"+fileID, nil)
			if err != nil {
				return err
			}

			var file FileSummary
			if err := json.Unmarshal(data, &file); err != nil {
				return fmt.Errorf("parse file: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(file)
			}

			fmt.Printf("ID:           %d\n", file.ID)
			fmt.Printf("Name:         %s\n", file.DisplayName)
			fmt.Printf("Filename:     %s\n", file.Filename)
			fmt.Printf("Type:         %s\n", file.ContentType)
			fmt.Printf("Size:         %s\n", formatSize(file.Size))
			fmt.Printf("Locked:       %v\n", file.Locked)
			fmt.Printf("Hidden:       %v\n", file.Hidden)
			if file.CreatedAt != "" {
				fmt.Printf("Created:      %s\n", file.CreatedAt)
			}
			if file.UpdatedAt != "" {
				fmt.Printf("Updated:      %s\n", file.UpdatedAt)
			}
			return nil
		},
	}

	cmd.Flags().String("file-id", "", "Canvas file ID (required)")
	return cmd
}

func newFilesDownloadCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download a file to disk",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			fileID, _ := cmd.Flags().GetString("file-id")
			output, _ := cmd.Flags().GetString("output")
			if fileID == "" {
				return fmt.Errorf("--file-id is required")
			}
			if output == "" {
				return fmt.Errorf("--output is required")
			}

			// First, fetch file metadata to get the download URL.
			data, err := client.Get(ctx, "/files/"+fileID, nil)
			if err != nil {
				return err
			}

			var file FileSummary
			if err := json.Unmarshal(data, &file); err != nil {
				return fmt.Errorf("parse file metadata: %w", err)
			}

			if file.URL == "" {
				return fmt.Errorf("no download URL returned for file %s", fileID)
			}

			// The download URL is a full URL, not a Canvas API path — fetch it directly.
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, file.URL, nil)
			if err != nil {
				return fmt.Errorf("build download request: %w", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("download file: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("download failed with HTTP %d", resp.StatusCode)
			}

			f, err := os.Create(output)
			if err != nil {
				return fmt.Errorf("create output file: %w", err)
			}
			defer f.Close()

			n, err := io.Copy(f, resp.Body)
			if err != nil {
				return fmt.Errorf("write file: %w", err)
			}

			fmt.Printf("Downloaded %s (%s) to %s\n", file.DisplayName, formatSize(n), output)
			return nil
		},
	}

	cmd.Flags().String("file-id", "", "Canvas file ID (required)")
	cmd.Flags().String("output", "", "Output file path (required)")
	return cmd
}

func newFilesFoldersCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "folders",
		Short: "List all folders for a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/folders", nil)
			if err != nil {
				return err
			}

			var folders []FolderSummary
			if err := json.Unmarshal(data, &folders); err != nil {
				return fmt.Errorf("parse folders: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(folders)
			}

			if len(folders) == 0 {
				fmt.Println("No folders found.")
				return nil
			}
			for _, folder := range folders {
				fmt.Printf("%-8d  files:%-4d  folders:%-4d  %s\n",
					folder.ID, folder.FilesCount, folder.FoldersCount, folder.FullName)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	return cmd
}

func newFilesFolderContentsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "folder-contents",
		Short: "List files in a specific folder",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			folderID, _ := cmd.Flags().GetString("folder-id")
			if folderID == "" {
				return fmt.Errorf("--folder-id is required")
			}

			limit, _ := cmd.Flags().GetInt("limit")
			params := url.Values{}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/folders/"+folderID+"/files", params)
			if err != nil {
				return err
			}

			var files []FileSummary
			if err := json.Unmarshal(data, &files); err != nil {
				return fmt.Errorf("parse files: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(files)
			}

			if len(files) == 0 {
				fmt.Println("No files found in folder.")
				return nil
			}
			for _, f := range files {
				locked := " "
				if f.Locked {
					locked = "L"
				}
				fmt.Printf("%-8d  [%s]  %-10s  %s\n", f.ID, locked, formatSize(f.Size), truncate(f.DisplayName, 60))
			}
			return nil
		},
	}

	cmd.Flags().String("folder-id", "", "Canvas folder ID (required)")
	cmd.Flags().Int("limit", 0, "Maximum number of files to return")
	return cmd
}
