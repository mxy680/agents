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
import { IconCheck, IconLoader2, IconPuzzle, IconExternalLink, IconDownload } from "@tabler/icons-react"

type Step = "detecting" | "not-installed" | "connecting" | "waiting" | "success"

interface ExtensionConnectDialogProps {
  provider: string
  providerName: string
  children: React.ReactNode
}

// ---------------------------------------------------------------------------
// Extension communication via content script relay
// ---------------------------------------------------------------------------

let messageCounter = 0

/**
 * Send a message to the extension's background service worker via the
 * content script relay. The content script listens for "emdash-to-extension"
 * CustomEvents on document and forwards them to chrome.runtime.sendMessage,
 * then dispatches "emdash-from-extension" CustomEvents with the response.
 */
function sendToExtension(payload: Record<string, unknown>): Promise<Record<string, unknown> | null> {
  return new Promise((resolve) => {
    const id = ++messageCounter

    const timeout = setTimeout(() => {
      document.removeEventListener("emdash-from-extension", handler)
      resolve(null)
    }, 3000)

    function handler(event: Event) {
      const detail = (event as CustomEvent).detail
      if (detail?.id !== id) return

      clearTimeout(timeout)
      document.removeEventListener("emdash-from-extension", handler)

      if (detail.error) {
        resolve(null)
      } else {
        resolve(detail.response)
      }
    }

    document.addEventListener("emdash-from-extension", handler)
    document.dispatchEvent(
      new CustomEvent("emdash-to-extension", { detail: { id, payload } })
    )
  })
}

/**
 * Check if the extension is installed by looking for the data attribute
 * the content script sets on <html>.
 */
function isExtensionInstalled(): boolean {
  return document.documentElement.hasAttribute("data-emdash-extension")
}

const PROVIDER_LOGIN_URLS: Record<string, string> = {
  instagram: "https://www.instagram.com/",
  linkedin: "https://www.linkedin.com/",
  x: "https://x.com/",
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function ExtensionConnectDialog({
  provider,
  providerName,
  children,
}: ExtensionConnectDialogProps) {
  const [open, setOpen] = useState(false)
  const [step, setStep] = useState<Step>("detecting")
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
      setStep("detecting")
      setError(null)
    }
  }

  // Step 1: Detect extension on dialog open
  useEffect(() => {
    if (!open || step !== "detecting") return

    async function detect() {
      // Check if the content script injected the extension ID
      if (!isExtensionInstalled()) {
        setStep("not-installed")
        return
      }

      // Verify the extension is responsive
      const resp = await sendToExtension({ type: "ping" })
      if (resp?.ok) {
        setStep("connecting")
      } else {
        setStep("not-installed")
      }
    }

    // Small delay to let the content script inject
    const timer = setTimeout(detect, 200)
    return () => clearTimeout(timer)
  }, [open, step])

  // Step 2: Auto-configure extension and trigger sync
  useEffect(() => {
    if (step !== "connecting") return

    async function configure() {
      try {
        // Generate a token from the portal API
        const tokenRes = await fetch("/api/integrations/extension/token", {
          method: "POST",
        })
        if (!tokenRes.ok) {
          setError("Failed to generate auth token")
          return
        }
        const { token } = await tokenRes.json()

        // Send config to extension — no copy/paste needed!
        const configResp = await sendToExtension({
          type: "configure",
          token,
          portalUrl: window.location.origin,
        })

        if (!configResp?.ok) {
          setError("Failed to configure extension")
          return
        }

        // Trigger a sync for this provider
        sendToExtension({ type: "sync", provider })

        setStep("waiting")
      } catch {
        setError("Failed to connect to extension")
      }
    }

    configure()
  }, [step, provider])

  // Step 3: Poll for integration to appear
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
      // Ignore polling errors
    }
  }, [provider])

  useEffect(() => {
    if (step !== "waiting") return

    pollForIntegration()
    intervalRef.current = setInterval(pollForIntegration, 3000)

    return () => stopPolling()
  }, [step, pollForIntegration])

  // Step 4: Auto-close on success
  useEffect(() => {
    if (step !== "success") return

    const timer = setTimeout(() => {
      setOpen(false)
      window.location.reload()
    }, 2000)

    return () => clearTimeout(timer)
  }, [step])

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        {step === "detecting" && (
          <>
            <DialogHeader>
              <DialogTitle>Connect {providerName}</DialogTitle>
              <DialogDescription>Detecting extension...</DialogDescription>
            </DialogHeader>
            <div className="flex flex-col items-center gap-4 py-6">
              <IconLoader2 className="size-8 animate-spin text-muted-foreground" />
            </div>
          </>
        )}

        {step === "not-installed" && (
          <>
            <DialogHeader>
              <DialogTitle>Install Extension</DialogTitle>
              <DialogDescription>
                The Emdash Chrome extension is needed to connect your {providerName} account.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col gap-4 py-2">
              <div className="flex items-start gap-3 rounded-lg border p-4">
                <IconPuzzle className="mt-0.5 size-5 shrink-0 text-muted-foreground" />
                <div className="flex flex-col gap-2 text-sm">
                  <p className="font-medium">One-time setup (30 seconds)</p>
                  <ol className="flex flex-col gap-1.5 text-muted-foreground">
                    <li>1. Download and unzip the extension (button below)</li>
                    <li>2. Open <code className="rounded bg-muted px-1 py-0.5 text-xs">chrome://extensions</code></li>
                    <li>3. Enable &quot;Developer mode&quot; (top right)</li>
                    <li>4. Click &quot;Load unpacked&quot; → select the unzipped folder</li>
                    <li>5. Come back here and click &quot;Try Again&quot;</li>
                  </ol>
                </div>
              </div>

              <Button
                variant="outline"
                className="w-full gap-2"
                onClick={() => window.open("/api/extension/download", "_blank")}
              >
                <IconDownload className="size-4" />
                Download Extension
              </Button>

              <div className="rounded-lg border border-blue-200 bg-blue-50 p-3 text-sm text-blue-800 dark:border-blue-900 dark:bg-blue-950 dark:text-blue-300">
                Once installed, all future connections are automatic — just click Connect and log in.
              </div>
            </div>

            {error && <p className="text-sm text-destructive">{error}</p>}

            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button onClick={() => window.location.reload()}>Try Again</Button>
            </DialogFooter>
          </>
        )}

        {step === "connecting" && (
          <>
            <DialogHeader>
              <DialogTitle>Connecting {providerName}...</DialogTitle>
              <DialogDescription>
                Setting up automatically.
              </DialogDescription>
            </DialogHeader>
            <div className="flex flex-col items-center gap-4 py-6">
              <IconLoader2 className="size-8 animate-spin text-muted-foreground" />
            </div>
            {error && (
              <DialogFooter className="flex-col gap-2">
                <p className="text-sm text-destructive">{error}</p>
                <Button variant="outline" onClick={() => setOpen(false)}>
                  Cancel
                </Button>
              </DialogFooter>
            )}
          </>
        )}

        {step === "waiting" && (
          <>
            <DialogHeader>
              <DialogTitle>Log in to {providerName}</DialogTitle>
              <DialogDescription>
                The extension is ready. Now just log in to sync your session.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col items-center gap-4 py-6">
              <div className="flex size-16 items-center justify-center rounded-full bg-muted">
                <IconLoader2 className="size-8 animate-spin text-muted-foreground" />
              </div>
              <p className="text-center text-sm text-muted-foreground">
                If you&apos;re already logged in to {providerName}, click Sync below.
                Otherwise, log in and this dialog closes automatically.
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
          <>
            <DialogHeader>
              <DialogTitle>Connected!</DialogTitle>
              <DialogDescription>
                Your {providerName} account has been connected successfully.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col items-center gap-4 py-6">
              <div className="flex size-16 items-center justify-center rounded-full bg-green-100 dark:bg-green-950">
                <IconCheck className="size-8 text-green-600 dark:text-green-400" />
              </div>
              <p className="text-center text-sm text-muted-foreground">
                Closing automatically...
              </p>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
