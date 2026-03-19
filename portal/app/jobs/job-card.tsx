"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { IconCalendarEvent, IconHistory, IconFileText, IconPlayerPlay, IconLoader2 } from "@tabler/icons-react"

type RunStatus = "pending" | "running" | "completed" | "failed" | "timed_out"

export interface JobCardProps {
  definitionId: string
  displayName: string
  description: string
  agentDisplayName: string
  schedule: string
  scheduleHuman: string
  lastRun: {
    id: string
    status: string
    startedAt: string | null
    completedAt: string | null
  } | null
  nextRun: string // ISO string
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

function StatusBadge({ status }: { status: RunStatus | null }) {
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

export function JobCard({
  definitionId,
  displayName,
  description,
  agentDisplayName,
  scheduleHuman,
  lastRun,
  nextRun,
}: JobCardProps) {
  const router = useRouter()
  const [triggering, setTriggering] = useState(false)

  const lastRunStatus = (lastRun?.status ?? null) as RunStatus | null
  const isInProgress = lastRunStatus === "pending" || lastRunStatus === "running"
  const playDisabled = isInProgress || triggering

  async function handleRunNow() {
    if (playDisabled) return
    setTriggering(true)
    try {
      const res = await fetch("/api/jobs/trigger", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ jobDefinitionId: definitionId }),
      })
      if (res.ok) {
        router.refresh()
      }
    } finally {
      setTriggering(false)
    }
  }

  return (
    <Card className="flex flex-col">
      <CardHeader>
        <div className="flex items-start gap-3">
          <div className="flex size-10 shrink-0 items-center justify-center bg-muted">
            <IconCalendarEvent className="size-5" />
          </div>
          <div className="flex-1 min-w-0">
            <CardTitle className="text-base">{displayName}</CardTitle>
            <p className="text-xs text-muted-foreground mt-0.5">{agentDisplayName}</p>
          </div>
        </div>
      </CardHeader>
      <CardContent className="flex flex-1 flex-col gap-3">
        <CardDescription className="flex-1">{description}</CardDescription>

        <div className="flex flex-col gap-1.5 text-xs">
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground">Schedule</span>
            <span className="font-medium">{scheduleHuman}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground">Last run</span>
            <div className="flex items-center gap-1.5">
              <StatusBadge status={lastRunStatus} />
              {lastRun && (
                <span className="text-muted-foreground">
                  {formatDateShort(lastRun.completedAt ?? lastRun.startedAt)}
                </span>
              )}
            </div>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-muted-foreground">Next run</span>
            <span>{formatDateShort(nextRun)}</span>
          </div>
        </div>

        <div className="flex gap-2">
          {lastRunStatus === "completed" && (
            <Button asChild size="sm" className="flex-1">
              <Link href={`/jobs/${lastRun!.id}`}>
                <IconFileText className="size-3.5" />
                View Output
              </Link>
            </Button>
          )}
          <Button asChild size="sm" variant="outline" className="flex-1">
            <Link href={`/jobs/history/${definitionId}`}>
              <IconHistory className="size-3.5" />
              View History
            </Link>
          </Button>
          <Button
            size="sm"
            variant="outline"
            disabled={playDisabled}
            onClick={handleRunNow}
            aria-label="Run now"
          >
            {triggering || isInProgress ? (
              <IconLoader2 className="size-3.5 animate-spin" />
            ) : (
              <IconPlayerPlay className="size-3.5" />
            )}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
