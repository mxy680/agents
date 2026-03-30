"use client"

import { use, useEffect, useState, useCallback } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { AppSidebar } from "@/components/app-sidebar"
import { LogViewer } from "@/components/log-viewer"

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
import {
  IconArrowLeft,
  IconLoader2,
  IconRefresh,
  IconFileSpreadsheet,
  IconFileText,
  IconExternalLink,
  IconDownload,
} from "@tabler/icons-react"

interface LocalJobRun {
  id: string
  agent_name?: string
  job_slug?: string
  status: string
  started_at: string | null
  completed_at: string | null
  deliverables: Record<string, string>
  log: string
  log_length: number
}

interface Deliverables {
  [key: string]: string
}

function formatDate(iso: string | null): string {
  if (!iso) return "—"
  return new Date(iso).toLocaleString(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
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

function StatusBadge({ status }: { status: string }) {
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
    case "running":
      return (
        <Badge variant="outline" className="bg-yellow-500/20 text-yellow-400 border-yellow-500/30 flex items-center gap-1.5">
          <IconLoader2 className="size-3 animate-spin" />
          Running
        </Badge>
      )
    case "pending":
      return (
        <Badge variant="outline" className="text-muted-foreground border-muted-foreground/30">
          Pending
        </Badge>
      )
    default:
      return (
        <Badge variant="outline" className="text-muted-foreground border-muted-foreground/30">
          {status}
        </Badge>
      )
  }
}

function DeliverableCards({ deliverables }: { deliverables: Deliverables }) {
  const entries = Object.entries(deliverables).filter(([, v]) => v && v.startsWith("http"))

  if (entries.length === 0) return null

  function isPortalArtifact(url: string): boolean {
    return url.includes("/storage/v1/object/public/job-artifacts/")
  }

  function getLabel(key: string, url: string): string {
    if (key.includes("pdf") || key.includes("report")) return "PDF Report"
    if (key.includes("sheet") || key.includes("spreadsheet")) {
      return isPortalArtifact(url) ? "Spreadsheet" : "Google Sheet"
    }
    if (key.includes("drive")) return "Google Drive"
    if (key.includes("dashboard")) return "Dashboard"
    return key.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase())
  }

  function getIcon(key: string) {
    if (key.includes("sheet") || key.includes("spreadsheet")) {
      return <IconFileSpreadsheet className="size-5 text-green-500" />
    }
    if (key.includes("pdf") || key.includes("report")) {
      return <IconFileText className="size-5 text-red-400" />
    }
    return <IconExternalLink className="size-5 text-blue-400" />
  }

  return (
    <div className="flex flex-col gap-3">
      <h2 className="text-sm font-semibold">Deliverables</h2>
      <div className="flex flex-wrap gap-3">
        {entries.map(([key, url]) => {
          const isDownload = isPortalArtifact(url)
          return (
            <a
              key={key}
              href={url}
              target="_blank"
              rel="noopener noreferrer"
              download={isDownload ? undefined : undefined}
              className="flex items-center gap-2 rounded-md border border-border bg-muted/30 px-4 py-3 hover:bg-muted/60 transition-colors"
            >
              {getIcon(key)}
              <span className="text-sm font-medium">{getLabel(key, url)}</span>
              {isDownload ? (
                <IconDownload className="size-3.5 text-muted-foreground ml-1" />
              ) : (
                <IconExternalLink className="size-3.5 text-muted-foreground ml-1" />
              )}
            </a>
          )
        })}
      </div>
    </div>
  )
}

export default function LocalJobRunPage({
  params,
}: {
  params: Promise<{ id: string }>
}) {
  const { id } = use(params)
  const router = useRouter()

  const [run, setRun] = useState<LocalJobRun | null>(null)
  const [loading, setLoading] = useState(true)
  const [deliverables, setDeliverables] = useState<Deliverables>({})
  const [liveStatus, setLiveStatus] = useState<string>("")
  const [rerunning, setRerunning] = useState(false)
  const [cancelling, setCancelling] = useState(false)

  useEffect(() => {
    async function fetchRun() {
      const res = await fetch(`/api/jobs/${id}`)
      if (res.ok) {
        const data = await res.json() as LocalJobRun
        setRun(data)
        setDeliverables(data.deliverables ?? {})
        setLiveStatus(data.status)
      }
      setLoading(false)
    }
    fetchRun()
  }, [id])

  const handleDone = useCallback((status: string, newDeliverables: Deliverables) => {
    setLiveStatus(status)
    setDeliverables(newDeliverables)
    // Refresh run metadata
    fetch(`/api/jobs/${id}`)
      .then((r) => r.json())
      .then((data: LocalJobRun) => setRun(data))
      .catch(() => {})
  }, [id])

  async function handleCancel() {
    if (cancelling) return
    setCancelling(true)
    try {
      const res = await fetch(`/api/jobs/${id}`, { method: "DELETE" })
      if (res.ok) {
        setLiveStatus("failed")
        // Refresh run metadata
        const refreshRes = await fetch(`/api/jobs/${id}`)
        if (refreshRes.ok) {
          const data = await refreshRes.json() as LocalJobRun
          setRun(data)
        }
      }
    } finally {
      setCancelling(false)
    }
  }

  async function handleRerun() {
    if (rerunning) return
    setRerunning(true)
    try {
      const res = await fetch("/api/jobs/run", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ agent: run?.agent_name ?? "real-estate", job: run?.job_slug ?? "off-market-scan" }),
      })
      if (res.ok) {
        const { runId } = await res.json() as { runId: string }
        router.push(`/jobs/local/${runId}`)
      }
    } finally {
      setRerunning(false)
    }
  }

  const currentStatus = liveStatus || run?.status || "pending"
  const isTerminal = currentStatus === "completed" || currentStatus === "failed"

  const runDate = run?.started_at
    ? new Date(run.started_at).toLocaleDateString(undefined, { dateStyle: "medium" })
    : "—"

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
                <BreadcrumbPage className="capitalize">{run?.job_slug?.replace(/-/g, " ") ?? "Job Run"}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>

        <div className="flex flex-1 flex-col gap-6 p-6 max-w-5xl">
          <div className="flex items-center justify-between">
            <Button asChild variant="outline" size="sm">
              <Link href="/jobs">
                <IconArrowLeft className="size-4" />
                Back to Jobs
              </Link>
            </Button>
            <div className="flex items-center gap-2">
              {(currentStatus === "running" || currentStatus === "pending") && (
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={handleCancel}
                  disabled={cancelling}
                >
                  {cancelling ? (
                    <IconLoader2 className="size-4 animate-spin" />
                  ) : (
                    "Cancel"
                  )}
                </Button>
              )}
              <Button
                variant="outline"
                size="sm"
                onClick={handleRerun}
                disabled={rerunning || currentStatus === "running" || currentStatus === "pending"}
              >
                {rerunning ? (
                  <IconLoader2 className="size-4 animate-spin" />
                ) : (
                  <IconRefresh className="size-4" />
                )}
                Rerun
              </Button>
            </div>
          </div>

          {loading ? (
            <div className="flex items-center gap-2 text-muted-foreground">
              <IconLoader2 className="size-4 animate-spin" />
              <span className="text-sm">Loading...</span>
            </div>
          ) : !run ? (
            <div className="text-sm text-muted-foreground">Run not found.</div>
          ) : (
            <>
              <div>
                <div className="flex items-center gap-3 mb-2">
                  <h1 className="text-2xl font-semibold tracking-tight">
                    <span className="capitalize">{run?.job_slug?.replace(/-/g, " ") ?? "Job Run"}</span> — {runDate}
                  </h1>
                  <StatusBadge status={currentStatus} />
                </div>
                <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
                  <span>Started: {formatDate(run.started_at)}</span>
                  {isTerminal && (
                    <>
                      <span>Completed: {formatDate(run.completed_at)}</span>
                      <span>Duration: {formatDuration(run.started_at, run.completed_at)}</span>
                    </>
                  )}
                </div>
              </div>

              <div className="flex flex-col gap-2">
                <h2 className="text-sm font-semibold">Logs</h2>
                <LogViewer
                  runId={id}
                  initialLog={run.log || ""}
                  initialStatus={run.status}
                  initialDeliverables={run.deliverables}
                  onDone={handleDone}
                />
              </div>

              {isTerminal && Object.keys(deliverables).length > 0 && (
                <DeliverableCards deliverables={deliverables} />
              )}
            </>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
