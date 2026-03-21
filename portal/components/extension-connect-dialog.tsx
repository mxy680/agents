"use client"

import { useState, useEffect, useRef, useCallback } from "react"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { IconCheck, IconLoader2, IconDownload, IconExternalLink } from "@tabler/icons-react"

type Step = "connect" | "install" | "waiting" | "success"

interface ExtensionConnectDialogProps {
  provider: string
  providerName: string
  children: React.ReactNode
}

// ---------------------------------------------------------------------------
// Extension communication via window.postMessage (crosses isolation boundary)
// ---------------------------------------------------------------------------

let messageCounter = 0

function sendToExtension(payload: Record<string, unknown>): Promise<Record<string, unknown> | null> {
  return new Promise((resolve) => {
    const id = ++messageCounter

    const timeout = setTimeout(() => {
      window.removeEventListener("message", handler)
      resolve(null)
    }, 2000)

    function handler(event: MessageEvent) {
      if (event.source !== window) return
      if (event.data?.direction !== "emdash-from-extension") return
      if (event.data?.id !== id) return

      clearTimeout(timeout)
      window.removeEventListener("message", handler)
      resolve(event.data.error ? null : event.data.response)
    }

    window.addEventListener("message", handler)
    window.postMessage({ direction: "emdash-to-extension", id, payload }, "*")
  })
}

function isExtensionInstalled(): boolean {
  return document.documentElement.hasAttribute("data-emdash-extension")
}

const PROVIDER_LOGIN_URLS: Record<string, string> = {
  instagram: "https://www.instagram.com/",
  linkedin: "https://www.linkedin.com/",
  x: "https://x.com/",
}

export function ExtensionConnectDialog({
  provider,
  providerName,
  children,
}: ExtensionConnectDialogProps) {
  const [open, setOpen] = useState(false)
  const [step, setStep] = useState<Step>("connect")
  const [error, setError] = useState<string | null>(null)
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
      setStep("connect")
      setError(null)
    }
  }

  // When dialog opens, immediately try to connect
  useEffect(() => {
    if (!open || step !== "connect") return

    async function tryConnect() {
      if (!isExtensionInstalled()) {
        setStep("install")
        return
      }

      // Verify extension responds via window.postMessage relay
      const ping = await sendToExtension({ type: "ping" })
      if (!ping?.ok) {
        setStep("install")
        return
      }

      // Generate token and configure extension automatically
      try {
        const tokenRes = await fetch("/api/integrations/extension/token", {
          method: "POST",
        })
        if (!tokenRes.ok) {
          setError("Failed to generate auth token")
          setStep("install")
          return
        }
        const { token } = await tokenRes.json()

        const configResp = await sendToExtension({
          type: "configure",
          token,
          portalUrl: window.location.origin,
        })

        if (!configResp?.ok) {
          setError("Failed to configure extension")
          setStep("install")
          return
        }

        // Trigger sync and move to waiting
        sendToExtension({ type: "sync", provider })
        setStep("waiting")
      } catch {
        setError("Connection failed")
        setStep("install")
      }
    }

    tryConnect()
  }, [open, step, provider])

  // Poll for integration
  const pollForIntegration = useCallback(async () => {
    try {
      const res = await fetch("/api/integrations")
      if (!res.ok) return
      const data = await res.json()
      const found = (data.integrations ?? []).some(
        (i: { provider: string; status: string }) =>
          i.provider === provider && i.status === "active"
      )
      if (found) {
        stopPolling()
        setStep("success")
      }
    } catch {
      // ignore
    }
  }, [provider])

  useEffect(() => {
    if (step !== "waiting") return
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

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        {step === "connect" && (
          <>
            <DialogHeader>
              <DialogTitle>Connecting {providerName}...</DialogTitle>
            </DialogHeader>
            <div className="flex items-center justify-center py-8">
              <IconLoader2 className="size-8 animate-spin text-muted-foreground" />
            </div>
          </>
        )}

        {step === "install" && (
          <>
            <DialogHeader>
              <DialogTitle>Connect {providerName}</DialogTitle>
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
                <p className="mt-2">The page will reload automatically after installation.</p>
              </div>

              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
            </DialogFooter>
          </>
        )}

        {step === "waiting" && (
          <>
            <DialogHeader>
              <DialogTitle>Log in to {providerName}</DialogTitle>
              <DialogDescription>
                {providerName} session will be synced automatically.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col items-center gap-4 py-6">
              <IconLoader2 className="size-8 animate-spin text-muted-foreground" />
              <p className="text-center text-sm text-muted-foreground">
                Already logged in? Click sync. Otherwise, log in and this closes automatically.
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => window.open(PROVIDER_LOGIN_URLS[provider], "_blank")}
                >
                  <IconExternalLink className="size-4" />
                  Open {providerName}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    sendToExtension({ type: "sync", provider })
                    pollForIntegration()
                  }}
                >
                  Sync Now
                </Button>
              </div>
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
            <p className="text-lg font-medium">{providerName} connected!</p>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
