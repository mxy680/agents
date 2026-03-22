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
 * Builds a bookmarklet that reads document.cookie on the Canvas page
 * and opens a portal callback page (same-origin) with cookies in the
 * URL fragment. The callback page handles auth via session cookie.
 */
function buildBookmarkletHref(portalOrigin: string) {
  const code = `
    (function(){
      var c={},d=document.cookie.split(';');
      for(var i=0;i<d.length;i++){var p=d[i].trim().split('=');if(p.length>=2)c[p[0]]=p.slice(1).join('=');}
      var u=encodeURIComponent(location.origin);
      var k=encodeURIComponent(JSON.stringify(c));
      window.open('${portalOrigin}/integrations/canvas/callback#base_url='+u+'&cookies='+k,'emdash_canvas','width=480,height=360,popup=1');
    })();
  `.replace(/\s+/g, " ").trim()

  return `javascript:${encodeURIComponent(code)}`
}

export function CanvasConnectDialog({ children }: CanvasConnectDialogProps) {
  const [open, setOpen] = useState(false)
  const [step, setStep] = useState<Step>("setup")
  const [label, setLabel] = useState("")
  const [bookmarkletHref, setBookmarkletHref] = useState("")
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
      setStep("setup")
      setLabel("")
      setBookmarkletHref("")
    }
  }

  // Build bookmarklet when dialog opens + snapshot count
  useEffect(() => {
    if (!open) return
    setBookmarkletHref(buildBookmarkletHref(window.location.origin))
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

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        {step === "setup" && (
          <>
            <DialogHeader>
              <DialogTitle>Connect Canvas LMS</DialogTitle>
              <DialogDescription>
                No extension needed. Use a bookmarklet to connect in one click.
              </DialogDescription>
            </DialogHeader>

            <div className="flex flex-col gap-4 py-2">
              <div className="rounded-lg bg-muted p-4">
                <p className="mb-3 text-sm font-medium text-foreground">Step 1: Save the bookmarklet</p>
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
                  Drag it to your bookmarks bar — don&apos;t click it here
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
              <Button onClick={() => setStep("waiting")}>
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
                Click the &quot;Connect Canvas&quot; bookmarklet while logged into Canvas.
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
