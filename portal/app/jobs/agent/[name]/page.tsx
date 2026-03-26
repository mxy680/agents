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
  BreadcrumbSeparator,
  BreadcrumbLink,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { IconArrowLeft } from "@tabler/icons-react"
import { toHumanReadable } from "@/lib/cron"
import { RunScanButton } from "../../run-scan-button"

interface LocalJobRun {
  id: string
  agent_name: string
  job_slug: string
  status: string
  started_at: string | null
  completed_at: string | null
  created_at: string
}

function formatDate(iso: string | null): string {
  if (!iso) return "—"
  return new Date(iso).toLocaleString(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  })
}

function formatDuration(
  startedAt: string | null,
  completedAt: string | null
): string {
  if (!startedAt || !completedAt) return "—"
  const ms =
    new Date(completedAt).getTime() - new Date(startedAt).getTime()
  const seconds = Math.round(ms / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remaining = seconds % 60
  return `${minutes}m ${remaining}s`
}

function StatusBadge({ status }: { status: string }) {
  switch (status) {
    case "completed":
      return (
        <Badge
          variant="outline"
          className="text-xs bg-green-500/20 text-green-400 border-green-500/30"
        >
          Completed
        </Badge>
      )
    case "failed":
      return (
        <Badge
          variant="outline"
          className="text-xs bg-red-500/20 text-red-400 border-red-500/30"
        >
          Failed
        </Badge>
      )
    case "running":
      return (
        <Badge
          variant="outline"
          className="text-xs bg-yellow-500/20 text-yellow-400 border-yellow-500/30"
        >
          Running
        </Badge>
      )
    default:
      return (
        <Badge
          variant="outline"
          className="text-xs text-muted-foreground border-muted-foreground/30"
        >
          {status}
        </Badge>
      )
  }
}

export default async function AgentJobsPage({
  params,
}: {
  params: Promise<{ name: string }>
}) {
  const { name } = await params

  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()
  if (!user) redirect("/login")

  const admin = createAdminClient()

  // Get the agent template
  const { data: template } = await admin
    .from("agent_templates")
    .select("id, name, display_name, description, status")
    .eq("name", name)
    .single()

  // Get job definitions for this agent
  const { data: jobDefs } = await admin
    .from("job_definitions")
    .select("id, slug, display_name, description, schedule, enabled")
    .eq("template_id", template?.id ?? "")
    .eq("enabled", true)

  const jobs = (jobDefs ?? []) as Array<{
    id: string
    slug: string
    display_name: string
    description: string
    schedule: string
  }>

  // Get all local runs for this agent
  const { data: localRuns } = await admin
    .from("local_job_runs")
    .select(
      "id, agent_name, job_slug, status, started_at, completed_at, created_at"
    )
    .eq("agent_name", name)
    .order("created_at", { ascending: false })
    .limit(20)

  const runs = (localRuns ?? []) as LocalJobRun[]

  const displayName = template?.display_name ?? name.replace(/-/g, " ")

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
                <BreadcrumbLink href="/jobs">Jobs</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage className="capitalize">
                  {displayName}
                </BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>

        <div className="flex flex-1 flex-col gap-6 p-6 max-w-4xl">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight capitalize">
                {displayName}
              </h1>
              {template?.description && (
                <p className="text-sm text-muted-foreground mt-1 max-w-xl">
                  {template.description}
                </p>
              )}
            </div>
            <Button asChild variant="outline" size="sm">
              <Link href="/jobs">
                <IconArrowLeft className="size-4" />
                All Jobs
              </Link>
            </Button>
          </div>

          {/* Job definitions */}
          {jobs.length > 0 && (
            <div className="flex flex-col gap-3">
              <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                Available Jobs
              </h2>
              {jobs.map((job) => (
                <div
                  key={job.id}
                  className="flex items-center justify-between border border-border rounded-lg p-4"
                >
                  <div>
                    <p className="text-sm font-semibold">{job.display_name}</p>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      {job.description}
                    </p>
                    <Badge variant="outline" className="text-xs mt-2">
                      {toHumanReadable(job.schedule)}
                    </Badge>
                  </div>
                  <RunScanButton agent={name} job={job.slug} />
                </div>
              ))}
            </div>
          )}

          {/* Run history */}
          <div className="flex flex-col gap-3">
            <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
              Run History
            </h2>
            {runs.length === 0 ? (
              <p className="text-sm text-muted-foreground py-8 text-center">
                No runs yet. Click Run to start the first one.
              </p>
            ) : (
              <div className="flex flex-col border border-border rounded-lg divide-y divide-border">
                {runs.map((run) => (
                  <Link
                    key={run.id}
                    href={`/jobs/local/${run.id}`}
                    className="flex items-center justify-between p-3 hover:bg-muted/40 transition-colors"
                  >
                    <div className="flex items-center gap-3">
                      <StatusBadge status={run.status} />
                      <div>
                        <p className="text-sm">
                          {run.job_slug.replace(/-/g, " ")}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {formatDate(run.started_at ?? run.created_at)}
                        </p>
                      </div>
                    </div>
                    <span className="text-xs text-muted-foreground">
                      {formatDuration(run.started_at, run.completed_at)}
                    </span>
                  </Link>
                ))}
              </div>
            )}
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
