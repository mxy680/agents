"use client"

import React, { useState, useRef, useEffect, useCallback, use } from "react"
import { useRouter } from "next/navigation"
import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"
import { IconSend, IconLoader2, IconPaperclip, IconFile, IconX, IconDownload, IconPhoto } from "@tabler/icons-react"

interface TextBlock { type: "text"; content: string }
interface ToolBlock { type: "tool"; id: string; name: string; finalInput: string; result?: string }
interface FileBlock { type: "file"; name: string; url: string; size?: number; fileType?: string }
type ContentBlock = TextBlock | ToolBlock | FileBlock

interface Message {
  id: string
  role: "user" | "assistant"
  blocks: ContentBlock[]
  isStreaming: boolean
}

function appendText(blocks: ContentBlock[], text: string): ContentBlock[] {
  const updated = [...blocks]
  const last = updated[updated.length - 1]
  if (last?.type === "text") {
    updated[updated.length - 1] = { ...last, content: last.content + text }
  } else {
    updated.push({ type: "text", content: text })
  }
  return updated
}

export default function ClientChatPage({
  params,
}: {
  params: Promise<{ code: string }>
}) {
  const { code } = use(params)
  const router = useRouter()

  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [clientName, setClientName] = useState("")
  const [agentName, setAgentName] = useState("")
  const [agentDisplayName, setAgentDisplayName] = useState("")
  const [conversationId, setConversationId] = useState<string | null>(null)
  const [verified, setVerified] = useState(false)

  const bottomRef = useRef<HTMLDivElement>(null)
  const abortRef = useRef<AbortController | null>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  // Verify code and get agent info on mount
  useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search)
    const agent = searchParams.get("agent") ?? ""

    async function verify() {
      try {
        const res = await fetch("/api/client/verify", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ code: decodeURIComponent(code) }),
        })
        if (res.ok) {
          const data = await res.json() as {
            clientName: string
            agents: Array<{ name: string; displayName: string }>
          }
          setClientName(data.clientName)

          // Find the selected agent
          const selected = data.agents.find((a) => a.name === agent) ?? data.agents[0]
          if (selected) {
            setAgentName(selected.name)
            setAgentDisplayName(selected.displayName)
          }
          setVerified(true)
        } else {
          router.push("/client")
        }
      } catch {
        router.push("/client")
      }
    }
    verify()
  }, [code, router])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [messages])

  const sendMessage = useCallback(async () => {
    const text = input.trim()
    if (!text || isLoading) return

    setInput("")
    setError(null)

    const userMsgId = crypto.randomUUID()
    const assistantMsgId = crypto.randomUUID()

    setMessages((prev) => [
      ...prev,
      { id: userMsgId, role: "user", blocks: [{ type: "text", content: text }], isStreaming: false },
      { id: assistantMsgId, role: "assistant", blocks: [], isStreaming: true },
    ])

    setIsLoading(true)
    const abort = new AbortController()
    abortRef.current = abort

    try {
      const res = await fetch("/api/client/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          code: decodeURIComponent(code),
          message: text,
          conversationId,
          agentName,
        }),
        signal: abort.signal,
      })

      if (!res.ok || !res.body) {
        throw new Error(await res.text().catch(() => "Request failed"))
      }

      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let lineBuffer = ""
      let currentEventType = ""
      let dataLines: string[] = []

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        lineBuffer += decoder.decode(value, { stream: true })
        const lines = lineBuffer.split("\n")
        lineBuffer = lines.pop() ?? ""

        for (const line of lines) {
          if (line.startsWith("event:")) {
            currentEventType = line.slice(6).trim()
          } else if (line.startsWith("data: ")) {
            dataLines.push(line.slice(6))
          } else if (line.startsWith("data:")) {
            dataLines.push(line.slice(5))
          } else if (line === "" && dataLines.length > 0) {
            handleSSE(currentEventType, dataLines.join("\n"))
            dataLines = []
          }
        }
      }
      if (dataLines.length > 0) handleSSE(currentEventType, dataLines.join("\n"))

      function handleSSE(event: string, data: string) {
        switch (event) {
          case "delta":
            setMessages((prev) =>
              prev.map((m) => m.id === assistantMsgId ? { ...m, blocks: appendText(m.blocks, data) } : m)
            )
            break
          case "conversation":
            setConversationId(data)
            break
          case "tool_start":
            try {
              const { name, id } = JSON.parse(data)
              setMessages((prev) =>
                prev.map((m) => m.id === assistantMsgId
                  ? { ...m, blocks: [...m.blocks, { type: "tool", id, name, finalInput: "", result: undefined }] }
                  : m
                )
              )
            } catch {}
            break
          case "tool_result":
            try {
              const { summary, toolUseId } = JSON.parse(data)
              setMessages((prev) =>
                prev.map((m) => {
                  if (m.id !== assistantMsgId) return m
                  return {
                    ...m,
                    blocks: m.blocks.map((b) =>
                      b.type === "tool" && b.id === toolUseId ? { ...b, result: summary } : b
                    ),
                  }
                })
              )
            } catch {}
            break
          case "file":
            try {
              const fileData = JSON.parse(data)
              setMessages((prev) =>
                prev.map((m) => m.id === assistantMsgId
                  ? { ...m, blocks: [...m.blocks, { type: "file", ...fileData }] }
                  : m
                )
              )
            } catch {}
            break
          case "result":
          case "error":
            if (event === "error") setError(data)
            setMessages((prev) =>
              prev.map((m) => m.id === assistantMsgId ? { ...m, isStreaming: false } : m)
            )
            break
        }
      }

      setMessages((prev) =>
        prev.map((m) => m.id === assistantMsgId ? { ...m, isStreaming: false } : m)
      )
    } catch (e) {
      if ((e as Error).name === "AbortError") return
      setError(e instanceof Error ? e.message : String(e))
      setMessages((prev) =>
        prev.map((m) => m.id === assistantMsgId ? { ...m, isStreaming: false } : m)
      )
    } finally {
      setIsLoading(false)
      abortRef.current = null
    }
  }, [code, input, isLoading, conversationId])

  if (!verified) {
    return (
      <div className="min-h-screen bg-[#0a0a0a] flex items-center justify-center">
        <IconLoader2 className="size-6 text-neutral-500 animate-spin" />
      </div>
    )
  }

  return (
    <div className="flex h-dvh flex-col bg-[#0a0a0a] text-white">
      {/* Header */}
      <header className="flex h-14 shrink-0 items-center justify-between border-b border-neutral-800 px-4">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">{agentDisplayName || "Emdash"}</span>
        </div>
        <span className="text-xs text-neutral-500">
          {clientName}
        </span>
      </header>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.length === 0 && (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <p className="text-lg font-medium text-neutral-300">Hi {clientName}</p>
            <p className="text-sm text-neutral-500 mt-1">Ask me anything about NYC real estate.</p>
          </div>
        )}

        {messages.map((msg) => (
          <div key={msg.id} className={`flex ${msg.role === "user" ? "justify-end" : "justify-start"}`}>
            <div className={`max-w-[80%] ${msg.role === "user" ? "bg-blue-600" : "bg-neutral-800"} rounded-lg px-3 py-2`}>
              {msg.blocks.map((block, i) => {
                if (block.type === "text") {
                  return (
                    <div key={i} className="text-sm prose prose-invert prose-sm max-w-none [&>*:first-child]:mt-0 [&>*:last-child]:mb-0">
                      {msg.role === "user" ? (
                        <p className="whitespace-pre-wrap m-0">{block.content}</p>
                      ) : (
                        <ReactMarkdown remarkPlugins={[remarkGfm]}>
                          {block.content || ""}
                        </ReactMarkdown>
                      )}
                      {msg.isStreaming && i === msg.blocks.length - 1 && (
                        <span className="inline-block w-1.5 h-3 ml-0.5 bg-current animate-pulse" />
                      )}
                    </div>
                  )
                }
                if (block.type === "tool") {
                  return (
                    <div key={i} className="text-xs text-neutral-400 border border-neutral-700 rounded px-2 py-1 my-1">
                      {block.result ? `${block.name} — done` : `Running ${block.name}...`}
                    </div>
                  )
                }
                if (block.type === "file") {
                  return (
                    <a key={i} href={block.url} target="_blank" rel="noopener noreferrer" download={block.name}
                      className="flex items-center gap-2 border border-neutral-700 rounded px-2 py-1.5 my-1 hover:bg-neutral-700/50 transition-colors">
                      <IconFile className="size-4 text-neutral-400 shrink-0" />
                      <span className="text-xs truncate flex-1">{block.name}</span>
                      <IconDownload className="size-3.5 text-neutral-500 shrink-0" />
                    </a>
                  )
                }
                return null
              })}
              {msg.isStreaming && msg.blocks.length === 0 && (
                <div className="flex items-center gap-2 text-xs text-neutral-400">
                  <IconLoader2 className="size-3 animate-spin" />
                  <span>Thinking...</span>
                </div>
              )}
            </div>
          </div>
        ))}

        {error && (
          <div className="text-xs text-red-400 bg-red-900/20 border border-red-800 rounded px-3 py-2">
            {error}
          </div>
        )}

        <div ref={bottomRef} />
      </div>

      {/* Input */}
      <div className="shrink-0 border-t border-neutral-800 p-4">
        <div className="flex items-end gap-2 max-w-3xl mx-auto">
          <textarea
            ref={textareaRef}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); sendMessage() }
            }}
            placeholder="Ask me anything..."
            rows={1}
            disabled={isLoading}
            className="flex-1 resize-none bg-neutral-900 border border-neutral-800 rounded-lg px-3 py-2.5 text-sm text-white placeholder:text-neutral-600 focus:outline-none focus:border-neutral-600 disabled:opacity-50 min-h-[40px] max-h-[200px]"
            style={{ height: "auto", overflowY: "auto" }}
            onInput={(e) => {
              const t = e.target as HTMLTextAreaElement
              t.style.height = "auto"
              t.style.height = Math.min(t.scrollHeight, 200) + "px"
            }}
          />
          {isLoading ? (
            <button onClick={() => abortRef.current?.abort()} className="p-2.5 bg-neutral-800 rounded-lg hover:bg-neutral-700">
              <IconLoader2 className="size-4 text-neutral-400 animate-spin" />
            </button>
          ) : (
            <button onClick={sendMessage} disabled={!input.trim()} className="p-2.5 bg-white text-black rounded-lg hover:bg-neutral-200 disabled:opacity-30 disabled:cursor-not-allowed">
              <IconSend className="size-4" />
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
