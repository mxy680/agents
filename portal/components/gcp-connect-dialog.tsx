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

interface GCPConnectDialogProps {
  children: React.ReactNode
}

export function GCPConnectDialog({ children }: GCPConnectDialogProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [token, setToken] = useState("")
  const [projectId, setProjectId] = useState("")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  async function handleSave() {
    setError("")
    setSaving(true)

    try {
      const res = await fetch("/api/integrations/save-cookies", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          provider: "gcp",
          label: "GCP Account",
          cookies: {
            token: token.trim(),
            ...(projectId.trim() ? { project_id: projectId.trim() } : {}),
          },
        }),
      })

      const data = await res.json()
      if (!res.ok) {
        setError(data.error || "Failed to save")
        return
      }

      setOpen(false)
      setToken("")
      setProjectId("")
      router.refresh()
    } catch {
      setError("Network error. Please try again.")
    } finally {
      setSaving(false)
    }
  }

  function handleReset() {
    setOpen(false)
    setToken("")
    setProjectId("")
    setError("")
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) handleReset(); else setOpen(true) }}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Connect Google Cloud Platform</DialogTitle>
          <DialogDescription>
            Enter a GCP access token (from <code>gcloud auth print-access-token</code> or a service account).
            Project ID is optional — set it to avoid passing <code>--project</code> on every command.
          </DialogDescription>
        </DialogHeader>
        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="gcp-token">Access Token</FieldLabel>
            <Input
              id="gcp-token"
              type="password"
              placeholder="ya29.xxx..."
              value={token}
              onChange={(e) => setToken(e.target.value)}
              autoFocus
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="gcp-project">Default Project ID (optional)</FieldLabel>
            <Input
              id="gcp-project"
              type="text"
              placeholder="my-project-id"
              value={projectId}
              onChange={(e) => setProjectId(e.target.value)}
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
          <Button onClick={handleSave} disabled={saving || !token.trim()}>
            {saving ? "Saving..." : "Connect"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
