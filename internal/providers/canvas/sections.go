package canvas

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// newSectionsCmd returns the parent "sections" command with all subcommands attached.
func newSectionsCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sections",
		Short:   "Manage Canvas course sections",
		Aliases: []string{"section", "sec"},
	}

	cmd.AddCommand(newSectionsListCmd(factory))
	cmd.AddCommand(newSectionsGetCmd(factory))
	cmd.AddCommand(newSectionsCreateCmd(factory))
	cmd.AddCommand(newSectionsUpdateCmd(factory))
	cmd.AddCommand(newSectionsDeleteCmd(factory))
	cmd.AddCommand(newSectionsCrosslistCmd(factory))
	cmd.AddCommand(newSectionsUncrosslistCmd(factory))

	return cmd
}

func newSectionsListCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sections for a course",
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

			limit, _ := cmd.Flags().GetInt("limit")
			params := url.Values{}
			if limit > 0 {
				params.Set("per_page", strconv.Itoa(limit))
			}

			data, err := client.Get(ctx, "/courses/"+courseID+"/sections", params)
			if err != nil {
				return err
			}

			var sections []SectionSummary
			if err := json.Unmarshal(data, &sections); err != nil {
				return fmt.Errorf("parse sections: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(sections)
			}

			if len(sections) == 0 {
				fmt.Println("No sections found.")
				return nil
			}
			for _, s := range sections {
				students := ""
				if s.TotalStudents > 0 {
					students = fmt.Sprintf(" (%d students)", s.TotalStudents)
				}
				fmt.Printf("%-6d  %s%s\n", s.ID, truncate(s.Name, 60), students)
			}
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().Int("limit", 0, "Maximum number of sections to return")
	return cmd
}

func newSectionsGetCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get details for a specific section",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			sectionID, _ := cmd.Flags().GetString("section-id")
			if sectionID == "" {
				return fmt.Errorf("--section-id is required")
			}

			data, err := client.Get(ctx, "/sections/"+sectionID, nil)
			if err != nil {
				return err
			}

			var section SectionSummary
			if err := json.Unmarshal(data, &section); err != nil {
				return fmt.Errorf("parse section: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(section)
			}

			fmt.Printf("ID:        %d\n", section.ID)
			fmt.Printf("Name:      %s\n", section.Name)
			if section.CourseID > 0 {
				fmt.Printf("Course ID: %d\n", section.CourseID)
			}
			if section.TotalStudents > 0 {
				fmt.Printf("Students:  %d\n", section.TotalStudents)
			}
			if section.StartAt != "" {
				fmt.Printf("Start:     %s\n", section.StartAt)
			}
			if section.EndAt != "" {
				fmt.Printf("End:       %s\n", section.EndAt)
			}
			if section.NonxlistCourseID > 0 {
				fmt.Printf("Cross-listed from course: %d\n", section.NonxlistCourseID)
			}
			return nil
		},
	}

	cmd.Flags().String("section-id", "", "Canvas section ID (required)")
	return cmd
}

func newSectionsCreateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new section in a course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			courseID, _ := cmd.Flags().GetString("course-id")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			startAt, _ := cmd.Flags().GetString("start-at")
			endAt, _ := cmd.Flags().GetString("end-at")

			sectionBody := map[string]any{
				"name": name,
			}
			if startAt != "" {
				sectionBody["start_at"] = startAt
			}
			if endAt != "" {
				sectionBody["end_at"] = endAt
			}

			body := map[string]any{"course_section": sectionBody}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("create section %q in course %s", name, courseID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/courses/"+courseID+"/sections", body)
			if err != nil {
				return err
			}

			var section SectionSummary
			if err := json.Unmarshal(data, &section); err != nil {
				return fmt.Errorf("parse created section: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(section)
			}
			fmt.Printf("Section created: %d — %s\n", section.ID, section.Name)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("name", "", "Section name (required)")
	cmd.Flags().String("start-at", "", "Section start date (RFC3339)")
	cmd.Flags().String("end-at", "", "Section end date (RFC3339)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newSectionsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing section",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			sectionID, _ := cmd.Flags().GetString("section-id")
			if sectionID == "" {
				return fmt.Errorf("--section-id is required")
			}

			sectionBody := map[string]any{}
			if cmd.Flags().Changed("name") {
				v, _ := cmd.Flags().GetString("name")
				sectionBody["name"] = v
			}
			if cmd.Flags().Changed("start-at") {
				v, _ := cmd.Flags().GetString("start-at")
				sectionBody["start_at"] = v
			}
			if cmd.Flags().Changed("end-at") {
				v, _ := cmd.Flags().GetString("end-at")
				sectionBody["end_at"] = v
			}

			body := map[string]any{"course_section": sectionBody}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("update section %s", sectionID), body)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Put(ctx, "/sections/"+sectionID, body)
			if err != nil {
				return err
			}

			var section SectionSummary
			if err := json.Unmarshal(data, &section); err != nil {
				return fmt.Errorf("parse updated section: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(section)
			}
			fmt.Printf("Section %d updated.\n", section.ID)
			return nil
		},
	}

	cmd.Flags().String("section-id", "", "Canvas section ID (required)")
	cmd.Flags().String("name", "", "New section name")
	cmd.Flags().String("start-at", "", "New start date (RFC3339)")
	cmd.Flags().String("end-at", "", "New end date (RFC3339)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newSectionsDeleteCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a section",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			sectionID, _ := cmd.Flags().GetString("section-id")
			if sectionID == "" {
				return fmt.Errorf("--section-id is required")
			}

			if err := confirmDestructive(cmd, "this will permanently delete the section"); err != nil {
				return err
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("delete section %s", sectionID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			_, err = client.Delete(ctx, "/sections/"+sectionID)
			if err != nil {
				return err
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(map[string]any{"deleted": true, "section_id": sectionID})
			}
			fmt.Printf("Section %s deleted.\n", sectionID)
			return nil
		},
	}

	cmd.Flags().String("section-id", "", "Canvas section ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm deletion")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newSectionsCrosslistCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crosslist",
		Short: "Cross-list a section into a different course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			sectionID, _ := cmd.Flags().GetString("section-id")
			newCourseID, _ := cmd.Flags().GetString("new-course-id")
			if sectionID == "" {
				return fmt.Errorf("--section-id is required")
			}
			if newCourseID == "" {
				return fmt.Errorf("--new-course-id is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("crosslist section %s into course %s", sectionID, newCourseID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Post(ctx, "/sections/"+sectionID+"/crosslist/"+newCourseID, nil)
			if err != nil {
				return err
			}

			var section SectionSummary
			if err := json.Unmarshal(data, &section); err != nil {
				return fmt.Errorf("parse crosslist result: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(section)
			}
			fmt.Printf("Section %d cross-listed into course %s.\n", section.ID, newCourseID)
			return nil
		},
	}

	cmd.Flags().String("section-id", "", "Canvas section ID (required)")
	cmd.Flags().String("new-course-id", "", "Destination course ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}

func newSectionsUncrosslistCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uncrosslist",
		Short: "Remove a section from cross-listing",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			sectionID, _ := cmd.Flags().GetString("section-id")
			if sectionID == "" {
				return fmt.Errorf("--section-id is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, fmt.Sprintf("uncrosslist section %s", sectionID), nil)
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			data, err := client.Delete(ctx, "/sections/"+sectionID+"/crosslist")
			if err != nil {
				return err
			}

			var section SectionSummary
			if err := json.Unmarshal(data, &section); err != nil {
				return fmt.Errorf("parse uncrosslist result: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(section)
			}
			fmt.Printf("Section %d removed from cross-listing.\n", section.ID)
			return nil
		},
	}

	cmd.Flags().String("section-id", "", "Canvas section ID (required)")
	cmd.Flags().Bool("dry-run", false, "Preview without executing")
	return cmd
}
