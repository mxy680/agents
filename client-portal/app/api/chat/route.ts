export const maxDuration = 1800

import { NextRequest } from "next/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { runLocal, buildSystemPrompt } from "@/lib/local-runner"
import { type ChatSSEEvent } from "@/lib/agent-events"
import { generateTitle } from "@/lib/auto-title"

type ContentBlock =
  | { type: "text"; content: string }
  | { type: "tool"; id: string; name: string; finalInput: string; result?: string }

export async function POST(request: NextRequest) {
  let code: string
  let message: string
  let conversationId: string | undefined
  let requestedAgent: string | undefined
  try {
    const body = await request.json()
    code = body.code
    message = body.message
    conversationId = body.conversationId
    requestedAgent = body.agentName
  } catch {
    return new Response(JSON.stringify({ error: "Invalid request" }), { status: 400 })
  }

  if (!code || !message) {
    return new Response(JSON.stringify({ error: "code and message required" }), { status: 400 })
  }

  if (message.length > 32_000) {
    return new Response(JSON.stringify({ error: "Message too long" }), { status: 400 })
  }

  const admin = createAdminClient()

  // Validate access code
  const { data: access } = await admin
    .from("client_access")
    .select("id, client_name, agent_name, agent_names, active")
    .eq("code", code)
    .eq("active", true)
    .single()

  if (!access) {
    return new Response(JSON.stringify({ error: "Invalid access code" }), { status: 401 })
  }

  // Use requested agent if valid, otherwise fall back to first assigned
  const allowedAgents: string[] = (access.agent_names as string[] | null)?.length
    ? (access.agent_names as string[])
    : [access.agent_name]
  const agentName = (requestedAgent && allowedAgents.includes(requestedAgent))
    ? requestedAgent
    : allowedAgents[0]

  // Create or validate conversation
  if (!conversationId) {
    const title = generateTitle(message)
    const { data: conv } = await admin
      .from("conversations")
      .insert({
        user_id: "00000000-0000-0000-0000-000000000000",
        agent_name: agentName,
        title,
      })
      .select("id")
      .single()
    conversationId = conv?.id
  }

  // Look up session_id for existing conversation
  let sessionId: string | undefined
  if (conversationId) {
    const { data: conv } = await admin
      .from("conversations")
      .select("session_id")
      .eq("id", conversationId)
      .single()
    if (conv?.session_id) {
      sessionId = conv.session_id
    }
  }

  // Persist user message
  if (conversationId) {
    await admin.from("conversation_messages").insert({
      conversation_id: conversationId,
      role: "user",
      blocks: [{ type: "text", content: message }],
    })
  }

  const systemPrompt = buildSystemPrompt(agentName, `You are chatting with ${access.client_name}. Be helpful, professional, and concise.`)

  // Stream response
  const stream = new ReadableStream({
    async start(controller) {
      const encoder = new TextEncoder()

      function send(event: string, data: string) {
        const lines = data.split("\n").map((l) => `data: ${l}`).join("\n")
        controller.enqueue(encoder.encode(`event: ${event}\n${lines}\n\n`))
      }

      const abortController = new AbortController()
      request.signal.addEventListener("abort", () => abortController.abort())

      const assistantBlocks: ContentBlock[] = []

      try {
        const localGen = runLocal({
          agentName,
          message,
          systemPrompt,
          sessionId,
          signal: abortController.signal,
          timeoutMs: 1_800_000,
        })

        for await (const sseEvent of localGen) {
          send(sseEvent.event, sseEvent.data)

          // Track text blocks for persistence
          if (sseEvent.event === "delta") {
            const last = assistantBlocks[assistantBlocks.length - 1]
            if (last?.type === "text") {
              assistantBlocks[assistantBlocks.length - 1] = { ...last, content: last.content + sseEvent.data }
            } else {
              assistantBlocks.push({ type: "text", content: sseEvent.data })
            }
          }

          // Track session_id
          if (sseEvent.event === "session" && conversationId) {
            admin.from("conversations")
              .update({ session_id: sseEvent.data })
              .eq("id", conversationId)
              .then(() => {})
          }

          // Send conversation_id to client on first event
          if (sseEvent.event === "session" && conversationId) {
            send("conversation", conversationId)
          }
        }

        // Persist assistant message
        if (conversationId && assistantBlocks.length > 0) {
          await admin.from("conversation_messages").insert({
            conversation_id: conversationId,
            role: "assistant",
            blocks: assistantBlocks,
          })
          await admin.from("conversations")
            .update({ updated_at: new Date().toISOString() })
            .eq("id", conversationId)
        }
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e)
        send("error", msg)
      } finally {
        controller.close()
      }
    },
  })

  return new Response(stream, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      Connection: "keep-alive",
    },
  })
}
