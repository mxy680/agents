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
          <DialogTitle>Zillow Proxy</DialogTitle>
          <DialogDescription>
            A residential proxy lets agents search Zillow without browser cookies. Requests rotate through residential IPs, bypassing bot detection. Works with laptop closed.
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
            <FieldLabel htmlFor="zillow-proxy-url">Proxy URL</FieldLabel>
            <Input
              id="zillow-proxy-url"
              placeholder="http://user:pass@proxy.example.com:8080"
              value={proxyUrl}
              onChange={(e) => setProxyUrl(e.target.value)}
            />
            <p className="text-xs text-muted-foreground mt-1">
              Supports http://, https://, socks4://, socks5://
            </p>
          </Field>
        </FieldGroup>
        {error && (
          <p className="text-sm text-destructive">{error}</p>
        )}
        <DialogFooter>
          <Button variant="outline" onClick={handleReset} disabled={saving}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={saving || !proxyUrl.trim()}>
            {saving ? "Saving..." : "Save"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
