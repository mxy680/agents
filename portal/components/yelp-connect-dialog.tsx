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

interface YelpConnectDialogProps {
  children: React.ReactNode
}

export function YelpConnectDialog({ children }: YelpConnectDialogProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [label, setLabel] = useState("")
  const [apiKey, setApiKey] = useState("")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  async function handleSave() {
    setError("")
    setSaving(true)

    try {
      const res = await fetch("/api/integrations/yelp/save", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          label: label.trim() || "Yelp",
          api_key: apiKey.trim(),
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
    setError("")
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) handleReset(); else setOpen(true) }}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Connect Yelp</DialogTitle>
          <DialogDescription>
            Enter your Yelp Fusion API key. You can find it in the{" "}
            <a
              href="https://www.yelp.com/developers/v3/manage_app"
              target="_blank"
              rel="noopener noreferrer"
              className="underline"
            >
              Yelp Developer portal
            </a>
            .
          </DialogDescription>
        </DialogHeader>
        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="yelp-label">Account name</FieldLabel>
            <Input
              id="yelp-label"
              placeholder="e.g. Yelp, My App"
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              autoFocus
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="yelp-api-key">API Key</FieldLabel>
            <Input
              id="yelp-api-key"
              type="password"
              placeholder="Enter your Yelp Fusion API key"
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
          <Button onClick={handleSave} disabled={saving || !apiKey.trim()}>
            {saving ? "Saving..." : "Connect"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
