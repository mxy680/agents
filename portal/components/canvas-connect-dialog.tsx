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
import { IconCheck, IconLoader2 } from "@tabler/icons-react"

type Step = "setup" | "waiting" | "success"

interface CanvasConnectDialogProps {
  children: React.ReactNode
}

/**
 * Builds the bookmarklet JavaScript that:
 * 1. Reads document.cookie on the Canvas page
 * 2. Opens a popup to our portal callback with cookies in the URL fragment
 *
 * The fragment (#) is used instead of query params so cookies never appear
 * in server logs or browser history.
 */
function buildBookmarkletHref(portalOrigin: string) {
  // Minified bookmarklet code
  const code = `
    (function(){
      var cookies={};
      document.cookie.split(';').forEach(function(c){
        var parts=c.trim().split('=');
        if(parts.length>=2) cookies[parts[0]]=parts.slice(1).join('=');
      });
      var base=location.origin;
      var frag='base_url='+encodeURIComponent(base)+'&cookies='+encodeURIComponent(JSON.stringify(cookies));
      window.open('${portalOrigin}/integrations/canvas/callback#'+frag,'_blank','width=480,height=360');
    })();
  `.replace(/\s+/g, " ").trim()

  return `javascript:${encodeURIComponent(code)}`
}

export function CanvasConnectDialog({ children }: CanvasConnectDialogProps) {
  const [open, setOpen] = useState(false)
  const [step, setStep] = useState<Step>("setup")
  const [label, setLabel] = useState("")
  const [initialCount, setInitialCount] = useState(0)
  const [bookmarkletHref, setBookmarkletHref] = useState("")
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
      setLabel("")
    }
  }

  // Build bookmarklet href when dialog opens
  useEffect(() => {
    if (!open) return
    setBookmarkletHref(buildBookmarkletHref(window.location.origin))

    // Snapshot integration count
    ;(async () => {
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
    })()
  }, [open])

  // Poll for new integration
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
    } catch {}
  }, [initialCount])

  useEffect(() => {
    if (step !== "waiting") return
    pollForIntegration()
    intervalRef.current = setInterval(pollForIntegration, 2000)
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

  function handleDone() {
    setStep("waiting")
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        {step === "setup" && (
          <>
            <DialogHeader>
              <DialogTitle>Connect Canvas LMS</DialogTitle>
              <DialogDescription>
                Connect your Canvas account in two steps. No extension required.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col gap-4 py-2">
              <div className="flex flex-col gap-1.5">
                <label className="text-sm font-medium" htmlFor="canvas-label">
                  Account name (optional)
                </label>
                <Input
                  id="canvas-label"
                  placeholder="e.g. CWRU Canvas"
                  value={label}
                  onChange={(e) => setLabel(e.target.value)}
                />
              </div>

              <div className="rounded-lg bg-muted p-4">
                <p className="mb-3 text-sm font-medium text-foreground">Step 1: Add the bookmarklet</p>
                <p className="mb-2 text-sm text-muted-foreground">
                  Drag this button to your bookmarks bar:
                </p>
                <div className="flex justify-center">
                  {/* eslint-disable-next-line @next/next/no-html-link-for-pages */}
                  <a
                    href={bookmarkletHref}
                    onClick={(e) => e.preventDefault()}
                    draggable
                    className="inline-flex items-center gap-2 rounded-md border border-primary bg-primary px-4 py-2 text-sm font-medium text-primary-foreground shadow-sm hover:opacity-90 cursor-grab active:cursor-grabbing"
                  >
                    Connect Canvas
                  </a>
                </div>
                <p className="mt-2 text-xs text-muted-foreground text-center">
                  Drag it — don&apos;t click it here
                </p>
              </div>

              <div className="rounded-lg bg-muted p-4">
                <p className="mb-2 text-sm font-medium text-foreground">Step 2: Use it on Canvas</p>
                <ol className="flex flex-col gap-1.5 text-sm text-muted-foreground">
                  <li>1. Go to your Canvas site and log in</li>
                  <li>2. Click the <strong>&quot;Connect Canvas&quot;</strong> bookmark</li>
                  <li>3. A popup will confirm the connection</li>
                </ol>
              </div>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button onClick={handleDone}>
                I&apos;ve clicked the bookmarklet
              </Button>
            </DialogFooter>
          </>
        )}

        {step === "waiting" && (
          <>
            <DialogHeader>
              <DialogTitle>Waiting for connection...</DialogTitle>
              <DialogDescription>
                Click the &quot;Connect Canvas&quot; bookmarklet while on your Canvas site.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col items-center gap-4 py-6">
              <IconLoader2 className="size-8 animate-spin text-muted-foreground" />
              <p className="text-center text-sm text-muted-foreground">
                Listening for your Canvas session...
              </p>
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setStep("setup")}>
                Back
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
