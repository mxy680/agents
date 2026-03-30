"use client"

import { useState } from "react"
import Link from "next/link"
import { Badge } from "@/components/ui/badge"
import { IconTrash, IconLoader2 } from "@tabler/icons-react"

interface Run {
  id: string
  job_slug: string
  status: string
  started_at: string | null
  completed_at: string | null
  created_at: string
}

function formatDate(iso: string | null): string {
  if (!iso) return "--"
  return new Date(iso).toLocaleString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  })
}

function formatDuration(
  startedAt: string | null,
  completedAt: string | null
): string {
  if (!startedAt || !completedAt) return "--"
  const ms = new Date(completedAt).getTime() - new Date(startedAt).getTime()
  const seconds = Math.round(ms / 1000)
  if (seconds < 60) return `${seconds}s`
  const minutes = Math.floor(seconds / 60)
  const remaining = seconds % 60
  return `${minutes}m ${remaining}s`
}

function StatusBadge({ status }: { status: string }) {
  const styles: Record<string, string> = {
    completed: "bg-green-500/20 text-green-400 border-green-500/30",
    failed: "bg-red-500/20 text-red-400 border-red-500/30",
    running: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30",
  }
  return (
    <Badge
      variant="outline"
      className={`text-xs ${styles[status] ?? "text-muted-foreground border-muted-foreground/30"}`}
    >
      {status}
    </Badge>
  )
}

export function RunHistory({ initialRuns }: { initialRuns: Run[] }) {
  const [runs, setRuns] = useState(initialRuns)
  const [deleting, setDeleting] = useState<string | null>(null)

  async function handleDelete(e: React.MouseEvent, runId: string) {
    e.preventDefault()
    e.stopPropagation()
    if (deleting) return
    setDeleting(runId)
    try {
      const res = await fetch(`/api/jobs/${runId}`, { method: "DELETE" })
      if (res.ok) {
        setRuns((prev) => prev.filter((r) => r.id !== runId))
      }
    } finally {
      setDeleting(null)
    }
  }

  if (runs.length === 0) {
    return (
      <p className="text-sm text-muted-foreground py-8 text-center">
        No runs yet. Click Run to start the first one.
      </p>
    )
  }

  return (
    <div className="flex flex-col border border-border rounded-lg divide-y divide-border">
      {runs.map((run) => (
        <div key={run.id} className="flex items-center justify-between p-3 hover:bg-muted/40 transition-colors">
          <Link
            href={`/jobs/local/${run.id}`}
            className="flex items-center gap-3 flex-1 min-w-0"
          >
            <StatusBadge status={run.status} />
            <div className="min-w-0">
              <p className="text-sm capitalize">
                {run.job_slug.replace(/-/g, " ")}
              </p>
              <p className="text-xs text-muted-foreground">
                {formatDate(run.started_at ?? run.created_at)}
              </p>
            </div>
          </Link>
          <div className="flex items-center gap-3 shrink-0">
            <span className="text-xs text-muted-foreground">
              {formatDuration(run.started_at, run.completed_at)}
            </span>
            <button
              onClick={(e) => handleDelete(e, run.id)}
              disabled={deleting === run.id}
              className="text-muted-foreground hover:text-red-400 transition-colors p-1"
              title="Delete run"
            >
              {deleting === run.id ? (
                <IconLoader2 className="size-3.5 animate-spin" />
              ) : (
                <IconTrash className="size-3.5" />
              )}
            </button>
          </div>
        </div>
      ))}
    </div>
  )
}
