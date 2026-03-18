import { spawn } from "child_process"
import { mkdirSync, writeFileSync } from "fs"
import path from "path"

export interface ChatSSEEvent {
  event: "delta" | "tool_start" | "tool_input" | "tool_result" | "session" | "result" | "error"
  data: string
}

export interface ContainerOptions {
  instancePath: string
  message: string
  sessionId?: string
  model?: string
  env: Record<string, string>
  signal?: AbortSignal
  timeoutMs?: number
}

/**
 * Spawns a Docker container per message, parses NDJSON stdout into ChatSSEEvents.
 * Returns an async generator of ChatSSEEvents.
 */
export async function* runContainer(opts: ContainerOptions): AsyncGenerator<ChatSSEEvent> {
  const { instancePath, message, sessionId, model, env, signal, timeoutMs } = opts

  // Ensure instance directory exists
  mkdirSync(instancePath, { recursive: true })

  // Write session.json
  const sessionJson: { prompt: string; sessionId?: string; model?: string } = {
    prompt: message,
  }
  if (sessionId) sessionJson.sessionId = sessionId
  if (model) sessionJson.model = model
  writeFileSync(path.join(instancePath, "session.json"), JSON.stringify(sessionJson))

  // Build docker run args
  const dockerArgs: string[] = [
    "run",
    "--rm",
    `-v`,
    `${instancePath}:/agent/workspace`,
    "--network=bridge",
    "--cpus=2",
    "--memory=2g",
  ]

  // Pass env vars as -e flags
  for (const [key, value] of Object.entries(env)) {
    dockerArgs.push("-e", `${key}=${value}`)
  }

  dockerArgs.push("ghcr.io/emdash-projects/agent-base:dev")

  const proc = spawn("docker", dockerArgs, {
    stdio: ["ignore", "pipe", "pipe"],
  })

  // Handle abort signal
  if (signal) {
    signal.addEventListener("abort", () => {
      proc.kill("SIGTERM")
    })
  }

  // Handle timeout
  let aborted = false
  let timeoutHandle: ReturnType<typeof setTimeout> | undefined
  if (timeoutMs) {
    timeoutHandle = setTimeout(() => {
      aborted = true
      proc.kill("SIGTERM")
    }, timeoutMs)
  }

  const queue: Array<ChatSSEEvent | null> = []
  let resolve: (() => void) | null = null

  function enqueue(event: ChatSSEEvent | null) {
    queue.push(event)
    resolve?.()
    resolve = null
  }

  async function waitForItem(): Promise<void> {
    if (queue.length > 0) return
    return new Promise((r) => {
      resolve = r
    })
  }

  // Collect stderr for error reporting
  const stderrChunks: Buffer[] = []
  proc.stderr.on("data", (chunk: Buffer) => {
    stderrChunks.push(chunk)
  })

  // Track tool input accumulation per tool use id
  const toolInputAccum: Record<string, string> = {}
  let currentToolId: string | null = null

  // Parse NDJSON stdout
  let lineBuffer = ""
  proc.stdout.on("data", (chunk: Buffer) => {
    lineBuffer += chunk.toString("utf8")
    const lines = lineBuffer.split("\n")
    lineBuffer = lines.pop() ?? ""

    for (const line of lines) {
      const trimmed = line.trim()
      if (!trimmed) continue

      let parsed: Record<string, unknown>
      try {
        parsed = JSON.parse(trimmed)
      } catch {
        continue
      }

      const events = mapAgentEvent(parsed, toolInputAccum, { get currentToolId() { return currentToolId }, set currentToolId(v) { currentToolId = v } })
      for (const e of events) {
        enqueue(e)
      }
    }
  })

  proc.on("close", () => {
    if (timeoutHandle) clearTimeout(timeoutHandle)
    enqueue(null) // sentinel
  })

  proc.on("error", (err) => {
    if (timeoutHandle) clearTimeout(timeoutHandle)
    enqueue({ event: "error", data: `Process error: ${err.message}` })
    enqueue(null)
  })

  // Yield events as they arrive
  try {
    while (true) {
      await waitForItem()
      while (queue.length > 0) {
        const item = queue.shift()!
        if (item === null) return
        yield item
      }
    }
  } finally {
    if (timeoutHandle) clearTimeout(timeoutHandle)
    if (!proc.killed) proc.kill("SIGTERM")
  }
}

interface ToolIdRef {
  currentToolId: string | null
}

function mapAgentEvent(
  raw: Record<string, unknown>,
  toolInputAccum: Record<string, string>,
  toolIdRef: ToolIdRef
): ChatSSEEvent[] {
  const rawType = raw.type as string | undefined

  // Agent SDK wraps stream events: {type: "stream_event", event: {type: "content_block_delta", ...}}
  // Unwrap to get the inner event for content parsing.
  let event: Record<string, unknown>
  let type: string | undefined
  if (rawType === "stream_event" && raw.event && typeof raw.event === "object") {
    event = raw.event as Record<string, unknown>
    type = event.type as string | undefined
  } else {
    event = raw
    type = rawType
  }

  switch (type) {
    case "content_block_start": {
      const block = event.content_block as Record<string, unknown> | undefined
      if (block?.type === "tool_use") {
        const id = block.id as string
        const name = block.name as string
        toolIdRef.currentToolId = id
        toolInputAccum[id] = ""
        return [{ event: "tool_start", data: JSON.stringify({ name, id }) }]
      }
      return []
    }

    case "content_block_delta": {
      const delta = event.delta as Record<string, unknown> | undefined
      if (!delta) return []

      if (delta.type === "text_delta") {
        const text = delta.text as string
        return [{ event: "delta", data: text }]
      }

      if (delta.type === "input_json_delta") {
        const partial = delta.partial_json as string
        if (toolIdRef.currentToolId) {
          toolInputAccum[toolIdRef.currentToolId] = (toolInputAccum[toolIdRef.currentToolId] ?? "") + partial
        }
        return [{ event: "tool_input", data: partial }]
      }

      return []
    }

    case "content_block_stop": {
      if (toolIdRef.currentToolId && toolInputAccum[toolIdRef.currentToolId] !== undefined) {
        const accumulated = toolInputAccum[toolIdRef.currentToolId]
        toolIdRef.currentToolId = null
        return [{ event: "tool_input", data: accumulated }]
      }
      return []
    }

    case "tool_use_summary": {
      const summary = event.summary as string | undefined
      return [{ event: "tool_result", data: JSON.stringify({ summary: summary ?? "" }) }]
    }

    case "result": {
      // result events come at top level from Agent SDK (not wrapped in stream_event)
      const sessionId = (raw.session_id ?? event.session_id) as string | undefined
      const events: ChatSSEEvent[] = []
      if (sessionId) {
        events.push({ event: "session", data: sessionId })
      }
      events.push({
        event: "result",
        data: JSON.stringify({
          sessionId,
          result: raw.result ?? event.result,
          costUsd: raw.total_cost_usd ?? event.total_cost_usd,
          numTurns: raw.num_turns ?? event.num_turns,
          stopReason: raw.stop_reason ?? event.stop_reason,
        }),
      })
      return events
    }

    default:
      return []
  }
}
