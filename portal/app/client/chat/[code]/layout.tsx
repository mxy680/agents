"use client"

import { useState, useEffect, use } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import { IconPlus, IconMessage, IconChevronLeft, IconMenu2, IconTrash } from "@tabler/icons-react"

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

export default function ClientChatLayout({
  children,
  params,
}: {
  children: React.ReactNode
  params: Promise<{ code: string }>
}) {
  const { code } = use(params)
  const router = useRouter()
  const searchParams = useSearchParams()
  const agent = searchParams.get("agent") ?? ""
  const convId = searchParams.get("conv") ?? ""

  const [conversations, setConversations] = useState<Conversation[]>([])
  const [sidebarOpen, setSidebarOpen] = useState(false)

  useEffect(() => {
    async function load() {
      try {
        const res = await fetch(
          `/api/client/conversations?code=${encodeURIComponent(code)}&agent=${agent}`
        )
        if (res.ok) {
          const data = await res.json() as { conversations: Conversation[] }
          setConversations(data.conversations)
        }
      } catch {}
    }
    load()

    // Refresh every 10 seconds when sidebar is open
    const interval = setInterval(load, 10000)
    return () => clearInterval(interval)
  }, [code, agent])

  function newChat() {
    router.push(`/client/chat/${encodeURIComponent(code)}?agent=${agent}`)
    setSidebarOpen(false)
    // Force reload to clear conversation state
    window.location.href = `/client/chat/${encodeURIComponent(code)}?agent=${agent}`
  }

  function openConversation(id: string) {
    router.push(`/client/chat/${encodeURIComponent(code)}?agent=${agent}&conv=${id}`)
    setSidebarOpen(false)
    window.location.href = `/client/chat/${encodeURIComponent(code)}?agent=${agent}&conv=${id}`
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
            onClick={() => router.push("/client")}
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
                        await fetch(`/api/client/conversations?id=${conv.id}&code=${encodeURIComponent(code)}`, { method: "DELETE" })
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
