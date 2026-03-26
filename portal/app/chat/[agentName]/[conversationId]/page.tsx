"use client"

import React, { useState, useRef, useEffect, useCallback, use } from "react"
import { Button } from "@/components/ui/button"
import {
  IconSend,
  IconRobot,
  IconArrowLeft,
  IconLoader2,
  IconPaperclip,
  IconX,
  IconFile,
  IconDownload,
  IconPhoto,
} from "@tabler/icons-react"
import { cn } from "@/lib/utils"
import { MessageBubble } from "@/components/chat/message-bubble"
import { PromptPalette } from "@/components/chat/prompt-palette"

// --- Types ---

interface TextBlock {
  type: "text"
  content: string
}

interface ToolBlock {
  type: "tool"
  id: string
  name: string
  finalInput: string
  result?: string
}

interface FileBlock {
  type: "file"
  name: string
  url: string
  size?: number
  fileType?: string
}

type ContentBlock = TextBlock | ToolBlock | FileBlock

interface PendingFile {
  file: File
  name: string
  url?: string
  uploading: boolean
  preview?: string
}

type MessageRole = "user" | "assistant"

interface Message {
  id: string
  role: MessageRole
  blocks: ContentBlock[]
  isStreaming: boolean
}

// --- Helper: append text to the last text block ---

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

// Inline ToolCallCard, ThinkingIndicator, MarkdownContent, and MessageBubble
// are now in portal/components/chat/




// --- DB message type for hydration ---

interface DBMessage {
  id: string
  conversation_id: string
  role: "user" | "assistant"
  blocks: ContentBlock[]
  created_at: string
}

// --- Main chat page ---

export default function ConversationPage({
  params,
}: {
  params: Promise<{ agentName: string; conversationId: string }>
}) {
  const { agentName, conversationId } = use(params)

  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [hydrated, setHydrated] = useState(false)
  const [pendingFiles, setPendingFiles] = useState<PendingFile[]>([])

  const bottomRef = useRef<HTMLDivElement>(null)
  const abortRef = useRef<AbortController | null>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  async function handleFileSelect(files: FileList | null) {
    if (!files) return
    const newFiles: PendingFile[] = Array.from(files).map((f) => ({
      file: f,
      name: f.name,
      uploading: true,
      preview: f.type.startsWith("image/") ? URL.createObjectURL(f) : undefined,
    }))
    setPendingFiles((prev) => [...prev, ...newFiles])

    // Upload each file
    for (const pf of newFiles) {
      try {
        const formData = new FormData()
        formData.append("file", pf.file)
        formData.append("conversationId", conversationId)
        const res = await fetch("/api/chat/upload", { method: "POST", body: formData })
        if (res.ok) {
          const data = await res.json() as { url: string; name: string }
          setPendingFiles((prev) =>
            prev.map((f) => f.file === pf.file ? { ...f, url: data.url, uploading: false } : f)
          )
        } else {
          setPendingFiles((prev) => prev.filter((f) => f.file !== pf.file))
        }
      } catch {
        setPendingFiles((prev) => prev.filter((f) => f.file !== pf.file))
      }
    }
  }

  function removePendingFile(file: File) {
    setPendingFiles((prev) => prev.filter((f) => f.file !== file))
  }

  // Load existing messages on mount
  useEffect(() => {
    async function loadMessages() {
      try {
        const res = await fetch(`/api/conversations/${conversationId}`)
        if (!res.ok) return
        const json = await res.json()
        const dbMessages: DBMessage[] = json.messages ?? []
        setMessages(
          dbMessages.map((m) => ({
            id: m.id,
            role: m.role,
            blocks: m.blocks,
            isStreaming: false,
          }))
        )
      } catch {
        // ignore — start with empty messages
      } finally {
        setHydrated(true)
      }
    }
    loadMessages()
  }, [conversationId])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [messages])

  const sendMessage = useCallback(async (overrideText?: string) => {
    const text = (overrideText ?? input).trim()
    if (!text || isLoading) return

    setInput("")
    setError(null)

    const userMsgId = crypto.randomUUID()
    const assistantMsgId = crypto.randomUUID()

    // Build user blocks: text + any attached files
    const userBlocks: ContentBlock[] = []
    const uploadedFiles = pendingFiles.filter((f) => f.url && !f.uploading)
    for (const f of uploadedFiles) {
      userBlocks.push({ type: "file", name: f.name, url: f.url!, size: f.file.size, fileType: f.file.type })
    }
    userBlocks.push({ type: "text", content: text })

    // Include file info in the message text so the agent knows about them
    let messageWithFiles = text
    if (uploadedFiles.length > 0) {
      const fileList = uploadedFiles.map((f) => `[Attached file: ${f.name} (${f.url})]`).join("\n")
      messageWithFiles = `${fileList}\n\n${text}`
    }

    setPendingFiles([])

    setMessages((prev) => [
      ...prev,
      {
        id: userMsgId,
        role: "user",
        blocks: userBlocks,
        isStreaming: false,
      },
      { id: assistantMsgId, role: "assistant", blocks: [], isStreaming: true },
    ])

    setIsLoading(true)
    const abort = new AbortController()
    abortRef.current = abort

    try {
      const res = await fetch(`/api/chat/${agentName}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ message: messageWithFiles, conversationId }),
        signal: abort.signal,
      })

      if (!res.ok || !res.body) {
        const errText = await res.text().catch(() => "Request failed")
        throw new Error(errText)
      }

      const reader = res.body.getReader()
      const decoder = new TextDecoder()

      let lineBuffer = ""
      let currentEventType = ""
      let activeToolId: string | null = null
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
            handleSSEData(currentEventType, dataLines.join("\n"))
            dataLines = []
          }
        }
      }

      if (dataLines.length > 0) {
        handleSSEData(currentEventType, dataLines.join("\n"))
      }

      function handleSSEData(eventType: string, data: string) {
        switch (eventType) {
          case "delta":
            setMessages((prev) =>
              prev.map((m) =>
                m.id === assistantMsgId
                  ? { ...m, blocks: appendText(m.blocks, data) }
                  : m
              )
            )
            break

          case "tool_start": {
            try {
              const { name, id } = JSON.parse(data) as { name: string; id: string }
              activeToolId = id
              const toolBlock: ToolBlock = {
                type: "tool",
                id,
                name,
                finalInput: "",
                result: undefined,
              }
              setMessages((prev) =>
                prev.map((m) =>
                  m.id === assistantMsgId
                    ? { ...m, blocks: [...m.blocks, toolBlock] }
                    : m
                )
              )
            } catch {
              // ignore
            }
            break
          }

          case "tool_input": {
            if (!activeToolId) break
            const toolId = activeToolId
            setMessages((prev) =>
              prev.map((m) => {
                if (m.id !== assistantMsgId) return m
                return {
                  ...m,
                  blocks: m.blocks.map((b) => {
                    if (b.type !== "tool" || b.id !== toolId) return b
                    return { ...b, finalInput: data }
                  }),
                }
              })
            )
            break
          }

          case "tool_result": {
            try {
              const parsed = JSON.parse(data) as {
                summary: string
                toolUseId?: string
              }
              const toolId = parsed.toolUseId ?? activeToolId
              if (toolId) {
                setMessages((prev) =>
                  prev.map((m) => {
                    if (m.id !== assistantMsgId) return m
                    return {
                      ...m,
                      blocks: m.blocks.map((b) => {
                        if (b.type !== "tool" || b.id !== toolId) return b
                        const resultStr =
                          typeof parsed.summary === "string"
                            ? parsed.summary
                            : JSON.stringify(parsed.summary)
                        return { ...b, result: resultStr }
                      }),
                    }
                  })
                )
                if (toolId === activeToolId) activeToolId = null
              }
            } catch {
              // ignore
            }
            break
          }

          case "result":
            setMessages((prev) =>
              prev.map((m) =>
                m.id === assistantMsgId ? { ...m, isStreaming: false } : m
              )
            )
            break

          case "error":
            setError(data)
            setMessages((prev) =>
              prev.map((m) =>
                m.id === assistantMsgId ? { ...m, isStreaming: false } : m
              )
            )
            break
        }
      }

      setMessages((prev) =>
        prev.map((m) =>
          m.id === assistantMsgId ? { ...m, isStreaming: false } : m
        )
      )
    } catch (e) {
      if ((e as Error).name === "AbortError") return
      const msg = e instanceof Error ? e.message : String(e)
      setError(msg)
      setMessages((prev) =>
        prev.map((m) =>
          m.id === assistantMsgId ? { ...m, isStreaming: false } : m
        )
      )
    } finally {
      setIsLoading(false)
      abortRef.current = null
    }
  }, [agentName, conversationId, input, isLoading])

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  const stopGeneration = () => {
    abortRef.current?.abort()
    setIsLoading(false)
  }

  return (
    <div className="flex h-full flex-col bg-background">
      {/* Header */}
      <header className="flex h-14 shrink-0 items-center gap-3 border-b px-4">
        <a
          href="/agents"
          className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors"
        >
          <IconArrowLeft className="size-3.5" />
          Agents
        </a>
        <div className="h-4 w-px bg-border" />
        <div className="flex items-center gap-2">
          <div className="flex size-6 items-center justify-center bg-muted">
            <IconRobot className="size-3.5" />
          </div>
          <span className="text-sm font-medium capitalize">
            {agentName.replace(/-/g, " ")}
          </span>
        </div>
      </header>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4">
        {hydrated && messages.length === 0 && (
          <PromptPalette
            agentName={agentName}
            agentDisplayName={agentName.replace(/-/g, " ").replace(/\b\w/g, (c) => c.toUpperCase())}
            onSelect={(prompt) => sendMessage(prompt)}
          />
        )}

        {messages.map((message) => (
          <MessageBubble
            key={message.id}
            role={message.role}
            blocks={message.blocks}
            isStreaming={message.isStreaming}
          />
        ))}

        {error && (
          <div className="mb-4 border border-destructive/50 bg-destructive/10 px-3 py-2 text-xs text-destructive">
            {error}
          </div>
        )}

        <div ref={bottomRef} />
      </div>

      {/* Input area */}
      <div className="shrink-0 border-t bg-background p-4">
        {/* Pending file previews */}
        {pendingFiles.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-2">
            {pendingFiles.map((pf, i) => (
              <div key={i} className="flex items-center gap-1.5 bg-muted/60 border border-border rounded px-2 py-1 text-xs">
                {pf.preview ? (
                  <img src={pf.preview} alt={pf.name} className="size-6 rounded object-cover" />
                ) : (
                  <IconFile className="size-4 text-muted-foreground" />
                )}
                <span className="max-w-[120px] truncate">{pf.name}</span>
                {pf.uploading && <IconLoader2 className="size-3 animate-spin text-muted-foreground" />}
                <button onClick={() => removePendingFile(pf.file)} className="text-muted-foreground hover:text-foreground">
                  <IconX className="size-3" />
                </button>
              </div>
            ))}
          </div>
        )}
        <div className="flex items-end gap-2">
          <input
            ref={fileInputRef}
            type="file"
            multiple
            className="hidden"
            onChange={(e) => handleFileSelect(e.target.files)}
          />
          <Button
            size="icon"
            variant="ghost"
            onClick={() => fileInputRef.current?.click()}
            disabled={isLoading}
            title="Attach file"
            className="shrink-0"
          >
            <IconPaperclip className="size-4" />
          </Button>
          <textarea
            ref={textareaRef}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Message the agent… (Enter to send, Shift+Enter for newline)"
            rows={1}
            disabled={isLoading}
            className={cn(
              "flex-1 resize-none bg-muted/40 border border-border px-3 py-2 text-xs outline-none",
              "placeholder:text-muted-foreground focus:border-ring focus:ring-1 focus:ring-ring/50",
              "disabled:opacity-50 min-h-[36px] max-h-[200px]"
            )}
            style={{ height: "auto", overflowY: "auto" }}
            onInput={(e) => {
              const target = e.target as HTMLTextAreaElement
              target.style.height = "auto"
              target.style.height = Math.min(target.scrollHeight, 200) + "px"
            }}
          />
          {isLoading ? (
            <Button
              size="icon"
              variant="outline"
              onClick={stopGeneration}
              title="Stop generation"
            >
              <IconLoader2 className="size-4 animate-spin" />
            </Button>
          ) : (
            <Button
              size="icon"
              onClick={() => sendMessage()}
              disabled={!input.trim() && pendingFiles.length === 0}
              title="Send message"
            >
              <IconSend className="size-4" />
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}
