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
	cmd.AddCommand(newPeerReviewsCreateCmd(factory))
	cmd.AddCommand(newPeerReviewsDeleteCmd(factory))

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

func newPeerReviewsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Assign a peer reviewer to a submission",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			assignmentID, _ := cmd.Flags().GetString("assignment-id")
			userID, _ := cmd.Flags().GetString("user-id")
			reviewerID, _ := cmd.Flags().GetString("reviewer-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if assignmentID == "" {
				return fmt.Errorf("--assignment-id is required")
			}
			if userID == "" {
				return fmt.Errorf("--user-id is required")
			}
			if reviewerID == "" {
				return fmt.Errorf("--reviewer-id is required")
			}

			path := "/courses/" + courseID + "/assignments/" + assignmentID + "/submissions/" + userID + "/peer_reviews"
			body := map[string]any{"user_id": reviewerID}
			data, err := client.Post(ctx, path, body)
			if err != nil {
				return err
			}

			var review map[string]any
			if err := json.Unmarshal(data, &review); err != nil {
				return fmt.Errorf("parse peer review: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(review)
			}
			assessorID, _ := review["assessor_id"]
			workflowState, _ := review["workflow_state"].(string)
			fmt.Printf("Peer reviewer %v assigned (state: %s)\n", assessorID, workflowState)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("assignment-id", "", "Canvas assignment ID (required)")
	cmd.Flags().String("user-id", "", "Submission owner user ID (required)")
	cmd.Flags().String("reviewer-id", "", "Reviewer user ID (required)")
	return cmd
}

func newPeerReviewsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Remove a peer review assignment",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			assignmentID, _ := cmd.Flags().GetString("assignment-id")
			userID, _ := cmd.Flags().GetString("user-id")
			reviewerID, _ := cmd.Flags().GetString("reviewer-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if assignmentID == "" {
				return fmt.Errorf("--assignment-id is required")
			}
			if userID == "" {
				return fmt.Errorf("--user-id is required")
			}
			if reviewerID == "" {
				return fmt.Errorf("--reviewer-id is required")
			}

			if err := confirmDestructive(cmd, "this will remove the peer review assignment"); err != nil {
				return err
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			path := "/courses/" + courseID + "/assignments/" + assignmentID + "/submissions/" + userID + "/peer_reviews"
			if _, err := client.Delete(ctx, path); err != nil {
				return err
			}

			fmt.Printf("Peer reviewer %s removed from submission\n", reviewerID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("assignment-id", "", "Canvas assignment ID (required)")
	cmd.Flags().String("user-id", "", "Submission owner user ID (required)")
	cmd.Flags().String("reviewer-id", "", "Reviewer user ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive action")
	return cmd
}
