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
import { RunScanButton } from "../run-scan-button"
import { RunHistory } from "../agent/[name]/run-history"

export default async function JobPage({
  params,
}: {
  params: Promise<{ slug: string }>
}) {
  const { slug } = await params

  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()
  if (!user) redirect("/login")

  const admin = createAdminClient()

  // Find the job definition by slug
  const { data: jobDef } = await admin
    .from("job_definitions")
    .select(
      "id, slug, display_name, description, schedule, estimated_minutes, template_id, agent_templates(name, display_name)"
    )
    .eq("slug", slug)
    .single()

  if (!jobDef) redirect("/jobs")

  const template = Array.isArray(jobDef.agent_templates)
    ? jobDef.agent_templates[0]
    : jobDef.agent_templates
  const agentName = (template as { name: string })?.name ?? "unknown"
  const agentDisplayName =
    (template as { display_name: string })?.display_name ?? "Unknown Agent"

  // Fetch runs for this specific job
  const { data: localRuns } = await admin
    .from("local_job_runs")
    .select(
      "id, agent_name, job_slug, status, started_at, completed_at, created_at"
    )
    .eq("agent_name", agentName)
    .eq("job_slug", slug)
    .order("created_at", { ascending: false })
    .limit(20)

  const runs = (localRuns ?? []) as Array<{
    id: string
    job_slug: string
    status: string
    started_at: string | null
    completed_at: string | null
    created_at: string
  }>

  return (
    <SidebarProvider>
      <AppSidebar />
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
                <BreadcrumbPage>{jobDef.display_name}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>

        <div className="flex flex-1 flex-col gap-6 p-6 max-w-4xl">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">
                {jobDef.display_name}
              </h1>
              <p className="text-sm text-muted-foreground mt-1">
                {jobDef.description}
              </p>
              <div className="flex items-center gap-3 mt-2">
                <Badge variant="outline" className="text-xs">
                  {agentDisplayName}
                </Badge>
                {jobDef.schedule && (
                  <Badge variant="outline" className="text-xs">
                    {toHumanReadable(jobDef.schedule)}
                  </Badge>
                )}
                {(jobDef as unknown as { estimated_minutes?: number }).estimated_minutes && (
                  <Badge variant="outline" className="text-xs">
                    ~{(jobDef as unknown as { estimated_minutes?: number }).estimated_minutes} min
                  </Badge>
                )}
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Button asChild variant="outline" size="sm">
                <Link href="/jobs">
                  <IconArrowLeft className="size-4" />
                  All Jobs
                </Link>
              </Button>
              <RunScanButton agent={agentName} job={slug} />
            </div>
          </div>

          <div className="flex flex-col gap-3">
            <h2 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
              Run History
            </h2>
            <RunHistory initialRuns={runs} />
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
