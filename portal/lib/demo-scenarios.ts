/**
 * Pre-built demo scenarios for each agent type.
 * Mark clicks one of these to trigger an impressive demo flow.
 */

export interface DemoScenario {
  id: string
  label: string
  description: string
  prompt: string
  icon: string // Tabler icon name
}

export const DEMO_SCENARIOS: Record<string, DemoScenario[]> = {
  "email-assistant": [
    {
      id: "inbox-summary",
      label: "Inbox Summary",
      description: "Summarize unread emails from the last 24 hours",
      prompt:
        "Check my unread emails from the last 24 hours and give me a summary of what's important.",
      icon: "IconMailOpened",
    },
    {
      id: "priority-triage",
      label: "Priority Triage",
      description: "Find urgent emails that need a response today",
      prompt:
        "Find any urgent emails that need a response today. Look for emails from real people (not newsletters or marketing) and prioritize them.",
      icon: "IconUrgent",
    },
    {
      id: "search-emails",
      label: "Search Emails",
      description: "Search for emails by sender or topic",
      prompt:
        "Search my recent emails for anything related to meetings or scheduling in the last week.",
      icon: "IconSearch",
    },
  ],
  "github-assistant": [
    {
      id: "pr-review",
      label: "PR Review",
      description: "List open PRs that need attention",
      prompt:
        "List all open pull requests across my repositories. For each one, show the title, author, and how many days it's been open.",
      icon: "IconGitPullRequest",
    },
    {
      id: "issue-dashboard",
      label: "Issue Dashboard",
      description: "Show open issues assigned to me",
      prompt:
        "Show me all open issues assigned to me across my repositories, sorted by most recently updated.",
      icon: "IconChecklist",
    },
    {
      id: "repo-overview",
      label: "Repo Overview",
      description: "Overview of my most active repositories",
      prompt:
        "List my 5 most recently updated repositories with their description, language, and star count.",
      icon: "IconBook",
    },
  ],
  "instagram-assistant": [
    {
      id: "engagement-report",
      label: "Engagement Report",
      description: "Check engagement on recent posts",
      prompt:
        "Show me my most recent Instagram posts and their engagement — likes, comments, and any notable interactions.",
      icon: "IconChartBar",
    },
    {
      id: "activity-feed",
      label: "Recent Activity",
      description: "Check notifications and recent activity",
      prompt:
        "Check my Instagram activity feed — show me recent likes, comments, and follows.",
      icon: "IconBell",
    },
    {
      id: "follower-insights",
      label: "Follower Insights",
      description: "Analyze followers and following",
      prompt:
        "Show me my follower count, following count, and list my most recent followers.",
      icon: "IconUsers",
    },
  ],
  "supabase-assistant": [
    {
      id: "health-check",
      label: "Health Check",
      description: "Check status of all Supabase projects",
      prompt:
        "List all my Supabase projects and check the health status of each one.",
      icon: "IconHeartbeat",
    },
    {
      id: "project-overview",
      label: "Project Overview",
      description: "Overview of projects and their configuration",
      prompt:
        "Give me an overview of my Supabase projects — name, region, plan, and status for each.",
      icon: "IconDashboard",
    },
  ],
}

export function getScenariosForAgent(agentName: string): DemoScenario[] {
  return DEMO_SCENARIOS[agentName] ?? []
}
