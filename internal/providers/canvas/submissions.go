package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newSubmissionsCmd returns the parent "submissions" command with all subcommands attached.
func newSubmissionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "submissions",
		Short:   "Manage Canvas submissions",
		Aliases: []string{"submission", "sub"},
	}

	cmd.AddCommand(newSubmissionsListCmd(factory))
	cmd.AddCommand(newSubmissionsGetCmd(factory))

	return cmd
}

func newSubmissionsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List submissions for an assignment",
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

			limit, _ := cmd.Flags().GetInt("limit")
			params := url.Values{}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			path := "/courses/" + courseID + "/assignments/" + assignmentID + "/submissions"
			data, err := client.Get(ctx, path, params)
			if err != nil {
				return err
			}

			var submissions []SubmissionSummary
			if err := json.Unmarshal(data, &submissions); err != nil {
				return fmt.Errorf("parse submissions: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(submissions)
			}

			if len(submissions) == 0 {
				fmt.Println("No submissions found.")
				return nil
			}
			for _, s := range submissions {
				grade := s.Grade
				if grade == "" {
					grade = "ungraded"
				}
				fmt.Printf("user:%-6d  state:%-10s  grade:%-8s  submitted:%s\n",
					s.UserID, s.WorkflowState, grade, s.SubmittedAt)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("assignment-id", "", "Canvas assignment ID (required)")
	cmd.Flags().Int("limit", 0, "Maximum number of submissions to return")
	return cmd
}

func newSubmissionsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a specific submission",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			assignmentID, _ := cmd.Flags().GetString("assignment-id")
			userID, _ := cmd.Flags().GetString("user-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if assignmentID == "" {
				return fmt.Errorf("--assignment-id is required")
			}
			if userID == "" {
				userID = "self"
			}

			path := "/courses/" + courseID + "/assignments/" + assignmentID + "/submissions/" + userID
			data, err := client.Get(ctx, path, nil)
			if err != nil {
				return err
			}

			var submission SubmissionSummary
			if err := json.Unmarshal(data, &submission); err != nil {
				return fmt.Errorf("parse submission: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(submission)
			}

			fmt.Printf("ID:           %d\n", submission.ID)
			fmt.Printf("User ID:      %d\n", submission.UserID)
			fmt.Printf("State:        %s\n", submission.WorkflowState)
			if submission.Grade != "" {
				fmt.Printf("Grade:        %s\n", submission.Grade)
				fmt.Printf("Score:        %.1f\n", submission.Score)
			}
			if submission.SubmittedAt != "" {
				fmt.Printf("Submitted:    %s\n", submission.SubmittedAt)
			}
			if submission.GradedAt != "" {
				fmt.Printf("Graded:       %s\n", submission.GradedAt)
			}
			if submission.Late {
				fmt.Println("Late:         yes")
			}
			if submission.Missing {
				fmt.Println("Missing:      yes")
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("assignment-id", "", "Canvas assignment ID (required)")
	cmd.Flags().String("user-id", "", "Canvas user ID (defaults to self)")
	return cmd
}
