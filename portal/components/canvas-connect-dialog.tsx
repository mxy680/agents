"use client"

import { useState, useEffect, useRef, useCallback } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import {
  IconCheck,
  IconLoader2,
  IconDownload,
  IconRefresh,
} from "@tabler/icons-react"

type Step =
  | "url-input"
  | "connecting"
  | "install"
  | "syncing"
  | "success"

interface CanvasConnectDialogProps {
  children: React.ReactNode
}

const EXTENSION_ID = "pkkpglobhebcecahhomkiniapgdpfico"

function sendToExtension(payload: Record<string, unknown>): Promise<Record<string, unknown> | null> {
  return new Promise((resolve) => {
    const timeout = setTimeout(() => resolve(null), 10000)

    try {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const chrome = (window as any).chrome
      if (!chrome?.runtime?.sendMessage) {
        clearTimeout(timeout)
        resolve(null)
        return
      }

      chrome.runtime.sendMessage(EXTENSION_ID, payload, (response: Record<string, unknown> | undefined) => {
        clearTimeout(timeout)
        if (chrome.runtime.lastError) {
          resolve(null)
        } else {
          resolve(response ?? null)
        }
      })
    } catch {
      clearTimeout(timeout)
      resolve(null)
    }
  })
}

export function CanvasConnectDialog({ children }: CanvasConnectDialogProps) {
  const [open, setOpen] = useState(false)
  const [step, setStep] = useState<Step>("url-input")
  const [error, setError] = useState<string | null>(null)
  const [canvasUrl, setCanvasUrl] = useState("")
  const [label, setLabel] = useState("")
  const [initialCount, setInitialCount] = useState(0)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  function stopPolling() {
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
  }

  function handleOpenChange(v: boolean) {
    setOpen(v)
    if (!v) {
      stopPolling()
      setStep("url-input")
      setError(null)
      setCanvasUrl("")
      setLabel("")
    }
  }

  // Poll for integration
  const pollForIntegration = useCallback(async () => {
    try {
      const res = await fetch("/api/integrations")
      if (!res.ok) return
      const data = await res.json()
      const count = (data.integrations ?? []).filter(
        (i: { provider: string; status: string }) =>
          i.provider === "canvas" && i.status === "active"
      ).length
      if (count > initialCount) {
        stopPolling()
        setStep("success")
      }
    } catch {
      // ignore
    }
  }, [initialCount])

  useEffect(() => {
    if (step !== "syncing") return
    pollForIntegration()
    intervalRef.current = setInterval(pollForIntegration, 3000)
    return () => stopPolling()
  }, [step, pollForIntegration])

  // Auto-close on success
  useEffect(() => {
    if (step !== "success") return
    const timer = setTimeout(() => {
      setOpen(false)
      window.location.reload()
    }, 1500)
    return () => clearTimeout(timer)
  }, [step])

  async function handleConnect() {
    setError(null)

    // Validate URL
    const url = canvasUrl.trim()
    if (!url) {
      setError("Canvas URL is required")
      return
    }
    try {
      new URL(url)
    } catch {
      setError("Invalid URL. Include https:// (e.g. https://canvas.case.edu)")
      return
    }

    setStep("connecting")

    // Snapshot current integration count
    try {
      const res = await fetch("/api/integrations")
      if (res.ok) {
        const data = await res.json()
        const count = (data.integrations ?? []).filter(
          (i: { provider: string; status: string }) =>
            i.provider === "canvas" && i.status === "active"
        ).length
        setInitialCount(count)
      }
    } catch {}

    // Ping extension
    const ping = await sendToExtension({ type: "ping" })
    if (!ping?.ok) {
      setStep("install")
      return
    }

    // Generate token and configure extension
    try {
      const tokenRes = await fetch("/api/integrations/extension/token", {
        method: "POST",
      })
      if (!tokenRes.ok) {
        const data = await tokenRes.json().catch(() => ({}))
        setError(data.error ?? "Failed to generate auth token")
        setStep("url-input")
        return
      }
      const { token } = await tokenRes.json()

      await sendToExtension({
        type: "configure",
        token,
        portalUrl: window.location.origin,
      })
    } catch {
      setError("Failed to configure extension")
      setStep("url-input")
      return
    }

    // Ask extension to sync Canvas cookies from the provided URL
    const syncResp = await sendToExtension({
      type: "sync-canvas",
      canvasUrl: url.replace(/\/+$/, ""),
      label: label.trim() || "Canvas LMS",
    })

    if (!syncResp?.ok) {
      setError((syncResp?.error as string) || "Failed to sync Canvas session. Make sure you are logged into Canvas in this browser.")
      setStep("url-input")
      return
    }

    setStep("syncing")
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        {step === "url-input" && (
          <>
            <DialogHeader>
              <DialogTitle>Connect Canvas LMS</DialogTitle>
              <DialogDescription>
                Enter your school&apos;s Canvas URL. Make sure you&apos;re logged into Canvas in this browser first.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col gap-3 py-2">
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium" htmlFor="canvas-url">
                  Canvas URL
                </label>
                <Input
                  id="canvas-url"
                  placeholder="https://canvas.case.edu"
                  value={canvasUrl}
                  onChange={(e) => setCanvasUrl(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleConnect()}
                  autoFocus
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium" htmlFor="canvas-label">
                  Account name (optional)
                </label>
                <Input
                  id="canvas-label"
                  placeholder="e.g. CWRU Canvas"
                  value={label}
                  onChange={(e) => setLabel(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleConnect()}
                />
              </div>
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleConnect} disabled={!canvasUrl.trim()}>
                Connect
              </Button>
            </DialogFooter>
          </>
        )}

        {step === "connecting" && (
          <>
            <DialogHeader>
              <DialogTitle>Connecting Canvas...</DialogTitle>
            </DialogHeader>
            <div className="flex items-center justify-center py-8">
              <IconLoader2 className="size-8 animate-spin text-muted-foreground" />
            </div>
          </>
        )}

        {step === "install" && (
          <>
            <DialogHeader>
              <DialogTitle>Connect Canvas LMS</DialogTitle>
              <DialogDescription>
                Install the Emdash extension to connect your account. This is a one-time step.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col gap-3 py-2">
              <Button
                className="w-full gap-2"
                onClick={() => window.open("/api/extension/download", "_blank")}
              >
                <IconDownload className="size-4" />
                Download Extension
              </Button>

              <div className="rounded-lg bg-muted p-3 text-sm text-muted-foreground">
                <p className="mb-2 font-medium text-foreground">After downloading:</p>
                <ol className="flex flex-col gap-1">
                  <li>1. Unzip the file</li>
                  <li>2. Go to <code className="rounded bg-background px-1 py-0.5 text-xs">chrome://extensions</code></li>
                  <li>3. Turn on &quot;Developer mode&quot; (top right)</li>
                  <li>4. Click &quot;Load unpacked&quot; and select the unzipped folder</li>
                </ol>
                <p className="mt-2">Then come back and click Connect again.</p>
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
            </DialogFooter>
          </>
        )}

        {step === "syncing" && (
          <>
            <DialogHeader>
              <DialogTitle>Syncing Canvas session...</DialogTitle>
              <DialogDescription>
                Verifying your Canvas credentials.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col items-center gap-4 py-6">
              <IconLoader2 className="size-8 animate-spin text-muted-foreground" />
              <p className="text-center text-sm text-muted-foreground">
                Waiting for confirmation...
              </p>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  sendToExtension({
                    type: "sync-canvas",
                    canvasUrl: canvasUrl.trim().replace(/\/+$/, ""),
                    label: label.trim() || "Canvas LMS",
                  })
                  pollForIntegration()
                }}
              >
                <IconRefresh className="size-4" />
                Retry Sync
              </Button>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
            </DialogFooter>
          </>
        )}

        {step === "success" && (
          <div className="flex flex-col items-center gap-4 py-8">
            <div className="flex size-16 items-center justify-center rounded-full bg-green-100 dark:bg-green-950">
              <IconCheck className="size-8 text-green-600 dark:text-green-400" />
            </div>
            <p className="text-lg font-medium">Canvas LMS connected!</p>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
