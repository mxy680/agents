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
			name, _ := cmd.Flags().GetString("name")
			if courseID == "" {
				return fmt.Errorf("--course-id is required")
			}
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				return dryRunResult(cmd, "create section: "+name, map[string]any{"course_id": courseID, "name": name})
			}

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			body := map[string]any{"name": name}
			data, err := client.Post(ctx, "/courses/"+courseID+"/sections", map[string]any{"course_section": body})
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
			fmt.Printf("Section %d created: %s\n", section.ID, section.Name)
			return nil
		},
	}

	cmd.Flags().String("course-id", "", "Canvas course ID (required)")
	cmd.Flags().String("name", "", "Section name (required)")
	return cmd
}

func newSectionsUpdateCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a section",
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

			section := map[string]any{}
			if name, _ := cmd.Flags().GetString("name"); name != "" {
				section["name"] = name
			}

			data, err := client.Put(ctx, "/sections/"+sectionID, map[string]any{"course_section": section})
			if err != nil {
				return err
			}

			var s SectionSummary
			if err := json.Unmarshal(data, &s); err != nil {
				return fmt.Errorf("parse section: %w", err)
			}

			if cli.IsJSONOutput(cmd) {
				return cli.PrintJSON(s)
			}
			fmt.Printf("Section %s updated\n", sectionID)
			return nil
		},
	}

	cmd.Flags().String("section-id", "", "Canvas section ID (required)")
	cmd.Flags().String("name", "", "New name")
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

			client, err := factory(ctx)
			if err != nil {
				return err
			}

			if _, err := client.Delete(ctx, "/sections/"+sectionID); err != nil {
				return err
			}

			fmt.Printf("Section %s deleted\n", sectionID)
			return nil
		},
	}

	cmd.Flags().String("section-id", "", "Canvas section ID (required)")
	cmd.Flags().Bool("confirm", false, "Confirm destructive action")
	return cmd
}

func newSectionsCrosslistCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crosslist",
		Short: "Cross-list a section into another course",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, err := factory(ctx)
			if err != nil {
				return err
			}

			sectionID, _ := cmd.Flags().GetString("section-id")
			newCourseID, _ := cmd.Flags().GetString("new-course-id")
			if sectionID == "" {
				return fmt.Errorf("--section-id is required")
			}
			if newCourseID == "" {
				return fmt.Errorf("--new-course-id is required")
			}

			path := "/sections/" + sectionID + "/crosslist/" + newCourseID
			data, err := client.Post(ctx, path, nil)
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
			fmt.Printf("Section %s cross-listed into course %s\n", sectionID, newCourseID)
			return nil
		},
	}

	cmd.Flags().String("section-id", "", "Canvas section ID (required)")
	cmd.Flags().String("new-course-id", "", "Target course ID (required)")
	return cmd
}

func newSectionsUncrosslistCmd(factory ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uncrosslist",
		Short: "Remove a section from cross-listing",
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

			path := "/sections/" + sectionID + "/crosslist"
			data, err := client.Delete(ctx, path)
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
			fmt.Printf("Section %s removed from cross-listing\n", sectionID)
			return nil
		},
	}

	cmd.Flags().String("section-id", "", "Canvas section ID (required)")
	return cmd
}
