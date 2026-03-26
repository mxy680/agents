"use client"

import * as React from "react"
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from "@/components/ui/table"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { IconPlus, IconCheck, IconX, IconCopy, IconTrash } from "@tabler/icons-react"

interface ClientAccess {
  id: string
  code: string
  client_name: string
  agent_name: string
  agent_names: string[] | null
  active: boolean
  created_at: string
}

interface Agent {
  name: string
  displayName: string
}

interface Props {
  initialClients: ClientAccess[]
  agents: Agent[]
}

function generateCode(): string {
  const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
  let code = ""
  for (let i = 0; i < 4; i++) code += chars[Math.floor(Math.random() * chars.length)]
  code += "-"
  for (let i = 0; i < 4; i++) code += chars[Math.floor(Math.random() * chars.length)]
  return code
}

export function ClientsTable({ initialClients, agents }: Props) {
  const [clients, setClients] = React.useState(initialClients)
  const [showForm, setShowForm] = React.useState(false)
  const [name, setName] = React.useState("")
  const [code, setCode] = React.useState(() => generateCode())
  const [selectedAgents, setSelectedAgents] = React.useState<string[]>([])
  const [saving, setSaving] = React.useState(false)
  const [copied, setCopied] = React.useState<string | null>(null)

  async function handleCreate() {
    if (!name.trim() || selectedAgents.length === 0) return
    setSaving(true)
    try {
      const res = await fetch("/api/admin/clients", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          code: code.trim(),
          client_name: name.trim(),
          agent_names: selectedAgents,
        }),
      })
      if (res.ok) {
        const { client } = await res.json() as { client: ClientAccess }
        setClients((prev) => [client, ...prev])
        setName("")
        setCode(generateCode())
        setSelectedAgents([])
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
      setClients((prev) => prev.map((c) => c.id === id ? { ...c, active: !active } : c))
    }
  }

  async function handleDelete(id: string) {
    const res = await fetch(`/api/admin/clients/${id}`, { method: "DELETE" })
    if (res.ok) {
      setClients((prev) => prev.filter((c) => c.id !== id))
    }
  }

  function copyCode(c: string) {
    navigator.clipboard.writeText(c)
    setCopied(c)
    setTimeout(() => setCopied(null), 2000)
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
          <div className="grid gap-3 sm:grid-cols-2">
            <Input placeholder="Client name" value={name} onChange={(e) => setName(e.target.value)} />
            <div className="flex items-center gap-2">
              <Input placeholder="Access code" value={code} onChange={(e) => setCode(e.target.value)} className="font-mono" />
              <Button size="sm" variant="outline" onClick={() => setCode(generateCode())} type="button">
                Regenerate
              </Button>
            </div>
          </div>
          <div>
            <p className="text-xs text-muted-foreground mb-2">Assign agents:</p>
            <div className="flex flex-wrap gap-2">
              {agents.map((agent) => (
                <button
                  key={agent.name}
                  onClick={() => setSelectedAgents((prev) =>
                    prev.includes(agent.name) ? prev.filter((a) => a !== agent.name) : [...prev, agent.name]
                  )}
                  className={`px-2.5 py-1 rounded text-xs border transition-colors ${
                    selectedAgents.includes(agent.name)
                      ? "bg-primary text-primary-foreground border-primary"
                      : "bg-muted/40 text-muted-foreground border-border hover:border-foreground/30"
                  }`}
                >
                  {agent.displayName}
                </button>
              ))}
            </div>
          </div>
          <div className="flex gap-2">
            <Button size="sm" onClick={handleCreate} disabled={saving || !name.trim() || selectedAgents.length === 0}>
              {saving ? "Saving..." : "Create"}
            </Button>
            <Button size="sm" variant="outline" onClick={() => setShowForm(false)}>Cancel</Button>
          </div>
        </div>
      )}

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Access Code</TableHead>
            <TableHead>Agents</TableHead>
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
          {clients.map((client) => {
            const agentList = client.agent_names?.length ? client.agent_names : [client.agent_name]
            return (
              <TableRow key={client.id}>
                <TableCell className="font-medium">{client.client_name}</TableCell>
                <TableCell>
                  <div className="flex items-center gap-1.5">
                    <code className="text-xs bg-muted px-1.5 py-0.5 rounded font-mono">{client.code}</code>
                    <button onClick={() => copyCode(client.code)} className="text-muted-foreground hover:text-foreground transition-colors" title="Copy code">
                      {copied === client.code ? <IconCheck className="size-3.5 text-green-500" /> : <IconCopy className="size-3.5" />}
                    </button>
                  </div>
                </TableCell>
                <TableCell>
                  <div className="flex flex-wrap gap-1">
                    {[...new Set(agentList)].map((a) => (
                      <Badge key={a} variant="outline" className="text-xs capitalize">{a.replace(/-/g, " ")}</Badge>
                    ))}
                  </div>
                </TableCell>
                <TableCell>
                  <Badge variant={client.active ? "default" : "secondary"}>{client.active ? "Active" : "Inactive"}</Badge>
                </TableCell>
                <TableCell className="text-sm text-muted-foreground">
                  {new Date(client.created_at).toLocaleDateString()}
                </TableCell>
                <TableCell>
                  <div className="flex items-center gap-1">
                    <Button size="sm" variant="ghost" onClick={() => handleToggleActive(client.id, client.active)} title={client.active ? "Deactivate" : "Activate"}>
                      {client.active ? <IconX className="size-4" /> : <IconCheck className="size-4" />}
                    </Button>
                    <Button size="sm" variant="ghost" onClick={() => handleDelete(client.id)} title="Delete" className="text-muted-foreground hover:text-red-400">
                      <IconTrash className="size-4" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}
