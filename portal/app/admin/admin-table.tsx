"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"
import { IconCheck, IconX, IconLoader2 } from "@tabler/icons-react"

interface ApprovalRequest {
  id: string
  user_id: string
  user_email: string
  template_id: string
  template_name: string
  template_display_name: string
  status: "pending" | "approved" | "rejected"
  acquired_at: string
  reviewed_at: string | null
  reviewer_note: string | null
}

interface AdminTableProps {
  initialRequests: ApprovalRequest[]
}

type TabValue = "pending" | "approved" | "rejected" | "all"

const STATUS_BADGE: Record<string, { label: string; className: string }> = {
  pending: { label: "Pending", className: "bg-yellow-500/20 text-yellow-400 border-yellow-500/30" },
  approved: { label: "Approved", className: "bg-green-500/20 text-green-400 border-green-500/30" },
  rejected: { label: "Rejected", className: "bg-red-500/20 text-red-400 border-red-500/30" },
}

export function AdminTable({ initialRequests }: AdminTableProps) {
  const [requests, setRequests] = useState<ApprovalRequest[]>(initialRequests)
  const [activeTab, setActiveTab] = useState<TabValue>("pending")
  const [processing, setProcessing] = useState<string | null>(null)
  const [rejectingId, setRejectingId] = useState<string | null>(null)
  const [rejectNote, setRejectNote] = useState("")

  const filtered = activeTab === "all"
    ? requests
    : requests.filter((r) => r.status === activeTab)

  const counts = {
    pending: requests.filter((r) => r.status === "pending").length,
    approved: requests.filter((r) => r.status === "approved").length,
    rejected: requests.filter((r) => r.status === "rejected").length,
    all: requests.length,
  }

  async function handleAction(id: string, action: "approve" | "reject", note?: string) {
    setProcessing(id)
    try {
      const res = await fetch(`/api/admin/agents/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action, note }),
      })
      if (res.ok) {
        const newStatus = action === "approve" ? "approved" : "rejected"
        setRequests((prev) =>
          prev.map((r) =>
            r.id === id
              ? { ...r, status: newStatus, reviewed_at: new Date().toISOString(), reviewer_note: note ?? null }
              : r
          )
        )
        setRejectingId(null)
        setRejectNote("")
      }
    } finally {
      setProcessing(null)
    }
  }

  function formatDate(iso: string) {
    return new Date(iso).toLocaleString(undefined, {
      dateStyle: "medium",
      timeStyle: "short",
    })
  }

  return (
    <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as TabValue)}>
      <TabsList className="mb-4">
        <TabsTrigger value="pending">
          Pending {counts.pending > 0 && <span className="ml-1 text-xs">({counts.pending})</span>}
        </TabsTrigger>
        <TabsTrigger value="approved">
          Approved {counts.approved > 0 && <span className="ml-1 text-xs">({counts.approved})</span>}
        </TabsTrigger>
        <TabsTrigger value="rejected">
          Rejected {counts.rejected > 0 && <span className="ml-1 text-xs">({counts.rejected})</span>}
        </TabsTrigger>
        <TabsTrigger value="all">
          All {counts.all > 0 && <span className="ml-1 text-xs">({counts.all})</span>}
        </TabsTrigger>
      </TabsList>

      {(["pending", "approved", "rejected", "all"] as const).map((tab) => (
        <TabsContent key={tab} value={tab}>
          {filtered.length === 0 ? (
            <p className="text-sm text-muted-foreground py-8 text-center">No requests.</p>
          ) : (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>User</TableHead>
                    <TableHead>Agent</TableHead>
                    <TableHead>Requested</TableHead>
                    {tab !== "pending" && <TableHead>Status</TableHead>}
                    {tab !== "pending" && <TableHead>Reviewed</TableHead>}
                    {tab !== "pending" && <TableHead>Note</TableHead>}
                    {tab === "pending" && <TableHead className="text-right">Actions</TableHead>}
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filtered.map((req) => (
                    <>
                      <TableRow key={req.id}>
                        <TableCell className="font-mono text-xs">{req.user_email}</TableCell>
                        <TableCell>{req.template_display_name}</TableCell>
                        <TableCell className="text-xs text-muted-foreground">
                          {formatDate(req.acquired_at)}
                        </TableCell>
                        {tab !== "pending" && (
                          <TableCell>
                            <Badge
                              variant="outline"
                              className={STATUS_BADGE[req.status]?.className}
                            >
                              {STATUS_BADGE[req.status]?.label ?? req.status}
                            </Badge>
                          </TableCell>
                        )}
                        {tab !== "pending" && (
                          <TableCell className="text-xs text-muted-foreground">
                            {req.reviewed_at ? formatDate(req.reviewed_at) : "—"}
                          </TableCell>
                        )}
                        {tab !== "pending" && (
                          <TableCell className="text-xs text-muted-foreground max-w-48 truncate">
                            {req.reviewer_note ?? "—"}
                          </TableCell>
                        )}
                        {tab === "pending" && (
                          <TableCell className="text-right">
                            <div className="flex items-center justify-end gap-2">
                              <Button
                                size="sm"
                                variant="outline"
                                className="gap-1 border-green-500/50 text-green-400 hover:bg-green-500/10"
                                disabled={processing === req.id}
                                onClick={() => handleAction(req.id, "approve")}
                              >
                                {processing === req.id ? (
                                  <IconLoader2 className="size-3.5 animate-spin" />
                                ) : (
                                  <IconCheck className="size-3.5" />
                                )}
                                Approve
                              </Button>
                              <Button
                                size="sm"
                                variant="outline"
                                className="gap-1 border-red-500/50 text-red-400 hover:bg-red-500/10"
                                disabled={processing === req.id}
                                onClick={() =>
                                  setRejectingId(rejectingId === req.id ? null : req.id)
                                }
                              >
                                <IconX className="size-3.5" />
                                Reject
                              </Button>
                            </div>
                          </TableCell>
                        )}
                      </TableRow>
                      {tab === "pending" && rejectingId === req.id && (
                        <TableRow key={`${req.id}-reject`}>
                          <TableCell colSpan={4} className="bg-muted/30 pb-3">
                            <div className="flex flex-col gap-2 max-w-md">
                              <p className="text-xs text-muted-foreground">
                                Optional: add a reason for rejecting this request.
                              </p>
                              <Textarea
                                placeholder="Rejection reason (optional)"
                                value={rejectNote}
                                onChange={(e) => setRejectNote(e.target.value)}
                                rows={2}
                                className="text-sm"
                              />
                              <div className="flex gap-2">
                                <Button
                                  size="sm"
                                  variant="destructive"
                                  disabled={processing === req.id}
                                  onClick={() => handleAction(req.id, "reject", rejectNote || undefined)}
                                >
                                  {processing === req.id ? (
                                    <IconLoader2 className="size-3.5 animate-spin" />
                                  ) : null}
                                  Confirm Reject
                                </Button>
                                <Button
                                  size="sm"
                                  variant="outline"
                                  onClick={() => {
                                    setRejectingId(null)
                                    setRejectNote("")
                                  }}
                                >
                                  Cancel
                                </Button>
                              </div>
                            </div>
                          </TableCell>
                        </TableRow>
                      )}
                    </>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </TabsContent>
      ))}
    </Tabs>
  )
}
