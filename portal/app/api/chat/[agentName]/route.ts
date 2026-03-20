import { NextRequest } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { resolveUserCredentials } from "@/lib/credentials"
import { runContainer } from "@/lib/container-runner"
import { generateTitle } from "@/lib/auto-title"
import path from "path"
import os from "os"

export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ agentName: string }> }
) {
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

  // Check that the user has acquired this agent
  const { data: template } = await supabase
    .from("agent_templates")
    .select("id")
    .eq("name", agentName)
    .eq("status", "active")
    .single()

  if (template) {
    const { data: userAgent } = await supabase
      .from("user_agents")
      .select("template_id, status")
      .eq("user_id", user.id)
      .eq("template_id", template.id)
      .maybeSingle()

    if (!userAgent) {
      return new Response(
        JSON.stringify({ error: "Agent not acquired. Visit the marketplace to get this agent." }),
        { status: 403 }
      )
    }

    if (userAgent.status === "pending") {
      return new Response(
        JSON.stringify({ error: "Your access to this agent is pending admin approval." }),
        { status: 403 }
      )
    }

    if (userAgent.status === "rejected") {
      return new Response(
        JSON.stringify({ error: "Your access to this agent was not approved." }),
        { status: 403 }
      )
    }
  }

  // Parse request body
  let message: string
  let sessionId: string | undefined
  let conversationId: string | undefined
  try {
    const body = await request.json()
    message = body.message
    sessionId = body.sessionId
    conversationId = body.conversationId
  } catch {
    return new Response(JSON.stringify({ error: "Invalid request body" }), { status: 400 })
  }

  if (!message || typeof message !== "string") {
    return new Response(JSON.stringify({ error: "message is required" }), { status: 400 })
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

  // Resolve credentials for all active integrations
  let credEnv: Record<string, string>
  try {
    credEnv = await resolveUserCredentials(user.id)
  } catch (e) {
    console.error("Failed to resolve user credentials:", e)
    return new Response(JSON.stringify({ error: "Failed to fetch integrations" }), { status: 500 })
  }

  // CLAUDE_CODE_OAUTH_TOKEN is required
  const claudeToken = process.env.CLAUDE_CODE_OAUTH_TOKEN
  if (!claudeToken) {
    return new Response(JSON.stringify({ error: "CLAUDE_CODE_OAUTH_TOKEN not configured" }), { status: 500 })
  }

  // Build instance directory
  const instancePath = path.join(
    os.tmpdir(),
    "agents",
    user.id,
    agentName,
    String(Date.now())
  )

  // Copy agent template files if they exist
  try {
    const { mkdirSync, copyFileSync, existsSync } = await import("fs")
    mkdirSync(instancePath, { recursive: true })

    const agentsDir = path.join(process.cwd(), "..", "agents", agentName)
    for (const file of ["role.md", "CLAUDE.md"]) {
      const src = path.join(agentsDir, file)
      if (existsSync(src)) {
        copyFileSync(src, path.join(instancePath, file))
      }
    }
  } catch (e) {
    console.error("Failed to copy agent template files:", e)
  }

  const containerEnv: Record<string, string> = {
    CLAUDE_CODE_OAUTH_TOKEN: claudeToken,
    ...credEnv,
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

      // If client disconnects, abort the container
      request.signal.addEventListener("abort", () => {
        abortController.abort()
      })

      // Collect assistant output for persistence
      type ContentBlock = { type: "text"; content: string } | { type: "tool"; id: string; name: string; finalInput: string; result?: string }
      const assistantBlocks: ContentBlock[] = []
      let activeToolId: string | null = null
      try {
        for await (const sseEvent of runContainer({
          instancePath,
          message,
          sessionId,
          env: containerEnv,
          signal: abortController.signal,
          timeoutMs: 5 * 60 * 1000, // 5 minute timeout
        })) {
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
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e)
        send("error", msg)
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
