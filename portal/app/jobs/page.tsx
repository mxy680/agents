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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { IconCalendarEvent, IconPlayerPlay } from "@tabler/icons-react"
import { toHumanReadable } from "@/lib/cron"

interface AgentTemplate {
  id: string
  name: string
  display_name: string
}

interface JobDefinition {
  id: string
  slug: string
  template_id: string
  display_name: string
  description: string
  schedule: string
  enabled: boolean
  agent_templates: AgentTemplate | AgentTemplate[]
}

interface LocalJobRun {
  id: string
  agent_name: string
  job_slug: string
  status: string
  started_at: string | null
  completed_at: string | null
}

function StatusDot({ status }: { status: string }) {
  const color =
    status === "completed"
      ? "bg-green-500"
      : status === "running"
        ? "bg-yellow-500"
        : status === "failed"
          ? "bg-red-500"
          : "bg-muted-foreground"
  return <span className={`inline-block size-2 rounded-full ${color}`} />
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

export default async function JobsPage() {
  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()

  if (!user) {
    redirect("/login")
  }

  const admin = createAdminClient()

  // Fetch all job definitions with their agent template
  const { data: jobDefs } = await admin
    .from("job_definitions")
    .select(
      "id, slug, template_id, display_name, description, schedule, estimated_minutes, enabled, agent_templates(id, name, display_name)"
    )
    .eq("enabled", true)

  const definitions = (jobDefs ?? []) as JobDefinition[]

  // Fetch latest local run per job slug
  const { data: localRuns } = await admin
    .from("local_job_runs")
    .select("id, agent_name, job_slug, status, started_at, completed_at")
    .order("created_at", { ascending: false })
    .limit(50)

  const latestRunByJob: Record<string, LocalJobRun> = {}
  for (const run of (localRuns ?? []) as LocalJobRun[]) {
    const key = `${run.agent_name}:${run.job_slug}`
    if (!latestRunByJob[key]) {
      latestRunByJob[key] = run
    }
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
              Scheduled tasks across all agents.
            </p>
          </div>

          {definitions.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-4 py-16 text-center">
              <IconCalendarEvent className="size-12 text-muted-foreground/50" />
              <div>
                <p className="text-sm font-medium">No jobs configured.</p>
                <p className="text-sm text-muted-foreground mt-0.5">
                  Jobs appear once agents are set up with templates.
                </p>
              </div>
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 auto-rows-fr">
              {definitions.map((def) => {
                const template = Array.isArray(def.agent_templates)
                  ? def.agent_templates[0]
                  : def.agent_templates
                const agentName = template?.name ?? "unknown"
                const agentDisplayName =
                  template?.display_name ?? "Unknown Agent"
                const latestRun = latestRunByJob[`${agentName}:${def.slug}`]

                return (
                  <Link
                    key={def.id}
                    href={`/jobs/${def.slug}`}
                    className="block"
                  >
                    <Card className="h-full hover:border-foreground/20 transition-colors cursor-pointer">
                      <CardHeader>
                        <div className="flex items-start gap-3">
                          <div className="flex size-10 shrink-0 items-center justify-center bg-muted">
                            <IconPlayerPlay className="size-5" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <CardTitle className="text-base">
                              {def.display_name}
                            </CardTitle>
                            <p className="text-xs text-muted-foreground mt-0.5">
                              {agentDisplayName}
                            </p>
                          </div>
                        </div>
                      </CardHeader>
                      <CardContent className="flex flex-col gap-3">
                        <p className="text-sm text-muted-foreground line-clamp-2">
                          {def.description}
                        </p>

                        <div className="flex flex-col gap-1.5 text-xs">
                          {def.schedule && (
                            <div className="flex items-center justify-between">
                              <span className="text-muted-foreground">
                                Schedule
                              </span>
                              <Badge variant="outline" className="text-xs">
                                {toHumanReadable(def.schedule)}
                              </Badge>
                            </div>
                          )}
                          {(def as Record<string, unknown>).estimated_minutes && (
                            <div className="flex items-center justify-between">
                              <span className="text-muted-foreground">
                                Est. time
                              </span>
                              <span>~{(def as Record<string, unknown>).estimated_minutes} min</span>
                            </div>
                          )}
                          {latestRun && (
                            <div className="flex items-center justify-between">
                              <span className="text-muted-foreground">
                                Last run
                              </span>
                              <div className="flex items-center gap-1.5">
                                <StatusDot status={latestRun.status} />
                                <span>
                                  {formatDateShort(
                                    latestRun.completed_at ??
                                      latestRun.started_at
                                  )}
                                </span>
                              </div>
                            </div>
                          )}
                        </div>
                      </CardContent>
                    </Card>
                  </Link>
                )
              })}
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
