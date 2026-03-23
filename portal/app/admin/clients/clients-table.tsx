"use client"

import * as React from "react"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { IconPlus, IconCheck, IconX } from "@tabler/icons-react"

interface Client {
  id: string
  name: string
  email: string | null
  notes: string | null
  active: boolean
  created_at: string
}

interface ClientsTableProps {
  initialClients: Client[]
}

export function ClientsTable({ initialClients }: ClientsTableProps) {
  const [clients, setClients] = React.useState(initialClients)
  const [showForm, setShowForm] = React.useState(false)
  const [name, setName] = React.useState("")
  const [email, setEmail] = React.useState("")
  const [notes, setNotes] = React.useState("")
  const [saving, setSaving] = React.useState(false)

  async function handleCreate() {
    if (!name.trim()) return
    setSaving(true)

    try {
      const res = await fetch("/api/admin/clients", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: name.trim(),
          email: email.trim() || null,
          notes: notes.trim() || null,
        }),
      })

      if (res.ok) {
        const { client } = await res.json()
        setClients((prev) => [client, ...prev])
        setName("")
        setEmail("")
        setNotes("")
        setShowForm(false)
      }
    } finally {
      setSaving(false)
    }
  }

  async function handleToggleActive(id: string, active: boolean) {
    const res = await fetch(`/api/admin/clients/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ active: !active }),
    })

    if (res.ok) {
      setClients((prev) =>
        prev.map((c) => (c.id === id ? { ...c, active: !active } : c))
      )
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex justify-end">
        <Button size="sm" onClick={() => setShowForm(!showForm)}>
          <IconPlus className="mr-1 size-4" />
          Add Client
        </Button>
      </div>

      {showForm && (
        <div className="flex flex-col gap-3 rounded-lg border p-4">
          <div className="grid gap-3 sm:grid-cols-3">
            <Input
              placeholder="Name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
            <Input
              placeholder="Email (optional)"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
            />
            <Input
              placeholder="Notes (optional)"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
            />
          </div>
          <div className="flex gap-2">
            <Button size="sm" onClick={handleCreate} disabled={saving || !name.trim()}>
              {saving ? "Saving..." : "Save"}
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => setShowForm(false)}
            >
              Cancel
            </Button>
          </div>
        </div>
      )}

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Notes</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Created</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {clients.length === 0 && (
            <TableRow>
              <TableCell colSpan={6} className="text-center text-muted-foreground">
                No clients yet. Add one to get started.
              </TableCell>
            </TableRow>
          )}
          {clients.map((client) => (
            <TableRow key={client.id}>
              <TableCell className="font-medium">{client.name}</TableCell>
              <TableCell>{client.email ?? "—"}</TableCell>
              <TableCell className="max-w-[200px] truncate">
                {client.notes ?? "—"}
              </TableCell>
              <TableCell>
                <Badge variant={client.active ? "default" : "secondary"}>
                  {client.active ? "Active" : "Inactive"}
                </Badge>
              </TableCell>
              <TableCell className="text-sm text-muted-foreground">
                {new Date(client.created_at).toLocaleDateString()}
              </TableCell>
              <TableCell>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => handleToggleActive(client.id, client.active)}
                  title={client.active ? "Deactivate" : "Activate"}
                >
                  {client.active ? (
                    <IconX className="size-4" />
                  ) : (
                    <IconCheck className="size-4" />
                  )}
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
