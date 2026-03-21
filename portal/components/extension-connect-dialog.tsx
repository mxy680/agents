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
  IconExternalLink,
  IconRefresh,
} from "@tabler/icons-react"

type Step =
  | "connect"
  | "install"
  | "choose-mode"
  | "label-input"
  | "enable-incognito"
  | "incognito-waiting"
  | "waiting"
  | "success"

interface ExtensionConnectDialogProps {
  provider: string
  providerName: string
  children: React.ReactNode
}

// Stable extension ID — derived from the `key` field in manifest.json.
// This never changes regardless of where the extension is loaded from.
const EXTENSION_ID = "pkkpglobhebcecahhomkiniapgdpfico"

// ---------------------------------------------------------------------------
// Direct messaging via chrome.runtime.sendMessage (externally_connectable)
// No content script needed — the page talks directly to the background worker.
// ---------------------------------------------------------------------------

function sendToExtension(payload: Record<string, unknown>): Promise<Record<string, unknown> | null> {
  return new Promise((resolve) => {
    const timeout = setTimeout(() => resolve(null), 2000)

    try {
      // chrome.runtime.sendMessage is available on pages matching externally_connectable
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

export function ExtensionConnectDialog({
  provider,
  providerName,
  children,
}: ExtensionConnectDialogProps) {
  const [open, setOpen] = useState(false)
  const [step, setStep] = useState<Step>("connect")
  const [error, setError] = useState<string | null>(null)
  const [incognitoAllowed, setIncognitoAllowed] = useState(false)
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
      setStep("connect")
      setError(null)
      setLabel("")
      setSessionId(null)
      setIncognitoAllowed(false)
    }
  }

  // When dialog opens, immediately try to connect
  useEffect(() => {
    if (!open || step !== "connect") return

    async function tryConnect() {
      // Ping extension directly — no content script needed
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
          const data = await tokenRes.json().catch(() => ({}))
          setError(data.error ?? "Failed to generate auth token")
          // Extension is installed but token failed — still show choose-mode
          // so user can retry or use sync-current-session
          const incognitoResp = await sendToExtension({ type: "check-incognito" })
          setIncognitoAllowed(incognitoResp?.allowed === true)
          setStep("choose-mode")
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
        }

        // Check incognito access and move to mode selection
        const incognitoResp = await sendToExtension({ type: "check-incognito" })
        setIncognitoAllowed(incognitoResp?.allowed === true)
        setStep("choose-mode")
      } catch {
        setError("Connection failed")
        setStep("choose-mode")
      }
    }

    tryConnect()
  }, [open, step, provider])

  // Poll for integration (used in "waiting" step — sync-current-session flow)
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
        setError("Window was closed")
        setStep("label-input")
      } else if (resp.status === "error") {
        stopPolling()
        setError((resp.error as string) || "Login failed")
        setStep("label-input")
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

  // Handlers
  function handleSyncCurrentSession() {
    sendToExtension({ type: "sync", provider })
    setStep("waiting")
  }

  function handleChooseIncognito() {
    if (incognitoAllowed) {
      setStep("label-input")
    } else {
      setStep("enable-incognito")
    }
  }

  async function handleStartLogin() {
    setError(null)
    const resp = await sendToExtension({
      type: "login",
      provider,
      label: label.trim() || `${providerName} Account`,
    })
    if (!resp?.ok) {
      setError((resp?.error as string) || "Failed to open incognito window")
      return
    }
    setSessionId(resp.sessionId as string)
    setStep("incognito-waiting")
  }

  async function handleIncognitoAccessDone() {
    setError(null)
    const resp = await sendToExtension({ type: "check-incognito" })
    if (resp?.allowed) {
      setIncognitoAllowed(true)
      setStep("label-input")
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
                <p className="mt-2">Then come back and click Connect again.</p>
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

        {step === "choose-mode" && (
          <>
            <DialogHeader>
              <DialogTitle>Connect {providerName}</DialogTitle>
              <DialogDescription>
                How would you like to connect your account?
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col gap-3 py-2">
              <button
                className="flex items-start gap-3 rounded-lg border bg-primary p-4 text-left text-primary-foreground transition-opacity hover:opacity-90"
                onClick={handleSyncCurrentSession}
              >
                <IconRefresh className="mt-0.5 size-5 shrink-0" />
                <div>
                  <p className="font-medium">Sync current session</p>
                  <p className="text-sm opacity-80">
                    Use the {providerName} account you&apos;re currently logged into
                  </p>
                </div>
              </button>

              <button
                className="flex items-start gap-3 rounded-lg border p-4 text-left transition-colors hover:bg-muted"
                onClick={handleChooseIncognito}
              >
                <IconExternalLink className="mt-0.5 size-5 shrink-0" />
                <div>
                  <p className="font-medium">Log in to another account</p>
                  <p className="text-sm text-muted-foreground">
                    Open a private window to log into a different account
                  </p>
                </div>
              </button>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
            </DialogFooter>
          </>
        )}

        {step === "label-input" && (
          <>
            <DialogHeader>
              <DialogTitle>Name this account</DialogTitle>
              <DialogDescription>
                Give this {providerName} account a label so you can tell them apart.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col gap-3 py-2">
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium" htmlFor="account-label">
                  Account name
                </label>
                <Input
                  id="account-label"
                  placeholder="e.g. Work, Personal"
                  value={label}
                  onChange={(e) => setLabel(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleStartLogin()}
                  autoFocus
                />
              </div>
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setStep("choose-mode")}>
                Back
              </Button>
              <Button onClick={handleStartLogin}>Start Login</Button>
            </DialogFooter>
          </>
        )}

        {step === "enable-incognito" && (
          <>
            <DialogHeader>
              <DialogTitle>Enable incognito access</DialogTitle>
              <DialogDescription>
                To log into a different account, enable incognito access for the extension:
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
              <Button variant="outline" onClick={() => setStep("choose-mode")}>
                Back
              </Button>
              <Button onClick={handleIncognitoAccessDone}>Done</Button>
            </DialogFooter>
          </>
        )}

        {step === "incognito-waiting" && (
          <>
            <DialogHeader>
              <DialogTitle>Log in to {providerName}</DialogTitle>
              <DialogDescription>
                An incognito window has opened. Log in and this dialog will close automatically.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col items-center gap-4 py-6">
              <IconLoader2 className="size-8 animate-spin text-muted-foreground" />
              <p className="text-center text-sm text-muted-foreground">
                Log in to {providerName} in the incognito window...
              </p>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={handleCancelLogin}>
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
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  sendToExtension({ type: "sync", provider })
                  pollForIntegration()
                }}
              >
                <IconRefresh className="size-4" />
                Sync Now
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
            <p className="text-lg font-medium">{providerName} connected!</p>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
