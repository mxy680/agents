package gmail

import (
	"context"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
	api "google.golang.org/api/gmail/v1"
)

// LabelInfo is the JSON-serializable representation of a Gmail label.
type LabelInfo struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Type                  string `json:"type"`
	MessagesTotal         int64  `json:"messagesTotal"`
	MessagesUnread        int64  `json:"messagesUnread"`
	ThreadsTotal          int64  `json:"threadsTotal"`
	ThreadsUnread         int64  `json:"threadsUnread"`
	LabelListVisibility   string `json:"labelListVisibility,omitempty"`
	MessageListVisibility string `json:"messageListVisibility,omitempty"`
}

// labelFromAPI converts a Gmail API Label to LabelInfo.
func labelFromAPI(l *api.Label) LabelInfo {
	return LabelInfo{
		ID:                    l.Id,
		Name:                  l.Name,
		Type:                  l.Type,
		MessagesTotal:         l.MessagesTotal,
		MessagesUnread:        l.MessagesUnread,
		ThreadsTotal:          l.ThreadsTotal,
		ThreadsUnread:         l.ThreadsUnread,
		LabelListVisibility:   l.LabelListVisibility,
		MessageListVisibility: l.MessageListVisibility,
	}
}

// newLabelsListCmd returns the `labels list` command.
func newLabelsListCmd(factory ServiceFactory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all labels in the mailbox",
		RunE:  makeRunLabelsList(factory),
	}
}

func makeRunLabelsList(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		resp, err := svc.Users.Labels.List("me").Do()
		if err != nil {
			return fmt.Errorf("listing labels: %w", err)
		}

		labels := make([]LabelInfo, 0, len(resp.Labels))
		for _, l := range resp.Labels {
			labels = append(labels, labelFromAPI(l))
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(labels)
		}

		if len(labels) == 0 {
			fmt.Println("No labels found.")
			return nil
		}

		lines := make([]string, 0, len(labels)+1)
		lines = append(lines, fmt.Sprintf("%-30s  %-10s  %-10s  %-8s", "NAME", "TYPE", "MESSAGES", "UNREAD"))
		for _, l := range labels {
			lines = append(lines, fmt.Sprintf("%-30s  %-10s  %-10d  %-8d", truncate(l.Name, 30), l.Type, l.MessagesTotal, l.MessagesUnread))
		}
		cli.PrintText(lines)
		return nil
	}
}

// newLabelsGetCmd returns the `labels get` command.
func newLabelsGetCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a label by ID",
		RunE:  makeRunLabelsGet(factory),
	}
	cmd.Flags().String("id", "", "Label ID (required)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunLabelsGet(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		l, err := svc.Users.Labels.Get("me", id).Do()
		if err != nil {
			return fmt.Errorf("getting label %s: %w", id, err)
		}

		info := labelFromAPI(l)

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(info)
		}

		lines := []string{
			fmt.Sprintf("ID:                    %s", info.ID),
			fmt.Sprintf("Name:                  %s", info.Name),
			fmt.Sprintf("Type:                  %s", info.Type),
			fmt.Sprintf("Messages Total:        %d", info.MessagesTotal),
			fmt.Sprintf("Messages Unread:       %d", info.MessagesUnread),
			fmt.Sprintf("Threads Total:         %d", info.ThreadsTotal),
			fmt.Sprintf("Threads Unread:        %d", info.ThreadsUnread),
			fmt.Sprintf("Label List Visibility: %s", info.LabelListVisibility),
			fmt.Sprintf("Message List Visibility: %s", info.MessageListVisibility),
		}
		cli.PrintText(lines)
		return nil
	}
}

// newLabelsCreateCmd returns the `labels create` command.
func newLabelsCreateCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new label",
		RunE:  makeRunLabelsCreate(factory),
	}
	cmd.Flags().String("name", "", "Label name (required)")
	cmd.Flags().String("label-list-visibility", "", "Label list visibility (labelShow, labelHide, labelShowIfUnread)")
	cmd.Flags().String("message-list-visibility", "", "Message list visibility (show, hide)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunLabelsCreate(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		name, _ := cmd.Flags().GetString("name")
		labelListVisibility, _ := cmd.Flags().GetString("label-list-visibility")
		messageListVisibility, _ := cmd.Flags().GetString("message-list-visibility")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would create label %q", name), map[string]string{
				"id":     "",
				"name":   name,
				"status": "created",
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		label := &api.Label{
			Name:                  name,
			LabelListVisibility:   labelListVisibility,
			MessageListVisibility: messageListVisibility,
		}

		created, err := svc.Users.Labels.Create("me", label).Do()
		if err != nil {
			return fmt.Errorf("creating label: %w", err)
		}

		result := map[string]string{"id": created.Id, "name": created.Name, "status": "created"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Label %q created (id: %s)\n", created.Name, created.Id)
		return nil
	}
}

// newLabelsUpdateCmd returns the `labels update` command.
func newLabelsUpdateCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a label (full replacement)",
		RunE:  makeRunLabelsUpdate(factory),
	}
	cmd.Flags().String("id", "", "Label ID (required)")
	cmd.Flags().String("name", "", "New label name (required)")
	_ = cmd.MarkFlagRequired("id")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func makeRunLabelsUpdate(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would update label %s to name %q", id, name), map[string]string{
				"id":     id,
				"name":   name,
				"status": "updated",
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		updated, err := svc.Users.Labels.Update("me", id, &api.Label{Name: name}).Do()
		if err != nil {
			return fmt.Errorf("updating label %s: %w", id, err)
		}

		result := map[string]string{"id": updated.Id, "name": updated.Name, "status": "updated"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Label %s updated\n", id)
		return nil
	}
}

// newLabelsPatchCmd returns the `labels patch` command.
func newLabelsPatchCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "patch",
		Short: "Partially update a label",
		RunE:  makeRunLabelsPatch(factory),
	}
	cmd.Flags().String("id", "", "Label ID (required)")
	cmd.Flags().String("name", "", "New label name")
	cmd.Flags().String("label-list-visibility", "", "Label list visibility (labelShow, labelHide, labelShowIfUnread)")
	cmd.Flags().String("message-list-visibility", "", "Message list visibility (show, hide)")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunLabelsPatch(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")
		labelListVisibility, _ := cmd.Flags().GetString("label-list-visibility")
		messageListVisibility, _ := cmd.Flags().GetString("message-list-visibility")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, fmt.Sprintf("Would patch label %s", id), map[string]string{
				"id":     id,
				"name":   name,
				"status": "patched",
			})
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		label := &api.Label{
			Name:                  name,
			LabelListVisibility:   labelListVisibility,
			MessageListVisibility: messageListVisibility,
		}

		patched, err := svc.Users.Labels.Patch("me", id, label).Do()
		if err != nil {
			return fmt.Errorf("patching label %s: %w", id, err)
		}

		result := map[string]string{"id": patched.Id, "name": patched.Name, "status": "patched"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Label %s patched\n", id)
		return nil
	}
}

// newLabelsDeleteCmd returns the `labels delete` command.
func newLabelsDeleteCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Permanently delete a label (IRREVERSIBLE)",
		RunE:  makeRunLabelsDelete(factory),
	}
	cmd.Flags().String("id", "", "Label ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm irreversible deletion")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func makeRunLabelsDelete(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		id, _ := cmd.Flags().GetString("id")

		if cli.IsDryRun(cmd) {
			return dryRunResult(cmd, "Would permanently delete label "+id, map[string]string{"id": id, "status": "deleted"})
		}

		if err := confirmDestructive(cmd); err != nil {
			return err
		}

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		err = svc.Users.Labels.Delete("me", id).Do()
		if err != nil {
			return fmt.Errorf("deleting label %s: %w", id, err)
		}

		result := map[string]string{"id": id, "status": "deleted"}
		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}
		fmt.Printf("Label %s permanently deleted\n", id)
		return nil
	}
}
