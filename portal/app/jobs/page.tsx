import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { redirect } from "next/navigation"
import { AppSidebar } from "@/components/app-sidebar"

import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { IconCalendarEvent } from "@tabler/icons-react"
import { getNextRun, toHumanReadable } from "@/lib/cron"
import { JobCard } from "./job-card"

interface AgentTemplate {
  id: string
  display_name: string
}

interface UserAgentRow {
  template_id: string
  status: string
}

interface JobDefinition {
  id: string
  template_id: string
  display_name: string
  description: string
  schedule: string
  agent_templates: AgentTemplate | AgentTemplate[]
}

interface JobRun {
  id: string
  status: "pending" | "running" | "completed" | "failed" | "timed_out"
  started_at: string | null
  completed_at: string | null
  created_at: string
}

interface EnrichedJob {
  definition: JobDefinition
  agentName: string
  lastRun: JobRun | null
  nextRun: Date
}

export default async function JobsPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    redirect("/login")
  }

  const admin = createAdminClient()

  // Fetch approved user agents to get template IDs
  const { data: userAgents } = await supabase
    .from("user_agents")
    .select("template_id, status")
    .eq("user_id", user.id)
    .eq("status", "approved")

  const approvedTemplateIds = ((userAgents ?? []) as UserAgentRow[]).map((ua) => ua.template_id)

  let enrichedJobs: EnrichedJob[] = []

  if (approvedTemplateIds.length > 0) {
    // Fetch enabled job definitions for the user's approved templates
    const { data: jobDefs } = await admin
      .from("job_definitions")
      .select("id, template_id, display_name, description, schedule, agent_templates(id, display_name)")
      .in("template_id", approvedTemplateIds)
      .eq("enabled", true)

    const definitions = (jobDefs ?? []) as JobDefinition[]

    // For each job, fetch the most recent job run for this user
    enrichedJobs = await Promise.all(
      definitions.map(async (def) => {
        const { data: runs } = await supabase
          .from("job_runs")
          .select("id, status, started_at, completed_at, created_at")
          .eq("job_definition_id", def.id)
          .eq("user_id", user.id)
          .order("created_at", { ascending: false })
          .limit(1)

        const lastRun = ((runs ?? []) as JobRun[])[0] ?? null

        const agentTemplate = Array.isArray(def.agent_templates)
          ? def.agent_templates[0]
          : def.agent_templates
        const agentName = agentTemplate?.display_name ?? "Unknown Agent"

        const nextRun = getNextRun(def.schedule)

        return { definition: def, agentName, lastRun, nextRun }
      })
    )
  }

  return (
    <SidebarProvider>
      <AppSidebar
        user={{
          email: user.email ?? undefined,
          name: user.user_metadata?.full_name ?? user.user_metadata?.name,
        }}
      />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator
            orientation="vertical"
            className="mr-2 data-vertical:h-4 data-vertical:self-auto"
          />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>Jobs</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Jobs</h1>
            <p className="text-sm text-muted-foreground">
              Scheduled tasks that run automatically on your connected integrations.
            </p>
          </div>

          {enrichedJobs.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-4 py-16 text-center">
              <IconCalendarEvent className="size-12 text-muted-foreground/50" />
              <div>
                <p className="text-sm font-medium">No scheduled jobs.</p>
                <p className="text-sm text-muted-foreground mt-0.5">
                  Jobs become available once you have approved agents with integrations connected.
                </p>
              </div>
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {enrichedJobs.map(({ definition, agentName, lastRun, nextRun }) => (
                <JobCard
                  key={definition.id}
                  definitionId={definition.id}
                  displayName={definition.display_name}
                  description={definition.description}
                  agentDisplayName={agentName}
                  schedule={definition.schedule}
                  scheduleHuman={toHumanReadable(definition.schedule)}
                  lastRun={lastRun ? {
                    id: lastRun.id,
                    status: lastRun.status,
                    startedAt: lastRun.started_at,
                    completedAt: lastRun.completed_at,
                  } : null}
                  nextRun={nextRun.toISOString()}
                />
              ))}
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
