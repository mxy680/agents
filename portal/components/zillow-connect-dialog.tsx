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

interface ZillowConnectDialogProps {
  children: React.ReactNode
}

export function ZillowConnectDialog({ children }: ZillowConnectDialogProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [label, setLabel] = useState("")
  const [proxyUrl, setProxyUrl] = useState("")
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  async function handleSave() {
    setError("")
    setSaving(true)

    try {
      const res = await fetch("/api/integrations/zillow/save", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          label: label.trim() || "Zillow",
          proxy_url: proxyUrl.trim(),
        }),
      })

      const data = await res.json()
      if (!res.ok) {
        setError(data.error || "Failed to save configuration")
        return
      }

      setOpen(false)
      setLabel("")
      setProxyUrl("")
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
    setProxyUrl("")
    setError("")
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) handleReset(); else setOpen(true) }}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Connect Zillow</DialogTitle>
          <DialogDescription>
            Zillow works without any credentials. Optionally configure a proxy to avoid rate limits in production.
          </DialogDescription>
        </DialogHeader>
        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="zillow-label">Account name</FieldLabel>
            <Input
              id="zillow-label"
              placeholder="e.g. Zillow, Real Estate Search"
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              autoFocus
            />
          </Field>
          <Field>
            <FieldLabel htmlFor="zillow-proxy-url">Proxy URL (optional)</FieldLabel>
            <Input
              id="zillow-proxy-url"
              placeholder="e.g. http://user:pass@proxy:8080 or socks5://..."
              value={proxyUrl}
              onChange={(e) => setProxyUrl(e.target.value)}
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
          <Button onClick={handleSave} disabled={saving}>
            {saving ? "Saving..." : "Connect"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
