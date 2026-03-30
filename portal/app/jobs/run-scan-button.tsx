"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { IconPlayerPlay, IconLoader2 } from "@tabler/icons-react"

interface RunScanButtonProps {
  agent?: string
  job?: string
}

export function RunScanButton({
  agent = "real-estate",
  job = "off-market-scan",
}: RunScanButtonProps) {
  const router = useRouter()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleRun() {
    if (loading) return
    setLoading(true)
    setError(null)

    try {
      const res = await fetch("/api/jobs/run", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ agent, job }),
      })

      if (!res.ok) {
        const body = (await res.json().catch(() => ({}))) as {
          error?: string
        }
        throw new Error(body.error ?? `HTTP ${res.status}`)
      }

      const { runId } = (await res.json()) as { runId: string }
      router.push(`/jobs/local/${runId}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to start")
      setLoading(false)
    }
  }

  return (
    <div className="flex flex-col gap-2">
      <Button onClick={handleRun} disabled={loading} size="sm">
        {loading ? (
          <IconLoader2 className="size-4 animate-spin" />
        ) : (
          <IconPlayerPlay className="size-4" />
        )}
        Run
      </Button>
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  )
}
