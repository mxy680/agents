"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { IconMessageCircle, IconDownload, IconLoader2 } from "@tabler/icons-react"

interface AgentDetailActionsProps {
  templateId: string
  agentName: string
  isAcquired: boolean
}

export function AgentDetailActions({
  templateId,
  agentName,
  isAcquired: initialIsAcquired,
}: AgentDetailActionsProps) {
  const [isAcquired, setIsAcquired] = useState(initialIsAcquired)
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
        setIsAcquired(true)
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

  return (
    <div className="flex flex-col gap-2">
      {error && (
        <p className="text-xs text-destructive">{error}</p>
      )}
      {isAcquired ? (
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
