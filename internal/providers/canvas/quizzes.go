package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newQuizzesCmd returns the parent "quizzes" command with all subcommands attached.
func newQuizzesCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "quizzes",
		Short:   "Manage Canvas quizzes",
		Aliases: []string{"quiz"},
	}

	cmd.AddCommand(newQuizzesListCmd(factory))
	cmd.AddCommand(newQuizzesGetCmd(factory))
	cmd.AddCommand(newQuizzesCreateCmd(factory))
	cmd.AddCommand(newQuizzesUpdateCmd(factory))
	cmd.AddCommand(newQuizzesDeleteCmd(factory))
	cmd.AddCommand(newQuizzesQuestionsCmd(factory))
	cmd.AddCommand(newQuizzesSubmissionsCmd(factory))

	return cmd
}

func newQuizzesListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List quizzes for a course",
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
			limit, _ := cmd.Flags().GetInt("limit")

			params := url.Values{}
			if search != "" {
				params.Set("search_term", search)
			}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/quizzes", params)
			if err != nil {
				return err
			}

			var quizzes []QuizSummary
			if err := json.Unmarshal(data, &quizzes); err != nil {
				return fmt.Errorf("parse quizzes: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(quizzes)
			}

			if len(quizzes) == 0 {
				fmt.Println("No quizzes found.")
				return nil
			}
			for _, q := range quizzes {
				published := "no"
				if q.Published {
					published = "yes"
				}
				fmt.Printf("%-6d  pub:%-4s  %-12s  %dq  %s\n",
					q.ID, published, q.QuizType, q.QuestionCount, truncate(q.Title, 50))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("search", "", "Search term to filter quizzes")
	cmd.Flags().Int("limit", 0, "Maximum number of quizzes to return")
	return cmd
}

func newQuizzesGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific quiz",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			quizID, _ := cmd.Flags().GetString("quiz-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if quizID == "" {
				return fmt.Errorf("--quiz-id is required")
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/quizzes/"+quizID, nil)
			if err != nil {
				return err
			}

			var quiz QuizSummary
			if err := json.Unmarshal(data, &quiz); err != nil {
				return fmt.Errorf("parse quiz: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(quiz)
			}

			fmt.Printf("ID:           %d\n", quiz.ID)
			fmt.Printf("Title:        %s\n", quiz.Title)
			fmt.Printf("Type:         %s\n", quiz.QuizType)
			fmt.Printf("Published:    %v\n", quiz.Published)
			fmt.Printf("Questions:    %d\n", quiz.QuestionCount)
			if quiz.PointsPossible > 0 {
				fmt.Printf("Points:       %.1f\n", quiz.PointsPossible)
			}
			if quiz.TimeLimit > 0 {
				fmt.Printf("Time Limit:   %d min\n", quiz.TimeLimit)
			}
			if quiz.DueAt != "" {
				fmt.Printf("Due:          %s\n", quiz.DueAt)
			}
			if quiz.Description != "" {
				fmt.Printf("Description:  %s\n", truncate(quiz.Description, 200))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("quiz-id", "", "Canvas quiz ID (required)")
	return cmd
}

func newQuizzesCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new quiz",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			title, _ := cmd.Flags().GetString("title")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if title == "" {
				return fmt.Errorf("--title is required")
			}

			quizBody := map[string]any{"title": title}
			if cmd.Flags().Changed("description") {
				v, _ := cmd.Flags().GetString("description")
				quizBody["description"] = v
			}
			if cmd.Flags().Changed("quiz-type") {
				v, _ := cmd.Flags().GetString("quiz-type")
				quizBody["quiz_type"] = v
			}
			if cmd.Flags().Changed("time-limit") {
				v, _ := cmd.Flags().GetInt("time-limit")
				quizBody["time_limit"] = v
			}
			if cmd.Flags().Changed("points") {
				v, _ := cmd.Flags().GetFloat64("points")
				quizBody["points_possible"] = v
			}
			if cmd.Flags().Changed("published") {
				v, _ := cmd.Flags().GetBool("published")
				quizBody["published"] = v
			}

			body := map[string]any{"quiz": quizBody}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("create quiz %q in course %s", title, courseID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/courses/"+courseID+"/quizzes", body)
			if err != nil {
				return err
			}

			var quiz QuizSummary
			if err := json.Unmarshal(data, &quiz); err != nil {
				return fmt.Errorf("parse quiz: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(quiz)
			}

			fmt.Printf("Created quiz %d: %s\n", quiz.ID, quiz.Title)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("title", "", "Quiz title (required)")
	cmd.Flags().String("description", "", "Quiz description")
	cmd.Flags().String("quiz-type", "", "Quiz type: practice_quiz, assignment, graded_survey, or survey")
	cmd.Flags().Int("time-limit", 0, "Time limit in minutes")
	cmd.Flags().Float64("points", 0, "Total points possible")
	cmd.Flags().Bool("published", false, "Publish the quiz immediately")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newQuizzesUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing quiz",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			quizID, _ := cmd.Flags().GetString("quiz-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if quizID == "" {
				return fmt.Errorf("--quiz-id is required")
			}

			quizBody := map[string]any{}
			if cmd.Flags().Changed("title") {
				v, _ := cmd.Flags().GetString("title")
				quizBody["title"] = v
			}
			if cmd.Flags().Changed("description") {
				v, _ := cmd.Flags().GetString("description")
				quizBody["description"] = v
			}
			if cmd.Flags().Changed("time-limit") {
				v, _ := cmd.Flags().GetInt("time-limit")
				quizBody["time_limit"] = v
			}
			if cmd.Flags().Changed("published") {
				v, _ := cmd.Flags().GetBool("published")
				quizBody["published"] = v
			}

			body := map[string]any{"quiz": quizBody}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("update quiz %s in course %s", quizID, courseID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Put(ctx, "/courses/"+courseID+"/quizzes/"+quizID, body)
			if err != nil {
				return err
			}

			var quiz QuizSummary
			if err := json.Unmarshal(data, &quiz); err != nil {
				return fmt.Errorf("parse quiz: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(quiz)
			}

			fmt.Printf("Updated quiz %d: %s\n", quiz.ID, quiz.Title)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("quiz-id", "", "Canvas quiz ID (required)")
	cmd.Flags().String("title", "", "New quiz title")
	cmd.Flags().String("description", "", "New quiz description")
	cmd.Flags().Int("time-limit", 0, "New time limit in minutes")
	cmd.Flags().Bool("published", false, "Publish or unpublish the quiz")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newQuizzesDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a quiz",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			quizID, _ := cmd.Flags().GetString("quiz-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if quizID == "" {
				return fmt.Errorf("--quiz-id is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the quiz"); err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("delete quiz %s in course %s", quizID, courseID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Delete(ctx, "/courses/"+courseID+"/quizzes/"+quizID)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"deleted": true, "quiz_id": quizID})
			}
			fmt.Printf("Quiz %s deleted.\n", quizID)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("quiz-id", "", "Canvas quiz ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newQuizzesQuestionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "questions",
		Short: "List questions for a quiz",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			quizID, _ := cmd.Flags().GetString("quiz-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if quizID == "" {
				return fmt.Errorf("--quiz-id is required")
			}

			limit, _ := cmd.Flags().GetInt("limit")
			params := url.Values{}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/quizzes/"+quizID+"/questions", params)
			if err != nil {
				return err
			}

			// Question structures vary widely; output raw JSON array.
			var questions []map[string]any
			if err := json.Unmarshal(data, &questions); err != nil {
				return fmt.Errorf("parse questions: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(questions)
			}

			if len(questions) == 0 {
				fmt.Println("No questions found.")
				return nil
			}
			for _, q := range questions {
				id, _ := q["id"]
				pos, _ := q["position"]
				qtype, _ := q["question_type"].(string)
				name, _ := q["question_name"].(string)
				pts, _ := q["points_possible"]
				fmt.Printf("%-6v  pos:%-4v  %-20s  pts:%-6v  %s\n", id, pos, qtype, pts, truncate(name, 40))
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("quiz-id", "", "Canvas quiz ID (required)")
	cmd.Flags().Int("limit", 0, "Maximum number of questions to return")
	return cmd
}

func newQuizzesSubmissionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submissions",
		Short: "List submissions for a quiz",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			courseID, _ := cmd.Flags().GetString("course-id")
			quizID, _ := cmd.Flags().GetString("quiz-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if quizID == "" {
				return fmt.Errorf("--quiz-id is required")
			}

			limit, _ := cmd.Flags().GetInt("limit")
			params := url.Values{}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/quizzes/"+quizID+"/submissions", params)
			if err != nil {
				return err
			}

			// Canvas wraps quiz submissions in {"quiz_submissions": [...]}.
			var envelope struct {
				QuizSubmissions []map[string]any `json:"quiz_submissions"`
			}
			if err := json.Unmarshal(data, &envelope); err != nil {
				return fmt.Errorf("parse quiz submissions: %w", err)
			}

			submissions := envelope.QuizSubmissions

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(submissions)
			}

			if len(submissions) == 0 {
				fmt.Println("No quiz submissions found.")
				return nil
			}
			for _, s := range submissions {
				id, _ := s["id"]
				userID, _ := s["user_id"]
				state, _ := s["workflow_state"].(string)
				score, _ := s["score"]
				attempt, _ := s["attempt"]
				submittedAt, _ := s["finished_at"].(string)
				fmt.Printf("id:%-6v  user:%-6v  state:%-12s  score:%-8v  attempt:%-4v  finished:%s\n",
					id, userID, state, score, attempt, submittedAt)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("quiz-id", "", "Canvas quiz ID (required)")
	cmd.Flags().Int("limit", 0, "Maximum number of submissions to return")
	return cmd
}
