"use client"

import React, { useState, useCallback } from "react"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import {
  IconPlus,
  IconDotsVertical,
  IconPencil,
  IconStar,
  IconStarFilled,
  IconTrash,
} from "@tabler/icons-react"
import { cn } from "@/lib/utils"

interface Conversation {
  id: string
  user_id: string
  agent_name: string
  title: string | null
  starred: boolean
  session_id: string | null
  created_at: string
  updated_at: string
}

interface Props {
  agentName: string
  conversations: Conversation[]
  activeId?: string
}

type GroupKey = "Starred" | "Today" | "Yesterday" | "Previous 7 Days" | "Older"

function groupConversations(conversations: Conversation[]): Map<GroupKey, Conversation[]> {
  const now = new Date()
  const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const startOfYesterday = new Date(startOfToday)
  startOfYesterday.setDate(startOfYesterday.getDate() - 1)
  const startOf7DaysAgo = new Date(startOfToday)
  startOf7DaysAgo.setDate(startOf7DaysAgo.getDate() - 7)

  const groups = new Map<GroupKey, Conversation[]>([
    ["Starred", []],
    ["Today", []],
    ["Yesterday", []],
    ["Previous 7 Days", []],
    ["Older", []],
  ])

  for (const conv of conversations) {
    if (conv.starred) {
      groups.get("Starred")!.push(conv)
      continue
    }
    const updatedAt = new Date(conv.updated_at)
    if (updatedAt >= startOfToday) {
      groups.get("Today")!.push(conv)
    } else if (updatedAt >= startOfYesterday) {
      groups.get("Yesterday")!.push(conv)
    } else if (updatedAt >= startOf7DaysAgo) {
      groups.get("Previous 7 Days")!.push(conv)
    } else {
      groups.get("Older")!.push(conv)
    }
  }

  return groups
}

interface ConversationItemProps {
  conv: Conversation
  agentName: string
  isActive: boolean
  onRename: (id: string, currentTitle: string) => void
  onDelete: (id: string) => void
  onStarToggle: (id: string, starred: boolean) => void
}

function ConversationItem({
  conv,
  agentName,
  isActive,
  onRename,
  onDelete,
  onStarToggle,
}: ConversationItemProps) {
  const title = conv.title || "Untitled"

  return (
    <div
      className={cn(
        "group flex items-center gap-1 px-2 py-1.5 text-xs hover:bg-muted/60 transition-colors",
        isActive && "bg-muted/80"
      )}
    >
      <a
        href={`/chat/${agentName}/${conv.id}`}
        className="flex-1 truncate text-foreground/80 hover:text-foreground"
        title={title}
      >
        {title}
      </a>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <button
            className="shrink-0 opacity-0 group-hover:opacity-100 focus:opacity-100 p-0.5 text-muted-foreground hover:text-foreground transition-opacity"
            aria-label="Conversation actions"
          >
            <IconDotsVertical className="size-3.5" />
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end">
          <DropdownMenuItem onClick={() => onRename(conv.id, conv.title ?? "")}>
            <IconPencil className="size-3.5" />
            Rename
          </DropdownMenuItem>
          <DropdownMenuItem onClick={() => onStarToggle(conv.id, conv.starred)}>
            {conv.starred ? (
              <>
                <IconStarFilled className="size-3.5" />
                Unstar
              </>
            ) : (
              <>
                <IconStar className="size-3.5" />
                Star
              </>
            )}
          </DropdownMenuItem>
          <DropdownMenuItem variant="destructive" onClick={() => onDelete(conv.id)}>
            <IconTrash className="size-3.5" />
            Delete
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  )
}

export function ChatSidebar({ agentName, conversations: initialConversations, activeId }: Props) {
  const router = useRouter()
  const [conversations, setConversations] = useState<Conversation[]>(initialConversations)

  // Rename state
  const [renameDialogOpen, setRenameDialogOpen] = useState(false)
  const [renamingId, setRenamingId] = useState<string | null>(null)
  const [renameValue, setRenameValue] = useState("")

  // Delete state
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deletingId, setDeletingId] = useState<string | null>(null)

  const handleRename = useCallback((id: string, currentTitle: string) => {
    setRenamingId(id)
    setRenameValue(currentTitle)
    setRenameDialogOpen(true)
  }, [])

  const submitRename = useCallback(async () => {
    if (!renamingId) return
    const title = renameValue.trim() || "Untitled"

    // Optimistic update
    setConversations((prev) =>
      prev.map((c) => (c.id === renamingId ? { ...c, title } : c))
    )
    setRenameDialogOpen(false)

    await fetch(`/api/conversations/${renamingId}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ title }),
    })
    router.refresh()
  }, [renamingId, renameValue, router])

  const handleStarToggle = useCallback(async (id: string, starred: boolean) => {
    const newStarred = !starred

    // Optimistic update
    setConversations((prev) =>
      prev.map((c) => (c.id === id ? { ...c, starred: newStarred } : c))
    )

    await fetch(`/api/conversations/${id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ starred: newStarred }),
    })
    router.refresh()
  }, [router])

  const handleDeleteRequest = useCallback((id: string) => {
    setDeletingId(id)
    setDeleteDialogOpen(true)
  }, [])

  const confirmDelete = useCallback(async () => {
    if (!deletingId) return

    // Optimistic removal
    setConversations((prev) => prev.filter((c) => c.id !== deletingId))
    setDeleteDialogOpen(false)

    await fetch(`/api/conversations/${deletingId}`, { method: "DELETE" })

    // If we deleted the active conversation, navigate to the agent chat root
    if (deletingId === activeId) {
      router.push(`/chat/${agentName}`)
    } else {
      router.refresh()
    }
  }, [deletingId, activeId, agentName, router])

  const groups = groupConversations(conversations)
  const groupOrder: GroupKey[] = ["Starred", "Today", "Yesterday", "Previous 7 Days", "Older"]

  return (
    <>
      <aside className="flex h-full w-56 shrink-0 flex-col border-r bg-background">
        {/* New Chat */}
        <div className="p-2 border-b">
          <a href={`/chat/${agentName}`}>
            <Button variant="outline" size="sm" className="w-full justify-start gap-2 text-xs">
              <IconPlus className="size-3.5" />
              New Chat
            </Button>
          </a>
        </div>

        {/* Conversation list */}
        <div className="flex-1 overflow-y-auto py-1">
          {groupOrder.map((groupKey) => {
            const items = groups.get(groupKey) ?? []
            if (items.length === 0) return null
            return (
              <div key={groupKey} className="mb-2">
                <p className="px-3 py-1 text-[10px] font-medium uppercase tracking-wider text-muted-foreground">
                  {groupKey}
                </p>
                {items.map((conv) => (
                  <ConversationItem
                    key={conv.id}
                    conv={conv}
                    agentName={agentName}
                    isActive={conv.id === activeId}
                    onRename={handleRename}
                    onDelete={handleDeleteRequest}
                    onStarToggle={handleStarToggle}
                  />
                ))}
              </div>
            )
          })}

          {conversations.length === 0 && (
            <p className="px-3 py-4 text-xs text-muted-foreground text-center">
              No conversations yet
            </p>
          )}
        </div>
      </aside>

      {/* Rename Dialog */}
      <Dialog open={renameDialogOpen} onOpenChange={setRenameDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Rename conversation</DialogTitle>
          </DialogHeader>
          <Input
            value={renameValue}
            onChange={(e) => setRenameValue(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") submitRename()
            }}
            placeholder="Conversation title"
            autoFocus
          />
          <DialogFooter>
            <Button variant="outline" size="sm" onClick={() => setRenameDialogOpen(false)}>
              Cancel
            </Button>
            <Button size="sm" onClick={submitRename}>
              Save
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirm Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete conversation?</DialogTitle>
          </DialogHeader>
          <p className="text-xs text-muted-foreground">
            This will permanently delete the conversation and all its messages.
          </p>
          <DialogFooter>
            <Button variant="outline" size="sm" onClick={() => setDeleteDialogOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" size="sm" onClick={confirmDelete}>
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
