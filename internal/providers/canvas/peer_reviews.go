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

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf(
					"assign reviewer %s to submission by user %s (assignment %s, course %s)",
					reviewerID, userID, assignmentID, courseID,
				), map[string]any{"user_id": reviewerID})
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			path := "/courses/" + courseID + "/assignments/" + assignmentID + "/submissions/" + userID + "/peer_reviews"
			params := url.Values{}
			params.Set("user_id", reviewerID)

			// Canvas expects user_id as a query parameter for the reviewer.
			fullPath := path + "?" + params.Encode()
			data, err := client.Post(ctx, fullPath, nil)
			if err != nil {
				return err
			}

			var result map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				return fmt.Errorf("parse created peer review: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(result)
			}
			fmt.Printf("Peer review assigned: reviewer %s → submission by user %s.\n", reviewerID, userID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("assignment-id", "", "Canvas assignment ID (required)")
	cmd.Flags().String("user-id", "", "User ID of the submission author (required)")
	cmd.Flags().String("reviewer-id", "", "User ID of the peer reviewer to assign (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newPeerReviewsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Remove a peer reviewer from a submission",
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

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf(
					"remove reviewer %s from submission by user %s (assignment %s, course %s)",
					reviewerID, userID, assignmentID, courseID,
				), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			params := url.Values{}
			params.Set("user_id", reviewerID)
			path := "/courses/" + courseID + "/assignments/" + assignmentID + "/submissions/" + userID + "/peer_reviews?" + params.Encode()

			_, err = client.Delete(ctx, path)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{
					"deleted":       true,
					"reviewer_id":   reviewerID,
					"user_id":       userID,
					"assignment_id": assignmentID,
				})
			}
			fmt.Printf("Peer review assignment removed: reviewer %s from submission by user %s.\n", reviewerID, userID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("assignment-id", "", "Canvas assignment ID (required)")
	cmd.Flags().String("user-id", "", "User ID of the submission author (required)")
	cmd.Flags().String("reviewer-id", "", "User ID of the peer reviewer to remove (required)")
	cmd.Flags().Bool("confirm", false, "Confirm removal")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}
