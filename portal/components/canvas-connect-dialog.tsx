"use client"

import { useState } from "react"
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

interface CanvasConnectDialogProps {
  children: React.ReactNode
}

export function CanvasConnectDialog({ children }: CanvasConnectDialogProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [label, setLabel] = useState("")
  const [baseUrl, setBaseUrl] = useState("")
  const [sessionCookie, setSessionCookie] = useState("")
  const [csrfToken, setCsrfToken] = useState("")
  const [logSessionId, setLogSessionId] = useState("")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  async function handleSave() {
    setError("")
    setSaving(true)

    try {
      const res = await fetch("/api/integrations/canvas/save", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          label: label.trim() || "Canvas LMS",
          base_url: baseUrl.trim(),
          session_cookie: sessionCookie.trim(),
          csrf_token: csrfToken.trim(),
          log_session_id: logSessionId.trim() || undefined,
        }),
      })

      const data = await res.json()
      if (!res.ok) {
        setError(data.error || "Failed to save credentials")
        return
      }

      setOpen(false)
      resetFields()
      router.refresh()
    } catch {
      setError("Network error. Please try again.")
    } finally {
      setSaving(false)
    }
  }

  function resetFields() {
    setLabel("")
    setBaseUrl("")
    setSessionCookie("")
    setCsrfToken("")
    setLogSessionId("")
    setError("")
  }

  function handleReset() {
    setOpen(false)
    resetFields()
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) handleReset(); else setOpen(true) }}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Connect Canvas LMS</DialogTitle>
          <DialogDescription>
            Enter your Canvas instance URL and session cookies. Open DevTools in your browser while logged into Canvas, go to Application &gt; Cookies, and copy the values below.
          </DialogDescription>
        </DialogHeader>
        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="canvas-label">Account name</FieldLabel>
            <Input
              id="canvas-label"
              placeholder="e.g. CWRU Canvas, My School"
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              autoFocus
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="canvas-base-url">Canvas URL</FieldLabel>
            <Input
              id="canvas-base-url"
              placeholder="https://canvas.university.edu"
              value={baseUrl}
              onChange={(e) => setBaseUrl(e.target.value)}
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="canvas-session">_normandy_session cookie</FieldLabel>
            <Input
              id="canvas-session"
              type="password"
              placeholder="Paste the _normandy_session cookie value"
              value={sessionCookie}
              onChange={(e) => setSessionCookie(e.target.value)}
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="canvas-csrf">_csrf_token cookie</FieldLabel>
            <Input
              id="canvas-csrf"
              type="password"
              placeholder="Paste the _csrf_token cookie value"
              value={csrfToken}
              onChange={(e) => setCsrfToken(e.target.value)}
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="canvas-log-session">log_session_id cookie (optional)</FieldLabel>
            <Input
              id="canvas-log-session"
              placeholder="Paste the log_session_id cookie value"
              value={logSessionId}
              onChange={(e) => setLogSessionId(e.target.value)}
            />
          </Field>
        </FieldGroup>
        {error && (
          <p className="text-sm text-destructive">{error}</p>
        )}
        <DialogFooter>
          <Button variant="outline" onClick={handleReset} disabled={saving}>
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={saving || !baseUrl.trim() || !sessionCookie.trim() || !csrfToken.trim()}
          >
            {saving ? "Connecting..." : "Connect"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
