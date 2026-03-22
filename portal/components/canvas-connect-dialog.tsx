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
} from "@tabler/icons-react"

type Step =
  | "url-input"
  | "connecting"
  | "install"
  | "enable-incognito"
  | "incognito-waiting"
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
  const [sessionId, setSessionId] = useState<string | null>(null)
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
      setSessionId(null)
    }
  }

  // Poll for incognito login session status
  const pollForLoginStatus = useCallback(async () => {
    if (!sessionId) return
    try {
      const resp = await sendToExtension({ type: "login-status", sessionId })
      if (!resp?.ok) return

      if (resp.status === "complete") {
        stopPolling()
        setStep("success")
      } else if (resp.status === "cancelled") {
        stopPolling()
        setError("Window was closed before login completed")
        setStep("url-input")
      } else if (resp.status === "error") {
        stopPolling()
        setError((resp.error as string) || "Login failed")
        setStep("url-input")
      }
    } catch {
      // ignore transient errors
    }
  }, [sessionId])

  useEffect(() => {
    if (step !== "incognito-waiting" || !sessionId) return
    intervalRef.current = setInterval(pollForLoginStatus, 2000)
    return () => stopPolling()
  }, [step, sessionId, pollForLoginStatus])

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

    // Check incognito access
    const incognitoResp = await sendToExtension({ type: "check-incognito" })
    if (!incognitoResp?.allowed) {
      setStep("enable-incognito")
      return
    }

    // Start incognito login
    await startLogin(url)
  }

  async function startLogin(url: string) {
    setError(null)
    const loginUrl = url.replace(/\/+$/, "") + "/login"
    const resp = await sendToExtension({
      type: "login",
      provider: "canvas",
      loginUrl,
      canvasUrl: url.replace(/\/+$/, ""),
      label: label.trim() || "Canvas LMS",
    })
    if (!resp?.ok || !resp?.sessionId) {
      setError((resp?.error as string) || "Failed to open incognito window")
      setStep("url-input")
      return
    }
    setSessionId(resp.sessionId as string)
    setStep("incognito-waiting")
  }

  async function handleIncognitoAccessDone() {
    setError(null)
    const resp = await sendToExtension({ type: "check-incognito" })
    if (resp?.allowed) {
      await startLogin(canvasUrl.trim())
    } else {
      setError("Incognito access is still not enabled. Please follow the steps above.")
    }
  }

  async function handleCancelLogin() {
    if (sessionId) {
      await sendToExtension({ type: "cancel-login", sessionId })
    }
    setOpen(false)
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
                Enter your school&apos;s Canvas URL. An incognito window will open for you to log in.
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

        {step === "enable-incognito" && (
          <>
            <DialogHeader>
              <DialogTitle>Enable incognito access</DialogTitle>
              <DialogDescription>
                To log into Canvas, enable incognito access for the extension:
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col gap-3 py-2">
              <ol className="flex flex-col gap-2 rounded-lg bg-muted p-4 text-sm">
                <li>1. Right-click the Emdash extension icon in the Chrome toolbar</li>
                <li>2. Click &quot;Manage extension&quot;</li>
                <li>3. Toggle &quot;Allow in Incognito&quot;</li>
              </ol>
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setStep("url-input")}>
                Back
              </Button>
              <Button onClick={handleIncognitoAccessDone}>Done</Button>
            </DialogFooter>
          </>
        )}

        {step === "incognito-waiting" && (
          <>
            <DialogHeader>
              <DialogTitle>Log in to Canvas</DialogTitle>
              <DialogDescription>
                An incognito window has opened. Log into Canvas and this dialog will close automatically.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col items-center gap-4 py-6">
              <IconLoader2 className="size-8 animate-spin text-muted-foreground" />
              <p className="text-center text-sm text-muted-foreground">
                Waiting for you to log in...
              </p>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={handleCancelLogin}>
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
