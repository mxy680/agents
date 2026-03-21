"use client"

import { useState, useTransition } from "react"
import { useRouter } from "next/navigation"
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
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field"
import { XBrowser } from "@/components/x-browser"

type Step = "label" | "browser"

interface XConnectDialogProps {
  children: React.ReactNode
}

export function XConnectDialog({ children }: XConnectDialogProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [label, setLabel] = useState("")
  const [step, setStep] = useState<Step>("label")
  const [wsUrl, setWsUrl] = useState<string | null>(null)
  const [wsToken, setWsToken] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isPending, startTransition] = useTransition()

  function handleOpenChange(v: boolean) {
    setOpen(v)
    if (!v) {
      // Reset state when dialog closes
      setLabel("")
      setStep("label")
      setWsUrl(null)
      setWsToken(null)
      setError(null)
    }
  }

  async function handleConnect() {
    const accountLabel = label.trim() || "X Account"
    setError(null)

    startTransition(async () => {
      try {
        const res = await fetch("/api/integrations/x/browser-session", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ label: accountLabel }),
        })

        if (!res.ok) {
          const data = await res.json().catch(() => ({}))
          setError(data.error ?? "Failed to start browser session")
          return
        }

        const { wsUrl: url, token } = await res.json()
        setWsUrl(url)
        setWsToken(token)
        setStep("browser")
      } catch {
        setError("Network error. Please try again.")
      }
    })
  }

  function handleComplete() {
    setOpen(false)
    startTransition(() => {
      router.refresh()
    })
  }

  function handleCancel() {
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent
        className={step === "browser" ? "!max-w-[90vw] w-[90vw] h-[85vh] overflow-hidden flex flex-col" : undefined}
      >
        {step === "label" && (
          <>
            <DialogHeader>
              <DialogTitle>Connect X</DialogTitle>
              <DialogDescription>
                Give this account a name so you can tell it apart from other X accounts.
              </DialogDescription>
            </DialogHeader>
            <FieldGroup>
              <Field>
                <FieldLabel htmlFor="x-account-label">Account name</FieldLabel>
                <Input
                  id="x-account-label"
                  placeholder="e.g. Work X, Personal"
                  value={label}
                  onChange={(e) => setLabel(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      e.preventDefault()
                      handleConnect()
                    }
                  }}
                  autoFocus
                />
              </Field>
            </FieldGroup>
            {error && <p className="text-sm text-destructive">{error}</p>}
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleConnect} disabled={isPending}>
                {isPending ? "Starting..." : "Connect"}
              </Button>
            </DialogFooter>
          </>
        )}

        {step === "browser" && wsUrl && wsToken && (
          <>
            <DialogHeader>
              <DialogTitle>Log in to X</DialogTitle>
              <DialogDescription>
                Log in to your X account in the browser below. The session will be
                saved automatically once you are logged in.
              </DialogDescription>
            </DialogHeader>
            <XBrowser
              wsUrl={wsUrl}
              token={wsToken}
              onComplete={handleComplete}
              onCancel={handleCancel}
            />
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
