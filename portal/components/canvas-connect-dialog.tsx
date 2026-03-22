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
  const [cookieData, setCookieData] = useState("")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  async function handleSave() {
    setError("")

    // Parse the pasted cookie JSON
    let parsed: Record<string, string>
    try {
      parsed = JSON.parse(cookieData.trim())
    } catch {
      setError("Invalid data. Click the extension icon on your Canvas site and paste what it copies.")
      return
    }

    // Extract site URL
    const baseUrl = parsed._site_url
    if (!baseUrl) {
      setError("Missing site URL. Make sure you copied from the Emdash Cookie Helper extension.")
      return
    }

    // Find session cookie
    const sessionCookie =
      parsed["_normandy_session"] ||
      parsed["canvas_session"] ||
      parsed["_legacy_normandy_session"]

    if (!sessionCookie) {
      setError("No Canvas session cookie found. Make sure you are logged into Canvas before copying.")
      return
    }

    setSaving(true)

    try {
      const res = await fetch("/api/integrations/canvas/save", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          label: label.trim() || "Canvas LMS",
          base_url: baseUrl,
          session_cookie: sessionCookie,
          csrf_token: parsed["_csrf_token"] || undefined,
          log_session_id: parsed["log_session_id"] || undefined,
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
    setCookieData("")
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
            Use the Emdash Cookie Helper extension to copy your Canvas session, then paste it here.
          </DialogDescription>
        </DialogHeader>
        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="canvas-label">Account name</FieldLabel>
            <Input
              id="canvas-label"
              placeholder="e.g. CWRU Canvas"
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              autoFocus
            />
          </Field>
          <div className="rounded-lg bg-muted p-3 text-sm text-muted-foreground">
            <ol className="flex flex-col gap-1">
              <li>1. Go to your Canvas site and log in</li>
              <li>2. Click the <strong>Emdash Cookie Helper</strong> extension icon</li>
              <li>3. Click <strong>&quot;Copy Cookies&quot;</strong></li>
              <li>4. Paste below</li>
            </ol>
          </div>
          <Field>
            <FieldLabel htmlFor="canvas-cookies">Paste cookies</FieldLabel>
            <Input
              id="canvas-cookies"
              placeholder='Click "Copy Cookies" in the extension, then paste here'
              value={cookieData}
              onChange={(e) => setCookieData(e.target.value)}
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
          <Button onClick={handleSave} disabled={saving || !cookieData.trim()}>
            {saving ? "Connecting..." : "Connect"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
