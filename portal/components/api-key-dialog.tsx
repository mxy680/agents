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

interface ApiKeyDialogProps {
  provider: string
  providerName: string
  children: React.ReactNode
}

export function ApiKeyDialog({ provider, providerName, children }: ApiKeyDialogProps) {
  const router = useRouter()
  const [open, setOpen] = useState(false)
  const [token, setToken] = useState("")
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
          provider,
          label: `${providerName} Account`,
          cookies: { token: token.trim() },
        }),
      })

      const data = await res.json()
      if (!res.ok) {
        setError(data.error || "Failed to save")
        return
      }

      setOpen(false)
      setToken("")
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
    setError("")
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) handleReset(); else setOpen(true) }}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Connect {providerName}</DialogTitle>
          <DialogDescription>
            Enter your {providerName} API token.
          </DialogDescription>
        </DialogHeader>
        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="api-token">API Token</FieldLabel>
            <Input
              id="api-token"
              type="password"
              placeholder={`Paste your ${providerName} token`}
              value={token}
              onChange={(e) => setToken(e.target.value)}
              autoFocus
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
