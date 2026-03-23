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
import { IconPlus, IconX } from "@tabler/icons-react"

interface Assignment {
  id: string
  client_id: string
  notes: string | null
  created_at: string
  clients: { id: string; name: string; email: string | null; active: boolean } | null
}

interface AvailableClient {
  id: string
  name: string
  email: string | null
}

interface AgentClientsProps {
  agentName: string
  initialAssignments: Assignment[]
  availableClients: AvailableClient[]
}

export function AgentClients({ agentName, initialAssignments, availableClients }: AgentClientsProps) {
  const [assignments, setAssignments] = React.useState(initialAssignments)
  const [showDropdown, setShowDropdown] = React.useState(false)
  const [saving, setSaving] = React.useState(false)

  const assignedClientIds = new Set(assignments.map((a) => a.client_id))
  const unassignedClients = availableClients.filter((c) => !assignedClientIds.has(c.id))

  async function handleAssign(client: AvailableClient) {
    setShowDropdown(false)
    setSaving(true)

    // Optimistic update
    const optimistic: Assignment = {
      id: `optimistic-${client.id}`,
      client_id: client.id,
      notes: null,
      created_at: new Date().toISOString(),
      clients: { id: client.id, name: client.name, email: client.email, active: true },
    }
    setAssignments((prev) => [optimistic, ...prev])

    try {
      const res = await fetch(`/api/agents/${agentName}/clients`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ client_id: client.id }),
      })

      if (res.ok) {
        const { assignment } = await res.json()
        setAssignments((prev) =>
          prev.map((a) =>
            a.id === optimistic.id
              ? { ...assignment, clients: optimistic.clients }
              : a
          )
        )
      } else {
        // Roll back
        setAssignments((prev) => prev.filter((a) => a.id !== optimistic.id))
      }
    } finally {
      setSaving(false)
    }
  }

  async function handleRemove(clientId: string) {
    const previous = assignments
    setAssignments((prev) => prev.filter((a) => a.client_id !== clientId))

    const res = await fetch(`/api/agents/${agentName}/clients`, {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ client_id: clientId }),
    })

    if (!res.ok) {
      setAssignments(previous)
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h2 className="text-base font-semibold">Assigned Clients</h2>
        <div className="relative">
          <Button
            size="sm"
            disabled={saving || unassignedClients.length === 0}
            onClick={() => setShowDropdown((v) => !v)}
          >
            <IconPlus className="size-4" />
            Assign Client
          </Button>
          {showDropdown && unassignedClients.length > 0 && (
            <div className="absolute right-0 z-10 mt-1 w-64 rounded-md border bg-popover shadow-md">
              {unassignedClients.map((client) => (
                <button
                  key={client.id}
                  className="flex w-full flex-col px-3 py-2 text-left text-sm hover:bg-accent"
                  onClick={() => handleAssign(client)}
                >
                  <span className="font-medium">{client.name}</span>
                  {client.email && (
                    <span className="text-xs text-muted-foreground">{client.email}</span>
                  )}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Assigned</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {assignments.length === 0 && (
            <TableRow>
              <TableCell colSpan={4} className="text-center text-muted-foreground">
                No clients assigned yet.
              </TableCell>
            </TableRow>
          )}
          {assignments.map((assignment) => (
            <TableRow key={assignment.id}>
              <TableCell className="font-medium">
                {assignment.clients?.name ?? "Unknown"}
              </TableCell>
              <TableCell>{assignment.clients?.email ?? "—"}</TableCell>
              <TableCell className="text-sm text-muted-foreground">
                {new Date(assignment.created_at).toLocaleDateString()}
              </TableCell>
              <TableCell>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => handleRemove(assignment.client_id)}
                  title="Remove"
                >
                  <IconX className="size-4" />
                </Button>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
