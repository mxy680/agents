"use client"

import React, { useState, useRef, useEffect, useCallback, use } from "react"
import ReactMarkdown from "react-markdown"
import { Button } from "@/components/ui/button"
import { IconSend, IconRobot, IconUser, IconChevronDown, IconChevronRight, IconTool, IconArrowLeft, IconLoader2 } from "@tabler/icons-react"
import { cn } from "@/lib/utils"

// --- Types ---

interface ToolCall {
  id: string
  name: string
  inputParts: string[]
  finalInput: string
  result?: string
}

type MessageRole = "user" | "assistant"

interface Message {
  id: string
  role: MessageRole
  text: string
  toolCalls: ToolCall[]
  isStreaming: boolean
}

// --- Tool call card component ---

function ToolCallCard({ tool }: { tool: ToolCall }) {
  const [expanded, setExpanded] = useState(false)
  const hasResult = Boolean(tool.result)

  return (
    <div className="my-1 border border-border/60 bg-muted/30 text-xs">
      <button
        onClick={() => setExpanded((v) => !v)}
        className="flex w-full items-center gap-1.5 px-2.5 py-1.5 text-left hover:bg-muted/60 transition-colors"
      >
        <IconTool className="size-3 shrink-0 text-muted-foreground" />
        <span className="font-mono font-medium">{tool.name}</span>
        {hasResult && (
          <span className="ml-1 text-muted-foreground">— done</span>
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

// --- Message bubble ---

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

function MarkdownContent({ content }: { content: string }) {
  return (
    <ReactMarkdown
      components={{
        p: ({ children }) => <p className="mb-2 last:mb-0">{children}</p>,
        strong: ({ children }) => <strong className="font-semibold">{children}</strong>,
        em: ({ children }) => <em className="italic">{children}</em>,
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
        blockquote: ({ children }) => (
          <blockquote className="my-2 border-l-2 border-border pl-3 text-muted-foreground italic">{children}</blockquote>
        ),
        hr: () => <hr className="my-3 border-border/60" />,
        a: ({ href, children }) => (
          <a href={href} target="_blank" rel="noopener noreferrer" className="text-primary underline underline-offset-2 hover:text-primary/80">
            {children}
          </a>
        ),
      }}
    >
      {content}
    </ReactMarkdown>
  )
}

function MessageBubble({ message }: { message: Message }) {
  const isUser = message.role === "user"
  const isThinking = message.isStreaming && message.text.length === 0

  return (
    <div className={cn("flex gap-2.5 mb-4", isUser && "flex-row-reverse")}>
      <div className={cn(
        "flex size-7 shrink-0 items-center justify-center mt-0.5",
        isUser ? "bg-primary text-primary-foreground" : "bg-muted"
      )}>
        {isUser ? <IconUser className="size-3.5" /> : <IconRobot className="size-3.5" />}
      </div>
      <div className={cn("flex flex-col gap-1 max-w-[80%]", isUser && "items-end")}>
        <div className={cn(
          "px-3 py-2 text-xs/relaxed",
          isUser
            ? "bg-primary text-primary-foreground"
            : "bg-muted/60 text-foreground border border-border/40"
        )}>
          {isThinking ? (
            <ThinkingIndicator />
          ) : isUser ? (
            message.text
          ) : (
            <MarkdownContent content={message.text} />
          )}
          {message.isStreaming && message.text.length > 0 && (
            <span className="inline-block w-1.5 h-3 ml-0.5 bg-current animate-pulse align-text-bottom" />
          )}
        </div>
        {message.toolCalls.length > 0 && (
          <div className="w-full max-w-full">
            {message.toolCalls.map((tool) => (
              <ToolCallCard key={tool.id} tool={tool} />
            ))}
          </div>
        )}
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
      { id: userMsgId, role: "user", text, toolCalls: [], isStreaming: false },
      { id: assistantMsgId, role: "assistant", text: "", toolCalls: [], isStreaming: true },
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

      // Tool tracking for the current assistant message
      const toolCallsMap = new Map<string, ToolCall>()
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
            // Blank line = event boundary — rejoin multi-line data with newlines
            handleSSEData(currentEventType, dataLines.join("\n"))
            dataLines = []
          }
        }
      }

      // Flush any remaining data
      if (dataLines.length > 0) {
        handleSSEData(currentEventType, dataLines.join("\n"))
      }

      function handleSSEData(eventType: string, data: string) {
        switch (eventType) {
          case "delta":
            setMessages((prev) =>
              prev.map((m) =>
                m.id === assistantMsgId
                  ? { ...m, text: m.text + data }
                  : m
              )
            )
            break

          case "tool_start": {
            try {
              const { name, id } = JSON.parse(data) as { name: string; id: string }
              activeToolId = id
              const tool: ToolCall = { id, name, inputParts: [], finalInput: "", result: undefined }
              toolCallsMap.set(id, tool)
              setMessages((prev) =>
                prev.map((m) =>
                  m.id === assistantMsgId
                    ? { ...m, toolCalls: [...m.toolCalls, tool] }
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
            // Last tool_input event from container-runner sends the accumulated full input
            // We only have one content_block_stop per tool use, but multiple input_json_delta
            // We track all incoming tool_input data - the last one from content_block_stop
            // contains the full accumulated input.
            setMessages((prev) =>
              prev.map((m) => {
                if (m.id !== assistantMsgId) return m
                const updatedTools = m.toolCalls.map((t) => {
                  if (t.id !== toolId) return t
                  return { ...t, finalInput: data }
                })
                return { ...m, toolCalls: updatedTools }
              })
            )
            break
          }

          case "tool_result": {
            try {
              const { summary } = JSON.parse(data) as { summary: string }
              // Match to the most recently started tool
              const lastToolId = activeToolId
              if (lastToolId) {
                setMessages((prev) =>
                  prev.map((m) => {
                    if (m.id !== assistantMsgId) return m
                    const updatedTools = m.toolCalls.map((t) => {
                      if (t.id !== lastToolId) return t
                      return { ...t, result: summary }
                    })
                    return { ...m, toolCalls: updatedTools }
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
            // Final result — mark streaming done
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

      // Mark done if not already
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
