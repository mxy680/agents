package canvas

import (
	"encoding/json"
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newPeerReviewsCmd returns the parent "peer-reviews" command with all subcommands attached.
func newPeerReviewsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "peer-reviews",
		Short:   "Manage Canvas peer reviews for assignments",
		Aliases: []string{"review"},
	}

	cmd.AddCommand(newPeerReviewsListCmd(factory))

	return cmd
}

func newPeerReviewsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List peer reviews for an assignment",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			assignmentID, _ := cmd.Flags().GetString("assignment-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if assignmentID == "" {
				return fmt.Errorf("--assignment-id is required")
			}

			path := "/courses/" + courseID + "/assignments/" + assignmentID + "/peer_reviews"
			data, err := client.Get(ctx, path, nil)
			if err != nil {
				return err
			}

			var reviews []map[string]any
			if err := json.Unmarshal(data, &reviews); err != nil {
				return fmt.Errorf("parse peer reviews: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(reviews)
			}

			if len(reviews) == 0 {
				fmt.Println("No peer reviews found.")
				return nil
			}
			for _, r := range reviews {
				assessorID, _ := r["assessor_id"]
				assetID, _ := r["asset_id"]
				workflowState, _ := r["workflow_state"]
				fmt.Printf("assessor_id=%-6v  asset_id=%-6v  state=%v\n", assessorID, assetID, workflowState)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("assignment-id", "", "Canvas assignment ID (required)")
	return cmd
}

