import { useCallback, useRef, useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Send,
  Bot,
  User,
  FolderGit2,
  Loader2,
  Square,
  Crown,
  Cpu,
  GitBranch,
  Circle,
  X,
  Trash2,
  ChevronRight,
  ChevronDown,
  Brain,
  Terminal,
  FileText,
  Pencil,
  Search,
  Eye
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Markdown } from '@/components/Markdown'

// --- Types ---

interface Project {
  readonly id: string
  readonly name: string
  readonly path: string
  readonly githubUrl: string | null
  readonly createdAt: string
}

type MessageRole = 'user' | 'assistant' | 'system' | 'thinking' | 'tool-use' | 'tool-result' | 'ask-user'

interface QuestionOption {
  readonly label: string
  readonly description?: string
}

interface Question {
  readonly header?: string
  readonly question: string
  readonly options: readonly QuestionOption[]
  readonly multiSelect?: boolean
}

interface Message {
  readonly id: string
  readonly role: MessageRole
  content: string
  readonly toolName?: string
  readonly toolInput?: Record<string, unknown>
  readonly questions?: readonly Question[]
  answered?: boolean
  readonly timestamp: Date
}

type MinionStatus = 'spawning' | 'working' | 'done' | 'error'

interface Minion {
  readonly id: string
  readonly name: string
  readonly task: string
  readonly branch: string
  readonly worktreePath: string
  status: MinionStatus
  logs: Message[]
}

interface MinionSummary {
  readonly id: string
  readonly name: string
  readonly branch: string
  readonly status: MinionStatus
}

interface ProjectChatProps {
  project: Project
  hidden?: boolean
  activePanel: string
  onStatusChange?: (status: {
    orchestratorStatus: 'idle' | 'working'
    minionCount: number
    activeMinionCount: number
    minions: readonly MinionSummary[]
    activePanel: string
  }) => void
}

// --- Spawn command parser ---

const SPAWN_REGEX = /\[SPAWN_MINION\s+name="([^"]+)"\s+task="([^"]+)"\]/g
const MESSAGE_REGEX = /\[MESSAGE_MINION\s+name="([^"]+)"\s+message="([^"]+)"\]/g

function parseSpawnCommands(text: string): Array<{ name: string; task: string }> {
  const commands: Array<{ name: string; task: string }> = []
  let match: RegExpExecArray | null
  while ((match = SPAWN_REGEX.exec(text)) !== null) {
    commands.push({ name: match[1], task: match[2] })
  }
  SPAWN_REGEX.lastIndex = 0
  return commands
}

function parseMessageCommands(text: string): Array<{ name: string; message: string }> {
  const commands: Array<{ name: string; message: string }> = []
  let match: RegExpExecArray | null
  while ((match = MESSAGE_REGEX.exec(text)) !== null) {
    commands.push({ name: match[1], message: match[2] })
  }
  MESSAGE_REGEX.lastIndex = 0
  return commands
}

// --- Main Component ---

export function ProjectChat({ project, hidden, activePanel, onStatusChange }: ProjectChatProps): React.JSX.Element {
  const [minions, setMinions] = useState<Minion[]>([])
  const [loaded, setLoaded] = useState(false)

  // processedCommands ref is below, near handleCommandsFromText

  // Orchestrator chat state
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [streaming, setStreaming] = useState(false)
  const [waiting, setWaiting] = useState(false)
  const [sessionId, setSessionId] = useState<string | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)
  const currentAssistantId = useRef<string | null>(null)
  const claudeSessionIdRef = useRef<string | null>(null)

  // Load saved session on mount
  useEffect(() => {
    window.api.session.load({ projectId: project.id }).then((result) => {
      if (result.data) {
        const data = result.data as {
          messages?: Array<{ id: string; role: MessageRole; content: string; timestamp: string }>
          minions?: Array<{ id: string; name: string; task: string; branch: string; worktreePath: string; projectPath: string; status: string; claudeSessionId: string | null; logs: Array<{ id: string; role: 'assistant' | 'system'; content: string; timestamp: string }> }>
          claudeSessionId?: string | null
        }
        if (data.messages) {
          setMessages(data.messages.map((m) => {
            if (m.role === 'tool-use') {
              try {
                const parsed = JSON.parse(m.content)
                return { ...m, content: '', toolName: parsed.toolName, toolInput: parsed.toolInput, timestamp: new Date(m.timestamp) }
              } catch {
                return { ...m, timestamp: new Date(m.timestamp) }
              }
            }
            if (m.role === 'ask-user') {
              try {
                const parsed = JSON.parse(m.content)
                return { ...m, content: '', questions: parsed.questions, answered: parsed.answered ?? true, timestamp: new Date(m.timestamp) }
              } catch {
                return { ...m, timestamp: new Date(m.timestamp) }
              }
            }
            return { ...m, timestamp: new Date(m.timestamp) }
          }))
        }
        if (data.minions) {
          setMinions(data.minions.map((m) => ({
            ...m,
            status: m.status as MinionStatus,
            logs: m.logs.map((l) => ({ ...l, timestamp: new Date(l.timestamp) }))
          })))
        }
        if (data.claudeSessionId) {
          claudeSessionIdRef.current = data.claudeSessionId
        }
      }
      setLoaded(true)
    })
  }, [project.id])

  // Save session when messages or minions change (debounced)
  useEffect(() => {
    if (!loaded) return
    const timer = setTimeout(() => {
      window.api.session.save({
        projectId: project.id,
        data: {
          messages: messages.map((m) => ({
            id: m.id,
            role: m.role,
            content: m.role === 'tool-use'
              ? JSON.stringify({ toolName: m.toolName, toolInput: m.toolInput })
              : m.role === 'ask-user'
                ? JSON.stringify({ questions: m.questions, answered: m.answered })
                : m.content,
            timestamp: m.timestamp instanceof Date ? m.timestamp.toISOString() : m.timestamp
          })),
          minions: minions.map((m) => ({
            id: m.id, name: m.name, task: m.task, branch: m.branch,
            worktreePath: m.worktreePath, projectPath: project.path,
            status: m.status, claudeSessionId: null,
            logs: m.logs.map((l) => ({ ...l, timestamp: l.timestamp instanceof Date ? l.timestamp.toISOString() : l.timestamp }))
          })),
          claudeSessionId: claudeSessionIdRef.current
        }
      })
    }, 1000)
    return () => clearTimeout(timer)
  }, [messages, minions, loaded, project.id, project.path])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  useEffect(() => {
    if (activePanel === 'orchestrator') inputRef.current?.focus()
  }, [activePanel])

  // Report status changes to parent (use ref to avoid re-render loops)
  const onStatusChangeRef = useRef(onStatusChange)
  onStatusChangeRef.current = onStatusChange

  useEffect(() => {
    const activeMinionCount = minions.filter((m) => m.status === 'working' || m.status === 'spawning').length
    onStatusChangeRef.current?.({
      orchestratorStatus: streaming ? 'working' : 'idle',
      minionCount: minions.length,
      activeMinionCount,
      minions: minions.map((m) => ({ id: m.id, name: m.name, branch: m.branch, status: m.status })),
      activePanel
    })
  }, [streaming, minions, activePanel])

  // Create orchestrator session (use saved session ID for resume)
  useEffect(() => {
    if (!loaded) return
    let id: string | null = null
    window.api.chat.createSession({
      projectPath: project.path,
      claudeSessionId: claudeSessionIdRef.current ?? undefined
    }).then((result) => {
      id = result.sessionId
      setSessionId(id)
    })
    return () => {
      if (id) window.api.chat.destroySession({ sessionId: id })
    }
  }, [project.path, loaded])

  // Listen for minion updates — filtered to this project only
  useEffect(() => {
    const unsubUpdate = window.api.minion.onUpdate((minionInfo) => {
      if (minionInfo.projectPath !== project.path) return
      setMinions((prev) => {
        const existing = prev.find((m) => m.id === minionInfo.id)
        if (existing) {
          return prev.map((m) => m.id === minionInfo.id ? { ...m, status: minionInfo.status } : m)
        }
        return [...prev, {
          ...minionInfo,
          logs: []
        }]
      })
    })

    const unsubLog = window.api.minion.onLog((data) => {
      if (data.projectPath !== project.path) return
      setMinions((prev) =>
        prev.map((m) =>
          m.id === data.minionId
            ? { ...m, logs: [...m.logs, { ...data.log, timestamp: new Date(data.log.timestamp) }] }
            : m
        )
      )
    })

    return () => { unsubUpdate(); unsubLog() }
  }, [project.path])

  // Track which commands we've already processed
  const processedCommands = useRef(new Set<string>())

  // Handle spawn and message commands from orchestrator output
  const handleCommandsFromText = useCallback((text: string) => {
    const spawnCmds = parseSpawnCommands(text)
    for (const cmd of spawnCmds) {
      const key = `spawn:${cmd.name}:${cmd.task}`
      if (processedCommands.current.has(key)) continue
      processedCommands.current.add(key)

      window.api.minion.spawn({
        projectPath: project.path,
        name: cmd.name,
        task: cmd.task
      })
    }

    const msgCmds = parseMessageCommands(text)
    for (const cmd of msgCmds) {
      const key = `msg:${cmd.name}:${cmd.message}`
      if (processedCommands.current.has(key)) continue
      processedCommands.current.add(key)

      // Find minion by name
      const minion = minions.find((m) => m.name === cmd.name)
      if (minion) {
        window.api.minion.message({ minionId: minion.id, message: cmd.message })
      }
    }
  }, [project.path, minions])

  // Stream listener
  useEffect(() => {
    const unsubStream = window.api.chat.onStream((data) => {
      if (data.sessionId !== sessionId) return
      const evt = data.event
      const type = evt.type as string | undefined

      // Capture claude session ID for persistence
      if (type === 'system' && (evt as Record<string, unknown>).subtype === 'init' && (evt as Record<string, unknown>).session_id) {
        claudeSessionIdRef.current = (evt as Record<string, unknown>).session_id as string
      }

      if (type === 'assistant' && evt.message) {
        const msg = evt.message as Record<string, unknown>
        const content = msg.content as Array<Record<string, unknown>> | undefined
        if (content) {
          for (const block of content) {
            if (block.type === 'thinking' && block.thinking) {
              addMessage({ role: 'thinking', content: block.thinking as string })
            } else if (block.type === 'tool_use' && block.name === 'AskUserQuestion') {
              const input = block.input as { questions?: Question[] }
              if (input.questions && input.questions.length > 0) {
                setWaiting(false)
                setMessages((prev) => [...prev, {
                  id: crypto.randomUUID(),
                  role: 'ask-user' as const,
                  content: '',
                  questions: input.questions,
                  answered: false,
                  timestamp: new Date()
                }])
              }
            } else if (block.type === 'tool_use') {
              addMessage({
                role: 'tool-use',
                content: '',
                toolName: block.name as string,
                toolInput: block.input as Record<string, unknown>
              })
            } else if (block.type === 'text' && block.text) {
              const text = block.text as string
              updateAssistantMessage(text)
              handleCommandsFromText(text)
            }
          }
        }
      }
      // Tool results come as 'user' type events with tool_result content
      if (type === 'user') {
        const content = evt.content as Array<Record<string, unknown>> | undefined
        if (Array.isArray(content)) {
          for (const block of content) {
            if (block.type === 'tool_result' && block.content) {
              const resultText = typeof block.content === 'string'
                ? block.content
                : JSON.stringify(block.content)
              addMessage({ role: 'tool-result', content: resultText })
            }
          }
        }
      }
      if (type === 'content_block_delta') {
        const delta = evt.delta as Record<string, unknown> | undefined
        if (delta?.type === 'text_delta') appendToAssistantMessage(delta.text as string)
      }
      if (type === 'result') {
        const result = evt.result as string | undefined
        if (result) {
          updateAssistantMessage(result)
          handleCommandsFromText(result)
        }
      }
      if (type === 'error') {
        appendToAssistantMessage(`\n\n**Error:** ${(evt.error as string) || 'Unknown error'}`)
      }
    })
    const unsubDone = window.api.chat.onDone((data) => {
      if (data.sessionId !== sessionId) return
      setStreaming(false)
      setWaiting(false)
      currentAssistantId.current = null
    })
    return () => { unsubStream(); unsubDone() }
  }, [sessionId, handleCommandsFromText])

  const addMessage = useCallback((msg: {
    role: MessageRole; content: string; toolName?: string; toolInput?: Record<string, unknown>
  }): void => {
    setMessages((prev) => [...prev, {
      id: crypto.randomUUID(),
      role: msg.role,
      content: msg.content,
      toolName: msg.toolName,
      toolInput: msg.toolInput,
      timestamp: new Date()
    }])
  }, [])

  const updateAssistantMessage = useCallback((text: string): void => {
    const trimmed = text.trim()
    if (!trimmed) return
    setWaiting(false)
    if (!currentAssistantId.current) {
      const id = crypto.randomUUID()
      currentAssistantId.current = id
      setMessages((prev) => [...prev, { id, role: 'assistant', content: trimmed, timestamp: new Date() }])
    } else {
      const targetId = currentAssistantId.current
      setMessages((prev) => prev.map((m) => (m.id === targetId ? { ...m, content: trimmed } : m)))
    }
  }, [])

  const appendToAssistantMessage = useCallback((text: string): void => {
    setWaiting(false)
    if (!currentAssistantId.current) {
      const id = crypto.randomUUID()
      currentAssistantId.current = id
      setMessages((prev) => [...prev, { id, role: 'assistant', content: text, timestamp: new Date() }])
    } else {
      const targetId = currentAssistantId.current
      setMessages((prev) => prev.map((m) => (m.id === targetId ? { ...m, content: m.content + text } : m)))
    }
  }, [])

  const handleSend = async (): Promise<void> => {
    const text = input.trim()
    if (!text || !sessionId || streaming) return
    setMessages((prev) => [...prev, { id: crypto.randomUUID(), role: 'user', content: text, timestamp: new Date() }])
    setInput('')
    setStreaming(true)
    setWaiting(true)
    currentAssistantId.current = null
    const result = await window.api.chat.send({ sessionId, message: text })
    if (result.error) {
      setMessages((prev) => [...prev, { id: crypto.randomUUID(), role: 'assistant', content: `**Error:** ${result.error}`, timestamp: new Date() }])
      setStreaming(false)
      setWaiting(false)
    }
  }

  const handleAnswerQuestion = async (messageId: string, answers: readonly string[]): Promise<void> => {
    if (!sessionId) return

    // Mark the question as answered
    setMessages((prev) => prev.map((m) =>
      m.id === messageId ? { ...m, answered: true } : m
    ))

    // Send the answer as the next message
    const answerText = answers.join(', ')
    setMessages((prev) => [...prev, { id: crypto.randomUUID(), role: 'user', content: answerText, timestamp: new Date() }])
    setStreaming(true)
    setWaiting(true)
    currentAssistantId.current = null
    const result = await window.api.chat.send({ sessionId, message: answerText })
    if (result.error) {
      setMessages((prev) => [...prev, { id: crypto.randomUUID(), role: 'assistant', content: `**Error:** ${result.error}`, timestamp: new Date() }])
      setStreaming(false)
      setWaiting(false)
    }
  }

  const handleAbort = async (): Promise<void> => {
    if (sessionId) await window.api.chat.abort({ sessionId })
  }

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>): void => {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); handleSend() }
  }

  const handleReset = async (): Promise<void> => {
    // Abort any streaming
    if (sessionId) await window.api.chat.abort({ sessionId })

    // Kill and remove all minions
    for (const minion of minions) {
      await window.api.minion.remove({ minionId: minion.id, cleanupWorktree: true })
    }
    setMinions([])

    // Destroy orchestrator session
    if (sessionId) await window.api.chat.destroySession({ sessionId })

    // Clear state
    setMessages([])
    setMinions([])
    setStreaming(false)
    setWaiting(false)
    currentAssistantId.current = null
    claudeSessionIdRef.current = null
    processedCommands.current.clear()

    // Delete saved session
    await window.api.session.delete({ projectId: project.id })

    // Create fresh session
    const result = await window.api.chat.createSession({ projectPath: project.path })
    setSessionId(result.sessionId)
  }

  const handleKillMinion = async (minionId: string): Promise<void> => {
    await window.api.minion.kill({ minionId })
  }

  const handleRemoveMinion = async (minionId: string): Promise<void> => {
    await window.api.minion.remove({ minionId, cleanupWorktree: true })
    setMinions((prev) => prev.filter((m) => m.id !== minionId))
  }

  const activeMinion = minions.find((m) => m.id === activePanel)

  return (
    <div className={cn('flex flex-1 flex-col overflow-hidden', hidden && 'hidden')}>
      {/* Header */}
      <div
        className="flex h-[52px] shrink-0 items-center gap-3 border-b px-4"
        style={{ WebkitAppRegion: 'drag' } as React.CSSProperties}
      >
        <div className="flex items-center gap-2">
          <FolderGit2 className="size-4 text-muted-foreground" />
          <span className="text-sm font-medium">{project.name}</span>
          {activeMinion && (
            <>
              <span className="text-muted-foreground">/</span>
              <span className="text-sm text-muted-foreground">{activeMinion.name}</span>
              <StatusBadge status={activeMinion.status} />
            </>
          )}
        </div>
        <div className="ml-auto flex items-center gap-2" style={{ WebkitAppRegion: 'no-drag' } as React.CSSProperties}>
          {streaming && activePanel === 'orchestrator' && (
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <Loader2 className="size-3 animate-spin" />
              Claude is working...
            </div>
          )}
          <button
            onClick={handleReset}
            className="rounded-md p-1 text-muted-foreground transition-colors hover:text-destructive"
            title="Clear session and kill all minions"
          >
            <Trash2 className="size-4" />
          </button>
        </div>
      </div>

        {/* Content */}
        {activePanel === 'orchestrator' ? (
          <>
            {/* Orchestrator chat */}
            <div className="flex-1 overflow-y-auto">
              {messages.length === 0 ? (
                <div className="flex h-full flex-col items-center justify-center gap-4 px-4">
                  <div className="flex size-16 items-center justify-center rounded-2xl bg-primary/15">
                    <Crown className="size-8 text-primary" />
                  </div>
                  <div className="text-center">
                    <h2 className="text-lg font-semibold">Orchestrator</h2>
                    <p className="mt-1 max-w-sm text-sm text-muted-foreground">
                      Tell me what to build. I'll spawn minions to work on{' '}
                      <span className="font-medium text-foreground">{project.name}</span> in parallel,
                      each in their own git worktree.
                    </p>
                  </div>
                </div>
              ) : (
                <div className="mx-auto max-w-3xl space-y-6 px-4 py-6">
                  {messages.map((msg) => <ChatMessage key={msg.id} message={msg} onAnswer={handleAnswerQuestion} />)}
                  {waiting && <ThinkingIndicator />}
                  <div ref={messagesEndRef} />
                </div>
              )}
            </div>

            {/* Input */}
            <div className="shrink-0 border-t p-4">
              <div className="mx-auto max-w-3xl">
                <div className="relative rounded-lg border bg-card shadow-sm">
                  <textarea
                    ref={inputRef}
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    onKeyDown={handleKeyDown}
                    placeholder={streaming ? 'Orchestrator is working...' : 'Tell the orchestrator what to do...'}
                    disabled={streaming}
                    rows={1}
                    className="block w-full resize-none bg-transparent px-4 py-3 pr-12 text-sm placeholder:text-muted-foreground focus:outline-none disabled:opacity-50"
                    style={{ minHeight: '44px', maxHeight: '200px' }}
                    onInput={(e) => {
                      const target = e.target as HTMLTextAreaElement
                      target.style.height = '44px'
                      target.style.height = `${Math.min(target.scrollHeight, 200)}px`
                    }}
                  />
                  {streaming ? (
                    <Button size="icon" variant="ghost" className="absolute bottom-1.5 right-1.5 size-8 text-destructive" onClick={handleAbort}>
                      <Square className="size-4" />
                    </Button>
                  ) : (
                    <Button size="icon" variant="ghost" className="absolute bottom-1.5 right-1.5 size-8" onClick={handleSend} disabled={!input.trim() || !sessionId}>
                      <Send className="size-4" />
                    </Button>
                  )}
                </div>
                <p className="mt-2 text-center text-[11px] text-muted-foreground">
                  Press Enter to send, Shift+Enter for new line
                </p>
              </div>
            </div>
          </>
        ) : activeMinion ? (
          <MinionLogView
            minion={activeMinion}
            onKill={() => handleKillMinion(activeMinion.id)}
            onRemove={() => handleRemoveMinion(activeMinion.id)}
          />
        ) : null}
    </div>
  )
}

// --- Minion Log View (read-only) ---

function MinionLogView({
  minion,
  onKill,
  onRemove
}: {
  minion: Minion
  onKill: () => void
  onRemove: () => void
}): React.JSX.Element {
  const logsEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [minion.logs])

  return (
    <div className="flex flex-1 flex-col overflow-y-auto">
      {/* Minion info header */}
      <div className="border-b bg-card/50 px-6 py-4">
        <div className="flex items-center gap-3">
          <div className="flex size-10 items-center justify-center rounded-lg bg-secondary">
            <Cpu className="size-5" />
          </div>
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <span className="font-medium">{minion.name}</span>
              <StatusBadge status={minion.status} />
            </div>
            <div className="text-xs text-muted-foreground">{minion.task}</div>
          </div>
          <div className="flex items-center gap-2">
            <div className="flex items-center gap-1.5 rounded-md bg-muted px-2 py-1 text-xs text-muted-foreground">
              <GitBranch className="size-3" />
              {minion.branch}
            </div>
            {minion.status === 'working' && (
              <Button variant="outline" size="sm" onClick={onKill}>
                <Square className="size-3" />
                Stop
              </Button>
            )}
            {(minion.status === 'done' || minion.status === 'error') && (
              <Button variant="outline" size="sm" onClick={onRemove}>
                <X className="size-3" />
                Remove
              </Button>
            )}
          </div>
        </div>
      </div>

      {/* Logs */}
      <div className="flex-1 overflow-y-auto">
        <div className="mx-auto max-w-3xl space-y-4 px-4 py-6">
          {minion.logs.map((log) => (
            <LogEntry key={log.id} log={log} />
          ))}
          <div ref={logsEndRef} />
          {(minion.status === 'spawning' || minion.status === 'working') && (
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <Loader2 className="size-3.5 animate-spin" />
              Minion is working...
            </div>
          )}
        </div>
      </div>

      {/* Read-only notice */}
      <div className="shrink-0 border-t px-4 py-3">
        <p className="text-center text-xs text-muted-foreground">
          Read-only view. Talk to the Orchestrator to give instructions.
        </p>
      </div>
    </div>
  )
}

function LogEntry({ log }: { log: Message }): React.JSX.Element {
  const isSystem = log.role === 'system'

  return (
    <div className={cn('flex gap-3', isSystem && 'opacity-70')}>
      <div className={cn(
        'flex size-6 shrink-0 items-center justify-center rounded-full',
        isSystem ? 'bg-primary/15' : 'bg-secondary'
      )}>
        {isSystem ? <Crown className="size-3 text-primary" /> : <Bot className="size-3" />}
      </div>
      <div className="flex-1">
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium">
            {isSystem ? 'System' : 'Minion'}
          </span>
          <span className="text-[10px] text-muted-foreground">
            {log.timestamp instanceof Date ? formatTime(log.timestamp) : formatTime(new Date(log.timestamp))}
          </span>
        </div>
        <div className="mt-0.5 text-sm text-foreground/90">
          <Markdown content={log.content} />
        </div>
      </div>
    </div>
  )
}

// --- Shared Components ---

const SPAWN_LINE_REGEX = /\[SPAWN_MINION\s+name="[^"]+"\s+task="[^"]+"\]\n*/g
const MESSAGE_LINE_REGEX = /\[MESSAGE_MINION\s+name="[^"]+"\s+message="[^"]+"\]\n*/g

function stripCommands(text: string): string {
  return text.replace(SPAWN_LINE_REGEX, '').replace(MESSAGE_LINE_REGEX, '').trim()
}

function ChatMessage({ message, onAnswer }: {
  message: Message
  onAnswer?: (messageId: string, answers: readonly string[]) => void
}): React.JSX.Element {
  if (message.role === 'thinking') return <ThinkingBlock content={message.content} />
  if (message.role === 'tool-use') return <ToolUseBlock name={message.toolName ?? ''} input={message.toolInput ?? {}} />
  if (message.role === 'tool-result') return <ToolResultBlock content={message.content} />
  if (message.role === 'ask-user') return <AskUserBlock message={message} onAnswer={onAnswer} />

  const isUser = message.role === 'user'
  const displayContent = isUser ? message.content : stripCommands(message.content)

  if (!isUser && !displayContent) return <></>

  return (
    <div className={cn('flex gap-3', isUser && 'flex-row-reverse')}>
      <div className={cn(
        'flex size-8 shrink-0 items-center justify-center rounded-full',
        isUser ? 'bg-primary text-primary-foreground' : 'bg-primary/15'
      )}>
        {isUser ? <User className="size-4" /> : <Crown className="size-4 text-primary" />}
      </div>
      <div className={cn(
        'max-w-[80%] rounded-lg px-4 py-2.5 text-sm',
        isUser ? 'bg-primary text-primary-foreground' : 'bg-secondary'
      )}>
        {isUser
          ? <p className="whitespace-pre-wrap">{displayContent}</p>
          : <Markdown content={displayContent} />
        }
      </div>
    </div>
  )
}

function ThinkingBlock({ content }: { content: string }): React.JSX.Element {
  return (
    <div className="flex gap-3">
      <div className="flex size-6 shrink-0 items-center justify-center rounded-full bg-muted">
        <Brain className="size-3 text-muted-foreground" />
      </div>
      <p className="flex-1 whitespace-pre-wrap text-xs text-muted-foreground/70 italic">
        {content}
      </p>
    </div>
  )
}

function ToolUseBlock({ name, input }: { name: string; input: Record<string, unknown> }): React.JSX.Element {
  const [open, setOpen] = useState(false)

  const icon = (() => {
    switch (name) {
      case 'Bash': return <Terminal className="size-3" />
      case 'Read': return <Eye className="size-3" />
      case 'Write': return <FileText className="size-3" />
      case 'Edit': return <Pencil className="size-3" />
      case 'Grep': case 'Glob': return <Search className="size-3" />
      default: return <Crown className="size-3" />
    }
  })()

  const summary = (() => {
    if (name === 'Bash' && input.command) return String(input.command).slice(0, 120)
    if ((name === 'Read' || name === 'Write' || name === 'Edit') && input.file_path) return String(input.file_path)
    if ((name === 'Grep' || name === 'Glob') && input.pattern) return String(input.pattern)
    return ''
  })()

  return (
    <div className="flex gap-3">
      <div className="flex size-6 shrink-0 items-center justify-center rounded-full" style={{ backgroundColor: 'var(--accent)' }}>
        {icon}
      </div>
      <div className="flex-1 min-w-0">
        <button
          onClick={() => setOpen((v) => !v)}
          className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground"
        >
          {open ? <ChevronDown className="size-3" /> : <ChevronRight className="size-3" />}
          <span className="font-medium text-foreground/80">{name}</span>
          {summary && <code className="truncate text-[11px] text-muted-foreground/60">{summary}</code>}
        </button>
        {open && (
          <pre className="mt-1 overflow-x-auto rounded bg-black/20 p-2 text-[11px] text-muted-foreground">
            {JSON.stringify(input, null, 2)}
          </pre>
        )}
      </div>
    </div>
  )
}

function ToolResultBlock({ content }: { content: string }): React.JSX.Element {
  const [open, setOpen] = useState(false)
  const truncated = content.length > 300
  const display = open ? content : content.slice(0, 300)

  return (
    <div className="ml-9">
      <button
        onClick={() => setOpen((v) => !v)}
        className="flex items-center gap-1 text-[11px] text-muted-foreground/50 hover:text-muted-foreground"
      >
        {open ? <ChevronDown className="size-3" /> : <ChevronRight className="size-3" />}
        Output {truncated && !open ? `(${content.length} chars)` : ''}
      </button>
      {open && (
        <pre className="mt-1 max-h-[300px] overflow-auto rounded bg-black/20 p-2 text-[11px] text-muted-foreground">
          {display}
        </pre>
      )}
    </div>
  )
}

function AskUserBlock({ message, onAnswer }: {
  message: Message
  onAnswer?: (messageId: string, answers: readonly string[]) => void
}): React.JSX.Element {
  const [page, setPage] = useState(0)
  const [selections, setSelections] = useState<Map<number, Set<string>>>(new Map())

  const questions = message.questions ?? []
  const answered = message.answered
  const total = questions.length
  const current = questions[page]
  const isSingle = total === 1

  const getSelected = (qi: number): Set<string> => selections.get(qi) ?? new Set()

  const handleSelect = (label: string): void => {
    if (answered) return
    setSelections((prev) => {
      const next = new Map(prev)
      const set = new Set(prev.get(page) ?? [])
      if (current?.multiSelect) {
        if (set.has(label)) set.delete(label)
        else set.add(label)
      } else {
        set.clear()
        set.add(label)
      }
      next.set(page, set)
      return next
    })
  }

  const handleNext = (): void => {
    if (page < total - 1) setPage(page + 1)
  }

  const handleBack = (): void => {
    if (page > 0) setPage(page - 1)
  }

  const handleSubmit = (): void => {
    if (answered) return
    const allAnswers: string[] = []
    for (let i = 0; i < total; i++) {
      const set = selections.get(i)
      if (set) allAnswers.push(...Array.from(set))
    }
    if (allAnswers.length > 0) onAnswer?.(message.id, allAnswers)
  }

  // Single question + single select: auto-submit on click
  const handleQuickSelect = (label: string): void => {
    if (answered) return
    if (isSingle && !current?.multiSelect) {
      onAnswer?.(message.id, [label])
    } else {
      handleSelect(label)
    }
  }

  const currentSelected = getSelected(page)
  const isLastPage = page === total - 1
  const canProceed = currentSelected.size > 0
  const allAnswered = Array.from({ length: total }, (_, i) => (selections.get(i)?.size ?? 0) > 0).every(Boolean)

  if (!current) return <></>

  return (
    <div className="flex gap-3">
      <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary/15">
        <Crown className="size-4 text-primary" />
      </div>
      <div className="max-w-[80%] flex-1">
        <div className="rounded-lg border bg-card p-4">
          {/* Step indicator */}
          {!isSingle && (
            <div className="mb-3 flex items-center gap-2">
              {questions.map((_, i) => (
                <div
                  key={i}
                  className={cn(
                    'size-2 rounded-full transition-colors',
                    i === page ? 'bg-primary' : (selections.get(i)?.size ?? 0) > 0 ? 'bg-primary/40' : 'bg-muted-foreground/20'
                  )}
                />
              ))}
              <span className="ml-1 text-[10px] text-muted-foreground">{page + 1} of {total}</span>
            </div>
          )}

          {current.header && (
            <p className="mb-1 text-xs font-semibold uppercase tracking-wider text-muted-foreground">{current.header}</p>
          )}
          <p className="mb-3 text-sm font-medium">{current.question}</p>

          <div className="space-y-2">
            {current.options.map((opt) => {
              const isSelected = currentSelected.has(opt.label)

              return (
                <button
                  key={opt.label}
                  onClick={() => handleQuickSelect(opt.label)}
                  disabled={answered}
                  className={cn(
                    'flex w-full items-start gap-3 rounded-lg border p-3 text-left text-sm transition-colors',
                    answered
                      ? 'opacity-50 cursor-default'
                      : isSelected
                        ? 'border-primary bg-primary/10'
                        : 'hover:bg-accent'
                  )}
                >
                  <div className={cn(
                    'mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-full border',
                    isSelected ? 'border-primary bg-primary' : 'border-muted-foreground/30'
                  )}>
                    {isSelected && <div className="size-1.5 rounded-full bg-primary-foreground" />}
                  </div>
                  <div>
                    <p className="font-medium">{opt.label}</p>
                    {opt.description && (
                      <p className="mt-0.5 text-xs text-muted-foreground">{opt.description}</p>
                    )}
                  </div>
                </button>
              )
            })}
          </div>

          {/* Navigation */}
          {!answered && !isSingle && (
            <div className="mt-4 flex items-center justify-between">
              <button
                onClick={handleBack}
                disabled={page === 0}
                className={cn(
                  'text-xs font-medium transition-colors',
                  page === 0 ? 'text-muted-foreground/30' : 'text-muted-foreground hover:text-foreground'
                )}
              >
                Back
              </button>
              {isLastPage ? (
                <button
                  onClick={handleSubmit}
                  disabled={!allAnswered}
                  className={cn(
                    'rounded-md px-4 py-1.5 text-xs font-medium transition-colors',
                    allAnswered
                      ? 'bg-primary text-primary-foreground hover:bg-primary/90'
                      : 'bg-muted text-muted-foreground'
                  )}
                >
                  Submit
                </button>
              ) : (
                <button
                  onClick={handleNext}
                  disabled={!canProceed}
                  className={cn(
                    'rounded-md px-4 py-1.5 text-xs font-medium transition-colors',
                    canProceed
                      ? 'bg-primary text-primary-foreground hover:bg-primary/90'
                      : 'bg-muted text-muted-foreground'
                  )}
                >
                  Next
                </button>
              )}
            </div>
          )}

          {/* Single question multi-select submit */}
          {!answered && isSingle && current.multiSelect && canProceed && (
            <button
              onClick={handleSubmit}
              className="mt-3 rounded-md bg-primary px-4 py-1.5 text-xs font-medium text-primary-foreground hover:bg-primary/90"
            >
              Submit
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

function ThinkingIndicator(): React.JSX.Element {
  return (
    <div className="flex gap-3">
      <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-primary/15">
        <Crown className="size-4 text-primary" />
      </div>
      <div className="flex items-center gap-1.5 rounded-lg bg-secondary px-4 py-3">
        <span className="size-2 animate-bounce rounded-full [animation-delay:0ms]" style={{ backgroundColor: 'var(--muted-foreground)' }} />
        <span className="size-2 animate-bounce rounded-full [animation-delay:150ms]" style={{ backgroundColor: 'var(--muted-foreground)' }} />
        <span className="size-2 animate-bounce rounded-full [animation-delay:300ms]" style={{ backgroundColor: 'var(--muted-foreground)' }} />
      </div>
    </div>
  )
}

function StatusDot({ status }: { status: MinionStatus }): React.JSX.Element {
  const colors: Record<MinionStatus, string> = {
    spawning: 'text-yellow-400 animate-pulse',
    working: 'text-blue-400 animate-pulse',
    done: 'text-emerald-400',
    error: 'text-destructive'
  }
  return <Circle className={cn('size-2.5 fill-current', colors[status])} />
}

function StatusBadge({ status }: { status: MinionStatus }): React.JSX.Element {
  const variants: Record<MinionStatus, 'default' | 'secondary' | 'success' | 'destructive'> = {
    spawning: 'default',
    working: 'default',
    done: 'success',
    error: 'destructive'
  }
  return <Badge variant={variants[status]} className="text-[10px]">{status}</Badge>
}

function formatTime(date: Date): string {
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
