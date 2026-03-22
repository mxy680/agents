import { NextRequest } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { deployAgent, streamAgentLogs, stopAgent, getAgentStatus } from "@/lib/orchestrator-client"
import { parseNDJSON } from "@/lib/agent-events"
import { generateTitle } from "@/lib/auto-title"
import { checkOrigin } from "@/lib/csrf"

// Terminal statuses — the agent will not produce more output after reaching these.
const TERMINAL_STATUSES = new Set(["completed", "failed", "stopped"])
// Maximum wait time for an agent to reach "running" before giving up.
const STATUS_POLL_TIMEOUT_MS = 60_000
const STATUS_POLL_INTERVAL_MS = 1_000

export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ agentName: string }> }
) {
  const csrfError = checkOrigin(request)
  if (csrfError) return csrfError

  const { agentName } = await params

  // Validate agentName: only allow alphanumeric, hyphens, underscores
  if (!/^[a-z0-9-_]+$/i.test(agentName)) {
    return new Response(JSON.stringify({ error: "Invalid agent name" }), { status: 400 })
  }

  // Authenticate user
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    return new Response(JSON.stringify({ error: "Unauthorized" }), { status: 401 })
  }

  // Get access token for orchestrator auth
  const { data: { session } } = await supabase.auth.getSession()
  const authToken = session?.access_token

  // Parse request body
  let message: string
  let sessionId: string | undefined
  let conversationId: string | undefined
  try {
    const body = await request.json() as { message?: unknown; sessionId?: unknown; conversationId?: unknown }
    message = body.message as string
    sessionId = body.sessionId as string | undefined
    conversationId = body.conversationId as string | undefined
  } catch {
    return new Response(JSON.stringify({ error: "Invalid request body" }), { status: 400 })
  }

  if (!message || typeof message !== "string") {
    return new Response(JSON.stringify({ error: "message is required" }), { status: 400 })
  }

  const MAX_MESSAGE_LENGTH = 32_000
  if (message.length > MAX_MESSAGE_LENGTH) {
    return new Response(JSON.stringify({ error: "Message too long" }), { status: 400 })
  }

  // If conversationId provided, look up existing session_id
  const admin = createAdminClient()
  if (conversationId && !sessionId) {
    const { data: conv } = await supabase
      .from("conversations")
      .select("session_id")
      .eq("id", conversationId)
      .eq("user_id", user.id)
      .single()
    if (conv?.session_id) {
      sessionId = conv.session_id
    }
  }

  // Check if this is the first exchange (no prior messages) for auto-titling
  let isFirstExchange = false
  if (conversationId) {
    const { data: existingMsgs } = await admin
      .from("conversation_messages")
      .select("id")
      .eq("conversation_id", conversationId)
      .limit(1)
    isFirstExchange = !existingMsgs || existingMsgs.length === 0
  }

  // Persist user message before streaming
  if (conversationId) {
    await admin.from("conversation_messages").insert({
      conversation_id: conversationId,
      user_id: user.id,
      role: "user",
      blocks: [{ type: "text", content: message }],
    })
  }

  // Look up the template by agent name
  const { data: template, error: templateError } = await admin
    .from("agent_templates")
    .select("id")
    .eq("name", agentName)
    .eq("status", "active")
    .single()

  if (templateError || !template) {
    return new Response(
      JSON.stringify({ error: `Agent template "${agentName}" not found` }),
      { status: 404 }
    )
  }

  // Set up SSE stream
  const stream = new ReadableStream({
    async start(controller) {
      const encoder = new TextEncoder()

      function send(event: string, data: string) {
        // SSE spec: multi-line data must be sent as multiple "data:" lines
        const lines = data.split("\n").map((l) => `data: ${l}`).join("\n")
        controller.enqueue(encoder.encode(`event: ${event}\n${lines}\n\n`))
      }

      const abortController = new AbortController()

      // If client disconnects, abort the stream
      request.signal.addEventListener("abort", () => {
        abortController.abort()
      })

      // Collect assistant output for persistence
      type ContentBlock = { type: "text"; content: string } | { type: "tool"; id: string; name: string; finalInput: string; result?: string }
      const assistantBlocks: ContentBlock[] = []
      let activeToolId: string | null = null
      let instanceId: string | null = null

      try {
        // Deploy the agent via the orchestrator, passing prompt + sessionId as config_overrides.
        // The orchestrator stores these; the pod_spec injects uppercase string config_overrides as env vars.
        const configOverrides: Record<string, unknown> = {
          AGENT_PROMPT: message,
        }
        if (sessionId) {
          configOverrides.AGENT_SESSION_ID = sessionId
        }

        const instance = await deployAgent(template.id, configOverrides, authToken)
        instanceId = instance.id

        // Poll until the instance is running or reaches a terminal state
        let status = instance.status
        const pollStart = Date.now()
        while (status === "pending" || status === "creating") {
          if (Date.now() - pollStart > STATUS_POLL_TIMEOUT_MS) {
            throw new Error("Timed out waiting for agent to start")
          }
          await new Promise<void>((r) => setTimeout(r, STATUS_POLL_INTERVAL_MS))
          const updated = await getAgentStatus(instance.id, authToken)
          status = updated.status
        }

        if (TERMINAL_STATUSES.has(status) && status !== "completed") {
          throw new Error(`Agent failed to start (status: ${status})`)
        }

        // Stream logs from the orchestrator
        const logStream = await streamAgentLogs(instance.id, authToken, abortController.signal)
        const reader = logStream.getReader()
        const decoder = new TextDecoder()

        // State for NDJSON parsing
        const toolInputAccum: Record<string, string> = {}
        const toolIdRef = { currentToolId: null as string | null }
        // Buffer for incomplete NDJSON lines
        let ndjsonLineBuffer = ""
        // Buffer for incomplete SSE frames from the orchestrator's log stream
        let sseFrameBuffer = ""

        try {
          while (true) {
            const { done, value } = await reader.read()
            if (done) break

            // The orchestrator sends SSE-formatted chunks: `data: <raw stdout bytes>\n\n`
            // We need to extract the raw stdout content from each SSE frame.
            sseFrameBuffer += decoder.decode(value, { stream: true })

            // Split on double-newline SSE frame boundaries
            const frames = sseFrameBuffer.split("\n\n")
            // Last element may be an incomplete frame — keep it in the buffer
            sseFrameBuffer = frames.pop() ?? ""

            for (const frame of frames) {
              // Extract content from "data: ..." lines, joining multi-line data
              const dataLines: string[] = []
              for (const line of frame.split("\n")) {
                if (line.startsWith("data: ")) {
                  dataLines.push(line.slice(6))
                } else if (line.startsWith("data:")) {
                  dataLines.push(line.slice(5))
                }
              }
              if (dataLines.length === 0) continue

              const rawContent = dataLines.join("\n")
              const { events, remainingBuffer } = parseNDJSON(
                rawContent,
                ndjsonLineBuffer,
                toolInputAccum,
                toolIdRef
              )
              ndjsonLineBuffer = remainingBuffer

              for (const sseEvent of events) {
                send(sseEvent.event, sseEvent.data)

                // Track assistant content for persistence
                switch (sseEvent.event) {
                  case "delta": {
                    const last = assistantBlocks[assistantBlocks.length - 1]
                    if (last?.type === "text") {
                      assistantBlocks[assistantBlocks.length - 1] = { ...last, content: last.content + sseEvent.data }
                    } else {
                      assistantBlocks.push({ type: "text", content: sseEvent.data })
                    }
                    break
                  }
                  case "tool_start": {
                    try {
                      const { name, id } = JSON.parse(sseEvent.data) as { name: string; id: string }
                      activeToolId = id
                      assistantBlocks.push({ type: "tool", id, name, finalInput: "" })
                    } catch { /* ignore */ }
                    break
                  }
                  case "tool_input": {
                    if (activeToolId) {
                      const toolId = activeToolId
                      const idx = assistantBlocks.findIndex((b) => b.type === "tool" && b.id === toolId)
                      if (idx !== -1) {
                        assistantBlocks[idx] = { ...assistantBlocks[idx] as Extract<ContentBlock, { type: "tool" }>, finalInput: sseEvent.data }
                      }
                    }
                    break
                  }
                  case "tool_result": {
                    try {
                      const parsed = JSON.parse(sseEvent.data) as { summary: string; toolUseId?: string }
                      const toolId: string | null = parsed.toolUseId ?? activeToolId
                      if (toolId) {
                        const idx = assistantBlocks.findIndex((b) => b.type === "tool" && b.id === toolId)
                        if (idx !== -1) {
                          const resultStr = typeof parsed.summary === "string" ? parsed.summary : JSON.stringify(parsed.summary)
                          assistantBlocks[idx] = { ...assistantBlocks[idx] as Extract<ContentBlock, { type: "tool" }>, result: resultStr }
                        }
                        if (toolId === activeToolId) activeToolId = null
                      }
                    } catch { /* ignore */ }
                    break
                  }
                  case "session": {
                    // Update conversation with session_id
                    if (conversationId) {
                      admin.from("conversations").update({ session_id: sseEvent.data }).eq("id", conversationId).then(() => {})
                    }
                    break
                  }
                }
              }
            }
          }

          // Flush any remaining NDJSON line buffer after stream ends
          if (ndjsonLineBuffer.trim()) {
            const { events } = parseNDJSON("", ndjsonLineBuffer, toolInputAccum, toolIdRef)
            for (const sseEvent of events) {
              send(sseEvent.event, sseEvent.data)
            }
          }
        } finally {
          reader.releaseLock()
        }
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e)
        send("error", msg)
        // Attempt to stop the agent if we deployed one and something went wrong
        if (instanceId) {
          stopAgent(instanceId, authToken).catch(() => {})
        }
      } finally {
        // Persist assistant message and update conversation
        if (conversationId && assistantBlocks.length > 0) {
          admin.from("conversation_messages").insert({
            conversation_id: conversationId,
            user_id: user.id,
            role: "assistant",
            blocks: assistantBlocks,
          }).then(() => {})

          admin.from("conversations").update({ updated_at: new Date().toISOString() }).eq("id", conversationId).then(() => {})

          // Auto-title on first exchange
          if (isFirstExchange) {
            const title = generateTitle(message)
            admin.from("conversations").update({ title }).eq("id", conversationId).then(() => {})
          }
        }

        controller.close()
      }
    },
  })

  return new Response(stream, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache, no-transform",
      Connection: "keep-alive",
      "X-Accel-Buffering": "no",
    },
  })
}
