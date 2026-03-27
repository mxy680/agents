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
  const [pendingFiles, setPendingFiles] = useState<Array<{ file: File; url?: string; uploading: boolean; preview?: string }>>([])

  const bottomRef = useRef<HTMLDivElement>(null)
  const abortRef = useRef<AbortController | null>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  async function handleFileSelect(files: FileList | null) {
    if (!files) return
    const newFiles = Array.from(files).map((f) => ({
      file: f, uploading: true,
      preview: f.type.startsWith("image/") ? URL.createObjectURL(f) : undefined,
    }))
    setPendingFiles((prev) => [...prev, ...newFiles])
    for (const pf of newFiles) {
      try {
        const fd = new FormData()
        fd.append("file", pf.file)
        fd.append("conversationId", conversationId ?? "client")
        const res = await fetch("/api/chat/upload", { method: "POST", body: fd })
        if (res.ok) {
          const data = await res.json() as { url: string }
          setPendingFiles((prev) => prev.map((f) => f.file === pf.file ? { ...f, url: data.url, uploading: false } : f))
        } else {
          setPendingFiles((prev) => prev.filter((f) => f.file !== pf.file))
        }
      } catch {
        setPendingFiles((prev) => prev.filter((f) => f.file !== pf.file))
      }
    }
  }

  // Verify code and get agent info on mount
  useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search)
    const agent = searchParams.get("agent") ?? ""
    const conv = searchParams.get("conv") ?? ""

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

          const selected = data.agents.find((a) => a.name === agent) ?? data.agents[0]
          if (selected) {
            setAgentName(selected.name)
            setAgentDisplayName(selected.displayName)
          }

          // Load existing conversation if conv param provided
          if (conv) {
            setConversationId(conv)
            try {
              const convRes = await fetch(`/api/client/conversations/${conv}?code=${encodeURIComponent(decodeURIComponent(code))}`)
              if (convRes.ok) {
                const convData = await convRes.json() as { messages: Array<{ id: string; role: "user" | "assistant"; blocks: ContentBlock[] }> }
                setMessages(
                  (convData.messages ?? []).map((m) => ({
                    id: m.id,
                    role: m.role,
                    blocks: m.blocks,
                    isStreaming: false,
                  }))
                )
              }
            } catch {}
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
    if (!text && pendingFiles.length === 0) return
    if (isLoading) return

    // Build message with file references
    const uploadedFiles = pendingFiles.filter((f) => f.url && !f.uploading)
    let messageText = text
    if (uploadedFiles.length > 0) {
      const fileList = uploadedFiles.map((f) => `[Attached file: ${f.file.name} (${f.url})]`).join("\n")
      messageText = fileList + (text ? "\n\n" + text : "")
    }

    setInput("")
    setError(null)
    setPendingFiles([])

    const userMsgId = crypto.randomUUID()
    const assistantMsgId = crypto.randomUUID()

    const userBlocks: ContentBlock[] = []
    for (const f of uploadedFiles) {
      userBlocks.push({ type: "file", name: f.file.name, url: f.url!, size: f.file.size, fileType: f.file.type })
    }
    if (text) userBlocks.push({ type: "text", content: text })

    setMessages((prev) => [
      ...prev,
      { id: userMsgId, role: "user", blocks: userBlocks, isStreaming: false },
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
          message: messageText || text,
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
            // Update URL so sidebar can highlight this conversation
            const url = new URL(window.location.href)
            url.searchParams.set("conv", data)
            window.history.replaceState({}, "", url.toString())
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

  // Split blocks into separate bubbles — each file block and text block is its own bubble
  function renderMessage(msg: Message) {
    const bubbles: React.ReactNode[] = []
    const isUser = msg.role === "user"
    const align = isUser ? "justify-end" : "justify-start"
    const bg = isUser ? "bg-blue-600" : "bg-neutral-800"

    // Group consecutive text/tool blocks into one bubble, file blocks get their own
    let textGroup: ContentBlock[] = []

    function flushTextGroup() {
      if (textGroup.length === 0) return
      const blocks = [...textGroup]
      textGroup = []
      bubbles.push(
        <div key={`text-${bubbles.length}`} className={`flex ${align}`}>
          <div className={`max-w-[80%] ${bg} rounded-lg px-3 py-2`}>
            {blocks.map((block, i) => {
              if (block.type === "text") {
                return (
                  <div key={i} className={isUser ? "text-sm" : "text-sm prose prose-invert prose-sm max-w-none prose-p:my-1.5 prose-headings:text-white prose-headings:mt-3 prose-headings:mb-1.5 prose-strong:text-white prose-a:text-blue-400 prose-a:no-underline hover:prose-a:underline prose-code:bg-black/30 prose-code:px-1 prose-code:py-0.5 prose-code:rounded prose-code:text-xs prose-code:before:content-none prose-code:after:content-none prose-pre:bg-black/30 prose-pre:rounded prose-pre:text-xs prose-table:text-xs prose-th:border prose-th:border-neutral-600 prose-th:px-2 prose-th:py-1.5 prose-th:bg-neutral-700/50 prose-td:border prose-td:border-neutral-700 prose-td:px-2 prose-td:py-1.5 prose-hr:border-neutral-600 prose-blockquote:border-neutral-500 prose-li:my-0.5"}>
                    {isUser ? (
                      <p className="whitespace-pre-wrap m-0">{block.content}</p>
                    ) : (
                      <ReactMarkdown remarkPlugins={[remarkGfm]}>
                        {block.content || ""}
                      </ReactMarkdown>
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
              return null
            })}
            {msg.isStreaming && blocks[blocks.length - 1]?.type === "text" && (
              <span className="inline-block w-1.5 h-3 ml-0.5 bg-current animate-pulse" />
            )}
          </div>
        </div>
      )
    }

    for (const block of msg.blocks) {
      if (block.type === "file") {
        flushTextGroup()
        const isImage = block.fileType?.startsWith("image/")
        bubbles.push(
          <div key={`file-${bubbles.length}`} className={`flex ${align}`}>
            <div className="max-w-[80%]">
              {isImage && block.url && (
                <img src={block.url} alt={block.name} className="max-w-xs rounded-lg mb-1" />
              )}
              <a
                href={block.url}
                target="_blank"
                rel="noopener noreferrer"
                download={block.name}
                className={`flex items-center gap-2.5 ${bg} rounded-lg px-3 py-2.5 hover:opacity-80 transition-opacity`}
              >
                <IconFile className="size-5 text-neutral-400 shrink-0" />
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-white truncate">{block.name}</p>
                  {block.size && (
                    <p className="text-xs text-neutral-400">
                      {block.size > 1024 * 1024
                        ? `${(block.size / (1024 * 1024)).toFixed(1)} MB`
                        : `${Math.round(block.size / 1024)} KB`}
                    </p>
                  )}
                </div>
                <IconDownload className="size-4 text-neutral-500 shrink-0" />
              </a>
            </div>
          </div>
        )
      } else {
        textGroup.push(block)
      }
    }
    flushTextGroup()

    // Thinking indicator
    if (msg.isStreaming && msg.blocks.length === 0) {
      bubbles.push(
        <div key="thinking" className={`flex ${align}`}>
          <div className={`${bg} rounded-lg px-3 py-2`}>
            <div className="flex items-center gap-2 text-xs text-neutral-400">
              <IconLoader2 className="size-3 animate-spin" />
              <span>Thinking...</span>
            </div>
          </div>
        </div>
      )
    }

    return bubbles
  }

  return (
    <div className="flex h-full flex-col bg-[#0a0a0a] text-white">
      {/* Header */}
      <header className="flex h-14 shrink-0 items-center justify-between border-b border-neutral-800 px-4 pl-12 md:pl-4">
        <span className="text-sm font-medium">{agentDisplayName || "Agent"}</span>
        <span className="text-xs text-neutral-500">{clientName}</span>
      </header>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-3">
        {messages.length === 0 && (
          <div className="flex flex-col items-center justify-center h-full text-center">
            <p className="text-lg font-medium text-neutral-300">Hi {clientName}</p>
            <p className="text-sm text-neutral-500 mt-1">How can I help you today?</p>
          </div>
        )}

        {messages.map((msg) => (
          <React.Fragment key={msg.id}>
            {renderMessage(msg)}
          </React.Fragment>
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
        {pendingFiles.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-2 max-w-3xl mx-auto">
            {pendingFiles.map((pf, i) => (
              <div key={i} className="flex items-center gap-1.5 bg-neutral-800 border border-neutral-700 rounded-lg px-2 py-1 text-xs text-neutral-300">
                {pf.preview ? (
                  <img src={pf.preview} alt={pf.file.name} className="size-5 rounded object-cover" />
                ) : (
                  <IconFile className="size-3.5 text-neutral-500" />
                )}
                <span className="max-w-[100px] truncate">{pf.file.name}</span>
                {pf.uploading && <IconLoader2 className="size-3 animate-spin text-neutral-500" />}
                <button onClick={() => setPendingFiles((prev) => prev.filter((f) => f.file !== pf.file))} className="text-neutral-500 hover:text-white">
                  <IconX className="size-3" />
                </button>
              </div>
            ))}
          </div>
        )}
        <div className="flex items-end gap-2 max-w-3xl mx-auto">
          <input ref={fileInputRef} type="file" multiple className="hidden" onChange={(e) => handleFileSelect(e.target.files)} />
          <button
            onClick={() => fileInputRef.current?.click()}
            disabled={isLoading}
            className="p-2.5 text-neutral-500 hover:text-white transition-colors disabled:opacity-30 shrink-0"
            title="Attach file"
          >
            <IconPaperclip className="size-4" />
          </button>
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
            <button onClick={() => abortRef.current?.abort()} className="p-2.5 bg-neutral-800 rounded-lg hover:bg-neutral-700 shrink-0">
              <IconLoader2 className="size-4 text-neutral-400 animate-spin" />
            </button>
          ) : (
            <button onClick={sendMessage} disabled={!input.trim() && pendingFiles.length === 0} className="p-2.5 bg-white text-black rounded-lg hover:bg-neutral-200 disabled:opacity-30 disabled:cursor-not-allowed shrink-0">
              <IconSend className="size-4" />
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
