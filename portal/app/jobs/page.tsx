import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { redirect } from "next/navigation"
import Link from "next/link"
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
import { Badge } from "@/components/ui/badge"
import { IconCalendarEvent } from "@tabler/icons-react"
import { getNextRun, toHumanReadable } from "@/lib/cron"
import { JobCard } from "./job-card"
import { RunScanButton } from "./run-scan-button"

interface LocalJobRun {
  id: string
  agent_name: string
  job_slug: string
  status: string
  started_at: string | null
  completed_at: string | null
  created_at: string
}

interface AgentTemplate {
  id: string
  display_name: string
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

function formatDateShort(iso: string | null): string {
  if (!iso) return "—"
  return new Date(iso).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  })
}

function formatDuration(startedAt: string | null, completedAt: string | null): string {
  if (!startedAt || !completedAt) return "—"
  const ms = new Date(completedAt).getTime() - new Date(startedAt).getTime()
  const seconds = Math.round(ms / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remaining = seconds % 60
  return `${minutes}m ${remaining}s`
}

function LocalRunStatusBadge({ status }: { status: string }) {
  switch (status) {
    case "completed":
      return (
        <Badge variant="outline" className="text-xs bg-green-500/20 text-green-400 border-green-500/30">
          Completed
        </Badge>
      )
    case "failed":
      return (
        <Badge variant="outline" className="text-xs bg-red-500/20 text-red-400 border-red-500/30">
          Failed
        </Badge>
      )
    case "running":
      return (
        <Badge variant="outline" className="text-xs bg-yellow-500/20 text-yellow-400 border-yellow-500/30">
          Running
        </Badge>
      )
    case "pending":
    default:
      return (
        <Badge variant="outline" className="text-xs text-muted-foreground border-muted-foreground/30">
          Pending
        </Badge>
      )
  }
}

function LocalRunRow({ run }: { run: LocalJobRun }) {
  return (
    <Link
      href={`/jobs/local/${run.id}`}
      className="flex items-center justify-between py-2.5 px-1 hover:bg-muted/40 rounded transition-colors text-sm"
    >
      <div className="flex items-center gap-3">
        <LocalRunStatusBadge status={run.status} />
        <span className="text-muted-foreground">
          {formatDateShort(run.started_at ?? run.created_at)}
        </span>
      </div>
      <span className="text-muted-foreground text-xs">
        {formatDuration(run.started_at, run.completed_at)}
      </span>
    </Link>
  )
}

export default async function JobsPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    redirect("/login")
  }

  const admin = createAdminClient()

  // Fetch all enabled job definitions (admin-only tool, no per-user filtering)
  const { data: jobDefs } = await admin
    .from("job_definitions")
    .select("id, template_id, display_name, description, schedule, agent_templates(id, display_name)")
    .eq("enabled", true)

  const definitions = (jobDefs ?? []) as JobDefinition[]

  // For each job, fetch the most recent job run for this user
  const enrichedJobs: EnrichedJob[] = await Promise.all(
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

  // Fetch recent local job runs (real-estate pipeline)
  const { data: localRuns } = await admin
    .from("local_job_runs")
    .select("id, agent_name, job_slug, status, started_at, completed_at, created_at")
    .eq("agent_name", "real-estate")
    .order("created_at", { ascending: false })
    .limit(20)

  const recentLocalRuns = (localRuns ?? []) as LocalJobRun[]

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

          {/* NYC Assemblage Scan — manual trigger */}
          <div className="flex flex-col gap-4 border border-border rounded-lg p-4">
            <div>
              <h2 className="text-base font-semibold">NYC Assemblage Scan</h2>
              <p className="text-sm text-muted-foreground mt-0.5">
                Run the full NYC assemblage intelligence pipeline locally with live log streaming.
              </p>
            </div>
            <RunScanButton />

            {recentLocalRuns.length > 0 && (
              <div className="flex flex-col gap-2">
                <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
                  Recent Runs
                </h3>
                <div className="flex flex-col divide-y divide-border">
                  {recentLocalRuns.map((run) => (
                    <LocalRunRow key={run.id} run={run} />
                  ))}
                </div>
              </div>
            )}
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
