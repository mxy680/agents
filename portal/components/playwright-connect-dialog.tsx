"use client"

import * as React from "react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { IconLoader2, IconCheck, IconAlertCircle, IconBrowser } from "@tabler/icons-react"

interface PlaywrightConnectDialogProps {
  provider: string
  providerName: string
  children: React.ReactNode
}

type Status = "idle" | "launching" | "waiting_login" | "capturing" | "done" | "error"

export function PlaywrightConnectDialog({
  provider,
  providerName,
  children,
}: PlaywrightConnectDialogProps) {
  const [open, setOpen] = React.useState(false)
  const [status, setStatus] = React.useState<Status>("idle")
  const [message, setMessage] = React.useState("")
  const [sessionId, setSessionId] = React.useState<string | null>(null)

  const pollRef = React.useRef<ReturnType<typeof setInterval> | null>(null)

  React.useEffect(() => {
    return () => {
      if (pollRef.current) clearInterval(pollRef.current)
    }
  }, [])

  async function handleLaunch() {
    setStatus("launching")
    setMessage(`Starting ${providerName} browser session...`)

    try {
      const res = await fetch("/api/integrations/playwright/connect", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ provider }),
      })

      if (!res.ok) {
        const err = await res.json()
        setStatus("error")
        setMessage(err.error || "Failed to start session")
        return
      }

      const { sessionId: sid } = await res.json()
      setSessionId(sid)
      setStatus("waiting_login")
      setMessage(`A browser window has opened. Log in to ${providerName} there.`)

      // Start polling for status
      pollRef.current = setInterval(async () => {
        try {
          const statusRes = await fetch(
            `/api/integrations/playwright/status?sessionId=${sid}`
          )
          if (!statusRes.ok) return

          const data = await statusRes.json()
          setMessage(data.message)

          if (data.status === "done") {
            setStatus("done")
            if (pollRef.current) clearInterval(pollRef.current)
            // Auto-close after 2 seconds
            setTimeout(() => {
              setOpen(false)
              setStatus("idle")
              window.location.reload()
            }, 2000)
          } else if (data.status === "error") {
            setStatus("error")
            if (pollRef.current) clearInterval(pollRef.current)
          } else if (data.status === "capturing") {
            setStatus("capturing")
          }
        } catch {
          // Ignore polling errors
        }
      }, 1500)
    } catch {
      setStatus("error")
      setMessage("Failed to connect to server")
    }
  }

  function handleOpenChange(next: boolean) {
    if (!next && pollRef.current) {
      clearInterval(pollRef.current)
    }
    if (!next) {
      setStatus("idle")
      setMessage("")
      setSessionId(null)
    }
    setOpen(next)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Connect {providerName}</DialogTitle>
          <DialogDescription>
            A browser window will open for you to log in. Your session cookies
            will be captured automatically.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col items-center gap-4 py-4">
          {status === "idle" && (
            <Button onClick={handleLaunch} className="w-full">
              <IconBrowser className="mr-2 size-4" />
              Launch Browser
            </Button>
          )}

          {status === "launching" && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <IconLoader2 className="size-4 animate-spin" />
              {message}
            </div>
          )}

          {status === "waiting_login" && (
            <div className="flex flex-col items-center gap-2">
              <div className="flex items-center gap-2 text-sm text-yellow-500">
                <IconBrowser className="size-4" />
                Browser opened
              </div>
              <p className="text-center text-sm text-muted-foreground">
                {message}
              </p>
            </div>
          )}

          {status === "capturing" && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <IconLoader2 className="size-4 animate-spin" />
              {message}
            </div>
          )}

          {status === "done" && (
            <div className="flex items-center gap-2 text-sm text-green-500">
              <IconCheck className="size-4" />
              {message}
            </div>
          )}

          {status === "error" && (
            <div className="flex flex-col items-center gap-2">
              <div className="flex items-center gap-2 text-sm text-destructive">
                <IconAlertCircle className="size-4" />
                {message}
              </div>
              <Button variant="outline" size="sm" onClick={handleLaunch}>
                Try again
              </Button>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
