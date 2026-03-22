package canvas

import (
	"fmt"

	"github.com/emdash-projects/agents/internal/cli"
	"github.com/spf13/cobra"
)

// CourseSummary is a condensed Canvas course representation.
type CourseSummary struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	CourseCode       string `json:"course_code,omitempty"`
	WorkflowState    string `json:"workflow_state,omitempty"`
	EnrollmentType   string `json:"enrollment_type,omitempty"`
	StartAt          string `json:"start_at,omitempty"`
	EndAt            string `json:"end_at,omitempty"`
	DefaultView      string `json:"default_view,omitempty"`
	TotalStudents    int    `json:"total_students,omitempty"`
}

// AssignmentSummary is a condensed Canvas assignment representation.
type AssignmentSummary struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description,omitempty"`
	DueAt           string   `json:"due_at,omitempty"`
	PointsPossible  float64  `json:"points_possible,omitempty"`
	GradingType     string   `json:"grading_type,omitempty"`
	SubmissionTypes []string `json:"submission_types,omitempty"`
	CourseID        int      `json:"course_id,omitempty"`
	Published       bool     `json:"published,omitempty"`
	Position        int      `json:"position,omitempty"`
	Locked          bool     `json:"locked_for_user,omitempty"`
}

// SubmissionSummary is a condensed Canvas submission representation.
type SubmissionSummary struct {
	ID              int     `json:"id"`
	AssignmentID    int     `json:"assignment_id"`
	UserID          int     `json:"user_id"`
	WorkflowState   string  `json:"workflow_state,omitempty"`
	Grade           string  `json:"grade,omitempty"`
	Score           float64 `json:"score,omitempty"`
	SubmittedAt     string  `json:"submitted_at,omitempty"`
	GradedAt        string  `json:"graded_at,omitempty"`
	Late            bool    `json:"late,omitempty"`
	Missing         bool    `json:"missing,omitempty"`
	SubmissionType  string  `json:"submission_type,omitempty"`
	Attempt         int     `json:"attempt,omitempty"`
}

// UserSummary is a condensed Canvas user representation.
type UserSummary struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	ShortName     string `json:"short_name,omitempty"`
	SortableName  string `json:"sortable_name,omitempty"`
	Email         string `json:"email,omitempty"`
	LoginID       string `json:"login_id,omitempty"`
	AvatarURL     string `json:"avatar_url,omitempty"`
	Locale        string `json:"locale,omitempty"`
	Bio           string `json:"bio,omitempty"`
}

// ModuleSummary is a condensed Canvas module representation.
type ModuleSummary struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Position          int    `json:"position,omitempty"`
	UnlockAt          string `json:"unlock_at,omitempty"`
	State             string `json:"state,omitempty"`
	ItemsCount        int    `json:"items_count,omitempty"`
	CompletedAt       string `json:"completed_at,omitempty"`
	RequireSequential bool   `json:"require_sequential_progress,omitempty"`
	Published         bool   `json:"published,omitempty"`
}

// ModuleItemSummary is a condensed Canvas module item representation.
type ModuleItemSummary struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Position     int    `json:"position,omitempty"`
	Type         string `json:"type,omitempty"`
	ContentID    int    `json:"content_id,omitempty"`
	HTMLURL      string `json:"html_url,omitempty"`
	ModuleID     int    `json:"module_id,omitempty"`
}

// DiscussionSummary is a condensed Canvas discussion topic representation.
type DiscussionSummary struct {
	ID             int    `json:"id"`
	Title          string `json:"title"`
	Message        string `json:"message,omitempty"`
	PostedAt       string `json:"posted_at,omitempty"`
	LastReplyAt    string `json:"last_reply_at,omitempty"`
	DiscussionType string `json:"discussion_type,omitempty"`
	Published      bool   `json:"published,omitempty"`
	Pinned         bool   `json:"pinned,omitempty"`
	Locked         bool   `json:"locked,omitempty"`
	UserName       string `json:"user_name,omitempty"`
}

// AnnouncementSummary is a condensed Canvas announcement representation.
type AnnouncementSummary struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Message   string `json:"message,omitempty"`
	PostedAt  string `json:"posted_at,omitempty"`
	Published bool   `json:"published,omitempty"`
	UserName  string `json:"user_name,omitempty"`
	CourseID  int    `json:"context_code,omitempty"`
}

// PageSummary is a condensed Canvas page representation.
type PageSummary struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	Published   bool   `json:"published,omitempty"`
	FrontPage   bool   `json:"front_page,omitempty"`
	EditingRoles string `json:"editing_roles,omitempty"`
}

// FileSummary is a condensed Canvas file representation.
type FileSummary struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Filename    string `json:"filename,omitempty"`
	ContentType string `json:"content-type,omitempty"`
	Size        int64  `json:"size,omitempty"`
	URL         string `json:"url,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	FolderID    int    `json:"folder_id,omitempty"`
	Locked      bool   `json:"locked,omitempty"`
	Hidden      bool   `json:"hidden,omitempty"`
}

// FolderSummary is a condensed Canvas folder representation.
type FolderSummary struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	FullName     string `json:"full_name,omitempty"`
	FilesCount   int    `json:"files_count,omitempty"`
	FoldersCount int    `json:"folders_count,omitempty"`
	ParentID     int    `json:"parent_folder_id,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
	Locked       bool   `json:"locked,omitempty"`
	Hidden       bool   `json:"hidden,omitempty"`
}

// EnrollmentSummary is a condensed Canvas enrollment representation.
type EnrollmentSummary struct {
	ID               int     `json:"id"`
	CourseID         int     `json:"course_id,omitempty"`
	UserID           int     `json:"user_id,omitempty"`
	Type             string  `json:"type,omitempty"`
	EnrollmentState  string  `json:"enrollment_state,omitempty"`
	Role             string  `json:"role,omitempty"`
	CurrentGrade     string  `json:"current_grade,omitempty"`
	CurrentScore     float64 `json:"current_score,omitempty"`
	FinalGrade       string  `json:"final_grade,omitempty"`
	FinalScore       float64 `json:"final_score,omitempty"`
	UserName         string  `json:"user_name,omitempty"`
}

// GradeSummary represents grade information from Canvas.
type GradeSummary struct {
	CurrentGrade string  `json:"current_grade,omitempty"`
	CurrentScore float64 `json:"current_score,omitempty"`
	FinalGrade   string  `json:"final_grade,omitempty"`
	FinalScore   float64 `json:"final_score,omitempty"`
}

// CalendarEventSummary is a condensed Canvas calendar event.
type CalendarEventSummary struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	StartAt     string `json:"start_at,omitempty"`
	EndAt       string `json:"end_at,omitempty"`
	Description string `json:"description,omitempty"`
	LocationName string `json:"location_name,omitempty"`
	AllDay       bool   `json:"all_day,omitempty"`
	ContextCode  string `json:"context_code,omitempty"`
	WorkflowState string `json:"workflow_state,omitempty"`
}

// ConversationSummary is a condensed Canvas conversation.
type ConversationSummary struct {
	ID              int      `json:"id"`
	Subject         string   `json:"subject,omitempty"`
	WorkflowState   string   `json:"workflow_state,omitempty"`
	LastMessage      string   `json:"last_message,omitempty"`
	LastMessageAt    string   `json:"last_message_at,omitempty"`
	MessageCount     int      `json:"message_count,omitempty"`
	Starred          bool     `json:"starred,omitempty"`
	Participants     []string `json:"participants,omitempty"`
}

// QuizSummary is a condensed Canvas quiz representation.
type QuizSummary struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Description   string  `json:"description,omitempty"`
	QuizType      string  `json:"quiz_type,omitempty"`
	PointsPossible float64 `json:"points_possible,omitempty"`
	TimeLimit      int     `json:"time_limit,omitempty"`
	Published      bool    `json:"published,omitempty"`
	DueAt          string  `json:"due_at,omitempty"`
	QuestionCount  int     `json:"question_count,omitempty"`
}

// truncate shortens s to at most max runes, appending "..." if truncated.
func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-3]) + "..."
}

// confirmDestructive returns an error if --confirm flag is absent or false.
func confirmDestructive(cmd *cobra.Command, msg string) error {
	confirmed, _ := cmd.Flags().GetBool("confirm")
	if !confirmed {
		return fmt.Errorf("%s; re-run with --confirm to proceed", msg)
	}
	return nil
}

// dryRunResult prints a dry-run preview and returns nil.
func dryRunResult(cmd *cobra.Command, action string, data any) error {
	if cli.IsJSONOutput(cmd) {
		return cli.PrintJSON(data)
	}
	fmt.Printf("[DRY RUN] %s\n", action)
	return nil
}

// GroupSummary is a condensed Canvas group representation.
type GroupSummary struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	MembersCount int    `json:"members_count,omitempty"`
	JoinLevel    string `json:"join_level,omitempty"`
	ContextType  string `json:"context_type,omitempty"`
	CourseID     int    `json:"course_id,omitempty"`
}

// RubricSummary is a condensed Canvas rubric representation.
type RubricSummary struct {
	ID             int     `json:"id"`
	Title          string  `json:"title"`
	PointsPossible float64 `json:"points_possible,omitempty"`
	RubricType     string  `json:"context_type,omitempty"`
	ReadOnly       bool    `json:"read_only,omitempty"`
}

// SectionSummary is a condensed Canvas section representation.
type SectionSummary struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	CourseID         int    `json:"course_id,omitempty"`
	StartAt          string `json:"start_at,omitempty"`
	EndAt            string `json:"end_at,omitempty"`
	TotalStudents    int    `json:"total_students,omitempty"`
	NonxlistCourseID int    `json:"nonxlist_course_id,omitempty"`
}

// formatSize formats a file size in bytes to a human-readable string.
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
