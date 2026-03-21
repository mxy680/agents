"use client"

import { useState, useEffect, useRef } from "react"
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
import { Input } from "@/components/ui/input"
import { IconCheck, IconCopy, IconLoader2 } from "@tabler/icons-react"

type Step = "setup" | "waiting" | "success"

interface ExtensionConnectDialogProps {
  provider: string
  providerName: string
  children: React.ReactNode
}

export function ExtensionConnectDialog({
  provider,
  providerName,
  children,
}: ExtensionConnectDialogProps) {
  const [open, setOpen] = useState(false)
  const [step, setStep] = useState<Step>("setup")
  const [token, setToken] = useState<string | null>(null)
  const [tokenError, setTokenError] = useState<string | null>(null)
  const [isGenerating, setIsGenerating] = useState(false)
  const [copied, setCopied] = useState(false)
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
      setStep("setup")
      setToken(null)
      setTokenError(null)
      setIsGenerating(false)
      setCopied(false)
    }
  }

  async function handleGenerateToken() {
    setIsGenerating(true)
    setTokenError(null)
    try {
      const res = await fetch("/api/integrations/extension/token", {
        method: "POST",
      })
      if (!res.ok) {
        const data = await res.json().catch(() => ({}))
        setTokenError(data.error ?? "Failed to generate token")
        return
      }
      const data = await res.json()
      setToken(data.token)
    } catch {
      setTokenError("Network error. Please try again.")
    } finally {
      setIsGenerating(false)
    }
  }

  async function handleCopyToken() {
    if (!token) return
    await navigator.clipboard.writeText(token)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  function handleContinue() {
    setStep("waiting")
  }

  useEffect(() => {
    if (step !== "waiting") return

    async function poll() {
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
    }

    poll()
    intervalRef.current = setInterval(poll, 3000)

    return () => stopPolling()
  }, [step, provider])

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
        {step === "setup" && (
          <>
            <DialogHeader>
              <DialogTitle>Connect {providerName}</DialogTitle>
              <DialogDescription>
                Use the Emdash Chrome extension to sync your {providerName} session
                automatically.
              </DialogDescription>
            </DialogHeader>

            <ol className="flex flex-col gap-4 text-sm">
              <li className="flex flex-col gap-1">
                <span className="font-medium">1. Install the Emdash Chrome extension</span>
                <span className="text-muted-foreground">
                  Load it unpacked from the{" "}
                  <code className="rounded bg-muted px-1 py-0.5 text-xs">
                    portal/extension/
                  </code>{" "}
                  directory in Chrome&apos;s extension settings.
                </span>
              </li>

              <li className="flex flex-col gap-2">
                <span className="font-medium">2. Generate an auth token</span>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleGenerateToken}
                    disabled={isGenerating}
                  >
                    {isGenerating ? (
                      <>
                        <IconLoader2 className="size-4 animate-spin" />
                        Generating...
                      </>
                    ) : (
                      "Generate token"
                    )}
                  </Button>
                </div>
                {token && (
                  <div className="flex gap-2">
                    <Input
                      readOnly
                      value={token}
                      className="font-mono text-xs"
                      onClick={(e) => (e.target as HTMLInputElement).select()}
                    />
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={handleCopyToken}
                      className="shrink-0"
                    >
                      {copied ? (
                        <IconCheck className="size-4" />
                      ) : (
                        <IconCopy className="size-4" />
                      )}
                    </Button>
                  </div>
                )}
                {tokenError && (
                  <p className="text-sm text-destructive">{tokenError}</p>
                )}
              </li>

              <li className="flex flex-col gap-1">
                <span className="font-medium">3. Paste the token into the extension popup</span>
                <span className="text-muted-foreground">
                  Open the Emdash extension popup and paste the token to link it to your
                  account.
                </span>
              </li>

              <li className="flex flex-col gap-1">
                <span className="font-medium">
                  4. Log in to {providerName} in your browser
                </span>
                <span className="text-muted-foreground">
                  The extension will detect your login and sync the session automatically.
                </span>
              </li>
            </ol>

            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleContinue}>Continue</Button>
            </DialogFooter>
          </>
        )}

        {step === "waiting" && (
          <>
            <DialogHeader>
              <DialogTitle>Waiting for {providerName} session...</DialogTitle>
              <DialogDescription>
                The extension will automatically detect your login.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col items-center gap-4 py-6">
              <div className="flex size-16 items-center justify-center rounded-full bg-muted">
                <IconLoader2 className="size-8 animate-spin text-muted-foreground" />
              </div>
              <p className="text-center text-sm text-muted-foreground">
                Log in to {providerName} in your browser tab. Once detected, this dialog
                will close automatically.
              </p>
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
