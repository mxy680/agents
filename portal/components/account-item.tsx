"use client"

import { useRouter } from "next/navigation"
import { useState, useRef, useEffect } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { IconX, IconPlayerPlay, IconCheck, IconAlertTriangle, IconLoader2, IconRefresh } from "@tabler/icons-react"

type ConnectType = string

interface AccountItemProps {
  id: string
  label: string
  status: string
  provider: string
  connectType: ConnectType
}

export function AccountItem({ id, label, status, provider, connectType }: AccountItemProps) {
  const router = useRouter()
  const [removing, setRemoving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<"pass" | "fail" | null>(null)
  const [testError, setTestError] = useState("")
  const [refreshing, setRefreshing] = useState(false)
  const [refreshStatus, setRefreshStatus] = useState<"idle" | "launching" | "waiting" | "done" | "error">("idle")
  const [refreshMessage, setRefreshMessage] = useState("")
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    return () => {
      if (pollRef.current) clearInterval(pollRef.current)
    }
  }, [])

  async function handleDisconnect() {
    setRemoving(true)
    const res = await fetch("/api/integrations/disconnect", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id }),
    })
    if (res.ok) {
      router.refresh()
    } else {
      setRemoving(false)
    }
  }

  async function handleTest() {
    setTesting(true)
    setTestResult(null)
    setTestError("")
    try {
      const res = await fetch("/api/integrations/test", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id }),
      })
      const data = await res.json()
      if (data.ok) {
        setTestResult("pass")
      } else {
        setTestResult("fail")
        setTestError(data.error || "Test failed")
      }
    } catch {
      setTestResult("fail")
      setTestError("Network error")
    } finally {
      setTesting(false)
      setTimeout(() => {
        setTestResult(null)
        setTestError("")
      }, 8000)
    }
  }

  async function handleRefresh() {
    setRefreshing(true)
    setRefreshStatus("launching")
    setRefreshMessage("")

    try {
      if (connectType === "oauth") {
        // OAuth: redirect to re-auth with same label
        window.location.href = `/api/integrations/${provider}/connect?label=${encodeURIComponent(label)}`
        return
      }

      if (connectType === "cookie-capture") {
        // Cookie capture: POST to refresh endpoint
        setRefreshMessage("Browser opening — solve CAPTCHA if prompted...")
        const controller = new AbortController()
        const timeout = setTimeout(() => controller.abort(), 180_000)
        const res = await fetch(`/api/integrations/${provider}/refresh`, {
          method: "POST",
          signal: controller.signal,
        })
        clearTimeout(timeout)
        const data = await res.json()
        if (data.ok) {
          setRefreshStatus("done")
          setRefreshMessage(`Captured ${data.cookieCount} cookies`)
          setTimeout(() => {
            setRefreshing(false)
            setRefreshStatus("idle")
            setRefreshMessage("")
            router.refresh()
          }, 1500)
        } else {
          setRefreshStatus("error")
          setRefreshMessage(data.error || "Failed to refresh")
        }
        return
      }

      if (connectType === "playwright") {
        // Playwright: launch session with same label, poll for completion
        setRefreshMessage(`Starting ${provider} browser session...`)
        const res = await fetch("/api/integrations/playwright/connect", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ provider, label }),
        })

        if (!res.ok) {
          const err = await res.json()
          setRefreshStatus("error")
          setRefreshMessage(err.error || "Failed to start session")
          return
        }

        const { sessionId } = await res.json()
        setRefreshStatus("waiting")
        setRefreshMessage("Browser opened — log in to refresh your session.")

        pollRef.current = setInterval(async () => {
          try {
            const statusRes = await fetch(
              `/api/integrations/playwright/status?sessionId=${sessionId}`
            )
            if (!statusRes.ok) return
            const data = await statusRes.json()
            setRefreshMessage(data.message)

            if (data.status === "saved") {
              setRefreshStatus("done")
              setRefreshMessage(data.message)
              if (pollRef.current) clearInterval(pollRef.current)
              setTimeout(() => {
                setRefreshing(false)
                setRefreshStatus("idle")
                setRefreshMessage("")
                router.refresh()
              }, 1500)
            } else if (data.status === "error") {
              setRefreshStatus("error")
              if (pollRef.current) clearInterval(pollRef.current)
            }
          } catch {
            // Ignore polling errors
          }
        }, 1500)
        return
      }

      // Framer / BlueBubbles: manual config — just disconnect and re-connect
      setRefreshStatus("error")
      setRefreshMessage("Disconnect and re-connect to update credentials.")
    } catch {
      setRefreshStatus("error")
      setRefreshMessage("Network error")
    } finally {
      if (connectType !== "playwright" && connectType !== "oauth") {
        // For non-polling flows, reset refreshing after completion
        // (playwright resets in the poll callback, oauth redirects away)
        if (refreshStatus !== "done") {
          setRefreshing(false)
        }
      }
    }
  }

  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center justify-between gap-2 border p-2 text-sm">
        <span className="truncate font-medium">{label}</span>
        <div className="flex items-center gap-1">
          <Badge variant={status === "active" ? "default" : "secondary"}>
            {status}
          </Badge>
          <Button
            variant="ghost"
            size="icon"
            className="size-6"
            aria-label="Test credentials"
            onClick={handleTest}
            disabled={testing || removing}
          >
            {testing ? (
              <IconLoader2 className="size-3 animate-spin" />
            ) : testResult === "pass" ? (
              <IconCheck className="size-3 text-green-500" />
            ) : testResult === "fail" ? (
              <IconAlertTriangle className="size-3 text-red-500" />
            ) : (
              <IconPlayerPlay className="size-3" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="size-6"
            aria-label="Refresh credentials"
            onClick={handleRefresh}
            disabled={refreshing || removing || testing}
          >
            {refreshing && refreshStatus !== "done" && refreshStatus !== "error" ? (
              <IconLoader2 className="size-3 animate-spin" />
            ) : refreshStatus === "done" ? (
              <IconCheck className="size-3 text-green-500" />
            ) : (
              <IconRefresh className="size-3" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="size-6"
            aria-label="Disconnect account"
            onClick={handleDisconnect}
            disabled={removing}
          >
            <IconX className="size-3" />
          </Button>
        </div>
      </div>
      {testError && (
        <pre className="px-2 py-1 text-xs text-red-500 font-mono whitespace-pre-wrap break-words max-h-24 overflow-y-auto border border-red-500/20 bg-red-500/5">
          {testError}
        </pre>
      )}
      {refreshMessage && (
        <div className={`px-2 py-1 text-xs ${refreshStatus === "error" ? "text-red-500" : refreshStatus === "done" ? "text-green-500" : "text-muted-foreground"}`}>
          {refreshMessage}
          {refreshStatus === "error" && (
            <Button
              variant="ghost"
              size="sm"
              className="ml-2 h-auto px-1 py-0 text-xs"
              onClick={() => { setRefreshing(false); setRefreshStatus("idle"); setRefreshMessage("") }}
            >
              Dismiss
            </Button>
          )}
        </div>
      )}
    </div>
  )
}
