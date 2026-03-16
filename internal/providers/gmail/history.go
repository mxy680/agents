package gmail

import (
	"context"
	"fmt"
	"strings"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// HistoryLabelChange records which labels were added or removed on a message.
type HistoryLabelChange struct {
	MessageID string   `json:"messageId"`
	LabelIDs  []string `json:"labelIds"`
}

// HistoryEntry summarises a single history record.
type HistoryEntry struct {
	ID              uint64               `json:"id"`
	MessagesAdded   []string             `json:"messagesAdded,omitempty"`
	MessagesDeleted []string             `json:"messagesDeleted,omitempty"`
	LabelsAdded     []HistoryLabelChange `json:"labelsAdded,omitempty"`
	LabelsRemoved   []HistoryLabelChange `json:"labelsRemoved,omitempty"`
}

// HistoryResult is the JSON-serializable response for history list.
type HistoryResult struct {
	History       []HistoryEntry `json:"history"`
	HistoryID     uint64         `json:"historyId"`
	NextPageToken string         `json:"nextPageToken,omitempty"`
}

// newHistoryListCmd returns the `history list` command.
func newHistoryListCmd(factory ServiceFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List mailbox changes since a given history ID",
		RunE:  makeRunHistoryList(factory),
	}
	cmd.Flags().Uint64("start-history-id", 0, "History ID to start listing from (required)")
	cmd.Flags().String("label-id", "", "Filter history by label ID")
	cmd.Flags().String("history-types", "", "Comma-separated history types to include (messageAdded,messageDeleted,labelAdded,labelRemoved)")
	cmd.Flags().Int64("limit", 0, "Maximum number of history records to return")
	cmd.Flags().String("page-token", "", "Page token for pagination")
	_ = cmd.MarkFlagRequired("start-history-id")
	return cmd
}

func makeRunHistoryList(factory ServiceFactory) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := context.Background()
		startID, _ := cmd.Flags().GetUint64("start-history-id")
		labelID, _ := cmd.Flags().GetString("label-id")
		historyTypesStr, _ := cmd.Flags().GetString("history-types")
		limit, _ := cmd.Flags().GetInt64("limit")
		pageToken, _ := cmd.Flags().GetString("page-token")

		svc, err := factory(ctx)
		if err != nil {
			return err
		}

		call := svc.Users.History.List("me").StartHistoryId(startID)

		if labelID != "" {
			call = call.LabelId(labelID)
		}
		if historyTypesStr != "" {
			types := strings.Split(historyTypesStr, ",")
			call = call.HistoryTypes(types...)
		}
		if limit > 0 {
			call = call.MaxResults(limit)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing history from %d: %w", startID, err)
		}

		entries := make([]HistoryEntry, 0, len(resp.History))
		for _, h := range resp.History {
			entry := HistoryEntry{ID: h.Id}

			for _, ma := range h.MessagesAdded {
				if ma.Message != nil {
					entry.MessagesAdded = append(entry.MessagesAdded, ma.Message.Id)
				}
			}
			for _, md := range h.MessagesDeleted {
				if md.Message != nil {
					entry.MessagesDeleted = append(entry.MessagesDeleted, md.Message.Id)
				}
			}
			for _, la := range h.LabelsAdded {
				if la.Message != nil {
					entry.LabelsAdded = append(entry.LabelsAdded, HistoryLabelChange{
						MessageID: la.Message.Id,
						LabelIDs:  la.LabelIds,
					})
				}
			}
			for _, lr := range h.LabelsRemoved {
				if lr.Message != nil {
					entry.LabelsRemoved = append(entry.LabelsRemoved, HistoryLabelChange{
						MessageID: lr.Message.Id,
						LabelIDs:  lr.LabelIds,
					})
				}
			}

			entries = append(entries, entry)
		}

		result := HistoryResult{
			History:       entries,
			HistoryID:     resp.HistoryId,
			NextPageToken: resp.NextPageToken,
		}

		if cli.IsJSONOutput(cmd) {
			return cli.PrintJSON(result)
		}

		if len(entries) == 0 {
			fmt.Println("No history records found.")
			return nil
		}

		lines := make([]string, 0, len(entries)+1)
		lines = append(lines, fmt.Sprintf("%-12s  %-8s  %-8s  %-10s  %-10s", "HISTORY_ID", "MSG_ADD", "MSG_DEL", "LBL_ADD", "LBL_REM"))
		for _, e := range entries {
			lines = append(lines, fmt.Sprintf("%-12d  %-8d  %-8d  %-10d  %-10d",
				e.ID,
				len(e.MessagesAdded),
				len(e.MessagesDeleted),
				len(e.LabelsAdded),
				len(e.LabelsRemoved),
			))
		}
		cli.PrintText(lines)
		return nil
	}
}
