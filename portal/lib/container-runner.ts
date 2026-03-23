import { spawn } from "child_process"
import { mkdirSync, writeFileSync } from "fs"
import path from "path"
import { type ChatSSEEvent, mapAgentEvent, type ToolIdRef } from "@/lib/agent-events"

export type { ChatSSEEvent }

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
