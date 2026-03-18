"use client"

import { useState } from "react"
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

interface ConnectDialogProps {
  provider: string
  providerName: string
  children: React.ReactNode
}

export function ConnectDialog({ provider, providerName, children }: ConnectDialogProps) {
  const [open, setOpen] = useState(false)
  const [label, setLabel] = useState("")

  function handleConnect() {
    const accountLabel = label.trim() || `${providerName} Account`
    window.location.href = `/api/integrations/${provider}/connect?label=${encodeURIComponent(accountLabel)}`
  }

  return (
    <Dialog open={open} onOpenChange={(v) => { setOpen(v); if (!v) setLabel("") }}>
      <DialogTrigger asChild>
        {children}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Connect {providerName}</DialogTitle>
          <DialogDescription>
            Give this account a name so you can tell it apart from other {providerName} accounts.
          </DialogDescription>
        </DialogHeader>
        <FieldGroup>
          <Field>
            <FieldLabel htmlFor="account-label">Account name</FieldLabel>
            <Input
              id="account-label"
              placeholder={`e.g. Work ${providerName}, Personal`}
              value={label}
              onChange={(e) => setLabel(e.target.value)}
              onKeyDown={(e) => { if (e.key === "Enter") { e.preventDefault(); handleConnect() } }}
              autoFocus
            />
          </Field>
        </FieldGroup>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <Button onClick={handleConnect}>
            Connect
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
