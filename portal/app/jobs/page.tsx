import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { redirect } from "next/navigation"
import Link from "next/link"
import { AppSidebar } from "@/components/app-sidebar"
import { isAdmin } from "@/lib/admin"
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
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { IconCalendarEvent, IconHistory, IconFileText } from "@tabler/icons-react"
import { getNextRun, toHumanReadable } from "@/lib/cron"

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

function formatDateShort(iso: string | null): string {
  if (!iso) return "—"
  return new Date(iso).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  })
}

function StatusBadge({ status }: { status: JobRun["status"] | null }) {
  if (!status) {
    return (
      <Badge variant="outline" className="text-xs text-muted-foreground border-muted-foreground/30">
        Never run
      </Badge>
    )
  }
  switch (status) {
    case "completed":
      return (
        <Badge variant="outline" className="text-xs bg-green-500/20 text-green-400 border-green-500/30">
          Completed
        </Badge>
      )
    case "failed":
    case "timed_out":
      return (
        <Badge variant="outline" className="text-xs bg-red-500/20 text-red-400 border-red-500/30">
          {status === "timed_out" ? "Timed out" : "Failed"}
        </Badge>
      )
    case "running":
      return (
        <Badge variant="outline" className="text-xs bg-yellow-500/20 text-yellow-400 border-yellow-500/30">
          Running
        </Badge>
      )
    case "pending":
      return (
        <Badge variant="outline" className="text-xs bg-muted text-muted-foreground border-muted-foreground/30">
          Pending
        </Badge>
      )
  }
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

  const userIsAdmin = isAdmin(user.email)

  return (
    <SidebarProvider>
      <AppSidebar
        user={{
          email: user.email ?? undefined,
          name: user.user_metadata?.full_name ?? user.user_metadata?.name,
        }}
        isAdmin={userIsAdmin}
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
                <Card key={definition.id} className="flex flex-col">
                  <CardHeader>
                    <div className="flex items-start gap-3">
                      <div className="flex size-10 shrink-0 items-center justify-center bg-muted">
                        <IconCalendarEvent className="size-5" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <CardTitle className="text-base">{definition.display_name}</CardTitle>
                        <p className="text-xs text-muted-foreground mt-0.5">{agentName}</p>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="flex flex-1 flex-col gap-3">
                    <CardDescription className="flex-1">{definition.description}</CardDescription>

                    <div className="flex flex-col gap-1.5 text-xs">
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">Schedule</span>
                        <span className="font-medium">{toHumanReadable(definition.schedule)}</span>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">Last run</span>
                        <div className="flex items-center gap-1.5">
                          <StatusBadge status={lastRun?.status ?? null} />
                          {lastRun && (
                            <span className="text-muted-foreground">
                              {formatDateShort(lastRun.completed_at ?? lastRun.started_at)}
                            </span>
                          )}
                        </div>
                      </div>
                      <div className="flex items-center justify-between">
                        <span className="text-muted-foreground">Next run</span>
                        <span>{formatDateShort(nextRun.toISOString())}</span>
                      </div>
                    </div>

                    <div className="flex gap-2">
                      {lastRun?.status === "completed" && (
                        <Button asChild size="sm" className="flex-1">
                          <Link href={`/jobs/${lastRun.id}`}>
                            <IconFileText className="size-3.5" />
                            View Output
                          </Link>
                        </Button>
                      )}
                      <Button asChild size="sm" variant="outline" className="flex-1">
                        <Link href={`/jobs/history/${definition.id}`}>
                          <IconHistory className="size-3.5" />
                          View History
                        </Link>
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
