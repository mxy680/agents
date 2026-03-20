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

interface FramerConnectDialogProps {
  children: React.ReactNode
}

export function FramerConnectDialog({ children }: FramerConnectDialogProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [label, setLabel] = useState("")
  const [apiKey, setApiKey] = useState("")
  const [projectUrl, setProjectUrl] = useState("")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  async function handleSave() {
    setError("")
    setSaving(true)

    try {
      const res = await fetch("/api/integrations/framer/save", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          label: label.trim() || "Framer Project",
          api_key: apiKey.trim(),
          project_url: projectUrl.trim(),
        }),
      })

      const data = await res.json()
      if (!res.ok) {
        setError(data.error || "Failed to save credentials")
        return
      }

      setOpen(false)
      setLabel("")
      setApiKey("")
      setProjectUrl("")
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
    setApiKey("")
    setProjectUrl("")
    setError("")
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) handleReset(); else setOpen(true) }}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Connect Framer</DialogTitle>
          <DialogDescription>
            Enter your Framer project API key. You can find it in your project settings under &quot;Server API&quot;.
          </DialogDescription>
        </DialogHeader>
        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="framer-label">Account name</FieldLabel>
            <Input
              id="framer-label"
              placeholder="e.g. Marketing Site, Portfolio"
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              autoFocus
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="framer-project-url">Project URL</FieldLabel>
            <Input
              id="framer-project-url"
              placeholder="https://framer.com/projects/Website--aabbccddeeff"
              value={projectUrl}
              onChange={(e) => setProjectUrl(e.target.value)}
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="framer-api-key">API Key</FieldLabel>
            <Input
              id="framer-api-key"
              type="password"
              placeholder="Enter your Framer API key"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
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
          <Button onClick={handleSave} disabled={saving || !apiKey.trim() || !projectUrl.trim()}>
            {saving ? "Saving..." : "Connect"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
