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

interface BlueBubblesConnectDialogProps {
  children: React.ReactNode
}

export function BlueBubblesConnectDialog({ children }: BlueBubblesConnectDialogProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [label, setLabel] = useState("")
  const [serverUrl, setServerUrl] = useState("")
  const [password, setPassword] = useState("")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  async function handleSave() {
    setError("")
    setSaving(true)

    try {
      const res = await fetch("/api/integrations/bluebubbles/save", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          label: label.trim() || "iMessage (BlueBubbles)",
          url: serverUrl.trim(),
          password: password.trim(),
        }),
      })

      const data = await res.json()
      if (!res.ok) {
        setError(data.error || "Failed to save credentials")
        return
      }

      setOpen(false)
      setLabel("")
      setServerUrl("")
      setPassword("")
      router.refresh()
    } catch {
      setError("Network error. Please try again.")
    } finally {
      setSaving(false)
    }
  }

  function handleReset() {
    setOpen(false)
    setLabel("")
    setServerUrl("")
    setPassword("")
    setError("")
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) handleReset(); else setOpen(true) }}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Connect iMessage</DialogTitle>
          <DialogDescription>
            Enter your BlueBubbles server URL and password. BlueBubbles must be running on a Mac signed into your Apple ID.
          </DialogDescription>
        </DialogHeader>
        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="bb-label">Account name</FieldLabel>
            <Input
              id="bb-label"
              placeholder="e.g. Personal iMessage, Work Mac"
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              autoFocus
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="bb-server-url">Server URL</FieldLabel>
            <Input
              id="bb-server-url"
              placeholder="https://your-mac.ngrok.io"
              value={serverUrl}
              onChange={(e) => setServerUrl(e.target.value)}
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="bb-password">Server password</FieldLabel>
            <Input
              id="bb-password"
              type="password"
              placeholder="Enter your BlueBubbles password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
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
          <Button onClick={handleSave} disabled={saving || !serverUrl.trim() || !password.trim()}>
            {saving ? "Connecting..." : "Connect"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
