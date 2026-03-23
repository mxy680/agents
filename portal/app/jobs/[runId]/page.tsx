import { createClient } from "@/lib/supabase/server"
import { redirect, notFound } from "next/navigation"
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
import { MarkdownContent } from "@/components/markdown-content"

interface JobRun {
  id: string
  status: "pending" | "running" | "completed" | "failed" | "timed_out"
  output_markdown: string | null
  error_message: string | null
  started_at: string | null
  completed_at: string | null
  duration_ms: number | null
  created_at: string
  job_definition_id: string
}

interface JobDefinition {
  id: string
  display_name: string
}

function formatDate(iso: string | null): string {
  if (!iso) return "—"
  return new Date(iso).toLocaleString(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  })
}

function formatDuration(ms: number | null): string {
  if (!ms) return "—"
  const seconds = Math.round(ms / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remaining = seconds % 60
  return `${minutes}m ${remaining}s`
}

function StatusBadge({ status }: { status: JobRun["status"] }) {
  switch (status) {
    case "completed":
      return (
        <Badge variant="outline" className="bg-green-500/20 text-green-400 border-green-500/30">
          Completed
        </Badge>
      )
    case "failed":
      return (
        <Badge variant="outline" className="bg-red-500/20 text-red-400 border-red-500/30">
          Failed
        </Badge>
      )
    case "timed_out":
      return (
        <Badge variant="outline" className="bg-red-500/20 text-red-400 border-red-500/30">
          Timed out
        </Badge>
      )
    case "running":
      return (
        <Badge variant="outline" className="bg-yellow-500/20 text-yellow-400 border-yellow-500/30">
          Running
        </Badge>
      )
    case "pending":
      return (
        <Badge variant="outline" className="text-muted-foreground border-muted-foreground/30">
          Pending
        </Badge>
      )
  }
}

export default async function JobRunPage({
  params,
}: {
  params: Promise<{ runId: string }>
}) {
  const { runId } = await params
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    redirect("/login")
  }

  // Fetch run — RLS ensures user can only see their own runs
  const { data: run } = await supabase
    .from("job_runs")
    .select("id, status, output_markdown, error_message, started_at, completed_at, duration_ms, created_at, job_definition_id")
    .eq("id", runId)
    .single()

  if (!run) {
    notFound()
  }

  const typedRun = run as JobRun

  // Fetch parent job definition
  const { data: jobDef } = await supabase
    .from("job_definitions")
    .select("id, display_name")
    .eq("id", typedRun.job_definition_id)
    .single()

  const typedJobDef = jobDef as JobDefinition | null

  const runDate = formatDate(typedRun.started_at ?? typedRun.created_at)

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
              {typedJobDef && (
                <>
                  <BreadcrumbSeparator />
                  <BreadcrumbItem>
                    <BreadcrumbLink href={`/jobs/history/${typedJobDef.id}`}>
                      {typedJobDef.display_name}
                    </BreadcrumbLink>
                  </BreadcrumbItem>
                </>
              )}
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage>Run {runDate}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6 max-w-4xl">
          <div className="flex items-center justify-between">
            <Button asChild variant="outline" size="sm">
              <Link href="/jobs">
                <IconArrowLeft className="size-4" />
                Back to Jobs
              </Link>
            </Button>
          </div>

          <div>
            <div className="flex items-center gap-3 mb-2">
              <h1 className="text-2xl font-semibold tracking-tight">
                {typedJobDef?.display_name ?? "Job Run"}
              </h1>
              <StatusBadge status={typedRun.status} />
            </div>
            <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
              <span>Started: {formatDate(typedRun.started_at)}</span>
              <span>Completed: {formatDate(typedRun.completed_at)}</span>
              <span>Duration: {formatDuration(typedRun.duration_ms)}</span>
            </div>
          </div>

          {typedRun.status === "completed" && typedRun.output_markdown && (
            <div className="border border-border bg-muted/20 p-6">
              <h2 className="text-sm font-semibold mb-4">Output</h2>
              <MarkdownContent content={typedRun.output_markdown} className="text-sm" />
            </div>
          )}

          {typedRun.status === "completed" && !typedRun.output_markdown && (
            <div className="border border-border bg-muted/20 p-4">
              <p className="text-sm text-muted-foreground">No output was produced.</p>
            </div>
          )}

          {(typedRun.status === "failed" || typedRun.status === "timed_out") && (
            <div className="border border-destructive/50 bg-destructive/10 p-4">
              <h2 className="text-sm font-semibold text-destructive mb-2">
                {typedRun.status === "timed_out" ? "Job Timed Out" : "Job Failed"}
              </h2>
              {typedRun.error_message ? (
                <p className="text-sm text-destructive/80 font-mono whitespace-pre-wrap">
                  {typedRun.error_message}
                </p>
              ) : (
                <p className="text-sm text-destructive/80">No error details available.</p>
              )}
            </div>
          )}

          {(typedRun.status === "pending" || typedRun.status === "running") && (
            <div className="border border-border bg-muted/20 p-4">
              <p className="text-sm text-muted-foreground">
                This job is currently {typedRun.status}. Refresh the page to see updated results.
              </p>
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
