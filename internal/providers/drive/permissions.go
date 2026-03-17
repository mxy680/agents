package drive

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/drive/v3"
)

func newPermissionsListCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List permissions on a file",
		RunE:  makeRunPermissionsList(factory),
	}
	cmd.Flags().String("file-id", "", "File ID (required)")
	_ = cmd.MarkFlagRequired("file-id")
	return cmd
}

func makeRunPermissionsList(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		fileID, _ := cmd.Flags().GetString("file-id")

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Permissions.List(fileID).
			Fields("permissions(id,role,type,emailAddress,domain,displayName)").
			SupportsAllDrives(true).
			Do()
		if err != nil {
			return fmt.Errorf("listing permissions for %s: %w", fileID, err)
		}

		perms := make([]PermissionInfo, 0, len(resp.Permissions))
		for _, p := range resp.Permissions {
			perms = append(perms, toPermissionInfo(p))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(perms)
		}

		if len(perms) == 0 {
			fmt.Println("No permissions found.")
			return nil
		}

		lines := make([]string, 0, len(perms)+1)
		lines = append(lines, fmt.Sprintf("%-20s  %-10s  %-10s  %s", "ID", "ROLE", "TYPE", "EMAIL/DOMAIN"))
		for _, p := range perms {
			target := p.EmailAddress
			if target == "" {
				target = p.Domain
			}
			lines = append(lines, fmt.Sprintf("%-20s  %-10s  %-10s  %s", truncate(p.ID, 20), p.Role, p.Type, target))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newPermissionsGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a specific permission on a file",
		RunE:  makeRunPermissionsGet(factory),
	}
	cmd.Flags().String("file-id", "", "File ID (required)")
	cmd.Flags().String("permission-id", "", "Permission ID (required)")
	_ = cmd.MarkFlagRequired("file-id")
	_ = cmd.MarkFlagRequired("permission-id")
	return cmd
}

func makeRunPermissionsGet(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		fileID, _ := cmd.Flags().GetString("file-id")
		permissionID, _ := cmd.Flags().GetString("permission-id")

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		perm, err := svc.Permissions.Get(fileID, permissionID).
			Fields("id,role,type,emailAddress,domain,displayName").
			SupportsAllDrives(true).
			Do()
		if err != nil {
			return fmt.Errorf("getting permission %s on file %s: %w", permissionID, fileID, err)
		}

		info := toPermissionInfo(perm)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("ID:      %s", info.ID),
			fmt.Sprintf("Role:    %s", info.Role),
			fmt.Sprintf("Type:    %s", info.Type),
		}
		if info.EmailAddress != "" {
			lines = append(lines, fmt.Sprintf("Email:   %s", info.EmailAddress))
		}
		if info.Domain != "" {
			lines = append(lines, fmt.Sprintf("Domain:  %s", info.Domain))
		}
		if info.DisplayName != "" {
			lines = append(lines, fmt.Sprintf("Name:    %s", info.DisplayName))
		}
		cli.PrintText(lines)
		return nil
	}
}

func newPermissionsCreateCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a permission (share a file)",
		RunE:  makeRunPermissionsCreate(factory),
	}
	cmd.Flags().String("file-id", "", "File ID (required)")
	cmd.Flags().String("role", "", "Permission role: owner, organizer, fileOrganizer, writer, commenter, reader (required)")
	cmd.Flags().String("type", "", "Grantee type: user, group, domain, anyone (required)")
	cmd.Flags().String("email", "", "Email address (for user/group type)")
	cmd.Flags().String("domain", "", "Domain (for domain type)")
	_ = cmd.MarkFlagRequired("file-id")
	_ = cmd.MarkFlagRequired("role")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func makeRunPermissionsCreate(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		fileID, _ := cmd.Flags().GetString("file-id")
		role, _ := cmd.Flags().GetString("role")
		permType, _ := cmd.Flags().GetString("type")
		email, _ := cmd.Flags().GetString("email")
		domain, _ := cmd.Flags().GetString("domain")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would share file %s with %s=%s role=%s", fileID, permType, email+domain, role), map[string]any{
				"action": "create_permission",
				"fileId": fileID,
				"role":   role,
				"type":   permType,
				"email":  email,
				"domain": domain,
			})
		}

		ctx := cmd.Context()
		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		perm := &api.Permission{
			Role:         role,
			Type:         permType,
			EmailAddress: email,
			Domain:       domain,
		}

		created, err := svc.Permissions.Create(fileID, perm).
			SupportsAllDrives(true).
			Do()
		if err != nil {
			return fmt.Errorf("creating permission on file %s: %w", fileID, err)
		}

		info := toPermissionInfo(created)
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}
		fmt.Printf("Created permission: %s (role=%s, type=%s)\n", info.ID, info.Role, info.Type)
		return nil
	}
}

func newPermissionsDeleteCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a permission from a file",
		RunE:  makeRunPermissionsDelete(factory),
	}
	cmd.Flags().String("file-id", "", "File ID (required)")
	cmd.Flags().String("permission-id", "", "Permission ID to delete (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("file-id")
	_ = cmd.MarkFlagRequired("permission-id")
	return cmd
}

func makeRunPermissionsDelete(factory ServiceFactory) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		fileID, _ := cmd.Flags().GetString("file-id")
		permissionID, _ := cmd.Flags().GetString("permission-id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would delete permission %s from file %s", permissionID, fileID), map[string]any{
				"action":       "delete_permission",
				"fileId":       fileID,
				"permissionId": permissionID,
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

		if err := svc.Permissions.Delete(fileID, permissionID).
			SupportsAllDrives(true).
			Do(); err != nil {
			return fmt.Errorf("deleting permission %s from file %s: %w", permissionID, fileID, err)
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(map[string]string{"status": "deleted", "fileId": fileID, "permissionId": permissionID})
		}
		fmt.Printf("Deleted permission: %s\n", permissionID)
		return nil
	}
}
