"use client"

import { useState, useEffect, Suspense } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { IconPlus, IconMessage, IconChevronLeft, IconMenu2, IconTrash, IconLogout } from "@tabler/icons-react"

interface Conversation {
  id: string
  title: string | null
  agent_name: string
  updated_at: string
}

function formatDate(iso: string): string {
  const d = new Date(iso)
  const now = new Date()
  const diffMs = now.getTime() - d.getTime()
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))
  if (diffDays === 0) return d.toLocaleTimeString(undefined, { hour: "2-digit", minute: "2-digit" })
  if (diffDays === 1) return "Yesterday"
  if (diffDays < 7) return d.toLocaleDateString(undefined, { weekday: "short" })
  return d.toLocaleDateString(undefined, { month: "short", day: "numeric" })
}

function ChatLayoutInner({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const agent = searchParams.get("agent") ?? ""
  const convId = searchParams.get("conv") ?? ""

  const [conversations, setConversations] = useState<Conversation[]>([])
  const [sidebarOpen, setSidebarOpen] = useState(false)

  useEffect(() => {
    async function load() {
      try {
        // Cookie is sent automatically — no code param needed
        const res = await fetch(`/api/conversations?agent=${agent}`)
        if (res.ok) {
          const data = await res.json() as { conversations: Conversation[] }
          setConversations(data.conversations)
        }
      } catch {}
    }
    load()

    // Refresh every 10 seconds
    const interval = setInterval(load, 10000)
    return () => clearInterval(interval)
  }, [agent])

  function newChat() {
    setSidebarOpen(false)
    window.location.href = `/chat?agent=${agent}`
  }

  function openConversation(id: string) {
    setSidebarOpen(false)
    window.location.href = `/chat?agent=${agent}&conv=${id}`
  }

  return (
    <div className="flex h-dvh bg-[#0a0a0a]">
      {/* Sidebar overlay on mobile */}
      {sidebarOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 md:hidden"
          onClick={() => setSidebarOpen(false)}
        />
      )}

      {/* Sidebar */}
      <div
        className={`${
          sidebarOpen ? "translate-x-0" : "-translate-x-full"
        } md:translate-x-0 fixed md:relative z-50 md:z-auto w-64 h-full bg-[#111] border-r border-neutral-800 flex flex-col transition-transform duration-200`}
      >
        <div className="flex items-center justify-between p-3 border-b border-neutral-800">
          <button
            onClick={() => router.push("/")}
            className="text-xs text-neutral-500 hover:text-white flex items-center gap-1 transition-colors"
          >
            <IconChevronLeft className="size-3" />
            Agents
          </button>
          <button
            onClick={newChat}
            className="flex items-center gap-1.5 px-2.5 py-1.5 bg-neutral-800 hover:bg-neutral-700 rounded text-xs text-white transition-colors"
          >
            <IconPlus className="size-3" />
            New
          </button>
        </div>

        <div className="flex-1 overflow-y-auto p-2">
          {conversations.length === 0 ? (
            <p className="text-xs text-neutral-600 text-center py-8">
              No conversations yet
            </p>
          ) : (
            <div className="flex flex-col gap-0.5">
              {conversations.map((conv) => (
                <div
                  key={conv.id}
                  className={`group flex items-center gap-1 px-2.5 py-2 rounded text-xs transition-colors ${
                    convId === conv.id
                      ? "bg-neutral-800 text-white"
                      : "text-neutral-400 hover:bg-neutral-800/50 hover:text-neutral-200"
                  }`}
                >
                  <button
                    onClick={() => openConversation(conv.id)}
                    className="flex-1 text-left min-w-0"
                  >
                    <div className="flex items-center gap-2">
                      <IconMessage className="size-3 shrink-0 text-neutral-600" />
                      <span className="truncate flex-1">
                        {conv.title ?? "New conversation"}
                      </span>
                    </div>
                    <span className="text-[10px] text-neutral-600 ml-5">
                      {formatDate(conv.updated_at)}
                    </span>
                  </button>
                  <button
                    onClick={async (e) => {
                      e.stopPropagation()
                      try {
                        // Cookie sent automatically — no code param needed
                        await fetch(`/api/conversations?id=${conv.id}`, { method: "DELETE" })
                        setConversations((prev) => prev.filter((c) => c.id !== conv.id))
                        if (convId === conv.id) newChat()
                      } catch {}
                    }}
                    className="opacity-0 group-hover:opacity-100 p-1 text-neutral-600 hover:text-red-400 transition-all shrink-0"
                    title="Delete conversation"
                  >
                    <IconTrash className="size-3" />
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Sign out */}
        <div className="p-3 border-t border-neutral-800">
          <button
            onClick={async () => {
              await fetch("/api/logout", { method: "POST" })
              window.location.href = "/"
            }}
            className="flex items-center gap-1.5 w-full px-2.5 py-1.5 text-xs text-neutral-500 hover:text-white hover:bg-neutral-800 rounded transition-colors"
          >
            <IconLogout className="size-3" />
            Sign out
          </button>
        </div>
      </div>

      {/* Main content */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Mobile menu button */}
        <button
          onClick={() => setSidebarOpen(true)}
          className="md:hidden absolute top-4 left-4 z-30 p-1.5 text-neutral-500 hover:text-white"
        >
          <IconMenu2 className="size-5" />
        </button>

        {children}
      </div>
    </div>
  )
}

export default function ClientChatLayout({ children }: { children: React.ReactNode }) {
  return (
    <Suspense fallback={<div className="flex h-dvh bg-[#0a0a0a] items-center justify-center"><div className="text-neutral-500 text-sm">Loading...</div></div>}>
      <ChatLayoutInner>{children}</ChatLayoutInner>
    </Suspense>
  )
}
