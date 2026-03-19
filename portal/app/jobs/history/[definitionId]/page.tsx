import { createClient } from "@/lib/supabase/server"
import { redirect, notFound } from "next/navigation"
import Link from "next/link"
import { AppSidebar } from "@/components/app-sidebar"
import { isAdmin } from "@/lib/admin"
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { IconCalendarEvent, IconExternalLink } from "@tabler/icons-react"

interface JobDefinition {
  id: string
  display_name: string
  description: string
  schedule: string
}

interface JobRun {
  id: string
  status: "pending" | "running" | "completed" | "failed" | "timed_out"
  output_markdown: string | null
  error_message: string | null
  started_at: string | null
  completed_at: string | null
  duration_ms: number | null
  created_at: string
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
    case "timed_out":
      return (
        <Badge variant="outline" className="text-xs bg-red-500/20 text-red-400 border-red-500/30">
          Timed out
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
        <Badge variant="outline" className="text-xs text-muted-foreground border-muted-foreground/30">
          Pending
        </Badge>
      )
  }
}

function outputPreview(run: JobRun): string {
  if (run.status === "failed" || run.status === "timed_out") {
    return run.error_message?.slice(0, 100) ?? "—"
  }
  if (!run.output_markdown) return "—"
  const preview = run.output_markdown.slice(0, 100)
  return run.output_markdown.length > 100 ? preview + "…" : preview
}

export default async function JobHistoryPage({
  params,
}: {
  params: Promise<{ definitionId: string }>
}) {
  const { definitionId } = await params
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    redirect("/login")
  }

  // Fetch job definition
  const { data: jobDef } = await supabase
    .from("job_definitions")
    .select("id, display_name, description, schedule")
    .eq("id", definitionId)
    .single()

  if (!jobDef) {
    notFound()
  }

  const typedJobDef = jobDef as JobDefinition

  // Fetch all runs for this user + definition, most recent first
  const { data: runs } = await supabase
    .from("job_runs")
    .select("id, status, output_markdown, error_message, started_at, completed_at, duration_ms, created_at")
    .eq("job_definition_id", definitionId)
    .eq("user_id", user.id)
    .order("created_at", { ascending: false })
    .limit(50)

  const typedRuns = (runs ?? []) as JobRun[]

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
                <BreadcrumbLink href="/jobs">Jobs</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbLink href="/jobs">{typedJobDef.display_name}</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage>History</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">{typedJobDef.display_name}</h1>
            <p className="text-sm text-muted-foreground mt-0.5">Run history (last 50 runs)</p>
          </div>

          {typedRuns.length === 0 ? (
            <div className="flex flex-col items-center justify-center gap-4 py-16 text-center">
              <IconCalendarEvent className="size-12 text-muted-foreground/50" />
              <div>
                <p className="text-sm font-medium">No runs yet.</p>
                <p className="text-sm text-muted-foreground mt-0.5">
                  This job has not run yet. Check back after the next scheduled time.
                </p>
              </div>
            </div>
          ) : (
            <div className="border border-border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Date</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Duration</TableHead>
                    <TableHead className="max-w-xs">Output Preview</TableHead>
                    <TableHead className="w-20">Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {typedRuns.map((run) => (
                    <TableRow key={run.id}>
                      <TableCell className="text-sm whitespace-nowrap">
                        {formatDate(run.started_at ?? run.created_at)}
                      </TableCell>
                      <TableCell>
                        <StatusBadge status={run.status} />
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground whitespace-nowrap">
                        {formatDuration(run.duration_ms)}
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground max-w-xs truncate">
                        {outputPreview(run)}
                      </TableCell>
                      <TableCell>
                        <Button asChild size="sm" variant="ghost">
                          <Link href={`/jobs/${run.id}`}>
                            <IconExternalLink className="size-3.5" />
                            View
                          </Link>
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
