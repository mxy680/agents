"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { IconMessageCircle, IconDownload, IconLoader2 } from "@tabler/icons-react"

type AcquisitionStatus = "pending" | "approved" | "rejected"

interface AgentDetailActionsProps {
  templateId: string
  agentName: string
  isAcquired: boolean
  initialAcquisitionStatus: AcquisitionStatus | null
}

export function AgentDetailActions({
  templateId,
  agentName,
  isAcquired: initialIsAcquired,
  initialAcquisitionStatus,
}: AgentDetailActionsProps) {
  const [isAcquired, setIsAcquired] = useState(initialIsAcquired)
  const [acquisitionStatus, setAcquisitionStatus] = useState<AcquisitionStatus | null>(
    initialAcquisitionStatus
  )
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleAcquire() {
    setIsLoading(true)
    setError(null)
    try {
      const res = await fetch("/api/agents/acquire", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ templateId }),
      })
      if (res.ok) {
        const data = await res.json().catch(() => ({}))
        setIsAcquired(true)
        setAcquisitionStatus(data.status ?? "pending")
      } else {
        const data = await res.json().catch(() => ({}))
        setError(data.error ?? "Failed to acquire agent")
      }
    } catch {
      setError("Network error. Please try again.")
    } finally {
      setIsLoading(false)
    }
  }

  const isPending = isAcquired && acquisitionStatus === "pending"
  const isApproved = isAcquired && acquisitionStatus === "approved"
  const isRejected = isAcquired && acquisitionStatus === "rejected"

  return (
    <div className="flex flex-col gap-2">
      {error && (
        <p className="text-xs text-destructive">{error}</p>
      )}
      {isPending ? (
        <div className="flex flex-col gap-1.5">
          <Button
            size="lg"
            className="w-full sm:w-fit bg-yellow-500/20 text-yellow-400 border border-yellow-500/30 hover:bg-yellow-500/20 cursor-default"
            disabled
          >
            Pending Approval
          </Button>
          <p className="text-xs text-muted-foreground">
            Your request is awaiting admin review.
          </p>
        </div>
      ) : isRejected ? (
        <div className="flex flex-col gap-1.5">
          <Button size="lg" className="w-full sm:w-fit" variant="outline" disabled>
            Access Denied
          </Button>
          <p className="text-xs text-muted-foreground">
            Your request to access this agent was not approved.
          </p>
        </div>
      ) : isApproved ? (
        <Button asChild size="lg" className="w-full sm:w-fit">
          <a href={`/chat/${agentName}`}>
            <IconMessageCircle className="size-5" />
            Open Chat
          </a>
        </Button>
      ) : (
        <Button
          size="lg"
          className="w-full sm:w-fit"
          onClick={handleAcquire}
          disabled={isLoading}
        >
          {isLoading ? (
            <IconLoader2 className="size-5 animate-spin" />
          ) : (
            <IconDownload className="size-5" />
          )}
          {isLoading ? "Getting…" : "Get for Free"}
        </Button>
      )}
    </div>
  )
}
