import { spawn } from "child_process"
import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "fs"
import path from "path"
import os from "os"
import { type ChatSSEEvent, type StreamState, mapAgentEvent } from "@/lib/agent-events"
import { resolveAdminCredentials } from "@/lib/credentials"

export type { ChatSSEEvent }

export interface LocalRunnerOptions {
  agentName: string
  message: string
  systemPrompt: string
  sessionId?: string
  model?: string
  signal?: AbortSignal
  timeoutMs?: number
}

// Resolve the repo root. In local dev, portal is one level down from repo root.
// In Docker/Fly, agents are copied to /app/agents/ alongside the portal.
const REPO_ROOT = (() => {
  const fromCwd = path.resolve(process.cwd(), "..")
  // Check if agents dir exists at the parent (local dev)
  try {
    const { statSync } = require("fs")
    statSync(path.join(fromCwd, "agents"))
    return fromCwd
  } catch {
    // Fallback: agents are in the same dir as portal (Docker)
    return process.cwd()
  }
})()

/**
 * Spawns `node entrypoint.mjs session.json` for a local agent (no Docker).
 * Streams NDJSON stdout as ChatSSEEvents.
 */
export async function* runLocal(opts: LocalRunnerOptions): AsyncGenerator<ChatSSEEvent> {
  const { agentName, message, systemPrompt, sessionId, model, signal, timeoutMs } = opts

  // Resolve credentials from Supabase + inherit all process.env (doppler vars already present).
  const credEnv = await resolveAdminCredentials()

  // Build env: inherit everything, add resolved creds, and set PATH / NODE_PATH.
  const binDir = path.join(REPO_ROOT, "bin")
  const existingPath = process.env.PATH ?? ""
  const npmGlobalRoot = "/opt/homebrew/lib/node_modules"
  const existingNodePath = process.env.NODE_PATH ?? ""

  const env: Record<string, string> = {
    ...process.env as Record<string, string>,
    ...credEnv,
    PATH: `${binDir}:${existingPath}`,
    NODE_PATH: existingNodePath ? `${npmGlobalRoot}:${existingNodePath}` : npmGlobalRoot,
  }

  // Remove platform infrastructure vars — agents must NOT access the platform database
  delete env.NEXT_PUBLIC_SUPABASE_URL
  delete env.NEXT_PUBLIC_SUPABASE_ANON_KEY
  delete env.SUPABASE_SERVICE_ROLE_KEY
  delete env.SUPABASE_DB_URL
  delete env.SUPABASE_JWT_SECRET
  delete env.ENCRYPTION_MASTER_KEY

  // Create a temp directory for the session file.
  const tmpDir = mkdtempSync(path.join(os.tmpdir(), `local-agent-${agentName}-`))
  mkdirSync(tmpDir, { recursive: true })

  // Write session.json — entrypoint.mjs reads: prompt, systemPrompt, model.
  const sessionData: { prompt: string; systemPrompt: string; model?: string; sessionId?: string } = {
    prompt: message,
    systemPrompt,
    model: model ?? "claude-opus-4-6",
  }
  if (sessionId) sessionData.sessionId = sessionId

  const sessionFile = path.join(tmpDir, "session.json")
  writeFileSync(sessionFile, JSON.stringify(sessionData))

  // Agent entrypoint path: <repo>/agents/<agentName>/entrypoint.mjs
  // NOTE: path constructed via array join to prevent Turbopack from analyzing it as a module
  const entrypointPath = [REPO_ROOT, "agents", agentName, "entrypoint.mjs"].join(path.sep)

  // Verify the entrypoint exists before spawning.
  try {
    readFileSync(entrypointPath)
  } catch {
    yield { event: "error", data: `Entrypoint not found: ${entrypointPath}` }
    return
  }

  const agentCwd = path.join(REPO_ROOT, "agents", agentName)

  console.log("[local-runner] Spawning:", entrypointPath, sessionFile)
  console.log("[local-runner] CWD:", agentCwd)
  console.log("[local-runner] PATH includes bin:", env.PATH?.includes("/bin"))
  console.log("[local-runner] NODE_PATH:", env.NODE_PATH)

  const proc = spawn("node", [entrypointPath, sessionFile], {
    stdio: ["ignore", "pipe", "pipe"],
    env: env as NodeJS.ProcessEnv,
    cwd: agentCwd,
  })

  // Handle abort signal.
  if (signal) {
    signal.addEventListener("abort", () => {
      proc.kill("SIGTERM")
    })
  }

  // Handle timeout.
  let timedOut = false
  let timeoutHandle: ReturnType<typeof setTimeout> | undefined
  if (timeoutMs) {
    timeoutHandle = setTimeout(() => {
      timedOut = true
      proc.kill("SIGTERM")
    }, timeoutMs)
  }

  // Producer/consumer queue so we can yield from an async generator
  // while the child process pushes events from callbacks.
  const queue: Array<ChatSSEEvent | null> = []
  let resolveWait: (() => void) | null = null

  function enqueue(event: ChatSSEEvent | null) {
    queue.push(event)
    resolveWait?.()
    resolveWait = null
  }

  function waitForItem(): Promise<void> {
    if (queue.length > 0) return Promise.resolve()
    return new Promise((r) => { resolveWait = r })
  }

  // Collect stderr for error reporting.
  const stderrChunks: Buffer[] = []
  proc.stderr.on("data", (chunk: Buffer) => {
    stderrChunks.push(chunk)
    console.log("[local-runner] STDERR:", chunk.toString("utf8").slice(0, 200))
  })

  // Track tool input accumulation per tool use id.
  const toolInputAccum: Record<string, string> = {}
  let currentToolId: string | null = null
  const streamState: StreamState = { hasReceivedDeltas: false }

  // Parse NDJSON stdout line by line.
  let lineBuffer = ""
  proc.stdout.on("data", (chunk: Buffer) => {
    console.log("[local-runner] STDOUT chunk:", chunk.toString("utf8").slice(0, 200))
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
        // Non-JSON stdout (e.g. ANSI terminal output from entrypoint.mjs) — skip silently.
        continue
      }

      const toolIdRef = {
        get currentToolId() { return currentToolId },
        set currentToolId(v: string | null) { currentToolId = v },
      }
      const events = mapAgentEvent(parsed, toolInputAccum, toolIdRef, streamState)
      for (const e of events) {
        enqueue(e)
      }
    }
  })

  proc.on("close", (code) => {
    if (timeoutHandle) clearTimeout(timeoutHandle)
    if (timedOut) {
      enqueue({ event: "error", data: "Agent timed out" })
    } else if (code !== 0 && code !== null) {
      const stderr = Buffer.concat(stderrChunks).toString("utf8").trim()
      const msg = stderr || `Process exited with code ${code}`
      enqueue({ event: "error", data: msg })
    }
    enqueue(null) // sentinel — generator stops here
  })

  proc.on("error", (err) => {
    if (timeoutHandle) clearTimeout(timeoutHandle)
    enqueue({ event: "error", data: `Process error: ${err.message}` })
    enqueue(null)
  })

  // Yield events as they arrive.
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
    // #7: Clean up the temp directory on every exit path
    try {
      rmSync(tmpDir, { recursive: true, force: true })
    } catch {
      // Ignore cleanup errors
    }
  }
}

/**
 * Reads role.md + CLAUDE.md from the agent directory and concatenates them
 * into a single system prompt string.
 */
export function buildSystemPrompt(agentName: string, contextPrefix?: string): string {
  const agentDir = path.join(REPO_ROOT, "agents", agentName)
  const parts: string[] = []

  if (contextPrefix?.trim()) {
    parts.push(contextPrefix.trim())
  }

  for (const filename of ["role.md", "CLAUDE.md"]) {
    try {
      const content = readFileSync(path.join(agentDir, filename), "utf-8").trim()
      if (content) parts.push(content)
    } catch {
      // File missing — skip.
    }
  }

  // Include skill files from agents/<name>/skills/
  const skillsDir = path.join(agentDir, "skills")
  try {
    const { readdirSync } = require("fs")
    const skillFiles = (readdirSync(skillsDir) as string[]).filter((f: string) => f.endsWith(".md")).sort()
    for (const sf of skillFiles) {
      try {
        const content = readFileSync(path.join(skillsDir, sf), "utf-8").trim()
        if (content) parts.push(content)
      } catch {
        // Skip unreadable skill files
      }
    }
  } catch {
    // Skills directory doesn't exist — skip
  }

  return parts.join("\n\n")
}
