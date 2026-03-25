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

