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
import { Button } from "@/components/ui/button"
import { IconArrowLeft } from "@tabler/icons-react"
import { RunScanButton } from "../../run-scan-button"
import { RunHistory } from "./run-history"

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

  const { data: template } = await admin
    .from("agent_templates")
    .select("id, name, display_name, description, status")
    .eq("name", name)
    .single()

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

  const { data: localRuns } = await admin
    .from("local_job_runs")
    .select(
      "id, agent_name, job_slug, status, started_at, completed_at, created_at"
    )
    .eq("agent_name", name)
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
                <BreadcrumbPage>{displayName}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>

        <div className="flex flex-1 flex-col gap-6 p-6 max-w-4xl">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">
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

          {/* Available jobs */}
          {jobs.length > 0 && (
            <div className="flex flex-col gap-3">
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
            <RunHistory initialRuns={runs} />
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
