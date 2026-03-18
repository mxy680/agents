"use client"

import React, { useState, useRef, useEffect, useCallback, use } from "react"
import ReactMarkdown from "react-markdown"
import remarkGfm from "remark-gfm"
import { Button } from "@/components/ui/button"
import { IconSend, IconRobot, IconUser, IconChevronDown, IconChevronRight, IconTool, IconArrowLeft, IconLoader2 } from "@tabler/icons-react"
import { cn } from "@/lib/utils"

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

type ContentBlock = TextBlock | ToolBlock

type MessageRole = "user" | "assistant"

interface Message {
  id: string
  role: MessageRole
  blocks: ContentBlock[]
  isStreaming: boolean
}

// --- Helper: get or create the last text block in a message ---

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

// --- Tool call card component ---

function ToolCallCard({ tool }: { tool: ToolBlock }) {
  const [expanded, setExpanded] = useState(false)
  const hasResult = Boolean(tool.result)

  return (
    <div className="my-1.5 border border-border/60 bg-muted/30 text-xs">
      <button
        onClick={() => setExpanded((v) => !v)}
        className="flex w-full items-center gap-1.5 px-2.5 py-1.5 text-left hover:bg-muted/60 transition-colors"
      >
        <IconTool className="size-3 shrink-0 text-muted-foreground" />
        <span className="font-mono font-medium">{tool.name}</span>
        {hasResult && (
          <span className="ml-1 text-muted-foreground">— done</span>
        )}
        {!hasResult && (
          <IconLoader2 className="ml-1 size-3 text-muted-foreground animate-spin" />
        )}
        <span className="ml-auto">
          {expanded ? (
            <IconChevronDown className="size-3 text-muted-foreground" />
          ) : (
            <IconChevronRight className="size-3 text-muted-foreground" />
          )}
        </span>
      </button>
      {expanded && (
        <div className="border-t border-border/60 px-2.5 py-2 space-y-2">
          {tool.finalInput && (
            <div>
              <p className="text-muted-foreground mb-1">Input</p>
              <pre className="whitespace-pre-wrap break-all font-mono text-xs bg-background/60 p-1.5 overflow-auto max-h-48">
                {(() => {
                  try {
                    return JSON.stringify(JSON.parse(tool.finalInput), null, 2)
                  } catch {
                    return tool.finalInput
                  }
                })()}
              </pre>
            </div>
          )}
          {tool.result && (
            <div>
              <p className="text-muted-foreground mb-1">Result</p>
              <pre className="whitespace-pre-wrap break-all font-mono text-xs bg-background/60 p-1.5 overflow-auto max-h-48">
                {(() => {
                  try {
                    return JSON.stringify(JSON.parse(tool.result), null, 2)
                  } catch {
                    return tool.result
                  }
                })()}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

// --- Thinking indicator ---

function ThinkingIndicator() {
  return (
    <div className="flex items-center gap-2 px-3 py-2 text-xs text-muted-foreground">
      <div className="flex gap-1">
        <span className="size-1.5 rounded-full bg-muted-foreground/60 animate-bounce" style={{ animationDelay: "0ms" }} />
        <span className="size-1.5 rounded-full bg-muted-foreground/60 animate-bounce" style={{ animationDelay: "150ms" }} />
        <span className="size-1.5 rounded-full bg-muted-foreground/60 animate-bounce" style={{ animationDelay: "300ms" }} />
      </div>
      <span className="italic">Thinking…</span>
    </div>
  )
}

// --- Markdown renderer ---

function MarkdownContent({ content }: { content: string }) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        p: ({ children }) => <p className="mb-2 last:mb-0">{children}</p>,
        strong: ({ children }) => <strong className="font-semibold">{children}</strong>,
        em: ({ children }) => <em className="italic">{children}</em>,
        del: ({ children }) => <del className="line-through text-muted-foreground">{children}</del>,
        ul: ({ children }) => <ul className="mb-2 ml-4 list-disc last:mb-0">{children}</ul>,
        ol: ({ children }) => <ol className="mb-2 ml-4 list-decimal last:mb-0">{children}</ol>,
        li: ({ children }) => <li className="mb-0.5">{children}</li>,
        code: ({ children, className }) => {
          const isBlock = className?.includes("language-")
          if (isBlock) {
            return (
              <pre className="my-2 overflow-auto rounded bg-background/80 p-2 text-[11px] font-mono">
                <code>{children}</code>
              </pre>
            )
          }
          return <code className="rounded bg-background/60 px-1 py-0.5 text-[11px] font-mono">{children}</code>
        },
        pre: ({ children }) => <>{children}</>,
        h1: ({ children }) => <h1 className="mb-2 text-base font-bold">{children}</h1>,
        h2: ({ children }) => <h2 className="mb-2 text-sm font-bold">{children}</h2>,
        h3: ({ children }) => <h3 className="mb-1 text-xs font-bold">{children}</h3>,
        h4: ({ children }) => <h4 className="mb-1 text-xs font-semibold">{children}</h4>,
        blockquote: ({ children }) => (
          <blockquote className="my-2 border-l-2 border-border pl-3 text-muted-foreground italic">{children}</blockquote>
        ),
        hr: () => <hr className="my-3 border-border/60" />,
        a: ({ href, children }) => (
          <a href={href} target="_blank" rel="noopener noreferrer" className="text-primary underline underline-offset-2 hover:text-primary/80">
            {children}
          </a>
        ),
        // Table support (GFM)
        table: ({ children }) => (
          <div className="my-2 overflow-x-auto">
            <table className="w-full border-collapse text-[11px]">{children}</table>
          </div>
        ),
        thead: ({ children }) => (
          <thead className="bg-background/60">{children}</thead>
        ),
        tbody: ({ children }) => <tbody>{children}</tbody>,
        tr: ({ children }) => (
          <tr className="border-b border-border/40">{children}</tr>
        ),
        th: ({ children }) => (
          <th className="px-2 py-1.5 text-left font-semibold border border-border/40">{children}</th>
        ),
        td: ({ children }) => (
          <td className="px-2 py-1.5 border border-border/40">{children}</td>
        ),
        // Task list support (GFM)
        input: ({ checked, ...props }) => (
          <input
            type="checkbox"
            checked={checked}
            readOnly
            className="mr-1.5 align-middle"
            {...props}
          />
        ),
      }}
    >
      {content}
    </ReactMarkdown>
  )
}

// --- Message bubble with ordered content blocks ---

function MessageBubble({ message }: { message: Message }) {
  const isUser = message.role === "user"
  const hasContent = message.blocks.length > 0
  const isThinking = message.isStreaming && !hasContent

  // Check if the last block is a streaming text block
  const lastBlock = message.blocks[message.blocks.length - 1]
  const isStreamingText = message.isStreaming && lastBlock?.type === "text"

  return (
    <div className={cn("flex gap-2.5 mb-4", isUser && "flex-row-reverse")}>
      <div className={cn(
        "flex size-7 shrink-0 items-center justify-center mt-0.5",
        isUser ? "bg-primary text-primary-foreground" : "bg-muted"
      )}>
        {isUser ? <IconUser className="size-3.5" /> : <IconRobot className="size-3.5" />}
      </div>
      <div className={cn("flex flex-col gap-0 max-w-[80%]", isUser && "items-end")}>
        {isThinking && (
          <div className="bg-muted/60 border border-border/40">
            <ThinkingIndicator />
          </div>
        )}

        {message.blocks.map((block, i) => {
          if (block.type === "text") {
            const isLast = i === message.blocks.length - 1
            return (
              <div
                key={i}
                className={cn(
                  "px-3 py-2 text-xs/relaxed",
                  isUser
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted/60 text-foreground border border-border/40"
                )}
              >
                {isUser ? block.content : <MarkdownContent content={block.content} />}
                {isLast && isStreamingText && (
                  <span className="inline-block w-1.5 h-3 ml-0.5 bg-current animate-pulse align-text-bottom" />
                )}
              </div>
            )
          }

          if (block.type === "tool") {
            return <ToolCallCard key={block.id} tool={block} />
          }

          return null
        })}
      </div>
    </div>
  )
}

// --- Main chat page ---

export default function ChatPage({ params }: { params: Promise<{ agentName: string }> }) {
  const { agentName } = use(params)

  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState("")
  const [isLoading, setIsLoading] = useState(false)
  const [sessionId, setSessionId] = useState<string | undefined>()
  const [error, setError] = useState<string | null>(null)

  const bottomRef = useRef<HTMLDivElement>(null)
  const abortRef = useRef<AbortController | null>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

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
      const res = await fetch(`/api/chat/${agentName}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ message: text, sessionId }),
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

      // SSE parser: accumulate data lines, dispatch on blank line (event boundary)
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
            // Append text to the last text block, or create a new one
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
              const toolBlock: ToolBlock = { type: "tool", id, name, finalInput: "", result: undefined }
              setMessages((prev) =>
                prev.map((m) =>
                  m.id === assistantMsgId
                    ? { ...m, blocks: [...m.blocks, toolBlock] }
                    : m
                )
              )
            } catch {
              // ignore parse error
            }
            break
          }

          case "tool_input": {
            if (!activeToolId) break
            const toolId = activeToolId
            setMessages((prev) =>
              prev.map((m) => {
                if (m.id !== assistantMsgId) return m
                const updatedBlocks = m.blocks.map((b) => {
                  if (b.type !== "tool" || b.id !== toolId) return b
                  return { ...b, finalInput: data }
                })
                return { ...m, blocks: updatedBlocks }
              })
            )
            break
          }

          case "tool_result": {
            try {
              const { summary } = JSON.parse(data) as { summary: string }
              if (activeToolId) {
                const toolId = activeToolId
                setMessages((prev) =>
                  prev.map((m) => {
                    if (m.id !== assistantMsgId) return m
                    const updatedBlocks = m.blocks.map((b) => {
                      if (b.type !== "tool" || b.id !== toolId) return b
                      return { ...b, result: summary }
                    })
                    return { ...m, blocks: updatedBlocks }
                  })
                )
                activeToolId = null
              }
            } catch {
              // ignore
            }
            break
          }

          case "session":
            setSessionId(data)
            break

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

      // Mark done
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
  }, [agentName, input, isLoading, sessionId])

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
    <div className="flex h-screen flex-col bg-background">
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
          <span className="text-sm font-medium capitalize">{agentName.replace(/-/g, " ")}</span>
        </div>
        {sessionId && (
          <>
            <div className="ml-auto h-4 w-px bg-border" />
            <span className="text-xs text-muted-foreground font-mono" title="Session ID">
              Session: {sessionId.slice(0, 8)}…
            </span>
          </>
        )}
      </header>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4">
        {messages.length === 0 && (
          <div className="flex flex-col items-center justify-center h-full gap-3 text-center">
            <div className="flex size-14 items-center justify-center bg-muted">
              <IconRobot className="size-7 text-muted-foreground" />
            </div>
            <div>
              <p className="text-sm font-medium capitalize">{agentName.replace(/-/g, " ")}</p>
              <p className="text-xs text-muted-foreground mt-0.5">Send a message to start a conversation.</p>
            </div>
          </div>
        )}

        {messages.map((message) => (
          <MessageBubble key={message.id} message={message} />
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
        <div className="flex items-end gap-2">
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
            style={{
              height: "auto",
              overflowY: "auto",
            }}
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
              onClick={sendMessage}
              disabled={!input.trim()}
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
